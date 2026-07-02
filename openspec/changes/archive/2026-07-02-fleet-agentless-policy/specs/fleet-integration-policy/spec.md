## MODIFIED Requirements

### Requirement: Shared modeling package adoption (Phase 1 refactor)

The `elasticstack_fleet_integration_policy` resource implementation (`internal/fleet/integration_policy/`) SHALL be refactored to import its inputs, streams, and vars typed modeling from the new shared package `internal/fleet/policyshape/` rather than defining those types inline. This refactor is behaviour-preserving for all existing attributes: no change to the persisted state format, no change to the acceptance test results, and no removal or rename of any existing attribute.

The following types and helpers SHALL be migrated to `policyshape` and re-exported or directly imported:

- `InputType` and associated value type
- `InputsType` and associated value type
- `VarsJsonType` and associated value type
- Defaults merging logic (`models_defaults.go`)
- Canonical JSON normalization logic
- Secret helpers (`secrets.go`)

The `integration_policy` package MAY retain thin wrappers or type aliases to avoid large import-path changes in existing callers within the same package, at implementer discretion.

The refactor carries one explicit, additive schema change: a `condition` attribute is added to inputs and streams (see the `condition` requirement below). This is non-breaking — no state upgrader is required, and existing configs and state are unaffected.

#### Scenario: Schema parity after refactor (additive `condition` only)

- **WHEN** the Phase 1 refactor is applied
- **THEN** `terraform providers schema -json` for `elasticstack_fleet_integration_policy` SHALL be identical before and after, except for the addition of an Optional `condition` attribute on input and stream elements
- **AND** no plan diff SHALL appear for any existing `elasticstack_fleet_integration_policy` resource in state (the new `condition` attribute is absent from existing state and is treated as unset/null)

#### Scenario: Acceptance tests pass after refactor

- **WHEN** the integration_policy acceptance tests are run against a compatible Kibana after Phase 1
- **THEN** all tests that passed before the refactor SHALL continue to pass
- **AND** no new test failures SHALL be introduced by the refactor

### Requirement: Full parity gate before Phase 2

Phase 2 (the new `elasticstack_fleet_agentless_policy` resource) SHALL NOT be merged until Phase 1 has passed acceptance test parity. The task list enforces this sequencing.

#### Scenario: Phase 1 acceptance tests gate Phase 2

- **WHEN** the refactor is under review
- **THEN** integration_policy acceptance tests MUST be verified passing before Phase 2 implementation begins

### Requirement: Additive `condition` attribute on inputs and streams

The Fleet package policy API exposes a `condition` field (agent condition expression) at three levels: integration-level (top of `KibanaHTTPAPIsSimplifiedCreatePackagePolicyRequest`), per-input (`PackagePolicyRequestMappedInput`), and per-stream (`PackagePolicyRequestMappedInputStream`); it round-trips through the read response (`PackagePolicyMappedInputs`). The existing `integration_policy` schema does not expose `condition` at any level and silently drops it. Phase 1 SHALL close this gap by adding `condition` as a first-class attribute on input and stream elements, sourced from the shared `policyshape` types.

`condition` SHALL be:

- Optional string (agent condition expression to evaluate whether to apply this input/stream).
- Absent from state and treated as unset/null for existing resources (no state upgrader required).
- Sent on Create/Update as the API `condition` field when set; omitted from the request body when unset.
- Read back from the API response when present; set to null in state when the API does not return it.

This is an additive, non-breaking schema change. Existing configs and state are unaffected.

#### Scenario: condition on an input is sent and read back

- **WHEN** an `inputs` block sets `condition = "host.os.family == 'linux'"` on an input element
- **THEN** the create/update request SHALL include that `condition` on the corresponding input
- **AND** state SHALL reflect the `condition` value returned by the API

#### Scenario: condition on a stream is sent and read back

- **WHEN** a stream element sets `condition = "data_stream.dataset == 'audit'"`
- **THEN** the create/update request SHALL include that `condition` on the corresponding stream
- **AND** state SHALL reflect the `condition` value returned by the API

#### Scenario: Existing resources unaffected by condition addition

- **WHEN** an existing `elasticstack_fleet_integration_policy` resource (with no `condition` set in config) is planned after the Phase 1 refactor
- **THEN** no plan diff SHALL appear
- **AND** `condition` SHALL be null in state
- **AND** the request body SHALL omit `condition`

### Requirement: Version gating for `condition`

The Fleet package policy API rejects the `condition` field with an "Additional properties are not allowed" HTTP 400 on Kibana 9.4.0 and 9.4.3 (both released versions at time of writing); it was confirmed empirically to work correctly (round-trips through create/read) on a 9.5.0-SNAPSHOT Kibana. The resource SHALL therefore enforce a minimum Kibana version of 9.5.0 for `condition` using the existing `EnforceMinVersion` pattern (see `internal/fleet/integration_policy/capabilities.go`, mirroring `SupportsPolicyIDs`/`SupportsOutputID`). This is a soft, attribute-scoped gate: it SHALL NOT affect any request that does not set `condition` on any input or stream, regardless of the connected Kibana version.

#### Scenario: Kibana version too old for condition

- **WHEN** the connected Kibana is older than 9.5.0
- **AND** `condition` is set on an input or a stream in the configuration
- **THEN** Create or Update SHALL return an attribute-scoped error diagnostic naming the minimum version (9.5.0)
- **AND** no API call SHALL be made

#### Scenario: Kibana version too old but condition unset

- **WHEN** the connected Kibana is older than 9.5.0
- **AND** no input or stream sets `condition`
- **THEN** Create or Update SHALL proceed normally with no error
- **AND** the request body SHALL omit `condition`, matching pre-existing behavior
