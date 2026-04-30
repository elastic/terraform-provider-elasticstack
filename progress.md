# Progress

## Task 1 — `elasticsearch-resource-envelope`

- [x] Implementation
- [x] Validation run
  - [x] `make build` — PASS
  - [x] `make lint` — PASS
  - [x] `go test -v ./internal/entitycore/...` — PASS
  - [x] `make check-openspec` — PASS
- [x] Coverage review
  - `resource_envelope.go` branches now covered for `Delete` state.Get error and delete callback diagnostics
  - Added tests for defensive Create/Update defaults and concrete method override behavior
- [x] Spec review (subtasks 1.1–1.8)
  - Findings written to `task1-spec-review.md`
  - Result: Task 1 remains spec-compliant after review fixes

## Review
- Correct: Schema injection clones the blocks map before adding `elasticsearch_connection`; Read/Delete follow the planned diagnostic gates and typed client resolution; ImportState passthrough matches the spec and doc updates describe the envelope pattern clearly.
- Fixed: Replaced silent no-op Create/Update stubs with defensive error diagnostics so a forgotten concrete override fails loudly instead of appearing to succeed with no state change.
- Fixed: Added regression tests covering the new defensive Create/Update behavior, confirmation that concrete overrides still win via method promotion, `Delete` state.Get short-circuit, and delete callback diagnostic propagation.
- Note: `plan.md` was not present at the requested path; review used the OpenSpec change docs under `openspec/changes/elasticsearch-resource-envelope/` as the task plan source.
- Note: Targeted validation re-run: `go test ./internal/entitycore/...` — PASS.

## Remaining Tasks

- [ ] Task 2: Migrate `elasticsearch_security_user`
- [ ] Task 3: Migrate `elasticsearch_security_system_user`
- [ ] Task 4: Migrate `elasticsearch_security_role`
- [ ] Task 5: Migrate `elasticsearch_security_role_mapping`
- [ ] Task 6: Verification (acceptance tests, generated docs diff)
