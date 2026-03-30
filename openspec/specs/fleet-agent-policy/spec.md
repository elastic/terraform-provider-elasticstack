# `elasticstack_fleet_agent_policy` — Schema and Functional Requirements

Resource implementation: `internal/fleet/agentpolicy`

## Purpose

Define schema and behavior for the Fleet agent policy resource: API usage, identity and import, lifecycle, compatibility version gates, and the mapping between Terraform configuration and the Fleet Agent Policy API.

## Schema

```hcl
resource "elasticstack_fleet_agent_policy" "example" {
  id        = <computed, string> # same as policy_id
  policy_id = <optional+computed, string> # force new; stable policy identifier

  name      = <required, string>
  namespace = <required, string>

  description          = <optional, string>
  data_output_id       = <optional, string>
  monitoring_output_id = <optional, string>
  fleet_server_host_id = <optional, string>
  download_source_id   = <optional, string>

  monitor_logs    = <optional+computed, bool>  # default false
  monitor_metrics = <optional+computed, bool>  # default false

  skip_destroy = <optional, bool> # when true, destroy removes from state only

  host_name_format = <optional+computed, string> # "hostname" (default) or "fqdn"; min 8.7.0

  supports_agentless = <optional, bool> # min 8.15.0

  sys_monitoring = <optional, bool> # force new; passed as query param at create

  inactivity_timeout   = <optional+computed, duration string> # min 8.7.0
  unenrollment_timeout = <optional+computed, duration string> # min 8.15.0

  global_data_tags = <optional+computed, map(object)> { # min 8.15.0
    "<tag_name>" = {
      string_value = <optional, string> # mutually exclusive with number_value
      number_value = <optional, float32> # mutually exclusive with string_value
    }
  }

  space_ids = <optional+computed, set(string)> # min 9.1.0

  required_versions = <optional+computed, map(int32)> # version → % (0–100); min 9.1.0

  advanced_settings = <optional+computed, object> { # min 8.17.0
    logging_level                  = <optional+computed, string> # debug|info|warning|error; default "info"
    logging_to_files               = <optional+computed, bool>   # default true
    logging_files_interval         = <optional+computed, duration string> # default "30s"
    logging_files_keepfiles        = <optional+computed, int32>  # default 7
    logging_files_rotateeverybytes = <optional+computed, int64>  # default 10485760
    logging_metrics_period         = <optional+computed, duration string> # default "30s"
    go_max_procs                   = <optional+computed, int32>  # default 0 (all CPUs)
    download_timeout               = <optional+computed, duration string> # default "2h"
    download_target_directory      = <optional+computed, string>
    monitoring_runtime_experimental = <optional+computed, string> # ""|"process"|"otel"
  }

  advanced_monitoring_options = <optional+computed, object> { # min 8.16.0
    http_monitoring_endpoint = <optional+computed, object> {
      enabled        = <optional+computed, bool>  # default false
      host           = <optional+computed, string> # default "localhost"
      port           = <optional+computed, int32>  # default 6791; 0–65535
      buffer_enabled = <optional+computed, bool>  # default false
      pprof_enabled  = <optional+computed, bool>  # default false
    }
    diagnostics = <optional+computed, object> {
      rate_limits = <optional+computed, object> {
        interval = <optional+computed, duration string> # default "1m"
        burst    = <optional+computed, int32>           # default 1
      }
      file_uploader = <optional+computed, object> {
        init_duration    = <optional+computed, duration string> # default "1s"
        backoff_duration = <optional+computed, duration string> # default "1m"
        max_retries      = <optional+computed, int32>           # default 10
      }
    }
  }
}
```

## Requirements

### Requirement: Fleet agent policy CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Fleet Create Agent Policy API (`POST /api/fleet/agent_policies`) to create agent policies. The resource SHALL use the Fleet Get Agent Policy API (`GET /api/fleet/agent_policies/{agentPolicyId}`) to read agent policies. The resource SHALL use the Fleet Update Agent Policy API (`PUT /api/fleet/agent_policies/{agentPolicyId}`) to update agent policies. The resource SHALL use the Fleet Delete Agent Policy API (`POST /api/fleet/agent_policies/delete`) to delete agent policies. When the Fleet API returns a non-success status for create, update, read, or delete operations (other than not found on read or delete), the resource SHALL surface the API error to Terraform diagnostics.

#### Scenario: API failure on create

- GIVEN the Fleet Create API returns a non-success status
- WHEN create runs
- THEN Terraform diagnostics SHALL include the API error

#### Scenario: Not found on delete is ignored

- GIVEN the Fleet Delete API returns 404
- WHEN delete runs
- THEN the resource SHALL treat the response as success and return no error

### Requirement: Identity (REQ-005)

The resource SHALL expose a computed `id` attribute whose value is set to the policy ID returned by the Fleet API. The resource SHALL also expose a computed `policy_id` attribute set to the same policy ID value. Both `id` and `policy_id` SHALL be equal to the API-assigned policy identifier after create or update.

#### Scenario: ID and policy_id equality

- GIVEN a successful create
- WHEN state is persisted
- THEN `id` and `policy_id` SHALL both equal the API-assigned policy ID

### Requirement: Import (REQ-006)

The resource SHALL support import. When the import ID is a composite string in the format `<space_id>/<policy_id>` (as produced by `clients.CompositeIDFromStrFw`), the resource SHALL set `policy_id` to the parsed resource ID and `space_ids` to `[<space_id>]` in state. When the import ID is a plain (non-composite) string, the resource SHALL treat the entire string as `policy_id` and SHALL NOT set `space_ids` from the import ID.

#### Scenario: Composite import ID

- GIVEN import with ID `"my-space/abc123"`
- WHEN import runs
- THEN `policy_id` SHALL be `"abc123"` and `space_ids` SHALL contain `"my-space"`

#### Scenario: Plain import ID

- GIVEN import with ID `"abc123"` (no `/` separator recognizable as composite)
- WHEN import runs
- THEN `policy_id` SHALL be `"abc123"` and `space_ids` SHALL NOT be set from the import ID

### Requirement: Lifecycle — force new attributes (REQ-007)

Changing `policy_id` SHALL require resource replacement. Changing `sys_monitoring` SHALL require resource replacement.

#### Scenario: policy_id change triggers replacement

- GIVEN an existing resource with `policy_id = "old-id"`
- WHEN `policy_id` is changed in config
- THEN Terraform SHALL plan a resource replacement

#### Scenario: sys_monitoring change triggers replacement

- GIVEN an existing resource
- WHEN `sys_monitoring` is changed
- THEN Terraform SHALL plan a resource replacement

### Requirement: Connection — provider client (REQ-008)

The resource SHALL use the provider-level Fleet client for all API calls. There is no resource-level connection override for this resource.

#### Scenario: Provider client used

- GIVEN a valid provider configuration
- WHEN any CRUD operation runs
- THEN the resource SHALL obtain its Fleet client from the provider

### Requirement: Compatibility — version-gated features (REQ-009–REQ-017)

When `supports_agentless` is configured, the resource SHALL verify the stack version is at least 8.15.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. When `inactivity_timeout` is configured, the resource SHALL verify the stack version is at least 8.7.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. When `unenrollment_timeout` is configured, the resource SHALL verify the stack version is at least 8.15.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. When `global_data_tags` is configured with one or more entries, the resource SHALL verify the stack version is at least 8.15.0, and if it is lower the resource SHALL fail with a "global_data_tags ES version error". When `host_name_format` is set to `"fqdn"`, the resource SHALL verify the stack version is at least 8.7.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. When `space_ids` is configured, the resource SHALL verify the stack version is at least 9.1.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. When `required_versions` is configured, the resource SHALL verify the stack version is at least 9.1.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. When `advanced_monitoring_options` is configured, the resource SHALL verify the stack version is at least 8.16.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error. When `advanced_settings` is configured, the resource SHALL verify the stack version is at least 8.17.0, and if it is lower the resource SHALL fail with an "Unsupported Elasticsearch version" error.

#### Scenario: global_data_tags on unsupported version

- GIVEN stack version < 8.15.0
- WHEN `global_data_tags` is set with one or more entries
- THEN the resource SHALL fail with a version error before calling the API

#### Scenario: host_name_format fqdn on unsupported version

- GIVEN stack version < 8.7.0
- WHEN `host_name_format = "fqdn"` is configured
- THEN the resource SHALL fail with an "Unsupported Elasticsearch version" error

#### Scenario: Supported version passes gate

- GIVEN stack version meets the minimum for a feature
- WHEN the feature is configured
- THEN the resource SHALL proceed with the API call

### Requirement: Create behavior (REQ-018–REQ-020)

On create, the resource SHALL build a `PostFleetAgentPolicies` request body from the plan model, applying all applicable version gates. The resource SHALL pass `sys_monitoring` as a query parameter (`sys_monitoring=true/false`) to the Create API. After a successful create response, if the response contains a valid policy ID, the resource SHALL perform a read-back (Get Agent Policy) to populate the full state, and SHALL use the read-back response in preference to the create response for state population.

#### Scenario: Read-back after create

- GIVEN a successful create response with a non-empty policy ID
- WHEN create completes
- THEN the resource SHALL call Get Agent Policy and use the result to populate state

#### Scenario: sys_monitoring query param

- GIVEN `sys_monitoring = true`
- WHEN create runs
- THEN the Create API SHALL be called with query parameter `sys_monitoring=true`

### Requirement: Update behavior (REQ-021–REQ-022)

On update, the resource SHALL read the current policy from the API (using the operational space from state) before building the update body, in order to preserve existing agent features not managed by this resource. The resource SHALL then submit a `PutFleetAgentPoliciesAgentpolicyid` request merging the new `host_name_format`-derived feature with any preserved existing agent features.

#### Scenario: Agent features preserved on update

- GIVEN the existing policy has agent features not related to `fqdn`
- WHEN update runs
- THEN those existing features SHALL be preserved in the update body alongside the managed `fqdn` feature

### Requirement: Read behavior (REQ-023–REQ-024)

On read (refresh), the resource SHALL determine the operational space by reading `space_ids` from state (not plan) and using the first space ID if present. The resource SHALL call Get Agent Policy using the `policy_id` from state. If the API returns nil (policy not found), the resource SHALL remove itself from state. Otherwise, the resource SHALL populate state from the API response.

#### Scenario: Not found removes from state

- GIVEN the policy no longer exists in Fleet
- WHEN read runs
- THEN the resource SHALL call `resp.State.RemoveResource` and not return an error

#### Scenario: Operational space from state

- GIVEN `space_ids` in state contains `["space-a"]`
- WHEN read or update runs
- THEN the API call SHALL use `space-a` as the space context

### Requirement: Delete behavior (REQ-025–REQ-026)

On delete, the resource SHALL read `policy_id` and `skip_destroy` from state. When `skip_destroy` is `true`, the resource SHALL skip the delete API call and remove only the Terraform state entry. When `skip_destroy` is `false` or not set, the resource SHALL call the Fleet Delete Agent Policy API. The resource SHALL determine the operational space from state for the delete call.

#### Scenario: skip_destroy suppresses delete

- GIVEN `skip_destroy = true`
- WHEN destroy runs
- THEN the Fleet Delete API SHALL NOT be called and the resource SHALL be removed from state

#### Scenario: Normal delete calls Fleet API

- GIVEN `skip_destroy = false`
- WHEN destroy runs
- THEN the Fleet Delete Agent Policy API SHALL be called with the policy ID

### Requirement: Space-aware API routing (REQ-027)

All API calls (create, read, update, delete) SHALL include a space-aware path prefix (`/s/{spaceID}`) when the operational space ID is non-empty and not `"default"`. When the space ID is empty or `"default"`, the resource SHALL use the standard API path without a space prefix.

#### Scenario: Non-default space routing

- GIVEN operational space ID is `"my-space"`
- WHEN any API call is made
- THEN the request path SHALL be prefixed with `/s/my-space`

#### Scenario: Default space routing

- GIVEN operational space ID is `""` or `"default"`
- WHEN any API call is made
- THEN the request path SHALL NOT include a space prefix

### Requirement: Mapping — monitoring flags (REQ-028)

The resource SHALL map `monitor_logs = true` to the `logs` entry in the `monitoring_enabled` array sent to the API, and `monitor_metrics = true` to the `metrics` entry. On read, if the API response `monitoring_enabled` contains `"logs"`, the resource SHALL set `monitor_logs = true` in state; otherwise it SHALL set `monitor_logs = false`. The same SHALL apply to `monitor_metrics` for `"metrics"`.

#### Scenario: Monitor logs on read

- GIVEN the API response includes `"logs"` in `monitoring_enabled`
- WHEN read populates state
- THEN `monitor_logs` SHALL be `true` in state

### Requirement: Mapping — host_name_format via agent features (REQ-029)

The resource SHALL represent `host_name_format` using the `agent_features` API field. When `host_name_format = "fqdn"`, the resource SHALL send `{"name": "fqdn", "enabled": true}` in `agent_features`. When `host_name_format = "hostname"`, the resource SHALL send `{"name": "fqdn", "enabled": false}`. On read, if `agent_features` contains an entry with `name = "fqdn"` and `enabled = true`, the resource SHALL set `host_name_format = "fqdn"` in state; otherwise it SHALL set `host_name_format = "hostname"`.

#### Scenario: FQDN feature mapping

- GIVEN `host_name_format = "fqdn"` in plan
- WHEN create or update runs
- THEN the API request SHALL include `agent_features: [{"name": "fqdn", "enabled": true}]`

#### Scenario: Hostname feature mapping

- GIVEN `host_name_format = "hostname"` in plan
- WHEN create or update runs
- THEN the API request SHALL include `agent_features: [{"name": "fqdn", "enabled": false}]`

### Requirement: Mapping — duration fields (REQ-030)

The resource SHALL accept `inactivity_timeout` and `unenrollment_timeout` as duration strings (e.g. `"30s"`, `"2m"`, `"1h"`). When sending to the API, these SHALL be converted to seconds as a float32. On read, the API-returned seconds value SHALL be converted back to a duration string and stored in state. When the API returns `null` for either field, the resource SHALL store a null value in state.

#### Scenario: Duration round-trip

- GIVEN `inactivity_timeout = "5m"` in config
- WHEN create runs
- THEN the API SHALL receive `inactivity_timeout = 300` (seconds)
- AND on subsequent read the state SHALL contain `inactivity_timeout = "5m0s"`

### Requirement: Mapping — global_data_tags (REQ-031)

The resource SHALL map each `global_data_tags` entry to a `{name, value}` object in the Fleet API's `global_data_tags` array. A tag entry with `string_value` set SHALL use the string variant of the API value union. A tag entry with `number_value` set SHALL use the numeric variant. On read, the resource SHALL reconstruct the map from the API array, matching each entry by name and placing the value into the appropriate `string_value` or `number_value` field. When `global_data_tags` is empty and the stack supports global data tags (≥ 8.15.0), the resource SHALL send an empty array to the API.

#### Scenario: String tag round-trip

- GIVEN `global_data_tags = { env = { string_value = "prod" } }`
- WHEN create runs
- THEN the API SHALL receive `[{"name": "env", "value": "prod"}]`

### Requirement: Mapping — required_versions (REQ-032)

The resource SHALL map `required_versions` entries to the Fleet API's `required_versions` array as `{version, percentage}` objects. Each map key is the version string and each value (0–100 integer) is the upgrade percentage. On read, the resource SHALL reconstruct the map by rounding the float32 percentage from the API response to the nearest integer. When `required_versions` is an empty map, the resource SHALL send an empty array to the API to clear any existing upgrade targets. When the API returns `null` for `required_versions`, the resource SHALL store a null map in state.

#### Scenario: Empty required_versions clears upgrades

- GIVEN `required_versions = {}`
- WHEN create or update runs
- THEN the API SHALL receive `required_versions: []`

### Requirement: Mapping — advanced_settings (REQ-033)

The resource SHALL map `advanced_settings` attributes to the Fleet API's `advanced_settings` object using the key names defined in the API schema (e.g. `agent_logging_level`, `agent_limits_go_max_procs`). Only attributes that are known (not null or unknown) SHALL be included in the API payload. On read, if the API response does not include `advanced_settings`, the resource SHALL store a null object in state.

#### Scenario: Null advanced_settings on read

- GIVEN the API response has no `advanced_settings` field
- WHEN read runs
- THEN `advanced_settings` SHALL be null in state

### Requirement: Mapping — advanced_monitoring_options (REQ-034)

The resource SHALL map `advanced_monitoring_options.http_monitoring_endpoint` to the API fields `monitoring_http` (enabled, host, port, buffer) and `monitoring_pprof_enabled`. The resource SHALL map `advanced_monitoring_options.diagnostics` to the API field `monitoring_diagnostics` (limit.interval, limit.burst, uploader.init_dur, uploader.max_dur, uploader.max_retries). On read, if none of `monitoring_http`, `monitoring_pprof_enabled`, or `monitoring_diagnostics` are present in the API response, the resource SHALL store a null object in state for `advanced_monitoring_options`.

#### Scenario: Null advanced_monitoring_options on read

- GIVEN the API response contains no monitoring HTTP, pprof, or diagnostics fields
- WHEN read runs
- THEN `advanced_monitoring_options` SHALL be null in state

### Requirement: Plan modifiers — use state for unknown (REQ-035)

The attributes `id`, `policy_id`, `inactivity_timeout`, `unenrollment_timeout`, `space_ids`, `required_versions`, `advanced_settings`, and `advanced_monitoring_options` SHALL use `UseStateForUnknown` plan modifiers so that known prior state values are preserved during plan when these attributes are not explicitly changed.

#### Scenario: Stable policy_id across plans

- GIVEN a resource with a known `policy_id` in state
- WHEN a plan is generated without changing `policy_id`
- THEN `policy_id` SHALL not be shown as unknown in the plan
