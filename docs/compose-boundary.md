# Compose Boundary

`compose` exists to turn an already-existing image into a publishable card-like asset. It is not a free-form design system and not a command-line Figma.

## What `compose` is

- A fixed-layout card renderer for creator and brand workflows
- A deterministic post-processing step for image-based covers, promo cards, and social graphics
- A CLI surface optimized for automation, not manual exploration

## What `compose` is not

- Not a scene graph engine
- Not an arbitrary canvas editor
- Not a drag-and-drop replacement
- Not a multi-artboard design tool
- Not a generic template marketplace runtime

## Hard boundaries

V1/V2 `compose` must stay within these limits:

- One primary foreground image
- One optional background image or a solid background color
- One title
- One subtitle
- One logo
- One badge
- One fixed safe-area model
- One fixed layout family per invocation

No support for:

- arbitrary `x/y` positioning
- arbitrary rotation
- arbitrary z-index control
- multiple text boxes beyond title/subtitle/badge
- unlimited overlays
- free-form layer stacks
- gradients, masks, blend modes, or custom vector scenes in the first version

## Layout philosophy

`compose` should expose a small number of layout families, not an open-ended layout grammar.

Examples:

- `poster`
- `cover`
- `quote-card`
- `product-card`

Each layout family defines fixed element slots and predictable placement rules. Callers choose the layout family and fill the available slots.

## Why this boundary exists

- Keeps the CLI contract stable and understandable
- Prevents flag explosion
- Preserves deterministic rendering for agents and servers
- Avoids turning a rendering engine into a half-designed authoring tool
- Forces visual consistency across batches and brand outputs

## Decision rule for future requests

If a requested feature makes `compose` more like a free-form editor than a fixed-layout renderer, it should be rejected, deferred, or moved to a future higher-level system outside `imgcli` core.
