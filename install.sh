#!/bin/bash
set -euo pipefail

REPO="Szotasz/connectors-cli"
# Default to the latest version known to this installer. Override with
# `VERSION=v0.3.0 ./install.sh` to install a different release.
DEFAULT_VERSION="v0.2.0"
VERSION="${VERSION:-$DEFAULT_VERSION}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

BINARY="connectors-${OS}-${ARCH}"
BASE_URL="https://github.com/${REPO}/releases/download/${VERSION}"
BIN_URL="${BASE_URL}/${BINARY}"
SUMS_URL="${BASE_URL}/checksums.txt"

# Per-invocation temp dir so we don't race anyone else on /tmp/connectors.
TMPDIR_INSTALL=$(mktemp -d)
trap 'rm -rf "$TMPDIR_INSTALL"' EXIT

# Strict curl: fail on HTTP errors (-f), HTTPS only, TLS 1.2+, fail on cert
# issues. -fsSL keeps stdout clean but still surfaces errors via exit code.
CURL_OPTS=(-fsSL --proto '=https' --tlsv1.2 --retry 2)

echo "Downloading connectors ${VERSION} for ${OS}/${ARCH}..."
if ! curl "${CURL_OPTS[@]}" "$BIN_URL" -o "$TMPDIR_INSTALL/$BINARY"; then
  echo "Error: failed to download $BIN_URL" >&2
  echo "Check that the release exists: https://github.com/${REPO}/releases/tag/${VERSION}" >&2
  exit 1
fi

echo "Verifying checksum..."
if ! curl "${CURL_OPTS[@]}" "$SUMS_URL" -o "$TMPDIR_INSTALL/checksums.txt"; then
  echo "Error: failed to download $SUMS_URL" >&2
  echo "Refusing to install an unverified binary." >&2
  exit 1
fi

# Pick the sha256/shasum tool the host actually has.
if command -v sha256sum >/dev/null 2>&1; then
  SHA_CMD="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA_CMD="shasum -a 256"
else
  echo "Error: neither sha256sum nor shasum found; cannot verify download." >&2
  exit 1
fi

# Extract the expected hash for our binary from the manifest. The manifest
# uses the standard "<hash>  <filename>" format. We intentionally do not
# `cd` into TMPDIR_INSTALL and pipe to `sha256sum -c` because filenames in
# the manifest may not match our locally chosen paths.
expected=$(awk -v f="$BINARY" '$2 == f || $2 == "*"f { print $1; exit }' "$TMPDIR_INSTALL/checksums.txt")
if [ -z "$expected" ]; then
  echo "Error: $BINARY not found in checksums.txt" >&2
  exit 1
fi

actual=$($SHA_CMD "$TMPDIR_INSTALL/$BINARY" | awk '{print $1}')

if [ "$expected" != "$actual" ]; then
  echo "Error: checksum mismatch for $BINARY" >&2
  echo "  expected: $expected" >&2
  echo "  actual:   $actual" >&2
  exit 1
fi

chmod +x "$TMPDIR_INSTALL/$BINARY"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMPDIR_INSTALL/$BINARY" "$INSTALL_DIR/connectors"
else
  sudo mv "$TMPDIR_INSTALL/$BINARY" "$INSTALL_DIR/connectors"
fi

echo "Installed connectors ${VERSION} to ${INSTALL_DIR}/connectors"
echo ""
echo "Next steps:"
echo "  export CONNECTORS_HU_TOKEN=cnk_your_api_key"
echo "  connectors sync"
