## Why

The `elasticstack_kibana_agentbuilder_agent` data source still implements Plugin Framework boilerplate locally: provider data conversion, client storage, metadata naming, config decode, scoped Kibana client resolution, and state persistence. The `entitycore` package already provides a generic Kibana data source envelope that centralizes this outer flow, but it does not yet own static per-model server version gates.

Migrating this data source to the envelope reduces bespoke data source wiring and makes the static Agent Builder API version requirement a reusable concern of the generic data source pipeline. The Agent Builder data source still needs local read logic for its entity-specific behavior: composite ID and space resolution, optional dependency export, tool expansion, and workflow lookup.

## What Changes

- Add an optional Kibana data source version-requirements interface to `internal/entitycore/data_source_envelope.go`.
- Teach the generic Kibana data source envelope to enforce model-provided minimum version requirements after resolving the Kibana scoped client and before invoking the entity read callback.
- Migrate the Agent Builder agent data source constructor to `entitycore.NewKibanaDataSource`.
- Refactor the current Agent Builder data source `Read` method into an envelope read callback.
- Move the unconditional Agent Builder API minimum version check into `agentDataSourceModel` via the optional version-requirements interface.
- Keep the conditional workflow-tool version check in the Agent Builder callback because it depends on API data discovered during the read.
- Convert the Agent Builder data source schema method into a schema factory and let the envelope inject `kibana_connection`.
- Update tests to cover the new envelope version hook and preserve the data source's Terraform type name and schema behavior.

## Capabilities

### New Capabilities

_(none - this is an internal implementation migration and shared data source envelope enhancement, not a user-visible Terraform capability)_

### Modified Capabilities

_(none - the `elasticstack_kibana_agentbuilder_agent` data source should preserve its existing user-visible schema and read behavior)_

## Impact

- **Shared implementation:** `internal/entitycore/data_source_envelope.go` gains optional pre-read version requirement enforcement for Kibana-backed envelope data sources.
- **Migrated data source:** `internal/kibana/agentbuilderagent/data_source.go`, `data_source_schema.go`, `data_source_read.go`, and `models.go` move from local data source orchestration to the generic Kibana envelope.
- **Tests:** `internal/entitycore/data_source_envelope_test.go` and Agent Builder data source tests should cover the new version hook and confirm schema/type-name compatibility.
- **No Terraform behavior change intended:** existing acceptance coverage for default space, explicit space, `kibana_connection`, `include_dependencies`, and workflow-tool export should continue to pass.
