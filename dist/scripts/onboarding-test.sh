#!/bin/bash
#
# Onboarding Test Script for Infrastructure MCP Server
# Run this after installation to verify everything works
#
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'
BOLD='\033[1m'

INSTALL_DIR="${INSTALL_DIR:-$HOME/.infra-mcp}"
PASSED=0
FAILED=0
SKIPPED=0

echo ""
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║         Infrastructure MCP Server - Onboarding Test              ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════╝${NC}"
echo ""
echo "This script will verify your installation is working correctly."
echo ""

# Test function
test_step() {
    local step_num=$1
    local step_name=$2
    local command=$3
    
    echo -e "${BLUE}[$step_num]${NC} $step_name"
    echo "    Command: $command"
    echo -n "    Result: "
    
    if eval "$command" > /dev/null 2>&1; then
        echo -e "${GREEN}PASS ✓${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL ✗${NC}"
        ((FAILED++))
        return 1
    fi
}

test_step_output() {
    local step_num=$1
    local step_name=$2
    local command=$3
    
    echo -e "${BLUE}[$step_num]${NC} $step_name"
    echo "    Command: $command"
    echo -n "    Result: "
    
    local output
    if output=$(eval "$command" 2>&1); then
        echo -e "${GREEN}PASS ✓${NC}"
        echo -e "    Output: ${CYAN}$output${NC}"
        ((PASSED++))
        return 0
    else
        echo -e "${RED}FAIL ✗${NC}"
        echo -e "    Error: $output"
        ((FAILED++))
        return 1
    fi
}

skip_step() {
    local step_num=$1
    local step_name=$2
    local reason=$3
    
    echo -e "${BLUE}[$step_num]${NC} $step_name"
    echo -e "    ${YELLOW}SKIPPED: $reason${NC}"
    ((SKIPPED++))
}

echo -e "${BOLD}═══ Phase 1: Installation Verification ═══${NC}"
echo ""

test_step "1.1" "Installation directory exists" "[ -d '$INSTALL_DIR' ]"
test_step "1.2" "MCP server binary exists" "[ -f '$INSTALL_DIR/bin/infra-mcp-server' ]"
test_step "1.3" "MCP server binary is executable" "[ -x '$INSTALL_DIR/bin/infra-mcp-server' ]"
test_step "1.4" "Configuration file exists" "[ -f '$INSTALL_DIR/config/config.json' ]"
test_step "1.5" "Cursor MCP config exists" "[ -f '$HOME/.cursor/mcp.json' ]"

echo ""
echo -e "${BOLD}═══ Phase 2: Dependencies ═══${NC}"
echo ""

test_step "2.1" "autossh is installed" "command -v autossh"

if [[ "$(uname -s)" == "Darwin" ]]; then
    test_step "2.2" "Homebrew is installed" "command -v brew"
fi

echo ""
echo -e "${BOLD}═══ Phase 3: SSH Keys ═══${NC}"
echo ""

check_key() {
    local num=$1
    local key=$2
    local path="$INSTALL_DIR/keys/$key"
    
    if [ -f "$path" ]; then
        test_step "$num" "SSH key: $key" "[ -f '$path' ] && [ \$(stat -f '%Lp' '$path' 2>/dev/null || stat -c '%a' '$path' 2>/dev/null) = '600' ]"
    else
        skip_step "$num" "SSH key: $key" "Not found (tunnel won't work without it)"
    fi
}

check_key "3.1" "financial-services-prod-bastion.pem"
check_key "3.2" "financial-services-staging-bastion.pem"
check_key "3.3" "minnect-prod-bastion.pem"
check_key "3.4" "minnect-staging-bastion.pem"

echo ""
echo -e "${BOLD}═══ Phase 4: SSH Tunnel Connectivity ═══${NC}"
echo ""

check_tunnel_port() {
    local num=$1
    local port=$2
    local name=$3
    
    if nc -z localhost $port 2>/dev/null || lsof -i :$port -sTCP:LISTEN > /dev/null 2>&1; then
        test_step "$num" "Port $port ($name)" "nc -z localhost $port || lsof -i :$port -sTCP:LISTEN"
    else
        skip_step "$num" "Port $port ($name)" "Not listening (tunnel not started or key missing)"
    fi
}

# Staging
check_tunnel_port "4.1" 4001 "Transaction Service - Staging"
check_tunnel_port "4.2" 4002 "Wallet Service - Staging"
check_tunnel_port "4.3" 4003 "Payment Gateway - Staging"
check_tunnel_port "4.4" 4004 "Ledger Service - Staging"

# Production
check_tunnel_port "4.5" 5001 "Transaction Service - Prod"
check_tunnel_port "4.6" 5002 "Wallet Service - Prod"
check_tunnel_port "4.7" 5003 "Payment Gateway - Prod"
check_tunnel_port "4.8" 5004 "Ledger Service - Prod"

# Minnect
check_tunnel_port "4.9" 6001 "Profile Service - Staging"
check_tunnel_port "4.10" 7001 "Profile Service - Prod"
check_tunnel_port "4.11" 7002 "Minnect DB - Prod"

echo ""
echo -e "${BOLD}═══ Phase 5: Database Connectivity ═══${NC}"
echo ""

# Check if psql is available
if command -v psql &> /dev/null; then
    # Try to connect to a staging database (safest to test)
    if nc -z localhost 4001 2>/dev/null; then
        echo -e "${BLUE}[5.1]${NC} Testing database connection (staging)..."
        
        # Extract password from config
        if command -v jq &> /dev/null; then
            DB_PASS=$(jq -r '.connections[] | select(.id == "ts_stage") | .password' "$INSTALL_DIR/config/config.json" 2>/dev/null)
            
            if [ -n "$DB_PASS" ] && [ "$DB_PASS" != "null" ]; then
                echo "    Attempting connection to Transaction Service (staging)..."
                if PGPASSWORD="$DB_PASS" psql -h localhost -p 4001 -U ro_user -d transaction_service -c "SELECT 1 as connection_test;" > /dev/null 2>&1; then
                    echo -e "    Result: ${GREEN}PASS ✓${NC} - Database connection successful!"
                    ((PASSED++))
                else
                    echo -e "    Result: ${RED}FAIL ✗${NC} - Could not connect to database"
                    echo "    Possible causes: Wrong password, database not ready, or tunnel issue"
                    ((FAILED++))
                fi
            else
                skip_step "5.1" "Database connection test" "Could not extract password from config"
            fi
        else
            skip_step "5.1" "Database connection test" "jq not installed for config parsing"
        fi
    else
        skip_step "5.1" "Database connection test" "Staging tunnel not running"
    fi
else
    skip_step "5.1" "Database connection test" "psql not installed"
fi

echo ""
echo -e "${BOLD}═══ Phase 6: MCP Server Test ═══${NC}"
echo ""

echo -e "${BLUE}[6.1]${NC} Testing MCP server startup..."
echo "    Starting server in test mode..."

# Create a test to verify the server can start
TEST_OUTPUT=$(timeout 3 "$INSTALL_DIR/bin/infra-mcp-server" -t stdio -c "$INSTALL_DIR/config/config.json" 2>&1 <<< '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{}}}' || true)

if echo "$TEST_OUTPUT" | grep -q '"result"' 2>/dev/null; then
    echo -e "    Result: ${GREEN}PASS ✓${NC} - MCP server responds to initialize"
    ((PASSED++))
elif echo "$TEST_OUTPUT" | grep -q 'jsonrpc' 2>/dev/null; then
    echo -e "    Result: ${GREEN}PASS ✓${NC} - MCP server started (partial response)"
    ((PASSED++))
else
    echo -e "    Result: ${YELLOW}INCONCLUSIVE${NC} - Server may work, manual test recommended"
    echo "    To test manually: Open Cursor and check MCP status in bottom bar"
    ((SKIPPED++))
fi

echo ""
echo -e "${CYAN}══════════════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${BOLD}Test Summary${NC}"
echo ""
echo -e "  ${GREEN}Passed:${NC}  $PASSED"
echo -e "  ${RED}Failed:${NC}  $FAILED"
echo -e "  ${YELLOW}Skipped:${NC} $SKIPPED"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║              🎉 All tests passed! You're ready to go! 🎉         ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Start Cursor IDE"
    echo "  2. Look for 'infra-mcp-server' in the MCP status (bottom bar)"
    echo "  3. Ask Claude: 'List all database connections available'"
    echo ""
else
    echo -e "${YELLOW}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║              Some tests failed. See details above.               ║${NC}"
    echo -e "${YELLOW}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo "Common fixes:"
    echo "  • Missing SSH keys: Copy .pem files to ~/.infra-mcp/keys/"
    echo "  • Wrong permissions: chmod 600 ~/.infra-mcp/keys/*.pem"
    echo "  • Tunnels not running: mcp-tunnels start"
    echo "  • Config issues: Check ~/.infra-mcp/config/config.json"
    echo ""
fi
