## Context

The canonical `provider-elasticsearch-connection` spec currently guarantees that covered Elasticsearch entities expose `elasticsearch_connection` and that their SDK and Plugin Framework definitions stay exactly aligned with the shared helper functions in `internal/schema/connection.go`. Those helper functions still attach a deprecation warning when called for entity schemas, so the current implementation can satisfy the existing spec while still surfacing a warning the provider no longer wants to show.

This change is limited to the entity-facing `elasticsearch_connection` schema used by covered Elasticsearch resources and data sources. It does not redesign the connection block, add or remove fields, or change how provider-level configuration is modeled.

## Goals / Non-Goals

**Goals:**
- Define spec-level behavior that covered entity `elasticsearch_connection` definitions are not deprecated.
- Preserve the existing helper-based source of truth so SDK and Plugin Framework entities still derive from shared schema builders.
- Require automated coverage tests to assert the absence of a deprecation warning in addition to presence and helper equivalence.

**Non-Goals:**
- Changing field names, validation, defaults, or authentication behavior inside the Elasticsearch connection schema.
- Expanding this change to Kibana or Fleet connection helpers.
- Reworking provider configuration semantics beyond keeping this change focused on entity schemas.

## Decisions

Treat warning removal as a requirement change in the existing `provider-elasticsearch-connection` capability.
The current spec already owns the contract for how covered entities expose `elasticsearch_connection`, so the new behavior belongs in that capability instead of a separate spec. This keeps helper equivalence and warning behavior in one place.

Alternative considered: create a new capability just for deprecation messaging.
Rejected because the warning is part of the same connection-schema contract already defined by `provider-elasticsearch-connection`.

Keep the helper functions as the schema source of truth while changing their entity behavior.
The implementation should continue using `GetEsConnectionSchema("elasticsearch_connection", false)` and `GetEsFWConnectionBlock(false)` as the canonical definitions for covered entities, but those helpers should stop assigning entity-level deprecation metadata. That lets the existing equality-based tests continue to guard drift.

Alternative considered: allow entities to override deprecation metadata outside the helpers.
Rejected because it would weaken the single-source-of-truth guarantee and make cross-entity drift easier to introduce.

Extend acceptance requirements rather than relying on helper equality alone.
Exact equality to the helpers is necessary but not sufficient for this behavioral change because the helpers themselves are what currently emit the warning. The acceptance criteria should explicitly require tests to assert that covered entity schemas have no deprecation message.

Alternative considered: rely only on updated helper equality checks.
Rejected because the spec should state the user-visible behavior directly, not only imply it through helper implementation details.

## Risks / Trade-offs

- Narrowing the helpers to remove entity deprecation metadata could unintentionally affect provider configuration if the behavior is not scoped carefully -> Keep the change limited to entity-facing helper output and preserve existing provider-configuration semantics.
- Additional test assertions may need different checks for SDK and Plugin Framework schema types -> Reuse the existing coverage test structure and add explicit assertions for each schema representation.
- The spec could become redundant if both helper equivalence and no-warning behavior are repeated carelessly -> Update only the requirement blocks that define source-of-truth and acceptance behavior so the delta stays focused.

## Migration Plan

- Update the `provider-elasticsearch-connection` delta spec to require non-deprecated entity connection schemas and corresponding acceptance assertions.
- Remove the entity-level deprecation warning metadata from the Elasticsearch connection helper output in `internal/schema/connection.go`.
- Update the provider connection coverage tests so SDK and Plugin Framework entities assert both helper equivalence and the absence of a deprecation warning.
- Validate the OpenSpec change after the delta is complete.

## Open Questions

- None. The desired behavior is to stop surfacing the entity-level deprecation warning while keeping the shared Elasticsearch connection schema pattern intact.
