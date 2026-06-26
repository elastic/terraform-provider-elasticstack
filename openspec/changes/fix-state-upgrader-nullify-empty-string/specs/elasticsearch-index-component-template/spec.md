# Delta spec: `elasticstack_elasticsearch_component_template` — state upgrade empty-string fix

## MODIFIED Requirements

### Requirement: State upgrade to schema version 1 (REQ-032–REQ-035)

The resource SHALL define schema version `1` and provide an upgrade path from version `0`. During
state upgrade from version `0`, the provider SHALL collapse legacy list-shaped `template` blocks to
the Plugin Framework object-or-null representation. During that upgrade, the provider SHALL ensure
the migrated `template` object contains explicit keys for `alias`, `mappings`, `settings`, and
`data_stream_options`, using null when absent. During that upgrade, the provider SHALL normalize
legacy alias state by converting SDK-style duplicated `index_routing` and `search_routing` values
into the Plugin Framework routing-only representation and by converting empty-string alias `filter`
values to null.

Immediately after ensuring the migrated `template` keys, the upgrader SHALL call
`stateutil.NullifyEmptyString` on the `template` map for the `mappings` and `settings` attributes.
The upgrader SHALL also call `stateutil.NullifyEmptyString` on the top-level state map for the
`metadata` attribute. Both calls SHALL convert any empty-string value (`""`) to null; keys that are
absent, already null, or non-empty SHALL be left unchanged.

This change ensures that SDK v2 state written with `"mappings": ""` or `"settings": ""` (produced
when the corresponding HCL attribute was omitted) is normalized to `null` before the Plugin
Framework decodes it, preventing JSON semantic-equality errors such as `unexpected end of JSON
input`.

#### Scenario: Upgrade legacy template state

- GIVEN version `0` state containing list-shaped `template` data and legacy alias routing values
- WHEN the provider upgrades state to schema version `1`
- THEN the provider SHALL collapse `template` to object-or-null form
- AND it SHALL preserve equivalent alias routing semantics without creating spurious diffs

#### Scenario: Settings-only template — empty-string mappings normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where `template.mappings` is `""` and
  `template.settings` contains a non-empty JSON object (a settings-only template with no `mappings`
  block in HCL)
- WHEN the provider upgrades state to schema version `1`
- THEN the upgraded state SHALL contain `template.mappings = null`
- AND `template.settings` SHALL be preserved unchanged
- AND subsequent `terraform plan` SHALL complete without a Semantic Equality Check Error

#### Scenario: Empty-string metadata normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where top-level `metadata` is `""`
- WHEN the provider upgrades state to schema version `1`
- THEN the upgraded state SHALL contain `metadata = null`
- AND the upgraded state SHALL decode against the v1 schema without error

#### Scenario: Mappings-only template — empty-string settings normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where `template.settings` is `""` and
  `template.mappings` contains a non-empty JSON object (a mappings-only template with no `settings`
  block in HCL)
- WHEN the provider upgrades state to schema version `1`
- THEN the upgraded state SHALL contain `template.settings = null`
- AND `template.mappings` SHALL be preserved unchanged
- AND subsequent `terraform plan` SHALL complete without a Semantic Equality Check Error

#### Scenario: Non-empty JSON strings are preserved

- GIVEN version `0` state where `template.mappings` is a non-empty JSON object string and
  `template.settings` is a non-empty JSON object string
- WHEN the provider upgrades state to schema version `1`
- THEN both `template.mappings` and `template.settings` SHALL be carried through unchanged

## ADDED Requirements

### Requirement: Acceptance test — settings-only SDK upgrade (REQ-036)

The acceptance test suite SHALL include a test `TestAccResourceComponentTemplateFromSDKSettingsOnly`
that verifies the state upgrade succeeds for a settings-only component template (no `mappings`
block).

- **Step 1** SHALL use registry provider `0.14.5` (the last Plugin SDK v2 release) to create a
  component template that includes only a `settings` block, with no `mappings` and no `alias`.
- **Step 2** SHALL re-apply the same logical configuration using the Plugin Framework provider
  (current in-tree). The provider SHALL complete the upgrade without error. The resulting state
  SHALL show `template.mappings` as null/empty.
- **Step 3** SHALL be a plan-only step asserting no diff (`ExpectNonEmptyPlan: false`).

#### Scenario: End-to-end SDK upgrade for settings-only template

- GIVEN a component template created by provider `0.14.5` using only a `settings` block
- WHEN the provider is upgraded to the Plugin Framework version and `terraform plan` runs
- THEN the plan SHALL succeed (no `Semantic Equality Check Error`)
- AND the plan SHALL show no diff (no spurious changes)
