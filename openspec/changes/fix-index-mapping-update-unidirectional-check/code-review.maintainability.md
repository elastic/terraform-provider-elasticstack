# Maintainability Review — fix-index-mapping-update-unidirectional-check

**Lane:** maintainability  
**Iteration:** 2 (post-implementation)  
**Status:** final — reviewing committed code on branch `openspec/fix-index-mapping-update-unidirectional-check`

---

## Summary

The implementation is clean, minimal, and faithful to the design. All four pre-implementation findings from iteration 1 are resolved. Three new suggestions emerged from reading the code; none rise to WARNING level. Verdict: **approve**.

---

## Previous Findings — Resolution Status

### W1 (Duplicate decode logic) — RESOLVED

`decodeMappingPair` was extracted as a private helper and called from both `StringSemanticEquals` and `RequiresMappingsUpdate`. The JSON-decode and error-diagnostic logic lives in one place.

### W2 (No argument-order coverage) — RESOLVED

The `TestIndexMappingsValue_RequiresMappingsUpdate` test suite includes asymmetric cases that catch argument reversal: "plan adds field not in state → update required" (field1And2 vs field1Only, want=true) and "state is superset of plan → no update" (field1Only vs field1Plus3, want=false). If `MappingsSemanticallyEqual(planMap, stateMap)` were accidentally swapped to `MappingsSemanticallyEqual(stateMap, planMap)`, both cases would return the wrong value and fail CI.

### S1 (StringSemanticEquals comment described unidirectional) — RESOLVED

The method comment at `mappings_value.go:124-126` now explicitly says "The check is bidirectional: extras on either side are tolerated."

### S2 (Type-level comment repeated the misdescription) — RESOLVED ADEQUATELY

The type-level comment (`mappings_value.go:102-106`) describes the design intent ("API value as a non-drifting superset of user intent"), while the `RequiresMappingsUpdate` method comment describes its semantics as unidirectional. Readers can distinguish the two.

---

## New Findings

### SUGGESTION — S3: Null/unknown check inconsistency within the same file

**Location:** `internal/elasticsearch/index/mappings_value.go:135–141` vs `:170–175`

`StringSemanticEquals` uses `v.IsNull() || v.IsUnknown()` for its null/unknown guard. `RequiresMappingsUpdate` uses `!typeutils.IsKnown(v)` — logically equivalent, but two different idioms in the same file forty lines apart. A future reader modifying the guards in one method may not realize they need to match the other. Using one consistent idiom (the existing `v.IsNull() || v.IsUnknown()` pattern, or the `typeutils.IsKnown` wrapper) would remove this friction.

No behavioral difference; strictly a readability issue.

---

### SUGGESTION — S4: Two test cases in `TestIndexMappingsValue_RequiresMappingsUpdate` exercise identical data

**Location:** `internal/elasticsearch/index/mappings_value_test.go:436–452`

```go
// case 1:
{name: "plan adds field not in state — update required", plan: field1And2, state: field1Only, want: true},
// case 2:
{name: "plan is strict superset of state — update required",
 plan:  field1And2,
 state: index.NewMappingsValue(`{"properties":{"field1":{"type":"text"}}}`), // identical to field1Only
 want:  true},
```

Case 2 creates a new `MappingsValue` inline whose JSON is byte-for-byte the same as the pre-declared `field1Only`. Both cases route through the same code path and assert the same outcome. The two test names describe conceptually different scenarios, but the data does not distinguish them — making the duplication difficult to justify. Collapsing to one case (reusing `field1Only`) reduces noise without losing any coverage.

---

### SUGGESTION — S5: `return nil` vs `return diags` inconsistency in `updateMappings`

**Location:** `internal/elasticsearch/index/index/update.go:202`

```go
if !requiresUpdate {
    return nil  // diags is nil at this point, so semantically identical
}
```

Both `updateAliases` (line 145) and `updateSettings` (line 185) return `diags` at their early-return points. The new `updateMappings` returns `nil`. The two are semantically identical since `diags` is nil here, but the inconsistency adds a small "why is this different?" pause for the next reader.

---

## Positive Observations

- **Concern separation is correct and complete.** `RequiresMappingsUpdate` is unidirectional, `StringSemanticEquals` is bidirectional; each is documented with why. The design risk of "two similarly-shaped oppositely-directioned predicates" is squarely addressed by the doc comment warning and by the asymmetric unit-test cases.
- **`decodeMappingPair` extraction.** The helper is documented with its precondition ("must only be called after null/unknown guards") and its error handling is consistent with the rest of the file. A future method that needs to compare two `MappingsValue` instances can reuse it rather than copy the decode pattern again.
- **`updateMappings` call-site clarity.** Replacing `StringSemanticEquals` with `RequiresMappingsUpdate` and inverting the branch makes the intent explicit: the function now reads "call the API when the plan requires an update" rather than "skip the API when the values are equal." This is easier to audit.
- **Acceptance test coverage.** `TestAccResourceIndexMappingsUpdateRegression` checks the live cluster via direct ES API read (`checkIndexHasField`), not just Terraform state — exactly the right approach to pin this class of bug where state can reflect planned values even when the API call was skipped.
- **`TestAccResourceIndexUseExistingAdoptMappings`** exercises both halves of the adoption asymmetry in one scenario (new field written, legacy field retained), matching the acceptance criteria in the proposal.

---

## Verdict

`approve`

All WARNINGs from the design-level review are resolved. Three suggestions (S3–S5) are minor style inconsistencies with no correctness impact. The implementation is well-scoped, the comment strategy is adequate for the argument-order risk, and the test suite meaningfully covers the new behavior.
