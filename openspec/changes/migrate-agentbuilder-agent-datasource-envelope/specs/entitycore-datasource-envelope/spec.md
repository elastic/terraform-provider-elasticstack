## ADDED Requirements

### Requirement: Kibana envelope enforces optional model version requirements

The Kibana data source envelope SHALL allow a decoded model to optionally declare pre-read server version requirements. When the model implements the optional version-requirements interface, the envelope SHALL evaluate those requirements after resolving the scoped Kibana client and before invoking the concrete read function.

Version requirements SHALL remain optional. A Kibana data source model that only satisfies the base `KibanaDataSourceModel` contract SHALL continue through the existing read flow without defining no-op version requirements.

#### Scenario: Model without version requirements reads normally

- **GIVEN** a Kibana envelope data source whose model does not implement the optional version-requirements interface
- **WHEN** `Read` successfully decodes config and resolves the scoped Kibana client
- **THEN** the envelope SHALL invoke the concrete read function without attempting model-specific version enforcement
- **AND** state persistence SHALL follow the existing envelope behavior

#### Scenario: Supported server invokes read function

- **GIVEN** a Kibana envelope data source whose model declares a minimum server version requirement
- **AND** the scoped Kibana client reports that the server satisfies that minimum version
- **WHEN** `Read` evaluates the version requirement
- **THEN** the envelope SHALL invoke the concrete read function
- **AND** the read result SHALL be used for state persistence according to existing envelope behavior

#### Scenario: Unsupported server stops before read function

- **GIVEN** a Kibana envelope data source whose model declares a minimum server version requirement with an error message
- **AND** the scoped Kibana client reports that the server does not satisfy that minimum version
- **WHEN** `Read` evaluates the version requirement
- **THEN** the envelope SHALL add an `Unsupported server version` diagnostic using the model-provided error message
- **AND** the concrete read function SHALL NOT be invoked
- **AND** Terraform state SHALL NOT be set from a read result

#### Scenario: Version requirement diagnostics stop read

- **GIVEN** a Kibana envelope data source whose model implements the optional version-requirements interface
- **AND** collecting or enforcing the requirements returns error diagnostics
- **WHEN** `Read` evaluates version requirements
- **THEN** the envelope SHALL append those diagnostics to the read response
- **AND** the concrete read function SHALL NOT be invoked
- **AND** Terraform state SHALL NOT be set from a read result
