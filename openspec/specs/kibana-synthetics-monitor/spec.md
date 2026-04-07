# `elasticstack_kibana_synthetics_monitor` — Schema and Functional Requirements

Resource implementation: `internal/kibana/synthetics/monitor`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_synthetics_monitor` resource: Kibana Synthetics Monitor HTTP APIs, composite identity and import, exactly-one monitor-type block validation, at-least-one location requirement, version-gated `labels` attribute, stable mapping between Terraform state and API payloads, and provider-level Kibana connection only.

## Schema

```hcl
resource "elasticstack_kibana_synthetics_monitor" "example" {
  # Identity
  id       = <computed, string>                    # composite: "<space_id>/<monitor_id>"; UseStateForUnknown; RequiresReplace
  space_id = <optional, computed, string>          # Kibana space; UseStateForUnknown; RequiresReplace; default "default"

  # Monitor definition (common fields)
  name      = <required, string>
  namespace = <optional, computed, string>         # data stream namespace; UseStateForUnknown; validated: no *, \, /, ?, ", <, >, |, whitespace, comma, #, :, or -
  schedule  = <optional, computed, int64>          # minutes; UseStateForUnknown; one of: 1, 3, 5, 10, 15, 30, 60, 120, 240

  # At least one of locations or private_locations must be set (AtLeastOneOf validator)
  locations         = <optional, list(string)>     # managed location names (validated against allowed set when location validation is enabled)
  private_locations = <optional, list(string)>     # private location names (by label)

  enabled          = <optional, computed, bool>    # UseStateForUnknown; default true
  tags             = <optional, list(string)>
  labels           = <optional, map(string)>       # key-value labels; requires server version >= 8.16.0
  service_name     = <optional, computed, string>  # APM service name; UseStateForUnknown
  timeout          = <optional, computed, int64>   # seconds; UseStateForUnknown; default 16
  params           = <optional, string>            # JSON (normalized) object; arbitrary monitor parameters; use jsonencode
  retest_on_failure = <optional, bool>

  alert {                                          # optional, computed; UseStateForUnknown; default { status: { enabled: true }, tls: { enabled: true } }
    status {
      enabled = <optional, computed, bool>         # UseStateForUnknown
    }
    tls {
      enabled = <optional, computed, bool>         # UseStateForUnknown
    }
  }

  # Exactly one of the following type blocks must be set (ExactlyOneOf validator)

  http {                                           # optional; HTTP monitor fields
    url                          = <required, string>
    max_redirects                = <optional, computed, int64>    # UseStateForUnknown; default 0
    mode                         = <optional, computed, string>   # UseStateForUnknown; one of: "any", "all"
    ipv4                         = <optional, computed, bool>     # UseStateForUnknown
    ipv6                         = <optional, computed, bool>     # UseStateForUnknown
    username                     = <optional, string>
    password                     = <optional, string>
    proxy_url                    = <optional, computed, string>   # UseStateForUnknown
    proxy_header                 = <optional, string>             # JSON (normalized) object; additional headers for CONNECT proxy requests
    response                     = <optional, string>             # JSON (normalized) object; controls indexing of HTTP response body
    check                        = <optional, string>             # JSON (normalized) object; check request settings
    ssl_verification_mode        = <optional, computed, string>   # UseStateForUnknown
    ssl_supported_protocols      = <optional, computed, list(string)> # UseStateForUnknown
    ssl_certificate_authorities  = <optional, list(string)>
    ssl_certificate              = <optional, computed, string>   # UseStateForUnknown
    ssl_key                      = <optional, computed, string>   # UseStateForUnknown; sensitive
    ssl_key_passphrase           = <optional, computed, string>   # UseStateForUnknown; sensitive
  }

  tcp {                                            # optional; TCP monitor fields
    host                         = <required, string>   # host:port format supported
    check_send                   = <optional, string>   # payload to send
    check_receive                = <optional, string>   # expected response string
    proxy_url                    = <optional, computed, string>   # UseStateForUnknown
    proxy_use_local_resolver     = <optional, computed, bool>     # UseStateForUnknown
    ssl_verification_mode        = <optional, computed, string>   # UseStateForUnknown
    ssl_supported_protocols      = <optional, computed, list(string)> # UseStateForUnknown
    ssl_certificate_authorities  = <optional, list(string)>
    ssl_certificate              = <optional, computed, string>   # UseStateForUnknown
    ssl_key                      = <optional, computed, string>   # UseStateForUnknown; sensitive
    ssl_key_passphrase           = <optional, computed, string>   # UseStateForUnknown; sensitive
  }

  icmp {                                           # optional; ICMP (ping) monitor fields
    host = <required, string>   # IP address or hostname
    wait = <optional, computed, int64>             # UseStateForUnknown; seconds; default 1
  }

  browser {                                        # optional; browser (Playwright) monitor fields
    inline_script       = <required, string>
    screenshots         = <optional, computed, string>   # UseStateForUnknown; one of: "on", "off", "only-on-failure"
    synthetics_args     = <optional, list(string)>       # CLI arguments passed to the synthetics agent
    ignore_https_errors = <optional, computed, bool>     # UseStateForUnknown
    playwright_options  = <optional, string>             # JSON (normalized) object
  }
}
```

Notes:

- The type name is built as `req.ProviderTypeName + "_kibana_synthetics_monitor"` (constant `synthetics.MetadataPrefix + "monitor"`).
- The `id` attribute has both `UseStateForUnknown` and `RequiresReplace` plan modifiers.
- `space_id` has `UseStateForUnknown` and `RequiresReplace`; changing it destroys and recreates.
- Managed location validation (enum of allowed location strings for `locations`) runs at validate time unless the environment variable `TF_ELASTICSTACK_SKIP_LOCATION_VALIDATION` is set to `true` at that moment.
- `params`, `proxy_header`, `response`, `check`, and `playwright_options` use `jsontypes.NormalizedType`.

## Requirements

### Requirement: Kibana synthetics monitor APIs (REQ-001)

The resource SHALL manage synthetics monitors through Kibana's Synthetics Monitor HTTP API: add monitor (create), get monitor (read), update monitor (update), and delete monitor (delete). Reference: [Kibana Synthetics Add Monitor API](https://www.elastic.co/guide/en/kibana/current/add-monitor-api.html).

#### Scenario: Create monitor

- GIVEN valid configuration with exactly one type block and at least one location
- WHEN create runs
- THEN the provider SHALL call the Kibana add-monitor API with the configured space and SHALL set state from the API response

#### Scenario: Read removes missing monitors

- GIVEN a read/refresh
- WHEN get returns HTTP 404 for the monitor
- THEN the provider SHALL remove the resource from state and SHALL NOT surface an error diagnostic

#### Scenario: Update monitor

- GIVEN a valid plan with changes to mutable fields
- WHEN update runs
- THEN the provider SHALL call the Kibana update-monitor API and SHALL set state from the response

#### Scenario: Delete monitor

- GIVEN a monitor in state
- WHEN delete runs
- THEN the provider SHALL call the Kibana delete-monitor API with the monitor ID and space from state

### Requirement: API error surfacing (REQ-002)

For create, update, and read, when the request fails at the transport layer or the API returns a non-success HTTP status, the resource SHALL surface error diagnostics to Terraform describing the failure. Delete SHALL surface errors when the API call fails regardless of cause.

#### Scenario: Non-success create/update

- GIVEN a non-success API response on create or update
- WHEN the operation completes
- THEN Terraform SHALL receive error diagnostics describing the failure

#### Scenario: Non-404 error on read

- GIVEN a non-success response that is not HTTP 404 on read
- WHEN the operation completes
- THEN Terraform SHALL receive error diagnostics describing the failure

### Requirement: Provider configuration and Kibana client (REQ-003)

On create, read, update, and delete, if the provider did not supply a usable API client, the resource SHALL return an "Unconfigured Client" error diagnostic and SHALL NOT proceed to the Kibana API.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an "Unconfigured Client" error diagnostic

### Requirement: Identity and composite `id` (REQ-004)

After a successful API call, the resource SHALL set `id` to the composite string `<space_id>/<monitor_id>`, where `space_id` is the Kibana space used for the operation and `monitor_id` is the monitor ID returned by the API.

#### Scenario: State matches composite id

- GIVEN a monitor returned by the API with monitor ID `abc123` and space `my-space`
- WHEN state is written
- THEN `id` SHALL equal `"my-space/abc123"`

### Requirement: Import (REQ-005)

The resource SHALL support Terraform import using an `id` in the form `<space_id>/<monitor_id>` (exactly one `/` separating two non-empty segments). The imported `id` SHALL be passed through to state as-is. If the `id` is not in that form (i.e. does not split into exactly two parts on `/`), the provider SHALL return a "Wrong resource ID" error diagnostic.

#### Scenario: Valid import id

- GIVEN an import id `default/mon-001`
- WHEN import runs
- THEN state SHALL hold `id = "default/mon-001"` and subsequent read SHALL use `default` as the space and `mon-001` as the monitor ID

#### Scenario: Invalid import id format

- GIVEN an import id without a `/` separator (e.g. `badid`)
- WHEN import runs
- THEN the provider SHALL return an error diagnostic with summary "Wrong resource ID."

### Requirement: Lifecycle — force replacement (REQ-006)

Changing `id` or `space_id` SHALL require destroying and recreating the resource rather than an in-place update.

#### Scenario: Replace on space_id change

- GIVEN a plan that changes only `space_id`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create) for the resource

### Requirement: Config validation — exactly one monitor type (REQ-007)

Exactly one of the `http`, `tcp`, `icmp`, or `browser` blocks SHALL be set. Configuring zero or more than one of these blocks SHALL be rejected at plan/validate time.

#### Scenario: No type block

- GIVEN a configuration with no `http`, `tcp`, `icmp`, or `browser` block
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

#### Scenario: Multiple type blocks

- GIVEN a configuration with both `http` and `tcp` blocks
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Config validation — at least one location (REQ-008)

At least one of `locations` or `private_locations` SHALL be set. Configuring neither SHALL be rejected at plan/validate time.

#### Scenario: No location provided

- GIVEN a configuration with neither `locations` nor `private_locations` set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Config validation — `locations` enum (REQ-009)

Unless `TF_ELASTICSTACK_SKIP_LOCATION_VALIDATION` is set to `true` in the environment at validate time, each element of `locations` SHALL be one of the recognized managed location identifiers. An unrecognized location string SHALL be rejected with a validation diagnostic.

#### Scenario: Invalid location string with validation enabled

- GIVEN `locations` contains a string not in the recognized set and `TF_ELASTICSTACK_SKIP_LOCATION_VALIDATION` is not `true`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic listing the allowed values

### Requirement: Config validation — `schedule` enum (REQ-010)

The `schedule` attribute, when set to a known value, SHALL be one of `1`, `3`, `5`, `10`, `15`, `30`, `60`, `120`, or `240` (minutes). Any other value SHALL be rejected at plan/validate time.

#### Scenario: Invalid schedule value

- GIVEN `schedule` set to `7` (not in the allowed set)
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Config validation — `namespace` format (REQ-011)

The `namespace` attribute, when set to a known value, SHALL match the pattern `^[^*\\/?\"<>|\s,#:-]*$`. Values containing any disallowed character SHALL be rejected at plan/validate time.

#### Scenario: Invalid namespace character

- GIVEN `namespace` set to `"my namespace"` (contains a space)
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Config validation — `browser.screenshots` enum (REQ-012)

The `browser.screenshots` attribute, when set to a known value, SHALL be one of `"on"`, `"off"`, or `"only-on-failure"`. Any other value SHALL be rejected at plan/validate time.

#### Scenario: Invalid screenshots value

- GIVEN `browser.screenshots` set to `"always"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Config validation — `http.mode` enum (REQ-013)

The `http.mode` attribute, when set to a known value, SHALL be one of `"any"` or `"all"`. Any other value SHALL be rejected at plan/validate time.

#### Scenario: Invalid mode value

- GIVEN `http.mode` set to `"both"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic

### Requirement: Compatibility — `labels` attribute (REQ-014)

When `labels` is configured with a known, non-null value, the resource SHALL verify the Elastic stack server version is at least **8.16.0** before creating or updating the monitor. If the server version is strictly below **8.16.0**, the resource SHALL fail with an "Unsupported version for `labels` attribute" error diagnostic naming the minimum required version and SHALL NOT call the Kibana API.

#### Scenario: labels on old stack

- GIVEN server version &lt; 8.16.0 and `labels` configured with a non-null value
- WHEN create or update runs
- THEN the provider SHALL return an "Unsupported version for `labels` attribute" error diagnostic

### Requirement: Read — map API response to state (REQ-015)

When reading a monitor from the API, the resource SHALL map every field from the API response to Terraform state according to the monitor's type. The `id` SHALL be set to `<space_id>/<api_monitor_id>`. Fields for the monitor type not returned (absent or empty from API) SHALL preserve the prior state value where the schema marks them as `UseStateForUnknown`. The `locations` list SHALL be preserved from prior state (not re-derived from the API response); the `private_locations` list SHALL be derived from the API response locations where `is_service_managed` is false, using the label of each such location.

#### Scenario: HTTP monitor read

- GIVEN the API returns a monitor of type `http`
- WHEN the provider maps the response to state
- THEN `http.url`, `http.mode`, and other HTTP fields SHALL be set from the API response; `tcp`, `icmp`, and `browser` SHALL be null in state

#### Scenario: Private location mapping

- GIVEN the API returns locations where some have `is_service_managed = false`
- WHEN the provider maps the response to state
- THEN `private_locations` SHALL contain the labels of those non-service-managed locations

### Requirement: Mapping — `params` JSON object (REQ-016)

When the API returns a non-null `params` object for the monitor, the resource SHALL serialize it to a normalized JSON string and store it in state. When the API returns null `params`, the prior state value for `params` SHALL be preserved.

#### Scenario: API returns params

- GIVEN the API response includes a non-null `params` object
- WHEN the provider maps the response to state
- THEN `params` in state SHALL be a normalized JSON string representation of the API params object

#### Scenario: API returns null params

- GIVEN the API response has null `params`
- WHEN the provider maps the response to state
- THEN `params` in state SHALL retain the prior state value

### Requirement: Mapping — `proxy_header` and `playwright_options` JSON objects (REQ-017)

When the API returns a non-null `proxy_headers` value for an HTTP monitor, the resource SHALL serialize it to a normalized JSON string and store it in `http.proxy_header`. When the API returns null or absent `proxy_headers`, the prior state value SHALL be preserved. The same preservation rule SHALL apply to `browser.playwright_options`.

#### Scenario: API returns null proxy_header

- GIVEN the API response has null or absent `proxy_headers`
- WHEN the provider maps the response to state
- THEN `http.proxy_header` SHALL retain the prior state value

### Requirement: Mapping — `http.check` and `http.response` preservation (REQ-018)

The `http.check` and `http.response` JSON object attributes SHALL not be overwritten from the API response; they SHALL always retain the prior state values. This ensures no spurious drift from Kibana not echoing these values back.

#### Scenario: Read does not overwrite check/response

- GIVEN prior state has `http.check` set to a non-null JSON string
- WHEN the provider reads the monitor and maps the API response to state
- THEN `http.check` SHALL retain the prior state value regardless of what the API returns

### Requirement: Mapping — integer fields from string API values (REQ-019)

Several API fields are returned as strings but stored in state as `int64`. The resource SHALL parse `schedule.number`, `timeout`, and `icmp.wait` as base-10 integers. If parsing fails, the resource SHALL surface a diagnostic describing the conversion failure and SHALL NOT store partial state.

#### Scenario: Invalid schedule string from API

- GIVEN the API returns a `schedule.number` that cannot be parsed as an integer
- WHEN the provider maps the response to state
- THEN the provider SHALL return an error diagnostic and SHALL NOT set state

### Requirement: Mapping — SSL fields (REQ-020)

For both HTTP and TCP monitors, the SSL configuration fields (`ssl_verification_mode`, `ssl_supported_protocols`, `ssl_certificate_authorities`, `ssl_certificate`, `ssl_key`, `ssl_key_passphrase`) SHALL be mapped from the API response to state. When sending to the API, SSL fields SHALL be sent only when they are known and non-null; `ssl_supported_protocols` and `ssl_verification_mode` each drive an `SSLConfig` object independently.

#### Scenario: SSL fields from API

- GIVEN the API returns non-empty `sslSupportedProtocols` and `sslVerificationMode`
- WHEN the provider maps the response to state
- THEN `ssl_supported_protocols` and `ssl_verification_mode` SHALL be set accordingly in state

### Requirement: Mapping — alert configuration (REQ-021)

When the API returns a non-null `alert` object, the resource SHALL map it to the `alert` block in state with nested `status.enabled` and `tls.enabled` sub-objects. When the API returns null `alert`, `alert` in state SHALL be null.

#### Scenario: Null alert from API

- GIVEN the API returns null for `alert`
- WHEN the provider maps the response to state
- THEN `alert` in state SHALL be null

### Requirement: Write path — type-specific API request fields (REQ-022)

On create and update, the provider SHALL build the API request from the one configured type block. The HTTP type SHALL send `url`, `ssl` config, `max_redirects`, `mode`, `ipv4`, `ipv6`, `username`, `password`, `proxy_header`, `proxy_url`, `response`, and `check`. The TCP type SHALL send `host`, `check_send`, `check_receive`, `proxy_url`, `proxy_use_local_resolver`, and `ssl` config. The ICMP type SHALL send `host` and `wait` (as a string). The browser type SHALL send `inline_script`, `screenshots`, `synthetics_args`, `ignore_https_errors`, and `playwright_options`. If the configured type block is absent (none of the four), the provider SHALL fail with an "Unsupported monitor type config" error diagnostic.

#### Scenario: HTTP fields sent to API

- GIVEN a configuration with an `http` block
- WHEN create or update runs
- THEN the provider SHALL build an `HTTPMonitorFields` request with the configured `url` and optional HTTP-specific fields

### Requirement: Write path — common config fields (REQ-023)

On create and update, the provider SHALL send common `SyntheticsMonitorConfig` fields: `name`, `schedule`, `locations` (as managed location strings), `private_locations` (as label strings), `enabled`, `tags`, `labels`, `alert`, `service_name` (as `apm_service_name`), `timeout`, `namespace`, `params`, and `retest_on_failure`. When `labels` is null or unknown, the provider SHALL send an empty map rather than null.

#### Scenario: Common fields in create request

- GIVEN a monitor config with `name`, `schedule`, `locations`, and `tags`
- WHEN create runs
- THEN the API request SHALL include `name`, `schedule`, `locations`, and `tags` matching the plan values

## Traceability (implementation index)

| Area | Primary files |
|------|---------------|
| Metadata, Configure, Import, ConfigValidators | `resource.go` |
| Schema, model types, mapping functions, version constraints | `schema.go` |
| Create | `create.go` |
| Read | `read.go` |
| Update | `update.go` |
| Delete | `delete.go` |
| Shared synthetics utilities (GetKibanaClient, GetCompositeID) | `internal/kibana/synthetics/api_client.go`, `internal/kibana/synthetics/schema.go` |
| Composite ID parsing | `internal/clients/api_client.go` (CompositeIDFromStr) |
