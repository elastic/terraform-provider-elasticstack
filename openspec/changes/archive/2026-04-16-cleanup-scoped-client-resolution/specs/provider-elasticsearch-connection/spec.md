## MODIFIED Requirements

### Requirement: Registry-driven Elasticsearch coverage (REQ-006)
For provider-level registry completeness, the provider connection-schema test SHALL enumerate every registered SDK and Plugin Framework entity whose type name starts with `elasticstack_elasticsearch_` and SHALL run a subtest that validates the `elasticsearch_connection` contract for that entity. When a registered Elasticsearch entity remains intentionally excluded from the `elasticsearch_connection` contract, that same test SHALL assert the exception explicitly so the completeness check can distinguish intentional exclusion from missing coverage.

#### Scenario: Registered Elasticsearch entity is validated by the registry-driven test
- **WHEN** a resource or data source registered by `provider.New(...)` or `provider.NewFrameworkProvider(...)` has type name `elasticstack_elasticsearch_*`
- **THEN** the provider connection-schema test SHALL run a subtest for that entity and validate the expected `elasticsearch_connection` contract

#### Scenario: Ingest processor data sources stay explicitly exempt
- **WHEN** a registered SDK data source has type name `elasticstack_elasticsearch_ingest_processor_*`
- **THEN** the provider connection-schema test SHALL still run a subtest for that entity and SHALL assert that both `elasticsearch_connection` and `kibana_connection` are absent

### Requirement: Automated enforcement (REQ-005)
Provider-level coverage SHALL be enforced by automated tests in the provider test suite so that adding a new covered entity without the expected connection classification or schema definition fails continuous integration. The registry-driven provider connection-schema test SHALL participate in the final completeness assertion so every registered provider entity is accounted for.

#### Scenario: New registered Elasticsearch entity without validation fails CI
- **GIVEN** a new `elasticstack_elasticsearch_*` resource or data source is registered
- **WHEN** the provider connection-schema test does not validate it correctly
- **THEN** the provider unit tests SHALL fail and identify the entity by name
