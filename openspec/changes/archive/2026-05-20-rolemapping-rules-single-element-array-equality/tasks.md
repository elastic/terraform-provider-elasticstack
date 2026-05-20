## 1. Custom type — `rules_value.go`

- [x] 1.1 Create `internal/elasticsearch/security/rolemapping/rules_value.go` with:
  - `NormalizedRulesType` struct embedding `jsontypes.NormalizedType`; implement `ValueFromString` to return a `NormalizedRulesValue`.
  - `NormalizedRulesValue` struct embedding `jsontypes.Normalized`; implement `Type()` returning `NormalizedRulesType{}`.
  - `StringSemanticEquals`: cast `other` to `NormalizedRulesValue`; normalize both sides via `normalizeRulesJSONString`; fall back to `jsontypes.Normalized.StringSemanticEquals` on null/unknown or parse error.
  - `NewNormalizedRulesValue(v string) NormalizedRulesValue` and `NewNormalizedRulesNull() NormalizedRulesValue` constructors.
  - `normalizeRulesJSONString(raw string) (string, error)`: unmarshal JSON string → apply `normalizeRuleNode` → marshal back.
  - `normalizeRuleNode(node any)`: the walk function (moved from `read.go`) that collapses single-element arrays inside `"field"` objects to plain string values.

## 2. Read path — `read.go`

- [x] 2.1 Remove `normalizeRuleNode` and `normalizeRoleMappingRules` functions from `read.go` (both are now in `rules_value.go`).
- [x] 2.2 In `readRoleMapping`, replace the `normalizeRoleMappingRules(roleMapping.Rules)` call with a plain `json.Marshal(roleMapping.Rules)`. Use `NewNormalizedRulesValue(string(rulesJSON))` to set `data.Rules`.
- [x] 2.3 Remove the now-unused `normalizeRuleNode` walk from `read.go` (already covered by 2.1). Verify no other callers exist in the package.

## 3. Model — `models.go`

- [x] 3.1 Change the `Rules` field type from `jsontypes.Normalized` to `NormalizedRulesValue`:
  ```go
  Rules NormalizedRulesValue `tfsdk:"rules"`
  ```

## 4. Schema — `schema.go`

- [x] 4.1 Change the `CustomType` on the `rules` attribute from `jsontypes.NormalizedType{}` to `NormalizedRulesType{}`:
  ```go
  "rules": schema.StringAttribute{
      // ...
      CustomType: NormalizedRulesType{},
  },
  ```
- [x] 4.2 Remove the `jsontypes` import from `schema.go` if it is no longer referenced; add no new import (the type lives in the same package).

## 5. Unit tests — `rules_value_test.go`

- [x] 5.1 Create `internal/elasticsearch/security/rolemapping/rules_value_test.go` with table-driven tests for `NormalizedRulesValue.StringSemanticEquals`:
  - Single-element array vs. plain string: `{"field":{"groups":["project1"]}}` == `{"field":{"groups":"project1"}}` → **equal**.
  - Both array form: `{"field":{"groups":["project1"]}}` == `{"field":{"groups":["project1"]}}` → **equal**.
  - Both string form: `{"field":{"groups":"project1"}}` == `{"field":{"groups":"project1"}}` → **equal**.
  - Multi-element array vs. different value: `{"field":{"groups":["a","b"]}}` != `{"field":{"groups":["a"]}}` → **not equal**.
  - Null vs. null → **equal**; null vs. non-null → **not equal**.
  - Whitespace difference only: `{"field":{"groups":"x"}}` == `{ "field": { "groups": "x" } }` → **equal** (inherited from `jsontypes.Normalized`).

## 6. Build verification

- [x] 6.1 Run `make build` to confirm the provider compiles without errors.
- [x] 6.2 Run `go test ./internal/elasticsearch/security/rolemapping/...` to confirm the unit tests pass.

## 7. OpenSpec

- [x] 7.1 Keep the delta spec `openspec/changes/rolemapping-rules-single-element-array-equality/specs/elasticsearch-security-role-mapping/spec.md` aligned with implementation.
- [ ] 7.2 After merge: sync into `openspec/specs/elasticsearch-security-role-mapping/spec.md` or archive the change per project workflow; run `make check-openspec`.
