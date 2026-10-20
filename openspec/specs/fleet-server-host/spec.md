# `elasticstack_fleet_server_host` — Schema and Functional Requirements

Resource implementation: `internal/fleet/serverhost`

## Purpose

Define schema and behavior for the Fleet server host resource. The resource manages Fleet server hosts via the Fleet Fleet Server Hosts API, including Kibana space awareness.

## Schema

```hcl
resource "elasticstack_fleet_server_host" "example" {
  id      = <computed, string>          # internal identifier, mirrors host_id
  host_id = <optional+computed, string> # Fleet-assigned or user-supplied host ID

  name    = <required, string>
  hosts   = <required, list(string)>    # at least one entry
  default = <optional+computed, bool>  # defaults to false when omitted
  proxy_id = <optional, string>         # references a Fleet proxy by ID; empty string rejected at validation
  space_ids = <optional+computed, set(string)>
}
```

## Requirements

### Requirement: Fleet Server Host CRUD APIs (REQ-001–REQ-004)

The resource SHALL use the Fleet Create fleet server host API to create server hosts. The resource SHALL use the Fleet Get fleet server host API to read server hosts. The resource SHALL use the Fleet Update fleet server host API to update server hosts. The resource SHALL use the Fleet Delete fleet server host API to delete server hosts. When the Fleet API returns a non-success status for any create, update, or delete operation, the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API error on create

- GIVEN the Fleet API returns an error on create
- WHEN the resource create runs
- THEN diagnostics SHALL contain the API error

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` attribute that mirrors `host_id`. On create, if `host_id` is not configured, the Fleet API SHALL assign the identifier; the resource SHALL store the API-assigned value in both `id` and `host_id`. When `host_id` is configured, the resource SHALL pass it to the create API as the desired identifier.

#### Scenario: Auto-assigned host_id

- GIVEN `host_id` is not set in config
- WHEN create completes successfully
- THEN `host_id` and `id` SHALL both be set to the API-returned identifier

### Requirement: Create behavior — omit id when unset

When `host_id` is not set in config (null or unknown), the resource SHALL omit the `id` field from the Fleet Create Fleet Server Host request body entirely, allowing Fleet to auto-generate an ID. The resource SHALL NOT send `"id": ""` to the API.

An explicit `host_id = ""` in config is rejected at plan time by the `host_id` validation requirement and therefore never reaches the create payload.

#### Scenario: host_id unset — id omitted from create body

- GIVEN `host_id` is not set in config
- WHEN create runs and the Fleet Create Fleet Server Host API is called
- THEN the request body SHALL NOT contain an `"id"` field
- AND Fleet SHALL auto-assign a host ID which is stored in `host_id` state

#### Scenario: host_id set — id sent in create body

- GIVEN `host_id = "my-host-id"` is set in config
- WHEN create runs and the Fleet Create Fleet Server Host API is called
- THEN the request body SHALL contain `"id": "my-host-id"`

### Requirement: host_id plan-time validation

The resource SHALL validate `host_id` at plan time when an explicit value is provided. A supplied `host_id` value SHALL satisfy all of the following constraints:

1. Length is between 1 and 255 characters (inclusive).
2. Does not contain `/` (path separator).
3. Does not contain `..` (traversal sequence).
4. Does not contain any of the reserved keys as a substring: `__proto__`, `constructor`, `prototype`.

When any constraint is violated, the resource SHALL surface a plan-time error diagnostic naming the violated constraint. The validator SHALL NOT produce an error for null or unknown values; those are treated as not set and the `id` field is omitted from the create payload. An explicit empty string (`host_id = ""`) SHALL be rejected as a length-0 violation.

#### Scenario: Valid explicit host_id passes validation

- GIVEN `host_id = "my-valid-host"` is set in config
- WHEN a plan is generated
- THEN no validation errors SHALL be produced

#### Scenario: Explicit empty host_id fails validation

- GIVEN `host_id = ""` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the length constraint

#### Scenario: Unset host_id passes validation

- GIVEN `host_id` is not set in config (null or unknown)
- WHEN a plan is generated
- THEN the plan-time validator SHALL NOT produce an error
- AND the `id` field SHALL be omitted from the create request body

#### Scenario: host_id with path separator fails validation

- GIVEN `host_id = "my/host"` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the `/` constraint

#### Scenario: host_id with traversal sequence fails validation

- GIVEN `host_id = "my..host"` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the `..` constraint

#### Scenario: host_id exceeds maximum length fails validation

- GIVEN `host_id` is set to a string of 256 or more characters
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the length constraint

#### Scenario: Reserved key host_id fails validation

- GIVEN `host_id = "__proto__"` (or `"constructor"` or `"prototype"`) is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the reserved-key constraint

#### Scenario: Reserved key as substring fails validation

- GIVEN `host_id = "my-__proto__-host"` (or any value containing `constructor` or `prototype` as a substring) is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the reserved-key constraint

### Requirement: Import (REQ-007)

The resource SHALL support import with both plain and composite import IDs.

When the import ID is a composite string in the format `<space_id>/<host_id>`, the resource SHALL set `host_id` to the parsed resource ID and `space_ids` to `[<space_id>]` in state, so that server hosts in non-default Kibana spaces can be imported successfully.

When the import ID is a plain (non-composite) string — i.e. it contains no `/` that can be parsed as a composite ID — the resource SHALL treat the entire string as `host_id` and SHALL NOT set `space_ids` from the import ID. This preserves existing behaviour for default-space imports.

On the subsequent read after import (regardless of ID form), the resource SHALL use the space from state to query the Fleet API and populate all remaining attributes.

#### Scenario: Import by composite space/host ID

- GIVEN an existing Fleet server host in Kibana space `"my-space"` with host ID `"abc-123"`
- WHEN `terraform import` is run with the composite ID `"my-space/abc-123"`
- THEN `host_id` SHALL be `"abc-123"` and `space_ids` SHALL contain `"my-space"`

#### Scenario: Import by plain host ID (default space)

- GIVEN an existing Fleet server host in the default Kibana space with host ID `"abc-123"`
- WHEN `terraform import` is run with the plain ID `"abc-123"` (no `/` separator)
- THEN `host_id` SHALL be `"abc-123"` and `space_ids` SHALL NOT be set from the import ID

### Requirement: Space-aware create (REQ-008)

On create, when `space_ids` is configured with at least one space ID, the resource SHALL pass the first space ID from `space_ids` to the Fleet create API as the space context. When `space_ids` is null or unknown, the resource SHALL call the create API without a space prefix (default space).

#### Scenario: Create in named space

- GIVEN `space_ids = ["my-space"]`
- WHEN create runs
- THEN the Fleet create API SHALL be called with `my-space` as the space context

### Requirement: Space-aware read and update using state (REQ-009)

On read and update, the resource SHALL derive the operational space from the `space_ids` stored in state (not the plan). If `space_ids` in state is null or empty, the resource SHALL query using the default space. Otherwise, the resource SHALL use the first space ID from state as the space context. On update, the resource SHALL send updated fields (including `space_ids` from the plan) in the request body so the Fleet API can adjust space membership.

#### Scenario: Read uses state space

- GIVEN state has `space_ids = ["space-a"]`
- WHEN read runs
- THEN the Fleet get API SHALL be called using `space-a` as the space context

### Requirement: Space-aware delete (REQ-010)

On delete, the resource SHALL use the first space ID from state as the space context (same logic as read). Deleting removes the server host from all spaces; to remove from specific spaces only, `space_ids` SHALL be updated instead.

#### Scenario: Delete uses state space

- GIVEN state has `space_ids = ["space-a"]`
- WHEN destroy runs
- THEN the Fleet delete API SHALL be called using `space-a` as the space context

### Requirement: Delete clears default before removal (REQ-010a)

The Fleet API rejects deletion of a server host that is currently marked as default. On delete, when state has `default = true`, the resource SHALL first issue an update to set `is_default = false` (using the same operational space as the delete) and then SHALL issue the delete request. If the pre-delete update returns an error, the resource SHALL surface it to diagnostics and SHALL NOT proceed with the delete.

#### Scenario: Destroying a default server host

- GIVEN state has `default = true`
- WHEN `terraform destroy` runs
- THEN the resource SHALL update the host with `is_default = false` before calling the Fleet delete API
- AND the delete SHALL succeed without manual intervention

### Requirement: Read — not found removes from state (REQ-011)

On read, if the Fleet API returns a nil response for the server host, the resource SHALL remove itself from state. When the Fleet API returns an error, the resource SHALL surface it to Terraform diagnostics.

#### Scenario: Server host deleted outside Terraform

- GIVEN the server host was manually deleted from Fleet
- WHEN read (refresh) runs
- THEN the resource SHALL be removed from state

### Requirement: State mapping (REQ-012)

On read, the resource SHALL map `id`, `host_id`, `name`, `hosts`, `default`, and `proxy_id` from the API response. The `default` attribute SHALL always carry a known boolean value in state — when the user omits it from configuration, it SHALL default to `false` so plan and post-apply state agree. `proxy_id` SHALL be mapped from the API response's `proxy_id` field as-is: present and non-empty when the server host has a proxy assigned, and null when it does not. An empty-string `proxy_id` in the API response SHALL also be mapped to null. `space_ids` is not returned by the Fleet API; if `space_ids` is unknown in state after the API call, the resource SHALL set it to explicit null. If `space_ids` has a configured value, the resource SHALL preserve it.

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
2. When the plan's `proxy_id` is null or unknown (an empty configured value is rejected at validation), AND the prior state's `proxy_id` was known and non-empty (i.e. the practitioner is clearing a previously-set `proxy_id`), the resource SHALL send an explicit empty string `""`.
3. When the plan's `proxy_id` is null or unknown, AND the prior state's `proxy_id` was also null, unknown, or empty (i.e. `proxy_id` was never set), the resource SHALL send `nil` (omit the field from the request body).

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
- AND the plan omits `proxy_id`
- WHEN update runs
- THEN the Fleet update API request body SHALL include `"proxy_id": ""`
- AND the resource SHALL NOT omit the `proxy_id` field from the request body

#### Scenario: Configured empty proxy_id is rejected

- GIVEN configuration sets `proxy_id = ""`
- WHEN the configuration is validated
- THEN validation SHALL fail because `proxy_id` must be at least 1 character long, and no API call SHALL be made

#### Scenario: Update with proxy_id never set

- GIVEN a server host with no `proxy_id` in prior state
- AND the plan does not set `proxy_id`
- WHEN update runs
- THEN the Fleet update API request body SHALL NOT include a `proxy_id` field

### Requirement: Provider-level Fleet client by default with optional scoped override

The `elasticstack_fleet_server_host` resource SHALL use the provider-configured Fleet client by default. When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Fleet client for its API calls.

#### Scenario: Provider client used by default

- GIVEN `kibana_connection` is not configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the provider-configured Fleet client

#### Scenario: Scoped Fleet connection

- GIVEN `kibana_connection` is configured on the resource
- WHEN an API call runs
- THEN the resource SHALL use the scoped Fleet client derived from that block

### Requirement: `host_id` plan modifiers (REQ-015)

The `host_id` attribute SHALL carry `UseStateForUnknown()` and `RequiresReplace()` plan modifiers.

`UseStateForUnknown()` ensures that when `host_id` is absent from the practitioner's configuration, the Plugin Framework carries the prior-state value (the Fleet-assigned UUID) into the update plan so that the `Update` handler receives the correct resource identifier. Without this modifier, `host_id` is `null` in the plan, causing the `Update` handler to construct a request URL with an empty path segment and Kibana to return 404.

`RequiresReplace()` ensures that if a practitioner explicitly provides a new `host_id` value that differs from the one in state, the resource is destroyed and recreated rather than attempting an in-place update with a mismatched ID.

Both modifiers align `fleet_server_host.host_id` with the identical pattern already applied to every other fleet resource ID attribute (`fleet_output.output_id`, `fleet_agent_policy.policy_id`, `fleet_proxy.proxy_id`, `fleet_agent_download_source.source_id`).

#### Scenario: Update succeeds when host_id is not configured

- GIVEN a `fleet_server_host` resource created with `host_id` omitted from config (computed by Fleet)
- WHEN the practitioner changes `name` or `hosts` and runs `terraform apply`
- THEN the `Update` handler SHALL receive the prior-state `host_id` value in the plan (not null)
- AND the Fleet update API SHALL be called with the correct resource identifier
- AND the apply SHALL succeed without a 404 error

#### Scenario: Explicit host_id change triggers replacement

- GIVEN a `fleet_server_host` resource with a known `host_id` in state
- WHEN the practitioner sets a different explicit `host_id` value in config
- THEN Terraform SHALL plan a destroy-and-recreate rather than an in-place update

### Requirement: Acceptance test UPDATE coverage for computed host_id (REQ-016)

The acceptance test `TestAccResourceFleetServerHost_computedID` SHALL include at least one UPDATE `resource.TestStep` that:

1. Changes a mutable attribute (`name` or `hosts`) while omitting `host_id` from the test config.
2. Asserts that the apply succeeds.
3. Asserts that `host_id` remains set to a non-empty value in state after the update.

This guards the `UseStateForUnknown()` fix against future regressions.

#### Scenario: Acceptance test UPDATE step passes with computed host_id

- GIVEN a `fleet_server_host` acceptance test that created a host without an explicit `host_id`
- WHEN an UPDATE step is applied that changes `name` or `hosts` and still omits `host_id`
- THEN the apply SHALL succeed and state SHALL contain a non-empty `host_id`

### Requirement: proxy_id unset-to-empty-string update semantics (REQ-017)

Because the generated Fleet client tags the update request body's `proxy_id` field as `json:"proxy_id,omitempty"`, a `nil` pointer value is dropped from the serialized request body entirely. Fleet's behavior for an omitted `proxy_id` on `PUT /fleet/fleet_server_hosts/{itemId}` was verified live against Kibana 9.5.5: omission clears the proxy assignment, and an explicit empty string `""` also clears it (the response echoes `""`, which the resource maps to null). CI acceptance runs of the `proxy_id` test passed from 8.15.5 through 9.6.0-SNAPSHOT; stack versions 8.7.1 through 8.15.4 have not been verified. The the sibling `agent_download_sources` endpoint is documented to leave omitted fields unchanged.

Consequently, to be robust across stack versions, the resource SHALL distinguish "practitioner never configured `proxy_id`" (send nothing — field omitted) from "practitioner explicitly cleared a previously-set `proxy_id`" (send an explicit empty string `""`, which Fleet SHALL interpret as clearing the proxy assignment) on every update request. The resource SHALL make this determination by comparing the plan's `proxy_id` against the prior state's `proxy_id`, not from the plan value alone.

The `proxy_id` attribute SHALL NOT introduce a resource-level minimum stack version beyond the resource's existing behavior (the resource has no production version requirement; the 8.6.0 floor exists only in acceptance tests). Support was verified in CI from 8.15.5 through 9.6.0-SNAPSHOT; 8.7.1 through 8.15.4 is unverified, and support below 8.7.1 is not expected. The Fleet proxy resource (required to obtain a real `proxy_id`) requires 8.7.1, so the acceptance test is gated at 8.7.1.

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

### Requirement: Optional ssl attribute

The resource SHALL expose an optional `ssl` single-nested-object attribute with sub-attributes `certificate` (string, optional), `key` (string, optional, sensitive), `certificate_authorities` (list of string, optional, 1-10 non-empty entries), and `client_auth` (string, optional, one of `"none"`, `"optional"`, `"required"`). The provider SHALL map `certificate`, `key`, `certificate_authorities`, and `client_auth` to and from the `ssl` object on the Fleet create, update, and read response bodies.

#### Scenario: Create sends ssl fields

- **WHEN** the practitioner sets `ssl.certificate`, `ssl.key`, `ssl.certificate_authorities`, and/or `ssl.client_auth` on a new resource
- **THEN** the provider SHALL include the corresponding fields in the create request body's `ssl` object
- **AND** any `ssl` sub-attribute left unset SHALL be omitted from the request body's `ssl` object rather than sent as an empty or zero value

#### Scenario: Update sends changed ssl fields

- **WHEN** the practitioner changes any `ssl` sub-attribute on an existing resource and applies
- **THEN** the provider SHALL include the corresponding fields in the update request body's `ssl` object

#### Scenario: Read hydrates certificate, certificate_authorities, and client_auth from the API response

- **WHEN** the provider reads an existing server host whose stored `ssl` object has a non-null `certificate`, `certificate_authorities`, and/or `client_auth`
- **THEN** state SHALL contain a non-null `ssl` object with those fields populated from the response

#### Scenario: Omitted ssl is null in state

- **WHEN** the practitioner does not set `ssl` and the API response's `ssl` object is absent or has no non-null fields
- **THEN** state SHALL contain a null `ssl` object

#### Scenario: key is marked sensitive

- **WHEN** the practitioner inspects the `ssl.key` attribute in the resource schema
- **THEN** it SHALL be marked `Sensitive`

#### Scenario: Invalid client_auth value rejected

- **WHEN** the practitioner configures `ssl.client_auth` to a value other than `"none"`, `"optional"`, or `"required"`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: client_auth omitted is not defaulted by the provider

- **WHEN** the practitioner configures `ssl` without setting `client_auth`
- **THEN** the provider SHALL omit `client_auth` from the request body's `ssl` object rather than sending a provider-chosen default
- **AND** an apply followed by refresh SHALL complete without an inconsistent-result error or a subsequent diff while configuration continues to omit `client_auth`

#### Scenario: client_auth is Optional-only because Kibana returns no default

- **GIVEN** Kibana returns no `client_auth` when it is omitted on create
- **WHEN** the practitioner applies configuration that omits `client_auth` and then refreshes
- **THEN** state SHALL contain a null `ssl.client_auth`
- **AND** the schema SHALL NOT mark `client_auth` as computed

### Requirement: No provider-enforced ssl default or pairing constraint

The provider SHALL NOT enforce a default value for `client_auth`, nor a "certificate and key must be set together" constraint, since the Kibana API declares neither at the schema level.

#### Scenario: certificate without key is accepted

- **WHEN** the practitioner configures `ssl.certificate` without `ssl.key`
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a validation error

#### Scenario: key without certificate is accepted

- **WHEN** the practitioner configures `ssl.key` without `ssl.certificate`
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a validation error

### Requirement: ssl empty-value validation

The resource SHALL reject, at plan time, any configuration where `ssl` is set but ambiguously empty (all sub-attributes null), rather than silently normalizing the ambiguity to null on a later read.

#### Scenario: Empty ssl object rejected

- **WHEN** the practitioner configures `ssl = {}`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Empty string certificate or key rejected

- **WHEN** the practitioner configures `ssl.certificate = ""` or `ssl.key = ""`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Empty certificate_authorities list rejected

- **WHEN** the practitioner configures `ssl.certificate_authorities = []`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Empty-string CA element rejected

- **WHEN** the practitioner configures `ssl.certificate_authorities` containing one or more empty-string elements
- **THEN** Terraform SHALL report a validation error before any API call

### Requirement: ssl certificate_authorities size limit

The `certificate_authorities` sub-attribute SHALL accept at most 10 entries, matching the Fleet API's `maxItems: 10` constraint.

#### Scenario: Exactly 10 CAs accepted

- **WHEN** the practitioner configures `ssl.certificate_authorities` with exactly 10 non-empty entries
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a size validation error

#### Scenario: 11 CAs rejected

- **WHEN** the practitioner configures `ssl.certificate_authorities` with 11 non-empty entries
- **THEN** Terraform SHALL report a validation error before any API call

### Requirement: ssl minimum Kibana version guard

The `ssl` attribute SHALL be guarded by its own attribute-scoped minimum Kibana version requirement of 9.1.0, enforced only when `ssl` is configured, independent of the resource's other attributes. Configurations that do not set `ssl` SHALL NOT be affected by this requirement, on any stack version. `serverHostModel` SHALL implement `GetVersionRequirements` to enforce this.

#### Scenario: ssl create or update against unsupported stack version

- **WHEN** the practitioner configures `ssl` with at least one sub-attribute set
- **AND** the connected Kibana is below version 9.1.0
- **AND** the practitioner applies a create or update
- **THEN** the provider SHALL emit an attribute-scoped diagnostic on `ssl` that it is not supported on this version before sending a Fleet mutation request

#### Scenario: ssl omitted against old stack version is unaffected

- **WHEN** the practitioner does not configure `ssl`
- **AND** the connected Kibana is below version 9.1.0
- **THEN** the provider SHALL NOT emit a version diagnostic related to `ssl`
- **AND** the apply SHALL proceed as it did before this change

### Requirement: Upgrade migration for pre-existing server-side ssl

The provider SHALL document that users upgrading with an existing server-side `ssl` object must add its settings to Terraform configuration as part of the upgrade. The provider's read hydration SHALL NOT be described as preserving unmanaged remote `ssl` when configuration omits it; an unchanged configuration that omits such an object may plan its removal or resource replacement, according to the tested PUT semantics.

#### Scenario: Unchanged configuration with no pre-existing ssl

- **GIVEN** a resource managed by the previous provider version has no server-side `ssl` object
- **WHEN** the provider is upgraded and the unchanged configuration still omits `ssl`
- **THEN** refresh and planning SHALL produce no diff attributable to `ssl`

#### Scenario: Unchanged configuration with pre-existing unmanaged ssl

- **GIVEN** a resource managed by the previous provider version has a server-side `ssl` object that its configuration omits
- **WHEN** the provider is upgraded and the unchanged configuration still omits `ssl`
- **THEN** planning SHALL surface the removal or replacement implied by the tested API semantics
- **AND** the resource documentation SHALL direct the practitioner to add the existing SSL settings to configuration to adopt them

### Requirement: Protect deferred SSL fields on update

The provider SHALL NOT overwrite non-null `agent_*` or `es_*` fields (including empty strings and empty lists) already stored in a host's `ssl` object when those fields are not exposed by the Terraform schema. The provider SHALL send `ssl` on an update only when the planned `ssl` differs from the prior state; an update that leaves `ssl` unchanged SHALL omit it, which Kibana treats as leaving the stored object untouched. Before a PUT that sends `ssl` (including one that removes `ssl`), the provider SHALL read the host and, when it stores non-null `agent_*` or `es_*` fields, reject the update with an actionable diagnostic before sending the mutation. The check and the PUT are separate requests and are not atomic.

#### Scenario: Update cannot erase pre-existing unexposed SSL fields

- **GIVEN** a host's stored `ssl` object contains one or more non-null `agent_*` or `es_*` fields
- **WHEN** an update would replace the stored `ssl` object but the provider cannot preserve those fields
- **THEN** the provider SHALL return an actionable diagnostic before sending the PUT
- **AND** the server-side `agent_*` and `es_*` values SHALL remain unchanged
- **AND** an acceptance test SHALL seed these fields and verify no mutating request is sent

#### Scenario: Removing ssl is blocked on a host with unexposed SSL fields

- **GIVEN** a host's stored `ssl` object contains one or more non-null `agent_*` or `es_*` fields
- **WHEN** the practitioner removes `ssl` from configuration and applies
- **THEN** the provider SHALL return the same diagnostic before sending the PUT

#### Scenario: Unrelated update on a host with unexposed SSL fields

- **GIVEN** a host's stored `ssl` object contains non-null `agent_*` or `es_*` fields and `ssl` is unchanged in configuration
- **WHEN** the practitioner changes another attribute such as `name` and applies
- **THEN** the update SHALL succeed without sending `ssl`, and the `agent_*` and `es_*` values SHALL remain unchanged

### Requirement: ssl removal and replacement semantics

Probing Kibana 9.5.5 showed that a PUT omitting `ssl` leaves the stored object unchanged, that `ssl: {}` and explicit `ssl: null` both clear it, and that a PUT replaces the whole stored object. The provider SHALL therefore remove `ssl` in place by sending an empty `ssl` object, and SHALL NOT force resource replacement for any `ssl` removal.

#### Scenario: Removing ssl clears it in place

- **GIVEN** the prior state has `ssl` and the configuration omits it
- **WHEN** the practitioner applies
- **THEN** the provider SHALL perform an in-place update whose PUT body contains `"ssl": {}`
- **AND** a subsequent read SHALL show `ssl` as null and Kibana SHALL no longer report any supported `ssl` field

#### Scenario: ssl absent from both plan and prior state is not sent

- **GIVEN** neither the plan nor the prior state has `ssl`
- **WHEN** the practitioner updates another attribute
- **THEN** the PUT body SHALL omit `ssl`

#### Scenario: Dropping one child is an in-place update

- **GIVEN** the prior state has `certificate`, `key`, `certificate_authorities`, and `client_auth`
- **WHEN** the practitioner removes exactly one of them from configuration and applies
- **THEN** the provider SHALL perform an in-place update sending the remaining children
- **AND** the removed child SHALL be null in state and absent on the server

### Requirement: Update tolerates Kibana's string-encoded ssl response

Kibana returns the `ssl` object of a PUT response as a JSON-encoded string rather than an object. The provider SHALL NOT depend on parsing the PUT response body and SHALL build state from a read after the update; a host that cannot be found on that read SHALL be reported as an error.

#### Scenario: Update succeeds when the PUT response encodes ssl as a string

- **GIVEN** Kibana answers a PUT with `"ssl":"{...}"` as a string
- **WHEN** the practitioner updates a host with `ssl` configured
- **THEN** the update SHALL succeed and state SHALL reflect the values read back from Kibana

### Requirement: ssl.key is hydrated verbatim

Kibana echoes `ssl.key` unredacted on create, update, and read responses (probed on 9.5.5), so the provider SHALL hydrate `ssl.key` from the API response after create, update, refresh, and import.

#### Scenario: key is read back from the API

- **WHEN** the provider populates state after create, update, refresh, or import
- **THEN** `ssl.key` SHALL equal the value returned by Kibana
