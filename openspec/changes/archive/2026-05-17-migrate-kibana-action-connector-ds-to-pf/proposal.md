## Why

The `elasticstack_kibana_action_connector` data source is the only remaining SDK-based entity in the connectors domain — its resource counterpart already uses the Plugin Framework. Completing the migration removes the last SDK dependency in this package and aligns with the provider-wide move away from `terraform-plugin-sdk/v2`.

## What Changes

- Add `internal/kibana/connectors/data_source.go` implementing the data source via `entitycore.NewKibanaDataSource`
- Register `connectors.NewDataSource` in `provider/plugin_framework.go`
- Remove `kibana.DataSourceConnector()` from `provider/provider.go`
- Delete `internal/kibana/connector_data_source.go`
- Move acceptance tests from `internal/kibana/connector_data_source_test.go` into `internal/kibana/connectors/acc_test.go`
- Add SDK upgrade test (`TestAccConnectorsDataSourceFromSDK`) with `VersionConstraint: "0.15.1"`
- Update the implementation path reference in `openspec/specs/kibana-action-connector/spec.md`

## Capabilities

### New Capabilities

None. The data source schema and behavior are unchanged.

### Modified Capabilities

- `kibana-action-connector`: Implementation path for data source changes from `internal/kibana/connector_data_source.go` to `internal/kibana/connectors/data_source.go`. No schema or behavioral requirement changes.

## Impact

- `internal/kibana/connectors/` — new `data_source.go` added to existing PF package
- `internal/kibana/connector_data_source.go` — deleted
- `provider/provider.go` — one entry removed from `DataSourcesMap`
- `provider/plugin_framework.go` — one entry added to `dataSources()`
- `openspec/specs/kibana-action-connector/spec.md` — implementation path updated
