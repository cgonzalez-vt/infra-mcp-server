#!/bin/bash
#
# Health Check for Infrastructure MCP Server
# Run this to diagnose issues
#
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

INSTALL_DIR="${INSTALL_DIR:-$HOME/.infra-mcp}"

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║           Infrastructure MCP Server Health Check                 ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

errors=0
warnings=0

check() {
    local name=$1
    local result=$2
    local msg=$3
    
    if [ "$result" = "ok" ]; then
        echo -e "  ${GREEN}✓${NC} $name"
    elif [ "$result" = "warn" ]; then
        echo -e "  ${YELLOW}⚠${NC} $name: $msg"
        ((warnings++))
    else
        echo -e "  ${RED}✗${NC} $name: $msg"
        ((errors++))
    fi
}

# Check OS
echo -e "${BLUE}System Information${NC}"
echo "  OS: $(uname -s) $(uname -m)"
echo "  User: $USER"
echo ""

# Check installation directory
echo -e "${BLUE}Installation Check${NC}"

if [ -d "$INSTALL_DIR" ]; then
    check "Installation directory" "ok"
else
    check "Installation directory" "fail" "Not found at $INSTALL_DIR"
fi

if [ -f "$INSTALL_DIR/bin/infra-mcp-server" ]; then
    check "MCP server binary" "ok"
else
    check "MCP server binary" "fail" "Not found"
fi

if [ -x "$INSTALL_DIR/bin/infra-mcp-server" ]; then
    check "Binary executable" "ok"
else
    check "Binary executable" "fail" "Not executable"
fi

if [ -f "$INSTALL_DIR/config/config.json" ]; then
    check "Configuration file" "ok"
else
    check "Configuration file" "fail" "Not found at $INSTALL_DIR/config/config.json"
fi

echo ""

# Check Cursor integration
echo -e "${BLUE}Cursor Integration${NC}"

if [ -f "$HOME/.cursor/mcp.json" ]; then
    check "Cursor MCP config" "ok"
    
    if grep -q "infra-mcp-server" "$HOME/.cursor/mcp.json"; then
        check "MCP server registered" "ok"
    else
        check "MCP server registered" "fail" "infra-mcp-server not found in mcp.json"
    fi
else
    check "Cursor MCP config" "fail" "Not found at ~/.cursor/mcp.json"
fi

echo ""

# Check SSH keys
echo -e "${BLUE}SSH Keys${NC}"

check_key() {
    local key=$1
    local path="$INSTALL_DIR/keys/$key"
    
    if [ -f "$path" ]; then
        local perms=$(stat -f "%Lp" "$path" 2>/dev/null || stat -c "%a" "$path" 2>/dev/null)
        if [ "$perms" = "600" ]; then
            check "$key" "ok"
        else
            check "$key" "warn" "Permissions should be 600, got $perms"
        fi
    else
        check "$key" "warn" "Not found (may not be needed)"
    fi
}

check_key "financial-services-prod-bastion.pem"
check_key "financial-services-staging-bastion.pem"
check_key "minnect-prod-bastion.pem"
check_key "minnect-staging-bastion.pem"

echo ""

# Check autossh
echo -e "${BLUE}Dependencies${NC}"

if command -v autossh &> /dev/null; then
    check "autossh" "ok"
else
    check "autossh" "fail" "Not installed"
fi

echo ""

# Check tunnels
echo -e "${BLUE}SSH Tunnels${NC}"

check_port() {
    local port=$1
    local name=$2
    
    if lsof -i :$port -sTCP:LISTEN > /dev/null 2>&1 || nc -z localhost $port 2>/dev/null; then
        check "$name (:$port)" "ok"
    else
        check "$name (:$port)" "warn" "Not listening"
    fi
}

# Check staging ports
check_port 4001 "Transaction Service (staging)"
check_port 4002 "Wallet Service (staging)"
check_port 4003 "Payment Gateway (staging)"
check_port 4004 "Ledger Service (staging)"

# Check production ports
check_port 5001 "Transaction Service (prod)"
check_port 5002 "Wallet Service (prod)"
check_port 5003 "Payment Gateway (prod)"
check_port 5004 "Ledger Service (prod)"

# Check Minnect ports
check_port 6001 "Profile Service (staging)"
check_port 7001 "Profile Service (prod)"
check_port 7002 "Minnect DB (prod)"

echo ""

# Check config validity
echo -e "${BLUE}Configuration Validation${NC}"

if [ -f "$INSTALL_DIR/config/config.json" ]; then
    if command -v jq &> /dev/null; then
        if jq . "$INSTALL_DIR/config/config.json" > /dev/null 2>&1; then
            check "JSON valid" "ok"
            
            conn_count=$(jq '.connections | length' "$INSTALL_DIR/config/config.json")
            echo "  📊 Configured connections: $conn_count"
        else
            check "JSON valid" "fail" "Invalid JSON syntax"
        fi
    else
        check "JSON validation" "warn" "jq not installed, skipping validation"
    fi
fi

echo ""

# Summary
echo "══════════════════════════════════════════════════════════════════"
if [ $errors -eq 0 ] && [ $warnings -eq 0 ]; then
    echo -e "${GREEN}All checks passed! ✓${NC}"
elif [ $errors -eq 0 ]; then
    echo -e "${YELLOW}$warnings warning(s), but should still work${NC}"
else
    echo -e "${RED}$errors error(s), $warnings warning(s)${NC}"
    echo ""
    echo "Run the installer to fix issues:"
    echo "  ./install.sh"
fi
echo ""
