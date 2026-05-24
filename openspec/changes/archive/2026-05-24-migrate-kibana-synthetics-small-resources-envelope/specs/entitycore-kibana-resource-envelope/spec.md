## MODIFIED Requirements

### Requirement: Envelope constructor produces shared Kibana resource behavior

The system SHALL provide a generic constructor `NewKibanaResource[T]()` that accepts a `KibanaResourceOptions[T]` options struct (not a positional callback list) and returns an envelope owning shared Kibana resource behavior. The envelope SHALL provide Metadata, Schema, Configure, Create, Read, Update, and Delete behavior, and SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`. Concrete resources SHALL embed the envelope and may choose to implement additional Plugin Framework interfaces such as ImportState or state upgrade support.

#### Scenario: Constructor returns complete resource envelope

- **WHEN** `NewKibanaResource[T](component, name, opts)` is called with a `KibanaResourceOptions[T]` containing non-nil required callbacks
- **THEN** the returned value SHALL satisfy `resource.Resource` and `resource.ResourceWithConfigure`
- **AND** the returned value SHALL NOT satisfy `resource.ResourceWithImportState`

#### Scenario: Metadata builds the Terraform type name

- **WHEN** an envelope is constructed via `NewKibanaResource[T](ComponentKibana, "maintenance_window", opts)`
- **THEN** its `Metadata` SHALL set the type name to `<provider_type_name>_kibana_maintenance_window`

#### Scenario: Synthetics migration uses the envelope without changing behavior

- **WHEN** the synthetics resources are migrated to embed `*entitycore.KibanaResource[...]` returned by `NewKibanaResource`
- **THEN** the resources SHALL continue to preserve their schema, import behavior, and Terraform state identity behavior
- **AND** the resources SHALL remain usable as Terraform `resource.Resource` and `resource.ResourceWithConfigure` implementations
