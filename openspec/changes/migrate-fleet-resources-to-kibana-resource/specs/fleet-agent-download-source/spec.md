## MODIFIED Requirements

### Requirement: Minimum Kibana version guard

The resource SHALL be guarded by a minimum Kibana/Fleet version that supports the Agent Binary Download Sources API. If the connected Kibana is below that version, the provider SHALL emit a clear diagnostic during Create, Read, or Update indicating that the resource is not supported. Version enforcement is not required on Delete.

#### Scenario: Unsupported stack version

- **WHEN** the connected Kibana is below the supported minimum version
- **AND** the operation is Create, Read, or Update
- **THEN** the provider SHALL emit a diagnostic that the resource is not supported on this version

#### Scenario: Delete on unsupported version

- **WHEN** the connected Kibana is below the supported minimum version
- **AND** the operation is Delete
- **THEN** the provider SHALL proceed with the delete without a version diagnostic
