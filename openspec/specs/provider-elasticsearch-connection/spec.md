# `provider-elasticsearch-connection` — Elasticsearch entity connection schema coverage

Provider implementation: `provider/provider.go`, `provider/plugin_framework.go`  
Schema helpers: `internal/schema` (`GetEsConnectionSchema`, `GetEsFWConnectionBlock`)

## Purpose

Define provider-level requirements so every Elasticsearch Terraform entity exposes a consistent `elasticsearch_connection` schema (SDK and Plugin Framework), with automated tests that enforce coverage in CI.

**In scope:** All `elasticstack_elasticsearch_*` resources and data sources registered on the provider.

**Out of scope:** `elasticstack_elasticsearch_ingest_processor*` data sources — they are schema-construction helpers and do not create or use Elasticsearch clients; they are excluded from coverage requirements.
## Requirements
### Requirement: SDK entity coverage (REQ-001)

For every SDK resource or data source registered in `provider.New(...)` whose type name has prefix `elasticstack_elasticsearch_`, except `elasticstack_elasticsearch_ingest_processor*` data sources, the entity schema SHALL define `elasticsearch_connection`.

#### Scenario: Ingest processor data sources are excluded

- GIVEN an SDK data source with type name matching `elasticstack_elasticsearch_ingest_processor*`
- WHEN evaluating SDK coverage requirements
- THEN that data source SHALL NOT be required to define `elasticsearch_connection`

### Requirement: Plugin Framework entity coverage (REQ-002)

For every Plugin Framework resource or data source returned by `Provider.Resources(...)` and `Provider.DataSources(...)` whose type name has prefix `elasticstack_elasticsearch_`, except `elasticstack_elasticsearch_ingest_processor*` data sources, the entity schema SHALL define `elasticsearch_connection`.

#### Scenario: Ingest processor data sources are excluded from Framework coverage

- GIVEN a Framework data source with type name matching `elasticstack_elasticsearch_ingest_processor*`
- WHEN evaluating Framework coverage requirements
- THEN that data source SHALL NOT be required to define `elasticsearch_connection`

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

### Requirement: Regression identification (AC-003)

When a covered entity lacks a matching `elasticsearch_connection` definition, the corresponding subtest SHALL fail and SHALL identify that entity by name.

#### Scenario: Failure names the offending entity

- GIVEN a regression where a covered entity is missing or mismatched `elasticsearch_connection`
- WHEN the connection coverage subtest for that entity runs
- THEN the failure output SHALL identify the entity by its type name

