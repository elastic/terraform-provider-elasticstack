## MODIFIED Requirements

### Requirement: State upgrade v0 → v1 — `params` empty-string normalization (REQ-029)

The v0 → v1 state upgrader SHALL nullify empty-string `params` values for both
the rule-level `params` field and each action's `params` field, in addition to the
transformations already specified in REQ-029–REQ-034.

The rule-level `params` field, when present in v0 state as an empty string (`""`),
SHALL be set to `null` in the upgraded state. For each action entry in `actions[]`,
the action-level `params` field, when present as an empty string (`""`), SHALL be
set to `null` in the upgraded state.

These fields are typed as `jsontypes.Normalized` in the Plugin Framework model. An
empty string is not valid JSON and causes an `Invalid JSON String Value` error when
loading state. Setting them to `null` is the correct normalized form for
"not configured". Non-empty, non-null values (including valid JSON such as `"{}"`)
SHALL be passed through the upgrader without modification.

#### Scenario: Rule-level params is empty string

- GIVEN v0 state with `params` equal to `""`
- WHEN v0 → v1 state upgrade runs
- THEN the upgraded state SHALL have `params` equal to `null`

#### Scenario: Rule-level params is valid JSON

- GIVEN v0 state with `params` equal to a non-empty JSON string (e.g. `"{}"`)
- WHEN v0 → v1 state upgrade runs
- THEN the upgraded state SHALL have `params` unchanged

#### Scenario: Rule-level params is null

- GIVEN v0 state with `params` equal to `null` (or absent)
- WHEN v0 → v1 state upgrade runs
- THEN the upgraded state SHALL have `params` equal to `null`

#### Scenario: Action-level params is empty string

- GIVEN v0 state with one or more actions, where an action's `params` is `""`
- WHEN v0 → v1 state upgrade runs
- THEN the upgraded state SHALL have that action's `params` equal to `null`

#### Scenario: Action-level params is valid JSON

- GIVEN v0 state with an action whose `params` is a non-empty JSON string
- WHEN v0 → v1 state upgrade runs
- THEN the upgraded state SHALL have that action's `params` unchanged

#### Scenario: Both rule-level and action-level params are empty strings

- GIVEN v0 state where `params` is `""` and at least one action also has `params`
  equal to `""`
- WHEN v0 → v1 state upgrade runs
- THEN both the rule-level `params` and the action-level `params` SHALL be `null`
  in the upgraded state
