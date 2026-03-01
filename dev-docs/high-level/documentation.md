# Documentation

Provider docs are generated; keep them in sync with code changes.

## Where docs come from

- Templates: `templates/`
- Examples: `examples/`
- Generated output: `docs/`

Canonical contributor guidance: “Updating Documentation” in [`CONTRIBUTING.md`](../../CONTRIBUTING.md#updating-documentation).

## Regenerating docs

- Generate docs: `make docs-generate`
- Note: `make lint` also runs `docs-generate` and will fail if generated docs are stale.

