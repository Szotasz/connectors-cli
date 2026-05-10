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

LATEST=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | cut -d'"' -f4)
if [ -z "$LATEST" ]; then
  echo "Could not determine latest version"
  exit 1
fi

BINARY="conn-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${BINARY}"

echo "Downloading conn ${LATEST} for ${OS}/${ARCH}..."
curl -sL "$URL" -o /tmp/conn
chmod +x /tmp/conn

if [ -w "$INSTALL_DIR" ]; then
  mv /tmp/conn "$INSTALL_DIR/conn"
else
  sudo mv /tmp/conn "$INSTALL_DIR/conn"
fi

echo "Installed conn ${LATEST} to ${INSTALL_DIR}/conn"
echo ""
echo "Next steps:"
echo "  export CONN_HU_TOKEN=cnk_your_api_key"
echo "  conn sync"
