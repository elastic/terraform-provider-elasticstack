## ADDED Requirements

### Requirement: Shared package structure

The shared package `internal/fleet/policyshape/` (working name; final name at implementer discretion) SHALL provide reusable Plugin Framework types and helpers for modeling Fleet package policy inputs, streams, and vars. The package SHALL be importable by both `internal/fleet/integration_policy/` and `internal/fleet/agentlesspolicy/` without creating import cycles.

#### Scenario: Package importable from multiple resources

- **WHEN** both `internal/fleet/integration_policy/` and `internal/fleet/agentlesspolicy/` import `internal/fleet/policyshape/`
- **THEN** `go build ./...` SHALL succeed with no import cycle errors

### Requirement: VarsJsonType — JSON-encoded vars

The package SHALL provide `VarsJsonType`, a Plugin Framework custom type for the top-level (integration-level) JSON-encoded vars attribute (`vars_json`). It SHALL:

- Accept any valid JSON object string in config.
- Normalize the JSON on read (stable key ordering, no extra whitespace) so that semantically equivalent JSON strings produce no plan diff.
- Be marked `Sensitive` so values are redacted from plan output.
- Support `UseStateForUnknown` so computed values are preserved across plan/apply cycles.

Input-level and stream-level vars use the `vars` attribute key (matching the existing `integration_policy` schema) with a normalized JSON string type; they do not use the contextual-defaults `VarsJsonType`.

> **Naming note:** the API field is `vars` at all levels. The Terraform top-level attribute is named `vars_json` to signal JSON encoding; input/stream attributes keep the existing `vars` name used by `integration_policy` so the shared type is behaviour-preserving for that resource.

#### Scenario: Semantically equivalent JSON produces no diff

- **WHEN** the API returns `{"b":2,"a":1}` and the config contains `jsonencode({a=1, b=2})`
- **THEN** the normalized forms SHALL be equal
- **AND** Terraform SHALL NOT produce a plan diff for the vars attribute

#### Scenario: Changed JSON produces a diff

- **WHEN** the API returns `{"a":1}` and the config changes to `jsonencode({a=2})`
- **THEN** Terraform SHALL produce a plan diff showing the vars change

### Requirement: InputType — single input config

The package SHALL provide `InputType`, a Plugin Framework custom type representing a single input configuration object. It SHALL model:

- `enabled` — Optional+Computed bool (default true).
- `condition` — Optional string (agent condition expression).
- `vars` — Optional sensitive JSON string (input-level variables; normalized on read). Named `vars` to match the existing `integration_policy` schema.
- `streams` — Optional map of stream objects, each with `enabled`, optionally `condition`, and `vars`.

> **Note:** `var_group_selections` is modeled at the top level (on the package policy) only. Per-stream `var_group_selections` is API-supported but intentionally **not** modeled in v1 (deferred); per-input is not supported by the simplified request format this provider uses (legacy typed-input format only).

#### Scenario: Disabled input round-trips correctly

- **WHEN** `inputs = { "my/input" = { enabled = false } }` is set in config
- **THEN** the API request SHALL include `"enabled": false` for that input
- **AND** state SHALL contain `enabled = false` after read

#### Scenario: Nested stream vars round-trip

- **WHEN** `inputs["my/input"].streams["my.stream"].vars = jsonencode({key = "val"})` is set
- **THEN** the API request SHALL include the stream vars
- **AND** the normalized value SHALL be preserved in state after read

### Requirement: InputsType — map of inputs

The package SHALL provide `InputsType`, a Plugin Framework custom type for the top-level `inputs` map attribute, keyed by input type ID string. It SHALL use `InputType` for each element's value type.

#### Scenario: Multiple inputs in config

- **WHEN** config contains two inputs with different input type IDs
- **THEN** both entries SHALL appear in the API request body
- **AND** both entries SHALL be preserved in state after read

### Requirement: Defaults merging

The package SHALL provide a defaults merging function that merges package-supplied default values (from the Fleet integration package metadata) into user-supplied input and stream configurations. User-supplied values SHALL take precedence over defaults.

#### Scenario: User value overrides default

- **WHEN** the package default for a var is `"us-east-1"` and the user sets `"eu-west-1"`
- **THEN** the merged value SHALL be `"eu-west-1"`

#### Scenario: Missing user value uses default

- **WHEN** the package default for a var is `"us-east-1"` and the user does not set the var
- **THEN** the merged value SHALL be `"us-east-1"`

### Requirement: Canonical JSON normalization

The package SHALL provide a normalization function that, given a JSON string, returns the semantically equivalent JSON string with stable key ordering (alphabetical) and no extraneous whitespace.

#### Scenario: Normalization is idempotent

- **WHEN** normalization is applied twice to the same input
- **THEN** the result SHALL be the same both times

#### Scenario: Normalization produces stable ordering

- **WHEN** the input JSON has keys in arbitrary order
- **THEN** the output SHALL have keys in alphabetical order

### Requirement: Secret helpers

The package SHALL provide helpers for detecting and handling secret-bearing vars (vars whose API representation includes a `secret_ref` instead of a plain value). The helpers SHALL:

- Detect whether a var value is a secret reference (i.e., contains `{id, isSecretRef}`).
- Preserve existing secret references when the user has not supplied a new value.

#### Scenario: Secret reference preserved on update

- **WHEN** a var was previously set as a secret and the user does not supply a new value for it on update
- **THEN** the secret reference SHALL be preserved in the update request body
- **AND** the raw secret value SHALL NOT appear in state or plan output
