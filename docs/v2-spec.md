# imgcli V2 Specification

## Goal

V2 upgrades `imgcli` from a small set of image utilities into a composable image-agent CLI. The target is not “more commands” by itself. The target is a stable automation surface that lets an agent:

1. inspect visual assets before acting
2. apply a predictable sequence of transforms
3. export multiple platform variants from one source
4. do all of the above with stable JSON contracts

Project boundary:

- `imgcli` does not generate images.
- Upstream generation belongs to external CLIs or APIs such as Gemini or GPT Image.
- `imgcli` starts from already-existing image assets and turns them into inspected, normalized, composed, packaged, and exported outputs.

V2 implementation status:

- implemented: `inspect`, `compose`, `convert`, `variants`, `run`

## Non-goals

- No model-dependent image editing in V2 core.
- No text-to-image or image generation provider integration in V2 core.
- No background removal, face detection, or “AI beautify” in V2.
- No plugin runtime yet.
- No breaking change to existing `smartpad`, `topdf`, or `stitch` JSON contracts.

## Shared CLI rules

- Default output remains JSON.
- Success writes to `stdout`.
- Failure writes structured JSON to `stderr` and exits non-zero.
- Input order must be deterministic.
- All V2 commands accept `--json` for explicit compatibility, even though JSON is default.
- All V2 commands must support `--output -` only after binary `stdout` mode is intentionally designed. V2 keeps file-path-first semantics.

Shared success envelope:

```json
{
  "ok": true,
  "command": "inspect",
  "data": {}
}
```

Shared error envelope:

```json
{
  "error": "reason",
  "code": "INVALID_ARGUMENT",
  "exit_code": 2
}
```

New shared error codes introduced in V2:

- `CONFIG_ERROR`
- `PLAN_INVALID`
- `FONT_LOAD_FAILED`
- `METADATA_EXTRACT_FAILED`
- `OUTPUT_CONFLICT`
- `SIZE_LIMIT_EXCEEDED`

## Command: `inspect`

### Purpose

Give an agent a reliable machine-readable summary of one or more images before making processing decisions.

### Example

```bash
imgcli inspect --input poster.png
imgcli inspect --input a.jpg --input b.png
imgcli inspect --input-dir ./assets
```

### Flags

- `--input <path>` repeatable
- `--input-dir <dir>`
- `--hash` include perceptual hash and SHA256
- `--color-stats` include dominant color and average color
- `--limit <n>` limit files when using `--input-dir`

### Output

Per image:

- path
- format
- width
- height
- orientation
- file size
- has alpha
- EXIF orientation if present
- average color
- dominant color
- optional pHash
- optional SHA256

Example:

```json
{
  "ok": true,
  "command": "inspect",
  "data": {
    "images": [
      {
        "path": "poster.png",
        "format": "png",
        "width": 1536,
        "height": 2752,
        "orientation": "portrait",
        "size_bytes": 5512312,
        "has_alpha": false,
        "dominant_color": "#d8c1a6",
        "average_color": "#cbb89b"
      }
    ]
  }
}
```

### Acceptance notes

- `inspect` must not decode more image data than needed for metadata and low-resolution color sampling.
- `input-dir` ordering must match the natural-sort rules already used by `topdf`.

## Command: `compose`

### Purpose

Create creator-facing card layouts from an existing image asset. `compose` is a fixed-layout renderer, not a free-form canvas engine.

### Example

```bash
imgcli compose \
  --input cover.jpg \
  --output card.jpg \
  --width 1080 \
  --height 1440 \
  --title "Spring Drop" \
  --subtitle "Limited edition" \
  --logo brand.png
```

### V2 scope

Supported building blocks:

- background color or background image
- one fitted foreground image
- title and subtitle text
- one badge in a fixed slot
- optional logo overlay
- padding and safe area
- rounded corners

Explicit boundaries:

- no arbitrary `x/y` placement
- no arbitrary rotation or z-index
- no multiple independent text boxes beyond title/subtitle/badge
- no unlimited overlays
- no scene graph, timeline, or template language in V2
- no goal of becoming a command-line Figma

### Flags

- `--input <path>`
- `--output <path>`
- `--width <px>`
- `--height <px>`
- `--background-color <hex>`
- `--background-image <path>`
- `--title <text>`
- `--subtitle <text>`
- `--title-size <px>`
- `--subtitle-size <px>`
- `--title-color <hex>`
- `--subtitle-color <hex>`
- `--font <path>` optional; if absent use embedded default
- `--logo <path>`
- `--badge <text>`
- `--padding <px>`
- `--radius <px>`
- `--safe-area <top,right,bottom,left>`
- `--layout poster|cover|quote-card|product-card`
- `--quality <1-100>`

Flags intentionally not supported:

- no `--x`, `--y`
- no `--rotate`
- no `--layer`
- no repeatable arbitrary element flags

### Output

- output path
- final canvas size
- chosen layout family
- elements rendered
- font source
- duration

### Acceptance notes

- V2 compose is intentionally template-light: enough structure for repeatable brand cards without inventing a full scene graph.
- V2 compose must be implemented as a small set of fixed layout families with named slots.
- If a future request requires arbitrary layout freedom, it should not be added to `compose` by default.
- Text overflow must fail deterministically or truncate according to an explicit rule. V2 default: truncate with ellipsis for single-line badge and wrap title/subtitle to configured text box height.

## Command: `variants`

### Purpose

Turn one input into multiple platform-specific outputs in one call.

### Example

```bash
imgcli variants \
  --input hero.jpg \
  --preset-set creator-basic \
  --output-dir ./dist
```

### Built-in preset sets

- `creator-basic`
  - `xiaohongshu`
  - `wechat_cover`
  - `square`
  - `story_9x16`
- `ecommerce-basic`
  - `product_square`
  - `detail_long`
  - `banner_16x9`

### Flags

- `--input <path>`
- `--output-dir <dir>`
- `--preset-set <name>`
- `--preset <name>` repeatable, alternative to preset-set
- `--background blur|solid`
- `--filename-template <template>`

### Output

- input path
- output dir
- generated files
- per file preset, size, path

### Acceptance notes

- `variants` reuses `smartpad` internals instead of reimplementing resize logic.
- naming must be deterministic. V2 default filename template: `{base}_{preset}{ext}`

## Command: `run`

### Purpose

Execute a declarative multi-step image workflow from JSON or YAML.

### Example

```bash
imgcli run --recipe recipe.json
imgcli run --recipe launch.yaml --dry-run
```

### Recipe model

Top-level fields:

- `version`
- `inputs`
- `steps`
- `outputs`

Supported step types in V2:

- `inspect`
- `smartpad`
- `compose`
- `convert`
- `variants`
- `stitch`
- `topdf`

### `--dry-run`

Returns the resolved execution plan without writing outputs.

Example:

```json
{
  "ok": true,
  "command": "run",
  "data": {
    "dry_run": true,
    "steps": [
      {
        "id": "hero_xhs",
        "type": "smartpad",
        "input": "hero.jpg",
        "output": "dist/hero_xiaohongshu.jpg"
      }
    ]
  }
}
```

### Acceptance notes

- `run` is the orchestration layer. It must call internal services, not shell out to subcommands.
- V2 only needs local-file recipes.
- Recipe schema validation must fail before any output is written.

## Command: `convert`

### Purpose

Normalize format, quality, compression, metadata, and color delivery.

### Example

```bash
imgcli convert \
  --input source.png \
  --output result.jpg \
  --quality 82 \
  --strip-metadata
```

### Flags

- `--input <path>`
- `--output <path>`
- `--quality <1-100>`
- `--strip-metadata`
- `--flatten-background <hex>` for alpha to JPEG workflows
- `--max-width <px>`
- `--max-height <px>`

### Acceptance notes

- V2 convert is deliberately not a full optimizer. It focuses on the highest-value delivery controls.
- `convert` becomes the normalization primitive used by future upload pipelines.

## Internal architecture changes required

### New packages

- `pkg/inspect`
- `pkg/compose`
- `pkg/variants`
- `pkg/runbook` or `pkg/recipe`
- `pkg/convert`

### Refactors

- Extract shared input collection from `topdf` into a reusable package.
- Extract shared output metadata helpers from existing commands.
- Introduce a small internal execution-plan type used by `run --dry-run`.
- Move preset definitions toward a single source of truth so `smartpad` and `variants` do not diverge.

## Testing requirements

- Unit tests for every new command package.
- JSON schema shape tests for success and failure payloads.
- e2e tests for:
  - `inspect` on mixed PNG/JPEG inputs
  - `compose` basic card generation
  - `variants` creator-basic
  - `run --dry-run`
  - `convert` PNG to JPEG with flattening
- real smoke should remain focused on existing command chain until V2 commands are implemented, then extend to cover one recipe-driven flow.

## Recommended implementation order

1. `inspect`
2. shared input/output utility refactor
3. `convert`
4. `compose`
5. `variants`
6. `run --dry-run`
7. `run` execution mode

## Why this order

- `inspect` reduces blind decisions for every later command.
- `convert` and shared utilities create reusable primitives instead of one-off code paths.
- `compose` and `variants` directly address creator workflows.
- `run` should come after the underlying primitives are stable enough to orchestrate.
