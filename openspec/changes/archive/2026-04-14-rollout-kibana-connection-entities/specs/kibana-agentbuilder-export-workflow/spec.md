## MODIFIED Requirements

### Requirement: API, client, and version errors (REQ-002)

When the data source cannot obtain the effective Kibana OpenAPI client, it SHALL return an error diagnostic. By default, the effective client SHALL be the provider's configured Kibana OpenAPI client. When `kibana_connection` is configured on the data source, the effective client SHALL be the scoped Kibana OpenAPI client derived from that block. The data source SHALL also verify that the Elastic Stack version is at least `9.4.0-SNAPSHOT`; if the version is lower, it SHALL fail with an `Unsupported server version` diagnostic. Transport errors and unexpected HTTP statuses from the workflow API SHALL be surfaced as diagnostics.

#### Scenario: Stack below minimum version

- **WHEN** the target Elastic Stack version is below `9.4.0-SNAPSHOT`
- **THEN** the read SHALL fail with an unsupported-version diagnostic before calling the workflow API

#### Scenario: Effective Kibana client unavailable

- **WHEN** the provider or scoped `kibana_connection` cannot supply a Kibana OpenAPI client
- **THEN** the read SHALL fail with an error diagnostic
