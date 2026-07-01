## MODIFIED Requirements

### Requirement: Shared modeling package adoption (Phase 1 refactor)

The `elasticstack_fleet_integration_policy` resource implementation (`internal/fleet/integration_policy/`) SHALL be refactored to import its inputs, streams, and vars typed modeling from the new shared package `internal/fleet/policyshape/` rather than defining those types inline. This refactor SHALL be behaviour-preserving: no user-visible schema change, no change to the persisted state format, and no change to the acceptance test results.

The following types and helpers SHALL be migrated to `policyshape` and re-exported or directly imported:
- `InputType` and associated value type
- `InputsType` and associated value type
- `VarsJsonType` and associated value type
- Defaults merging logic (`models_defaults.go`)
- Canonical JSON normalization logic
- Secret helpers (`secrets.go`)

The `integration_policy` package MAY retain thin wrappers or type aliases to avoid large import-path changes in existing callers within the same package, at implementer discretion.

#### Scenario: Schema unchanged after refactor
- **WHEN** the Phase 1 refactor is applied
- **THEN** `terraform providers schema -json` SHALL produce identical output for `elasticstack_fleet_integration_policy` before and after the refactor
- **AND** no plan diff SHALL appear for any existing `elasticstack_fleet_integration_policy` resource in state

#### Scenario: Acceptance tests pass after refactor
- **WHEN** the integration_policy acceptance tests are run against a compatible Kibana after Phase 1
- **THEN** all tests that passed before the refactor SHALL continue to pass
- **AND** no new test failures SHALL be introduced by the refactor

### Requirement: Full parity gate before Phase 2

Phase 2 (the new `elasticstack_fleet_agentless_policy` resource) SHALL NOT be merged until Phase 1 has passed acceptance test parity. The task list enforces this sequencing.

#### Scenario: Phase 1 acceptance tests gate Phase 2
- **WHEN** the refactor is under review
- **THEN** integration_policy acceptance tests MUST be verified passing before Phase 2 implementation begins
