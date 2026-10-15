## MODIFIED Requirements

### Requirement: Mapping — advanced_settings (REQ-033)

The resource SHALL map `advanced_settings` attributes to the Fleet API's `advanced_settings` object using the key names defined in the API schema (e.g. `agent_logging_level`, `agent_limits_go_max_procs`, `agent_features_disable_policy_change_acks_enabled`, `agent_features_include_tags_in_events_enabled`, `agent_internal`). Only attributes that are known (not null or unknown) SHALL be included in the API payload. On read, if the API response does not include `advanced_settings`, the resource SHALL store a null object in state.

The Terraform attributes `features_disable_policy_change_acks_enabled` (bool), `features_include_tags_in_events_enabled` (bool), and `internal` (string, YAML content) SHALL map to the API's `agent_features_disable_policy_change_acks_enabled`, `agent_features_include_tags_in_events_enabled`, and `agent_internal` keys respectively, with the `agent_` prefix stripped per the resource's existing naming convention (e.g. `agent_logging_level` → `logging_level`). All three SHALL be `Optional` (not `Computed`), since the API returns `null`, not a server-assigned default, when any of them is unset.

#### Scenario: Null advanced_settings on read

- GIVEN the API response has no `advanced_settings` field
- WHEN read runs
- THEN `advanced_settings` SHALL be null in state

#### Scenario: Agent feature toggles mapped

- GIVEN `advanced_settings.features_disable_policy_change_acks_enabled = true` and `advanced_settings.features_include_tags_in_events_enabled = false` are configured
- WHEN create or update runs
- THEN the API payload's `advanced_settings` object SHALL include `agent_features_disable_policy_change_acks_enabled: true` and `agent_features_include_tags_in_events_enabled: false`

#### Scenario: Omitted agent feature toggles excluded from payload

- GIVEN `advanced_settings.features_disable_policy_change_acks_enabled` and `advanced_settings.features_include_tags_in_events_enabled` are not configured
- WHEN create or update runs
- THEN the API payload's `advanced_settings` object SHALL NOT include `agent_features_disable_policy_change_acks_enabled` or `agent_features_include_tags_in_events_enabled`

#### Scenario: internal YAML mapped

- GIVEN `advanced_settings.internal` is configured with a valid YAML string
- WHEN create or update runs
- THEN the API payload's `advanced_settings` object SHALL include `agent_internal` set to that string

#### Scenario: internal YAML semantic equality

- GIVEN a resource was created with `advanced_settings.internal` set to a valid YAML string
- WHEN the configuration is reapplied with YAML content that is textually different but parses to the same data (e.g. different indentation, comment placement, or key order)
- THEN the apply SHALL succeed and Terraform SHALL report no persistent diff for `advanced_settings.internal` after apply, and a refresh in which Kibana returns textually different but equivalent YAML SHALL NOT produce a diff

## ADDED Requirements

### Requirement: Compatibility — advanced_settings.internal version gate

When `advanced_settings.internal` is configured, the resource SHALL verify the stack version is at least 9.3.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. This gate is independent of the overall `advanced_settings` minimum version (8.17.0, REQ-009–REQ-017).

#### Scenario: internal below minimum version

- GIVEN stack version < 9.3.0
- WHEN `advanced_settings.internal` is configured
- THEN the resource SHALL fail with a version error before calling the API

### Requirement: Compatibility — advanced_settings.features_disable_policy_change_acks_enabled version gate

When `advanced_settings.features_disable_policy_change_acks_enabled` is configured, the resource SHALL verify the stack version is at least 9.4.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. This gate is independent of the overall `advanced_settings` minimum version (8.17.0, REQ-009–REQ-017).

#### Scenario: features_disable_policy_change_acks_enabled below minimum version

- GIVEN stack version < 9.4.0
- WHEN `advanced_settings.features_disable_policy_change_acks_enabled` is configured, including an explicit `false` value
- THEN the resource SHALL fail with a version error before calling the API

### Requirement: Compatibility — advanced_settings.features_include_tags_in_events_enabled version gate

When `advanced_settings.features_include_tags_in_events_enabled` is configured, the resource SHALL verify the stack version is at least 9.6.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. This gate is independent of the overall `advanced_settings` minimum version (8.17.0, REQ-009–REQ-017).

#### Scenario: features_include_tags_in_events_enabled below minimum version

- GIVEN stack version < 9.6.0
- WHEN `advanced_settings.features_include_tags_in_events_enabled` is configured, including an explicit `false` value
- THEN the resource SHALL fail with a version error before calling the API

### Requirement: Compatibility — unrelated advanced_settings fields unaffected by new gates

The three new per-attribute version gates (`internal`, `features_disable_policy_change_acks_enabled`, `features_include_tags_in_events_enabled`) SHALL NOT apply to any other `advanced_settings` attribute.

#### Scenario: Unrelated advanced_settings fields unaffected

- GIVEN stack version is at least 8.17.0 but below 9.3.0, 9.4.0, and 9.6.0
- WHEN `advanced_settings` is configured without `internal`, `features_disable_policy_change_acks_enabled`, or `features_include_tags_in_events_enabled`
- THEN the resource SHALL NOT fail due to any of these three version gates
