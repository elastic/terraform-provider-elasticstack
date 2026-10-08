# Delta Spec: `proxy_id` support for `elasticstack_fleet_server_host`

Base spec: `openspec/specs/fleet-server-host/spec.md`

## Schema refinements

Adds an optional `proxy_id` attribute, mirroring the existing `elasticstack_fleet_agent_download_source.proxy_id`:

```hcl
resource "elasticstack_fleet_server_host" "example" {
  id      = <computed, string>          # internal identifier, mirrors host_id
  host_id = <optional+computed, string> # Fleet-assigned or user-supplied host ID

  name      = <required, string>
  hosts     = <required, list(string)>    # at least one entry
  default   = <optional+computed, bool>  # defaults to false when omitted
  proxy_id  = <optional, string>          # references a Fleet proxy by ID
  space_ids = <optional+computed, set(string)>
}
```

## MODIFIED Requirements

### Requirement: State mapping (REQ-012)

On read, the resource SHALL map `id`, `host_id`, `name`, `hosts`, `default`, and `proxy_id` from the API response. The `default` attribute SHALL always carry a known boolean value in state — when the user omits it from configuration, it SHALL default to `false` so plan and post-apply state agree. `proxy_id` SHALL be mapped from the API response's `proxy_id` field as-is: present and non-empty when the server host has a proxy assigned, and null when it does not. `space_ids` is not returned by the Fleet API; if `space_ids` is unknown in state after the API call, the resource SHALL set it to explicit null. If `space_ids` has a configured value, the resource SHALL preserve it.

#### Scenario: default omitted from config

- GIVEN config does not set `default`
- WHEN apply completes
- THEN state SHALL have `default = false` (matching the Fleet API response) with no inconsistent-result error

#### Scenario: space_ids preserved after read

- GIVEN `space_ids = ["my-space"]` in state
- WHEN read runs
- THEN `space_ids` SHALL remain `["my-space"]` after the refresh

#### Scenario: proxy_id round-trips through state

- GIVEN a server host was created with `proxy_id = "my-proxy"`
- WHEN read runs
- THEN state SHALL have `proxy_id = "my-proxy"`

#### Scenario: proxy_id absent from API response

- GIVEN the server host has no proxy assigned
- WHEN read runs
- THEN `proxy_id` SHALL be null in state

### Requirement: Create API body (REQ-013)

On create, the resource SHALL submit `host_urls` (from `hosts`), `name`, `is_default` (from `default`), optionally `id` (from `host_id` when configured), and optionally `proxy_id` (from `proxy_id` when configured) in the create request body.

#### Scenario: Create with all fields

- GIVEN name, hosts, and default are all configured
- WHEN create runs
- THEN the Fleet API SHALL receive host_urls, name, and is_default in the request body

#### Scenario: Create with proxy_id

- GIVEN `proxy_id = "my-proxy"` is configured
- WHEN create runs
- THEN the Fleet create API request body SHALL include `"proxy_id": "my-proxy"`

#### Scenario: Create without proxy_id

- GIVEN `proxy_id` is not configured
- WHEN create runs
- THEN the Fleet create API request body SHALL NOT include a `proxy_id` field (or SHALL include it as absent/null, consistent with the generated client's `omitempty` behavior)

### Requirement: Update API body (REQ-014)

On update, the resource SHALL submit `host_urls`, `name`, and `is_default`, and SHALL conditionally submit `proxy_id` according to the rules below, using `host_id` from the plan as the resource identifier.

The `proxy_id` value sent on update SHALL be computed from both the plan value and the prior state value, per the following rule (see REQ-017 for the full rationale):

1. When the plan's `proxy_id` is known and non-empty, the resource SHALL send that value.
2. When the plan's `proxy_id` is null, unknown, or empty, AND the prior state's `proxy_id` was known and non-empty (i.e. the practitioner is clearing a previously-set `proxy_id`), the resource SHALL send an explicit empty string `""`.
3. When the plan's `proxy_id` is null, unknown, or empty, AND the prior state's `proxy_id` was also null, unknown, or empty (i.e. `proxy_id` was never set), the resource SHALL send `nil` (omit the field from the request body).

#### Scenario: Update name

- GIVEN a server host with a new `name` in plan
- WHEN update runs
- THEN the Fleet update API SHALL be called with the new name

#### Scenario: Update sets proxy_id

- GIVEN a server host with no prior `proxy_id` in state
- AND the plan sets `proxy_id = "my-proxy"`
- WHEN update runs
- THEN the Fleet update API request body SHALL include `"proxy_id": "my-proxy"`

#### Scenario: Update changes proxy_id to a different value

- GIVEN a server host with `proxy_id = "proxy-a"` in prior state
- AND the plan sets `proxy_id = "proxy-b"`
- WHEN update runs
- THEN the Fleet update API request body SHALL include `"proxy_id": "proxy-b"`

#### Scenario: Update clears a previously-set proxy_id

- GIVEN a server host with `proxy_id = "my-proxy"` in prior state
- AND the plan omits `proxy_id` (or sets it to an empty value)
- WHEN update runs
- THEN the Fleet update API request body SHALL include `"proxy_id": ""`
- AND the resource SHALL NOT omit the `proxy_id` field from the request body

#### Scenario: Update with proxy_id never set

- GIVEN a server host with no `proxy_id` in prior state
- AND the plan does not set `proxy_id`
- WHEN update runs
- THEN the Fleet update API request body SHALL NOT include a `proxy_id` field

## ADDED Requirements

### Requirement: proxy_id unset-to-empty-string update semantics (REQ-017)

Because the generated Fleet client tags the update request body's `proxy_id` field as `json:"proxy_id,omitempty"`, a `nil` pointer value is dropped from the serialized request body entirely, causing the field to be omitted. Omitting `proxy_id` on `PUT /fleet/fleet_server_hosts/{itemId}` SHALL be treated by the resource as equivalent to Fleet's own "leave unchanged" semantics for omitted fields — the same behavior already documented and relied upon for `PUT /fleet/agent_download_sources/{sourceId}` on the sibling `elasticstack_fleet_agent_download_source` resource.

Consequently, the resource SHALL distinguish "practitioner never configured `proxy_id`" (send nothing — field omitted, prior value if any is left unchanged by Fleet) from "practitioner explicitly cleared a previously-set `proxy_id`" (send an explicit empty string `""`, which Fleet SHALL interpret as clearing the proxy assignment) on every update request. The resource SHALL make this determination by comparing the plan's `proxy_id` against the prior state's `proxy_id`, not from the plan value alone.

#### Scenario: Clearing proxy_id produces an empty string, not an omitted field

- GIVEN a server host with `proxy_id = "my-proxy"` in prior state
- WHEN the practitioner removes the `proxy_id` line from configuration and applies
- THEN the resource SHALL send `"proxy_id": ""` in the update request body
- AND state after apply SHALL have `proxy_id` null
- AND the apply SHALL NOT produce a Terraform "inconsistent result after apply" error

#### Scenario: A configuration that never set proxy_id does not send the field on unrelated updates

- GIVEN a server host with no `proxy_id` ever configured
- WHEN the practitioner changes `name` and applies
- THEN the resource SHALL NOT send a `proxy_id` field in the update request body
