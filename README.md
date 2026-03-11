# imgcli

[![CI](https://github.com/geekjourneyx/imgcli/actions/workflows/ci.yml/badge.svg)](https://github.com/geekjourneyx/imgcli/actions/workflows/ci.yml)
[![Release](https://github.com/geekjourneyx/imgcli/actions/workflows/release.yml/badge.svg)](https://github.com/geekjourneyx/imgcli/actions/workflows/release.yml)
[![Go Version](https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go)](https://go.dev/)
[![CGO Free](https://img.shields.io/badge/cgo-free-2F855A)](https://go.dev/)

[简体中文](./README.zh-CN.md) | [English](./README.md)

`imgcli` is a CGO-free, agent-native image processing CLI for high-throughput automation. It is built for scripts, workers, and digital employees that need stable JSON output, deterministic file handling, and low-friction deployment across Linux and macOS.

## Project principles

- `imgcli` is a post-processing and packaging CLI, not an image generation CLI.
- Image generation should stay in external tools such as Gemini, GPT Image, or other provider-specific CLIs and APIs.
- `imgcli` is responsible for inspecting, normalizing, composing, packaging, and exporting images that already exist.
- The core must stay deterministic, local-file-first, and suitable for stable automation.

## Why imgcli

Typical image tooling breaks down in automation for predictable reasons:

- CLI output is optimized for humans instead of machines.
- Large image batches consume too much memory.
- Social-media resizing is handled with crude crops that damage composition.
- PDF and watermark workflows are often coupled to GUI tools or heavyweight runtimes.

`imgcli` is intentionally narrower:

- default JSON output for machine consumers
- stable error codes and non-zero exit codes
- CGO-free builds for easy cross-platform distribution
- memory-conscious, sequential processing for batch image jobs

## Core commands

- `inspect`
  - inspect image metadata before transforming assets
  - optionally include color statistics, SHA256, and perceptual hash
  - support single files or deterministic directory scans
- `compose`
  - render a fixed-layout creator card from one existing image
  - support title, subtitle, logo, badge, safe area, and rounded image corners
  - intentionally limited to named layout families such as `poster` and `cover`
- `convert`
  - normalize format, resize limits, and JPEG delivery settings
  - support background flattening for transparent-to-JPEG workflows
  - re-encode through a deterministic metadata-stripping path
- `run`
  - execute JSON or YAML recipes through internal services instead of shelling out to subcommands
  - support `input:<name>` and `step:<id>` references
  - support `--dry-run` to inspect the resolved plan before writing outputs
- `variants`
  - export multiple platform-specific variants from one source image
  - support built-in preset sets such as `creator-basic`
  - keep deterministic output naming and reuse `smartpad` internals
- `smartpad`
  - resize an image into a preset such as `xiaohongshu` or `wechat_cover`
  - preserve composition by fitting the foreground into a padded canvas
  - support blurred or solid-color background fill
- `topdf`
  - pack multiple images into a PDF in deterministic order
  - support visible text watermarking during page generation
  - stream pages one by one instead of retaining the full batch in memory
- `stitch`
  - vertically stitch multiple images to a fixed width
  - auto-split very tall outputs into multiple parts
  - return all output paths in JSON

## Installation

Recommended:

```bash
curl -fsSL https://raw.githubusercontent.com/geekjourneyx/imgcli/main/scripts/install.sh | bash
```

The install script:

- downloads the correct release asset for `linux` or `darwin`
- verifies the asset against `SHA256SUMS`
- installs to `~/.local/bin`
- prints a `PATH` hint if needed

Manual download:

- Releases: `https://github.com/geekjourneyx/imgcli/releases`
- Assets:
  - `imgcli-linux-amd64`
  - `imgcli-linux-arm64`
  - `imgcli-darwin-amd64`
  - `imgcli-darwin-arm64`
  - `SHA256SUMS`

Verify:

```bash
imgcli version
```

## Quick start

```bash
make build
./bin/imgcli version

./bin/imgcli inspect \
  --input in.jpg \
  --hash \
  --color-stats

./bin/imgcli compose \
  --input in.jpg \
  --output poster.jpg \
  --width 1080 \
  --height 1440 \
  --layout poster \
  --title "Launch Day" \
  --subtitle "A fixed-layout creator card"

./bin/imgcli convert \
  --input source.png \
  --output normalized.jpg \
  --flatten-background "#ffffff" \
  --max-width 1600 \
  --quality 82 \
  --strip-metadata

./bin/imgcli run \
  --recipe recipe.json \
  --dry-run

./bin/imgcli variants \
  --input poster.jpg \
  --output-dir dist \
  --preset-set creator-basic

./bin/imgcli smartpad \
  --input in.jpg \
  --output out.jpg \
  --preset xiaohongshu

./bin/imgcli topdf \
  --input page1.jpg \
  --input page2.jpg \
  --output bundle.pdf \
  --watermark-text "internal"

./bin/imgcli stitch \
  --input a.jpg \
  --input b.jpg \
  --output stitched.jpg \
  --width 1080
```

## JSON contract

`imgcli` defaults to JSON output. Success goes to `stdout`; failures return a non-zero exit code and structured JSON on `stderr`.

Success example:

```json
{
  "ok": true,
  "command": "smartpad",
  "data": {
    "input": "in.jpg",
    "output": "out.jpg",
    "target_width": 1080,
    "target_height": 1440
  }
}
```

Error example:

```json
{
  "error": "preset \"foo\" not found",
  "code": "PRESET_NOT_FOUND",
  "exit_code": 2
}
```

Contract rules:

- output fields are additive by default
- callers must not parse human help text
- error handling must key off `code`, not message strings

## Recipe references

`run` recipes stay file-path-first and only add two explicit reference forms:

- `input:<name>` resolves a top-level named input from the recipe
- `step:<id>` resolves file outputs produced by a previous step

Minimal example:

```json
{
  "version": "v1",
  "inputs": {
    "hero": "hero.jpg",
    "logo": "logo.png"
  },
  "steps": [
    {
      "id": "card",
      "type": "compose",
      "input": "input:hero",
      "output": "dist/card.jpg",
      "width": 1080,
      "height": 1440,
      "layout": "poster",
      "title": "Launch Day",
      "logo": "input:logo"
    },
    {
      "id": "web",
      "type": "convert",
      "input": "step:card",
      "output": "dist/card_web.jpg",
      "max_width": 720,
      "max_height": 720,
      "quality": 80,
      "strip_metadata": true
    }
  ]
}
```

## Presets

- `xiaohongshu`: `1080x1440`
- `wechat_cover`: `900x383`
- `square`: `1080x1080`
- `story_9x16`: `1080x1920`
- `product_square`: `1200x1200`
- `detail_long`: `1080x2160`
- `banner_16x9`: `1600x900`

Preset sets:

- `creator-basic`: `xiaohongshu`, `wechat_cover`, `square`, `story_9x16`
- `ecommerce-basic`: `product_square`, `detail_long`, `banner_16x9`

## Real smoke

If local `baoyu-image-gen` is configured, run the end-to-end smoke:

```bash
make real-smoke
```

This flow:

- builds `imgcli`
- uses an external image-generation skill to create fresh portrait, landscape, and square test images
- runs `inspect` on those generated images
- runs `compose`, `convert`, `variants`, `run`, `smartpad`, `topdf`, and `stitch`
- verifies output file types and prints the artifact directory

Useful overrides:

- `SKILL_DIR=/custom/skill/path make real-smoke`
- `SMOKE_ROOT=/tmp/custom-smokes make real-smoke`
- `RUN_ID=manual-check make real-smoke`

## Skill for agents

- Skill file: `skills/imgcli/SKILL.md`
- This repository also includes `AGENTS.md` for engineering and review policy

## V2 status

The next planned layer is documented in [docs/v2-spec.md](/root/go/src/imgcli/docs/v2-spec.md).
The `compose` boundary is documented in [docs/compose-boundary.md](/root/go/src/imgcli/docs/compose-boundary.md).

V2 core surface is now implemented: `inspect`, `compose`, `convert`, `variants`, and `run`.
`compose` remains intentionally scoped as a fixed-layout card renderer, not a command-line Figma.

## Development

Required quality gates:

```bash
gofmt -l .
go vet ./...
golangci-lint run
CGO_ENABLED=1 go test -count=1 ./...
make release-check
make build
```

Common commands:

```bash
make fmt
make vet
make lint
make test
make release-check
make build
make real-smoke
```
