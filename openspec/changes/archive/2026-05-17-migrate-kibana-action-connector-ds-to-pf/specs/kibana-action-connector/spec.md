## MODIFIED Requirements

### Requirement: Implementation path (metadata)
The data source SHALL be implemented at `internal/kibana/connectors/data_source.go` using `entitycore.NewKibanaDataSource`. The previous implementation at `internal/kibana/connector_data_source.go` is removed.

All data source behavioral requirements (REQ-DS-001 through REQ-DS-008) are unchanged.

#### Scenario: Data source registration
- **WHEN** the provider is initialized
- **THEN** the `elasticstack_kibana_action_connector` data source SHALL be registered via the Plugin Framework provider (not the SDK provider)
