# SQLTools Database Connections Setup

This guide explains how to set up SQLTools in VS Code / Cursor with our team's database connections.

## Prerequisites

- VS Code or Cursor IDE
- Access to our database servers (VPN if required)

## Installation

### Step 1: Install Extensions

Install the following extensions in VS Code / Cursor:

1. **SQLTools** - `mtxr.sqltools`
2. **SQLTools PostgreSQL Driver** - `mtxr.sqltools-driver-pg`

You can install via command line:

```bash
# For Cursor
cursor --install-extension mtxr.sqltools
cursor --install-extension mtxr.sqltools-driver-pg

# For VS Code
code --install-extension mtxr.sqltools
code --install-extension mtxr.sqltools-driver-pg
```

Or search for "SQLTools" in the Extensions panel and install both.

### Step 2: Add Connections

1. Open your VS Code / Cursor settings:
   - Press `Ctrl+Shift+P` (or `Cmd+Shift+P` on Mac)
   - Type "Preferences: Open Settings (JSON)"
   - Select it to open `settings.json`

2. Copy the contents of `sqltools-connections.json` and paste it into your `settings.json`

   If you already have content in `settings.json`, merge the `"sqltools.connections"` array into your existing settings.

   **Example:**
   ```json
   {
     "editor.fontSize": 14,
     "sqltools.connections": [
       // ... paste connections here
     ]
   }
   ```

3. Save the file

### Step 3: Reload

Reload the window to apply changes:
- Press `Ctrl+Shift+P` → "Developer: Reload Window"

## Usage

1. Click the **SQLTools icon** (database cylinder) in the left sidebar
2. You'll see connections organized by group:
   - **local** - Local development databases
   - **staging** - Staging environment databases
   - **production** - Production databases (read-only)
3. Click on a connection to connect
4. Right-click on tables to view data, run queries, etc.

## Connection Groups

| Group | Description | Use Case |
|-------|-------------|----------|
| `local` | Local Docker databases | Development & testing |
| `staging` | Staging environment | QA & integration testing |
| `production` | Production databases | Read-only queries, debugging |

## Available Databases

### Local
- Interview AI Local
- Performance AI Local
- Transaction Service Local
- Wallet Service Local
- Payment Gateway Local
- Ledger Service Local

### Staging
- Transaction Service Staging
- Wallet Service Staging
- Payment Gateway Staging
- Ledger Service Staging
- Profile Service Staging
- Interview AI Staging

### Production
- Transaction Service Production
- Wallet Service Production
- Payment Gateway Production
- Ledger Service Production
- Profile Service Production
- Interview AI Production

## Notes

- All connections use **read-only** database users
- Production connections should be used carefully - avoid running expensive queries
- Local connections require the respective services to be running locally
- Staging/Production connections may require VPN access

## Troubleshooting

### "Driver not installed"
Make sure you installed `mtxr.sqltools-driver-pg` and reloaded the window.

### Connection timeout
- Check if you're connected to VPN (for staging/production)
- Verify the service is running (for local)
- Check the port is correct and not blocked

### "Permission denied"
The read-only user may not have access to certain tables. Contact the DBA team if you need additional access.

## Updating Connections

If database configurations change, get the updated `sqltools-connections.json` file and replace the `sqltools.connections` array in your settings.

---

*Generated from `bin/config.json` using `scripts/generate-sqltools-config.js`*

