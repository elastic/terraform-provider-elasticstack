# Spec-Compliance Review: fix-index-mapping-update-unidirectional-check

**Lane:** spec-compliance  
**Verdict:** iterate  
**Change:** fix-index-mapping-update-unidirectional-check

---

## Summary

The core implementation required by this change has not been delivered. The `RequiresMappingsUpdate` method does not exist and `update.go` still uses the old bidirectional `StringSemanticEquals` gate, meaning every spec scenario in the "Update flow" requirement fails. Unit and acceptance tests are also absent. The `sort.missing`/`sort.mode` static-settings comparison is already satisfied by pre-existing code.

---

## Findings

### CRITICAL — `RequiresMappingsUpdate` method not implemented

**File:** `internal/elasticsearch/index/mappings_value.go`  
**Requirement:** Update flow (REQ-015–REQ-018) — "Mapping changes SHALL be applied using a **unidirectional** decision"

The spec requires a `RequiresMappingsUpdate(ctx context.Context, state MappingsValue) (bool, diag.Diagnostics)` method on `MappingsValue` that returns `!MappingsSemanticallyEqual(planMap, stateMap)`. This method does not exist in the file. Searching the entire codebase confirms it is absent.

**Impact:** The call site in `update.go` cannot be changed without this method; the entire fix is blocked.

---

### CRITICAL — `updateMappings` still uses bidirectional `StringSemanticEquals`

**File:** `internal/elasticsearch/index/index/update.go:194`  
**Requirement:** Update flow — "This decision SHALL NOT use the bidirectional semantic-equality check"

```go
areEqual, diags := planMappings.StringSemanticEquals(ctx, stateMappings)
```

This line must be replaced with `planMappings.RequiresMappingsUpdate(ctx, stateMappings)` (with the branch inverted). The current bidirectional check treats "plan is a superset of state" as equal, silently dropping added fields. Scenarios "Adding a mapping field calls the Put Mapping API" and "State already covering plan skips the Put Mapping API" both fail against this code.

---

### CRITICAL — No unit tests for `RequiresMappingsUpdate`

**File:** `internal/elasticsearch/index/mappings_value_test.go`  
**Requirement:** Tasks 1.3–1.4

`TestMappingsValue_RequiresMappingsUpdate` is absent. The six sub-cases listed in the tasks (plan adds field, plan is superset, state is superset / template extras, plan equals state, null/unknown plan, null/unknown state) have no coverage.

---

### CRITICAL — No acceptance test for regression (`elastic/protections-cloud#19769`)

**File:** `internal/elasticsearch/index/index/acc_test.go`  
**Requirement:** Scenario "Adding a mapping field calls the Put Mapping API" / Tasks 2.3

No test creates an index, adds a field to `mappings` in a second step, applies, then asserts via a direct Elasticsearch API read that the new field is present in the live cluster. Without this test the regression has no automated guard.

---

### CRITICAL — No `use_existing` mixed add/omit acceptance test

**File:** `internal/elasticsearch/index/index/acc_test.go`  
**Requirement:** Scenario "Adoption writes a field the plan adds and retains a field the plan omits" / Tasks 2.4

No test near `TestAccResourceIndexUseExistingAdopt` covers the asymmetric case (plan adds `new_field`, plan omits `legacy_field` that exists in live mapping). The adoption mapping path calls the shared `updateMappings` helper (create.go:183), which also suffers from the unfixed bidirectional gate.

---

## Passing items

### PASS — `sort.missing` and `sort.mode` static-settings comparison

**Files:** `internal/elasticsearch/index/settings_keys.go:67-68`, `internal/elasticsearch/index/index/use_existing.go:216-230`

`SettingSortMissing` and `SettingSortMode` are present in `StaticSettingsKeys` and handled in `compareStaticPlanAndES`. `toIndexSettings` (models.go:347) correctly extracts these values from the `sort` list-nested-attribute. The scenario "Adoption compares `sort.missing` against existing index" is already satisfied by pre-existing code.

---

## Required actions before approval

1. Add `RequiresMappingsUpdate` to `internal/elasticsearch/index/mappings_value.go` (tasks 1.1–1.2).
2. Replace `StringSemanticEquals` with `RequiresMappingsUpdate` (inverted branch) in `internal/elasticsearch/index/index/update.go:194` (task 2.1).
3. Add `TestMappingsValue_RequiresMappingsUpdate` unit tests (task 1.3).
4. Add regression acceptance test (adding a field to live index) (task 2.3).
5. Add `use_existing` mixed add/omit acceptance test (task 2.4).
