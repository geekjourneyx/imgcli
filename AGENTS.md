# AGENTS

## Engineering policy

- Keep command behavior backward-compatible unless a breaking change is explicitly approved.
- Preserve JSON schema compatibility with additive changes only.
- Keep the CLI CGO-free.
- Keep image generation out of `imgcli` core; generation belongs to external provider CLIs or APIs.
- Keep `compose` narrow: it is a fixed-layout card renderer, not a free-form design engine.
- Never commit real credentials, tokens, generated secrets, or private service endpoints.
- Prefer deterministic image fixtures in tests; use live generation only for explicit smoke validation.

## Required checks before merge

- `gofmt -l .` returns no output.
- `go vet ./...` passes.
- `golangci-lint run` passes.
- `CGO_ENABLED=1 go test -count=1 ./...` passes.
- `make release-check` passes.
- `make build` passes.

## When to run real smoke

- Run `make real-smoke` when changing `smartpad`, `topdf`, `stitch`, image encoding, release packaging, or the smoke script itself.
- If real smoke cannot be run, state the blocker and keep the last known passing command/result in the delivery note.

## Version and release rules

- Keep `Makefile`, `pkg/version/version.go`, `scripts/install.sh`, and `CHANGELOG.md` version values aligned.
- Keep release asset names stable:
  - `imgcli-linux-amd64`
  - `imgcli-linux-arm64`
  - `imgcli-darwin-amd64`
  - `imgcli-darwin-arm64`
  - `SHA256SUMS`
- Do not change release artifact naming or checksum behavior without updating install and workflow logic together.

## CLI contract rules

- Default output is JSON.
- Failures must emit structured JSON on `stderr` and exit non-zero.
- Machine-consumed commands must stay stable across versions.
- Error handling must use structured error codes, not string matching.
- Input ordering must remain deterministic for repeated `--input` flags and directory scans.
- New commands must focus on post-generation processing, packaging, and orchestration rather than model invocation.
- Reject or redesign requests that would turn `compose` into a command-line Figma: arbitrary coordinates, arbitrary layer stacks, free-form scene graphs, or unlimited overlay elements.

## Documentation rules

- `README.md` must describe value proposition, installation, JSON contract, and core workflows.
- `skills/imgcli/SKILL.md` must remain executable and reflect the current CLI surface.
- If a change affects installation, release assets, smoke flow, or command flags, update docs in the same change.
- If `compose` scope changes, update both `docs/v2-spec.md` and `docs/compose-boundary.md` in the same change.
