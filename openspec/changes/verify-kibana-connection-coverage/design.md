## Context

The repository already protects `elasticsearch_connection` in two ways: provider-level schema coverage tests under `provider/`, and a custom golangci-lint analyzer that proves Elasticsearch entity code reaches sink calls only through approved helper-derived client-resolution paths. `kibana_connection` currently has neither protection.

Because Kibana and Fleet entities use different sink surfaces than Elasticsearch entities, the verification change needs to mirror the same intent without pretending the current Elasticsearch analyzer can simply be reused unchanged.

## Goals / Non-Goals

**Goals:**
- Define provider-level coverage tests that fail when in-scope Kibana or Fleet entities are missing `kibana_connection` or drift from the shared schema helper.
- Define lint enforcement that proves Kibana and Fleet entity code uses approved helper-derived client resolution where `kibana_connection` is in scope.
- Require both protections to run in standard local and CI lint/test workflows.

**Non-Goals:**
- Implementing the `kibana_connection` helper path itself.
- Rolling the block out to specific entities in this change.
- Replacing the existing Elasticsearch analyzer with a fully generic analyzer if a narrower addition is simpler.

## Decisions

Mirror the Elasticsearch schema-coverage pattern with a dedicated provider-level coverage test.
The new test should iterate the provider's registered in-scope Kibana and Fleet entities, assert that `kibana_connection` exists, and verify it exactly matches the shared helper output. This keeps the test story consistent with `provider/elasticsearch_connection_schema_test.go`.

Alternative considered: rely on entity-by-entity unit tests only.
Rejected because provider-level enumeration is what catches newly registered entities that forget to add the block.

Add Kibana/Fleet-specific client-resolution lint rather than overloading the Elasticsearch analyzer first.
The repository can reuse the analyzer pattern, but the sink scope for Kibana and Fleet is different: Kibana client getters, Kibana OpenAPI getters, Fleet client getters, and entity methods on scoped `*clients.APIClient`. A dedicated capability keeps the contract clear and avoids conflating Kibana/Fleet sink rules with Elasticsearch sink rules.

Alternative considered: immediately generalize `esclienthelper` into one analyzer for all API-client uses.
Rejected because that expands the implementation surface significantly and is not required to achieve parity with the existing Elasticsearch verification model.

Keep lint enforcement source-based and conservative.
As with the Elasticsearch rule, the lint capability should approve only helper-derived client sources, explicitly allowlisted wrappers, or fact-proven wrappers. If the analyzer cannot prove provenance, it should fail conservatively at the sink.

Alternative considered: lint only for presence of helper calls anywhere in a file.
Rejected because that would not prove the client used at the actual API sink came from the approved helper path.

## Risks / Trade-offs

- [Risk] Kibana and Fleet use several different client sinks, which can make the analyzer scope fuzzy -> Mitigation: define sink scope explicitly in the capability and keep it narrow and type-driven.
- [Risk] Coverage tests may need to make an explicit scoping choice around experimental entities -> Mitigation: align the tests with the same normal provider constructors used for the rollout capability.
- [Risk] Adding multiple analyzers can increase maintenance cost -> Mitigation: keep the lint capability focused on helper provenance and reuse the Elasticsearch analyzer structure where practical.

## Migration Plan

1. Add a new provider-level coverage capability for `kibana_connection`.
2. Add a new Kibana/Fleet client-resolution lint capability modeled after the Elasticsearch lint contract.
3. Wire the coverage test and analyzer(s) into the normal provider test and lint workflows.
4. Use the resulting protections to guard the rollout implementation work.

## Open Questions

- Whether the lint implementation is best delivered as one combined Kibana/Fleet analyzer or as two sibling analyzers can be left to implementation as long as the capability contract is satisfied.
