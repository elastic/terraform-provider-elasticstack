# Verification notes (migrate-changelog-engine-to-go)

## Tasks 9.3 and 9.4 — pragmatic parity

**9.3 (PR-body / `validate-pr-section` verdicts)**  
The pre-migration JavaScript is removed from the tree, so an online A/B against the old runtime is not reproducible from this branch. Parity is asserted by porting the `.test.mjs` tables into Go: `internal/section/parse_test.go`, `internal/section/render_test.go`, and `internal/prcheck/validate_test.go`, including the spec-pinned mismatch string `RuleCBreakingOnlyWhenBreakingImpactMsg` from `internal/section` (exported and asserted in tests).

**9.4 (`changelog-generation` / `CHANGELOG.md`)**  
Full workflow execution is validated in CI when the migration PR runs. Locally, wiring is validated by successful `run-engine`, rewriter, and engine tests (`internal/engine/engine_test.go`, `internal/rewriter/section_test.go`, …) which pin expected `CHANGELOG.md` mutations relative to fixtures. There is no remaining JS baseline for a byte-diff in the working tree.

## Local command checkpoint (manual re-run)

- `make workflow-test` — must exit `0`.
- `go test ./scripts/changelog/... -count=1` — all tests green.
- `make build` and `make check-lint` — expected clean on CI; a workstation with stray **untracked** files in the repo root can cause `make check-fmt` (part of `check-lint`) to fail because its guard treats any non-empty `git status --porcelain` as failure.
