# Tasks Review Synthesis — fix-index-mapping-update-unidirectional-check

**Overall verdict: done**

All three review lanes approved. The tasks.md was restructured after the previous `iterate` verdict and now satisfies completeness, vertical-slice, and traceability requirements.

---

## Lane verdicts

| Lane | Verdict | Bead |
|---|---|---|
| Completeness | approve | ma-lzj |
| Vertical slices | approve | ma-slu |
| Traceability | approve | ma-3ti |

---

## What changed since the previous iteration

The prior synthesis (iteration 0) required tasks 1–4 to be restructured from horizontal layers into paired vertical slices, and task 6.2 to be removed or folded into slice checkpoints. The revised tasks.md implements both:

- **Slice A** (section 1): `RequiresMappingsUpdate` method + unit tests + `go test` checkpoint.
- **Slice B** (section 2): wire helper into call sites + acceptance tests + `TF_ACC=1 go test` checkpoint.
- **Spec sync** (section 3): spec delta update + `openspec validate` checkpoint.
- **Validate** (section 4): `go build` + `go vet`.
- Section 6 (the redundant "run tests" block) is absent from the revised tasks.md.

---

## Residual / informational findings

### [SUGGESTION] Traceability: pre-existing sort.missing/sort.mode delta content (ma-3ti)

The spec delta includes `sort.missing`/`sort.mode` content carried over in full-replacement format from a prior change. This content is not part of the current change's scope and is flagged as a suggestion only. It does not require any action for this change.

### [INFORMATIONAL] Slice review: task 2.2 is a no-code-change confirmation (ma-slu)

Task 2.2 confirms that `adoptExistingIndexOnCreate` picks up the fix via `updateMappings` without an additional code change. It is a checklist verification step, not an implementation task. The lane accepted this as cosmetic; no restructuring required.

### [INFORMATIONAL] Acceptance test negative-scenario coverage (ma-3ti, carried from iteration 0)

Two negative scenarios (template-injected extras, state-superset-of-plan) are exercised by unit tests in 1.3 but are not separately covered by acceptance tests. The traceability lane accepted the unit-test mitigation. Track as optional follow-up if acceptance test coverage is extended.

---

## Confirmed clean

- **Completeness**: all What Changes items (add `RequiresMappingsUpdate`, wire into `updateMappings`, acceptance tests, `use_existing` adoption coverage) and both modified Capabilities are covered. Unit, acceptance, spec-validation, and build/vet tasks are all present.
- **Traceability**: every task maps to a requirement (REQ-015–REQ-018, `use_existing` adoption unidirectional decision). No orphan tasks. No uncovered requirements.
- **Vertical slices**: each of the four sections ends with a runnable verification checkpoint. No horizontal layering remains.

---

## Summary for implementer

Tasks are approved as written. Implement in slice order:

1. **Slice A** — add `RequiresMappingsUpdate` and unit tests; verify with `go test ./internal/elasticsearch/index/...`.
2. **Slice B** — wire helper, add acceptance tests; verify with `TF_ACC=1 go test ./internal/elasticsearch/index/index/...`.
3. **Spec sync** — update spec delta; validate with `openspec validate`.
4. **Validate** — `go build ./...` and `go vet ./...`.

No blocking changes required before implementation begins.
