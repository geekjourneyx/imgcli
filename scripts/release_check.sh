#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

version_go="$(sed -n 's/.*Version = "\(.*\)"/\1/p' "$ROOT_DIR/pkg/version/version.go")"
version_make="$(sed -n 's/^VERSION := \(.*\)$/\1/p' "$ROOT_DIR/Makefile")"
version_install="$(sed -n 's/^VERSION="\([^"]*\)"/\1/p' "$ROOT_DIR/scripts/install.sh")"
version_changelog="$(sed -n 's/^## \[\([^]]*\)\].*/\1/p' "$ROOT_DIR/CHANGELOG.md" | head -n1)"

if [[ -z "$version_go" || -z "$version_make" || -z "$version_install" || -z "$version_changelog" ]]; then
  echo "version check failed: missing version source"
  exit 1
fi

if [[ "$version_go" != "$version_make" || "$version_go" != "$version_install" || "$version_go" != "$version_changelog" ]]; then
  echo "version mismatch: go=$version_go make=$version_make install=$version_install changelog=$version_changelog"
  exit 1
fi

for required in \
  "$ROOT_DIR/.github/workflows/ci.yml" \
  "$ROOT_DIR/.github/workflows/release.yml" \
  "$ROOT_DIR/scripts/install.sh" \
  "$ROOT_DIR/skills/imgcli/SKILL.md"; do
  if [[ ! -f "$required" ]]; then
    echo "release-check failed: missing required file $required"
    exit 1
  fi
done

echo "release-check ok: $version_go"
