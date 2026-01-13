# Distribution Package Builder

This directory contains scripts to build and package the Infrastructure MCP Server for distribution to team members.

## Quick Build

```bash
# From project root
make package VERSION=1.0.1

# Or directly
./dist/package.sh
```

## What Gets Built

The package script creates:

```
releases/infra-mcp-server-VERSION.zip
releases/infra-mcp-server-VERSION.tar.gz
```

Each package contains:

| File | Description |
|------|-------------|
| `install.sh` | Cross-platform installer (macOS/Linux) |
| `SETUP.md` | User documentation |
| `bin/server` | Wrapper script that detects platform |
| `bin/server-darwin-arm64` | macOS Apple Silicon binary |
| `bin/server-darwin-amd64` | macOS Intel binary |
| `bin/server-linux-amd64` | Linux binary |
| `scripts/onboarding-test.sh` | Post-install verification |
| `scripts/health-check.sh` | Diagnostic tool |
| `scripts/generate-config.sh` | Interactive config wizard |
| `templates/config.json.example` | Example configuration |

## Directory Structure

```
dist/
├── README.md              # This file
├── SETUP.md               # User-facing documentation (included in package)
├── install.sh             # Main installer script
├── package.sh             # Builds the distribution package
├── scripts/
│   ├── generate-config.sh # Interactive config generator
│   ├── health-check.sh    # Quick health check
│   └── onboarding-test.sh # Full verification suite
├── templates/             # Config templates (auto-generated)
└── launchd/               # macOS launchd templates (if any)

releases/                  # OUTPUT directory (gitignored)
├── infra-mcp-server-1.0.1.tar.gz
└── infra-mcp-server-1.0.1.zip
```

## Distributing to Team Members

1. **Build the package:**
   ```bash
   make package VERSION=1.0.1
   ```

2. **Share the zip file** via:
   - Google Drive / Dropbox
   - Slack (if under 1GB)
   - USB drive
   - Internal file server

3. **Share SSH keys separately** (securely!):
   - 1Password / LastPass
   - Encrypted email
   - In-person USB transfer

4. **Point users to `SETUP.md`** in the package

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.1 | 2026-01-13 | Fixed binary copying bug, added shell reload instructions |
| 1.0.0 | 2026-01-13 | Initial release |

## Customization

### Adding New SSH Tunnels

Edit `install.sh` and add new launchd plist entries in the `setup_tunnels_macos()` function.

### Changing Default Ports

Update both:
1. The launchd plist entries in `install.sh`
2. The port checks in `scripts/health-check.sh` and `scripts/onboarding-test.sh`

### Adding New Databases

Update the `scripts/generate-config.sh` wizard to include new database options.
