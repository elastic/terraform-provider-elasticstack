## 1. Convert kibanaoapi diagnostics to Plugin Framework

- [ ] 1.1 Convert `internal/clients/kibanaoapi/status.go`: change return type from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`, replace `diagutil.SDKErrorDiag` with `fwdiag.NewErrorDiagnostic`, replace `diagutil.SDKDiagsFromFramework(...)` with direct PF diagnostics
- [ ] 1.2 Convert `internal/clients/kibanaoapi/security_role.go`: change all three public functions (`GetSecurityRole`, `PutSecurityRole`, `DeleteSecurityRole`) to return `fwdiag.Diagnostics`; remove `sdkdiag` import
- [ ] 1.3 Convert `internal/clients/kibanaoapi/connector.go`: change `SearchConnectors` to return `fwdiag.Diagnostics`; remove `sdkdiag` import
- [ ] 1.4 Convert `internal/clients/kibanaoapi/spaces.go`: change `CreateSpace`, `UpdateSpace`, `DeleteSpace` to return `fwdiag.Diagnostics`; remove `sdkdiag` import
- [ ] 1.5 Verify `go build ./internal/clients/kibanaoapi/...` succeeds

## 2. Update callers of kibanaoapi functions

- [ ] 2.1 Update `internal/clients/kibana_scoped_client.go`: remove `diagutil.FrameworkDiagsFromSDK` wrapping around `kibanaoapi.GetKibanaStatus` calls in `ServerVersion`, `ServerFlavor`, `EnforceMinVersion`
- [ ] 2.2 Update `internal/kibana/security_role/read.go`: remove `diagutil.FrameworkDiagsFromSDK` wrapping around `kibanaoapi.GetSecurityRole`
- [ ] 2.3 Update `internal/kibana/security_role/create.go`: remove `diagutil.FrameworkDiagsFromSDK` wrapping around `kibanaoapi.PutSecurityRole`
- [ ] 2.4 Update `internal/kibana/security_role/update.go`: remove `diagutil.FrameworkDiagsFromSDK` wrapping around `kibanaoapi.PutSecurityRole`
- [ ] 2.5 Update `internal/kibana/security_role/delete.go`: remove `diagutil.FrameworkDiagsFromSDK` wrapping around `kibanaoapi.DeleteSecurityRole`
- [ ] 2.6 Search for and update any other PF consumers of kibanaoapi functions (`internal/kibana/spaces/`, `internal/kibana/connectors/`, etc.)
- [ ] 2.7 Verify `go build ./internal/kibana/... ./internal/clients/...` succeeds

## 3. Remove dead translation helpers

- [ ] 3.1 Remove `FrameworkDiagsFromSDK` from `internal/diagutil/translation.go`
- [ ] 3.2 Remove `SDKDiagsFromFramework` from `internal/diagutil/translation.go`
- [ ] 3.3 Remove `SDKErrorDiag` from `internal/diagutil/translation.go`
- [ ] 3.4 Remove `sdkdiag` import from `internal/diagutil/translation.go`
- [ ] 3.5 Verify `go build ./internal/diagutil/...` succeeds

## 4. Remove SDK provider and mux wiring

- [ ] 4.1 Delete `provider/provider.go`
- [ ] 4.2 Rewrite `provider/factory.go`: replace `ProtoV6ProviderServerFactory` with a direct PF server factory function that returns `func() tfprotov6.ProviderServer` using `providerserver.NewProtocol6(NewFrameworkProvider(version))`
- [ ] 4.3 Update `main.go` to call the new direct factory function
- [ ] 4.4 Run `go mod tidy` to remove unused `terraform-plugin-mux` dependency
- [ ] 4.5 Verify `go build` at project root succeeds

## 5. Remove dead SDK configuration code

- [ ] 5.1 Delete `internal/clients/config/sdk.go`
- [ ] 5.2 Remove `NewAPIClientFuncFromSDK` and `newAPIClientFromSDK` from `internal/clients/api_client.go`; remove `sdkdiag` and `schema` imports if unused
- [ ] 5.3 Remove `GetKibanaClientFromSDK`, `GetElasticsearchClientFromSDK`, `ConvertMetaToFactory` from `internal/clients/provider_client_factory.go`; remove `sdkdiag` and `schema` imports if unused
- [ ] 5.4 Remove `FromSDK` helper functions from `internal/clients/config/base.go`, `elasticsearch.go`, `kibana_oapi.go`, `fleet.go` if they are unused after step 5.1
- [ ] 5.5 Verify `go build ./internal/clients/...` succeeds

## 6. Clean up dead utility code

- [ ] 6.1 Remove SDK-dependent functions from `internal/utils/utils.go` (`MergeSchemaMaps`, `AddConnectionSchema`, `ExpandIndividuallyDefinedSettings`); keep `ConvertSettingsKeyToTFFieldKey` temporarily
- [ ] 6.2 Move `ConvertSettingsKeyToTFFieldKey` to `internal/utils/typeutils` (or similar non-SDK package)
- [ ] 6.3 Update `internal/elasticsearch/index/indices/models.go` import from `schemautil "github.com/elastic/terraform-provider-elasticstack/internal/utils"` to the new location
- [ ] 6.4 Delete `internal/utils/utils.go` if now empty; otherwise rename package if confused with `internal/utils/` directory
- [ ] 6.5 Delete `internal/tfsdkutils/diffs.go`
- [ ] 6.6 Delete `internal/elasticsearch/index/commons.go`
- [ ] 6.7 Delete `internal/elasticsearch/index/template_sdk_shared.go`
- [ ] 6.8 Verify `go build ./internal/utils/... ./internal/tfsdkutils/... ./internal/elasticsearch/index/...` succeeds

## 7. Update tests

- [ ] 7.1 Rewrite `provider/provider_test.go` `TestProvider`: validate PF provider directly instead of calling `provider.New("dev").InternalValidate()`
- [ ] 7.2 Update or remove `provider/factory_test.go` `TestMuxServer`: if keeping, rewrite to test direct PF server factory; if removing, delete the test file
- [ ] 7.3 Rewrite `provider/connection_schema_test.go`: remove SDK entity enumeration and validation; keep PF entity connection block validation only
- [ ] 7.4 Update `provider/connection_schema_test_helpers_test.go`: remove `sortedSDKEntityNames` and `sdkschema` references if no longer needed
- [ ] 7.5 Update `internal/clients/kibanaoapi/status_test.go`: change assertions from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`
- [ ] 7.6 Update `internal/clients/kibanaoapi/security_role_test.go`: change assertions from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`
- [ ] 7.7 Search for any other test files referencing removed SDK functions and update/remove
- [ ] 7.8 Verify `go test ./provider/... ./internal/clients/kibanaoapi/...` passes

## 8. Final verification

- [ ] 8.1 Run `make build` at project root
- [ ] 8.2 Run `make check-lint`
- [ ] 8.3 Search for remaining references to `terraform-plugin-sdk/v2/diag` in non-test production code; verify all are legitimate (e.g., only in test files or truly necessary)
- [ ] 8.4 Search for remaining references to `schema.ResourceData` in non-test production code; verify all are legitimate
- [ ] 8.5 Run acceptance tests for affected kibana resources (`security_role`, `spaces`, `connectors`) against a running stack
- [ ] 8.6 Verify `go mod tidy` produces a clean diff with no unexpected re-adds
