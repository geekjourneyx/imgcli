#!/usr/bin/env bash
set -euo pipefail

VERSION="1.0.0"
REPO="${REPO:-imgcli}"
OWNER="${OWNER:-geekjourneyx}"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "unsupported arch: $arch" >&2; exit 1 ;;
esac
case "$os" in
  linux|darwin) ;;
  *) echo "unsupported os: $os" >&2; exit 1 ;;
esac

asset="imgcli-${os}-${arch}"
base_url="https://github.com/${OWNER}/${REPO}/releases/download/v${VERSION}"
url="${base_url}/${asset}"
checksums_url="${base_url}/SHA256SUMS"

tmpdir="$(mktemp -d)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT

checksum_file() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | awk '{print $1}'
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
    return
  fi
  echo "missing checksum tool: need sha256sum or shasum" >&2
  exit 1
}

mkdir -p "$INSTALL_DIR"

echo "platform: ${os}/${arch}"
echo "download: $url"
curl -fsSL "$url" -o "$tmpdir/$asset"
curl -fsSL "$checksums_url" -o "$tmpdir/SHA256SUMS"

expected="$(grep "  ${asset}\$" "$tmpdir/SHA256SUMS" | awk '{print $1}')"
if [[ -z "$expected" ]]; then
  echo "checksum entry not found for $asset" >&2
  exit 1
fi
actual="$(checksum_file "$tmpdir/$asset")"
if [[ "$expected" != "$actual" ]]; then
  echo "checksum mismatch for $asset" >&2
  exit 1
fi

install -m 0755 "$tmpdir/$asset" "$INSTALL_DIR/imgcli"
echo "installed: $INSTALL_DIR/imgcli"
if [[ ":$PATH:" != *":${INSTALL_DIR}:"* ]]; then
  echo "add PATH: export PATH=\"\$PATH:${INSTALL_DIR}\""
fi
