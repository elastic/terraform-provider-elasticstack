## 1. Align the lint contract

- [x] 1.1 Update the canonical `openspec/specs/makefile-workflows/spec.md` requirement for `golangci-lint` so it defines repository-wide `./...` linting instead of `internal/`-only linting.
- [x] 1.2 Keep the requirement text explicit that `lint` uses fix mode while `check-lint` remains check-only for golangci-lint.

## 2. Align implementation and verification

- [x] 2.1 Update the root `Makefile` if needed so the `golangci-lint` target runs against `./...` and the aggregate lint targets inherit that scope.
- [x] 2.2 Run `make lint` after the lint-scope change so the broader repository-wide target is exercised end to end.
- [x] 2.3 Fix any lint errors surfaced by `make lint` so the repository passes under the expanded Go package scope.
- [x] 2.4 Run the relevant OpenSpec and lint-oriented checks to confirm the updated spec and Makefile stay aligned with repository behavior.
