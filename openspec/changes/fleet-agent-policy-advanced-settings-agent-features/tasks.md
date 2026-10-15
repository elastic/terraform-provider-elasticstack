## 1. Pre-implementation verification

- [x] 1.1 Against a live Kibana/Fleet stack at 9.3.0, confirm `agent_internal` is accepted on
  create/update and round-trips on read; confirm it is rejected (or absent) below 9.3.0
- [x] 1.2 Against a live Kibana/Fleet stack at 9.6.0, confirm `agent_features_include_tags_in_events_enabled`
  is accepted; confirm the 9.4.0 rejection (`Additional properties are not allowed`) still reproduces
  below 9.6.0
- [x] 1.3 Record any deviation from the 9.3.0 / 9.4.0 / 9.6.0 minimums in design.md before proceeding

## 2. Constants and capability flags

- [x] 2.1 Add `MinVersionAgentInternal` (9.3.0), `MinVersionFeaturesDisablePolicyChangeAcksEnabled`
  (9.4.0), `MinVersionFeaturesIncludeTagsInEventsEnabled` (9.6.0) to
  `internal/fleet/agentpolicy/resource.go`, alongside the existing `MinVersion...` constants
- [x] 2.2 Add `SupportsAgentInternal`, `SupportsFeaturesDisablePolicyChangeAcksEnabled`,
  `SupportsFeaturesIncludeTagsInEventsEnabled` to `agentPolicyFeatures` in
  `internal/fleet/agentpolicy/capabilities.go`, resolved via `client.EnforceMinVersion` in
  `resolveAgentPolicyFeatures`

## 3. Schema

- [x] 3.1 Add `features_disable_policy_change_acks_enabled` (`schema.BoolAttribute`, Optional) and
  `features_include_tags_in_events_enabled` (`schema.BoolAttribute`, Optional) to the
  `advanced_settings` nested attribute in `internal/fleet/agentpolicy/schema.go`
- [x] 3.2 Add `internal` (`schema.StringAttribute`, Optional, `CustomType: customtypes.NormalizedYamlType{}`)
  to the same block; schema description SHALL note the field holds free-form YAML with no stability
  guarantee, per Kibana's own UI copy
- [x] 3.3 Confirm `advancedSettingsAttrTypes()` picks up the three new attributes automatically (it
  derives types from the schema)

## 4. Model and mapping

- [x] 4.1 Add `FeaturesDisablePolicyChangeAcksEnabled types.Bool`, `FeaturesIncludeTagsInEventsEnabled types.Bool`,
  and `Internal customtypes.NormalizedYamlValue` fields (with matching `tfsdk` tags) to
  `advancedSettingsModel` in `internal/fleet/agentpolicy/models_advanced_settings.go`
- [x] 4.2 Extend `hasValues` to include all three new fields
- [x] 4.3 Extend `populateAdvancedSettingsFromAPI` to populate the two booleans via
  `typeutils.BoolFromAny(...)` and `Internal` via the string-from-`any` equivalent already used for
  other string fields in that function
- [x] 4.4 Extend `convertAdvancedSettingsToAPI` to emit `agent_features_disable_policy_change_acks_enabled`,
  `agent_features_include_tags_in_events_enabled`, and `agent_internal` only when the corresponding
  `agentPolicyFeatures.Supports...` flag is true and the attribute is known
  (`typeutils.IsKnown`), mirroring the existing `monitoring_runtime_experimental` gating in that
  function

## 5. Version requirements

- [x] 5.1 In `version_requirements.go`'s `GetVersionRequirements`, after decoding `settings`, add one
  `entitycore.NewAttributeVersionRequirement` call per new field
  (`path.Root("advanced_settings").AtName("features_disable_policy_change_acks_enabled")`, etc.),
  guarded by `typeutils.IsKnown(settings.<Field>)`, using the corresponding `MinVersion...` constant

## 6. Unit tests

- [x] 6.1 Extend `TestPopulateAdvancedSettingsFromAPI` (`models_advanced_settings_test.go`) with
  explicit-value and explicit-null subtests for all three new fields
- [x] 6.2 Extend `TestConvertAdvancedSettingsToAPI` with table-driven cases per field: explicit
  `true`/`false` (both booleans), a non-empty YAML string (`internal`), and known-but-unsupported
  (capability flag false) asserting the JSON key is omitted — new coverage, not mirrored from an
  existing case, per the issue's research comment
- [x] 6.3 Extend `TestAgentPolicyVersionRequirements` (`version_requirements_test.go`) with an entry
  per new field asserting its `path`/`MinVersion` pair appears in the returned requirements slice,
  and a case exercising the check when the field is unset (no requirement emitted)

## 7. Acceptance tests

- [x] 7.1 Extend `TestAccResourceAgentPolicyWithAdvancedSettings` (`acc_test.go`)'s `create_with_logging`
  step to also set all three new fields (two booleans `true`, `internal` a short valid YAML string);
  assert via `resource.TestCheckResourceAttr`
- [x] 7.2 Extend the existing `update_settings` step to flip both booleans to `false` and change
  `internal`'s YAML content; assert the new values, following the existing `logging_to_files`
  true→false precedent
- [x] 7.3 Extend the `remove_settings`/`set_to_defaults` steps to omit the three new fields; assert
  the resource still applies without a persistent diff
- [x] 7.4 Confirm the existing `import` step's `ImportStateVerify` round-trips all three new fields
- [x] 7.5 Add a `SkipFunc` per field (9.3.0 / 9.4.0 / 9.6.0), following the
  `monitoring_runtime_experimental` version-aware `SkipFunc` precedent (lines 1039-1052)
- [x] 7.6 Add an apply-time negative version-gating case per field: configuring the field (including
  explicit `false`) against a stack below its confirmed minimum SHALL surface the diagnostic
  `entitycore.EnforceVersionRequirements` raises from `Create`/`Update`, asserted as an `ExpectError`
  acceptance step
- [x] 7.7 Add an `internal` YAML-specific case asserting semantic equality: reformatted-but-equivalent
  YAML (different indentation/comment placement, same parsed content) produces no plan diff after
  apply

## 8. Docs and spec sync

- [ ] 8.1 Regenerate `docs/resources/fleet_agent_policy.md` via `make docs-generate`
- [ ] 8.2 Add a CHANGELOG entry following the repo's existing format
- [ ] 8.3 Sync this change's delta spec into `openspec/specs/fleet-agent-policy/spec.md` (schema
  block and REQ-033) when archiving

## 9. Validation and cleanup

- [ ] 9.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 9.2 Run `make check-openspec` — confirm the change validates
- [ ] 9.3 Run targeted acceptance tests against real Kibana/Fleet stacks at 9.3.0, 9.4.0, and 9.6.0
  (per `dev-docs/high-level/testing.md`)
- [ ] 9.4 Run the full unit test suite for `internal/fleet/agentpolicy`
- [ ] 9.5 Self-review with the `requirements-verification` skill against this change's specs
