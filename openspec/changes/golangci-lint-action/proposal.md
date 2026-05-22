## Why

The provider CI `lint` job currently runs all lint checks sequentially in a single job: OpenSpec validation, golangci-lint (with a custom binary), format checking, code generation, and docs generation. This is slow and means a fast check like `check-fmt` has to wait for golangci-lint to finish before it runs.

`golangci/golangci-lint-action` is the official GitHub Action for golangci-lint. It provides better caching of the linter binary and lint cache between runs, and adds inline PR annotations. It natively supports the module plugin system — it detects `.custom-gcl.yml`, builds the custom binary via `golangci-lint custom`, and runs that binary automatically.

Splitting golangci-lint into its own job lets it run in parallel with the remaining lint checks, reducing overall CI wall-clock time.

## What Changes

- **New `golangci-lint` job** in `.github/workflows/provider.yml`: a dedicated job that runs `golangci/golangci-lint-action` with the pinned version (`v2.12.2`), `pull-requests: read` permission for inline annotations, and `needs: classify` (same gate as the existing `lint` job).

- **Modified `lint` job**: removes the `golangci-lint` Make target call and calls the remaining targets directly (`make check-openspec check-fmt gen check-docs`). The Node, Go, and Terraform setup steps are unchanged.

- **`make check-lint` target removed** from the CI workflow invocation — the targets are called directly. The Makefile target itself is not changed (it remains as a local dev convenience).

- **`gate` job updated**: adds `golangci-lint` to its `needs` list alongside `build`, `lint`, and `test`.

- **`gate-provider.js` updated**: adds `golangciLintResult` as a required input alongside `lintResult`. Both must pass (or both must be skipped) for the gate to pass.

- **`.custom-gcl.yml` updated**: bumps `version` from `v2.11.4` to `v2.12.2` to align with the action's `version` input and the Makefile's installed binary version.

## Capabilities

### New Capabilities

None — this is a CI infrastructure change only.

### Modified Capabilities

None — no spec-level behaviour changes.

## Impact

- `.github/workflows/provider.yml`: `golangci-lint` job added; `lint` job steps changed; `gate` job `needs` and env updated.
- `.github/scripts/workflows/lib/gate-provider.js`: signature and logic extended to accept and validate `golangciLintResult`.
- `.github/scripts/workflows/lib/runners/gate.js`: `fields` array for the `provider` gate updated to include `GOLANGCI_LINT_RESULT`.
- `.custom-gcl.yml`: version bump to `v2.12.2`.
