# `elasticstack_fleet_agent_download_source` — Schema and Functional Requirements

Resource implementation: `internal/fleet/agentdownloadsource`

## Purpose

Define schema and behavior for the Terraform resource `elasticstack_fleet_agent_download_source`, which manages Fleet **Agent Binary Download Source** objects via the Kibana Fleet API, including optional Kibana space scoping.

## Schema

```hcl
resource "elasticstack_fleet_agent_download_source" "example" {
  # Required identity/config
  name = "My Agent Download Source"
  host = "https://artifacts.example.com/elastic-agent"

  # Optional identity
  source_id = "custom-download-source-id" # optional, string; if omitted, Kibana generates an ID

  # Optional arguments
  default  = false          # optional, bool; when true, marks this as the default download source
  proxy_id = "proxy-123456" # optional, string; references a Fleet proxy by ID

  # Optional authentication for a private artifact registry (requires Kibana 9.4.0+)
  auth = {                     # optional, single nested object
    username = "svc-account"   # optional, string; requires password
    password = "..."           # optional, string, sensitive; requires username; conflicts with api_key
    api_key  = "..."           # optional, string, sensitive; conflicts with password
    headers = [                # optional, list of { key, value }; max 100 entries
      { key = "X-Custom-Header", value = "value" },
    ]
  }

  # Space scoping
  space_ids = ["default"] # optional+computed, set(string); when set, the resource is managed in the first space ID
}
```

## Requirements

### Requirement: Fleet Agent Download Source CRUD APIs

The resource SHALL use the Kibana Fleet **Agent binary download sources** APIs to manage Agent Binary Download Source objects: create with `POST /api/fleet/agent_download_sources`, list with `GET /api/fleet/agent_download_sources`, read by ID with `GET /api/fleet/agent_download_sources/{source_id}`, update with `PUT /api/fleet/agent_download_sources/{source_id}`, and delete with `DELETE /api/fleet/agent_download_sources/{source_id}`.

#### Scenario: Create calls POST

- **WHEN** the practitioner applies a new `elasticstack_fleet_agent_download_source` that does not yet exist in state
- **THEN** the provider SHALL send a create request to the Fleet agent download sources create endpoint

#### Scenario: Delete calls DELETE

- **WHEN** the practitioner destroys the resource or replaces it in a way that removes the download source
- **THEN** the provider SHALL send a delete request to `DELETE /api/fleet/agent_download_sources/{source_id}`

### Requirement: Generated Kibana client and Fleet wrapper

The resource SHALL use the generated Kibana OpenAPI client in `generated/kbapi` via the Fleet wrapper client (`internal/clients/fleet`) for all HTTP interactions, including the space-aware request helpers where applicable.

#### Scenario: No ad hoc HTTP outside kbapi

- **WHEN** the provider performs any Fleet agent download source API call
- **THEN** the call SHALL be made through the Fleet wrapper using the generated client types

### Requirement: Terraform state identity

The Terraform state `id` attribute SHALL store the Fleet download source ID returned by Kibana. The `source_id` attribute SHALL mirror this value and SHALL be used as the path parameter for read, update, and delete calls.

#### Scenario: Read uses stored ID

- **WHEN** the provider refreshes or reads an existing managed download source
- **THEN** the provider SHALL use the download source ID from `id` / `source_id` as `{source_id}` in the read request path

### Requirement: Optional user-specified source ID on create

When `source_id` is set in configuration, the create request SHALL pass it through to Kibana so that the download source is created with that ID. When `source_id` is omitted, the provider SHALL accept the server-generated ID from the create response and persist it to both `id` and `source_id` in state.

#### Scenario: Omitted source ID uses server ID

- **WHEN** the practitioner creates a resource without setting `source_id`
- **THEN** the provider SHALL store the ID returned by Kibana in `id` and `source_id`

#### Scenario: Set source ID on create

- **WHEN** the practitioner sets `source_id` in configuration for a new resource
- **THEN** the provider SHALL include that ID in the create request so Kibana creates the object with that ID

### Requirement: Required attributes name and host

The resource SHALL require `name` (string) and `host` (string) and SHALL map them to the corresponding fields in the Fleet API request and response bodies.

#### Scenario: Missing required attribute

- **WHEN** configuration omits `name` or `host`
- **THEN** Terraform SHALL report a validation error before any API call

### Requirement: Optional attributes default and proxy_id

The resource SHALL expose optional attributes `default` (bool), mapped to `is_default` in the API, and `proxy_id` (string), mapped to `proxy_id` in the API.

#### Scenario: Defaults mapped on write

- **WHEN** the practitioner sets `default = true` and `proxy_id` to a Fleet proxy ID
- **THEN** the provider SHALL send the corresponding `is_default` and `proxy_id` fields expected by the API

### Requirement: Kibana space scoping via space_ids

The resource SHALL expose `space_ids` as an optional+computed `set(string)`. When `space_ids` is set, the first value SHALL determine the Kibana space used for create, read, update, and delete operations via the space-aware Fleet client helpers. When `space_ids` is unset or empty, the resource SHALL operate in the default space. Since the Fleet download sources API does not return space information, `space_ids` SHALL be preserved from Terraform state and SHALL NOT be derived from the API.

#### Scenario: Non-default space

- **WHEN** `space_ids` contains a non-default space as its first element
- **THEN** the provider SHALL issue Fleet API requests in that space

#### Scenario: space_ids not from API

- **WHEN** the provider completes a successful read of a download source
- **THEN** the provider SHALL NOT overwrite `space_ids` solely from API response data

### Requirement: Terraform import

The resource SHALL support `terraform import` by accepting either `<space_id>/<source_id>` or `<source_id>`. The Fleet download source ID SHALL populate `source_id` in state. When a space is provided, the space ID SHALL populate `space_ids` as a single-entry collection. When only `<source_id>` is provided (no `/` separator), `space_ids` SHALL NOT be set from the import ID; the subsequent read cycle SHALL populate all remaining attributes using the default space.

#### Scenario: Import by composite ID

- **WHEN** the practitioner runs import with a valid `<space_id>/<source_id>` identifier
- **THEN** state SHALL contain the parsed resource ID in `source_id`, and `space_ids` SHALL contain exactly the provided `space_id`

#### Scenario: Import by source ID only

- **WHEN** the practitioner runs import with `<source_id>` and no explicit space prefix
- **THEN** state SHALL contain `source_id` set to the provided value, and `space_ids` SHALL NOT be set from the import ID

### Requirement: Error handling and read failures

When the Fleet API returns a non-success status code for create, update, or delete, the resource SHALL surface the error (HTTP status and message or body) via Terraform diagnostics. For read operations, a `404 Not Found` for a specific `source_id` SHALL cause the resource to be removed from state. Other non-success statuses on read SHALL be reported as diagnostics.

#### Scenario: Mutation error surfaces diagnostic

- **WHEN** create, update, or delete returns a non-success HTTP status
- **THEN** the provider SHALL return a Terraform diagnostic with the error details

#### Scenario: Read 404 removes from state

- **WHEN** a read for the configured `source_id` returns HTTP 404
- **THEN** the provider SHALL remove the resource from state

### Requirement: Create and update SHALL converge through read path

After successful create and update operations, the provider SHALL read the download source back from the API using the same read path used during refresh/read and SHALL derive Terraform state from that read result instead of directly trusting mutation response payloads.

#### Scenario: Create uses shared read path for final state

- **WHEN** create returns success
- **THEN** the provider SHALL perform a follow-up read by `source_id` and populate state from the read response

#### Scenario: Update uses shared read path for final state

- **WHEN** update returns success
- **THEN** the provider SHALL perform a follow-up read by `source_id` and populate state from the read response

### Requirement: In-place updates and replacement for source_id

The resource SHALL support in-place updates for changes to `name`, `host`, `default`, and `proxy_id`, issuing a `PUT` to the Fleet API. Changes to `source_id` SHALL force replacement.

#### Scenario: In-place update

- **WHEN** the practitioner changes only `name`, `host`, `default`, or `proxy_id`
- **THEN** the provider SHALL perform an update via `PUT` without replacing the resource

#### Scenario: source_id change forces replacement

- **WHEN** the practitioner changes `source_id` in configuration for an existing resource
- **THEN** Terraform SHALL plan a replace operation

### Requirement: Auth and secrets not in v1 schema

The resource SHALL expose an optional `auth` nested attribute mapping to the Fleet API's `auth` request/response field: `username` (string), `password` (string, sensitive), `api_key` (string, sensitive), and `headers` (list of `{key, value}`, up to 100 entries). The resource SHALL NOT expose the Fleet download source `secrets.auth` reference mechanism (referencing an existing Fleet secret store entry by ID) or the top-level `ssl` object (CA, client cert, client key material) as Terraform schema attributes; those remain managed outside this resource's schema.

#### Scenario: Auth attribute present in schema

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `auth` attribute exposing `username`, `password`, `api_key`, and `headers`

#### Scenario: No sensitive auth attributes

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be no Terraform attributes exposing the `secrets.auth` `{id}` reference mechanism or the `ssl` object (CA, client cert, client key material) for this resource

### Requirement: Minimum Kibana version guard

The resource SHALL be guarded by a minimum Kibana/Fleet version that supports the Agent Binary Download Sources API. If the connected Kibana is below that version, the provider SHALL emit a clear diagnostic during Create, Read, Update, or Delete indicating that the resource is not supported, consistent with the shared `KibanaResource[T]` envelope's uniform version enforcement across all lifecycle operations.

#### Scenario: Unsupported stack version

- **WHEN** the connected Kibana is below the supported minimum version
- **AND** the operation is Create, Read, Update, or Delete
- **THEN** the provider SHALL emit a diagnostic that the resource is not supported on this version

### Requirement: Auth credential-mode validation

The `auth` attribute's `username` and `password` SHALL be set together (each requires the other), and `password` SHALL conflict with `api_key` (mutually exclusive). `headers` SHALL carry no dependency on `username`, `password`, or `api_key`, so an `auth` block containing only `headers` SHALL be valid, and so SHALL an empty `auth = {}` block.

#### Scenario: Username without password rejected

- **WHEN** `auth = { username = "svc" }` is configured with no `password`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Password without username rejected

- **WHEN** `auth = { password = "secret" }` is configured with no `username`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Password and api_key together rejected

- **WHEN** `auth` sets both `password` and `api_key`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Headers-only auth block accepted

- **WHEN** `auth = { headers = [{ key = "X-A", value = "1" }] }` is configured with no credentials
- **THEN** Terraform SHALL accept the configuration and send the request with only `headers` populated under `auth`

#### Scenario: Empty auth object accepted

- **WHEN** `auth = {}` is configured
- **THEN** Terraform SHALL accept the configuration

### Requirement: Auth headers size limit

The `headers` list within `auth` SHALL accept at most 100 entries.

#### Scenario: Header count at limit

- **WHEN** `auth.headers` is configured with exactly 100 entries
- **THEN** Terraform SHALL accept the configuration

#### Scenario: Header count over limit

- **WHEN** `auth.headers` is configured with 101 entries
- **THEN** Terraform SHALL report a validation error before any API call

### Requirement: Auth mapped on create and update

When `auth` is present in configuration, the create and update requests SHALL include the corresponding `auth` object with `username`, `password`, `api_key`, and `headers` populated from configuration. When `auth` is absent from configuration on create, the create request SHALL omit `auth`.

#### Scenario: Create includes auth

- **WHEN** the practitioner creates a resource with `auth` configured
- **THEN** the create request body SHALL include an `auth` object matching the configured fields

#### Scenario: Create omits auth when unset

- **WHEN** the practitioner creates a resource without configuring `auth`
- **THEN** the create request body SHALL NOT include an `auth` key

### Requirement: Auth removal clears credentials on update

When the planned configuration omits `auth` but the prior state had `auth` configured, the update request SHALL send an explicit empty `auth` object (`"auth":{}`) rather than omitting the key, so Fleet clears the existing credentials and headers instead of merging. See scenarios below for the omit-when-never-configured and full-replacement-on-change cases.

#### Scenario: Removing the auth block clears it server-side

- **GIVEN** a resource with `auth.username`/`auth.password` configured and applied
- **WHEN** the practitioner removes the `auth` block entirely and applies
- **THEN** the update request body SHALL include `"auth":{}`
- **AND** a subsequent read SHALL show no credentials or headers configured

#### Scenario: Omitting auth from an update that never configured it is a no-op for auth

- **GIVEN** a resource created without `auth`
- **WHEN** the practitioner updates an unrelated attribute (e.g. `name`) without ever configuring `auth`
- **THEN** the update request body SHALL NOT include an `auth` key

#### Scenario: Setting a new credential replaces the full credential set

- **GIVEN** a resource with `auth.username`/`auth.password` configured
- **WHEN** the practitioner changes configuration to `auth.api_key` only and applies
- **THEN** the update request body SHALL include an `auth` object containing only `api_key` (and any configured `headers`), and the prior `username`/`password` SHALL no longer be present server-side

### Requirement: Auth password and api_key are never echoed on read

Kibana never returns `auth.password` or `auth.api_key` in any response; it stores the value in its secret store and returns only an opaque, unmapped reference. When the API omits either field, the resource SHALL preserve it from a fallback instead of writing null. `username`/`headers` always come from the API directly. See scenarios for the per-path fallback source.

#### Scenario: Password survives the post-create read

- **WHEN** the practitioner creates a resource with `auth.password` configured
- **THEN** the mandatory read-after-create SHALL populate state's `auth.password` with the configured value, not null, even though the create response omits it

#### Scenario: Rotated password survives the post-update read

- **GIVEN** a resource with `auth.password` already set
- **WHEN** the practitioner changes `auth.password` to a new value and applies
- **THEN** the mandatory read-after-update SHALL populate state's `auth.password` with the newly configured value, not the prior value and not null

#### Scenario: Refresh preserves the previously known password

- **GIVEN** state already has `auth.password` populated from a prior apply
- **WHEN** Terraform performs a plain refresh (no configuration change)
- **THEN** state's `auth.password` SHALL remain the previously known value, since the API read omits it

#### Scenario: Username and headers always come from the API

- **WHEN** the resource is read after create, update, or during refresh
- **THEN** state's `auth.username` and `auth.headers` SHALL reflect the API response directly, not a preserved/fallback value

### Requirement: Auth headers null vs empty list distinction

When the API response includes `auth.headers` as an empty list, state SHALL reflect an empty list, not null. When the API response omits `auth.headers` entirely (or omits `auth` entirely), state SHALL reflect null for `headers` (and for the whole `auth` object, respectively).

#### Scenario: Configured empty headers list round-trips as empty, not null

- **WHEN** the practitioner configures `auth = { headers = [] }`
- **THEN** state's `auth.headers` SHALL be an empty list after apply, not null

### Requirement: Auth minimum Kibana version

When `auth` is configured (non-null), the provider SHALL enforce a minimum Kibana version of 9.4.0 for the `auth` attribute, independent of the resource's overall minimum version. When `auth` is not configured, the resource's existing overall minimum version requirement applies unchanged.

#### Scenario: Auth rejected below minimum version

- **WHEN** the connected Kibana is below 9.4.0
- **AND** the practitioner configures `auth` on this resource
- **THEN** the provider SHALL emit a diagnostic that `auth` is not supported on this version

#### Scenario: Resource without auth still works below the auth minimum version

- **WHEN** the connected Kibana is at or above the resource's overall minimum version but below 9.4.0
- **AND** the practitioner's configuration omits `auth`
- **THEN** the provider SHALL NOT emit an `auth`-related version diagnostic

### Requirement: Import does not populate auth secrets

On `terraform import`, the resource SHALL NOT populate `auth.password` or `auth.api_key` from any prior Terraform state, since none exists for a freshly imported resource. The subsequent read cycle SHALL leave those two fields null whenever the API omits them (which it always does), regardless of whether the imported download source was originally created through Terraform or the Kibana UI. `auth.username` and `auth.headers` SHALL populate normally from the API response on import.

#### Scenario: Imported resource has null auth secrets

- **WHEN** the practitioner imports a download source that has `auth.password` configured server-side
- **THEN** state's `auth.password` SHALL be null after import
- **AND** state's `auth.username` and `auth.headers` (if configured server-side) SHALL be populated from the API response
