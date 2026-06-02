## 1. fleet/serverhost

- [x] 1.1 Add `GetID()`, `GetResourceID()` (→ `HostID`), `GetSpaceID()` (→ first of `SpaceIDs` or `""`), `GetKibanaConnection()` value-receiver methods to `serverHostModel`
- [x] 1.2 Add `IsUnscopedSpace() bool` returning `true` to `serverHostModel` (implements `entitycore.KibanaUnscopedSpace`)
- [x] 1.3 Remove `kibana_connection` block from `schema.go` schema function
- [x] 1.4 Extract `Create` method body to `func createServerHost(ctx, *KibanaScopedClient, KibanaWriteRequest[serverHostModel]) (KibanaWriteResult[serverHostModel], diag.Diagnostics)`
- [x] 1.5 Extract `Read` method body to `func readServerHost(ctx, *KibanaScopedClient, resourceID, spaceID string, model serverHostModel) (serverHostModel, bool, diag.Diagnostics)`
- [x] 1.6 Extract `Update` method body to `func updateServerHost(ctx, *KibanaScopedClient, KibanaWriteRequest[serverHostModel]) (KibanaWriteResult[serverHostModel], diag.Diagnostics)` — use `req.Prior.GetSpaceID()` as operational space for the API call
- [x] 1.7 Extract `Delete` method body to `func deleteServerHost(ctx, *KibanaScopedClient, resourceID, spaceID string, model serverHostModel) diag.Diagnostics` — retain the pre-delete clear-`IsDefault` logic inside the callback
- [x] 1.8 Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[serverHostModel]` in the resource struct and constructor; pass callbacks via `KibanaResourceOptions`; retain `*fleet.SpaceImporter` on the wrapper struct
- [x] 1.9 Add `entitycore_contract_test.go` asserting the resource embeds `KibanaResource[serverHostModel]` (follow pattern in `internal/fleet/agentdownloadsource/`)
- [x] 1.10 Run `make build` and `go test ./internal/fleet/serverhost/...`

## 2. fleet/output

- [ ] 2.1 Add `GetID()`, `GetResourceID()` (→ `OutputID`), `GetSpaceID()` (→ first of `SpaceIDs` or `""`), `GetKibanaConnection()` value-receiver methods to `outputModel`
- [ ] 2.2 Add `IsUnscopedSpace() bool` returning `true` to `outputModel`
- [ ] 2.3 Add `GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics)` to `outputModel`: emit `MinVersionOutputKafka` (8.13.0) when `Type == "kafka"`; emit `MinVersionOutputSSLVerificationMode` (8.10.0) when `Ssl` is known and `ssl.verification_mode` is set
- [ ] 2.4 Delete `assertKafkaSupport` and `assertSSLVerificationModeSupport` helpers from `models.go`; remove their call sites inside `toAPICreateModel`/`toAPIUpdateModel`/`buildCommonNewOutput`/`buildCommonUpdateOutput`; drop `*clients.KibanaScopedClient` parameters from these helpers if no other version-gated logic remains
- [ ] 2.5 Remove `kibana_connection` block from `schema.go` schema function
- [ ] 2.6 Extract `Create` method body to `func createOutput(ctx, *KibanaScopedClient, KibanaWriteRequest[outputModel]) (KibanaWriteResult[outputModel], diag.Diagnostics)`
- [ ] 2.7 Extract `Read` method body to `func readOutput(ctx, *KibanaScopedClient, resourceID, spaceID string, model outputModel) (outputModel, bool, diag.Diagnostics)`
- [ ] 2.8 Extract `Update` method body to `func updateOutput(ctx, *KibanaScopedClient, KibanaWriteRequest[outputModel]) (KibanaWriteResult[outputModel], diag.Diagnostics)` — use `req.Prior.GetSpaceID()` as operational space
- [ ] 2.9 Extract `Delete` method body to `func deleteOutput(ctx, *KibanaScopedClient, resourceID, spaceID string, model outputModel) diag.Diagnostics`
- [ ] 2.10 Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[outputModel]`; retain `*fleet.SpaceImporter` and `UpgradeState` on wrapper struct
- [ ] 2.11 Add `entitycore_contract_test.go`
- [ ] 2.12 Run `make build` and `go test ./internal/fleet/output/...`

## 3. fleet/customintegration

- [ ] 3.1 Add `GetID()`, `GetResourceID()` (→ `getPackageID(PackageName, PackageVersion)` when known, else `ID`), `GetSpaceID()` (→ `SpaceID`), `GetKibanaConnection()` value-receiver methods to `customIntegrationModel`
- [ ] 3.2 Add `GetVersionRequirements() ([]entitycore.VersionRequirement, diag.Diagnostics)` to `customIntegrationModel`: always emit `minVersionCustomPackageGet` (8.2.0) with the existing error message
- [ ] 3.3 Delete inline `EnforceMinVersion(ctx, minVersionCustomPackageGet)` checks from `create.go`, `read.go`, and `update.go`
- [ ] 3.4 Remove `kibana_connection` block from `schema.go` schema function
- [ ] 3.5 Extract `Create` method body to `func createCustomIntegration(ctx, *KibanaScopedClient, KibanaWriteRequest[customIntegrationModel]) (KibanaWriteResult[customIntegrationModel], diag.Diagnostics)`
- [ ] 3.6 Extract `Read` method body to `func readCustomIntegration(ctx, *KibanaScopedClient, resourceID, spaceID string, model customIntegrationModel) (customIntegrationModel, bool, diag.Diagnostics)` — read callback uses `model.PackageName`/`model.PackageVersion` directly, ignores `resourceID` parameter
- [ ] 3.7 Extract `Update` method body to `func updateCustomIntegration(ctx, *KibanaScopedClient, KibanaWriteRequest[customIntegrationModel]) (KibanaWriteResult[customIntegrationModel], diag.Diagnostics)`
- [ ] 3.8 Extract `Delete` method body to `func deleteCustomIntegration(ctx, *KibanaScopedClient, resourceID, spaceID string, model customIntegrationModel) diag.Diagnostics`
- [ ] 3.9 Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[customIntegrationModel]`; retain `ModifyPlan` on wrapper struct; no `SpaceImporter` (no ImportState support)
- [ ] 3.10 Update existing `entitycore_contract_test.go` to assert `KibanaResource[customIntegrationModel]` embed (replacing current `ResourceBase` assertion)
- [ ] 3.11 Run `make build` and `go test ./internal/fleet/customintegration/...`

## 4. Final validation

- [ ] 4.1 Run `make build` across the full provider
- [ ] 4.2 Run `make lint` (or `make check-lint`) and fix any issues
- [ ] 4.3 Verify no schema changes: confirm `make docs` produces no diff for the three affected resources
