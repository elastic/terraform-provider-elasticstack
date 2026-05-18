## MODIFIED Requirements

### Requirement: Implementation path (metadata)
The resource SHALL be implemented at `internal/kibana/security_role/resource.go` using `entitycore.NewKibanaResource`. The data source SHALL be implemented at `internal/kibana/security_role/data_source.go` using `entitycore.NewKibanaDataSource`. The previous implementations at `internal/kibana/role.go` and `internal/kibana/role_data_source.go` are removed.

All behavioral requirements (REQ-001 through REQ-026) are unchanged.

#### Scenario: Resource and data source registration
- **WHEN** the provider is initialized
- **THEN** both `elasticstack_kibana_security_role` resource and data source SHALL be registered via the Plugin Framework provider (not the SDK provider)
