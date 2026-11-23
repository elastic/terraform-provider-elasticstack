## Why

`elasticstack_fleet_agent_policy`'s `advanced_settings` nested attribute already mirrors most of
Kibana's agent-policy `advanced_settings` sub-object (logging, download, `go_max_procs`,
`monitoring_runtime_experimental`), but is missing three keys that exist identically on the
`Kibana_HTTP_APIs_new_agent_policy` and `Kibana_HTTP_APIs_update_agent_policy_request_body` OpenAPI
schemas:

- `agent_features_disable_policy_change_acks_enabled`
- `agent_features_include_tags_in_events_enabled`
- `agent_internal`

The generated client (`generated/kbapi/kibana.gen.go`) already carries all three as `interface{}`
fields, and the provider's own shadow struct `advancedSettingsAPIValues`
(`internal/fleet/agentpolicy/models_advanced_settings.go`) already declares the matching JSON keys as
`any`. Nothing currently populates or reads them — the gap is isolated to the Terraform
schema/model layer, not the generated client.

This closes the gap flagged in #5274.

## What Changes

- Add three optional attributes inside the existing `advanced_settings` single-nested attribute in
  `internal/fleet/agentpolicy/schema.go`, with the `agent_` prefix stripped (matching the existing
  `agent_logging_level` → `logging_level` convention):
  - `features_disable_policy_change_acks_enabled` (bool)
  - `features_include_tags_in_events_enabled` (bool)
  - `internal` (string, YAML content, semantic equality via the existing
    `internal/utils/customtypes.NormalizedYamlType`/`NormalizedYamlValue`)
- Map all three to/from `advancedSettingsAPIValues`'s existing `AgentFeaturesDisablePolicyChangeAcksEnabled`,
  `AgentFeaturesIncludeTagsInEventsEnabled`, and `AgentInternal` JSON keys — no `generated/kbapi`
  regeneration required.
- Add three new attribute-scoped minimum-version requirements (one per field, each independently
  adjustable, following the `monitoring_runtime_experimental` precedent rather than reusing the
  shared `MinVersionAdvancedSettings` floor):
  - `features_disable_policy_change_acks_enabled`: 9.4.0
  - `features_include_tags_in_events_enabled`: 9.6.0
  - `internal`: 9.3.0
- Make all three `Optional` (not `Optional+Computed`), since Kibana returns `null` (not a
  server-side default) when any of them is unset, keeping "absent" and "null" identical in state.
- Extend `TestAccResourceAgentPolicyWithAdvancedSettings`, `TestPopulateAdvancedSettingsFromAPI`,
  `TestConvertAdvancedSettingsToAPI`, and `TestAgentPolicyVersionRequirements` with cases for all
  three fields, each version-gated per field's confirmed minimum.

## Capabilities

### New Capabilities

<!-- None — this change extends an existing capability only. -->

### Modified Capabilities

- `fleet-agent-policy`: extends the `advanced_settings` mapping requirement (REQ-033) and the
  version-gate requirement with three additional fields.

## Impact

- Affected code: `internal/fleet/agentpolicy/schema.go`, `models_advanced_settings.go`,
  `resource.go` (new `MinVersion...` constants), `version_requirements.go`, `capabilities.go`,
  `acc_test.go`, `models_advanced_settings_test.go`, `version_requirements_test.go`.
- Affected spec: `openspec/specs/fleet-agent-policy/spec.md` (schema block and REQ-033).
- Docs: `docs/resources/fleet_agent_policy.md` (regenerated via `make docs-generate` during
  implementation; out of scope for this proposal-only change).
- No breaking changes: all three new attributes are optional and additive; existing configurations
  are unaffected.
