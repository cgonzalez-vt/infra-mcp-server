#!/bin/bash
#
# Package the Infrastructure MCP Server for distribution
#
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
VERSION="${VERSION:-$(date +%Y%m%d)}"
PACKAGE_NAME="infra-mcp-server-${VERSION}"
OUTPUT_DIR="${OUTPUT_DIR:-$PROJECT_ROOT/releases}"

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║           Infrastructure MCP Server Packager                     ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Version: $VERSION"
echo "Output: $OUTPUT_DIR"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Create temporary directory for packaging
TEMP_DIR=$(mktemp -d)
PACKAGE_DIR="$TEMP_DIR/$PACKAGE_NAME"
mkdir -p "$PACKAGE_DIR"

echo "Building binaries..."

cd "$PROJECT_ROOT"

# Build for macOS (Intel)
echo "  → Building for macOS (amd64)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o "$PACKAGE_DIR/bin/server-darwin-amd64" ./cmd/server/main.go

# Build for macOS (Apple Silicon)
echo "  → Building for macOS (arm64)..."
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o "$PACKAGE_DIR/bin/server-darwin-arm64" ./cmd/server/main.go

# Build for Linux
echo "  → Building for Linux (amd64)..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$PACKAGE_DIR/bin/server-linux-amd64" ./cmd/server/main.go

# Create a wrapper script that selects the right binary
cat > "$PACKAGE_DIR/bin/server" << 'WRAPPEREOF'
#!/bin/bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

case "$(uname -s)-$(uname -m)" in
    Darwin-arm64) exec "$SCRIPT_DIR/server-darwin-arm64" "$@" ;;
    Darwin-x86_64) exec "$SCRIPT_DIR/server-darwin-amd64" "$@" ;;
    Linux-x86_64) exec "$SCRIPT_DIR/server-linux-amd64" "$@" ;;
    *)
        echo "Unsupported platform: $(uname -s)-$(uname -m)"
        exit 1
        ;;
esac
WRAPPEREOF
chmod +x "$PACKAGE_DIR/bin/server"

echo "Copying distribution files..."

# Copy dist files
cp "$SCRIPT_DIR/install.sh" "$PACKAGE_DIR/"
cp "$SCRIPT_DIR/SETUP.md" "$PACKAGE_DIR/"
mkdir -p "$PACKAGE_DIR/scripts"
cp "$SCRIPT_DIR/scripts/"*.sh "$PACKAGE_DIR/scripts/" 2>/dev/null || true

# Copy config templates
mkdir -p "$PACKAGE_DIR/templates"
cat > "$PACKAGE_DIR/templates/config.json.example" << 'CONFIGEOF'
{
  "listen_address": "0.0.0.0",
  "port": 9092,
  "connections": [
    {
      "id": "example_db",
      "type": "postgres",
      "host": "localhost",
      "port": 5432,
      "name": "mydb",
      "user": "myuser",
      "password": "mypassword",
      "display_name": "Example Database",
      "project": "my-project",
      "environment": "local",
      "description": "Example database connection",
      "tags": ["local", "example"]
    }
  ],
  "aws_profiles": []
}
CONFIGEOF

# Copy logo/assets
mkdir -p "$PACKAGE_DIR/assets"
cp "$PROJECT_ROOT/assets/logo.svg" "$PACKAGE_DIR/assets/" 2>/dev/null || true

echo "Creating archives..."

cd "$TEMP_DIR"

# Create tar.gz
tar -czvf "$OUTPUT_DIR/${PACKAGE_NAME}.tar.gz" "$PACKAGE_NAME"

# Create zip
zip -r "$OUTPUT_DIR/${PACKAGE_NAME}.zip" "$PACKAGE_NAME"

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                    Packaging Complete! 📦                        ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""
echo "Created packages:"
echo "  • $OUTPUT_DIR/${PACKAGE_NAME}.tar.gz"
echo "  • $OUTPUT_DIR/${PACKAGE_NAME}.zip"
echo ""
echo "To distribute:"
echo "  1. Share the package file with your team"
echo "  2. Include the SSH keys separately (securely!)"
echo "  3. Point them to SETUP.md for instructions"
