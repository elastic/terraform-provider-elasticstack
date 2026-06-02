## MODIFIED Requirements

### Requirement: Focused package-policy envelope and fixed package identity (REQ-003)

The resource SHALL expose a familiar package-policy envelope with `id`, `policy_id`, `name`,
`namespace`, `agent_policy_id`, `agent_policy_ids`, `description`, `enabled`, `force`,
`integration_version`, and `space_ids`. The resource SHALL always target package name `endpoint`
and SHALL NOT expose a user-configurable `integration_name`. The resource SHALL NOT expose the
generic `vars_json`, generic `inputs`, generic `streams`, or `output_id` surfaces from
`elasticstack_fleet_integration_policy` in v1.

#### Scenario: Package name is fixed to Elastic Defend

- GIVEN a valid `elasticstack_fleet_elastic_defend_integration_policy` configuration
- WHEN create or update builds the API request
- THEN the request body SHALL target package name `endpoint`
- AND there SHALL be no user-configurable `integration_name` in the Terraform schema

## ADDED Requirements

### Requirement: Support multiple agent policies via `agent_policy_ids` (REQ-014)

The resource SHALL expose an `agent_policy_ids` Optional `list(string)` attribute that allows a
single Elastic Defend package policy to be linked to multiple agent policies. This attribute SHALL
conflict with `agent_policy_id`; the two attributes MUST NOT be set together. The `agent_policy_ids`
attribute SHALL require at least one element when set. Use of `agent_policy_ids` SHALL be
runtime-gated on Elastic Stack ≥ 8.15.0.

`agent_policy_id` SHALL change from Required to Optional. Existing configurations using
`agent_policy_id` SHALL continue to work without modification. No schema version bump or state
upgrader is required.

When `agent_policy_ids` is configured, the provider SHALL set `PolicyIds` to the full list and
`PolicyId` to the first element (for compatibility with older Kibana artifact-manifest endpoints)
in both the bootstrap and finalize request phases.

On read, the provider SHALL populate whichever of `agent_policy_id` or `agent_policy_ids` was
originally in state, to avoid Terraform detecting inconsistent results.

#### Scenario: Single agent policy — unchanged behavior

- GIVEN a configuration with `agent_policy_id = "policy-abc"`
- WHEN create or update runs
- THEN the behavior SHALL be identical to the pre-change resource
- AND the request SHALL set `PolicyId = "policy-abc"` and leave `PolicyIds` unset

#### Scenario: Multiple agent policies via `agent_policy_ids`

- GIVEN a configuration with `agent_policy_ids = ["policy-abc", "policy-xyz"]`
- WHEN create runs on Elastic Stack ≥ 8.15.0
- THEN the bootstrap request SHALL set `PolicyIds = ["policy-abc", "policy-xyz"]` and
  `PolicyId = "policy-abc"` (first element, for compatibility)
- AND the finalize request SHALL set the same fields
- AND the created package policy SHALL be associated with both agent policies
- AND a subsequent read SHALL populate `agent_policy_ids` from the API response `PolicyIds` field

#### Scenario: `agent_policy_ids` rejected below minimum stack version

- GIVEN a configuration with `agent_policy_ids = ["policy-abc"]`
- WHEN create or update runs against Elastic Stack < 8.15.0
- THEN the provider SHALL return an error diagnostic stating that `agent_policy_ids` requires
  Elastic Stack ≥ 8.15.0
- AND no package policy SHALL be created or modified

#### Scenario: `agent_policy_ids` and `agent_policy_id` conflict

- GIVEN a configuration that sets both `agent_policy_id` and `agent_policy_ids`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error
- AND no CRUD operation SHALL be attempted

#### Scenario: `agent_policy_ids` must have at least one element

- GIVEN a configuration with `agent_policy_ids = []`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation error
