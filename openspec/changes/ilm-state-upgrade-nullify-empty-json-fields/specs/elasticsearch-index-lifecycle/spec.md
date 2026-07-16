# Delta spec: `elasticstack_elasticsearch_index_lifecycle` — state upgrade empty-string JSON fix

## MODIFIED Requirements

### Requirement: Plugin Framework nested-block shape and state upgrade (REQ-030–REQ-031)

The resource SHALL model each phase block and each ILM action block as a Plugin Framework
`SingleNestedBlock`, so state stores them as objects instead of singleton lists. The resource SHALL
use schema version `1` and implement state upgrade from schema version `0`, unwrapping legacy
singleton-list phase values and legacy singleton-list action values into object values. The upgrade
SHALL leave `elasticsearch_connection` list-shaped.

After unwrapping all singleton lists, the state upgrader SHALL normalise empty-string JSON
attributes to null by calling `stateutil.NullifyEmptyString`:

1. At the **top level**, for the `metadata` attribute.
2. For each **phase block** that contains an `allocate` action, on the `include`, `exclude`, and
   `require` attributes of that `allocate` object.

`stateutil.NullifyEmptyString` is idempotent: keys that are absent, already null, or contain a
non-empty string SHALL be left unchanged.

This normalisation prevents `Invalid JSON String Value` errors caused by Plugin SDK v2 serialising
unset optional JSON attributes as `""` rather than null.

#### Scenario: Upgrade old SDK-shaped nested values

- GIVEN persisted schema version `0` state with a phase stored as `[ { ... } ]`
- WHEN Terraform runs the state upgrader
- THEN the upgraded state SHALL store that phase as a single object value

#### Scenario: Connection block not rewritten

- GIVEN persisted state with `elasticsearch_connection` stored as a list
- WHEN the ILM state upgrader runs
- THEN `elasticsearch_connection` SHALL remain list-shaped

#### Scenario: Empty-string metadata normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where the top-level `metadata` attribute is `""`
  (because it was never set in HCL)
- WHEN the provider runs the v0 → v1 state upgrader
- THEN the upgraded state SHALL contain `metadata = null`
- AND subsequent `terraform plan` SHALL complete without an `Invalid JSON String Value` error

#### Scenario: Empty-string allocate routing attributes normalized to null

- GIVEN version `0` state produced by Plugin SDK v2 where a warm or cold phase contains an `allocate`
  block with one or more of `include`, `exclude`, or `require` set to `""` (because those attributes
  were not set in HCL)
- WHEN the provider runs the v0 → v1 state upgrader
- THEN each of those attributes that is `""` SHALL become `null` in the upgraded state
- AND attributes that already contain a non-empty JSON string SHALL be preserved unchanged
- AND subsequent `terraform plan` SHALL complete without an `Invalid JSON String Value` error

#### Scenario: Non-empty JSON strings are preserved

- GIVEN version `0` state where `metadata` is a non-empty JSON object string and an `allocate` block
  has non-empty values for `include`, `exclude`, and `require`
- WHEN the provider runs the v0 → v1 state upgrader
- THEN all four attributes SHALL be carried through unchanged

#### Scenario: Combined empty-string metadata and allocate routing attributes

- GIVEN version `0` state with `"metadata": ""` at top level and a cold phase whose `allocate` block
  has `"include": ""`, `"exclude": ""`, and `"require": ""`
- WHEN the provider runs the v0 → v1 state upgrader
- THEN `metadata`, `include`, `exclude`, and `require` SHALL all become `null`
- AND the upgraded state SHALL decode against the v1 schema without error

## ADDED Requirements

### Requirement: Acceptance test — SDK upgrade without metadata (REQ-034)

The acceptance test suite SHALL include a test `TestAccResourceILMFromSDKNoMetadata` that verifies
the state upgrade succeeds for an ILM policy created without `metadata` and with a warm phase
`allocate` block that omits `include`, `exclude`, and `require`.

- **Step 1** SHALL use registry provider `0.14.5` (the last Plugin SDK v2 release) to create an ILM
  policy with a hot rollover phase and a warm phase with an `allocate` block specifying only
  `number_of_replicas`, with no `metadata`, `include`, `exclude`, or `require` set.
- **Step 2** SHALL re-apply the same logical configuration using the Plugin Framework provider
  (current in-tree). The provider SHALL complete the upgrade without error.
- **Step 3** SHALL be a plan-only step asserting no diff (`ExpectNonEmptyPlan: false`).

#### Scenario: End-to-end SDK upgrade — no metadata, allocate without routing filters

- GIVEN an ILM policy created by provider `0.14.5` with a warm phase `allocate` block and no
  `metadata`, `include`, `exclude`, or `require`
- WHEN the provider is upgraded to the Plugin Framework version and `terraform plan` runs
- THEN the plan SHALL succeed without an `Invalid JSON String Value` error
- AND the plan SHALL show no diff
