# Delta spec: `fleet-elastic-defend-integration-policy` — `additional_datastreams_permissions`

Extends the main spec at `openspec/specs/fleet-elastic-defend-integration-policy/spec.md`.

## MODIFIED Requirements

### Requirement: Focused package-policy envelope and fixed package identity (REQ-003)

The resource SHALL expose a familiar package-policy envelope with `id`, `policy_id`, `name`, `namespace`, `agent_policy_id`, `agent_policy_ids`, `description`, `enabled`, `force`, `integration_version`, `advanced_settings`, `additional_datastreams_permissions`, and `space_ids`. The resource SHALL always target package name `endpoint` and SHALL NOT expose a user-configurable `integration_name`. The resource SHALL NOT expose the generic `vars_json`, generic `inputs`, generic `streams`, or `output_id` surfaces from `elasticstack_fleet_integration_policy` in v1.

#### Scenario: Package name is fixed to Elastic Defend

- GIVEN a valid `elasticstack_fleet_elastic_defend_integration_policy` configuration
- WHEN create or update builds the API request
- THEN the request body SHALL target package name `endpoint`
- AND there SHALL be no user-configurable `integration_name` in the Terraform schema

## ADDED Requirements

### Requirement: additional_datastreams_permissions attribute (REQ-016)

The resource SHALL expose an `additional_datastreams_permissions` attribute of type `list(string)`, `Optional` and `Computed`, with a `UseStateForUnknown` plan modifier. Each element SHALL be a data stream name or pattern (for example `logs-nginx-*`) for which the Defend package policy's agent policy is granted write access in addition to the data streams the `endpoint` package itself declares.

Unlike `elasticstack_fleet_integration_policy`'s attribute of the same name, this attribute SHALL NOT be rejected at plan time when configured as an empty list.

The attribute SHALL be mapped identically on the bootstrap create request and the finalize/update request (the latter is also used by the plain update path): a known value (including an empty list) SHALL always be included in the request body; a null or unknown value SHALL always be omitted.

On read, a non-empty API list SHALL populate the attribute preserving API order. An absent or empty API list SHALL populate the attribute as an empty list when the model already held a known, non-null empty list, and as null otherwise.

#### Scenario: Permissions configured on create

- GIVEN a configuration with `additional_datastreams_permissions = ["logs-nginx-*"]`
- WHEN create runs (bootstrap followed by finalize)
- THEN both the bootstrap and finalize request bodies SHALL include `additional_datastreams_permissions` with that single element
- AND a subsequent read SHALL populate the attribute with that value from the API response

#### Scenario: Permissions updated to a new list

- GIVEN an existing Defend policy with `additional_datastreams_permissions = ["logs-nginx-*"]`
- WHEN the configuration changes to `["logs-nginx-*", "metrics-custom-*"]` and update runs
- THEN the update request body SHALL include both elements
- AND a subsequent read SHALL reflect both elements in state

#### Scenario: Attribute omitted when never configured

- GIVEN a configuration that never sets `additional_datastreams_permissions`
- WHEN create or update runs
- THEN neither the bootstrap, finalize, nor plain update request body SHALL include `additional_datastreams_permissions`
- AND a subsequent read SHALL set the attribute to null in state when the API response omits or empties the field

### Requirement: additional_datastreams_permissions — no revoke-on-unset; explicit empty list clears (REQ-017)

Unlike `elasticstack_fleet_integration_policy`, this resource SHALL NOT send an empty array in place of an omitted `additional_datastreams_permissions` field when the attribute is unset. Removing the attribute from configuration SHALL leave any previously granted permissions on the server unchanged. An explicit `additional_datastreams_permissions = []` SHALL be the supported way to clear previously granted permissions.

#### Scenario: Removing the attribute from config does not revoke permissions

- GIVEN an existing Defend policy with `additional_datastreams_permissions = ["logs-nginx-*"]` granted through Terraform
- WHEN the attribute is removed from configuration (not set to `[]`) and update runs, changing no other attribute
- THEN the update request body SHALL NOT include `additional_datastreams_permissions`
- AND the server-side permissions SHALL remain `["logs-nginx-*"]` after the update
- AND a subsequent read SHALL populate the attribute from the still-present API value rather than null

#### Scenario: Explicit empty list clears permissions

- GIVEN an existing Defend policy with `additional_datastreams_permissions = ["logs-nginx-*"]`
- WHEN the configuration changes to `additional_datastreams_permissions = []` and update runs
- THEN the update request body SHALL include the attribute as an empty array
- AND the server-side permissions SHALL be cleared
- AND a subsequent read SHALL populate the attribute as an empty list (not null), because the configuration explicitly holds a known empty list

### Requirement: Compatibility — additional_datastreams_permissions (REQ-018)

When `additional_datastreams_permissions` is configured with a known, non-null value, the resource SHALL verify the server version is at least 9.1.0, via an attribute-scoped version requirement evaluated independently of the existing `agent_policy_ids` version requirement (`MinVersionPolicyIDs`, 8.15.0), so that configuring one of the two attributes does not suppress the version check for the other. If the server version is lower, the resource SHALL return an attribute-level error diagnostic and SHALL NOT call the Fleet API.

#### Scenario: additional_datastreams_permissions rejected below minimum stack version

- GIVEN a configuration with `additional_datastreams_permissions = ["logs-nginx-*"]`
- WHEN create or update runs against Elastic Stack < 9.1.0
- THEN the provider SHALL return an attribute-scoped error diagnostic stating that `additional_datastreams_permissions` requires Elastic Stack >= 9.1.0
- AND no package policy SHALL be created or modified

#### Scenario: Omitting the attribute succeeds below minimum stack version

- GIVEN a configuration that does not set `additional_datastreams_permissions`
- WHEN create or update runs against Elastic Stack < 9.1.0
- THEN the operation SHALL succeed, since no version requirement is evaluated for a null/unknown attribute

#### Scenario: additional_datastreams_permissions and agent_policy_ids version requirements are independent

- GIVEN a configuration that sets `agent_policy_ids` but not `additional_datastreams_permissions`, or vice versa, or both, or neither
- WHEN the provider evaluates version requirements before create or update
- THEN each attribute's version requirement SHALL be evaluated solely based on whether that attribute itself is known (and, for `additional_datastreams_permissions`, non-null)
- AND configuring one of the two attributes SHALL NOT suppress the version check for the other
