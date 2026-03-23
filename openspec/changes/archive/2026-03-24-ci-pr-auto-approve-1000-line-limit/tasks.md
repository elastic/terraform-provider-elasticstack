## 1. Implementation

- [x] 1.1 Set `maxEditedLines` to `1000` in `scripts/auto-approve/evaluator.go`.
- [x] 1.2 Update Copilot diff threshold cases in `scripts/auto-approve/evaluator_test.go`: reject when `additions + deletions` is 1000 or more, expect `edited lines must be < 1000`; rename case descriptions accordingly.
- [x] 1.3 Add a table-driven case that approves a Copilot PR with total edits strictly between 300 and 1000 (e.g. 400 additions + 500 deletions) so the new band is covered.

## 2. Canonical spec and verification

- [x] 2.1 Update REQ-009 and the large-PR scenario in `openspec/specs/ci-pr-auto-approve/spec.md` to use `1000` instead of `300` (match the delta in this change).
- [x] 2.2 Run `go test ./scripts/auto-approve/...` and `make check-openspec`.
