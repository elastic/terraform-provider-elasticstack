## Context

The repository already protects `elasticsearch_connection` with provider-level schema coverage tests under `provider/`. `kibana_connection` currently does not have equivalent provider-wide verification, which leaves newly registered Kibana or Fleet entities free to omit the block or drift from the shared helper unnoticed.

## Goals / Non-Goals

**Goals:**
- Define provider-level coverage tests that fail when in-scope Kibana or Fleet entities are missing `kibana_connection` or drift from the shared schema helper.
- Require the coverage protection to run in standard local and CI test workflows.

**Non-Goals:**
- Implementing the `kibana_connection` helper path itself.
- Rolling the block out to specific entities in this change.
- Adding Kibana or Fleet client-resolution lint enforcement; that behavior will be proposed separately.

## Decisions

Mirror the Elasticsearch schema-coverage pattern with a dedicated provider-level coverage test.
The new test should iterate the provider's registered in-scope Kibana and Fleet entities, assert that `kibana_connection` exists, and verify it exactly matches the shared helper output. This keeps the test story consistent with `provider/elasticsearch_connection_schema_test.go`.

Alternative considered: rely on entity-by-entity unit tests only.
Rejected because provider-level enumeration is what catches newly registered entities that forget to add the block.

## Risks / Trade-offs

- [Risk] Coverage tests may need to make an explicit scoping choice around experimental entities -> Mitigation: align the tests with the same normal provider constructors used for the rollout capability.

## Migration Plan

1. Add a new provider-level coverage capability for `kibana_connection`.
2. Wire the coverage tests into the normal provider test workflows.
3. Use the resulting protections to guard the rollout implementation work.

## Open Questions

<!-- None. -->
