# `elasticstack_kibana_alerting_rule` — Schema and Functional Requirements

Resource implementation: `internal/kibana/alertingrule`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_alerting_rule` resource: Kibana Alerting HTTP APIs, composite identity and import, provider-level Kibana connection only, version-gated features, plan-time validation of `params`, stable mapping between Terraform state and API payloads (including avoiding spurious drift on `params` and partially populated read responses), follow-up enable/disable when an update alone does not match desired enabled state, and migration of saved state from older provider formats to the current schema.

## Schema

```hcl
resource "elasticstack_kibana_alerting_rule" "example" {
  # Identity
  id          = <computed, string> # composite: "<space_id>/<rule_id>"; UseStateForUnknown
  rule_id     = <optional, computed, string> # RequiresReplace; UseStateForUnknown; see schema description for Kibana ID format by version
  space_id    = <optional, computed, string> # default "default"; RequiresReplace

  # Rule definition
  name         = <required, string>
  consumer     = <required, string> # RequiresReplace
  rule_type_id = <required, string> # RequiresReplace
  interval     = <required, string> # alerting duration validator (seconds/minutes/hours/days)
  params       = <required, json string, normalized> # validated in ValidateConfig when rule_type_id and params are known

  notify_when = <optional, computed, string> # one of: onActionGroupChange | onActiveAlert | onThrottleInterval
  enabled     = <optional, computed, bool>   # default true
  tags        = <optional, set(string)>
  throttle    = <optional, string>             # alerting duration validator when set

  # Version-gated / API-populated
  alert_delay = <optional, computed, int64> # UseStateForUnknown; server support validated vs stack version

  # Read-only from API
  scheduled_task_id     = <computed, string> # UseStateForUnknown; preserved when API omits on re-read
  last_execution_status = <computed, string>
  last_execution_date   = <computed, string>   # formatted from API timestamp when present

  actions {
    group   = <optional, computed, string> # default "default"
    id      = <required, string>             # connector saved object id
    params  = <required, json string, normalized>

    frequency {
      summary     = <optional, computed, bool>
      notify_when = <optional, computed, string> # same enum as rule-level notify_when
      throttle    = <optional, string>         # alerting duration validator when set
    }
    # If `frequency` is present, `summary` and `notify_when` MUST both be set (objectvalidator.AlsoRequires).

    alerts_filter {
      kql = <optional, string>
      timeframe {
        days        = <optional, list(number)> # elements 1–7 (weekdays)
        timezone    = <optional, string>
        hours_start = <optional, string>       # StringIsHours
        hours_end   = <optional, string>       # StringIsHours
      }
      # If `timeframe` is present, all of days, timezone, hours_start, hours_end MUST be set (AlsoRequires).
    }
  }
}
```

Notes:

- Top-level and nested Markdown descriptions for several attributes are embedded from `internal/kibana/alertingrule/descriptions/*.md` and `resource-description.md` (includes link to Elastic create-rule API docs and an `api_key` auth note for stack 8.8.0+).
- Resource schema version is **1** (`schema.Schema.Version`); state upgrade handles **0 → 1**.

## Requirements

### Requirement: Kibana alerting rule APIs (REQ-001–REQ-004)

The resource SHALL manage rules through Kibana’s alerting rule HTTP API: create rule, get rule, update rule, and delete rule. When an update response does not reflect the desired enabled/disabled state, the provider SHALL perform the follow-up enable or disable operation Kibana expects so the rule ends up in the intended state. Reference: [Create rule API](https://www.elastic.co/guide/en/kibana/master/create-rule-api.html) (as linked from the resource description).

#### Scenario: Create then authoritative read

- GIVEN a successful create API response
- WHEN create completes
- THEN the provider SHALL re-fetch the rule with get and SHALL fail with an error if the rule cannot be read back

#### Scenario: Update then authoritative read

- GIVEN a successful update API response
- WHEN update completes
- THEN the provider SHALL re-fetch the rule with get and SHALL fail with an error if the rule cannot be read back

#### Scenario: Delete idempotency

- GIVEN delete is called
- WHEN the API returns not found for the rule
- THEN the provider SHALL treat delete as successful (no error diagnostic)

#### Scenario: Read removes missing rules

- GIVEN a read/refresh
- WHEN get returns not found
- THEN the provider SHALL remove the resource from state

### Requirement: API error surfacing (REQ-005)

For create, update, and read, when the request fails at the transport layer or the API returns an unexpected HTTP status or an empty successful body where rule data is required, the resource SHALL surface clear error diagnostics to Terraform. Delete SHALL surface errors except when the rule is already absent (not found), which SHALL be treated as success.

#### Scenario: Non-success create/update/read

- GIVEN a non-success response (other than read not-found handled above)
- WHEN the operation completes
- THEN Terraform SHALL receive error diagnostics describing the failure

### Requirement: Provider configuration and Kibana client (REQ-006)

On create, read, update, and delete, if the provider did not supply a usable API client for this resource, the resource SHALL return a configuration error diagnostic. The resource SHALL use the provider’s configured Kibana HTTP client for all operations and SHALL fail with a diagnostic if that client cannot be obtained.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

### Requirement: Stack version for feature gates (REQ-007)

Before create and update, the resource SHALL obtain the Elastic stack version from the provider and SHALL fail if that version cannot be read when needed to evaluate feature support. The compatibility rules in the **Compatibility** requirements SHALL use that version.

#### Scenario: Version fetch failure

- GIVEN the provider cannot determine the stack version when create or update needs it
- WHEN create or update runs
- THEN the provider SHALL surface diagnostics and SHALL not proceed to the alerting rule API for that operation

### Requirement: Identity and composite `id` (REQ-008–REQ-010)

After a successful read from the API, the resource SHALL set `id` to the composite string `<space_id>/<rule_id>`, where `space_id` is the Kibana space and `rule_id` is the rule id returned by the API. On update, the resource SHALL keep the same `id`, `rule_id`, and `space_id` as in prior state when reconciling from the API. A computed `id` that is unknown at plan time SHALL not by itself force replacement of an existing managed rule.

#### Scenario: State matches composite id

- GIVEN a rule in state returned by the API
- WHEN state is written
- THEN `id` SHALL equal `<space_id>/<rule_id>` for that rule

### Requirement: Import (REQ-011–REQ-012)

The resource SHALL support Terraform import using an id of the form `<space_id>/<rule_id>` (exactly one `/` separating two non-empty segments). On success, it SHALL populate `space_id` and `rule_id` from those segments and set `id` to the full import string. If the import id is not in that form, the provider SHALL return an error diagnostic describing the required format.

#### Scenario: Valid import id

- GIVEN an import id `my-space/my-rule`
- WHEN import runs
- THEN state SHALL hold `space_id = "my-space"`, `rule_id = "my-rule"`, and `id = "my-space/my-rule"`

### Requirement: Lifecycle — force replacement (REQ-013)

Changing any of `rule_id`, `space_id`, `consumer`, or `rule_type_id` SHALL require destroying and recreating the resource rather than an in-place update.

#### Scenario: Replace on immutable field

- GIVEN an in-place plan change only to `consumer`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create) for the resource

### Requirement: Compatibility — `notify_when` before 8.6 (REQ-014)

When the stack version is strictly below **8.6.0**, if `notify_when` is unknown, null, or empty string, create/update SHALL fail with the diagnostic summary `notify_when is required until v8.6`.

#### Scenario: Old stack without notify_when

- GIVEN server version &lt; 8.6.0 and `notify_when` not set to a non-empty value
- WHEN create or update runs
- THEN the provider SHALL return the required `notify_when` error

### Requirement: Compatibility — per-action `frequency` (REQ-015)

When the stack version is strictly below **8.6.0**, if any `actions[*].frequency` block is non-null with known values, create/update SHALL fail with an error stating that `actions.frequency` is only supported for Kibana v8.6 or higher.

#### Scenario: Frequency on old stack

- GIVEN server version &lt; 8.6.0 and a configured `frequency` object
- WHEN create or update runs
- THEN the provider SHALL return the frequency unsupported error

### Requirement: Compatibility — `actions.alerts_filter` (REQ-016)

When the stack version is strictly below **8.9.0**, if any `actions[*].alerts_filter` object is non-null with known values, create/update SHALL fail with an error stating that `actions.alerts_filter` is only supported for Kibana v8.9 or higher.

#### Scenario: Alerts filter on old stack

- GIVEN server version &lt; 8.9.0 and a configured `alerts_filter` object
- WHEN create or update runs
- THEN the provider SHALL return the alerts_filter unsupported error

### Requirement: Compatibility — `alert_delay` (REQ-017)

When the stack version is strictly below **8.13.0**, if `alert_delay` is known and non-null, create/update SHALL fail with a diagnostic that states `alert_delay` is not supported on older stacks and names the minimum version the provider enforces (**8.13**).

#### Scenario: Alert delay on old stack

- GIVEN server version &lt; 8.13.0 and `alert_delay` set
- WHEN create or update runs
- THEN the provider SHALL return the alert_delay unsupported error

### Requirement: Plan-time `params` validation (REQ-018–REQ-021)

When both `params` and `rule_type_id` are known during configuration validation (before apply), the resource SHALL parse `params` as JSON. If parsing fails, it SHALL report an attribute error on `params` with summary `Invalid params JSON`.

For **supported** rule types (those for which the provider encodes a known params shape), the resource SHALL verify that `params` is a JSON object that matches that shape: required fields for that rule type MUST be present, property names outside the allowed set MUST be rejected, and where a rule type allows more than one payload shape (for example DSL vs KQL vs ESQL for the same `rule_type_id`), the value MUST match exactly one of those shapes. The provider MAY allow specific extra property names when Kibana accepts them but the encoded shape does not yet list them. If validation fails, the resource SHALL report an attribute error on `params` whose summary is `Invalid params for rule_type_id "<id>"` and whose detail explains what was wrong.

For **unsupported** rule types (any `rule_type_id` the provider does not treat as having a known params shape), the resource SHALL perform no structural check of `params` beyond JSON syntax when known; compatibility with Kibana is left to the API.

At apply time, if `params` cannot be decoded as JSON for the request, the resource SHALL surface a diagnostic rather than calling the API with invalid JSON. Structural rules checked at plan time for supported types MUST NOT need to be re-stated as duplicate diagnostics on a normal successful plan.

#### Scenario: Invalid JSON

- GIVEN known `params` that are not valid JSON
- WHEN configuration is validated before apply
- THEN the provider SHALL report an error on `params` with `Invalid params JSON`

#### Scenario: Unsupported rule type passes structural checks

- GIVEN a `rule_type_id` the provider does not treat as having a known params shape
- WHEN configuration is validated before apply
- THEN the provider SHALL not reject `params` for unknown property names or missing keys solely because the rule type is outside the provider’s known-params set

#### Scenario: Supported rule type with wrong shape

- GIVEN a supported `rule_type_id` and `params` that are valid JSON but omit a required key or include a disallowed key
- WHEN configuration is validated before apply
- THEN the provider SHALL report an error on `params` referencing that `rule_type_id` and describing the validation failure

### Requirement: `.index-threshold` param defaults in API model (REQ-022)

When creating or updating a rule with `rule_type_id` **`.index-threshold`**, if the practitioner omits or nulls `groupBy` in `params`, the provider SHALL send `groupBy` as **`"all"`** so Kibana receives a defined value. If `aggType` is omitted or null, the provider SHALL send **`"count"`** as `aggType` and SHALL not send `aggField`, matching Kibana’s expectations for a count aggregation.

#### Scenario: Threshold rule without groupBy/aggType

- GIVEN `.index-threshold` and params without `groupBy` and without `aggType`
- WHEN the rule is created or updated against Kibana
- THEN the request SHALL include `groupBy: "all"` and `aggType: "count"` and SHALL omit `aggField`

### Requirement: State mapping — `params` drift control (REQ-023)

When writing `params` to state after a read, the provider SHALL avoid **spurious perpetual drift**: if Kibana returns keys the practitioner never represented in the prior `params` value (for example server-injected defaults), those keys SHALL NOT appear in the new state value unless they were already present in prior state for the same nested path. The comparison SHALL follow the structure of JSON objects and arrays so nested injected defaults are treated the same way. On the first persist after create, when there is no meaningful prior `params` to compare, the provider SHALL store the API payload as returned (subject to successful JSON serialization). If reconciling shapes fails in an unexpected way, the provider SHALL fall back to storing the API params rather than dropping user data. If the reconciled value cannot be serialized into state, the provider SHALL return a diagnostic.

#### Scenario: Kibana-injected default keys

- GIVEN prior state `params` omits a key that Kibana adds when returning the rule
- WHEN read maps the API response to state
- THEN stored `params` SHALL still omit that key so the next plan shows no unnecessary change solely because of that key

#### Scenario: First read after create

- GIVEN a newly created rule and no prior `params` in state to compare
- WHEN the provider writes state from the API response
- THEN `params` SHALL reflect the rule as returned by Kibana without stripping keys solely for drift control

### Requirement: State mapping — tags, throttle, notify_when, enabled (REQ-024)

From API to state: non-empty `tags` SHALL become a non-null set; an empty tag list SHALL become a null set. If the API returns a throttle value, state SHALL hold that string; otherwise `throttle` SHALL be null. An empty or absent `notify_when` from the API SHALL become null in state. If the response does not convey `enabled`, state SHALL treat the rule as enabled (`true`).

#### Scenario: Empty tags from API

- GIVEN the API returns no tags
- WHEN state is written
- THEN `tags` SHALL be null (not an empty set)

### Requirement: State mapping — execution metadata (REQ-025)

`last_execution_status` SHALL reflect the rule’s last execution status from the API when present. When the API provides a last-execution timestamp in RFC3339 form, `last_execution_date` SHALL be a non-null string in state derived from that instant using a single consistent formatting rule for this resource; if the timestamp is missing or cannot be parsed, `last_execution_date` SHALL be null.

#### Scenario: Execution fields from API read

- GIVEN a get-rule response with non-empty `execution_status.status` and a parseable `last_execution_date` in RFC3339 form
- WHEN the provider maps the response to Terraform state
- THEN `last_execution_status` SHALL match the API status and `last_execution_date` SHALL be a stable string representation of that time

### Requirement: State mapping — `scheduled_task_id` and `alert_delay` (REQ-026)

If the API returns `scheduled_task_id`, state SHALL store it. If the API omits `scheduled_task_id` and the prior state value was unknown, the provider SHALL set null; if the API omits it and the prior value was known, the provider SHALL keep the prior known value so a partial response does not erase it. The same persistence rules SHALL apply to `alert_delay` when the API returns or omits the rule’s alert-delay value.

#### Scenario: Preserve known scheduled_task_id when API omits field

- GIVEN prior state has a known non-null `scheduled_task_id` and the API response omits `scheduled_task_id`
- WHEN the provider maps the response to state
- THEN `scheduled_task_id` in state SHALL remain the prior known value

### Requirement: Actions mapping (REQ-027)

When reading from Kibana, each action SHALL appear in state as a list element with connector id, JSON `params`, optional `group` (defaulting to **`"default"`** when the API omits it), and optional nested `frequency` and `alerts_filter` blocks matching Terraform’s schema. The API’s nested hour range SHALL map to `hours_start` and `hours_end` in state; weekday constraints SHALL map to the `days` list. When writing to Kibana, `frequency` SHALL be sent only when both `summary` and `notify_when` are set; throttle on the frequency SHALL be sent when non-empty. When the practitioner configures `alerts_filter`, the provider SHALL send filter and timeframe fields that correspond to the configured KQL and timeframe values.

#### Scenario: API action without group uses default

- GIVEN an API action with no `group` field (or empty)
- WHEN the provider maps actions to Terraform state
- THEN the stored `group` attribute SHALL be `"default"`

### Requirement: Interval and throttle validation (REQ-028)

`interval`, rule-level `throttle`, and action-level duration fields (including frequency throttle) SHALL accept only strings in the alerting duration forms Kibana uses (seconds, minutes, hours, days). `hours_start` and `hours_end` SHALL be valid 24-hour clock times (`hh:mm`). Each element of `timeframe.days` SHALL be an integer from **1** through **7** inclusive (weekday indices).

#### Scenario: Invalid interval rejected at plan

- GIVEN `interval` is set to a string that is not a valid alerting duration
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for `interval`

### Requirement: State upgrade v0 → v1 (REQ-029–REQ-034)

Saved state at schema version **0** (legacy provider format) SHALL be upgraded automatically to version **1** when loaded:

- Missing or nil raw state SHALL produce an error diagnostic `Invalid raw state`.
- If stored JSON cannot be unmarshaled, the provider SHALL error with `Failed to unmarshal raw state`.
- If `id` is not already composite (`<space_id>/<rule_id>`), the provider SHALL rewrite it to that form using `space_id` from state (default **`"default"`**) and `rule_id` from state when set, otherwise the previous bare `id` value.
- Empty string values for `notify_when` and `throttle` SHALL become null in upgraded state.
- For each action, nested `frequency`, `alerts_filter`, and `timeframe` values that were stored as single-element lists in v0 SHALL become the single nested object shape expected by the current schema (or null when the list was empty).
- If upgrade changes state and re-serialization fails, the provider SHALL error with `Failed to marshal upgraded state`.

#### Scenario: Legacy non-composite id

- GIVEN v0 state with `id` equal to a bare rule id and `space_id` set
- WHEN upgrade runs
- THEN upgraded state SHALL have composite `id` `<space_id>/<rule_id>`

### Requirement: Authenticated access note (REQ-035)

The resource documentation embedded in schema SHALL state that `api_key` authentication for alerting rules is only supported from Elastic stack **8.8.0** and SHALL describe the class of error when unsupported.

#### Scenario: Schema description mentions api_key support window

- GIVEN the resource schema’s embedded Markdown description
- WHEN a reader inspects provider documentation for this resource
- THEN the text SHALL mention stack **8.8.0** (or later) for `api_key` support and SHALL mention the unsupported-scheme class of error

## Traceability (implementation index)

| Area | Primary files |
|------|----------------|
| Schema | `schema.go` |
| Metadata / Configure / Import | `resource.go` |
| CRUD orchestration | `create.go`, `read.go`, `update.go`, `delete.go` |
| Model mapping, version gates, params normalization | `models.go` |
| Params validate config | `validate.go` |
| State upgrade | `state_upgrade.go` |
| HTTP client, enable/disable, response parsing | `internal/clients/kibanaoapi/alerting_rule.go` |
| Composite id parsing | `internal/clients/api_client.go` (`CompositeID`, `CompositeIDFromStr`, `CompositeIDFromStrFw`) |
