## 1. Pattern A — `WithVersionRequirements` on the model (add, then drop inline check)

- [x] 1.1 Add `GetVersionRequirements()` to the `security_enable_rule` model (`internal/kibana/security_enable_rule/models.go`), returning the existing `8.11.0` minimum with the existing "Security detection rules bulk actions are not supported until Elastic Stack v8.11.0" error message
- [x] 1.2 Add a unit test `TestModel_GetVersionRequirements` for `security_enable_rule` covering: empty model, populated model — both return the same single requirement
- [x] 1.3 Remove inline `ServerVersion`+`LessThan` blocks in `internal/kibana/security_enable_rule/{read,update,delete}.go`
- [x] 1.4 Add `GetVersionRequirements()` to `internal/kibana/prebuilt_rules/models.go` returning `8.0.0` with the existing message
- [x] 1.5 Unit test for prebuilt_rules `GetVersionRequirements`
- [x] 1.6 Remove inline checks in `internal/kibana/prebuilt_rules/{read,update}.go`
- [x] 1.7 Add `GetVersionRequirements()` to `internal/kibana/slo/models.go` — returns two conditional requirements: `SLOSupportsPreventInitialBackfillMinVersion` when `Settings.PreventInitialBackfill` set; `SLOSupportsDataViewIDMinVersion` when `hasDataViewID()`
- [x] 1.8 Unit test for slo `GetVersionRequirements` covering all four combinations of the two conditions
- [x] 1.9 Remove inline checks in `internal/kibana/slo/{create,update}.go`
- [x] 1.10 Add `GetVersionRequirements()` to `internal/kibana/connectors/models.go` — returns a single requirement (`MinVersionSupportingPreconfiguredIDs`) when `connector_id` is configured, empty otherwise
- [x] 1.11 Unit test for connectors `GetVersionRequirements` covering connector_id set/unset
- [x] 1.12 Remove inline check in `internal/kibana/connectors/create.go`

## 2. Pattern B — `EnforceMinVersion` boolean for feature toggles

- [x] 2.1 `internal/kibana/synthetics/parameter/delete.go`: replace `kibanaVersion.LessThan(minKibanaPerIDDeleteVersion)` with `client.EnforceMinVersion(ctx, minKibanaPerIDDeleteVersion)`; if true use per-ID endpoint, else bulk
- [x] 2.2 `internal/kibana/agentbuilderagent/data_source.go`: replace `!serverVersion.LessThan(minVersionAdvancedAgentConfig)` with `client.EnforceMinVersion(ctx, minVersionAdvancedAgentConfig)`; assign result to `supportsAdvancedConfig`

## 3. Pattern C — `alertingRuleFeatures` capability struct

- [ ] 3.1 Enumerate every `*version.Version` comparison inside `internal/kibana/alertingrule/models*.go` and `internal/kibana/alertingrule/toAPIModel`; produce a definitive list of capability bits
- [ ] 3.2 Define `type alertingRuleFeatures struct { ... }` in `internal/kibana/alertingrule/features.go` with one field per identified capability bit, named `SupportsX`
- [ ] 3.3 Implement `resolveAlertingRuleFeatures(ctx, *clients.KibanaScopedClient) (alertingRuleFeatures, diag.Diagnostics)` calling `client.EnforceMinVersion` once per bit
- [ ] 3.4 Unit test `TestResolveAlertingRuleFeatures` using a stubbed `KibanaScopedClient` covering: server below all thresholds (all false), server above all thresholds (all true), serverless (all true)
- [ ] 3.5 Change `alertingrule.toAPIModel` signature from `(ctx, serverVersion *version.Version)` to `(ctx, features alertingRuleFeatures)`; rewrite all version comparisons inside as `features.SupportsX` lookups
- [ ] 3.6 Update unit tests under `internal/kibana/alertingrule/` that previously constructed `*version.Version` values to pass `alertingRuleFeatures{...}` instead
- [ ] 3.7 Update callers `internal/kibana/alertingrule/create.go` and `internal/kibana/alertingrule/update.go` to call `resolveAlertingRuleFeatures(ctx, client)` and pass the result into `toAPIModel`
- [ ] 3.8 Delete the now-unused `client.ServerVersion(ctx)` call in those files

## 4. Acceptance test sweep

- [ ] 4.1 `rg "ServerVersion|ServerFlavor" internal/kibana/**/acc_test.go` — for every hit, migrate to `versionutils.CheckIfNotServerless` or `EnforceMinVersion`. (If acceptance tests legitimately need the version itself for skip logic, route through `versionutils.SkipIfUnsupported` instead.)

## 5. Tests in `internal/clients/`

- [ ] 5.1 Rewrite `internal/clients/kibana_scoped_client_test.go`: remove `TestKibanaScopedClient_ServerVersion_*` and `TestKibanaScopedClient_ServerFlavor_MissingEndpoint`. Add tests for `EnforceMinVersion` and `EnforceVersionCheck` covering: missing endpoint, stateful below min, stateful at min, stateful above min, serverless short-circuit, malformed version response, status API error
- [ ] 5.2 In `internal/clients/provider_client_factory_test.go` replace `TestKibanaScopedClient_ServerFlavor_ViaFactory` with `TestKibanaScopedClient_EnforceMinVersion_ViaFactory` asserting the serverless short-circuit through the factory-obtained client

## 6. Remove the public accessors

- [ ] 6.1 `rg "\.ServerVersion\(|\.ServerFlavor\(" internal/kibana internal/clients` — confirm zero non-test references remain. Tests in `internal/clients/kibana_scoped_client_test.go` should now exercise only the public surface; if any direct method calls remain, return to the corresponding section above
- [ ] 6.2 Delete `ServerVersion` method from `internal/clients/kibana_scoped_client.go`
- [ ] 6.3 Delete `ServerFlavor` method from `internal/clients/kibana_scoped_client.go`
- [ ] 6.4 Run `make build` — must compile cleanly

## 7. Spec update

- [ ] 7.1 Apply the `provider-client-factory` delta from this change to `openspec/specs/provider-client-factory/spec.md` (the `MODIFIED Requirements` block replaces the existing "Kibana scoped client contract" requirement)
- [ ] 7.2 Apply the `elasticsearch-client-pf-diagnostics` delta to `openspec/specs/elasticsearch-client-pf-diagnostics/spec.md` (replaces the existing "KibanaScopedClient methods return Plugin Framework diagnostics" requirement and its scenarios)

## 8. Validation

- [ ] 8.1 `make check-openspec` (or `openspec validate`) passes
- [ ] 8.2 `make build` succeeds
- [ ] 8.3 `go test ./internal/kibana/... ./internal/clients/... ./internal/entitycore/...` passes
- [ ] 8.4 If acceptance environment is available (default `TF_ACC` variables), run targeted acc tests for the four Kibana resources whose write paths changed: `security_enable_rule`, `prebuilt_rules`, `slo`, `connectors`, `alertingrule`, `synthetics/parameter`, `agentbuilderagent` (data source)
