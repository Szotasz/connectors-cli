#!/bin/bash
set -euo pipefail

REPO="Szotasz/conn-cli"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

BINARY="conn-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${BINARY}"

echo "Downloading conn for ${OS}/${ARCH}..."
curl -sL "$URL" -o /tmp/conn
chmod +x /tmp/conn

if [ -w "$INSTALL_DIR" ]; then
  mv /tmp/conn "$INSTALL_DIR/conn"
else
  sudo mv /tmp/conn "$INSTALL_DIR/conn"
fi

echo "Installed conn to ${INSTALL_DIR}/conn"
echo ""
echo "Next steps:"
echo "  export CONN_HU_TOKEN=cnk_your_api_key"
echo "  conn sync"
