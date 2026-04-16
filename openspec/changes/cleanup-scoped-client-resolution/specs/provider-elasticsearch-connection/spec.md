## MODIFIED Requirements

### Requirement: Fixture ownership inventory (REQ-006)
For provider-level registry completeness, `provider/elasticsearch_connection_schema_test.go` SHALL own every registered SDK and Plugin Framework entity whose type name starts with `elasticstack_elasticsearch_`. When a registered Elasticsearch entity remains excluded from the `elasticsearch_connection` contract, that fixture SHALL still classify it explicitly as Elasticsearch-owned so the completeness check can distinguish intentional exclusion from missing coverage.

#### Scenario: Registered Elasticsearch entity is owned by the Elasticsearch fixture
- **WHEN** a resource or data source registered by `provider.New(...)` or `provider.NewFrameworkProvider(...)` has type name `elasticstack_elasticsearch_*`
- **THEN** `provider/elasticsearch_connection_schema_test.go` SHALL classify that entity into its owned set for coverage and completeness enforcement

### Requirement: Automated enforcement (REQ-005)
Provider-level coverage SHALL be enforced by automated tests in the provider test suite so that adding a new covered entity without the expected connection classification or schema definition fails continuous integration. The Elasticsearch fixture SHALL participate in the combined fixture partition with `provider/kibana_connection_schema_test.go` so every registered provider entity is accounted for exactly once.

#### Scenario: New registered Elasticsearch entity without fixture ownership fails CI
- **GIVEN** a new `elasticstack_elasticsearch_*` resource or data source is registered
- **WHEN** neither the Elasticsearch fixture nor the combined completeness check accounts for it correctly
- **THEN** the provider unit tests SHALL fail and identify the entity by name
