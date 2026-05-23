## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL continue to provide a generic constructor `NewKibanaResource[T]()` for Kibana-backed Terraform resources, and additional Kibana resources that follow the envelope lifecycle SHALL migrate to it while preserving existing Terraform-visible behavior.

#### Scenario: Action connector resource migrates to the Kibana envelope

- **WHEN** `newResource()` is called for `internal/kibana/connectors`
- **THEN** the returned resource SHALL embed `*entitycore.KibanaResource[tfModel]`
- **AND** the resource SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the wrapper SHALL continue to implement `resource.ResourceWithImportState` and `resource.ResourceWithUpgradeState`
- **AND** the migration SHALL preserve the existing Terraform schema shape, composite ID behavior, import behavior, state upgrade behavior, and version-gated connector validation behavior
