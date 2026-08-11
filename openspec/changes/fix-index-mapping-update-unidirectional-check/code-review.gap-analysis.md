# Gap Analysis — fix-index-mapping-update-unidirectional-check

**Lane:** gap-analysis
**Iteration:** 2 (post-implementation)
**Status:** implementation reviewed — branch `openspec/fix-index-mapping-update-unidirectional-check`, commit `e99e11c5f`

---

## Summary

All five findings from the iteration-1 design-level gap analysis (C1, W1, W2, S1, S2) are
resolved by the committed implementation. Adversarial re-analysis of the actual code finds no
new CRITICAL or WARNING gaps. Three SUGGESTION-level observations are reported below.

---

## Iteration-1 Findings — Resolution Status

### C1: Mirror-guard bug in null/unknown handling — RESOLVED

The concern was that copying `StringSemanticEquals`'s return expressions verbatim would produce
wrong results (null+null → `true`, causing a spurious PUT with empty body; state-null+plan-known
→ `false`, suppressing a needed PUT).

**Implementation at `mappings_value.go:170–175`:**

```go
if !typeutils.IsKnown(v) {
    return false, diags // nothing planned to add
}
if !typeutils.IsKnown(state) {
    return true, diags // state has nothing, plan adds everything
}
```

The guards follow the design's intent table directly, not `StringSemanticEquals`'s
`Normalized.Equal` return expressions. The correct return values are confirmed:
- plan null/unknown → false (no PUT) ✓
- state null/unknown, plan known → true (PUT) ✓
- both null → false (no PUT, no empty-body API call) ✓

### W1: plan=null, state=null test row absent — RESOLVED

`TestIndexMappingsValue_RequiresMappingsUpdate` at `mappings_value_test.go:466–470` includes:

```go
{
    name:  "plan null, state null — no update, not spurious PUT",
    plan:  index.NewMappingsNull(),
    state: index.NewMappingsNull(),
    want:  false,
},
```

This row catches the C1 mirror-guard bug: `null.Equal(null)` = true (the wrong return) vs. the
correct `false`. CI would fail this test if the guard were copied incorrectly.

### W2: Direct ES API verification pattern was unspecified — RESOLVED

`checkIndexHasField` at `acc_test.go:1168–1191` performs a direct Elasticsearch API read via
`getElasticsearchIndexStateByName`, checks `idxState.Mappings.Properties[fieldName]`, and
fails if the field is absent. This is independent of Terraform state — it catches the exact
failure mode where `readIndex` reflects the planned value even when PUT was skipped.

Both `TestAccResourceIndexMappingsUpdateRegression` and `TestAccResourceIndexUseExistingAdoptMappings`
use this helper.

### S1: Inline argument-order comment — RESOLVED

`mappings_value.go:183`:
```go
// planMap is receiver (plan intent), stateMap is argument (API state). Reversing reintroduces #19769.
return !MappingsSemanticallyEqual(planMap, stateMap), diags
```

### S2: Defense-in-depth null guard at call sites — RESOLVED

`update.go:91`:
```go
if typeutils.IsKnown(planModel.Mappings) {
    resp.Diagnostics.Append(r.updateMappings(...))
```

The guard uses `typeutils.IsKnown` (excludes both null and unknown) rather than just `!IsNull()`.
This is MORE conservative than S2 proposed and makes the null-plan skip explicit at the call boundary.

---

## Adversarial Review of Committed Code

### Confirmed safe: `dynamic_templates` and non-`properties` key handling

The correctness review (W2, iteration 1) noted that `MappingsSemanticallyEqual` uses
`reflect.DeepEqual` for top-level keys other than `"properties"` (e.g., `runtime`,
`dynamic_templates`, `_meta`). The concern was that ES-injected extras in those keys might
cause spurious PUTs under `RequiresMappingsUpdate`.

Adversarial verification shows **no regression**. For any configuration of plan and state
values at a non-`properties` key, `MappingsSemanticallyEqual` and `StringSemanticEquals` agree
on whether the keys are "equal" (both use the same `scalarSemanticEqual` path). The only
difference between old and new is the `||` in `StringSemanticEquals`; but since
`scalarSemanticEqual` on two structurally different maps returns `false` in both argument
orders, adding `||` would not have helped suppress the spurious PUT in the old code either.
The pre-existing `dynamic_templates`/`runtime`/`_meta` limitation applies equally to the old
and new code.

### Confirmed safe: plan modifier composition

The `mappingsPlanModifier.mergeMappingsForPlan` merges state-only extras into the plan value
before Terraform calls Update. `updateMappings` therefore receives a merged plan that includes
both user-specified fields and state extras.

`RequiresMappingsUpdate(merged_plan, state)` with merged_plan = {user_fields ∪ state_extras}
and state = {state_fields ∪ state_extras}:
- `MappingsSemanticallyEqual(merged_plan_map, state_map)` iterates over merged_plan keys.
  If user added a new field (e.g., field2), merged_plan has field2 but state does not →
  returns false → `RequiresMappingsUpdate` returns true → PUT called. ✓
- If user added no new fields, merged_plan is a subset of state (by construction of the
  plan modifier) → returns true → `RequiresMappingsUpdate` returns false → no PUT. ✓

The composition is correct. The gap analysis (iteration 1)'s "Integration point confirmed
correct" finding is confirmed against the actual implementation.

### Confirmed safe: argument order at both call sites

- `update.go:197`: `planMappings.RequiresMappingsUpdate(ctx, stateMappings)` — receiver=plan ✓
- `create.go:183` via `r.updateMappings(ctx, client, concreteName, plan.Mappings, synthetic.Mappings)`:
  first arg is plan (user intent), second is `synthetic.Mappings` (live existing index read). ✓

No call site has the argument order swapped.

### Confirmed safe: `RequiresMappingsUpdate` is not called from any other location

Grep across the entire repository shows `RequiresMappingsUpdate` is referenced at exactly
three locations: the implementation site (`mappings_value.go:167`), the call site
(`update.go:197`), and the unit test (`mappings_value_test.go:494`). No other caller exists
that could invert the argument order.

---

## New Findings

### SUGGESTION — S3: `checkIndexHasField` verifies field existence but not field type

**Location:** `internal/elasticsearch/index/index/acc_test.go:1186`

```go
if _, ok := idxState.Mappings.Properties[fieldName]; !ok {
    return fmt.Errorf("...")
}
```

The helper confirms the field is present in the live cluster mapping but does not assert the
field's type matches the plan. For the regression test's purpose (verifying PUT was not
skipped), existence is sufficient. A type mismatch would cause a separate apply-time error
from the ES API, which is outside this change's scope.

Low-risk: ES's additive PUT Mapping semantics mean a field with the wrong type cannot appear
silently — ES returns an error on mismatched type updates. No action required for this change.

---

### SUGGESTION — S4: Regression test does not assert `field1` is retained after the update step

**Location:** `TestAccResourceIndexMappingsUpdateRegression`, step 2 check
(`acc_test.go:1114–1121`)

The update step verifies only that `field2` was added (`checkIndexHasField(..., "field2")`).
It does not verify that `field1` (from the create step) is still present after the PUT.
ES's additive-only PUT Mapping semantics guarantee `field1` cannot be removed, and
`TestAccResourceIndexUseExistingAdoptMappings` exercises the "retain field not in plan" path
more explicitly, so this gap does not represent a real risk. Adding
`checkIndexHasField(..., "field1")` to the update step would make the test more self-contained
as documentation.

No action required for this change; defer if desired.

---

### SUGGESTION — S5: `decodeMappingPair` precondition is documented but not enforced

**Location:** `internal/elasticsearch/index/mappings_value.go:187–189`

```go
// decodeMappingPair JSON-decodes both MappingsValue receivers to raw maps.
// Must only be called after null/unknown guards — calling it on a null or
// unknown value produces an unmarshal error on the empty string.
```

A future developer adding a third method that reuses `decodeMappingPair` could forget the
precondition and receive an `AddError` diagnostic rather than a compile-time or panic-time
signal. The comment is adequate; the failure mode is recoverable (error returned, not panic).
No change needed.

---

## Verdict

**`approve`**

All CRITICAL and WARNING findings from the iteration-1 gap analysis are resolved in the
committed implementation. No new CRITICAL or WARNING gaps were found. The implementation is
narrowly scoped, the argument-order invariant is protected by documentation, inline comment,
and asymmetric unit tests, and the regression tests verify the fix via direct API reads
independent of Terraform state.
