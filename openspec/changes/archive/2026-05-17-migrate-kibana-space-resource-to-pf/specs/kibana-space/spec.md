## MODIFIED Requirements

### Requirement: Implementation path (metadata)
The resource SHALL be implemented at `internal/kibana/spaces/resource.go` using `entitycore.NewKibanaResource`. The previous implementation at `internal/kibana/space.go` is removed.

All behavioral requirements (REQ-001 through REQ-020) are unchanged.

#### Scenario: Resource registration
- **WHEN** the provider is initialized
- **THEN** the `elasticstack_kibana_space` resource SHALL be registered via the Plugin Framework provider (not the SDK provider)
