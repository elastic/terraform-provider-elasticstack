## 1. Custom set type

- [ ] 1.1 Create `internal/elasticsearch/ml/datafeed/expand_wildcards_type.go`. Define `ExpandWildcardsType` implementing `basetypes.SetTypable` (embed `basetypes.SetType{ElemType: types.StringType}`). Define `ExpandWildcardsValue` implementing `basetypes.SetValuableWithSemanticEquals` (embed `basetypes.SetValue`). Add constructor helpers: `NewExpandWildcardsNull()`, `NewExpandWildcardsUnknown()`, `NewExpandWildcardsValue(elements []attr.Value) (ExpandWildcardsValue, diag.Diagnostics)`.
- [ ] 1.2 Implement `SetSemanticEquals` on `ExpandWildcardsValue`. The method MUST: handle null/unknown conservatively; normalize both sides by expanding `"all"` → `{"open","closed","hidden"}`; compare normalized sets for equality; leave `"none"` and all other tokens as literals. Return `(bool, diag.Diagnostics)` per the Plugin Framework contract.
- [ ] 1.3 Add required interface assertions in the same file: `var _ basetypes.SetTypable = (*ExpandWildcardsType)(nil)` and `var _ basetypes.SetValuableWithSemanticEquals = (*ExpandWildcardsValue)(nil)`.
- [ ] 1.4 Create `internal/elasticsearch/ml/datafeed/expand_wildcards_type_test.go` with unit tests covering:
  - `{"all"}` == `{"open","closed","hidden"}` (both directions)
  - `{"open","closed","hidden"}` == `{"hidden","open","closed"}` (order insensitivity)
  - `{"all"}` != `{"open","closed"}` (partial expansion not equal)
  - `{"none"}` == `{"none"}`, `{"none"}` != `{"open"}`
  - null == null, unknown == unknown
  - null != non-null, unknown != known

## 2. Schema update

- [ ] 2.1 In `internal/elasticsearch/ml/datafeed/schema.go`, change `indices_options.expand_wildcards` from `schema.ListAttribute` to `schema.SetAttribute`. Replace `ElementType: types.StringType` with `CustomType: ExpandWildcardsType{SetType: basetypes.SetType{ElemType: types.StringType}}`. Remove `listplanmodifier.UseStateForUnknown()` and `listvalidator.*` imports if no longer needed. The `stringvalidator.OneOf` constraint on values remains (applied via `setvalidator.ValueStringsAre`).
- [ ] 2.2 Remove unused imports from `schema.go` (`listplanmodifier`, `listvalidator`) if they are no longer referenced.

## 3. Model update

- [ ] 3.1 In `internal/elasticsearch/ml/datafeed/models.go`, change `IndicesOptions.ExpandWildcards` from `types.List` to `ExpandWildcardsValue`.
- [ ] 3.2 In `ToAPIModel` (models.go), replace `indicesOptions.ExpandWildcards.ElementsAs(ctx, &expandWildcards, false)` with the equivalent call on `ExpandWildcardsValue` (it embeds `basetypes.SetValue` so `ElementsAs` is available directly; no cast needed).
- [ ] 3.3 In `FromAPIModel` (models.go), replace `types.ListValueFrom(ctx, types.StringType, apiModel.IndicesOptions.ExpandWildcards)` with element construction into `ExpandWildcardsValue` using `NewExpandWildcardsValue`. When the API returns an empty slice, use `NewExpandWildcardsNull()`.
- [ ] 3.4 In `GetIndicesOptionsAttrTypes()` (schema.go) and both uses of `map[string]attr.Type{...}` in `FromAPIModel` that reference `"expand_wildcards"`, replace `types.ListType{ElemType: types.StringType}` with `ExpandWildcardsType{SetType: basetypes.SetType{ElemType: types.StringType}}`.
- [ ] 3.5 Update `types.ObjectValueFrom` / `types.ObjectNull` call sites in `FromAPIModel` that build `IndicesOptions` objects to use the updated attribute type map.

## 4. Acceptance test update

- [ ] 4.1 In `internal/elasticsearch/ml/datafeed/acc_test.go`, replace any `resource.TestCheckResourceAttr("…", "indices_options.expand_wildcards.0", …)` style index-based assertions with `resource.TestCheckTypeSetElemAttr("…", "indices_options.expand_wildcards.*", …)`.
- [ ] 4.2 If any test fixture HCL sets `expand_wildcards = ["all"]`, add or update a test step that asserts the normalized element set (`"open"`, `"closed"`, `"hidden"`) is present without a perpetual diff. The fixture value `["all"]` must be preserved as written.
- [ ] 4.3 Verify state upgrader is not required: run `TestAccResourceDatafeed` with a state snapshot written under the old list schema. If a diagnostic error occurs during state decode, proceed to add a state upgrader (see task 5).

## 5. State upgrader (conditional)

- [ ] 5.1 **Only if task 4.3 reveals a state decode error:** add `StateUpgraders` on the datafeed resource with a schema version bump from 0 to 1. The upgrader reads `indices_options.expand_wildcards` as a list and writes it back as a set. Update `.SchemaVersion` on the resource.

## 6. Spec update (delta)

- [ ] 6.1 The delta spec at `openspec/changes/datafeed-expand-wildcards-set-type/specs/elasticsearch-ml-datafeed/spec.md` (created alongside this task file) documents the updated schema and plan behavior. Review it for accuracy before implementation begins and update it if design decisions change.

## 7. Verification

- [ ] 7.1 `make build` passes with no compilation errors.
- [ ] 7.2 `go test ./internal/elasticsearch/ml/datafeed/...` (unit tests) passes.
- [ ] 7.3 `make check-lint` passes.
- [ ] 7.4 `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate datafeed-expand-wildcards-set-type --type change` passes.
- [ ] 7.5 If an Elasticsearch stack is available: run `TF_ACC=1 go test ./internal/elasticsearch/ml/datafeed/... -run TestAccResourceDatafeed -v` and confirm no perpetual diff for `expand_wildcards = ["all"]` configurations.
