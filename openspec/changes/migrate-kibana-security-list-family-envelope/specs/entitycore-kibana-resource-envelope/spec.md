## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL continue to provide a generic constructor `NewKibanaResource[T]()` for Kibana-backed Terraform resources, and additional Kibana resources that follow the envelope lifecycle SHALL migrate to it while preserving existing Terraform-visible behavior.

#### Scenario: Security list family resources migrate to the Kibana envelope

- **WHEN** `newResource()` is called for `internal/kibana/securitylist`, `internal/kibana/securitylistitem`, `internal/kibana/securityexceptionlist`, and `internal/kibana/security_list_data_streams`
- **THEN** each returned resource SHALL embed `*entitycore.KibanaResource[...]`
- **AND** each resource SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** any existing wrapper-level interfaces such as `resource.ResourceWithImportState` SHALL remain implemented where they were implemented before migration
- **AND** the migration SHALL preserve the existing Terraform schema shape, state ID format, and import behavior for each resource
