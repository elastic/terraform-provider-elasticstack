## 1. Apply spec delta

- [ ] 1.1 Run `make check-openspec` (or `openspec validate`) on the change and fix any structural issues reported for the delta spec.
- [ ] 1.2 Archive or apply the change so `openspec/specs/ci-pr-auto-approve/spec.md` matches the delta (REQ-001 removed; REQ-002–REQ-014 renumbered to REQ-001–REQ-013) per project OpenSpec workflow.

## 2. Remove draft PR filter in script

- [ ] 2.1 Remove draft exclusion from `scripts/auto-approve` (e.g. `evaluator.go` and any related types/fixtures).
- [ ] 2.2 Update unit tests (`evaluator_test.go`, `main_test.go` as needed) so draft PRs are no longer expected to fail with a draft scope reason.

## 3. Follow-up cleanup

- [ ] 3.1 Search the repo for references to `REQ-001` / “Evaluation scope” / draft filtering for auto-approve and update docs or comments if needed.
- [ ] 3.2 Re-run `make check-openspec` and targeted `go test` for `scripts/auto-approve` after the canonical spec and code are updated.
