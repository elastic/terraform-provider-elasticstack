## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL continue to provide a generic constructor `NewKibanaResource[T]()` for Kibana-backed Terraform resources, and additional Kibana resources that follow the envelope lifecycle SHALL migrate to it while preserving existing Terraform-visible behavior.

#### Scenario: Small synthetics resources migrate to the Kibana envelope

- **WHEN** `newResource()` is called for `internal/kibana/synthetics/parameter` and `internal/kibana/synthetics/privatelocation`
- **THEN** each returned resource SHALL embed `*entitycore.KibanaResource[...]`
- **AND** each resource SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** any existing wrapper-level import support SHALL remain implemented after migration
- **AND** the migration SHALL preserve the existing Terraform schema shape, state ID format, and import behavior for each resource
