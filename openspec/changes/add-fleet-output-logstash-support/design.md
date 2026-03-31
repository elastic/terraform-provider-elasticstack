## Context

`elasticstack_fleet_output` already supports multiple Fleet output types in schema and state mapping patterns, but current implementation paths for `logstash` are incomplete compared to `elasticsearch` and `kafka`. This creates a behavior gap where provider users cannot reliably manage all Fleet-supported output types from Terraform.

The change targets the Fleet output resource in `internal/fleet/output` and related tests/docs. Existing patterns for type-dispatched request/state mapping and acceptance coverage will be reused to minimize risk.

## Goals / Non-Goals

**Goals:**
- Enable full CRUD behavior for `type = "logstash"` in the Fleet output resource.
- Ensure request construction and read/state mapping handle `logstash` consistently with other supported types.
- Add automated tests that prevent regressions for `logstash` create/update/read/import flows.
- Keep documentation aligned with provider behavior for Fleet outputs.

**Non-Goals:**
- Adding support for new output types beyond `logstash`.
- Expanding Terraform schema with every optional Logstash tuning field from Elastic Agent documentation in this change.
- Refactoring unrelated Fleet output code paths.

## Decisions

- Reuse existing output type dispatch architecture and add/complete `logstash` branches for create/update payload mapping and read conversion.
  - Rationale: this keeps behavior consistent with current `elasticsearch`/`kafka` handling and reduces maintenance complexity.
  - Alternative considered: introducing a new generic dynamic type mapping layer; rejected as unnecessary scope for a targeted capability addition.

- Add explicit acceptance coverage for `logstash` lifecycle scenarios using the existing Fleet test harness.
  - Rationale: resource behavior issues often appear only against real API responses; acceptance tests provide confidence for provider users.
  - Alternative considered: unit-only coverage; rejected because it can miss API contract mismatches.

- Keep schema changes minimal and focused on making `logstash` usable with existing common output fields (`name`, `type`, `hosts`, SSL/common fields).
  - Rationale: minimizes backward compatibility risk while closing the functional gap requested.
  - Alternative considered: adding many Logstash-specific configuration fields now; deferred to a follow-up change if needed.

## Risks / Trade-offs

- [API model mismatch for `logstash` requests or responses] -> Mitigation: add unit tests around mapping helpers and acceptance tests for end-to-end behavior.
- [Behavior divergence with existing `fleet-output` spec wording] -> Mitigation: include delta spec updates that explicitly capture `logstash` support expectations.
- [Insufficient test environment support for Logstash-specific endpoints] -> Mitigation: scope tests to Fleet output API behavior and common fields required for a valid `logstash` output definition.
