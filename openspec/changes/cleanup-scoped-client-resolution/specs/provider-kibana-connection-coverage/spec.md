## MODIFIED Requirements

### Requirement: Kibana-fixture entity coverage
`provider/kibana_connection_schema_test.go` SHALL own the registered SDK and Plugin Framework entities whose connection contract is Kibana-derived: all registered `elasticstack_kibana_*` entities, all registered `elasticstack_fleet_*` entities, and `elasticstack_apm_agent_configuration`. For those owned entities, the fixture SHALL continue to assert the expected `kibana_connection` helper equivalence and non-deprecated schema metadata for the entity's implementation style.

#### Scenario: APM agent configuration is covered by the Kibana fixture
- **WHEN** `elasticstack_apm_agent_configuration` is registered by `provider.NewFrameworkProvider(...)`
- **THEN** `provider/kibana_connection_schema_test.go` SHALL include that entity in its owned set and assert the shared `kibana_connection` block contract

### Requirement: Registry partition completeness
When the provider coverage tests enumerate entities from `provider.New(...)` and `provider.NewFrameworkProvider(...)`, every registered entity SHALL be owned by exactly one of `provider/kibana_connection_schema_test.go` or `provider/elasticsearch_connection_schema_test.go`. The automated coverage checks SHALL fail if an entity is not owned by either fixture or is owned by both.

#### Scenario: Newly registered entity is not assigned to a fixture
- **WHEN** a new provider entity is registered and neither connection-schema fixture claims it
- **THEN** the provider coverage tests SHALL fail and identify that entity as uncovered

#### Scenario: Entity is claimed by both fixtures
- **WHEN** the same registered provider entity is classified into both connection-schema fixtures
- **THEN** the provider coverage tests SHALL fail and identify that entity as doubly covered
