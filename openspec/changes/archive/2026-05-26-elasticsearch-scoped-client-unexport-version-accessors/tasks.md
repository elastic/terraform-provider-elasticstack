## 1. New public APIs on `ElasticsearchScopedClient`

- [x] 1.1 Add `IsServerless(ctx) (bool, fwdiag.Diagnostics)` to `internal/clients/elasticsearch_scoped_client.go` — reads `serverInfo().Version.BuildFlavor`, returns `flavor == ServerlessFlavor`
- [x] 1.2 Add `EnforceVersionCheck(ctx, check func(*version.Version) bool) (bool, fwdiag.Diagnostics)` mirroring the Kibana client signature (serverless short-circuits to `true`, otherwise calls `check(serverVersion)`)
- [x] 1.3 Add unit tests in `internal/clients/elasticsearch_scoped_client_test.go` for both new methods: missing endpoint, info API error, stateful below/at/above min, serverless short-circuit, predicate returning false on stateful, predicate true on stateful

## 2. Pattern A — `security/role/update.go` → `WithVersionRequirements`

- [x] 2.1 Read `internal/elasticsearch/security/role/models.go`; identify the conditional gates (`description` → `MinSupportedDescriptionVersion`, `remote_indices` → `MinSupportedRemoteIndicesVersion`)
- [x] 2.2 Add `GetVersionRequirements()` to the role model, returning a slice of `entitycore.VersionRequirement` whose entries are only present when the relevant attribute is configured
- [x] 2.3 Unit test `TestModel_GetVersionRequirements` covering all four combinations of (description set/unset, remote_indices empty/non-empty)
- [x] 2.4 Remove inline `client.ServerVersion(ctx)` + `LessThan` blocks from `internal/elasticsearch/security/role/update.go`

## 3. Pattern B — transform per-setting `EnforceMinVersion`

- [x] 3.1 Change `internal/elasticsearch/transform/version_gating.go::isSettingAllowed` signature from `(ctx, settingName, *version.Version) bool` to `(ctx, settingName, *clients.ElasticsearchScopedClient) (bool, diag.Diagnostics)` (or equivalent — call `client.EnforceMinVersion` against `settingsRequiredVersions[settingName]`)
- [x] 3.2 Update all `isSettingAllowed` call sites in `internal/elasticsearch/transform/write.go`
- [x] 3.3 Delete the `serverVersion` fetch in `transform/write.go`
- [x] 3.4 Update existing transform unit tests that exercise `isSettingAllowed` with explicit versions to stub a client or test via the public path

## 4. Pattern B — `index/index` flavor → `IsServerless`

- [x] 4.1 Change `internal/elasticsearch/index/index/models.go::toPutIndexParams` signature from `(serverFlavor string)` to `(isServerless bool)`; flip the inner check to `if !isServerless { ... }`
- [x] 4.2 Update `internal/elasticsearch/index/index/create.go` to call `client.IsServerless(ctx)` and pass the boolean
- [x] 4.3 Update `Test_tfModel_toPutIndexParams` in `models_test.go` — already iterates `isServerless := []bool{true, false}` so just remove the flavor-string conversion

## 5. Pattern C — apikey capability struct + private-state migration

- [x] 5.1 Define `type apikeyCapabilities struct { SupportsUpdate, SupportsRoleDescriptors, SupportsRestriction bool }` in a new file `internal/elasticsearch/security/apikey/capabilities.go`
- [x] 5.2 Implement `resolveAPIKeyCapabilities(ctx, *clients.ElasticsearchScopedClient) (apikeyCapabilities, diag.Diagnostics)` calling `client.EnforceMinVersion` once per bit against the three existing constants (`MinVersionWithUpdate`, `MinVersionReturningRoleDescriptors`, `MinVersionWithRestriction`)
- [x] 5.3 Replace `runtime_validation.go`'s direct `ServerVersion` + `LessThan(MinVersionWithRestriction)` with `resolveAPIKeyCapabilities(ctx, client)` then a check on `caps.SupportsRestriction`
- [x] 5.4 Replace `models.go` line 415's `serverVersion.GreaterThanOrEqual(MinVersionReturningRoleDescriptors)` with a `caps.SupportsRoleDescriptors` parameter — change the function signature to accept `apikeyCapabilities` rather than `*version.Version`
- [x] 5.5 Rename `clusterVersionPrivateData` → keep the JSON struct shape compatible at the byte level. Add a new `apikeyCapabilitiesPrivateData` JSON struct
- [x] 5.6 Rename `saveClusterVersion` → `saveAPIKeyCapabilities`. It calls `resolveAPIKeyCapabilities` then marshals the capabilities struct into the existing `cluster-version` private-state slot (slot key unchanged for state-compat)
- [x] 5.7 Rename `postReadPersistClusterVersion` → `postReadPersistAPIKeyCapabilities`
- [x] 5.8 Rename `clusterVersionOfLastRead` → `apikeyCapabilitiesOfLastRead`. Logic:
  - Read raw bytes from the `cluster-version` private-state slot
  - Try `json.Unmarshal` into `apikeyCapabilitiesPrivateData`; if any boolean field is true, return it
  - On failure or all-false (ambiguous with legacy `{"Version":""}`), fall back to the legacy `{Version string}` shape; if `Version` is non-empty, parse with `version.NewVersion`, synthesize capabilities by comparing against the three constants, return them
  - Return `nil, nil` only when both shapes fail to yield useful data
- [x] 5.9 Update `schema.go::requiresReplaceIfUpdateNotSupported` to read `caps.SupportsUpdate` rather than `ver.LessThan(MinVersionWithUpdate)`
- [x] 5.10 New unit test `TestApikeyCapabilitiesOfLastRead_LegacyVersionBlob` covering:
  - empty private state → returns zero value, no error
  - legacy `{"Version":"7.0.0"}` blob → capabilities with all `false`
  - legacy `{"Version":"8.20.0"}` blob → all `true`
  - new `{"SupportsUpdate":true,...}` blob → returned as-is
- [x] 5.11 New unit test `TestSaveAPIKeyCapabilities` covering serverless (all true) and a known-stateful version (mixed)
- [x] 5.12 Migrate `internal/elasticsearch/security/apikey/resource/acc_test.go::350` predicate (`LessThan || GreaterThanOrEqual`) to `client.EnforceVersionCheck(ctx, func(v) { ... })`

## 6. `internal/clients` acceptance-test escape hatch

- [x] 6.1 Add `AcceptanceServerInfo(ctx, *ElasticsearchScopedClient) (*version.Version, bool, fwdiag.Diagnostics)` in a new file `internal/clients/acceptance_testing_version.go` — calls the package-private `serverInfo` helper and returns version + serverless boolean. Documented "test-only"
- [x] 6.2 Add unit test in `internal/clients/elasticsearch_scoped_client_test.go` exercising `AcceptanceServerInfo` against the existing HTTP fixture

## 7. `versionutils/testutils.go` migration

- [x] 7.1 Replace `fetchAcceptanceServerInfo` body — call `clients.AcceptanceServerInfo(ctx, client)` instead of `client.ServerVersion` + `client.ServerFlavor`
- [x] 7.2 Update `CheckIfNotServerless` to call `client.IsServerless(ctx)` and invert the boolean
- [x] 7.3 Run `go test ./internal/versionutils/... ./internal/acctest/...`

## 8. Rewrite `internal/clients` tests

- [x] 8.1 Remove `TestElasticsearchScopedClient_ServerVersion`, `TestElasticsearchScopedClient_ServerFlavor` from `internal/clients/elasticsearch_scoped_client_test.go`. Behaviour now tested via `EnforceMinVersion`, `EnforceVersionCheck`, `IsServerless`
- [x] 8.2 Update `provider_client_factory_test.go` — replace any test calling raw accessors with equivalents through `EnforceMinVersion` or `IsServerless`

## 9. Remove the public accessors

- [x] 9.0 Migrate `internal/elasticsearch/index/ilm/{create,update}.go` callers (proposal missed these)
- [x] 9.1 `rg "\.ServerVersion\(|\.ServerFlavor\(" internal/elasticsearch internal/clients internal/versionutils internal/acctest` — confirm no production references remain; only `internal/clients/acceptance_testing_version.go` and the package-private helpers may use them
- [x] 9.2 Delete `ServerVersion` and `ServerFlavor` public methods from `internal/clients/elasticsearch_scoped_client.go` (or unexport to `serverVersion` / `serverFlavor` if a private call site needs them — note that `EnforceMinVersion` already calls `e.ServerVersion(ctx)` and `e.ServerFlavor(ctx)` internally and needs corresponding rewrites to the package-private form)
- [x] 9.3 Rewrite `EnforceMinVersion`, `EnforceVersionCheck`, `IsServerless` to call `e.serverInfo(ctx)` directly and read `info.Version.Int` / `info.Version.BuildFlavor` once each — eliminates the redundant cached calls and removes any remaining references to the unexported helpers
- [x] 9.4 Run `make build` — must compile cleanly

## 10. Spec updates

- [x] 10.1 Apply the `provider-client-factory` delta to `openspec/specs/provider-client-factory/spec.md` (adds the "Elasticsearch scoped client serverless-safe version surface" requirement)
- [x] 10.2 Apply the `elasticsearch-client-pf-diagnostics` delta to `openspec/specs/elasticsearch-client-pf-diagnostics/spec.md` (replaces the existing "ElasticsearchScopedClient methods return Plugin Framework diagnostics" requirement)

## 11. Validation

- [x] 11.1 `make check-openspec` (or `openspec validate`) passes
- [x] 11.2 `make build` succeeds
- [x] 11.3 `go test ./internal/elasticsearch/... ./internal/clients/... ./internal/entitycore/... ./internal/versionutils/...` passes
- [ ] 11.4 If acceptance environment is available, run targeted acc tests for the affected resources: `security/role`, `security/apikey`, `index/index`, `transform`. Verify legacy private-state apikey resources still refresh successfully against both stateful and serverless clusters
