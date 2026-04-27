## 1. Custom types

- [ ] 1.1 Add `internal/utils/customtypes/index_settings_value.go` implementing `IndexSettingsType` / `IndexSettingsValue` with `StringValuableWithSemanticEquals` (port `flattenMap` + `normalizeIndexSettings` from `internal/tfsdkutils/diffs.go`) and `xattr.ValidateableAttribute` for JSON-object validation.
- [ ] 1.2 Add unit tests in `internal/utils/customtypes/index_settings_value_test.go`, lifting the cases from `internal/utils/utils_test.go` covering dotted-vs-nested keys, `index.` prefixing, and stringified value comparison.
- [ ] 1.3 Implement the alias element custom type (`AliasObjectType` / `AliasObjectValue`) inside the new template package, satisfying `basetypes.ObjectValuable` and `basetypes.ObjectValuableWithSemanticEquals`. The `ObjectSemanticEquals` rule MUST match the strict predicate in design.md §2.
- [ ] 1.4 Add unit tests for the alias element semantic equality covering: identical values, differing `name`, differing `routing`, derived `index_routing`/`search_routing` echoes, differing `filter` JSON whitespace, null vs empty string equivalence, and the asymmetric prior-vs-new direction.

## 2. Client layer

- [ ] 2.1 Change `internal/clients/elasticsearch.PutIndexTemplate`, `GetIndexTemplate`, and `DeleteIndexTemplate` to return `fwdiag.Diagnostics`. Use `diagutil.CheckErrorFromFW()` for HTTP responses and `diagutil.FrameworkDiagFromError()` for other errors.
- [ ] 2.2 Confirm via grep that the only callers are `internal/elasticsearch/index/template.go` and `internal/elasticsearch/index/template_test.go`; both move in this change so no `SDKDiagsFromFramework` shim is needed.

## 3. Plugin Framework package skeleton

- [ ] 3.1 Create `internal/elasticsearch/index/template/` and add `resource.go` defining `Resource{*resourcecore.Core}` with `newResource()` / `NewResource()` / `ImportState` (passthrough on `id`).
- [ ] 3.2 Add `data_source.go` with explicit `Configure` / `Metadata` / `Read`, registered as `<provider>_elasticsearch_index_template`.
- [ ] 3.3 Wire interface assertions for `resource.ResourceWithConfigure`, `ResourceWithImportState`, `ResourceWithValidateConfig` (used for plan-time predicates that don't need the server version), `ResourceWithUpgradeState`, and `datasource.DataSourceWithConfigure`.

## 4. Schema

- [ ] 4.1 Add `internal/elasticsearch/index/template/schema.go` with the full resource schema mirroring `internal/elasticsearch/index/template.go`. Use `providerschema.GetEsFWConnectionBlock()` for the connection block. Set `Version: 1` on the resource schema.
- [ ] 4.2 Declare each of the six single-item containers as `schema.SingleNestedBlock`: `data_stream`, `template`, `template.lifecycle`, `template.data_stream_options`, `template.data_stream_options.failure_store`, `template.data_stream_options.failure_store.lifecycle`. The `template.alias` collection remains a `SetNestedBlock`.
- [ ] 4.3 Use `jsontypes.Normalized` for `metadata`, `template.mappings`, and `template.alias.filter`.
- [ ] 4.4 Use `customtypes.IndexSettingsType` for `template.settings`.
- [ ] 4.5 Use the alias `CustomType` on the `SetNestedBlock` `NestedBlockObject` for `template.alias`.
- [ ] 4.6 Implement `ValidateConfig` enforcing the existing REQ-032 rule that `data_stream_options` configured without `failure_store` is rejected at plan time (replaces the SDK `Required: true` flag, which `SingleNestedBlock` does not expose).
- [ ] 4.7 Add `data_source_schema.go` returning a Computed-only mirror schema (mirror the same `SingleNestedBlock` shape on the data source); share descriptions via constants in `descriptions.go`.

## 5. Models, expand, flatten

- [ ] 5.1 Add `internal/elasticsearch/index/template/models.go` with typed plan/state structs for the resource and the alias custom type.
- [ ] 5.2 Add `expand.go` translating typed models into `models.IndexTemplate` (metadata JSON parse, mappings/settings JSON parse, alias map keyed by name, lifecycle, data_stream, data_stream_options).
- [ ] 5.3 Add `flatten.go` translating Get response into the typed model. Drop the `extractAliasRoutingFromTemplateState`/`preserveAliasRoutingInFlattenedAliases` workarounds — semantic equality replaces them.
- [ ] 5.4 Add `version_gating.go` with `validateIgnoreMissingComponentTemplatesVersion` and `validateDataStreamOptionsVersion` returning `fwdiag.Diagnostics`.

## 6. CRUD

- [ ] 6.1 `create.go`: load plan, fetch server version, run version gating, expand, call `PutIndexTemplate`, set `id` (`<cluster_uuid>/<name>`), call shared read to refresh state.
- [ ] 6.2 `read.go`: extract a package-private `readIndexTemplate(ctx, client, name) (Model, fwdiag.Diagnostics)` used by the resource Read and the data source Read. Remove from state when not found (resource path) / leave defaulted (data source path, matching today's SDK behavior).
- [ ] 6.3 `update.go`: load prior state and plan, fetch server version, run version gating, apply the 8.x `allow_custom_routing` workaround by inspecting prior state, expand, call `PutIndexTemplate`, refresh.
- [ ] 6.4 `delete.go`: parse composite ID, call `DeleteIndexTemplate`.
- [ ] 6.5 `data_source.go`: implement Read by calling `readIndexTemplate` with the configured `name` and setting `id`.

## 7. State upgrade (v0 → v1)

- [ ] 7.1 Add `state_upgrade.go` registering an `UpgradeState` map with version `0` mapped to a `StateUpgrader` that uses `RawState` JSON.
- [ ] 7.2 Implement a JSON-tree walk that collapses each of the six `MaxItems: 1` paths from `[]`/`[obj]` to `null`/`obj`, applied top-down: `data_stream`, `template`, `template.lifecycle`, `template.data_stream_options`, `template.data_stream_options.failure_store`, `template.data_stream_options.failure_store.lifecycle`. The walker must collapse a parent before descending into it.
- [ ] 7.3 Return an error diagnostic with the offending path if any of these paths is found with more than one element (defensive against corrupt prior state).
- [ ] 7.4 Use a fresh PF state encode for the v1 schema (set the resulting `RawState` via the response so the framework re-decodes against the v1 schema).
- [ ] 7.5 Add unit tests in `state_upgrade_test.go` covering, for each path: missing/null, empty list `[]`, single-element list `[obj]`, multi-element list (error), and round-trip encode→upgrade→decode against the v1 schema. Use the `ilm` package's `state_upgrade_test.go` as a structural reference.

## 8. Provider wiring

- [ ] 8.1 Register `template.NewResource` in `provider/plugin_framework.go` `resources()`.
- [ ] 8.2 Register `template.NewDataSource` in `provider/plugin_framework.go` `dataSources()`.
- [ ] 8.3 Remove `"elasticstack_elasticsearch_index_template"` entries from both `ResourcesMap` and `DataSourcesMap` in `provider/provider.go`.

## 9. Acceptance tests

- [ ] 9.1 Move `internal/elasticsearch/index/template_test.go` into `internal/elasticsearch/index/template/acc_test.go` (package `template_test`). Update imports and any helper references (`checkResourceDestroy`).
- [ ] 9.2 Move `internal/elasticsearch/index/template_data_source_test.go` into the new package as `data_source_acc_test.go` (or merged into `acc_test.go`).
- [ ] 9.3 Move test data: `internal/elasticsearch/index/testdata/TestAccResourceIndexTemplate*` into the new package's `testdata/`.
- [ ] 9.4 Add `TestAccResourceIndexTemplateFromSDK` per REQ-042. Two-step `resource.TestCase`: Step 1 pins the last SDK-based provider release via `ExternalProviders` and `VersionConstraint` and applies a config exercising every collapsed block (`data_stream`, `template`, `template.alias`, `template.lifecycle`, `template.mappings`, `template.settings`, `template.data_stream_options`, `template.data_stream_options.failure_store`, `template.data_stream_options.failure_store.lifecycle`) plus the routing-only alias scenario; Step 2 switches to `ProtoV6ProviderFactories` and re-applies the equivalent configuration with `PlanOnly` / `ExpectNonEmptyPlan: false` (or equivalent) to assert a no-op upgrade. Guard the `data_stream_options` portion of the configuration with the `9.1.0` stack-version skip; keep the rest of the upgrade coverage running on older stacks. See `dev-docs/high-level/sdk-to-pf-migration.md` and the `internal/elasticsearch/security/user/` and `internal/elasticsearch/index/ilm/` precedents for the pattern.
- [ ] 9.5 Resolve the `VersionConstraint` value used in 9.4 by inspecting the latest tagged provider release at the time of merge (e.g. `0.14.x`) and confirming that release still implemented `elasticstack_elasticsearch_index_template` on Plugin SDK v2. Capture the chosen version in the test source as a comment so future readers understand the pin.

## 10. Schema coverage

- [ ] 10.1 Run the `schema-coverage` skill against the new package and add tests for any attribute/block lacking coverage. Pay particular attention to `data_stream_options.failure_store.lifecycle.data_retention`, `composed_of`, `ignore_missing_component_templates`, `priority`, `version`, and the alias derived-routing edge cases.
- [ ] 10.2 Add an explicit `TestAccResourceIndexTemplate_importState` if not already present.

## 11. Cleanup

- [ ] 11.1 Delete `internal/elasticsearch/index/template.go`.
- [ ] 11.2 Delete `internal/elasticsearch/index/template_data_source.go`.
- [ ] 11.3 Verify `internal/elasticsearch/index/component_template.go` still compiles unchanged. Confirm that the SDK helpers (`expandTemplate`, `flattenTemplateData`, `extractAliasRoutingFromTemplateState`, `preserveAliasRoutingInFlattenedAliases`, `hashAliasByName`, `stringIsJSONObject`) remain referenced exclusively by component template; do not delete them.
- [ ] 11.4 Move `MinSupportedIgnoreMissingComponentTemplateVersion` and `MinSupportedDataStreamOptionsVersion` from `internal/elasticsearch/index/template.go` into the new package; if `templateilmattachment` references them, expose via a small constants file.

## 12. Verification

- [ ] 12.1 `make build`.
- [ ] 12.2 `go test ./internal/elasticsearch/index/template/... -v`.
- [ ] 12.3 `go test ./internal/utils/customtypes/... -v`.
- [ ] 12.4 Acceptance tests: `go test ./internal/elasticsearch/index/template/... -v -count=1 -run TestAcc` against a live stack (see `dev-docs/high-level/testing.md`).
- [ ] 12.5 Regression: `go test ./internal/elasticsearch/index/templateilmattachment/... -v -count=1 -run TestAcc` to exercise the downstream resource that references index templates.
- [ ] 12.6 `make check-openspec`.
- [ ] 12.7 Regenerate documentation if affected (`make docs-generate`); confirm the new resource and data source pages render correctly.
