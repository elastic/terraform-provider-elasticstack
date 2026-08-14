## 1. Implement `writeSecurityEnableRule` callback

- [ ] 1.1 In `update.go`, remove the `Update` wrapper method (`func (r *EnableRuleResource) Update(ctx, req, resp)`) that shadows the envelope's promoted method
- [ ] 1.2 In `update.go`, remove the `upsert` helper method (`func (r *EnableRuleResource) upsert(ctx, plan, state)`) and its imports
- [ ] 1.3 In `update.go`, add a package-level `writeSecurityEnableRule` function with signature `func writeSecurityEnableRule(ctx context.Context, client *clients.KibanaScopedClient, req entitycore.KibanaWriteRequest[enableRuleModel]) (entitycore.KibanaWriteResult[enableRuleModel], diag.Diagnostics)` that: preserves the `DisableOnDestroy` null-default, computes `model.ID`, calls `kibanaoapi.EnableRulesByTag`, sets `model.AllRulesEnabled = types.BoolValue(true)`, and returns `KibanaWriteResult{Model: model, SkipReadAfterWrite: true}`

## 2. Remove `Create` wrapper override

- [ ] 2.1 Delete `create.go` entirely — it contains only the `Create` wrapper method that shadows the envelope's promoted `Create`; the envelope will now dispatch through `writeSecurityEnableRule`

## 3. Clean up delete callback

- [ ] 3.1 In `delete.go`, remove the `entitycore.EnforceVersionRequirements` call and its `if diags.HasError() { return diags }` guard from `deleteSecurityEnableRule` — the envelope's `baseResourceEnvelope.Delete` already enforces version requirements before invoking the callback
- [ ] 3.2 In `delete.go`, remove the now-unused `entitycore` import

## 4. Wire callbacks in resource constructor

- [ ] 4.1 In `resource.go`, remove the `placeholder` variable and replace the `PlaceholderKibanaWriteCallback` usages with `writeSecurityEnableRule` for both `Create` and `Update` in `KibanaResourceOptions`
- [ ] 4.2 In `resource.go`, remove the now-unused `entitycore` import (only needed for the placeholder)

## 5. Build and verify

- [ ] 5.1 Run `make build` and confirm the package compiles without errors
- [ ] 5.2 Run `go test ./internal/kibana/security_enable_rule/...` and confirm all unit and contract tests pass (including `TestEnableRuleResource_embedsEntityCoreKibanaResource`)
