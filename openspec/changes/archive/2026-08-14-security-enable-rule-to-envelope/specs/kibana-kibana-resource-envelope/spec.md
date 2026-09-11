## MODIFIED Requirements

### Requirement: Kibana resources fully implement envelope CRUD callbacks

Kibana resources that embed `entitycore.KibanaResource[T]` SHALL supply non-placeholder callbacks for `Create`, `Read`, `Update`, and `Delete` via `KibanaResourceOptions`. The wrapper struct SHALL NOT override the envelope's `Create` or `Update` methods. `PlaceholderKibanaWriteCallback` SHALL NOT be used by `security_exception_item`, `security_detection_rule`, `alertingrule`, or `security_enable_rule`.

#### Scenario: Lifecycle dispatch goes through envelope callbacks

- **WHEN** Terraform invokes Create, Read, Update, or Delete on any of the four migrated resources
- **THEN** the envelope's lifecycle dispatcher SHALL invoke the corresponding callback supplied via `KibanaResourceOptions`
- **AND** no wrapper-struct `Create` or `Update` method SHALL be present to shadow the envelope's promoted method
