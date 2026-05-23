## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL continue to provide a generic constructor `NewKibanaResource[T]()` for Kibana-backed Terraform resources, and additional Kibana resources that follow the envelope lifecycle SHALL migrate to it while preserving existing Terraform-visible behavior.

#### Scenario: Small Kibana control resources migrate to the Kibana envelope

- **WHEN** `newResource()` is called for `internal/kibana/defaultdataview`, `internal/kibana/security_enable_rule`, and `internal/kibana/prebuilt_rules`
- **THEN** each returned resource SHALL embed `*entitycore.KibanaResource[...]`
- **AND** each resource SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** `internal/kibana/prebuilt_rules` SHALL continue to implement `resource.ResourceWithModifyPlan` on the wrapper
- **AND** the migration SHALL preserve the existing Terraform schema shape, state behavior, and delete semantics for each resource
