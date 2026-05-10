#!/bin/bash
set -euo pipefail

REPO="Szotasz/connectors-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY="connectors-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

echo "Downloading connectors for ${OS}/${ARCH}..."
curl -sL "$URL" -o /tmp/connectors
chmod +x /tmp/connectors

if [ -w "$INSTALL_DIR" ]; then
  mv /tmp/connectors "$INSTALL_DIR/connectors"
else
  sudo mv /tmp/connectors "$INSTALL_DIR/connectors"
fi

echo "Installed connectors to ${INSTALL_DIR}/connectors"
echo ""
echo "Next steps:"
echo "  export CONNECTORS_HU_TOKEN=cnk_your_api_key"
echo "  connectors sync"
