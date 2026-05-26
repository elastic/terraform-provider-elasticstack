## Why

The `elasticstack_elasticsearch_security_role_mapping` resource shows a perpetual diff on every `plan` and `apply` when the `rules` JSON contains a single-element array for a `field` value. For example:

```hcl
rules = jsonencode({
  field = { groups = ["project1"] }
})
```

Elasticsearch normalizes single-element arrays to strings on storage (`{"groups":"project1"}`). The typed client reverses this during unmarshal (it always decodes field values as `[]string`). The existing read path calls `normalizeRuleNode` to re-collapse them to strings so state matches ES storage. This means state holds `"project1"` (string) while the user's config still encodes `["project1"]` (array).

Because `rules` uses `jsontypes.Normalized` — which does JSON semantic equality (whitespace and key-order insensitive) but treats `["x"]` and `"x"` as distinct — Terraform detects a diff on every run. The resource is perpetually dirty even though the logical value is unchanged.

Multi-element arrays (`["a","b"]`) are stored as-is by ES and are unaffected.

## What Changes

- **New custom type** `NormalizedRulesValue` / `NormalizedRulesType` (new file `rules_value.go`) in the `rolemapping` package. Overrides `StringSemanticEquals` to normalize both sides via the existing single-element-array collapsing logic before comparing; plan and state values are never mutated.
- **Read path simplified**: remove the `normalizeRuleNode` collapsing call from `readRoleMapping` in `read.go`. State will now store whatever the typed client returns (array form). Semantic equality handles both the new case (array vs. array) and the transition case (string-form state from earlier provider versions vs. array-form config).
- **`models.go`**: change the `Rules` field type from `jsontypes.Normalized` to `NormalizedRulesValue`.
- **`schema.go`**: change the `CustomType` on the `rules` attribute from `jsontypes.NormalizedType{}` to `NormalizedRulesType{}`.
- **Unit tests** (`rules_value_test.go`): cover semantic equality between string and array forms, multi-element arrays, null/unknown handling.

No state migration is required; the custom type's `StringSemanticEquals` handles any string/array combination already in state.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`elasticsearch-security-role-mapping`**: `rules` attribute custom type upgraded to `NormalizedRulesValue` providing single-element-array semantic equality; read path simplified by removing the collapsing normalizer.

## Impact

- **Users**: Perpetual-diff bug on single-element `groups` (and any other single-element array field rule) is fixed without any HCL changes. Users who worked around the bug via `jsondecode(length(var.list)==1 ? ... : ...)` can remove that workaround.
- **State**: Existing string-form state is handled transparently by semantic equality on the first plan after upgrade; subsequent reads will store array form.
- **Code**: `internal/elasticsearch/security/rolemapping/` only — fix is scoped to this package as directed.
- **Maintenance**: `normalizeRuleNode` moves to `rules_value.go`; `normalizeRoleMappingRules` removed from `read.go`.
