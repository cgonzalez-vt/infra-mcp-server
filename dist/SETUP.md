# 🚀 Infrastructure MCP Server - Quick Setup Guide

This guide will help you set up the Infrastructure MCP Server with Cursor IDE on **macOS** or **Linux**.

## What You'll Get

After setup, you'll have:
- **AI-powered database access** in Cursor IDE
- **SSH tunnels** that auto-connect to staging/production databases
- **Read-only access** to all service databases
- **Query execution** via natural language through Claude

## Prerequisites

| Requirement | Description |
|-------------|-------------|
| **Cursor IDE** | Latest version with MCP support |
| **SSH Keys** | Bastion host PEM files (get from team lead) |
| **Homebrew** (macOS) | Package manager for dependencies |

## Quick Install (5 minutes)

### Step 1: Run the Installer

```bash
# Clone or download the distribution package
cd ~/Downloads/infra-mcp-server

# Run the installer
./dist/install.sh
```

The installer will:
- Install `autossh` for persistent tunnels
- Create configuration directories
- Set up Cursor MCP integration
- Create tunnel management scripts

### Step 2: Reload Your Shell

**Important!** The installer adds commands to your PATH, but you need to reload your shell first:

**Option A:** Open a **new terminal window** (easiest)

**Option B:** Run this in your current terminal:
```bash
# For zsh (default on macOS)
source ~/.zshrc

# For bash
source ~/.bashrc
```

### Step 3: Copy SSH Keys

Get the SSH keys from your team lead and copy them:

```bash
# Copy keys to the keys directory
cp ~/Downloads/financial-services-prod-bastion.pem ~/.infra-mcp/keys/
cp ~/Downloads/financial-services-staging-bastion.pem ~/.infra-mcp/keys/
cp ~/Downloads/minnect-prod-bastion.pem ~/.infra-mcp/keys/
cp ~/Downloads/minnect-staging-bastion.pem ~/.infra-mcp/keys/

# Set correct permissions (required!)
chmod 600 ~/.infra-mcp/keys/*.pem
```

### Step 4: Configure Database Access

**Option A: Use the Config Generator (Recommended)**

```bash
~/.infra-mcp/bin/generate-config.sh
```

This will interactively ask which databases you want to access.

**Option B: Manual Configuration**

Edit `~/.infra-mcp/config/config.json` and update the passwords. Ask your team lead for the read-only database password.

### Step 5: Start SSH Tunnels

```bash
# Start all tunnels
mcp-tunnels start

# Check status
mcp-tunnels status
```

You should see output like:
```
Tunnel Status (macOS):
========================
  ✅ fs-prod (PID: 12345)
  ✅ fs-staging (PID: 12346)
  ✅ minnect-prod (PID: 12347)
  ✅ minnect-staging (PID: 12348)

Port Status:
------------
  ✅ :4001 - Transaction Service (staging)
  ✅ :4002 - Wallet Service (staging)
  ✅ :5001 - Transaction Service (prod)
  ...
```

### Step 6: Restart Cursor

1. **Quit Cursor completely** (Cmd+Q on macOS)
2. **Reopen Cursor**
3. **Check MCP status** in the bottom bar - you should see "infra-mcp-server" connected

## Using the MCP Server

Once connected, you can ask Claude in Cursor:

```
"Show me the schema for the users table in transaction service staging"

"Query the last 10 failed transactions from production"

"What indexes exist on the payments table?"

"Compare the user count between staging and production"
```

## Tunnel Management Commands

```bash
# Start all tunnels
mcp-tunnels start

# Start specific tunnel
mcp-tunnels start fs-prod

# Stop all tunnels
mcp-tunnels stop

# Check status
mcp-tunnels status

# View logs
mcp-tunnels logs fs-prod
```

## Troubleshooting

### MCP Server Not Connecting

1. Check Cursor's MCP config exists:
   ```bash
   cat ~/.cursor/mcp.json
   ```

2. Verify the binary exists and is executable:
   ```bash
   ls -la ~/.infra-mcp/bin/infra-mcp-server
   ```

3. Check MCP server logs:
   ```bash
   ls ~/.infra-mcp/logs/
   tail -f ~/.infra-mcp/logs/mcp-server-*.log
   ```

### Tunnels Not Starting

1. Check if keys have correct permissions:
   ```bash
   ls -la ~/.infra-mcp/keys/
   # Should show -rw------- (600)
   ```

2. Test SSH connection manually:
   ```bash
   ssh -i ~/.infra-mcp/keys/financial-services-staging-bastion.pem \
       ec2-user@44.193.202.29 echo "Connection successful"
   ```

3. Check tunnel logs:
   ```bash
   mcp-tunnels logs fs-staging
   ```

### Database Connection Errors

1. Verify tunnel is running:
   ```bash
   mcp-tunnels status
   ```

2. Test database connectivity:
   ```bash
   # Test staging transaction service
   psql -h localhost -p 4001 -U ro_user -d transaction_service -c "SELECT 1"
   ```

3. Check config.json has correct credentials:
   ```bash
   cat ~/.infra-mcp/config/config.json | grep password
   ```

## Directory Structure

After installation:

```
~/.infra-mcp/
├── bin/
│   ├── infra-mcp-server    # The MCP server binary
│   ├── mcp-tunnels         # Tunnel management script
│   └── mcp-start           # Server start script
├── config/
│   └── config.json         # Database connection config
├── keys/
│   ├── financial-services-prod-bastion.pem
│   ├── financial-services-staging-bastion.pem
│   ├── minnect-prod-bastion.pem
│   └── minnect-staging-bastion.pem
└── logs/
    ├── mcp-server-*.log
    └── tunnel-*.log

~/Library/LaunchAgents/  (macOS only)
├── com.infra-mcp.tunnel.fs-prod.plist
├── com.infra-mcp.tunnel.fs-staging.plist
├── com.infra-mcp.tunnel.minnect-prod.plist
└── com.infra-mcp.tunnel.minnect-staging.plist

~/.cursor/
└── mcp.json               # Cursor MCP configuration
```

## Port Reference

| Port | Service | Environment |
|------|---------|-------------|
| 4001 | Transaction Service | Staging |
| 4002 | Wallet Service | Staging |
| 4003 | Payment Gateway | Staging |
| 4004 | Ledger Service | Staging |
| 5001 | Transaction Service | Production |
| 5002 | Wallet Service | Production |
| 5003 | Payment Gateway | Production |
| 5004 | Ledger Service | Production |
| 6001 | Profile Service | Staging |
| 7001 | Profile Service | Production |
| 7002 | Minnect DB | Production |

## Security Notes

⚠️ **Important Security Considerations:**

1. **Never commit** SSH keys or config.json to git
2. **Production databases** are configured as **read-only** for safety
3. **Keep keys secure** - permissions should be 600
4. **Tunnels bind to 0.0.0.0** - only use on trusted networks

## Getting Help

- Check the logs: `~/.infra-mcp/logs/`
- Ask in the #eng-infrastructure Slack channel
- Contact your team lead for SSH keys or credentials

## Updating

To update to a new version:

```bash
# Download new distribution
cd ~/Downloads/infra-mcp-server-new

# Re-run installer (preserves your config)
./dist/install.sh
```

Your configuration and keys will be preserved.
