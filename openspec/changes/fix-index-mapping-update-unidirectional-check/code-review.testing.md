# Testing Review — fix-index-mapping-update-unidirectional-check

**Lane:** testing  
**Iteration:** 2 (post-implementation)  
**Commit reviewed:** e99e11c5f

---

## Summary

Implementation is complete. All three test tasks from iteration 1's C1 finding have been executed. The unit tests for `RequiresMappingsUpdate` are comprehensive and include the asymmetric argument-order case that directly pins the core bug. The two new acceptance tests use direct Elasticsearch API reads independent of Terraform state — a high-quality check that would have caught `elastic/protections-cloud#19769` had it existed before the bug.

Two pre-existing test gaps are now visible through the delta spec's scenario list. They are not failures of this change and do not block approval.

---

## Resolved findings from iteration 1

### C1 — RESOLVED: Three test tasks executed and meaningful

**Task 1.3 — `TestIndexMappingsValue_RequiresMappingsUpdate`**  
`internal/elasticsearch/index/mappings_value_test.go:422`

Nine cases covering: plan adds field (→ true), plan is strict superset (→ true), state is superset of plan (template-injected extras, → false), plan equals state (→ false), plan null (→ false), plan null and state null (→ false), plan unknown (→ false), state null with non-null plan (→ true), state unknown with non-null plan (→ true).

The asymmetric "state is superset of plan → false" case is the direct pin for the argument-order bug. Reversing `MappingsSemanticallyEqual(planMap, stateMap)` to `MappingsSemanticallyEqual(stateMap, planMap)` would fail this case because the test verifies `false` not `true`.

**Task 2.3 — `TestAccResourceIndexMappingsUpdateRegression`**  
`internal/elasticsearch/index/index/acc_test.go:1093`

Two-step test: create with `field1`, update adding `field2`, assert `field2` is present in the live cluster via `checkIndexHasField`. The direct ES read (not Terraform state) is exactly what distinguishes this from a plan-level check. The test comment at `:1117–1119` explains why: Terraform state can reflect the planned field even when the Put Mapping API call was skipped.

**Task 2.4 — `TestAccResourceIndexUseExistingAdoptMappings`**  
`internal/elasticsearch/index/index/acc_test.go:1130`

Pre-creates an index OOB with `legacy_field`, adopts with a plan specifying only `new_field`. Asserts both fields via `checkIndexHasField`: `new_field` was written by the Put Mapping API, `legacy_field` was retained (the API cannot delete mapping fields). Covers the full "add and retain" contract from the spec scenario.

---

## Remaining findings

### WARNING — W2: `compareStaticSettings` has no unit test for `sort_missing` or `sort_mode` mismatch (pre-existing gap)

**Location:** `internal/elasticsearch/index/index/use_existing_test.go` (entire file, 582 lines — no `sortMissing` or `sortMode` case)  
**Code handles it:** `internal/elasticsearch/index/index/use_existing.go:216–229`

The delta spec adds `sort.missing` and `sort.mode` to the formal list of static settings checked during adoption (scenario: "Adoption compares `sort.missing` against existing index"). The production code in `compareStaticPlanAndES` handles them at lines 216 and 224, using the same `planStringSliceForSortOrder` / `stringSliceOrderedFromAny` / `slicesEqual` pattern as `sort_order`. No test pins the mismatch path for either setting.

This is a pre-existing gap: `use_existing.go` was not changed by this commit. The code is structurally identical to the `sort_order` case which does have tests (`Test_compareStaticSettings_sortOrder_orderedEquality`, `Test_compareStaticSettings_sortOrder_orderMatters`). The risk of a latent bug is low, but the spec scenario has zero test anchor.

**Minimum required test (follow-up, not blocking this PR):**
```go
func Test_compareStaticSettings_sortMissing_mismatch(t *testing.T) {
    // plan: sort.missing = ["_first"], existing sort.missing = ["_last"]
    // compareStaticSettings should return one mismatch on "sort_missing"
}
func Test_compareStaticSettings_sortMode_mismatch(t *testing.T) {
    // plan: sort.mode = "min", existing sort.mode = "max"
    // compareStaticSettings should return one mismatch on "sort_mode"
}
```

---

### WARNING — W3: "Removed dynamic setting is nulled" scenario has no acceptance test (pre-existing gap)

**Location:** `internal/elasticsearch/index/index/acc_test.go` — no test removes a dynamic setting between steps

The delta spec formalises the "Removed dynamic setting is nulled" scenario. The production code is at `internal/elasticsearch/index/index/update.go:170–183` (unchanged by this commit). No acceptance test removes a setting between Terraform apply steps and verifies the live cluster value is cleared.

This gap predates this change. The `updateSettings` logic is correct, but CI cannot catch a future regression where the null-pass is dropped.

**Suggested test (follow-up, not blocking):** Add a step to `TestAccResourceIndexSettings` (or a separate test) that removes a dynamic setting (e.g. `refresh_interval`) from config and calls `checkIndexSetting` to confirm the live cluster no longer reports it.

---

### NOTE — W4: Template-drift assertion is plan-level (acceptable)

`TestAccResourceIndexTemplateNoMappingDrift` and `TestAccResourceIndexTemplateUserMappingNoDrift` use `plancheck.ExpectEmptyPlan()`. This is the conventional approach for no-op drift tests and is acceptable. The regression test from task 2.3 implicitly demonstrates correct apply-time behaviour for the add-field path, partially compensating for the absence of an apply-time assertion in the no-op path.

---

## Scenario coverage matrix

| Spec scenario | Unit test | Acceptance test | Gap |
|---|---|---|---|
| Template-injected mappings do not cause mapping update | `TestIndexMappingsValue_RequiresMappingsUpdate` "state is superset" | `TestAccResourceIndexTemplateNoMappingDrift`, `TestAccResourceIndexTemplateUserMappingNoDrift` (plan-level) | W4 (apply-time, acceptable) |
| Adding a mapping field calls the Put Mapping API | `TestIndexMappingsValue_RequiresMappingsUpdate` "plan adds field" | `TestAccResourceIndexMappingsUpdateRegression` (direct ES read) | **Fully covered** |
| State already covering plan skips the Put Mapping API | `TestIndexMappingsValue_RequiresMappingsUpdate` "state superset", "equal" | Plan-level no-op tests | Adequate |
| Removed alias is deleted | — | `TestAccResourceIndexUseExistingAdoptAliasReconcile` | Adequate |
| Removed dynamic setting is nulled | — | None | **W3 — pre-existing gap** |
| Adoption compares `sort.missing` against existing index | None | None | **W2 — pre-existing gap** |
| Adoption writes a field the plan adds and retains a field the plan omits | `TestIndexMappingsValue_RequiresMappingsUpdate` "plan adds field", "state superset" | `TestAccResourceIndexUseExistingAdoptMappings` (direct ES read for both fields) | **Fully covered** |

---

## Verdict

`approve` — the core bug and all new spec scenarios are properly tested. The two remaining warnings are pre-existing gaps exposed by the expanded spec but not introduced by this change. They should be addressed in a follow-up but do not block this PR.
