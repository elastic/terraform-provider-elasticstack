## Context

Canonical requirements for this resource live in [`openspec/specs/elasticsearch-security-role-mapping/spec.md`](../../specs/elasticsearch-security-role-mapping/spec.md). Implementation lives in [`internal/elasticsearch/security/rolemapping/`](../../../internal/elasticsearch/security/rolemapping/).

The codebase already has multiple well-established examples of the custom-type pattern:
- `internal/utils/customtypes/json_with_defaults_value.go` — `JSONWithDefaultsValue[TModel]` with `StringSemanticEquals`
- `internal/utils/customtypes/index_settings_value.go` — `IndexSettingsValue`
- `internal/utils/customtypes/normalized_yaml_value.go` — `NormalizedYAMLValue`

`NormalizedRulesValue` follows the same shape: embed `jsontypes.Normalized`, implement `attr.Type` via a companion `NormalizedRulesType`, override `StringSemanticEquals`.

## Goals / Non-Goals

**Goals:**

- Fix the perpetual-diff bug for `rules` attributes containing single-element arrays.
- Preserve the user's config value verbatim in `terraform plan` output (no plan-value mutation).
- Handle both directions: array-in-config + string-in-state (transition case) and array-in-config + array-in-state (steady state after fix).
- Remove the now-redundant `normalizeRuleNode` call from the read path.

**Non-goals:**

- Changing Elasticsearch server behavior (string normalization is an ES API characteristic).
- Fixing the data source drift (data source `rules` is computed-only; no plan comparison occurs — no change needed).
- State migration (no schema version bump required; `StringSemanticEquals` handles the shape mismatch transparently).
- Extracting the rule DSL normalization to a shared package (can be done later; keep local per operator direction).
- Changing behavior for multi-element arrays (already stored correctly by ES and provider).

## Decisions

- **Custom type, not plan modifier**: `planmodifier.String` (Approach A) would silently rewrite the user's config to string form in plan output, which is surprising. `StringSemanticEquals` is the idiomatic Plugin Framework mechanism for "two representations are logically identical" and leaves the plan value untouched.

- **Remove `normalizeRuleNode` from read path**: Per operator direction ("Simplify by removing it"). The typed client already returns array form; state will store array form. Semantic equality handles the transition for users with existing string-form state.

- **Keep local to `rolemapping` package**: Per operator direction ("Keep it local for now, we can move it easily"). `normalizeRuleNode` and related helpers move from `read.go` to `rules_value.go`.

- **`normalizeRoleNode` collapsing logic**: The `normalizeRuleNode` function remains as the normalization primitive inside `rules_value.go`, but is no longer called from the read path. `normalizeRoleMappingRules` (the marshal+walk+unmarshal wrapper) can be removed from `read.go`; the custom type uses a string-in / string-out variant.

## Implementation Sketch

### `rules_value.go` (new)

```go
package rolemapping

import (
    "context"
    "encoding/json"

    "github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
    "github.com/hashicorp/terraform-plugin-framework/attr"
    "github.com/hashicorp/terraform-plugin-framework/diag"
    "github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// NormalizedRulesType is the attr.Type companion for NormalizedRulesValue.
type NormalizedRulesType struct{ jsontypes.NormalizedType }

func (t NormalizedRulesType) ValueFromString(_ context.Context, v basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
    return NormalizedRulesValue{Normalized: jsontypes.NewNormalizedValue(v.ValueString())}, nil
}

// NormalizedRulesValue is a jsontypes.Normalized subtype that treats
// single-element arrays and plain strings as semantically equal inside
// "field" rule objects to handle the ES normalization behavior.
type NormalizedRulesValue struct{ jsontypes.Normalized }

func (v NormalizedRulesValue) Type(_ context.Context) attr.Type { return NormalizedRulesType{} }

func (v NormalizedRulesValue) StringSemanticEquals(ctx context.Context, other basetypes.StringValuable) (bool, diag.Diagnostics) {
    otherRules, ok := other.(NormalizedRulesValue)
    if !ok {
        return v.Normalized.StringSemanticEquals(ctx, other)
    }
    if v.IsNull() || v.IsUnknown() || otherRules.IsNull() || otherRules.IsUnknown() {
        return v.Normalized.StringSemanticEquals(ctx, otherRules.Normalized)
    }
    thisNorm, err1 := normalizeRulesJSONString(v.ValueString())
    thatNorm, err2 := normalizeRulesJSONString(otherRules.ValueString())
    if err1 != nil || err2 != nil {
        return v.Normalized.StringSemanticEquals(ctx, otherRules.Normalized)
    }
    return jsontypes.NewNormalizedValue(thisNorm).StringSemanticEquals(ctx, jsontypes.NewNormalizedValue(thatNorm))
}

func NewNormalizedRulesValue(v string) NormalizedRulesValue {
    return NormalizedRulesValue{Normalized: jsontypes.NewNormalizedValue(v)}
}

func NewNormalizedRulesNull() NormalizedRulesValue {
    return NormalizedRulesValue{Normalized: jsontypes.NewNormalizedNull()}
}

// normalizeRulesJSONString parses a JSON string and collapses single-element
// arrays inside "field" objects to plain string values.
func normalizeRulesJSONString(raw string) (string, error) {
    var tree map[string]any
    if err := json.Unmarshal([]byte(raw), &tree); err != nil {
        return "", err
    }
    normalizeRuleNode(tree)
    out, err := json.Marshal(tree)
    if err != nil {
        return "", err
    }
    return string(out), nil
}

// normalizeRuleNode walks a parsed JSON rule tree and collapses
// single-element arrays inside "field" objects to plain string values.
func normalizeRuleNode(node any) {
    switch v := node.(type) {
    case map[string]any:
        if field, ok := v["field"]; ok {
            if fieldMap, ok := field.(map[string]any); ok {
                for key, val := range fieldMap {
                    if arr, ok := val.([]any); ok && len(arr) == 1 {
                        fieldMap[key] = arr[0]
                    }
                }
            }
        }
        for _, child := range v {
            normalizeRuleNode(child)
        }
    case []any:
        for _, child := range v {
            normalizeRuleNode(child)
        }
    }
}
```

### `read.go` changes

- Remove `normalizeRuleNode` and `normalizeRoleMappingRules` functions.
- In `readRoleMapping`, replace:
  ```go
  rulesJSON, err := normalizeRoleMappingRules(roleMapping.Rules)
  // ...
  data.Rules = jsontypes.NewNormalizedValue(rulesJSON)
  ```
  with:
  ```go
  rulesJSON, err := json.Marshal(roleMapping.Rules)
  // ...
  data.Rules = NewNormalizedRulesValue(string(rulesJSON))
  ```

### `models.go` changes

```go
Rules NormalizedRulesValue `tfsdk:"rules"`
```

### `schema.go` changes

```go
"rules": schema.StringAttribute{
    // ...
    CustomType: NormalizedRulesType{},
},
```

## Risks / Trade-offs

- **Transition for existing state**: Users with string-form state (`"project1"`) will see no diff on plan (semantic equality normalizes to string form on both sides). After the first apply, state refreshes to array form. Subsequent plans compare array vs. array — always equal. Transparent, no manual action needed.
- **Data source**: No change needed — `rules` is computed-only on the data source, so there is no plan comparison.
- **`normalizeRulesType{}` interface compliance**: The `NormalizedRulesType` must implement enough of `attr.Type` for Plugin Framework to use it; embedding `jsontypes.NormalizedType` delegates all other methods.

## Open Questions

- Should the `normalizeRuleNode` collapsing behavior on read be **removed** now that semantic equality handles the discrepancy? **Answered**: Remove it (per operator direction).
- Does the ES API guarantee single-element arrays are always stored as strings, or is this version-dependent behavior? **Answered**: Non-issue either way (per operator direction).
- Should the fix be gated to the `rolemapping`-local package, or does the rule DSL normalization logic belong in a shared utility that other potential rule-bearing resources could reuse? **Answered**: Keep local for now (per operator direction).
