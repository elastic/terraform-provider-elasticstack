## MODIFIED Requirements

### Requirement: State upgraders nullify empty-string JSON fields (REQ-044–REQ-045 extended)

The v0 → v1 and v1 → v2 state upgraders SHALL nullify empty-string values for the
`metadata` and `role_descriptors` fields before decoding state into the typed model.

Both `metadata` (typed as `jsontypes.NormalizedType{}`) and `role_descriptors`
(typed with a custom type embedding `jsontypes.Normalized` validation) are JSON
string attributes. An empty string is not valid JSON and causes an
`Invalid JSON String Value` error when the Plugin Framework decodes state. The SDKv2
provider stored `""` for optional string attributes when not configured; upgrading
to the Plugin Framework provider without this normalization would fail for any legacy
state where either field is an empty string.

The upgraders SHALL use the raw-state pattern:
1. Unmarshal `req.RawState.JSON` into `map[string]any` using `stateutil.UnmarshalStateMap`.
2. Call `stateutil.NullifyEmptyString(stateMap, "metadata", "role_descriptors")`.
3. Apply remaining normalization (see below) to the raw map.
4. Re-marshal with `stateutil.MarshalStateMap`.

The v0 → v1 upgrader SHALL additionally convert `expiration = ""` to `null`
(using `stateutil.NullifyEmptyString(stateMap, "expiration", "metadata", "role_descriptors")`).

The v1 → v2 upgrader SHALL additionally default `type` to `"rest"` when the field
is absent, `null`, or `""` in the raw state map.

Non-empty, non-null values for `metadata`, `role_descriptors`, and `expiration`
SHALL be passed through the upgrader without modification.

#### Scenario: v0 state — metadata is empty string

- GIVEN v0 state with `metadata` equal to `""`
- WHEN the v0 → v1 state upgrader runs
- THEN the upgraded state SHALL have `metadata` equal to `null`

#### Scenario: v0 state — role_descriptors is empty string

- GIVEN v0 state with `role_descriptors` equal to `""`
- WHEN the v0 → v1 state upgrader runs
- THEN the upgraded state SHALL have `role_descriptors` equal to `null`

#### Scenario: v0 state — both metadata and role_descriptors are empty strings

- GIVEN v0 state where both `metadata` and `role_descriptors` are `""`
- WHEN the v0 → v1 state upgrader runs
- THEN the upgraded state SHALL have both `metadata` and `role_descriptors` equal to `null`

#### Scenario: v0 state — metadata is valid JSON

- GIVEN v0 state with `metadata` equal to a non-empty JSON string (e.g. `"{}"`)
- WHEN the v0 → v1 state upgrader runs
- THEN the upgraded state SHALL have `metadata` unchanged

#### Scenario: v0 state — metadata is null or absent

- GIVEN v0 state where `metadata` is `null` or not present
- WHEN the v0 → v1 state upgrader runs
- THEN the upgraded state SHALL have `metadata` remain `null` (or absent)

#### Scenario: v1 state — metadata is empty string

- GIVEN v1 state with `metadata` equal to `""`
- WHEN the v1 → v2 state upgrader runs
- THEN the upgraded state SHALL have `metadata` equal to `null`

#### Scenario: v1 state — role_descriptors is empty string

- GIVEN v1 state with `role_descriptors` equal to `""`
- WHEN the v1 → v2 state upgrader runs
- THEN the upgraded state SHALL have `role_descriptors` equal to `null`

#### Scenario: v1 state — both metadata and role_descriptors are empty strings

- GIVEN v1 state where both `metadata` and `role_descriptors` are `""`
- WHEN the v1 → v2 state upgrader runs
- THEN the upgraded state SHALL have both `metadata` and `role_descriptors` equal to `null`

#### Scenario: v1 state — valid JSON fields are unchanged

- GIVEN v1 state with `metadata` or `role_descriptors` equal to valid JSON
- WHEN the v1 → v2 state upgrader runs
- THEN those fields SHALL be unchanged in the upgraded state
