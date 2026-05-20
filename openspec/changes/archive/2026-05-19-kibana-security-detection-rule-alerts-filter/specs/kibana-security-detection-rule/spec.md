## MODIFIED Requirements

### Requirement: Schema — `actions.alerts_filter` structured block (REQ-080)

The `alerts_filter` attribute on `actions` SHALL be a `SingleNestedBlock` (not a `MapAttribute`). It SHALL contain:

- A `query` nested block (optional) with:
  - `kql` — optional string. Defines a KQL query filter that determines whether the action runs.
  - `filters_json` — optional `jsontypes.Normalized` JSON string. Encodes the Kibana filter DSL array (same type as `params`). Use `jsonencode([])` for an empty filter list. Named `filters_json` to allow a future typed `filters` list attribute to be added without conflict.
- A `timeframe` nested block (optional) with:
  - `days` — optional list of int64 (values 1–7, where 1=Monday, 7=Sunday).
  - `timezone` — optional string. ISO time zone name (e.g., `"UTC"`, `"Europe/London"`).
  - `hours_start` — optional string. Start of active hours in 24-hour `HH:MM` notation.
  - `hours_end` — optional string. End of active hours in 24-hour `HH:MM` notation.

The old `MapAttribute(string)` form of `alerts_filter` is removed.

Updated schema sketch for the affected block:

```hcl
  actions {                       # optional list
    action_type_id = <required, string>
    id             = <required, string>
    params         = <required, JSON-normalized string>
    group          = <optional, string>
    uuid           = <optional, computed, string>

    alerts_filter {               # optional block
      query {                     # optional block
        kql          = <optional, string>
        filters_json = <optional, JSON-normalized string>  # jsonencode([]) for empty
      }
      timeframe {                 # optional block
        days        = <optional, list(int64)>  # 1–7
        timezone    = <optional, string>
        hours_start = <optional, string>       # HH:MM
        hours_end   = <optional, string>       # HH:MM
      }
    }

    frequency {                   # optional+computed
      notify_when = <required, string>
      summary     = <required, bool>
      throttle    = <required, string>
    }
  }
```

#### Scenario: alerts_filter with query and no timeframe

- GIVEN a detection rule action with `alerts_filter.query.kql = "event.action : \"test\""` and `alerts_filter.query.filters_json = jsonencode([])`
- WHEN the resource is created or updated
- THEN the Kibana API request SHALL include `alertsFilter.query.kql` and `alertsFilter.query.filters` in the expected nested JSON shape

#### Scenario: alerts_filter with timeframe

- GIVEN a detection rule action with `alerts_filter.query.kql` set and `alerts_filter.timeframe` block with all four attributes
- WHEN the resource is created or updated
- THEN the Kibana API request SHALL include `alertsFilter.timeframe.days`, `alertsFilter.timeframe.timezone`, and `alertsFilter.timeframe.hours.start` / `alertsFilter.timeframe.hours.end`

#### Scenario: no alerts_filter

- GIVEN a detection rule action without an `alerts_filter` block
- WHEN the resource is created or updated
- THEN the Kibana API request SHALL omit `alertsFilter` for that action

---

## ADDED Requirements

### Requirement: Validation — `timeframe` attributes required together (REQ-081)

When the `alerts_filter.timeframe` block is present, all four attributes (`days`, `timezone`, `hours_start`, `hours_end`) SHALL be required. The provider SHALL enforce this via `objectvalidator.AlsoRequires` (or equivalent). Omitting any one of the four attributes while the block is present SHALL be a validation error.

#### Scenario: timeframe with missing timezone

- GIVEN an `alerts_filter.timeframe` block with `days` and `hours_start` and `hours_end` set but `timezone` absent
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for the missing attribute

---

### Requirement: Write path — `alerts_filter` serialization (REQ-082)

When `alerts_filter` is configured, the provider SHALL serialize it to the Kibana API payload as:

```json
{
  "query": {
    "kql": "<kql value>",
    "filters": <parsed filters_json array>
  },
  "timeframe": {
    "days": [...],
    "timezone": "<tz>",
    "hours": {
      "start": "<hours_start>",
      "end": "<hours_end>"
    }
  }
}
```

- `query` SHALL be included when the `query` block is present with at least one of `kql` or `filters_json` set.
- `timeframe` SHALL be included when the `timeframe` block is present.
- `filters_json` SHALL be unmarshaled from a JSON string into a native array before serialization (not sent as a raw string).

#### Scenario: filters_json marshaled correctly

- GIVEN `alerts_filter.query.filters_json = jsonencode([{"meta": {"alias": null}}])`
- WHEN the provider serializes the action
- THEN `alertsFilter.query.filters` in the API request SHALL be the parsed array, not a JSON string

---

### Requirement: Read path — `alerts_filter` deserialization (REQ-083)

When the Kibana API returns an `alertsFilter` object on a detection rule action, the provider SHALL:

- Map `alertsFilter.query.kql` to `alerts_filter.query.kql` (string).
- Marshal `alertsFilter.query.filters` (array) to a normalized JSON string for `alerts_filter.query.filters_json`.
- Map `alertsFilter.timeframe.days`, `.timezone`, `.hours.start`, `.hours.end` to the corresponding `alerts_filter.timeframe` attributes.
- Set `alerts_filter` to null when the API response omits `alertsFilter` for that action.

The provider SHALL NOT produce a Go map-literal string (e.g., `"map[filters:[] kql:...]"`) in any state attribute.

#### Scenario: round-trip correctness

- GIVEN a detection rule with `alerts_filter.query.kql = "event.action : \"test\""` and `alerts_filter.query.filters_json = jsonencode([])`
- WHEN the resource is applied and then refreshed
- THEN `terraform plan` SHALL show no diff for `alerts_filter`

---

### Requirement: Schema version bump and state migration (REQ-084)

The `elasticstack_kibana_security_detection_rule` resource schema version SHALL be bumped from 1 to 2.

A `StateUpgraders` entry for version 1 → 2 SHALL be registered. Because the previous `alerts_filter` implementation was functionally broken (invalid read and write paths), no valid `alerts_filter` state can exist in practice. The upgrade function SHALL be a no-op that returns state unchanged (discards any stored `alerts_filter` map data without error).

#### Scenario: state upgrade from v1

- GIVEN existing Terraform state at schema version 1 (no valid `alerts_filter` value)
- WHEN the provider is upgraded to the version that introduces schema version 2
- THEN Terraform state is upgraded to version 2 without error

---

### Requirement: Acceptance tests — `alerts_filter` (REQ-085)

The acceptance test suite for `elasticstack_kibana_security_detection_rule` SHALL include:

1. A test case that creates a detection rule with an action using `alerts_filter.query.kql` and `alerts_filter.query.filters_json = jsonencode([])`, applies, then asserts no plan diff (round-trip correctness).
2. A test step that updates `alerts_filter.query.kql` to a different value and asserts apply succeeds and state reflects the change.
3. A test step that includes `alerts_filter.timeframe` with all four attributes and asserts round-trip correctness.
4. A test case (or step) that uses an action without `alerts_filter` and asserts no regression in create/update/read behavior.

#### Scenario: create with alerts_filter — no diff on refresh

- GIVEN a detection rule configured with `alerts_filter.query.kql` and `alerts_filter.query.filters_json`
- WHEN the resource is applied and then `terraform plan` is re-run
- THEN no diff is shown for `alerts_filter`

#### Scenario: update kql

- GIVEN an applied detection rule with `alerts_filter.query.kql = "old_value"`
- WHEN the configuration is updated to `alerts_filter.query.kql = "new_value"` and applied
- THEN the apply succeeds and state shows the new kql value

#### Scenario: timeframe round-trip

- GIVEN a detection rule with `alerts_filter.timeframe { days = [1,2,3,4,5], timezone = "UTC", hours_start = "08:00", hours_end = "17:00" }`
- WHEN the resource is applied and refreshed
- THEN all four timeframe attributes in state match configuration

---

### Requirement: No regression — actions without alerts_filter (REQ-086)

Existing detection rule actions that do not include `alerts_filter` SHALL continue to work without changes. Create, update, and read operations for such actions SHALL produce no errors and no unexpected plan diffs.

#### Scenario: action without alerts_filter creates successfully

- GIVEN a detection rule action with `id`, `action_type_id`, and `params` set but no `alerts_filter` block
- WHEN the resource is created and then refreshed
- THEN apply succeeds and `terraform plan` shows no diff for the action

---

### Requirement: Documentation — `alerts_filter` block (REQ-087)

The provider documentation for `elasticstack_kibana_security_detection_rule` SHALL document:

- The `alerts_filter` nested block and its sub-attributes (`query.kql`, `query.filters_json`, `timeframe.*`).
- A usage example showing `jsonencode([])` for `filters_json`.
- The `filters_json` attribute name rationale is not required in docs; clear examples suffice.

#### Scenario: docs include alerts_filter example

- GIVEN the generated resource documentation
- WHEN a practitioner reads the `actions.alerts_filter` section
- THEN the docs show a `query.filters_json = jsonencode([])` example and describe each sub-attribute
