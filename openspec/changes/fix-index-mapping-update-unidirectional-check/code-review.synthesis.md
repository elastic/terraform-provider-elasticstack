# Code Review Synthesis — fix-index-mapping-update-unidirectional-check

**Iteration:** 2  
**Verdict:** `done`  
**Lanes:** correctness · gap-analysis · testing · spec-drift · spec-compliance · maintainability · standards · security · api-contract · performance

---

## Executive summary

All 10 lanes approve the iteration-2 implementation. All 9 required fixes from the iteration-1 synthesis (R1–R9) are resolved. `openspec validate fix-index-mapping-update-unidirectional-check` passes. No required fixes remain.

The implementation is narrowly scoped, correct, and well-tested:
- `RequiresMappingsUpdate` unidirectionally checks whether plan has content absent from state.
- The bidirectional `StringSemanticEquals` is untouched, preserving plan-time semantics.
- Both call sites (update and adoption) route through the shared `updateMappings` helper.
- Regression acceptance tests use direct Elasticsearch API reads (not Terraform state).
- All spec scenarios are covered at unit and/or acceptance level.

---

## Required fixes — none

All iteration-1 required fixes are resolved:

| Fix | Resolution |
|---|---|
| R1 — Re-run implementation | All 13 tasks complete; `RequiresMappingsUpdate` implemented |
| R2 — Correct null/unknown guards | Guards follow design table: plan null/unknown → `false`, state null/unknown → `true` |
| R3 — `plan=null, state=null → false` test row | Present at `mappings_value_test.go:466` |
| R4 — `typeutils.IsKnown` usage | Used at `mappings_value.go:170` and `:173` |
| R5 — Test name `TestIndexMappingsValue_RequiresMappingsUpdate` | Correct name at `mappings_value_test.go:422` |
| R6 — Extract `decodeMappingPair` helper | Extracted at `mappings_value.go:190–202` |
| R7 — Doc comment states receiver=plan, argument=state | Present at `mappings_value.go:155–164`; names `#19769` |
| R8 — Null guard at `update.go:90` | Guard added at `update.go:91–96` using `typeutils.IsKnown` |
| R9 — Acceptance test uses direct ES API read | `checkIndexHasField` at `acc_test.go:1168`; used by both regression tests |

---

## Optional / deferred follow-up

These findings were raised by lanes but are non-blocking. All lanes approved despite them.

### Standards: `.gitignore` entries unrelated to the fix (standards:W1)

**Location:** `.gitignore:70–79`

Commit `e99e11c5f` includes ten Gas City/Beads tooling entries (`.beads/`, `.gc/`, `*.db`, etc.) unrelated to `RequiresMappingsUpdate`. The standards lane approved but flags this as hygiene debt: extracting to a separate commit or PR would keep the change surface clean.

**Defer:** follow-up commit targeting main.

---

### Pre-existing test gaps exposed by the delta spec (testing:W2, testing:W3, spec-compliance:WARNING)

**W2 — `sort.missing` / `sort.mode` mismatch unit tests** (`use_existing_test.go`)  
`compareStaticPlanAndES` handles `SettingSortMissing` and `SettingSortMode` but no unit test pins the mismatch path. The spec delta formally scenarios `sort.missing` comparison, and the code pattern is structurally identical to the already-tested `sort_order` case — low regression risk.

Suggested tests:
```go
func Test_compareStaticSettings_sortMissing_mismatch(t *testing.T) { ... }
func Test_compareStaticSettings_sortMode_mismatch(t *testing.T) { ... }
```

**W3 — "Removed dynamic setting is nulled" acceptance test** (`acc_test.go`)  
`updateSettings` at `update.go:170–183` handles null-dynamic-settings correctly, but no acceptance test removes a setting between steps and verifies the live cluster value is cleared.

Both gaps pre-date this change and do not block this PR.

---

### Minor style suggestions (not blocking)

| Source | Finding | Location |
|---|---|---|
| correctness:S1 | Doc comment on `RequiresMappingsUpdate` doesn't note non-`properties` scope limitation of "template extras no PUT" guarantee | `mappings_value.go:155–166` |
| correctness:S2 | Dead `var diags diag.Diagnostics` at line 195 (`updateMappings`); `:=` at line 197 re-declares it | `update.go:195` |
| gap-analysis:S3 | `checkIndexHasField` checks field existence only, not field type | `acc_test.go:1186` |
| gap-analysis:S4 | Regression test step 2 doesn't assert `field1` retained after adding `field2` | `acc_test.go:1114–1121` |
| gap-analysis:S5 | `decodeMappingPair` precondition documented but not enforced at runtime | `mappings_value.go:187–189` |
| maintainability:S3 | `StringSemanticEquals` uses `v.IsNull() || v.IsUnknown()`; `RequiresMappingsUpdate` uses `!typeutils.IsKnown` — two idioms in the same file | `mappings_value.go:135` vs `:170` |
| maintainability:S4 | Two test cases in `TestIndexMappingsValue_RequiresMappingsUpdate` use byte-identical data but different names | `mappings_value_test.go:436–452` |
| maintainability:S5 | `updateMappings` returns `nil` on early exit; `updateAliases`/`updateSettings` return `diags` | `update.go:202` |
| standards:S1 | Inline comment at `mappings_value.go:183` repeats the doc comment above it | `mappings_value.go:183` |
| security:SUGGESTION | `decodeMappingPair` precondition not enforced; a future caller could violate it | `mappings_value.go:190` |

---

## Lane verdicts summary

| Lane | Verdict | Required fixes | Notable findings |
|---|---|---|---|
| correctness | approve | none | S1: doc comment gap; S2: dead `var` |
| testing | approve | none | W2, W3: pre-existing test gaps |
| gap-analysis | approve | none | S3–S5: low-priority style suggestions |
| spec-compliance | approve | none | WARNING: `sort.missing` unit test missing (pre-existing) |
| spec-drift | approve | none | No drift detected |
| maintainability | approve | none | S3–S5: style suggestions |
| standards | approve | none | W1: `.gitignore` entries (defer to follow-up) |
| security | approve | none | SUGGESTION: enforce `decodeMappingPair` precondition |
| api-contract | approve | none | W1 from iter-1 fully resolved |
| performance | approve | none | Marginal improvement over old code |

---

## Correctness confirmations

The following correctness properties were verified across multiple lanes and are confirmed sound:

1. **Argument order pin** — `!MappingsSemanticallyEqual(planMap, stateMap)` with plan as receiver. Reversing reintroduces `#19769`. Protected by inline comment, doc comment, and asymmetric unit tests.
2. **Null/unknown handling** — all three cases correct: plan null/unknown → `false`; state null/unknown with known plan → `true`; both null/unknown → `false` (via first guard).
3. **Branch inversion at wire site** — old `if areEqual { return nil }` replaced by `if !requiresUpdate { return nil }`. Logically correct.
4. **Plan modifier composition** — `mergeMappingsForPlan` produces merged plan (user + state extras); `RequiresMappingsUpdate(merged_plan, state)` correctly returns `true` only when user added a field not in state.
5. **Adoption path** — `create.go` delegates to shared `updateMappings`; fix propagates automatically; no separate call site.
6. **`StringSemanticEquals` untouched** — plan-time semantic equality and `RequiresReplace` behaviour unchanged.
7. **No other call sites** — `RequiresMappingsUpdate` is referenced at exactly three locations (definition, call site, unit test).
