# `elasticstack_fleet_proxy` — Schema and Functional Requirements

Resource implementation: `internal/fleet/proxy`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_fleet_proxy` resource, including Fleet Proxies API usage, identity/import, connection handling, state mapping, and proxy header lifecycle.

## Schema

```hcl
resource "elasticstack_fleet_proxy" "example" {
  # Identity
  id       = <computed, string>  # "<space_id>/<proxy_id>"; UseStateForUnknown
  proxy_id = <optional, computed, string>  # server-assigned if omitted; RequiresReplace; UseStateForUnknown
  space_id = <optional, computed, string>  # default: "default"; RequiresReplace; UseStateForUnknown

  # Required
  name = <required, string>  # LengthAtLeast(1)
  url  = <required, string>

  # Optional TLS
  certificate             = <optional, string>
  certificate_authorities = <optional, string>
  certificate_key         = <optional, string, sensitive>

  # Optional headers
  proxy_headers = <optional, map(string)>

  # Computed
  is_preconfigured = <computed, bool>

  # Connection override
  kibana_connection {
    endpoints    = <optional, list(string), sensitive>
    username     = <optional, string>
    password     = <optional, string, sensitive>
    api_key      = <optional, string, sensitive>
    bearer_token = <optional, string, sensitive>
    ca_certs     = <optional, list(string)>
    insecure     = <optional, bool>
  }
}
```

## Requirements

### Requirement: Identity and composite ID

The resource SHALL set `id` as the composite string `"<space_id>/<proxy_id>"` after every create and update. `proxy_id` SHALL be preserved from state (`UseStateForUnknown`) and populated from the API response. `space_id` SHALL default to `"default"` when not specified.

#### Scenario: Create with auto-assigned proxy_id
- **WHEN** `proxy_id` is not set in config and the resource is created
- **THEN** `proxy_id` SHALL be populated from the API-assigned ID
- **AND** `id` SHALL equal `"default/<proxy_id>"`

#### Scenario: Create with explicit proxy_id
- **WHEN** `proxy_id = "my-proxy"` is set in config and the resource is created
- **THEN** the API SHALL be called with `id: "my-proxy"`
- **AND** `proxy_id` in state SHALL equal `"my-proxy"`

### Requirement: Replacement on identity change

Changing `proxy_id` or `space_id` SHALL trigger resource replacement (`RequiresReplace`).

#### Scenario: proxy_id change forces replacement
- **WHEN** `proxy_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

#### Scenario: space_id change forces replacement
- **WHEN** `space_id` is changed in config
- **THEN** Terraform SHALL destroy and recreate the resource

### Requirement: Create

The resource SHALL call `POST /api/fleet/proxies` (space-aware) with `name`, `url`, and any optional TLS and header fields. The response body SHALL be decoded into the generated `kbapi.FleetProxyItem` type. State SHALL be set from the decoded response.

#### Scenario: Successful create
- **WHEN** a proxy resource is applied for the first time
- **THEN** `POST /api/fleet/proxies` SHALL be called with the configured fields
- **AND** state SHALL reflect the values returned in the API response

### Requirement: Read

The resource SHALL call `GET /api/fleet/proxies/{id}` (space-aware). On HTTP 404 the resource SHALL be removed from state. The response body SHALL be decoded into the generated `kbapi.FleetProxyItem` type.

#### Scenario: Resource deleted out of band
- **WHEN** the API returns HTTP 404 on Read
- **THEN** the resource SHALL be removed from state without error

### Requirement: Update

The resource SHALL call `PUT /api/fleet/proxies/{id}` (space-aware). `name` and `url` SHALL always be included. `proxy_headers` SHALL always be included in the request body — as `{}` when null or empty — so that existing headers are cleared when removed from config.

#### Scenario: Clearing proxy_headers
- **WHEN** `proxy_headers` is removed from config on an existing proxy that had headers
- **THEN** the PUT body SHALL include `"proxy_headers": {}`
- **AND** state after apply SHALL have `proxy_headers` as null

#### Scenario: Updating name and url
- **WHEN** `name` or `url` is changed in config
- **THEN** `PUT /api/fleet/proxies/{id}` SHALL be called with the new values
- **AND** state SHALL reflect the updated values

### Requirement: Delete

The resource SHALL call `DELETE /api/fleet/proxies/{id}` (space-aware). HTTP 404 on delete SHALL be treated as success.

#### Scenario: Successful delete
- **WHEN** the resource is destroyed
- **THEN** `DELETE /api/fleet/proxies/{id}` SHALL be called
- **AND** no error SHALL be returned

#### Scenario: Already-deleted resource
- **WHEN** the API returns HTTP 404 on Delete
- **THEN** the resource SHALL be removed from state without error

### Requirement: Import

The resource SHALL support import via the composite ID `"<space_id>/<proxy_id>"`. On import, Read SHALL parse the composite ID to derive `space_id` and `proxy_id` for the API call.

#### Scenario: Import by composite ID
- **WHEN** `terraform import elasticstack_fleet_proxy.x "default/my-proxy-id"` is run
- **THEN** `proxy_id` SHALL be set to `"my-proxy-id"`
- **AND** `space_id` SHALL be set to `"default"`
- **AND** all other attributes SHALL be populated from the API response

### Requirement: Proxy headers encoding

`proxy_headers` SHALL be modelled as `map(string)` in the schema. On write, each string value SHALL be sent through the generated `kbapi.FleetProxyHeaderValue` union via its `FromFleetProxyHeaderValueString` helper. On read, each header value SHALL be decoded from `kbapi.FleetProxyHeaderValue`; string values SHALL be stored verbatim, while boolean and numeric values (which the Fleet API also accepts) SHALL be stringified into state so that the `map(string)` schema can represent them.

#### Scenario: String header round-trip
- **WHEN** `proxy_headers = { "X-My-Header" = "value" }` is set
- **THEN** the API SHALL receive `"proxy_headers": { "X-My-Header": "value" }`
- **AND** state SHALL contain `proxy_headers.X-My-Header = "value"` after apply

#### Scenario: Non-string header from API
- **WHEN** the API returns a boolean or numeric value for a proxy header key
- **THEN** that value SHALL be stringified (e.g. `true` → `"true"`, `42` → `"42"`)
- **AND** the resulting string SHALL be stored under that key in `proxy_headers`

### Requirement: TLS fields sensitivity

`certificate_key` SHALL be marked sensitive. Empty strings returned by the API for TLS fields SHALL be treated as null in state.

#### Scenario: certificate_key not shown in plan
- **WHEN** `certificate_key` is set
- **THEN** its value SHALL be redacted in Terraform plan and apply output

#### Scenario: Empty TLS field from API
- **WHEN** the API returns an empty string for `certificate`, `certificate_authorities`, or `certificate_key`
- **THEN** the corresponding attribute SHALL be null in state

### Requirement: is_preconfigured default

`is_preconfigured` is a computed boolean. It SHALL be set to `false` when the API returns null or omits the field.

#### Scenario: is_preconfigured not set by API
- **WHEN** the API response omits `is_preconfigured`
- **THEN** `is_preconfigured` in state SHALL be `false`

### Requirement: Connection

The resource SHALL obtain its Fleet client via `kibana_connection` if provided, otherwise via the provider-level Kibana configuration. Space-aware requests SHALL use `space_id` via the `spaceAwarePathRequestEditor` helper.

#### Scenario: Resource-level connection override
- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all API calls SHALL use that connection instead of the provider-level Kibana connection
