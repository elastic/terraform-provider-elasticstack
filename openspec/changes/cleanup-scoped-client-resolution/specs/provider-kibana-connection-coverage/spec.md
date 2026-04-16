## MODIFIED Requirements

### Requirement: Registry-driven provider connection coverage
The provider test suite SHALL cover registered non-Elasticsearch entities with a single registry-driven connection-schema test. That test SHALL enumerate entities registered by `provider.New(...)` and `provider.NewFrameworkProvider(...)`, run one subtest per registered entity, and for each registered non-Elasticsearch entity SHALL assert the expected `kibana_connection` helper equivalence and non-deprecated schema metadata for the entity's implementation style.

#### Scenario: APM agent configuration is covered by the registry-driven test
- **WHEN** `elasticstack_apm_agent_configuration` is registered by `provider.NewFrameworkProvider(...)`
- **THEN** the provider connection-schema test SHALL run a subtest for that entity and assert the shared `kibana_connection` block contract

### Requirement: Registry completeness
When the provider connection-schema test enumerates entities from `provider.New(...)` and `provider.NewFrameworkProvider(...)`, it SHALL track which registered entities were validated and SHALL finish with a completeness assertion that fails if any registered entity was not validated.

#### Scenario: Newly registered entity is not validated by the test
- **WHEN** a new provider entity is registered and the provider connection-schema test does not validate it
- **THEN** the provider test suite SHALL fail and identify that entity as unvalidated
