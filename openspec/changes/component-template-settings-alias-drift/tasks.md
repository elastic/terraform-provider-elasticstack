## 1. Extract alias helpers into `aliasutil`

- [x] 1.1 Move `alias_type.go` from `internal/elasticsearch/index/template/` into `internal/elasticsearch/index/aliasutil/` — export `AliasObjectType`, `AliasObjectValue`, `NewAliasObjectValue`, `ObjectSemanticEquals`, `NewAliasObjectType`, `aliasObjectFromAttr`, `aliasElementModelsSemanticallyEqual`, `fillUnknownAliasModelFieldsFromOther` (capitalise as needed for export)
- [x] 1.2 Move `alias_canonicalize.go` into `aliasutil/` — export `CanonicalizeAliasObjectForState`, `CanonicalizeAliasSetElements`, `CanonicalizeTemplateAliasSetInModel` (or equivalent names)
- [x] 1.3 Move `alias_reconcile.go` into `aliasutil/` — export `MergePlanAliasSetWithPriorState`, `ProjectConfigAliasMatchesOntoPlan`, `MergeAliasSetPreferReferenceEncoding`, `ApplyTemplateAliasReconciliationFromReference`
- [x] 1.4 Update all existing callers in `internal/elasticsearch/index/template/` to reference the exported `aliasutil.*` symbols; remove or thin the original files (keep only thin shims if needed to avoid circular imports, otherwise delete)
- [x] 1.5 Ensure existing `alias_type_test.go` (if it references package-level symbols) still compiles; move or update as needed

## 2. Extract settings reconciliation into `templateutil`

- [x] 2.1 Add `ReconcileSettingsIfSemanticallyEqual` (or equivalent) to `internal/elasticsearch/index/templateutil/` — takes `(ctx, planSettings, stateSettings customtypes.IndexSettingsValue)` and returns the canonical value to use in the plan plus a `changed bool` plus diagnostics
- [x] 2.2 Update `internal/elasticsearch/index/template/modify_plan.go` to call the shared helper instead of the inline logic; remove duplicated code

## 3. Adopt `aliasutil.AliasObjectType` in `component_template` schema

- [ ] 3.1 In `internal/elasticsearch/index/componenttemplate/schema.go`, change the `alias` set nested block's element object type from plain `types.ObjectType{AttrTypes: aliasAttrTypes()}` to `aliasutil.AliasObjectType` (or `aliasutil.NewAliasObjectType()`)
- [ ] 3.2 Verify `AttrTypes` map is identical between the old type and `aliasutil.AliasObjectType` — if any field differs, reconcile before proceeding

## 4. Add read-time alias reconciliation to `component_template`

- [ ] 4.1 In `internal/elasticsearch/index/componenttemplate/read.go` (or `flatten.go`), after `flattenToData`, call `aliasutil.ApplyTemplateAliasReconciliationFromReference` + `aliasutil.CanonicalizeTemplateAliasSetInModel` using the prior state's alias set as the reference — mirroring `internal/elasticsearch/index/template/read.go:47-54`
- [ ] 4.2 Confirm `extractAliasRoutingFromData` is still used and the routing-preservation logic in `flattenTemplateBlock` remains compatible

## 5. Add `ModifyPlan` to `component_template`

- [ ] 5.1 Create `internal/elasticsearch/index/componenttemplate/modify_plan.go`
- [ ] 5.2 Implement `func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse)` — guard on null plan/state (create/destroy paths), load plan/state/config into `Data` models
- [ ] 5.3 Call `templateutil.ReconcileSettingsIfSemanticallyEqual` for `template.settings`; if the settings differ only in encoding, replace plan settings with the state's canonical value
- [ ] 5.4 Call `aliasutil.MergePlanAliasSetWithPriorState` + `aliasutil.ProjectConfigAliasMatchesOntoPlan` + `aliasutil.CanonicalizeAliasSetElements` for `template.alias` — mirroring the index_template `ModifyPlan` alias section
- [ ] 5.5 Register `resource.ResourceWithModifyPlan` in the `var _ ...` block in `internal/elasticsearch/index/componenttemplate/resource.go`

## 6. Spec sync

- [ ] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec sync component-template-settings-alias-drift` (or equivalent) to merge REQ-037 and REQ-038 from the delta spec into `openspec/specs/elasticsearch-index-component-template/spec.md`

## 7. Tests

- [ ] 7.1 Confirm whether `TestAccResourceComponentTemplateDottedSettingsNoDrift` already exists in `internal/elasticsearch/index/componenttemplate/acc_drift_test.go`; if so keep it as the regression guard for REQ-037; if not, write it: apply with dotted-key settings → plan-only step expecting no diff → import step with `ImportStateVerify: true`
- [ ] 7.2 Add `TestAccResourceComponentTemplateAliasRoutingNoDrift` (or similar): apply a component template with a `routing`-only alias, verify no `Provider produced inconsistent result after apply` and no perpetual diff on subsequent plan (guards REQ-038)
- [ ] 7.3 Add unit tests for new `aliasutil` exported helpers, porting the existing `TestReconcilePlanWithPriorStateForSemanticDrift_settingsNestedPlanDottedState` from `internal/elasticsearch/index/template/modify_plan_test.go` to exercise both the shared `templateutil` helper and the component_template `ModifyPlan`
- [ ] 7.4 Run `make build` and `go vet ./...` to confirm compilation; run `go test ./internal/elasticsearch/index/...` (unit tests only, no `TF_ACC`) to confirm no regressions
- [ ] 7.5 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate component-template-settings-alias-drift --type change` and resolve any reported errors

## 8. Pre-merge checklist

- [ ] 8.1 State migration risk: if an old provider binary cannot read state written by the new binary (due to custom type wrapper on alias elements), bump `SchemaVersion` and add a state upgrader for the alias element type change; otherwise confirm no migration needed
- [ ] 8.2 `templateilmattachment` sanity check: read `internal/elasticsearch/index/templateilmattachment/` code and confirm the new `ModifyPlan` path does not interact with ILM attachment operations
- [ ] 8.3 Run `make check-openspec` (includes linting of spec files)
