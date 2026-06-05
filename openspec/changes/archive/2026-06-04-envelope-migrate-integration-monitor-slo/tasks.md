## 1. fleet/integration migration

- [x] 1.1 Add `KibanaResourceModel` interface methods to `integrationModel` (`GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`)
- [x] 1.2 Add `IsUnscopedSpace()` to `integrationModel` returning true when SpaceID is null or unknown
- [x] 1.3 Replace `*entitycore.ResourceBase` embed with `*entitycore.KibanaResource[integrationModel]` in `resource.go`
- [x] 1.4 Remove explicit `kibana_connection` schema block from `schema.go` (envelope injects it)
- [x] 1.5 Rewrite `create.go` as a `KibanaWriteFunc[integrationModel]` (install logic + set ID/SpaceID, return `KibanaWriteResult[T]`)
- [x] 1.6 Rewrite `update.go` as a `KibanaWriteFunc[integrationModel]` (delegate to same install logic as create)
- [x] 1.7 Rewrite `read.go` as a `readIntegration(ctx, client, resourceID, spaceID, model) (integrationModel, bool, diag.Diagnostics)` callback
- [x] 1.8 Rewrite `delete.go` as a `deleteIntegration(ctx, client, resourceID, spaceID, model) diag.Diagnostics` callback
- [x] 1.9 Wire all callbacks into `entitycore.NewKibanaResource[integrationModel]` in `resource.go`
- [x] 1.10 Add `entitycore_contract_test.go` asserting `*entitycore.KibanaResource[integrationModel]` embedding
- [x] 1.11 Verify `make build` passes; run fleet/integration acceptance tests

## 2. kibana/synthetics/monitor migration

- [x] 2.1 Add `KibanaResourceModel` interface methods to `tfModelV0` (`GetID` returning composite, `GetResourceID` parsing composite to return monitorID, `GetSpaceID`, `GetKibanaConnection`)
- [x] 2.2 Remove `_ synthetics.ESAPIClient = newResource()` compile-time assertion from `resource.go`
- [x] 2.3 Remove `GetClient()` method from the concrete `Resource` type
- [x] 2.4 Delete `synthetics.ESAPIClient` interface and `synthetics.GetKibanaOAPIClient(ESAPIClient, dg)` from `synthetics/api_client.go`
- [x] 2.5 Replace `*entitycore.ResourceBase` embed with `*entitycore.KibanaResource[tfModelV0]` in `resource.go`
- [x] 2.6 Rewrite `create.go` as a `KibanaWriteFunc[tfModelV0]` (call `CreateMonitor`, populate model via `toModelV0`, return `KibanaWriteResult[T]`)
- [x] 2.7 Rewrite `update.go` as a `KibanaWriteFunc[tfModelV0]` (call `UpdateMonitor`, populate model via `toModelV0`, return `KibanaWriteResult[T]`)
- [x] 2.8 Rewrite `read.go` as a `readMonitor(ctx, client, resourceID, spaceID, model) (tfModelV0, bool, diag.Diagnostics)` callback (receives separate resourceID and spaceID; no composite parsing needed)
- [x] 2.9 Rewrite `delete.go` as a `deleteMonitor(ctx, client, resourceID, spaceID, model) diag.Diagnostics` callback
- [x] 2.10 Wire all callbacks into `entitycore.NewKibanaResource[tfModelV0]` in `resource.go`
- [x] 2.11 Add `entitycore_contract_test.go` asserting `*entitycore.KibanaResource[tfModelV0]` embedding
- [x] 2.12 Verify `make build` passes; run synthetics/monitor acceptance tests

## 3. kibana/slo migration

- [x] 3.1 Add `KibanaResourceModel` interface methods to `tfModel` (`GetID`, `GetResourceID` parsing composite, `GetSpaceID`, `GetKibanaConnection`)
- [x] 3.2 Promote `readSloFromAPI` from a `*Resource` method to a package-level function; update its signature to `readSloFromAPI(ctx, apiClient, resourceID, spaceID string, model *tfModel) (bool, diag.Diagnostics)` — removing composite-ID parsing from inside the function since the read callback receives `resourceID`/`spaceID` from the envelope directly
- [x] 3.3 Promote `readAndPopulate` from a `*Resource` method to a package-level function `readSloAndPopulate(ctx, apiClient, model *tfModel, diags *diag.Diagnostics)`; update it to call the promoted `readSloFromAPI`, reconstructing `resourceID`/`spaceID` from `model.ID`
- [x] 3.4 Replace `*entitycore.ResourceBase` embed with `*entitycore.KibanaResource[tfModel]` in `resource.go`; retain `ConfigValidators` and `UpgradeState` on the concrete type; remove explicit `kibana_connection` schema block (envelope injects it)
- [x] 3.5 Rewrite `create.go` as a `KibanaWriteFunc[tfModel]`: convert `req.Plan` → `apiModel` → call `resolveGroupBySupport(ctx, client, &apiModel, &diags)` (note: `EnforceVersionRequirements` is handled by the envelope since `tfModel` implements `WithVersionRequirements`) → call `CreateSlo` → set `model.ID` (composite) → call `readSloAndPopulate` (intermediate read: create response only returns `id`, not `enabled`) → if `planEnabled != serverEnabled`: call Enable/Disable API → call `readSloFromAPI` → return `KibanaWriteResult[T]{Model: model}`
- [x] 3.6 Rewrite `update.go` as a `KibanaWriteFunc[tfModel]`: convert `req.Plan` → `apiModel` → call `resolveGroupBySupport(ctx, client, &apiModel, &diags)` → call `UpdateSlo` → call `readSloAndPopulate` → reconcile enabled if needed → return `KibanaWriteResult[T]{Model: model}`
- [x] 3.7 Rewrite `read.go` as a `readSlo(ctx, client, resourceID, spaceID, model) (tfModel, bool, diag.Diagnostics)` callback; reconstruct composite ID from envelope-provided `resourceID`/`spaceID`, then call promoted `readSloFromAPI`
- [x] 3.8 Rewrite `delete.go` as a `deleteSlo(ctx, client, resourceID, spaceID, model) diag.Diagnostics` callback
- [x] 3.9 Wire all callbacks into `entitycore.NewKibanaResource[tfModel]` in `resource.go`
- [x] 3.10 Remove now-dead `*Resource` methods: `reconcileSloEnabledAfterWrite` (logic moved into write callbacks), `readAndPopulate`, and `readSloFromAPI` (both promoted to package-level); also remove the now-redundant explicit `entitycore.EnforceVersionRequirements` calls from the old create/update (envelope handles these)
- [x] 3.11 Add `entitycore_contract_test.go` asserting `*entitycore.KibanaResource[tfModel]` embedding
- [x] 3.12 Verify `make build` passes; run slo acceptance tests

## 4. Final validation

- [x] 4.1 Run `make build` across the full provider (all three resources in one build)
- [x] 4.2 Confirm no remaining `ResourceBase`-only resources exist in `internal/kibana/` or `internal/fleet/` trees (within the scope of this change: `fleet/agentpolicy`, `fleet/elastic_defend_integration_policy`, `fleet/integration_policy`, and `kibana/import_saved_objects` remain as explicit non-goals or out-of-scope)
