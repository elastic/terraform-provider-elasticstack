# Delta spec: `elasticstack_elasticsearch_index_template` — state upgrade empty-string fix

## MODIFIED Requirements

### Requirement: State schema version 0 → 1 upgrader (REQ-041)

The Plugin Framework resource SHALL declare schema `Version` `1` and SHALL implement
`resource.ResourceWithUpgradeState` registering an upgrader for prior schema version `0` (the version
under which Plugin SDK v2 wrote state). The upgrader SHALL transform tfstate written by the SDK
implementation into the Plugin Framework v1 shape by collapsing each of the following paths from a
list-/set-shaped value to a single-object shape:

- `data_stream`
- `template`
- `template.lifecycle`
- `template.data_stream_options`
- `template.data_stream_options.failure_store`
- `template.data_stream_options.failure_store.lifecycle`

For each listed path, after parent paths have already been collapsed, the upgrader SHALL apply the
following rule:

- If the value is null or absent, leave it unchanged.
- If the value is an empty array (`[]`), set it to null.
- If the value is a single-element array (`[obj]`), replace it with `obj`.
- If the value is an array with more than one element, return an error diagnostic identifying the
  path; the upgrader SHALL NOT silently drop elements.

After collapsing `template`, the provider SHALL ensure the migrated `template` object contains
explicit keys for `alias`, `mappings`, `settings`, `lifecycle`, and `data_stream_options`, using null
when absent. Immediately after ensuring those keys, the upgrader SHALL call
`stateutil.NullifyEmptyString` on the `template` map for the `mappings` and `settings` attributes.
The upgrader SHALL also call `stateutil.NullifyEmptyString` on the top-level state map for the
`metadata` attribute. Both calls SHALL convert any empty-string value (`""`) to null; keys that are
absent, already null, or non-empty SHALL be left unchanged.

All other non-converted attributes (including non-empty JSON-string attributes such as `metadata`,
`template.mappings`, `template.settings`, and `template.alias.filter`) SHALL be carried through
unchanged. After the upgrader runs, Terraform SHALL be able to decode the resulting state against the
v1 schema without further migration.

This change prevents decode and semantic-equality errors such as `missing expected {` or
`unexpected end of JSON input` reported for index templates during provider upgrades from ≤0.14.x to
≥0.15.0.

#### Scenario: Upgrade single-element list to object

- GIVEN tfstate written by Plugin SDK v2 with `data_stream = [{"hidden": true, "allow_custom_routing": false}]`
- WHEN the v0 → v1 upgrader runs
- THEN the upgraded state SHALL contain `data_stream = {"hidden": true, "allow_custom_routing": false}`

#### Scenario: Upgrade empty list to null

- GIVEN tfstate written by Plugin SDK v2 with `template.data_stream_options = []`
- WHEN the v0 → v1 upgrader runs
- THEN the upgraded state SHALL contain `template.data_stream_options = null`

#### Scenario: Upgrade nested single-element collections

- GIVEN tfstate written by Plugin SDK v2 with `template = [{"data_stream_options": [{"failure_store": [{"enabled": true, "lifecycle": [{"data_retention": "30d"}]}]}]}]`
- WHEN the v0 → v1 upgrader runs
- THEN the upgraded state SHALL contain `template = {"data_stream_options": {"failure_store": {"enabled": true, "lifecycle": {"data_retention": "30d"}}}}`

#### Scenario: Upgrade preserves non-collapsed attributes

- GIVEN tfstate written by Plugin SDK v2 that includes non-empty `metadata`, `composed_of`,
  `index_patterns`, and `template.alias` populated
- WHEN the v0 → v1 upgrader runs
- THEN those attributes SHALL be carried through byte-equivalent in the upgraded state

#### Scenario: Refuse multi-element arrays at collapsed paths

- GIVEN tfstate at one of the collapsed paths contains an array with two or more elements (a state
  corruption that should not occur because the SDK enforced `MaxItems: 1`)
- WHEN the v0 → v1 upgrader runs
- THEN it SHALL return an error diagnostic that identifies the offending path
- AND the upgrader SHALL NOT silently discard elements

#### Scenario: Settings-only index template — empty-string mappings normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where `template.mappings` is `""` and
  `template.settings` contains a non-empty JSON object (a settings-only template with no `mappings`
  block in HCL)
- WHEN the provider upgrades state to schema version `1`
- THEN the upgraded state SHALL contain `template.mappings = null`
- AND `template.settings` SHALL be preserved unchanged
- AND subsequent `terraform plan` SHALL complete without error

#### Scenario: Empty-string metadata normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where top-level `metadata` is `""`
- WHEN the provider upgrades state to schema version `1`
- THEN the upgraded state SHALL contain `metadata = null`
- AND the upgraded state SHALL decode against the v1 schema without error

#### Scenario: Mappings-only index template — empty-string settings normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where `template.settings` is `""` and
  `template.mappings` contains a non-empty JSON object (a mappings-only template with no `settings`
  block in HCL)
- WHEN the provider upgrades state to schema version `1`
- THEN the upgraded state SHALL contain `template.settings = null`
- AND `template.mappings` SHALL be preserved unchanged
- AND subsequent `terraform plan` SHALL complete without error

#### Scenario: Non-empty JSON strings are preserved

- GIVEN version `0` state where `template.mappings` is a non-empty JSON object string and
  `template.settings` is a non-empty JSON object string
- WHEN the provider upgrades state to schema version `1`
- THEN both `template.mappings` and `template.settings` SHALL be carried through unchanged

#### Scenario: End-to-end upgrade from prior SDK release

- GIVEN a resource created by the last Plugin SDK v2 release exercising every collapsed block
- WHEN the same configuration is re-applied with the new Plugin Framework provider
- THEN the v0 → v1 upgrader SHALL run automatically as part of refresh
- AND the subsequent plan SHALL show no diff

## ADDED Requirements

### Requirement: Acceptance test — settings-only SDK upgrade (REQ-048)

The acceptance test suite SHALL include a test `TestAccResourceIndexTemplateFromSDKSettingsOnly`
that verifies the state upgrade succeeds for a settings-only index template (no `mappings` block).

- **Step 1** SHALL use registry provider `0.14.5` (the last Plugin SDK v2 release) to create an
  index template with `index_patterns` and only a `settings` block, with no `mappings`, no
  `data_stream`, and no `alias`.
- **Step 2** SHALL re-apply the same logical configuration using the Plugin Framework provider
  (current in-tree). The provider SHALL complete the upgrade without error. The resulting state
  SHALL show `template.mappings` as null/empty.
- **Step 3** SHALL be a plan-only step asserting no diff (`ExpectNonEmptyPlan: false`).

#### Scenario: End-to-end SDK upgrade for settings-only index template

- GIVEN an index template created by provider `0.14.5` using only a `settings` block
- WHEN the provider is upgraded to the Plugin Framework version and `terraform plan` runs
- THEN the plan SHALL succeed (no decode error, no Semantic Equality Check Error)
- AND the plan SHALL show no diff (no spurious changes)
