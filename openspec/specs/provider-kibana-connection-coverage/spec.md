# provider-kibana-connection-coverage Specification

## Purpose
TBD - created by archiving change verify-kibana-connection-coverage. Update Purpose after archive.
## Requirements
### Requirement: Kibana entity coverage
For every Terraform resource or data source registered by the provider whose type name has prefix `elasticstack_kibana_`, where coverage scope is determined by enumerating the SDK provider returned by `provider.New(...)` through its resource and data source maps and the Plugin Framework provider returned by `provider.NewFrameworkProvider(...)` through its resource and data source enumeration methods, the entity schema SHALL define `kibana_connection` using the shared provider schema helper for that implementation style.

#### Scenario: New Kibana entity without `kibana_connection` fails coverage
- **WHEN** a covered `elasticstack_kibana_*` entity is registered without the shared `kibana_connection` definition
- **THEN** the provider coverage tests SHALL fail and identify that entity by name

#### Scenario: APM agent configuration is covered by the registry-driven test
- **WHEN** `elasticstack_apm_agent_configuration` is registered by `provider.NewFrameworkProvider(...)`
- **THEN** the provider connection-schema test SHALL run a subtest for that entity and assert the shared `kibana_connection` block contract

### Requirement: Fleet entity coverage
For every Terraform resource or data source registered by the provider whose type name has prefix `elasticstack_fleet_`, where coverage scope is determined by enumerating the SDK provider returned by `provider.New(...)` through its resource and data source maps and the Plugin Framework provider returned by `provider.NewFrameworkProvider(...)` through its resource and data source enumeration methods, the entity schema SHALL define `kibana_connection` using the shared provider schema helper for that implementation style.

#### Scenario: New Fleet entity without `kibana_connection` fails coverage
- **WHEN** a covered `elasticstack_fleet_*` entity is registered without the shared `kibana_connection` definition
- **THEN** the provider coverage tests SHALL fail and identify that entity by name

### Requirement: Shared-helper equivalence
Covered SDK entities SHALL expose a `kibana_connection` schema exactly equivalent to `internal/schema.GetKibanaEntityConnectionSchema()`, and covered Plugin Framework entities SHALL expose a `kibana_connection` block exactly equivalent to `internal/schema.GetKbFWConnectionBlock()`. Covered entity-local definitions SHALL NOT expose deprecation metadata.

#### Scenario: Covered entity matches shared helper
- **WHEN** a covered Kibana or Fleet entity is examined by the provider coverage tests
- **THEN** its `kibana_connection` definition SHALL exactly match the shared helper output for that implementation style

### Requirement: Registry completeness
When the provider connection-schema test enumerates entities from `provider.New(...)` and `provider.NewFrameworkProvider(...)`, it SHALL track which registered entities were validated and SHALL finish with a completeness assertion that fails if any registered entity was not validated.

#### Scenario: Newly registered entity is not validated by the test
- **WHEN** a new provider entity is registered and the provider connection-schema test does not validate it
- **THEN** the provider test suite SHALL fail and identify that entity as unvalidated

### Requirement: Provider-level automated enforcement
The provider SHALL enforce `kibana_connection` coverage through automated tests that enumerate the covered Kibana and Fleet entities returned by the provider constructors and run per-entity subtests for presence, helper equivalence, and non-deprecated schema or block metadata.

#### Scenario: Coverage tests enforce the contract in normal workflows
- **WHEN** provider unit tests run in local development or CI
- **THEN** the `kibana_connection` coverage tests SHALL fail if a covered Kibana or Fleet entity is missing the expected schema definition

