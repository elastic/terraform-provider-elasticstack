# `elasticstack_kibana_slo` — Schema and Functional Requirements

Resource implementation: `internal/kibana/slo`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_slo` resource: Kibana SLO HTTP APIs, composite identity and import, provider-level Kibana connection only, version-gated features (`group_by`, multiple `group_by`, `prevent_initial_backfill`, and `data_view_id`), exactly one indicator block enforced at config-validation time, read-after-create and read-after-update for computed fields, and migration of saved state from schema versions 0 and 1 to the current schema version 2.

## Schema

```hcl
resource "elasticstack_kibana_slo" "example" {
  # Identity
  id       = <computed, string>                   # composite: "<space_id>/<slo_id>"; UseStateForUnknown
  slo_id   = <optional, computed, string>         # 8–48 chars, [a-zA-Z0-9_-]; RequiresReplace; UseStateForUnknown
  space_id = <optional, computed, string>         # default "default"; RequiresReplace; UseStateForUnknown

  # Core definition
  name             = <required, string>
  description      = <required, string>
  budgeting_method = <required, string>           # one of: "occurrences" | "timeslices"

  # Grouping and tagging
  group_by = <optional, computed, list(string)>   # custom GroupByType; elements must be non-empty strings
  tags     = <optional, list(string)>

  time_window {                                   # exactly 1 block (SizeBetween 1,1)
    duration = <required, string>
    type     = <required, string>
  }

  objective {                                     # exactly 1 block (SizeBetween 1,1)
    target           = <required, float64>
    timeslice_target = <optional, float64>
    timeslice_window = <optional, string>
  }

  settings {                                      # optional single nested block; UseStateForUnknown on the object
    sync_delay              = <optional, computed, string>
    frequency               = <optional, computed, string>
    prevent_initial_backfill = <optional, computed, bool>  # requires stack >= 8.15.0
  }

  # Exactly one of the following indicator blocks must be specified
  # (enforced by ExactlyOneOf config validator)

  apm_latency_indicator {                         # exactly 1 block when present
    index            = <required, string>
    filter           = <optional, string>
    service          = <required, string>
    environment      = <required, string>
    transaction_type = <required, string>
    transaction_name = <required, string>
    threshold        = <required, int64>
  }

  apm_availability_indicator {                    # exactly 1 block when present
    index            = <required, string>
    filter           = <optional, string>
    service          = <required, string>
    environment      = <required, string>
    transaction_type = <required, string>
    transaction_name = <required, string>
  }

  kql_custom_indicator {                          # exactly 1 block when present
    index           = <required, string>
    data_view_id    = <optional, string>          # requires stack >= 8.15.0
    filter          = <optional, string>
    good            = <optional, computed, string> # default ""
    total           = <optional, computed, string> # default ""
    timestamp_field = <optional, computed, string> # default "@timestamp"
  }

  metric_custom_indicator {                       # exactly 1 block when present
    index           = <required, string>
    data_view_id    = <optional, string>          # requires stack >= 8.15.0
    filter          = <optional, string>
    timestamp_field = <optional, computed, string> # default "@timestamp"

    good {                                        # exactly 1 block
      equation = <required, string>
      metrics {                                   # at least 1 block
        name        = <required, string>
        aggregation = <required, string>
        field       = <required, string>
        filter      = <optional, string>
      }
    }

    total {                                       # exactly 1 block
      equation = <required, string>
      metrics {                                   # at least 1 block
        name        = <required, string>
        aggregation = <required, string>
        field       = <required, string>
        filter      = <optional, string>
      }
    }
  }

  histogram_custom_indicator {                    # exactly 1 block when present
    index           = <required, string>
    data_view_id    = <optional, string>          # requires stack >= 8.15.0
    filter          = <optional, string>
    timestamp_field = <optional, computed, string> # default "@timestamp"

    good {                                        # exactly 1 block
      aggregation = <required, string>            # one of: "value_count" | "range"
      field       = <required, string>
      filter      = <optional, string>
      from        = <optional, float64>
      to          = <optional, float64>
    }

    total {                                       # exactly 1 block
      aggregation = <required, string>            # one of: "value_count" | "range"
      field       = <required, string>
      filter      = <optional, string>
      from        = <optional, float64>
      to          = <optional, float64>
    }
  }

  timeslice_metric_indicator {                    # exactly 1 block when present
    index           = <required, string>
    data_view_id    = <optional, string>          # requires stack >= 8.15.0
    timestamp_field = <required, string>
    filter          = <optional, string>

    metric {                                      # exactly 1 block
      equation   = <required, string>
      comparator = <required, string>             # one of: "GT" | "GTE" | "LT" | "LTE"
      threshold  = <required, float64>

      metrics {                                   # at least 1 block
        name        = <required, string>
        aggregation = <required, string>          # one of: sum | avg | min | max | value_count | last_value | cardinality | std_deviation | percentile | doc_count
        field       = <optional, string>          # required for all aggregations except doc_count; must NOT be set for doc_count
        percentile  = <optional, float64>         # required when aggregation is "percentile"; must NOT be set otherwise
        filter      = <optional, string>          # supported for all aggregations except doc_count
      }
    }
  }
}
```

Notes:

- Schema version is **2**; state upgraders handle **0 → 2** (via v0→v1 then v1→v2) and **1 → 2**.
- Exactly one indicator block must be present per resource (enforced by `ResourceWithConfigValidators`).

## Requirements

### Requirement: Kibana SLO APIs (REQ-001–REQ-004)

The resource SHALL manage SLOs through Kibana's SLO HTTP APIs: create SLO, get SLO, update SLO, and delete SLO. After create and after update, the resource SHALL perform a read-back (get SLO) to populate computed fields into state; if the SLO cannot be found after create or update the resource SHALL fail with an error.

#### Scenario: Create then authoritative read

- GIVEN a successful create API response
- WHEN create completes
- THEN the provider SHALL re-fetch the SLO with get and SHALL fail with a "SLO not found" error if the SLO cannot be read back

#### Scenario: Update then authoritative read

- GIVEN a successful update API response
- WHEN update completes
- THEN the provider SHALL re-fetch the SLO with get and SHALL fail with a "SLO not found" error if the SLO cannot be read back

#### Scenario: Read removes missing SLOs

- GIVEN a read/refresh
- WHEN get returns HTTP 404 (not found)
- THEN the provider SHALL remove the resource from state

### Requirement: API error surfacing (REQ-005)

For create, update, and read, when the request fails at the transport layer or the API returns an unexpected status, the resource SHALL surface error diagnostics to Terraform. Delete SHALL surface errors from the API.

#### Scenario: Non-success API response

- GIVEN a non-success response from the Kibana SLO API (other than read not-found handled above)
- WHEN the operation completes
- THEN Terraform SHALL receive error diagnostics describing the failure

### Requirement: Provider configuration and Kibana client (REQ-006)

On create, read, update, and delete, if the provider did not supply a usable API client for this resource, the resource SHALL return a configuration error diagnostic ("Provider not configured"). The resource SHALL use the provider's configured Kibana HTTP client for all operations.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

### Requirement: Stack version for feature gates (REQ-007)

Before create and update, the resource SHALL obtain the Elastic Stack version from the provider client and SHALL fail with an error diagnostic if that version cannot be determined. The version SHALL be used to evaluate all feature-compatibility checks defined in the Compatibility requirements.

#### Scenario: Version fetch failure

- GIVEN the provider client cannot determine the Elastic Stack version
- WHEN create or update runs
- THEN the provider SHALL surface diagnostics and SHALL NOT proceed to call the SLO API

### Requirement: Identity and composite `id` (REQ-008–REQ-009)

The resource SHALL expose a computed `id` in the format `<space_id>/<slo_id>` where `space_id` is the Kibana space identifier and `slo_id` is the SLO object id returned by the API. On create the `id` SHALL be built from `space_id` and the `slo_id` returned in the create response. After each read the `id` SHALL be rebuilt from the `space_id` and `slo_id` fields in the API response.

#### Scenario: Composite id after create

- GIVEN a successful create call that returns `slo_id = "my-slo"` for `space_id = "default"`
- WHEN state is written
- THEN `id` SHALL equal `"default/my-slo"`

#### Scenario: Composite id preserved after read

- GIVEN state with `id = "my-space/my-slo"`
- WHEN a read/refresh completes
- THEN `id` SHALL be rebuilt as `<space_id>/<slo_id>` from the API response

### Requirement: Import (REQ-010)

The resource SHALL support Terraform import. Import SHALL use `ImportStatePassthroughID`, passing the raw import value directly as the `id` attribute. The expected format is `<space_id>/<slo_id>` (a composite id), which is parsed at the next read.

#### Scenario: Valid import id

- GIVEN an import id `"default/my-slo"`
- WHEN import runs
- THEN the provider SHALL persist the value as `id` in state, from which the subsequent read derives `space_id` and `slo_id`

### Requirement: Lifecycle — force replacement (REQ-011)

Changing `slo_id` or `space_id` SHALL require destroying and recreating the resource rather than an in-place update.

#### Scenario: Replace on slo_id change

- GIVEN an existing SLO with `slo_id = "old-id"`
- WHEN the practitioner changes `slo_id` to `"new-id"` in configuration
- THEN Terraform SHALL plan a replacement (destroy + create)

#### Scenario: Replace on space_id change

- GIVEN an existing SLO with `space_id = "space-a"`
- WHEN the practitioner changes `space_id` to `"space-b"`
- THEN Terraform SHALL plan a replacement

### Requirement: Exactly one indicator (REQ-012)

The resource SHALL enforce that exactly one of the following blocks is configured: `metric_custom_indicator`, `histogram_custom_indicator`, `apm_latency_indicator`, `apm_availability_indicator`, `kql_custom_indicator`, or `timeslice_metric_indicator`. Configuring zero or more than one SHALL produce a configuration error at plan time via the `ExactlyOneOf` config validator.

#### Scenario: No indicator block

- GIVEN a configuration with no indicator block
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation error

#### Scenario: Multiple indicator blocks

- GIVEN a configuration with two indicator blocks (e.g. both `kql_custom_indicator` and `apm_latency_indicator`)
- WHEN Terraform validates configuration
- THEN the provider SHALL return a validation error

### Requirement: Connection — provider-level Kibana client (REQ-013)

The resource SHALL use the provider's configured Kibana client by default. There is no resource-level connection override for this resource.

#### Scenario: All operations use provider Kibana client

- GIVEN a configured provider with a Kibana client
- WHEN any CRUD operation runs
- THEN the operation SHALL use the provider's Kibana client for all API calls

### Requirement: Compatibility — `group_by` (REQ-014)

When `group_by` is configured with at least one element, the resource SHALL verify the Elastic Stack version is at least **8.10.0**; if it is lower the resource SHALL fail with an "Unsupported Elastic Stack version" error stating that `group_by` requires 8.10.0 or higher. When the stack version is below 8.10.0 and `group_by` is empty or not set, the resource SHALL omit `group_by` from the API request body entirely.

#### Scenario: group_by on old stack

- GIVEN server version &lt; 8.10.0 and `group_by` set to one or more non-empty strings
- WHEN create or update runs
- THEN the provider SHALL return an unsupported version error for `group_by`

#### Scenario: group_by omitted below minimum version

- GIVEN server version &lt; 8.10.0 and `group_by` not configured or empty
- WHEN create or update runs
- THEN `group_by` SHALL NOT appear in the API request body

### Requirement: Compatibility — multiple `group_by` fields (REQ-015)

When `group_by` contains more than one element, the resource SHALL verify the Elastic Stack version is at least **8.14.0**; if it is lower the resource SHALL fail with an "Unsupported Elastic Stack version" error stating that multiple `group_by` fields require 8.14.0 or higher.

#### Scenario: Multiple group_by on old stack

- GIVEN server version &lt; 8.14.0 and `group_by` set to two or more strings
- WHEN create or update runs
- THEN the provider SHALL return an unsupported version error for multiple `group_by` fields

### Requirement: Compatibility — `settings.prevent_initial_backfill` (REQ-016)

When `settings.prevent_initial_backfill` is configured with a known value, the resource SHALL verify the Elastic Stack version is at least **8.15.0**; if it is lower the resource SHALL fail with an "Unsupported Elastic Stack version" error stating that `prevent_initial_backfill` requires 8.15.0 or higher.

#### Scenario: prevent_initial_backfill on old stack

- GIVEN server version &lt; 8.15.0 and `settings.prevent_initial_backfill` set
- WHEN create or update runs
- THEN the provider SHALL return an unsupported version error

### Requirement: Compatibility — `data_view_id` (REQ-017)

When any indicator block has `data_view_id` configured with a non-empty known value, the resource SHALL verify the Elastic Stack version is at least **8.15.0**; if it is lower the resource SHALL fail with an "Unsupported Elastic Stack version" error stating that `data_view_id` is not supported below 8.15.0.

#### Scenario: data_view_id on old stack

- GIVEN server version &lt; 8.15.0 and `data_view_id` set to a non-empty string in any indicator
- WHEN create or update runs
- THEN the provider SHALL return an unsupported version error for `data_view_id`

### Requirement: Create flow (REQ-018)

On create, the resource SHALL convert the Terraform plan to an API model and call the Kibana Create SLO API. If `slo_id` is configured with a non-empty known value, the resource SHALL pass it as the requested SLO id; otherwise the API SHALL generate the id. The `slo_id` returned in the create response SHALL be stored in state as `slo_id`. The resource SHALL then perform a read-back to populate all computed fields before writing final state.

#### Scenario: slo_id omitted — server-generated id used

- GIVEN `slo_id` is not set in configuration
- WHEN create completes successfully
- THEN `slo_id` in state SHALL hold the id returned by the Kibana API

#### Scenario: slo_id provided — sent to API

- GIVEN `slo_id = "custom-id"` in configuration
- WHEN create runs
- THEN the create request body SHALL include `id = "custom-id"`

### Requirement: Update flow (REQ-019)

On update, the resource SHALL convert the Terraform plan to an API model and call the Kibana Update SLO API using the `slo_id` and `space_id` from the current composite `id`. The resource SHALL perform a read-back after a successful update to populate computed fields into state.

#### Scenario: Update calls API and reads back

- GIVEN an existing SLO with a changed `name` in the Terraform plan
- WHEN update runs
- THEN the provider SHALL call the Kibana Update SLO API and SHALL perform a subsequent get to populate computed fields in state

### Requirement: Read flow (REQ-020)

On read, the resource SHALL parse the composite `id` from state to extract `space_id` and `slo_id`, then call the Kibana Get SLO API. If the API returns HTTP 404, the resource SHALL remove itself from state without error. On a successful response, the resource SHALL update all state attributes from the API response.

#### Scenario: Successful read maps all attributes

- GIVEN a valid get-SLO API response
- WHEN read completes
- THEN all attributes (name, description, budgeting_method, time_window, objective, indicator, settings, group_by, tags, slo_id, space_id) SHALL be updated in state from the response

### Requirement: Delete flow (REQ-021)

On delete, the resource SHALL parse the composite `id` from state to extract `space_id` and `slo_id`, then call the Kibana Delete SLO API with both identifiers.

#### Scenario: Delete uses composite id

- GIVEN state with `id = "my-space/my-slo"`
- WHEN delete runs
- THEN the delete request SHALL be made for `slo_id = "my-slo"` in space `"my-space"`

### Requirement: Mapping — `slo_id` and `space_id` defaults (REQ-022)

If `space_id` is not explicitly set in configuration, it SHALL default to `"default"`. If `slo_id` is not set or is empty in configuration, it SHALL be left to the server to generate; after create the server-generated id SHALL be stored as the computed `slo_id` value.

#### Scenario: Default space_id

- GIVEN a configuration without `space_id`
- WHEN the resource is created
- THEN `space_id` in state SHALL be `"default"`

### Requirement: Mapping — `group_by` wire format (REQ-023)

When the stack version is at least **8.14.0** (supports multiple group-by fields), the resource SHALL send `group_by` as a JSON array of strings. When the stack version is at least **8.10.0** but below **8.14.0**, the resource SHALL send `group_by` as a single string (only one element is permitted per REQ-015). When reading from the API, regardless of whether the response contains a string or an array, the resource SHALL normalize `group_by` to a list of strings in state.

#### Scenario: Single group_by on 8.10–8.13

- GIVEN server version in [8.10.0, 8.14.0) and `group_by = ["field.name"]`
- WHEN create or update runs
- THEN the API request SHALL include `group_by` as a single string `"field.name"`

### Requirement: Mapping — `settings` block (REQ-024)

The `settings` block uses `UseStateForUnknown` on its object plan modifier. When the `settings` block is configured, the resource SHALL send `sync_delay`, `frequency`, and `prevent_initial_backfill` (where known) to the API. When the `settings` block is not configured, no settings SHALL be sent. After reading from the API, if the `settings` block was previously configured in state, the resource SHALL update the `settings` object in state from the API response; if it was not configured, `settings` SHALL remain null in state.

#### Scenario: Settings omitted when not configured

- GIVEN a configuration without a `settings` block
- WHEN create runs
- THEN the create request SHALL NOT include a `settings` payload

### Requirement: Mapping — indicator type round-trip (REQ-025)

On read, the resource SHALL map the API indicator type to exactly one indicator block in state and SHALL clear all other indicator blocks to null/empty. The mapping from Kibana API type strings to Terraform blocks SHALL be:

- `sli.apm.transactionDuration` → `apm_latency_indicator`
- `sli.apm.transactionErrorRate` → `apm_availability_indicator`
- `sli.kql.custom` → `kql_custom_indicator`
- `sli.metric.custom` → `metric_custom_indicator`
- `sli.histogram.custom` → `histogram_custom_indicator`
- `sli.metric.timeslice` → `timeslice_metric_indicator`

If the API response contains an indicator type not in this mapping, the resource SHALL return an "Unexpected API response" error.

#### Scenario: Indicator type set to apm_latency

- GIVEN an API response with indicator type `sli.apm.transactionDuration`
- WHEN read maps the response to state
- THEN `apm_latency_indicator` SHALL be populated and all other indicator blocks SHALL be null or empty

### Requirement: Mapping — `tags` and `group_by` null handling (REQ-026)

When the API returns a null or absent `group_by`, the resource SHALL store null for `group_by` in state. When the API returns a non-null `group_by` list (including an empty list), the resource SHALL store it as a list in state. For `tags`, the resource SHALL store null when the API returns no tags, and SHALL store the list of tag strings when tags are present.

#### Scenario: Null group_by from API

- GIVEN an API response with no `group_by` value
- WHEN read maps the response to state
- THEN `group_by` SHALL be null in state

#### Scenario: No tags from API

- GIVEN an API response with no tags
- WHEN read maps the response to state
- THEN `tags` SHALL be null in state

### Requirement: Mapping — `objective` timeslice fields (REQ-027)

When the API returns a null `timeslice_target` or `timeslice_window`, the resource SHALL store null (not zero) for those fields in state.

#### Scenario: timeslice_target absent from API

- GIVEN an API response without `timeslice_target`
- WHEN read maps the response to state
- THEN `objective.timeslice_target` SHALL be null in state

### Requirement: StateUpgrade — v0 to v2 (REQ-028)

Saved state at schema version **0** SHALL be automatically upgraded to version **2** by first applying the v0→v1 migration and then the v1→v2 migration. If the v0→v1 step fails, the v1→v2 step SHALL NOT run. If the raw state or its JSON payload is nil, the upgrader SHALL return an error diagnostic `"Invalid raw state"`. If the JSON cannot be unmarshaled, the upgrader SHALL return an error diagnostic `"Failed to unmarshal raw state"`. If re-serialization fails after changes, the upgrader SHALL return an error diagnostic `"Failed to marshal raw state"`.

#### Scenario: v0 group_by string promoted to list

- GIVEN v0 state with `group_by` stored as a non-empty JSON string `"field.name"`
- WHEN the v0→v1 upgrader runs
- THEN `group_by` in the upgraded state SHALL be `["field.name"]` (a single-element JSON array) and a warning diagnostic SHALL be added

#### Scenario: v0 group_by absent or null — no change

- GIVEN v0 state where `group_by` is null or absent
- WHEN the v0→v1 upgrader runs
- THEN `group_by` SHALL remain null or absent without error

#### Scenario: v0→v2 chains both migrations

- GIVEN v0 state
- WHEN the v0→v2 upgrader runs
- THEN both the v0→v1 and v1→v2 transformations SHALL be applied in order

### Requirement: StateUpgrade — v1 to v2 (REQ-029)

Saved state at schema version **1** SHALL be automatically upgraded to version **2**. The v1→v2 upgrade converts `settings` from a nested block (stored as a list) to a single nested object. If `settings` is absent or null, the upgrader SHALL leave it unchanged. If `settings` is already an object (map), the upgrader SHALL leave it unchanged. If `settings` is a list with one element, the upgrader SHALL replace the list with the first element (the object). If `settings` is an empty list, the upgrader SHALL set `settings` to null. A warning diagnostic `"Upgraded legacy settings state"` SHALL be added when the list form is converted.

#### Scenario: v1 settings empty list becomes null

- GIVEN v1 state with `settings = []`
- WHEN the v1→v2 upgrader runs
- THEN `settings` in upgraded state SHALL be null and a warning diagnostic SHALL be added

#### Scenario: v1 settings single-element list becomes object

- GIVEN v1 state with `settings = [{"sync_delay": "1m", ...}]`
- WHEN the v1→v2 upgrader runs
- THEN `settings` in upgraded state SHALL be the inner object `{"sync_delay": "1m", ...}` and a warning diagnostic SHALL be added

#### Scenario: v1 settings already an object — no change

- GIVEN v1 state with `settings` already in object form
- WHEN the v1→v2 upgrader runs
- THEN `settings` SHALL remain unchanged

### Requirement: Plan/State — `id` and `slo_id` stability (REQ-030)

The `id` attribute SHALL use `UseStateForUnknown` so it is preserved across plans once set. `slo_id` SHALL also use `UseStateForUnknown` so a server-generated id from create is preserved on subsequent plans without requiring replacement.

#### Scenario: slo_id preserved after create

- GIVEN an SLO where `slo_id` was not configured and was server-generated
- WHEN Terraform plans a subsequent non-replacement change
- THEN `slo_id` SHALL show as known (the server-generated value) in the plan, not unknown

## Traceability (implementation index)

| Area | Primary files |
|------|---------------|
| Schema | `internal/kibana/slo/schema.go` |
| Metadata / Configure / Import / Config validators | `internal/kibana/slo/resource.go` |
| Create | `internal/kibana/slo/create.go` |
| Read | `internal/kibana/slo/read.go` |
| Update | `internal/kibana/slo/update.go` |
| Delete | `internal/kibana/slo/delete.go` |
| Model mapping (tfModel ↔ API), indicator mapping | `internal/kibana/slo/models.go` |
| Indicator models | `internal/kibana/slo/models_*_indicator.go` |
| Version constants | `internal/kibana/slo/constants.go` |
| group_by custom type | `internal/kibana/slo/group_by_type.go` |
| State upgrade | `internal/kibana/slo/state_upgrade.go` |
| HTTP client (create/read/update/delete, group_by wire format) | `internal/clients/kibana/slo.go` |
| Composite id parsing | `internal/clients/api_client.go` (`CompositeID`, `CompositeIDFromStrFw`) |
