## ADDED Requirements

### Requirement: Resource type supports remote Elasticsearch output
The `elasticstack_fleet_output` resource SHALL accept `remote_elasticsearch` as a valid `type` value in addition to existing supported resource types.

#### Scenario: Configure remote Elasticsearch output type
- **WHEN** a resource configuration sets `type = "remote_elasticsearch"`
- **THEN** schema validation SHALL accept the type value

### Requirement: Remote Elasticsearch authentication is required and sensitive
When `type = "remote_elasticsearch"`, the resource SHALL require service-token based authentication input and SHALL treat authentication material as sensitive state.

#### Scenario: Missing remote authentication
- **WHEN** a resource configuration sets `type = "remote_elasticsearch"` without required service token authentication
- **THEN** the provider SHALL return a validation or API diagnostic and SHALL NOT complete create

#### Scenario: Remote authentication stored as sensitive
- **WHEN** a resource with `type = "remote_elasticsearch"` is planned or applied
- **THEN** service token authentication fields SHALL be marked sensitive in Terraform schema/state

### Requirement: Remote Elasticsearch optional TLS and mTLS settings are supported
When `type = "remote_elasticsearch"`, the resource SHALL accept optional TLS and client-certificate settings supported by Fleet output APIs and SHALL map them into API requests.

#### Scenario: Configure TLS certificate authorities
- **WHEN** a remote Elasticsearch output configuration includes certificate authority settings
- **THEN** the provider SHALL send those TLS settings in create/update requests

#### Scenario: Configure client certificate and key
- **WHEN** a remote Elasticsearch output configuration includes client certificate and key material
- **THEN** the provider SHALL send mTLS settings in create/update requests

### Requirement: Remote Elasticsearch CRUD and read mapping behavior
The resource SHALL create, read, update, and delete remote Elasticsearch outputs through Fleet output APIs, and SHALL map remote Elasticsearch responses back into Terraform state using the same identity and space-context behavior as other output types.

#### Scenario: Read remote Elasticsearch output
- **WHEN** read is executed for a resource whose Fleet output type is remote Elasticsearch
- **THEN** the provider SHALL map common output fields and remote Elasticsearch-specific fields into state

### Requirement: Secret-preserving read behavior for remote output fields
For `remote_elasticsearch` outputs, when Fleet read responses omit or redact secret fields, the provider SHALL preserve previously configured secret values from state to avoid unintended drift.

#### Scenario: Secret redacted in API response
- **WHEN** a read response for a remote Elasticsearch output omits a configured secret authentication field
- **THEN** the provider SHALL retain the prior state value for that secret field

### Requirement: Remote Elasticsearch integration sync and related toggles
When `type = "remote_elasticsearch"`, the resource SHALL expose attributes aligned with Fleet for automatic integration synchronization and related remote-output options (`sync_integrations`, `sync_uninstalled_integrations`, `write_to_logs_streams` where supported by the API), and SHALL map them on create, read, and update.

#### Scenario: Configure integration sync on remote output
- **WHEN** a resource sets `type = "remote_elasticsearch"` and configures sync-related boolean attributes
- **THEN** the provider SHALL send and read back those fields through Fleet output APIs

#### Scenario: Sync fields only valid for remote type
- **WHEN** `type` is not `remote_elasticsearch` and sync-related attributes are set
- **THEN** schema validation SHALL reject the configuration
