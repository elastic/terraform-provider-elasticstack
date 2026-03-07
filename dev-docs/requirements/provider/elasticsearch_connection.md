# `elasticstack` provider Elasticsearch entities — `elasticsearch_connection` coverage requirements

Provider implementation: `provider/provider.go`, `provider/plugin_framework.go`

## Scope

This document defines provider-level requirements that ensure all Elasticsearch Terraform entities expose a consistent `elasticsearch_connection` schema.

Exception in scope:
- `elasticstack_elasticsearch_ingest_processor*` data sources are excluded from these coverage requirements because they are schema-construction helpers and do not create or use Elasticsearch clients.

## Requirements

- **[REQ-001] (Coverage / SDK)**: For every SDK resource or data source registered in `provider.New(...)` with type name prefix `elasticstack_elasticsearch_` (excluding `elasticstack_elasticsearch_ingest_processor*` data sources), the entity schema shall define `elasticsearch_connection`.
- **[REQ-002] (Coverage / Framework)**: For every Plugin Framework resource or data source returned by `Provider.Resources(...)` and `Provider.DataSources(...)` with type name prefix `elasticstack_elasticsearch_` (excluding `elasticstack_elasticsearch_ingest_processor*` data sources), the entity schema shall define `elasticsearch_connection`.
- **[REQ-003] (Schema source of truth / SDK)**: For SDK entities covered by REQ-001, the `elasticsearch_connection` schema definition shall be exactly equivalent to `internal/schema.GetEsConnectionSchema("elasticsearch_connection", false)`.
- **[REQ-004] (Schema source of truth / Framework)**: For Framework entities covered by REQ-002, the `elasticsearch_connection` block definition shall be exactly equivalent to `internal/schema.GetEsFWConnectionBlock(false)`.
- **[REQ-005] (Consistency)**: This provider-level coverage requirement shall be enforced through automated tests in the provider test suite so that adding a new `elasticstack_elasticsearch_` entity without the expected `elasticsearch_connection` definition fails CI.

## Acceptance criteria

- **[AC-001] (SDK unit test)**: Given a provider from `provider.New("dev")`, when iterating SDK `ResourcesMap` and `DataSourcesMap`, then each covered `elasticstack_elasticsearch_` entity (excluding `elasticstack_elasticsearch_ingest_processor*` data sources) shall run as its own subtest and assert:
  - `elasticsearch_connection` exists in schema.
  - Its schema is exactly equal to `internal/schema.GetEsConnectionSchema("elasticsearch_connection", false)`.
- **[AC-002] (Framework unit test)**: Given a provider from `provider.NewFrameworkProvider("dev")`, when iterating framework resources and data sources and resolving entity type names from metadata, then each covered `elasticstack_elasticsearch_` entity (excluding `elasticstack_elasticsearch_ingest_processor*` data sources) shall run as its own subtest and assert:
  - `elasticsearch_connection` exists in schema blocks.
  - Its block definition is exactly equal to `internal/schema.GetEsFWConnectionBlock(false)`.
- **[AC-003] (Regression behavior)**: When a new `elasticstack_elasticsearch_` entity is added without a matching `elasticsearch_connection` definition, the corresponding subtest shall fail and identify that entity by name.
