# Delta Spec: Fleet Agent Download Source

On sync or archive, this content is intended to land in `openspec/specs/fleet-agent-download-source/spec.md`.

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

  # Space scoping
  space_ids = ["default"] # optional+computed, set(string); when set, the resource is managed in the first space ID
}
```

## ADDED Requirements

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

The resource SHALL support `terraform import` by accepting either `<space_id>/<source_id>` or `<source_id>`. The Fleet download source ID SHALL populate both `id` and `source_id`. When a space is provided, the Space ID SHALL populate `space_ids` as a single-entry collection. When only `<source_id>` is provided, `space_ids` SHALL default to `["default"]`.

#### Scenario: Import by composite ID

- **WHEN** the practitioner runs import with a valid `<space_id>/<source_id>` identifier
- **THEN** state SHALL contain `source_id` in both `id` and `source_id`, and `space_ids` SHALL contain exactly the provided `space_id`

#### Scenario: Import by source ID only

- **WHEN** the practitioner runs import with `<source_id>` and no explicit space prefix
- **THEN** state SHALL contain `source_id` in both `id` and `source_id`, and `space_ids` SHALL contain exactly `default`

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

The initial implementation SHALL NOT expose the Fleet download source `auth` and `secrets` fields (API key, username/password, SSL key material) as Terraform schema attributes. Those fields remain managed outside this resource’s schema for the first version.

#### Scenario: No sensitive auth attributes

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be no Terraform attributes exposing `auth` or `secrets` material for this resource in v1

### Requirement: Minimum Kibana version guard

The resource SHALL be guarded by a minimum Kibana/Fleet version that supports the Agent Binary Download Sources API. If the connected Kibana is below that version, the provider SHALL emit a clear diagnostic during plan or apply indicating that the resource is not supported.

#### Scenario: Unsupported stack version

- **WHEN** the connected Kibana is below the supported minimum version
- **THEN** the provider SHALL emit a diagnostic that the resource is not supported on this version
