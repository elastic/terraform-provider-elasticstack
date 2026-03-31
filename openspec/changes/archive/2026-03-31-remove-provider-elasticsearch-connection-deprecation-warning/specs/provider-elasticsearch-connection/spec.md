## MODIFIED Requirements

### Requirement: SDK schema source of truth (REQ-003)

For SDK entities covered by the SDK coverage requirement, the `elasticsearch_connection` schema definition SHALL be exactly equivalent to `internal/schema.GetEsConnectionSchema("elasticsearch_connection", false)` and SHALL NOT expose a deprecation warning on the entity schema.

#### Scenario: SDK connection attribute matches helper without deprecation

- GIVEN a covered SDK `elasticstack_elasticsearch_*` entity
- WHEN its schema is compared to `GetEsConnectionSchema("elasticsearch_connection", false)` and its deprecation metadata is evaluated
- THEN the definitions SHALL be equal
- AND the `elasticsearch_connection` schema SHALL NOT be deprecated

### Requirement: Plugin Framework block source of truth (REQ-004)

For Framework entities covered by the Framework coverage requirement, the `elasticsearch_connection` block definition SHALL be exactly equivalent to `internal/schema.GetEsFWConnectionBlock(false)` and SHALL NOT expose a deprecation message on the entity block.

#### Scenario: Framework connection block matches helper without deprecation

- GIVEN a covered Framework `elasticstack_elasticsearch_*` entity
- WHEN its `elasticsearch_connection` block is compared to `GetEsFWConnectionBlock(false)` and its deprecation metadata is evaluated
- THEN the definitions SHALL be equal
- AND the `elasticsearch_connection` block SHALL NOT expose a deprecation message

### Requirement: SDK unit test acceptance (AC-001)

Given a provider from `provider.New("dev")`, when iterating SDK `ResourcesMap` and `DataSourcesMap`, each covered `elasticstack_elasticsearch_*` entity (excluding `elasticstack_elasticsearch_ingest_processor*` data sources) SHALL run as its own subtest that asserts `elasticsearch_connection` exists in the schema, that its schema is exactly equal to `internal/schema.GetEsConnectionSchema("elasticsearch_connection", false)`, and that the schema does not expose a deprecation warning.

#### Scenario: SDK subtests assert helper equality and no warning per covered entity

- GIVEN `provider.New("dev")` and each covered SDK Elasticsearch entity
- WHEN the SDK connection coverage tests run
- THEN each entity SHALL have a subtest asserting presence and equality of `elasticsearch_connection` to `GetEsConnectionSchema("elasticsearch_connection", false)`
- AND that subtest SHALL assert the schema is not deprecated

### Requirement: Framework unit test acceptance (AC-002)

Given a provider from `provider.NewFrameworkProvider("dev")`, when iterating framework resources and data sources and resolving entity type names from metadata, each covered `elasticstack_elasticsearch_*` entity (excluding `elasticstack_elasticsearch_ingest_processor*` data sources) SHALL run as its own subtest that asserts `elasticsearch_connection` exists in schema blocks, that its block definition is exactly equal to `internal/schema.GetEsFWConnectionBlock(false)`, and that the block does not expose a deprecation message.

#### Scenario: Framework subtests assert helper equality and no warning per covered entity

- GIVEN `provider.NewFrameworkProvider("dev")` and each covered Framework Elasticsearch entity
- WHEN the Framework connection coverage tests run
- THEN each entity SHALL have a subtest asserting presence in blocks and equality to `GetEsFWConnectionBlock(false)`
- AND that subtest SHALL assert the block exposes no deprecation message
