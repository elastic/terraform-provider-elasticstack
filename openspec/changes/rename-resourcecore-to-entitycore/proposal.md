## Why

`internal/resourcecore` provides the shared Plugin Framework substrate for Configure/Metadata/client wiring across every Plugin Framework **resource** in this provider (~30 resources today). The same boilerplate — a `*clients.ProviderClientFactory` field, a `Configure` that calls `clients.ConvertProviderDataToFactory`, and a `Metadata` that produces `<provider>_<component>_<name>` — is also re-implemented by hand in every Plugin Framework **data source** (~13 today) because the package was scoped to resources.

Now that the core has stabilized and proven its embed contract via the `resourcecore_contract_test.go` pattern, it should be widened to cover both Terraform entity kinds. To make that growth honest the package needs a name that reflects what it owns, and consumers need two embedded shapes (one per kind) that are clearly distinguished at the call site.

This change does the rename, introduces a parallel `DataSourceBase`, and proves the `DataSourceBase` shape on one minimal-diff pilot per stack component that has Plugin Framework data sources today (elasticsearch, kibana, fleet). APM has no Plugin Framework data sources and is intentionally excluded from the data-source pilot.

## What Changes

- Rename the Go package `internal/resourcecore` to `internal/entitycore`. Move all files (`core.go`, `doc.go`, `core_test.go`, `configure_test.go`) accordingly and update every import across the repo.
- Rename the embedded type `resourcecore.Core` to `entitycore.ResourceBase`. The constructor becomes `NewResourceBase`. The package-level component constants (`ComponentElasticsearch`, `ComponentKibana`, `ComponentFleet`, `ComponentAPM`) and the `Component` type are preserved unchanged. The factory-stored client and `Client()` accessor are preserved unchanged.
- Update every embed site in the repo (~30 resources plus `internal/kibana/import_saved_objects/schema.go`) from `*resourcecore.Core` to `*entitycore.ResourceBase`. Update every constructor call from `resourcecore.New(...)` to `entitycore.NewResourceBase(...)`. Update the six `resourcecore_contract_test.go` files (`internal/elasticsearch/index/alias`, `internal/kibana/dashboard`, `internal/kibana/maintenance_window`, `internal/fleet/customintegration`, `internal/fleet/agentdownloadsource`, `internal/kibana/connectors`) to assert `*entitycore.ResourceBase`, rename them to `entitycore_contract_test.go`, and rename the asserted field accordingly.
- Add a parallel embedded type `entitycore.DataSourceBase` whose surface mirrors `ResourceBase` but uses the Plugin Framework `datasource` request/response types. It SHALL store a `*clients.ProviderClientFactory`, expose it via `Client()`, configure it via `clients.ConvertProviderDataToFactory` with the same diagnostics-failure rule as `ResourceBase`, and produce Terraform type names of the form `<provider_type_name>_<component>_<data_source_name>` from a typed `Component` and a literal data-source-name suffix. `DataSourceBase` SHALL NOT define `Read`, `Schema`, or `ConfigValidators`; concrete data sources retain those.
- Migrate one Plugin Framework data source per stack component to embed `*entitycore.DataSourceBase`, picked for minimal diff:
  - `internal/elasticsearch/enrich/data_source.go` (`enrichPolicyDataSource` → embed `*entitycore.DataSourceBase`, type name `<p>_elasticsearch_enrich_policy`).
  - `internal/kibana/spaces/data_source.go` (`dataSource` → embed `*entitycore.DataSourceBase`, type name `<p>_kibana_spaces`; literal suffix `spaces`).
  - `internal/fleet/enrollmenttokens/data_source.go` (`enrollmentTokensDataSource` → embed `*entitycore.DataSourceBase`, type name `<p>_fleet_enrollment_tokens`).
- Add an `entitycore_contract_test.go` next to each migrated data source asserting that the concrete data source type embeds `*entitycore.DataSourceBase` and that the embedded field is anonymous, mirroring the existing resource-side contract test pattern.
- Rename the canonical spec capability from `provider-framework-resource-core` to `provider-framework-entity-core`. Move `openspec/specs/provider-framework-resource-core/spec.md` to `openspec/specs/provider-framework-entity-core/spec.md` (handled at archive time via the delta in this change). Update prose references in active (unarchived) OpenSpec change artifacts only; archived change artifacts are left intact.

The change is purely internal: there are no Terraform schema changes, no resource type-name changes, no data-source type-name changes, no provider-config surface changes, and no state shape changes. CRUD, schema, validators, and import behavior on every concrete resource and data source are preserved. Acceptance tests are not added or modified by this change.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `provider-framework-entity-core` (renamed from `provider-framework-resource-core`): broadens the canonical shared Plugin Framework substrate from "resource core" to "entity core", introducing a parallel `DataSourceBase` with the same Configure/Metadata/Client contract for Plugin Framework data sources, and renaming the resource-side embedded type from `Core` to `ResourceBase`. The user-observable contract for every existing resource (type names, configure diagnostics behavior, import behavior) is preserved exactly. The newly-stated requirements for compatible data sources match the existing requirements for compatible resources, applied to the data-source variant.

## Impact

- Affected code is provider-internal: `internal/entitycore/` (renamed from `internal/resourcecore/`), every `internal/**/resource.go` that embeds the core, every `resourcecore_contract_test.go` (renamed to `entitycore_contract_test.go`), `internal/kibana/import_saved_objects/schema.go`, plus the three pilot data sources and their new contract tests. No public Terraform schema, type name, identity format, or import behavior is affected. No client-facing code path changes.
- The rename touches a large number of files (~35) but each touch is mechanical (package name, import path, embedded type spelling, constructor name). The substance of the change lives in `internal/entitycore/data_source_base.go` and the three pilot data-source migrations.
- Risk surface is the embed promotion: `DataSourceBase.Configure` and `DataSourceBase.Metadata` use `datasource.ConfigureRequest`/`MetadataRequest`, which are distinct framework types from their resource counterparts, so the resource-side and data-source-side `Configure`/`Metadata` methods cannot share an implementation body without generics. The two implementations are deliberately written as a small, near-symmetric pair to keep diff review obvious.
- No backward-compatibility concerns for users: state, plans, and existing acceptance tests continue to behave identically. The change is intended to land as a single PR that compiles and passes the existing test suite without modification beyond the file renames.
