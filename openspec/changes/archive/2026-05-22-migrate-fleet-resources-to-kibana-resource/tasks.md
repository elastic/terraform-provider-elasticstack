## 1. Migrate `fleet/proxy`

- [x] 1.1 Add `KibanaResourceModel` interface methods to `proxyModel` (`GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`)
- [x] 1.2 Add `GetVersionRequirements()` to `proxyModel` implementing `entitycore.WithVersionRequirements` (returns min version 8.7.1 with error message matching current `assertVersionSupported`)
- [x] 1.3 Remove `kibana_connection` block from `getSchema()` (envelope injects it)
- [x] 1.4 Convert `Create` receiver method to package-level `KibanaCreateFunc[proxyModel]` callback
- [x] 1.5 Convert `Read` receiver method to package-level `kibanaReadFunc[proxyModel]` callback
- [x] 1.6 Convert `Update` receiver method to package-level `KibanaUpdateFunc[proxyModel]` callback
- [x] 1.7 Convert `Delete` receiver method to package-level `kibanaDeleteFunc[proxyModel]` callback
- [x] 1.8 Rewrite `resource.go`: embed `*entitycore.KibanaResource[proxyModel]`, wire `NewKibanaResource` with all four callbacks, retain `ImportState` concrete method
- [x] 1.9 Delete `version.go`
- [x] 1.10 Verify `make build` passes and unit tests in the package pass

## 2. Migrate `fleet/agentdownloadsource`

- [x] 2.1 Add `GetVersionRequirements()` to `model` implementing `entitycore.WithVersionRequirements` (returns min version 8.13.0 with error message matching current `assertVersionSupported`)
- [x] 2.2 Add `KibanaResourceModel` interface methods to `model`: `GetID` (returns `m.ID`), `GetResourceID` (returns `m.SourceID`), `GetKibanaConnection` (returns `m.KibanaConnection`), and `GetSpaceID` (returns first element of `SpaceIDs` or `"default"` when null/empty/unknown)
- [x] 2.3 Remove `kibana_connection` block from the schema function
- [x] 2.4 Convert `Read` receiver method to a package-level `kibanaReadFunc[model]` callback (wraps `readAndHydrateState`; callback receives the prior model to preserve `SpaceIDs` and `KibanaConnection`)
- [x] 2.5 Convert `Delete` receiver method to a package-level `kibanaDeleteFunc[model]` callback (uses `resourceID` and `spaceID` from envelope arguments)
- [x] 2.6 Update `Create` receiver method: replace `assertVersionSupported(ctx, apiClient)` with `entitycore.EnforceVersionRequirements(ctx, apiClient, &plan)`
- [x] 2.7 Update `Update` receiver method: replace `assertVersionSupported(ctx, apiClient)` with `entitycore.EnforceVersionRequirements(ctx, apiClient, &plan)`
- [x] 2.8 Rewrite `resource.go`: embed `*entitycore.KibanaResource[model]`, wire `NewKibanaResource` with read/delete callbacks and `PlaceholderKibanaWriteCallbacks` for create/update; update `SpaceImporter` to set both `path.Root("source_id")` and `path.Root("id")`
- [x] 2.9 Delete `version.go`
- [x] 2.10 Update `entitycore_contract_test.go`: replace `TestResource_embedsEntityCoreResourceBase` with a check that `Resource` embeds `*entitycore.KibanaResource[model]`; add assertion that `id` is set after import in the import state test
- [x] 2.11 Verify `make build` passes and unit tests in the package pass

## 3. Final verification

- [x] 3.1 Run `make check-lint` (includes `make check-openspec`) and resolve any issues
- [x] 3.2 Confirm neither package imports `assertVersionSupported` or the `version` file references remain
