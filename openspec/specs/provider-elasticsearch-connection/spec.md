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

### Requirement: Automated enforcement (REQ-005)

Provider-level coverage SHALL be enforced by automated tests in the provider test suite so that adding a new covered `elasticstack_elasticsearch_*` entity without the expected `elasticsearch_connection` definition fails continuous integration.

#### Scenario: New entity without connection fails CI

- GIVEN a new `elasticstack_elasticsearch_*` resource or data source is registered without a matching `elasticsearch_connection` definition
- WHEN the provider unit tests run
- THEN the test suite SHALL fail and identify the entity by name

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
