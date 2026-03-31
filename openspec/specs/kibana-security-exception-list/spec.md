# `elasticstack_kibana_security_exception_list` — Schema and Functional Requirements

Resource implementation: `internal/kibana/securityexceptionlist`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_security_exception_list` resource: manage Kibana exception lists via the Kibana Security Exceptions HTTP API, composite identity and import, provider-level Kibana connection only, stable mapping between Terraform state and API payloads, and authoritative read-after-write to prevent plan drift.

## Schema

```hcl
resource "elasticstack_kibana_security_exception_list" "example" {
  # Identity
  id             = <computed, string>                   # composite: "<space_id>/<list_uuid>"; UseStateForUnknown
  space_id       = <optional, computed, string>         # default "default"; RequiresReplace
  list_id        = <optional, computed, string>         # human-readable identifier; RequiresReplaceIfConfigured

  # Required fields
  name           = <required, string>
  description    = <required, string>
  type           = <required, string>                   # one of: detection | endpoint | endpoint_trusted_apps | endpoint_events | endpoint_host_isolation_exceptions | endpoint_blocklists; RequiresReplace

  # Optional fields
  namespace_type = <optional, computed, string>         # "single" (default) or "agnostic"; RequiresReplace
  os_types       = <optional, set(string)>              # elements: "linux" | "macos" | "windows"
  tags           = <optional, set(string)>
  meta           = <optional, JSON (normalized) string>

  # Computed (read-only from API)
  created_at     = <computed, string>
  created_by     = <computed, string>
  updated_at     = <computed, string>
  updated_by     = <computed, string>
  immutable      = <computed, bool>
  tie_breaker_id = <computed, string>
}
```

## Requirements

### Requirement: Kibana Security Exceptions API (REQ-001–REQ-004)

The resource SHALL manage exception lists through the Kibana Security Exceptions HTTP API: create exception list, read exception list, update exception list, and delete exception list. Reference: [Kibana Exceptions API documentation](https://www.elastic.co/docs/api/doc/kibana/group/endpoint-security-exceptions-api).

#### Scenario: Create then authoritative read

- GIVEN a successful create API response
- WHEN create completes
- THEN the resource SHALL re-fetch the exception list with a read and SHALL fail with an error diagnostic if the list cannot be read back

#### Scenario: Update then authoritative read

- GIVEN a successful update API response
- WHEN update completes
- THEN the resource SHALL re-fetch the exception list with a read and SHALL fail with an error diagnostic if the list cannot be read back

#### Scenario: Read removes missing lists

- GIVEN a read/refresh is triggered
- WHEN the API returns not found for the exception list
- THEN the resource SHALL remove itself from state

### Requirement: API error surfacing (REQ-005)

For create, update, and read, when the request fails or the API returns an empty successful body where list data is required, the resource SHALL surface clear error diagnostics to Terraform. Delete SHALL surface all API errors.

#### Scenario: Non-success create/update

- GIVEN a non-success API response during create or update
- WHEN the operation completes
- THEN Terraform SHALL receive error diagnostics describing the failure

#### Scenario: Empty create response

- GIVEN the create API returns a nil or empty body
- WHEN create runs
- THEN the resource SHALL surface an error diagnostic with summary "Failed to create exception list"

### Requirement: Provider configuration and Kibana client (REQ-006)

On create, read, update, and delete, the resource SHALL use the provider's configured Kibana OAPI HTTP client. If that client cannot be obtained, the resource SHALL return an error diagnostic with summary "Failed to get Kibana client" and SHALL not proceed to the API.

#### Scenario: Unconfigured provider

- GIVEN the provider did not supply a usable Kibana API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic before making any API call

### Requirement: Identity and composite `id` (REQ-007–REQ-008)

After a successful read from the API, the resource SHALL set `id` to the composite string `<space_id>/<list_uuid>`, where `space_id` is the Kibana space and `list_uuid` is the exception list's UUID returned by the API. On read, the resource SHALL parse `space_id` from the composite `id` stored in state.

#### Scenario: State matches composite id

- GIVEN an exception list returned by the API
- WHEN state is written after create or update
- THEN `id` SHALL equal `<space_id>/<list_uuid>` for that list

### Requirement: Import (REQ-009)

The resource SHALL support Terraform import using an `id` in the format `<space_id>/<list_uuid>`. On import, `id` SHALL be passed through directly to state via `ImportStatePassthroughID` and the subsequent read SHALL populate all attributes.

#### Scenario: Valid import id

- GIVEN an import id of the form `my-space/<uuid>`
- WHEN import runs
- THEN state SHALL hold `id = "my-space/<uuid>"` and subsequent read SHALL populate all attributes from the API

#### Scenario: Agnostic list import fallback

- GIVEN an import where `namespace_type` is not yet known in state
- WHEN read is called after import and the first read (without namespace_type) returns not found
- THEN the resource SHALL retry the read with `namespace_type=agnostic` before removing from state

### Requirement: Lifecycle — force replacement (REQ-010)

Changing any of `space_id`, `type`, or `namespace_type` SHALL require destroying and recreating the resource rather than an in-place update. Changing `list_id` when it was previously configured SHALL also require replacement (`RequiresReplaceIfConfigured`).

#### Scenario: Replace on immutable field

- GIVEN a plan that changes only `type`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create) for the resource

### Requirement: Schema defaults (REQ-011)

When `space_id` is not provided in configuration, the resource SHALL default to `"default"`. When `namespace_type` is not provided, the resource SHALL default to `"single"`.

#### Scenario: Default space_id

- GIVEN no `space_id` in configuration
- WHEN the resource is planned
- THEN `space_id` SHALL be `"default"` in the plan

### Requirement: `type` validation (REQ-012)

The `type` attribute SHALL only accept one of the following values: `detection`, `endpoint`, `endpoint_trusted_apps`, `endpoint_events`, `endpoint_host_isolation_exceptions`, `endpoint_blocklists`. Any other value SHALL be rejected at plan time with a validation diagnostic.

#### Scenario: Invalid type rejected

- GIVEN `type` set to a string not in the allowed set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for `type`

### Requirement: `namespace_type` validation (REQ-013)

The `namespace_type` attribute SHALL only accept `"single"` or `"agnostic"`. Any other value SHALL be rejected at plan time with a validation diagnostic.

#### Scenario: Invalid namespace_type rejected

- GIVEN `namespace_type` set to a value other than `"single"` or `"agnostic"`
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for `namespace_type`

### Requirement: `os_types` validation (REQ-014)

Each element of `os_types` SHALL be one of `"linux"`, `"macos"`, or `"windows"`. Any other value SHALL be rejected at plan time with a validation diagnostic.

#### Scenario: Invalid os_types element rejected

- GIVEN `os_types` containing a string not in the allowed set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for `os_types`

### Requirement: Mapping — optional fields on create and update (REQ-015)

When `list_id` is set and known, the resource SHALL include it in the create request body. When `namespace_type` is known, the resource SHALL include it in create and update request bodies. When `os_types` is known and non-empty, the resource SHALL include it in create and update request bodies. When `tags` is known and non-empty, the resource SHALL include it in create and update request bodies. When `meta` is known, the resource SHALL include it (parsed from JSON) in create and update request bodies.

#### Scenario: Omitting unset optional fields

- GIVEN `tags` and `meta` are not configured
- WHEN the create request is built
- THEN the request body SHALL omit `tags` and `meta` rather than sending null or empty values

### Requirement: State mapping — `os_types` and `tags` (REQ-016)

When the API returns a non-empty `os_types` array, state SHALL store it as a non-null set. When the API returns an empty or absent `os_types`, state SHALL store a null set. The same behavior SHALL apply to `tags`.

#### Scenario: Empty tags from API

- GIVEN the API returns no tags for the list
- WHEN state is written
- THEN `tags` SHALL be null in state (not an empty set)

#### Scenario: Non-empty os_types from API

- GIVEN the API returns `os_types: ["linux"]`
- WHEN state is written
- THEN `os_types` SHALL be a non-null set containing `"linux"`

### Requirement: State mapping — `meta` (REQ-017)

When the API returns a non-null `meta` object, the resource SHALL marshal it to a JSON string and store it as a normalized JSON value in state. When the API returns a null `meta`, state SHALL store a normalized JSON null. If marshalling `meta` fails, the resource SHALL return an error diagnostic.

#### Scenario: Null meta from API

- GIVEN the API returns no `meta` field
- WHEN state is written
- THEN `meta` SHALL be null in state

### Requirement: State mapping — audit fields (REQ-018)

The resource SHALL store `created_at`, `created_by`, `updated_at`, and `updated_by` from the API response as strings in state, formatted as `2006-01-02T15:04:05.000Z`. It SHALL also store `immutable` and `tie_breaker_id` from the API response.

#### Scenario: Audit fields stored from API

- GIVEN an API response with all audit fields
- WHEN state is written
- THEN `created_at`, `created_by`, `updated_at`, `updated_by`, `immutable`, and `tie_breaker_id` SHALL reflect the API values

### Requirement: Update request includes `type` (REQ-019)

When building the update request body, the resource SHALL include the `type` field even though `type` has `RequiresReplace` in the schema, because the Kibana API requires `type` on updates.

#### Scenario: Type included in update body

- GIVEN an in-place update of `name` on an existing exception list
- WHEN the update request is built
- THEN the request body SHALL include the `type` field matching the current state value

## Traceability (implementation index)

| Area | Primary files |
|------|---------------|
| Schema | `schema.go` |
| Metadata / Configure / Import | `resource.go` |
| Create | `create.go` |
| Read | `read.go` |
| Update | `update.go` |
| Delete | `delete.go` |
| Model mapping | `models.go` |
