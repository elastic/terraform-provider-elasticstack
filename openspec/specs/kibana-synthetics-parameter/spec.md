# `elasticstack_kibana_synthetics_parameter` — Schema and Functional Requirements

Resource implementation: `internal/kibana/synthetics/parameter`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_synthetics_parameter` resource, which manages Kibana Synthetics parameters (key/value pairs used in synthetic monitor configurations). The resource covers space-aware Synthetics Parameters API routing, composite identity and import, provider-level Kibana client usage, read-after-write on create and update, and the mapping of `share_across_spaces` to Kibana namespaces.

## Schema

```hcl
resource "elasticstack_kibana_synthetics_parameter" "example" {
  id                  = <computed, string>          # "<space_id>/<parameter_uuid>"; UseStateForUnknown; RequiresReplace
  key                 = <required, string>          # Parameter key; UseStateForUnknown
  value               = <required, string, sensitive> # Parameter value; UseStateForUnknown
  space_id            = <optional, computed, string> # Default "default"; UseStateForUnknown; RequiresReplace
  description         = <optional, computed, string> # Default ""; UseStateForUnknown
  tags                = <optional, computed, list(string)> # Default []; UseStateForUnknown
  share_across_spaces = <optional, computed, bool>  # Default false; UseStateForUnknown; RequiresReplace
}
```

Notes:

- The resource uses the provider-level Kibana OpenAPI (`kbapi`) client for all CRUD operations (create, read, update, and delete).
- For create and update, the provider serializes the request DTO with `encoding/json` and sends it with the `WithBody` request methods due to oapi-codegen oneOf limitations (see [oapi-codegen#1620](https://github.com/oapi-codegen/oapi-codegen/issues/1620)).
- `share_across_spaces` is not sent to the API on update calls; it is only sent on create.
- The `id` field has both `UseStateForUnknown` and `RequiresReplace` plan modifiers. Because `id` is computed and set by Kibana, any change that causes a new parameter to be created will also generate a new `id`.
- There is no schema version or state upgrade defined for this resource.

## Requirements

### Requirement: Space-scoped Synthetics Parameters API (REQ-001)

The resource SHALL manage Synthetics parameters through Kibana's Synthetics Parameters API using the space-aware path pattern: create via `POST /s/{space_id}/api/synthetics/params`, read via `GET /s/{space_id}/api/synthetics/params/{id}`, update via `PUT /s/{space_id}/api/synthetics/params/{id}`, and delete via `DELETE /s/{space_id}/api/synthetics/params` (or `DELETE /s/{space_id}/api/synthetics/params/{id}` for Kibana ≥ 8.17.0). When `space_id` is `"default"` or empty, the path SHALL remain unchanged (no `/s/{space_id}` prefix injected). All operations SHALL use `kibanautil.SpaceAwarePathRequestEditor(spaceID)` to rewrite the request URL path before the call is sent. All operations SHALL use the same Kibana OpenAPI (`kbapi`) HTTP transport for authentication and headers.

#### Scenario: CRUD uses space-aware Synthetics Parameters APIs

- GIVEN a managed Synthetics parameter with `space_id = "my-space"`
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the corresponding Kibana Synthetics Parameters API with the `/s/my-space/` path prefix

#### Scenario: Default space uses unscoped path

- GIVEN a managed Synthetics parameter with `space_id = "default"` (or `space_id` unset)
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the Kibana Synthetics Parameters API without a space path prefix

### Requirement: API and client error surfacing (REQ-002)

The resource SHALL fail with an error diagnostic when it cannot obtain the Kibana OpenAPI client. Transport errors and unexpected API responses for create, read, update, and delete SHALL be surfaced as error diagnostics. On read, a 404 response SHALL cause the resource to be removed from state rather than returning an error.

#### Scenario: Missing Kibana client

- GIVEN the resource cannot obtain a Kibana client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Read returns 404

- GIVEN a parameter that no longer exists in Kibana
- WHEN read calls the API and receives a 404 response
- THEN the provider SHALL remove the resource from state

#### Scenario: Create transport or API error

- GIVEN a create request
- WHEN the API returns a transport error or an unexpected response
- THEN the provider SHALL surface an error diagnostic

### Requirement: Identity and computed `id` (REQ-003)

The resource SHALL expose a computed `id` stored as `<space_id>/<parameter_uuid>` after successful create or update (see REQ-013). The `id` SHALL be preserved across reads using `UseStateForUnknown`. Because `id` also carries `RequiresReplace`, any operation that changes the parameter's Kibana identity SHALL trigger replacement.

#### Scenario: Composite `id` preserved on read

- GIVEN a managed parameter with a composite `id` in state
- WHEN read runs successfully
- THEN the provider SHALL preserve the composite `id` in state using `UseStateForUnknown`

### Requirement: Import by composite or bare id (REQ-004)

The resource SHALL support Terraform import using either a composite `<space_id>/<parameter_uuid>` or a bare `<parameter_uuid>` as the import identifier. When a bare UUID is provided (no `/`), the provider SHALL treat it as belonging to the default space and SHALL populate `space_id = "default"` in state. When a composite identifier is provided, the provider SHALL extract `<space_id>` and `<parameter_uuid>` and populate both `id` and `space_id` in state. On import, a composite `id` where the resource-UUID segment is empty SHALL return an error diagnostic.

#### Scenario: Import with composite id

- GIVEN an import id in the format `<space_id>/<parameter_uuid>`
- WHEN import runs and read is performed
- THEN the provider SHALL set `space_id` to `<space_id>` and `id` to `<space_id>/<parameter_uuid>` in state, and SHALL call the Synthetics Parameters API under the correct space path

#### Scenario: Import with bare UUID

- GIVEN an import id that contains no `/`
- WHEN import runs
- THEN the provider SHALL treat the id as a default-space parameter, set `space_id = "default"`, and set `id = "default/<parameter_uuid>"` in state

#### Scenario: Import with empty resource segment

- GIVEN an import id of the form `<space_id>/` (empty UUID segment)
- WHEN import runs
- THEN the provider SHALL return an error diagnostic

### Requirement: Provider-level Kibana client by default (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI (`kbapi`) client by default for all parameter API operations (create, read, update, and delete). When `kibana_connection` is configured on the resource, the resource SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana OpenAPI client for all of those operations.

#### Scenario: Standard provider connection

- **WHEN** `kibana_connection` is not configured on the resource
- **THEN** all parameter API operations SHALL use the provider-level Kibana OpenAPI client

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource
- **THEN** all parameter API operations SHALL use the scoped Kibana OpenAPI client derived from that block

### Requirement: Read-after-write on create and update (REQ-006)

After a successful create or update API call, the resource SHALL perform a follow-up read of the parameter by id and SHALL use the read response to populate state. This is required because Kibana's create response omits the `value` field and Kibana's update response returns the old `value`. The follow-up GET SHALL use the same space-aware path rules as REQ-001 for the parameter's effective `space_id` (via `SpaceAwarePathRequestEditor`).

#### Scenario: Read-after-write uses space-aware GET in named space

- GIVEN a successful create or update for a parameter with effective `space_id = "my-space"`
- WHEN the provider performs the follow-up read
- THEN the provider SHALL call `GET /s/my-space/api/synthetics/params/{id}` and populate state from the GET response

#### Scenario: Read-after-write uses unscoped GET in default space

- GIVEN a successful create or update for a parameter in the default space
- WHEN the provider performs the follow-up read
- THEN the provider SHALL call `GET /api/synthetics/params/{id}` without a space path prefix and populate state from the GET response

### Requirement: `share_across_spaces` create-only field (REQ-007)

The `share_across_spaces` attribute SHALL be sent to Kibana only on create requests. On update, `share_across_spaces` SHALL be omitted from the request body. Because this field is `RequiresReplace`, any change to it in configuration SHALL trigger resource replacement rather than an in-place update.

#### Scenario: `share_across_spaces` omitted on update

- GIVEN a parameter update that does not change `share_across_spaces`
- WHEN the update request is built
- THEN the request body SHALL NOT include `share_across_spaces`

#### Scenario: Replace on `share_across_spaces` change

- GIVEN an existing managed parameter with `share_across_spaces = false`
- WHEN configuration changes `share_across_spaces` to `true`
- THEN Terraform SHALL plan replacement for the resource

### Requirement: `share_across_spaces` to namespaces mapping (REQ-008)

When reading a parameter from Kibana, the resource SHALL map the API `namespaces` field to the `share_across_spaces` attribute. If `namespaces` equals `["*"]` (all spaces), `share_across_spaces` SHALL be set to `true`; otherwise it SHALL be set to `false`.

#### Scenario: All-spaces parameter maps to `share_across_spaces = true`

- GIVEN Kibana returns a parameter with `namespaces = ["*"]`
- WHEN the provider maps the response to state
- THEN `share_across_spaces` SHALL be `true`

#### Scenario: Single-space parameter maps to `share_across_spaces = false`

- GIVEN Kibana returns a parameter with `namespaces = ["default"]`
- WHEN the provider maps the response to state
- THEN `share_across_spaces` SHALL be `false`

### Requirement: Tags and description defaults (REQ-009)

When `tags` is not configured, the resource SHALL default it to an empty list and SHALL marshal it as an empty JSON array (not null) in API requests. When `description` is not configured, the resource SHALL default it to an empty string and SHALL include it in API requests.

#### Scenario: `tags` defaults to empty list

- GIVEN a parameter configured without `tags`
- WHEN Terraform plans the resource
- THEN `tags` SHALL default to `[]`

#### Scenario: `description` defaults to empty string

- GIVEN a parameter configured without `description`
- WHEN Terraform plans the resource
- THEN `description` SHALL default to `""`

### Requirement: Replacement on `id` change (REQ-010)

Because `id` carries `RequiresReplace`, any configuration or plan change that results in a new computed `id` value SHALL trigger replacement of the resource rather than an in-place update.

#### Scenario: Replace on `id` change

- GIVEN an existing managed parameter
- WHEN the planned `id` differs from the state `id`
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Create and update request bodies encoded with manual JSON (REQ-011)

For create and update, the resource SHALL serialize the parameter request DTO with `encoding/json` and send it with `Content-Type: application/json` through the generated client's `WithBody` request methods, rather than relying on the generated union request types alone, until oapi-codegen correctly encodes the oneOf request body for this API.

#### Scenario: Create uses marshalled JSON body

- GIVEN a parameter create
- WHEN the provider issues the POST request
- THEN the request body SHALL be produced by JSON-marshalling the request DTO and the call SHALL use the OpenAPI client's body-based POST method for parameters

#### Scenario: Update uses marshalled JSON body

- GIVEN a parameter update
- WHEN the provider issues the PUT request
- THEN the request body SHALL be produced by JSON-marshalling the request DTO and the call SHALL use the OpenAPI client's body-based PUT method for parameters

### Requirement: `space_id` attribute (REQ-012)

The resource SHALL expose an **optional, computed** `space_id` string attribute defined via the canonical `kbschema.ResourceSpaceIDAttribute()` helper: `Default` of `"default"` (`clients.DefaultSpaceID`), with `UseStateForUnknown` and `RequiresReplace` plan modifiers. When `space_id` is not configured, the schema default SHALL materialize `"default"` before create or update. The model SHALL NOT implement `KibanaUnscopedSpace`; the envelope's normal non-empty `space_id` validation SHALL apply and SHALL be satisfied by the schema default.

#### Scenario: `space_id` defaults to "default"

- GIVEN a parameter configured without `space_id`
- WHEN create runs
- THEN `space_id` SHALL be set to `"default"` in state

#### Scenario: `space_id` routes to named space

- GIVEN a parameter configured with `space_id = "ops-team"`
- WHEN create runs
- THEN the provider SHALL POST to `/s/ops-team/api/synthetics/params` and store `space_id = "ops-team"` in state

#### Scenario: Replace on `space_id` change

- GIVEN an existing managed parameter with `space_id = "default"`
- WHEN configuration changes `space_id` to `"ops-team"`
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Composite `id` encoding (REQ-013)

The resource `id` SHALL be stored as a composite `<space_id>/<parameter_uuid>` string (via `clients.CompositeID`) after a successful create or update. This encoding enables `resolveKibanaResourceIdentity` to recover both the UUID and the space from state without requiring a separate read of the `space_id` field. On read-after-write, `modelFromOAPI` SHALL accept the `spaceID` from the write context and assemble the composite `id`. A legacy bare-UUID `id` (no `/`) in prior state SHALL continue to resolve correctly: it SHALL fall back to the bare UUID as the resource id and the default space, and SHALL be rewritten to composite form on the next create, update, or refresh. No schema-version bump or `StateUpgraders` migration SHALL be introduced.

#### Scenario: `id` set to composite after create

- GIVEN a successful parameter create in space `"my-space"`
- WHEN Kibana returns the new parameter UUID `"abc-123"`
- THEN the provider SHALL store `id = "my-space/abc-123"` in state

#### Scenario: `id` set to composite in default space

- GIVEN a successful parameter create with no explicit `space_id`
- WHEN Kibana returns the new parameter UUID `"abc-123"`
- THEN the provider SHALL store `id = "default/abc-123"` in state

### Requirement: Backward compatibility for legacy bare-UUID state (REQ-014)

The resource SHALL remain compatible with existing state that stores a bare-UUID `id` without any schema-version bump or `StateUpgraders` migration. A legacy default-space parameter SHALL continue to be readable, updatable, and deletable: `resolveKibanaResourceIdentity` SHALL parse the bare UUID (no `/`) as the resource id with the default space, and CRUD SHALL route to the unscoped path. The bare-UUID `id` SHALL be rewritten to the composite `"default/<uuid>"` form on the next successful create, update, or refresh, without destroying or recreating the resource.

#### Scenario: Legacy bare-UUID state resolves to default space

- GIVEN existing state containing a `elasticstack_kibana_synthetics_parameter` with a bare UUID `id` and no `space_id`
- WHEN read, update, or delete runs
- THEN the provider SHALL treat it as a default-space parameter, route to the unscoped Synthetics Parameters path, and (on write/refresh) rewrite `id` to `"default/<uuid>"` with no destroy/recreate

### Requirement: Non-default space at existing API baseline (REQ-015)

Kibana **v8.12.0** documents space-prefixed Synthetics Parameters CRUD routes alongside unscoped routes (Parameters API baseline; see `design.md`). That **8.12.0** marker describes documented API availability for non-default space routing—not a new or raised runtime `GetVersionRequirements` floor introduced by this change. The provider SHALL NOT add a `GetVersionRequirements` entry solely to gate non-default `space_id` values (contrast Synthetics private location, which requires a higher stack version for non-default space).

#### Scenario: Non-default space without extra version gate

- GIVEN a parameter configured with a non-default `space_id`
- WHEN create, read, update, or delete runs on a Kibana deployment where space-prefixed Parameters routes are available
- THEN the provider SHALL route via the space-prefixed Synthetics Parameters API path and SHALL NOT fail with a version diagnostic introduced solely for non-default `space_id` routing

## Traceability

| Area | Primary files |
|------|---------------|
| Schema and model | `internal/kibana/synthetics/parameter/schema.go` |
| Metadata / Configure / Import | `internal/kibana/synthetics/parameter/resource.go` |
| Create | `internal/kibana/synthetics/parameter/create.go` |
| Read | `internal/kibana/synthetics/parameter/read.go` |
| Update | `internal/kibana/synthetics/parameter/update.go` |
| Delete | `internal/kibana/synthetics/parameter/delete.go` |
| Shared client helpers | `internal/kibana/synthetics/api_client.go` |
| Shared utilities | `internal/kibana/synthetics/schema.go` |
