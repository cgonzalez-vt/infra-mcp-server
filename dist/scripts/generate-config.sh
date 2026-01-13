#!/bin/bash
#
# Configuration Generator for Infrastructure MCP Server
# Interactive wizard to create config.json
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
CONFIG_DIR="$INSTALL_DIR/config"
OUTPUT_FILE="${1:-$CONFIG_DIR/config.json}"

echo -e "${CYAN}"
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║           Infrastructure MCP Server Config Generator             ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Collect database password
echo -e "${BOLD}Database Configuration${NC}"
echo ""
read -sp "Enter the read-only database password: " DB_PASSWORD
echo ""
echo ""

# Ask which environments to enable
echo -e "${BOLD}Which environments do you want to enable?${NC}"
echo ""

read -p "Enable Financial Services Staging? (y/N): " ENABLE_FS_STAGING
read -p "Enable Financial Services Production? (y/N): " ENABLE_FS_PROD
read -p "Enable Minnect Staging? (y/N): " ENABLE_MINNECT_STAGING
read -p "Enable Minnect Production? (y/N): " ENABLE_MINNECT_PROD
read -p "Enable Interview AI (Supabase)? (y/N): " ENABLE_IAI

echo ""

# Start building config
connections=()

# Financial Services Staging
if [[ "$ENABLE_FS_STAGING" =~ ^[Yy]$ ]]; then
    connections+='{
      "id": "ts_stage",
      "type": "postgres",
      "host": "localhost",
      "port": 4001,
      "name": "transaction_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "display_name": "Transaction Service Staging",
      "project": "transaction-service",
      "environment": "staging",
      "description": "Staging database for transaction service",
      "tags": ["staging", "transactions"]
    }'
    connections+='{
      "id": "ws_stage",
      "type": "postgres",
      "host": "localhost",
      "port": 4002,
      "name": "wallet_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "display_name": "Wallet Service Staging",
      "project": "wallet-service",
      "environment": "staging",
      "description": "Staging database for wallet service",
      "tags": ["staging", "wallet"]
    }'
    connections+='{
      "id": "pg_stage",
      "type": "postgres",
      "host": "localhost",
      "port": 4003,
      "name": "payment_gateway",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "display_name": "Payment Gateway Staging",
      "project": "payment-gateway",
      "environment": "staging",
      "description": "Staging database for payment gateway",
      "tags": ["staging", "payment"]
    }'
    connections+='{
      "id": "ls_stage",
      "type": "postgres",
      "host": "localhost",
      "port": 4004,
      "name": "ledger_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "display_name": "Ledger Service Staging",
      "project": "ledger-service",
      "environment": "staging",
      "description": "Staging database for ledger service",
      "tags": ["staging", "ledger"]
    }'
fi

# Financial Services Production
if [[ "$ENABLE_FS_PROD" =~ ^[Yy]$ ]]; then
    connections+='{
      "id": "ts_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 5001,
      "name": "transaction_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "options": {
        "default_transaction_read_only": "on"
      },
      "display_name": "Transaction Service Production",
      "project": "transaction-service",
      "environment": "production",
      "description": "Production database (read-only)",
      "tags": ["production", "critical", "transactions"]
    }'
    connections+='{
      "id": "ws_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 5002,
      "name": "wallet_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "options": {
        "default_transaction_read_only": "on"
      },
      "display_name": "Wallet Service Production",
      "project": "wallet-service",
      "environment": "production",
      "description": "Production database (read-only)",
      "tags": ["production", "wallet"]
    }'
    connections+='{
      "id": "pg_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 5003,
      "name": "payment_gateway",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "options": {
        "default_transaction_read_only": "on"
      },
      "display_name": "Payment Gateway Production",
      "project": "payment-gateway",
      "environment": "production",
      "description": "Production database (read-only)",
      "tags": ["production", "payment"]
    }'
    connections+='{
      "id": "ls_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 5004,
      "name": "ledger_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "options": {
        "default_transaction_read_only": "on"
      },
      "display_name": "Ledger Service Production",
      "project": "ledger-service",
      "environment": "production",
      "description": "Production database (read-only)",
      "tags": ["production", "ledger"]
    }'
fi

# Minnect Staging
if [[ "$ENABLE_MINNECT_STAGING" =~ ^[Yy]$ ]]; then
    connections+='{
      "id": "ps_stage",
      "type": "postgres",
      "host": "localhost",
      "port": 6001,
      "name": "profile_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "display_name": "Profile Service Staging",
      "project": "profile-service",
      "environment": "staging",
      "description": "Staging database for profile service",
      "tags": ["staging", "profiles"]
    }'
fi

# Minnect Production
if [[ "$ENABLE_MINNECT_PROD" =~ ^[Yy]$ ]]; then
    read -sp "Enter Minnect production database password: " MINNECT_PASSWORD
    echo ""
    
    connections+='{
      "id": "ps_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 7001,
      "name": "profile_service",
      "user": "ro_user",
      "password": "'"$DB_PASSWORD"'",
      "options": {
        "default_transaction_read_only": "on"
      },
      "display_name": "Profile Service Production",
      "project": "profile-service",
      "environment": "production",
      "description": "Production database for profile service",
      "tags": ["production", "profiles"]
    }'
    connections+='{
      "id": "minnect_prod",
      "type": "postgres",
      "host": "localhost",
      "port": 7002,
      "name": "valuetainment",
      "user": "valuetainment_backend",
      "password": "'"$MINNECT_PASSWORD"'",
      "options": {
        "default_transaction_read_only": "on"
      },
      "display_name": "Minnect Production",
      "project": "minnect",
      "environment": "production",
      "description": "Production database for Minnect",
      "tags": ["production", "minnect", "critical"]
    }'
fi

# Interview AI (Supabase)
if [[ "$ENABLE_IAI" =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${YELLOW}Interview AI uses Supabase (no SSH tunnel needed)${NC}"
    read -p "Staging Supabase user (e.g., ro_user.uvpghgakjckdoluqnqga): " IAI_STAGE_USER
    read -p "Production Supabase user (e.g., ro_user.gypnutyegqxelvsqjedu): " IAI_PROD_USER
    read -sp "Supabase password: " IAI_PASSWORD
    echo ""
    
    if [ -n "$IAI_STAGE_USER" ]; then
        connections+='{
          "id": "iai_stage",
          "type": "postgres",
          "host": "aws-0-us-east-1.pooler.supabase.com",
          "port": 6543,
          "name": "postgres",
          "user": "'"$IAI_STAGE_USER"'",
          "password": "'"$IAI_PASSWORD"'",
          "display_name": "Interview AI Staging",
          "project": "interview-ai",
          "environment": "staging",
          "description": "Staging database for Interview AI",
          "tags": ["staging", "interview-ai"]
        }'
    fi
    
    if [ -n "$IAI_PROD_USER" ]; then
        connections+='{
          "id": "iai_prod",
          "type": "postgres",
          "host": "aws-0-us-east-1.pooler.supabase.com",
          "port": 6543,
          "name": "postgres",
          "user": "'"$IAI_PROD_USER"'",
          "password": "'"$IAI_PASSWORD"'",
          "options": {
            "default_transaction_read_only": "on"
          },
          "display_name": "Interview AI Production",
          "project": "interview-ai",
          "environment": "production",
          "description": "Production database for Interview AI",
          "tags": ["production", "interview-ai"]
        }'
    fi
fi

# Build final JSON
echo ""
echo -e "${BLUE}Generating configuration...${NC}"

# Join connections with commas
IFS=','
connections_json="${connections[*]}"
IFS=' '

mkdir -p "$(dirname "$OUTPUT_FILE")"

cat > "$OUTPUT_FILE" << EOF
{
  "listen_address": "0.0.0.0",
  "port": 9092,
  "connections": [
    $connections_json
  ],
  "aws_profiles": []
}
EOF

# Validate JSON
if command -v jq &> /dev/null; then
    if jq . "$OUTPUT_FILE" > /dev/null 2>&1; then
        # Pretty-print the JSON
        jq . "$OUTPUT_FILE" > "$OUTPUT_FILE.tmp" && mv "$OUTPUT_FILE.tmp" "$OUTPUT_FILE"
        echo -e "${GREEN}✓ Configuration generated and validated${NC}"
    else
        echo -e "${RED}⚠ JSON validation failed. Please check the file manually.${NC}"
    fi
else
    echo -e "${YELLOW}Note: Install 'jq' for JSON validation${NC}"
fi

echo ""
echo -e "${GREEN}Configuration saved to: $OUTPUT_FILE${NC}"
echo ""
echo -e "${BOLD}Remember to:${NC}"
echo "  1. Copy SSH keys to $INSTALL_DIR/keys/"
echo "  2. Start tunnels: mcp-tunnels start"
echo "  3. Restart Cursor IDE"
