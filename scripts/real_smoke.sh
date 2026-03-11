#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SKILL_DIR="${SKILL_DIR:-$HOME/.openclaw/skills/baoyu-image-gen}"
IMGCLI_BIN="${IMGCLI_BIN:-$ROOT_DIR/bin/imgcli}"
GOCACHE_DIR="${GOCACHE_DIR:-/tmp/go-build}"
SMOKE_ROOT="${SMOKE_ROOT:-/tmp/imgcli-real-smoke}"
RUN_ID="${RUN_ID:-$(date -u +%Y%m%dT%H%M%SZ)}"
RUN_DIR="${SMOKE_ROOT}/${RUN_ID}"
GEN_SCRIPT="${SKILL_DIR}/scripts/main.ts"

PORTRAIT_PROMPT="${PORTRAIT_PROMPT:-A premium skincare campaign poster, soft daylight, minimalist studio, elegant product arrangement, editorial quality}"
LANDSCAPE_PROMPT="${LANDSCAPE_PROMPT:-A modern tea brand homepage hero banner, cinematic light, premium packaging, calm composition}"
SQUARE_PROMPT="${SQUARE_PROMPT:-A colorful stationery flat lay for social media, clean background, crisp shadows, high detail}"

log() {
  printf '[real-smoke] %s\n' "$*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    printf 'missing required command: %s\n' "$1" >&2
    exit 1
  fi
}

check_skill() {
  if [[ ! -f "$GEN_SCRIPT" ]]; then
    printf 'image generation skill not found: %s\n' "$GEN_SCRIPT" >&2
    exit 1
  fi
  if [[ ! -f "$HOME/.baoyu-skills/baoyu-image-gen/EXTEND.md" && ! -f "$ROOT_DIR/.baoyu-skills/baoyu-image-gen/EXTEND.md" ]]; then
    printf 'missing baoyu-image-gen EXTEND.md; generation is blocked until preferences are configured\n' >&2
    exit 1
  fi
}

build_bin() {
  log "building imgcli"
  mkdir -p "$(dirname "$IMGCLI_BIN")"
  GOCACHE="$GOCACHE_DIR" CGO_ENABLED=0 go build -o "$IMGCLI_BIN" "$ROOT_DIR"
}

generate_image() {
  local prompt="$1"
  local output="$2"
  local ratio="$3"

  log "generating $(basename "$output") with aspect ratio ${ratio}"
  npx -y bun "$GEN_SCRIPT" \
    --prompt "$prompt" \
    --image "$output" \
    --ar "$ratio" \
    --json
}

run_cli() {
  log "running $*"
  "$IMGCLI_BIN" "$@"
}

main() {
  require_cmd go
  require_cmd npx
  require_cmd file
  check_skill

  mkdir -p "$RUN_DIR"

  local portrait="${RUN_DIR}/portrait.png"
  local landscape="${RUN_DIR}/landscape.png"
  local square="${RUN_DIR}/square.png"
  local compose_out="${RUN_DIR}/poster.jpg"
  local convert_out="${RUN_DIR}/poster_web.jpg"
  local variants_dir="${RUN_DIR}/variants"
  local recipe_path="${RUN_DIR}/run_recipe.json"
  local recipe_card="${RUN_DIR}/run_card.jpg"
  local recipe_web="${RUN_DIR}/run_web.jpg"
  local recipe_variants_dir="${RUN_DIR}/run_variants"
  local smartpad_out="${RUN_DIR}/portrait_xhs.jpg"
  local pdf_out="${RUN_DIR}/bundle.pdf"
  local stitch_out="${RUN_DIR}/stitched.jpg"

  build_bin

  generate_image "$PORTRAIT_PROMPT" "$portrait" "9:16"
  generate_image "$LANDSCAPE_PROMPT" "$landscape" "16:9"
  generate_image "$SQUARE_PROMPT" "$square" "1:1"

  run_cli inspect \
    --input "$portrait" \
    --input "$landscape" \
    --input "$square" \
    --hash \
    --color-stats

  run_cli compose \
    --input "$portrait" \
    --output "$compose_out" \
    --width 1080 \
    --height 1440 \
    --layout poster \
    --title "Image Agent Cover" \
    --subtitle "Composed from an existing asset with fixed slots" \
    --logo "$square" \
    --badge "V2"

  run_cli convert \
    --input "$compose_out" \
    --output "$convert_out" \
    --quality 78 \
    --max-width 720 \
    --max-height 720 \
    --strip-metadata

  run_cli variants \
    --input "$compose_out" \
    --output-dir "$variants_dir" \
    --preset-set creator-basic

  cat > "$recipe_path" <<EOF
{
  "version": "v1",
  "inputs": {
    "hero": "$portrait",
    "logo": "$square"
  },
  "steps": [
    {
      "id": "inspect-source",
      "type": "inspect",
      "inputs": ["input:hero"],
      "include_hash": true
    },
    {
      "id": "card",
      "type": "compose",
      "input": "input:hero",
      "output": "$recipe_card",
      "width": 1080,
      "height": 1440,
      "layout": "poster",
      "title": "Recipe Driven Cover",
      "subtitle": "Executed through imgcli run",
      "logo": "input:logo",
      "badge": "RUN"
    },
    {
      "id": "web",
      "type": "convert",
      "input": "step:card",
      "output": "$recipe_web",
      "quality": 78,
      "max_width": 720,
      "max_height": 720,
      "strip_metadata": true
    },
    {
      "id": "social",
      "type": "variants",
      "input": "step:web",
      "output_dir": "$recipe_variants_dir",
      "preset_set": "creator-basic"
    }
  ]
}
EOF

  run_cli run \
    --recipe "$recipe_path" \
    --dry-run

  run_cli run \
    --recipe "$recipe_path"

  run_cli smartpad \
    --input "$portrait" \
    --output "$smartpad_out" \
    --preset xiaohongshu

  run_cli topdf \
    --input "$portrait" \
    --input "$landscape" \
    --input "$square" \
    --output "$pdf_out" \
    --watermark-text "imgcli smoke" \
    --watermark-position tile

  run_cli stitch \
    --input "$portrait" \
    --input "$landscape" \
    --input "$square" \
    --output "$stitch_out" \
    --width 1080

  log "verifying output file types"
  file "$portrait" "$landscape" "$square" "$compose_out" "$convert_out" "$recipe_card" "$recipe_web" "$smartpad_out" "$pdf_out" "$stitch_out"
  file "$variants_dir"/*
  file "$recipe_variants_dir"/*

  log "artifacts saved in ${RUN_DIR}"
  ls -lh "$RUN_DIR"
  ls -lh "$variants_dir"
  ls -lh "$recipe_variants_dir"
}

main "$@"
