## MODIFIED Requirements

### Requirement: Focused package-policy envelope and fixed package identity (REQ-003)

The resource SHALL expose a familiar package-policy envelope with `id`, `policy_id`, `name`,
`namespace`, `agent_policy_id`, `agent_policy_ids`, `description`, `enabled`, `force`,
`integration_version`, `advanced_settings`, and `space_ids`. The resource SHALL always target
package name `endpoint` and SHALL NOT expose a user-configurable `integration_name`. The resource
SHALL NOT expose the generic `vars_json`, generic `inputs`, generic `streams`, or `output_id`
surfaces from `elasticstack_fleet_integration_policy` in v1.

#### Scenario: Package name is fixed to Elastic Defend

- GIVEN a valid `elasticstack_fleet_elastic_defend_integration_policy` configuration
- WHEN create or update builds the API request
- THEN the request body SHALL target package name `endpoint`
- AND there SHALL be no user-configurable `integration_name` in the Terraform schema

### Requirement: Create finalizes the modeled policy after bootstrap (REQ-009)

After the bootstrap create succeeds, the resource SHALL submit a Defend-specific update request
that applies the configured typed `policy` settings and any configured `advanced_settings` to
the new package policy. The finalized request SHALL include the provider-modeled Defend `policy`
payload (including advanced settings merged into `policy.{os}.advanced`), the configured `preset`
mapped under `config.integration_config.value.endpointConfig.preset`, and the top-level package
policy `version`. The `artifact_manifest` SHALL NOT be included — Kibana manages it server-side
and rejects it when present in typed input config.

#### Scenario: Create applies modeled policy settings after bootstrap

- GIVEN a Defend resource configuration with non-default policy settings
- WHEN create completes
- THEN the provider SHALL apply those settings through a follow-up Defend package policy update
- AND Terraform users SHALL NOT need to supply server-managed Defend payloads directly

#### Scenario: Create applies advanced settings after bootstrap

- GIVEN a Defend resource configuration with `advanced_settings` containing
  `linux.advanced.artifacts.global.base_url`
- WHEN create completes
- THEN the finalize update SHALL include that value under `policy.linux.advanced` in the typed
  Defend config payload

### Requirement: Update preserves opaque server-managed Defend payloads (REQ-010)

On update, the resource SHALL send the Defend-specific typed package policy shape required by
Kibana, including the latest provider-modeled `preset`, `policy`, and `advanced_settings` values.
The provider SHALL preserve and resend the top-level package policy `version` (for optimistic
concurrency control) without exposing it in the public Terraform schema. The `artifact_manifest`
SHALL NOT be included in update requests — Kibana manages it server-side and rejects it when
present. The provider captures `artifact_manifest` from API responses into private state but does
not send it on update.

#### Scenario: Update succeeds without exposing artifact manifest

- GIVEN an existing Defend resource that was previously created or imported
- WHEN a user changes a modeled policy setting
- THEN the provider SHALL include the package policy `version` token in the update request for
  optimistic concurrency control
- AND the Terraform schema SHALL not expose `artifact_manifest` as a configurable field
- AND the provider SHALL NOT send `artifact_manifest` in the update request body

### Requirement: Read and import map only modeled fields to state (REQ-011)

On read and import, the resource SHALL parse the Defend-specific package policy response and
populate only the modeled Terraform schema fields. The provider SHALL map `preset` from the Defend
`integration_config` payload, SHALL map the typed `policy` payload into the corresponding
operating-system nested attributes, and SHALL map `policy.{os}.advanced` nested objects into
`advanced_settings` when that attribute is managed in configuration. The provider SHALL ignore
unmodeled server-managed Defend payloads in Terraform state, except for preserving any opaque
values required for future updates in internal provider-managed state.

#### Scenario: Read ignores unmodeled server-managed Defend fields

- GIVEN a Defend package policy response that includes `artifact_manifest` and other
  server-managed Defend data
- WHEN the resource reads or imports that package policy
- THEN Terraform state SHALL include only the modeled schema fields
- AND the provider SHALL preserve any required opaque update data internally

#### Scenario: Read maps advanced settings when configured

- GIVEN a Defend package policy whose API payload includes `policy.linux.advanced.artifacts.global.base_url`
- AND Terraform configuration that sets `advanced_settings`
- WHEN the resource reads state
- THEN `advanced_settings` SHALL include
  `linux.advanced.artifacts.global.base_url` with the API value

## ADDED Requirements

### Requirement: Advanced settings map attribute (REQ-015)

The resource SHALL expose an optional `advanced_settings` attribute of type `map(string)`. Keys
SHALL use Elastic's documented advanced-setting names with OS prefix and dot notation (for example
`linux.advanced.artifacts.global.base_url`, `windows.advanced.kernel.connect`,
`mac.advanced.harden.self_protect`) as defined in
[Elastic Defend advanced settings](https://www.elastic.co/docs/reference/security/defend-advanced-settings).
Values SHALL be opaque strings passed through to the Defend policy payload; the provider SHALL NOT
attempt to validate setting semantics or coerce types beyond string encoding.

The provider SHALL translate flat Terraform map keys into the nested `policy.{os}.advanced`
structure expected by the Fleet typed Defend config envelope, and SHALL flatten nested advanced
objects back into dot-notation keys on read. Keys MUST begin with one of `linux.`, `mac.`, or
`windows.` followed by `.advanced.`; keys that do not match this pattern SHALL produce a plan-time
or apply-time validation error.

When `advanced_settings` is null or unset in configuration, the provider SHALL omit advanced
settings from create/update payloads and SHALL NOT clear advanced settings that exist only on the
server (matching the unmanaged-field pattern used for `description`). When `advanced_settings` is
configured (including an empty map), the provider SHALL send only the configured keys on
create/update and SHALL treat absent keys as unmanaged within the advanced-settings surface.

#### Scenario: Air-gapped artifact base URL

- GIVEN `advanced_settings = { "linux.advanced.artifacts.global.base_url" = "http://10.0.0.33" }`
- WHEN create or update runs
- THEN the typed Defend policy payload SHALL include
  `policy.linux.advanced.artifacts.global.base_url = "http://10.0.0.33"`

#### Scenario: Multiple OS advanced settings in one map

- GIVEN `advanced_settings` with distinct keys for `linux`, `mac`, and `windows` artifact base URLs
- WHEN create or update runs
- THEN each value SHALL be placed under the corresponding `policy.{os}.advanced` subtree
- AND keys for different operating systems SHALL NOT overwrite one another

#### Scenario: Invalid advanced setting key rejected

- GIVEN `advanced_settings` containing key `artifacts.global.base_url` without an OS prefix
- WHEN Terraform validates or applies the configuration
- THEN the provider SHALL return an error diagnostic
- AND no package policy SHALL be modified

#### Scenario: Unset advanced settings leaves server values intact

- GIVEN an existing Defend package policy with advanced settings configured in Kibana
- AND a Terraform configuration that does not set `advanced_settings`
- WHEN update runs changing only a modeled `policy` field
- THEN the update payload SHALL NOT include an `advanced` subtree derived from Terraform
- AND server-side advanced settings outside Terraform management SHALL remain unchanged

#### Scenario: Empty advanced settings map clears managed keys

- GIVEN a Terraform configuration with `advanced_settings = {}`
- WHEN update runs
- THEN the provider SHALL send advanced settings as empty for each OS that had previously managed
  advanced keys in state
- AND keys previously managed by Terraform SHALL be removed from the Defend policy
