## ADDED Requirements

### Requirement: Kibana entity coverage
For every Terraform resource or data source registered by the provider whose type name has prefix `elasticstack_kibana_`, except entities explicitly excluded by the rollout contract, the entity schema SHALL define `kibana_connection` using the shared provider schema helper for that implementation style.

#### Scenario: New Kibana entity without `kibana_connection` fails coverage
- **WHEN** a covered `elasticstack_kibana_*` entity is registered without the shared `kibana_connection` definition
- **THEN** the provider coverage tests SHALL fail and identify that entity by name

### Requirement: Fleet entity coverage
For every Terraform resource or data source registered by the provider whose type name has prefix `elasticstack_fleet_`, except entities explicitly excluded by the rollout contract, the entity schema SHALL define `kibana_connection` using the shared provider schema helper for that implementation style.

#### Scenario: New Fleet entity without `kibana_connection` fails coverage
- **WHEN** a covered `elasticstack_fleet_*` entity is registered without the shared `kibana_connection` definition
- **THEN** the provider coverage tests SHALL fail and identify that entity by name

### Requirement: Shared-helper equivalence
Covered SDK entities SHALL expose a `kibana_connection` schema exactly equivalent to the shared SDK helper, and covered Plugin Framework entities SHALL expose a `kibana_connection` block exactly equivalent to the shared Framework helper. Covered entity-local definitions SHALL NOT expose deprecation metadata.

#### Scenario: Covered entity matches shared helper
- **WHEN** a covered Kibana or Fleet entity is examined by the provider coverage tests
- **THEN** its `kibana_connection` definition SHALL exactly match the shared helper output for that implementation style

### Requirement: Provider-level automated enforcement
The provider SHALL enforce `kibana_connection` coverage through automated tests that enumerate the covered Kibana and Fleet entities returned by the provider constructors and run per-entity subtests for presence, helper equivalence, and non-deprecated schema or block metadata.

#### Scenario: Coverage tests enforce the contract in normal workflows
- **WHEN** provider unit tests run in local development or CI
- **THEN** the `kibana_connection` coverage tests SHALL fail if a covered Kibana or Fleet entity is missing the expected schema definition
