## Why

The provider has two parallel data source wiring patterns: the generic `entitycore` envelope (`NewKibanaDataSource` / `NewElasticsearchDataSource`) and manual struct-based wiring (`DataSourceBase` embedding or fully bespoke `Configure`/`Metadata`/`Read`). The envelope eliminates duplicated boilerplate (config decode, scoped client resolution, state persistence, connection block schema injection) and is already proven by the Agent Builder POC. Migrating all remaining Plugin Framework data sources to the envelope unifies the codebase, reduces maintenance surface, and makes `DataSourceBase` — which will have zero consumers — removable.

## What Changes

- **Migrate 9 remaining PF data sources** to the generic envelope:
  - **Kibana-backed** (use `NewKibanaDataSource`): `spaces`, `export_saved_objects`
  - **Fleet-backed Kibana client** (use `NewKibanaDataSource` with `ComponentFleet`): `fleet_integration`, `fleet_output`, `enrollment_tokens`
  - **Elasticsearch-backed** (use `NewElasticsearchDataSource`): `indices`, `index_template`, `enrich_policy`, `security_role_mapping`
- **Update `agentbuilder_agent` model** to embed `entitycore.KibanaConnectionField` instead of a hand-rolled `GetKibanaConnection()` accessor, aligning it with other envelope data sources. The optional `GetVersionRequirements()` method is preserved because this data source still requires pre-read version gating.
- **Remove `DataSourceBase`** from `internal/entitycore/data_source_base.go` and all dependent `entitycore_contract_test.go` files, since every PF data source will use the envelope.
- **Convert schema methods to schema factories** for all migrated data sources: extract the current `Schema` method body into a package-level `func() dsschema.Schema` that omits the connection block (the envelope injects it).
- **Convert `Read` methods to pure read callbacks** for all migrated data sources: strip envelope-owned orchestration (`req.Config.Get`, client resolution, `resp.State.Set`) and keep only the entity-specific API call and model population logic.
- **Preserve all existing Terraform type names and schemas** — no user-visible breaking changes.

## Capabilities

### New Capabilities
_(none — this is an internal refactoring with no new user-visible capabilities)_

### Modified Capabilities
_(none — no spec-level requirements change; all migrated data sources preserve their existing schemas and read behavior)_

## Impact

- **`internal/entitycore/`**: `data_source_envelope.go` gains the sole PF data source pattern. `data_source_base.go` and `data_source_base_test.go` are removed.
- **`internal/kibana/spaces/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope. `entitycore_contract_test.go` removed.
- **`internal/kibana/exportsavedobjects/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope.
- **`internal/fleet/integrationds/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope.
- **`internal/fleet/outputds/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope.
- **`internal/fleet/enrollmenttokens/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope. `entitycore_contract_test.go` removed.
- **`internal/elasticsearch/index/indices/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope.
- **`internal/elasticsearch/index/template/`**: `data_source.go`, `data_source_schema.go`, `read.go`, and `models.go` refactored to envelope.
- **`internal/elasticsearch/enrich/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope. `entitycore_contract_test.go` removed.
- **`internal/elasticsearch/security/rolemapping/`**: `data_source.go`, `schema.go`, `read.go`, and `models.go` refactored to envelope.
- **`internal/kibana/agentbuilderagent/`**: `models.go` updated to embed `KibanaConnectionField`; hand-rolled `GetKibanaConnection()` removed. `GetVersionRequirements()` is preserved because the data source still requires pre-read version gating.
- **Provider registration** (`provider/plugin_framework.go`): No changes needed — `NewDataSource` constructors remain stable.
- **Tests**: All existing acceptance tests should pass without modification. Unit tests that assert on schema block presence may need minor updates since the block is now injected by the envelope.
