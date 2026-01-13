#!/bin/bash
#
# Infrastructure MCP Server - Cross-Platform Installer
# Supports: macOS (Intel/Apple Silicon), Linux
#
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Configuration
INSTALL_DIR="${INSTALL_DIR:-$HOME/.infra-mcp}"
CONFIG_DIR="$INSTALL_DIR/config"
KEYS_DIR="$INSTALL_DIR/keys"
LOGS_DIR="$INSTALL_DIR/logs"
BIN_DIR="$INSTALL_DIR/bin"

print_banner() {
    echo -e "${CYAN}"
    echo "╔══════════════════════════════════════════════════════════════════╗"
    echo "║                                                                  ║"
    echo "║           🚀 Infrastructure MCP Server Installer 🚀              ║"
    echo "║                                                                  ║"
    echo "║   Multi-Database Access for AI Assistants via SSH Tunnels       ║"
    echo "║                                                                  ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_os() {
    case "$(uname -s)" in
        Darwin*)
            OS="macos"
            if [[ $(uname -m) == "arm64" ]]; then
                ARCH="arm64"
            else
                ARCH="amd64"
            fi
            ;;
        Linux*)
            OS="linux"
            ARCH="amd64"
            ;;
        *)
            log_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    log_info "Detected OS: $OS ($ARCH)"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing=()
    
    # Check for autossh
    if ! command -v autossh &> /dev/null; then
        missing+=("autossh")
    fi
    
    if [ ${#missing[@]} -gt 0 ]; then
        log_warn "Missing dependencies: ${missing[*]}"
        
        if [ "$OS" == "macos" ]; then
            if command -v brew &> /dev/null; then
                log_info "Installing dependencies via Homebrew..."
                for dep in "${missing[@]}"; do
                    brew install "$dep"
                done
            else
                log_error "Homebrew not found. Please install Homebrew first:"
                echo "  /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
                exit 1
            fi
        elif [ "$OS" == "linux" ]; then
            log_info "Installing dependencies via apt..."
            sudo apt-get update
            sudo apt-get install -y "${missing[@]}"
        fi
    fi
    
    log_success "All dependencies installed"
}

create_directories() {
    log_info "Creating installation directories..."
    
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$KEYS_DIR"
    mkdir -p "$LOGS_DIR"
    mkdir -p "$BIN_DIR"
    
    # Set proper permissions for keys directory
    chmod 700 "$KEYS_DIR"
    
    log_success "Directories created at $INSTALL_DIR"
}

copy_binary() {
    log_info "Installing MCP server binary..."
    
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    
    # Determine the source bin directory
    local SOURCE_BIN_DIR=""
    if [ -d "$SCRIPT_DIR/../bin" ]; then
        SOURCE_BIN_DIR="$SCRIPT_DIR/../bin"
    elif [ -d "$SCRIPT_DIR/bin" ]; then
        SOURCE_BIN_DIR="$SCRIPT_DIR/bin"
    elif [ -f "$SCRIPT_DIR/server" ]; then
        SOURCE_BIN_DIR="$SCRIPT_DIR"
    else
        log_warn "Binary not found in distribution. Will attempt to download..."
        download_binary
        return
    fi
    
    # Copy the wrapper script as infra-mcp-server
    if [ -f "$SOURCE_BIN_DIR/server" ]; then
        cp "$SOURCE_BIN_DIR/server" "$BIN_DIR/infra-mcp-server"
        chmod +x "$BIN_DIR/infra-mcp-server"
        log_success "Wrapper script installed to $BIN_DIR/infra-mcp-server"
    fi
    
    # Copy ALL platform-specific binaries (required by the wrapper)
    local binaries_copied=0
    for binary in server-darwin-arm64 server-darwin-amd64 server-linux-amd64; do
        if [ -f "$SOURCE_BIN_DIR/$binary" ]; then
            cp "$SOURCE_BIN_DIR/$binary" "$BIN_DIR/"
            chmod +x "$BIN_DIR/$binary"
            log_info "  Copied $binary"
            ((binaries_copied++))
        fi
    done
    
    if [ $binaries_copied -eq 0 ]; then
        log_error "No platform binaries found in $SOURCE_BIN_DIR"
        log_error "Expected: server-darwin-arm64, server-darwin-amd64, or server-linux-amd64"
        exit 1
    fi
    
    log_success "Installed $binaries_copied platform binary(ies) to $BIN_DIR/"
}

download_binary() {
    log_info "Downloading pre-built binary..."
    
    # For now, we'll build from source or use Docker
    # In production, you'd host binaries on GitHub Releases
    
    if [ "$OS" == "macos" ]; then
        if [ "$ARCH" == "arm64" ]; then
            BINARY_URL="https://github.com/FreePeak/infra-mcp-server/releases/latest/download/server-darwin-arm64"
        else
            BINARY_URL="https://github.com/FreePeak/infra-mcp-server/releases/latest/download/server-darwin-amd64"
        fi
    else
        BINARY_URL="https://github.com/FreePeak/infra-mcp-server/releases/latest/download/server-linux-amd64"
    fi
    
    # Attempt download (this URL is placeholder - you'd set up actual releases)
    if curl -fsSL -o "$BIN_DIR/infra-mcp-server" "$BINARY_URL" 2>/dev/null; then
        chmod +x "$BIN_DIR/infra-mcp-server"
        log_success "Binary downloaded successfully"
    else
        log_warn "Could not download binary. Please build from source:"
        echo "  cd /path/to/infra-mcp-server && make build"
        echo "  Then copy bin/server to $BIN_DIR/infra-mcp-server"
    fi
}

setup_config_template() {
    log_info "Creating configuration template..."
    
    cat > "$CONFIG_DIR/config.json.template" << 'CONFIGEOF'
{
  "listen_address": "0.0.0.0",
  "port": 9092,
  "connections": [
    {
      "id": "ts_stage",
      "type": "postgres",
      "host": "localhost",
      "port": 4001,
      "name": "transaction_service",
      "user": "ro_user",
      "password": "YOUR_PASSWORD_HERE",
      "display_name": "Transaction Service Staging",
      "project": "transaction-service",
      "environment": "staging",
      "description": "Staging database for transaction service",
      "tags": ["staging", "transactions"]
    },
    {
      "id": "ts_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 5001,
      "name": "transaction_service",
      "user": "ro_user",
      "password": "YOUR_PASSWORD_HERE",
      "options": {
        "default_transaction_read_only": "on"
      },
      "display_name": "Transaction Service Production",
      "project": "transaction-service",
      "environment": "production",
      "description": "Production database (read-only)",
      "tags": ["production", "transactions"]
    }
  ],
  "aws_profiles": []
}
CONFIGEOF

    if [ ! -f "$CONFIG_DIR/config.json" ]; then
        cp "$CONFIG_DIR/config.json.template" "$CONFIG_DIR/config.json"
        log_warn "Created config.json - please edit with your actual credentials"
    else
        log_info "Existing config.json preserved"
    fi
    
    log_success "Configuration template created"
}

setup_tunnels_macos() {
    log_info "Setting up SSH tunnel services for macOS (launchd)..."
    
    LAUNCHD_DIR="$HOME/Library/LaunchAgents"
    mkdir -p "$LAUNCHD_DIR"
    
    # Create template for Financial Services Production tunnel
    cat > "$LAUNCHD_DIR/com.infra-mcp.tunnel.fs-prod.plist" << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.infra-mcp.tunnel.fs-prod</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/autossh</string>
        <string>-M</string>
        <string>0</string>
        <string>-o</string>
        <string>ExitOnForwardFailure=yes</string>
        <string>-o</string>
        <string>ServerAliveInterval=30</string>
        <string>-o</string>
        <string>ServerAliveCountMax=3</string>
        <string>-i</string>
        <string>$KEYS_DIR/financial-services-prod-bastion.pem</string>
        <string>-L</string>
        <string>0.0.0.0:5001:prod-transaction-service-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432</string>
        <string>-L</string>
        <string>0.0.0.0:5002:prod-wallet-service-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432</string>
        <string>-L</string>
        <string>0.0.0.0:5003:prod-payment-gateway-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432</string>
        <string>-L</string>
        <string>0.0.0.0:5004:prod-ledger-service-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432</string>
        <string>-N</string>
        <string>ec2-user@3.93.179.5</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>AUTOSSH_GATETIME</key>
        <string>0</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$LOGS_DIR/tunnel-fs-prod.log</string>
    <key>StandardErrorPath</key>
    <string>$LOGS_DIR/tunnel-fs-prod.error.log</string>
</dict>
</plist>
PLISTEOF

    # Create template for Financial Services Staging tunnel
    cat > "$LAUNCHD_DIR/com.infra-mcp.tunnel.fs-staging.plist" << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.infra-mcp.tunnel.fs-staging</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/autossh</string>
        <string>-M</string>
        <string>0</string>
        <string>-o</string>
        <string>ExitOnForwardFailure=yes</string>
        <string>-o</string>
        <string>ServerAliveInterval=30</string>
        <string>-o</string>
        <string>ServerAliveCountMax=3</string>
        <string>-i</string>
        <string>$KEYS_DIR/financial-services-staging-bastion.pem</string>
        <string>-L</string>
        <string>0.0.0.0:4001:stg-transaction-service-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432</string>
        <string>-L</string>
        <string>0.0.0.0:4002:stg-wallet-service-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432</string>
        <string>-L</string>
        <string>0.0.0.0:4003:stg-payment-gateway-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432</string>
        <string>-L</string>
        <string>0.0.0.0:4004:stg-ledger-service-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432</string>
        <string>-N</string>
        <string>ec2-user@44.193.202.29</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>AUTOSSH_GATETIME</key>
        <string>0</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$LOGS_DIR/tunnel-fs-staging.log</string>
    <key>StandardErrorPath</key>
    <string>$LOGS_DIR/tunnel-fs-staging.error.log</string>
</dict>
</plist>
PLISTEOF

    # Create template for Minnect Production tunnel
    cat > "$LAUNCHD_DIR/com.infra-mcp.tunnel.minnect-prod.plist" << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.infra-mcp.tunnel.minnect-prod</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/autossh</string>
        <string>-M</string>
        <string>0</string>
        <string>-o</string>
        <string>ExitOnForwardFailure=yes</string>
        <string>-o</string>
        <string>ServerAliveInterval=30</string>
        <string>-o</string>
        <string>ServerAliveCountMax=3</string>
        <string>-i</string>
        <string>$KEYS_DIR/minnect-prod-bastion.pem</string>
        <string>-L</string>
        <string>0.0.0.0:7001:prod-profiles-service.cluster-ci2a7bvoh5dz.us-east-1.rds.amazonaws.com:5432</string>
        <string>-L</string>
        <string>0.0.0.0:7002:valuetainment-production.ci2a7bvoh5dz.us-east-1.rds.amazonaws.com:5432</string>
        <string>-N</string>
        <string>ec2-user@18.212.176.229</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>AUTOSSH_GATETIME</key>
        <string>0</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$LOGS_DIR/tunnel-minnect-prod.log</string>
    <key>StandardErrorPath</key>
    <string>$LOGS_DIR/tunnel-minnect-prod.error.log</string>
</dict>
</plist>
PLISTEOF

    # Create template for Minnect Staging tunnel
    cat > "$LAUNCHD_DIR/com.infra-mcp.tunnel.minnect-staging.plist" << PLISTEOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.infra-mcp.tunnel.minnect-staging</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/autossh</string>
        <string>-M</string>
        <string>0</string>
        <string>-o</string>
        <string>ExitOnForwardFailure=yes</string>
        <string>-o</string>
        <string>ServerAliveInterval=30</string>
        <string>-o</string>
        <string>ServerAliveCountMax=3</string>
        <string>-i</string>
        <string>$KEYS_DIR/minnect-staging-bastion.pem</string>
        <string>-L</string>
        <string>0.0.0.0:6001:staging-profiles-service.cluster-cdwsi6ucuxab.us-east-1.rds.amazonaws.com:5432</string>
        <string>-N</string>
        <string>ec2-user@3.84.134.13</string>
    </array>
    <key>EnvironmentVariables</key>
    <dict>
        <key>AUTOSSH_GATETIME</key>
        <string>0</string>
    </dict>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$LOGS_DIR/tunnel-minnect-staging.log</string>
    <key>StandardErrorPath</key>
    <string>$LOGS_DIR/tunnel-minnect-staging.error.log</string>
</dict>
</plist>
PLISTEOF

    log_success "LaunchAgent plist files created"
    log_warn "Tunnel services created but NOT started. See post-install instructions."
}

setup_tunnels_linux() {
    log_info "Setting up SSH tunnel services for Linux (systemd)..."
    
    SYSTEMD_DIR="/etc/systemd/system"
    
    # Create Financial Services Production tunnel service
    sudo tee "$SYSTEMD_DIR/autossh-fs-production-mcp-tunnel.service" > /dev/null << SERVICEEOF
[Unit]
Description=AutoSSH Multi-RDS Tunnel for MCP Server Prod
After=network.target

[Service]
User=$USER
Environment="AUTOSSH_GATETIME=0"
ExecStart=/usr/bin/autossh -M 0 \\
  -o "ExitOnForwardFailure=yes" \\
  -o "ServerAliveInterval=30" \\
  -o "ServerAliveCountMax=3" \\
  -i $KEYS_DIR/financial-services-prod-bastion.pem \\
  -L 0.0.0.0:5001:prod-transaction-service-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432 \\
  -L 0.0.0.0:5002:prod-wallet-service-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432 \\
  -L 0.0.0.0:5003:prod-payment-gateway-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432 \\
  -L 0.0.0.0:5004:prod-ledger-service-rds.cluster-ro-ct8uk8aumjtb.us-east-1.rds.amazonaws.com:5432 \\
  ec2-user@3.93.179.5 -N
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
SERVICEEOF

    # Create Financial Services Staging tunnel service
    sudo tee "$SYSTEMD_DIR/autossh-fs-staging-mcp-tunnel.service" > /dev/null << SERVICEEOF
[Unit]
Description=AutoSSH Multi-RDS Tunnel for MCP Server Staging
After=network.target

[Service]
User=$USER
Environment="AUTOSSH_GATETIME=0"
ExecStart=/usr/bin/autossh -M 0 \\
  -o "ExitOnForwardFailure=yes" \\
  -o "ServerAliveInterval=30" \\
  -o "ServerAliveCountMax=3" \\
  -i $KEYS_DIR/financial-services-staging-bastion.pem \\
  -L 0.0.0.0:4001:stg-transaction-service-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432 \\
  -L 0.0.0.0:4002:stg-wallet-service-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432 \\
  -L 0.0.0.0:4003:stg-payment-gateway-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432 \\
  -L 0.0.0.0:4004:stg-ledger-service-rds.cluster-ro-ckd6e2ke242p.us-east-1.rds.amazonaws.com:5432 \\
  ec2-user@44.193.202.29 -N
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
SERVICEEOF

    sudo systemctl daemon-reload
    
    log_success "Systemd service files created"
}

setup_cursor_config() {
    log_info "Setting up Cursor IDE integration..."
    
    CURSOR_CONFIG_DIR="$HOME/.cursor"
    mkdir -p "$CURSOR_CONFIG_DIR"
    
    # Create/update mcp.json
    if [ -f "$CURSOR_CONFIG_DIR/mcp.json" ]; then
        log_info "Existing mcp.json found, creating backup..."
        cp "$CURSOR_CONFIG_DIR/mcp.json" "$CURSOR_CONFIG_DIR/mcp.json.backup.$(date +%Y%m%d%H%M%S)"
    fi
    
    cat > "$CURSOR_CONFIG_DIR/mcp.json" << MCPEOF
{
  "mcpServers": {
    "infra-mcp-server": {
      "command": "$BIN_DIR/infra-mcp-server",
      "args": ["-t", "stdio", "-c", "$CONFIG_DIR/config.json"]
    }
  }
}
MCPEOF

    log_success "Cursor MCP configuration created at $CURSOR_CONFIG_DIR/mcp.json"
}

create_control_scripts() {
    log_info "Creating control scripts..."
    
    # Create tunnel control script
    cat > "$BIN_DIR/mcp-tunnels" << 'CONTROLEOF'
#!/bin/bash
#
# MCP Tunnel Control Script
#

INSTALL_DIR="${INSTALL_DIR:-$HOME/.infra-mcp}"
KEYS_DIR="$INSTALL_DIR/keys"
LOGS_DIR="$INSTALL_DIR/logs"

case "$(uname -s)" in
    Darwin*) OS="macos" ;;
    Linux*)  OS="linux" ;;
esac

usage() {
    echo "Usage: mcp-tunnels <command> [tunnel-name]"
    echo ""
    echo "Commands:"
    echo "  start [name]     Start tunnel(s). If name omitted, starts all."
    echo "  stop [name]      Stop tunnel(s). If name omitted, stops all."
    echo "  restart [name]   Restart tunnel(s)"
    echo "  status           Show status of all tunnels"
    echo "  logs [name]      Show logs for a tunnel"
    echo ""
    echo "Tunnel names: fs-prod, fs-staging, minnect-prod, minnect-staging"
    echo ""
    echo "Examples:"
    echo "  mcp-tunnels start              # Start all tunnels"
    echo "  mcp-tunnels start fs-prod      # Start only FS production tunnel"
    echo "  mcp-tunnels status             # Check all tunnel status"
}

get_tunnel_names() {
    echo "fs-prod fs-staging minnect-prod minnect-staging"
}

start_tunnel_macos() {
    local name=$1
    local plist="$HOME/Library/LaunchAgents/com.infra-mcp.tunnel.${name}.plist"
    
    if [ ! -f "$plist" ]; then
        echo "Tunnel $name not configured. Plist not found: $plist"
        return 1
    fi
    
    # Check if key exists
    local key_name
    case "$name" in
        fs-prod)      key_name="financial-services-prod-bastion.pem" ;;
        fs-staging)   key_name="financial-services-staging-bastion.pem" ;;
        minnect-prod) key_name="minnect-prod-bastion.pem" ;;
        minnect-staging) key_name="minnect-staging-bastion.pem" ;;
    esac
    
    if [ ! -f "$KEYS_DIR/$key_name" ]; then
        echo "⚠️  SSH key not found: $KEYS_DIR/$key_name"
        echo "   Please copy the key file before starting this tunnel."
        return 1
    fi
    
    launchctl load "$plist" 2>/dev/null || true
    launchctl start "com.infra-mcp.tunnel.${name}" 2>/dev/null || true
    echo "✓ Started tunnel: $name"
}

stop_tunnel_macos() {
    local name=$1
    launchctl stop "com.infra-mcp.tunnel.${name}" 2>/dev/null || true
    launchctl unload "$HOME/Library/LaunchAgents/com.infra-mcp.tunnel.${name}.plist" 2>/dev/null || true
    echo "✓ Stopped tunnel: $name"
}

start_tunnel_linux() {
    local name=$1
    local service
    
    case "$name" in
        fs-prod)      service="autossh-fs-production-mcp-tunnel" ;;
        fs-staging)   service="autossh-fs-staging-mcp-tunnel" ;;
        minnect-prod) service="autossh-minnect-production-mcp-tunnel" ;;
        minnect-staging) service="autossh-minnect-staging-mcp-tunnel" ;;
    esac
    
    sudo systemctl start "$service"
    echo "✓ Started tunnel: $name"
}

stop_tunnel_linux() {
    local name=$1
    local service
    
    case "$name" in
        fs-prod)      service="autossh-fs-production-mcp-tunnel" ;;
        fs-staging)   service="autossh-fs-staging-mcp-tunnel" ;;
        minnect-prod) service="autossh-minnect-production-mcp-tunnel" ;;
        minnect-staging) service="autossh-minnect-staging-mcp-tunnel" ;;
    esac
    
    sudo systemctl stop "$service"
    echo "✓ Stopped tunnel: $name"
}

status_macos() {
    echo "Tunnel Status (macOS):"
    echo "========================"
    
    for name in $(get_tunnel_names); do
        local label="com.infra-mcp.tunnel.${name}"
        if launchctl list 2>/dev/null | grep -q "$label"; then
            local pid=$(launchctl list 2>/dev/null | grep "$label" | awk '{print $1}')
            if [ "$pid" != "-" ] && [ -n "$pid" ]; then
                echo "  ✅ $name (PID: $pid)"
            else
                echo "  ⚠️  $name (loaded but not running)"
            fi
        else
            echo "  ❌ $name (not loaded)"
        fi
    done
    
    echo ""
    echo "Port Status:"
    echo "------------"
    
    check_port() {
        local port=$1
        local desc=$2
        if lsof -i :$port -sTCP:LISTEN > /dev/null 2>&1; then
            echo "  ✅ :$port - $desc"
        else
            echo "  ❌ :$port - $desc (not listening)"
        fi
    }
    
    # Staging ports (4xxx)
    check_port 4001 "Transaction Service (staging)"
    check_port 4002 "Wallet Service (staging)"
    check_port 4003 "Payment Gateway (staging)"
    check_port 4004 "Ledger Service (staging)"
    
    # Production ports (5xxx)
    check_port 5001 "Transaction Service (prod)"
    check_port 5002 "Wallet Service (prod)"
    check_port 5003 "Payment Gateway (prod)"
    check_port 5004 "Ledger Service (prod)"
    
    # Minnect ports (6xxx, 7xxx)
    check_port 6001 "Profile Service (staging)"
    check_port 7001 "Profile Service (prod)"
    check_port 7002 "Minnect DB (prod)"
}

status_linux() {
    echo "Tunnel Status (Linux):"
    echo "======================"
    
    for name in $(get_tunnel_names); do
        local service
        case "$name" in
            fs-prod)      service="autossh-fs-production-mcp-tunnel" ;;
            fs-staging)   service="autossh-fs-staging-mcp-tunnel" ;;
            minnect-prod) service="autossh-minnect-production-mcp-tunnel" ;;
            minnect-staging) service="autossh-minnect-staging-mcp-tunnel" ;;
        esac
        
        if systemctl is-active --quiet "$service" 2>/dev/null; then
            echo "  ✅ $name (active)"
        else
            echo "  ❌ $name (inactive)"
        fi
    done
    
    echo ""
    echo "Port Status:"
    echo "------------"
    ss -tlnp 2>/dev/null | grep -E ':400[1-4]|:500[1-4]|:600[1]|:700[1-2]' | while read line; do
        echo "  $line"
    done
}

case "$1" in
    start)
        if [ -n "$2" ]; then
            [ "$OS" == "macos" ] && start_tunnel_macos "$2" || start_tunnel_linux "$2"
        else
            for name in $(get_tunnel_names); do
                [ "$OS" == "macos" ] && start_tunnel_macos "$name" || start_tunnel_linux "$name"
            done
        fi
        ;;
    stop)
        if [ -n "$2" ]; then
            [ "$OS" == "macos" ] && stop_tunnel_macos "$2" || stop_tunnel_linux "$2"
        else
            for name in $(get_tunnel_names); do
                [ "$OS" == "macos" ] && stop_tunnel_macos "$name" || stop_tunnel_linux "$name"
            done
        fi
        ;;
    restart)
        $0 stop $2
        sleep 2
        $0 start $2
        ;;
    status)
        [ "$OS" == "macos" ] && status_macos || status_linux
        ;;
    logs)
        if [ -n "$2" ]; then
            tail -f "$LOGS_DIR/tunnel-$2.log" "$LOGS_DIR/tunnel-$2.error.log"
        else
            echo "Please specify a tunnel name. Example: mcp-tunnels logs fs-prod"
        fi
        ;;
    *)
        usage
        ;;
esac
CONTROLEOF
    chmod +x "$BIN_DIR/mcp-tunnels"

    # Create main launcher script
    cat > "$BIN_DIR/mcp-start" << 'STARTEOF'
#!/bin/bash
#
# Start MCP Server for Cursor
#

INSTALL_DIR="${INSTALL_DIR:-$HOME/.infra-mcp}"
CONFIG_DIR="$INSTALL_DIR/config"
BIN_DIR="$INSTALL_DIR/bin"
LOGS_DIR="$INSTALL_DIR/logs"

# Generate timestamp for log file
TIMESTAMP=$(date +"%Y%m%d-%H%M%S")
LOG_FILE="$LOGS_DIR/mcp-server-$TIMESTAMP.log"

echo "Starting Infrastructure MCP Server..." >&2
echo "Config: $CONFIG_DIR/config.json" >&2
echo "Log: $LOG_FILE" >&2

exec "$BIN_DIR/infra-mcp-server" \
    -t stdio \
    -c "$CONFIG_DIR/config.json" \
    2> >(tee -a "$LOG_FILE" >&2)
STARTEOF
    chmod +x "$BIN_DIR/mcp-start"

    log_success "Control scripts created"
}

add_to_path() {
    log_info "Adding tools to PATH..."
    
    local shell_rc
    case "$SHELL" in
        */zsh)  shell_rc="$HOME/.zshrc" ;;
        */bash) shell_rc="$HOME/.bashrc" ;;
        *)      shell_rc="$HOME/.profile" ;;
    esac
    
    local path_line="export PATH=\"\$PATH:$BIN_DIR\""
    
    if ! grep -q "$BIN_DIR" "$shell_rc" 2>/dev/null; then
        echo "" >> "$shell_rc"
        echo "# Infrastructure MCP Server" >> "$shell_rc"
        echo "$path_line" >> "$shell_rc"
        log_success "Added $BIN_DIR to PATH in $shell_rc"
    else
        log_info "PATH already configured in $shell_rc"
    fi
}

print_post_install() {
    echo ""
    echo -e "${GREEN}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║                  Installation Complete! 🎉                       ║${NC}"
    echo -e "${GREEN}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BOLD}Next Steps:${NC}"
    echo ""
    echo -e "${YELLOW}1. Reload your shell (REQUIRED for commands to work):${NC}"
    echo ""
    echo "   Option A: Open a NEW terminal window"
    echo ""
    echo "   Option B: Run this in your current terminal:"
    case "$SHELL" in
        */zsh)  echo "      source ~/.zshrc" ;;
        */bash) echo "      source ~/.bashrc" ;;
        *)      echo "      source ~/.profile" ;;
    esac
    echo ""
    echo -e "${YELLOW}2. Copy SSH keys to:${NC}"
    echo "   $KEYS_DIR/"
    echo ""
    echo "   Required keys:"
    echo "   - financial-services-prod-bastion.pem"
    echo "   - financial-services-staging-bastion.pem"
    echo "   - minnect-prod-bastion.pem"
    echo "   - minnect-staging-bastion.pem"
    echo ""
    echo "   Then set permissions: chmod 600 $KEYS_DIR/*.pem"
    echo ""
    echo -e "${YELLOW}3. Copy your configuration:${NC}"
    echo "   cp /path/to/your/config.json $CONFIG_DIR/config.json"
    echo ""
    echo -e "${YELLOW}4. Start SSH tunnels:${NC}"
    echo "   mcp-tunnels start"
    echo ""
    echo -e "${YELLOW}5. Restart Cursor IDE${NC}"
    echo "   The MCP server will auto-connect when Cursor starts."
    echo ""
    echo -e "${BOLD}Useful Commands:${NC}"
    echo "   mcp-tunnels status    # Check tunnel status"
    echo "   mcp-tunnels start     # Start all tunnels"
    echo "   mcp-tunnels stop      # Stop all tunnels"
    echo "   mcp-tunnels logs X    # View tunnel logs"
    echo ""
    echo -e "${CYAN}Installation Directory: $INSTALL_DIR${NC}"
    echo ""
}

# Main installation flow
main() {
    print_banner
    detect_os
    check_dependencies
    create_directories
    copy_binary
    setup_config_template
    
    if [ "$OS" == "macos" ]; then
        setup_tunnels_macos
    else
        setup_tunnels_linux
    fi
    
    setup_cursor_config
    create_control_scripts
    add_to_path
    print_post_install
}

# Run main
main "$@"
