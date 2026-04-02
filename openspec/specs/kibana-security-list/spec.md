# `elasticstack_kibana_security_list` — Schema and Functional Requirements

Resource implementation: `internal/kibana/securitylist`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_security_list` resource: managing Kibana value lists (security lists) via the Kibana Security Lists API. The resource supports creating, reading, updating, and deleting value lists within a Kibana space, composite identity from space and list identifiers, import by composite `id`, mutable list metadata, and accurate mapping between Terraform state and the API response including timestamp and audit fields.

## Schema

```hcl
resource "elasticstack_kibana_security_list" "example" {
  # Identity
  id             = <optional+computed, string> # composite: "<space_id>/<list_id>"; RequiresReplace when changed
  space_id       = <optional+computed, string> # default "default"; RequiresReplace
  list_id        = <optional+computed, string> # human-readable identifier; RequiresReplace; UseStateForUnknown

  # List definition
  name        = <required, string>
  description = <required, string>
  type        = <required, string> # RequiresReplace; one of: binary | boolean | byte | date | date_nanos | date_range | double | double_range | float | float_range | geo_point | geo_shape | half_float | integer | integer_range | ip | ip_range | keyword | long | long_range | shape | short | text

  # Optional
  meta    = <optional, JSON (normalized) string>
  version = <optional+computed, int64>

  # Computed from API
  version_id     = <computed, string>
  immutable      = <computed, bool>
  created_at     = <computed, string>
  created_by     = <computed, string>
  updated_at     = <computed, string>
  updated_by     = <computed, string>
  tie_breaker_id = <computed, string>
}
```

Notes:

- The resource's Markdown description is embedded from `internal/kibana/securitylist/resource-description.md`.
- The `type` attribute's description is embedded from `internal/kibana/securitylist/type-description.md`.
- `meta` uses `jsontypes.NormalizedType` and is marshaled/unmarshaled as a JSON object.

## Requirements

### Requirement: Kibana Security Lists API

The resource SHALL manage value lists through the Kibana Security Lists API: create list, read list, update list, and delete list. After a successful create or update API call, the resource SHALL re-fetch the list using the read API and SHALL fail with an error diagnostic if the re-read returns an empty or nil response.

#### Scenario: Create then authoritative read

- GIVEN a successful create API response
- WHEN create completes
- THEN the provider SHALL re-fetch the list with the read API and SHALL fail with an error if the read returns nil

#### Scenario: Update then authoritative read

- GIVEN a successful update API response
- WHEN update completes
- THEN the provider SHALL re-fetch the list with the read API and SHALL fail with an error if the read returns nil

#### Scenario: Read removes missing lists

- GIVEN a read/refresh
- WHEN the read API returns not found (nil response)
- THEN the provider SHALL remove the resource from state

#### Scenario: Delete idempotency

- GIVEN delete is called
- WHEN the API returns not found for the list
- THEN the provider SHALL treat delete as successful (no error diagnostic)

### Requirement: API error surfacing

For create, read, and update, when the transport layer fails or the API returns an unexpected HTTP status code, the resource SHALL surface error diagnostics to Terraform. Delete SHALL surface errors except when the list is already absent (not found), which SHALL be treated as success.

#### Scenario: Non-success create/update/read

- GIVEN a non-success API response (other than read not-found handled by state removal)
- WHEN the operation completes
- THEN Terraform SHALL receive error diagnostics describing the failure including the HTTP status code

### Requirement: Provider configuration and Kibana client

On create, read, update, and delete, the resource SHALL obtain the Kibana OpenAPI client from the provider. If the client cannot be obtained, the resource SHALL return an error diagnostic and SHALL not proceed to the API.

#### Scenario: Unconfigured provider

- GIVEN the resource has no provider-supplied API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

### Requirement: Identity and composite `id`

After reading from the API, the resource SHALL set `id` to the composite string `<space_id>/<list_id>`, where `space_id` is the Kibana space and `list_id` is the list identifier returned by the API. The `list_id` attribute in state SHALL hold the list's identifier as returned by the API. When the practitioner does not supply `list_id`, Kibana auto-generates it and the resource SHALL persist the generated value to state using `UseStateForUnknown`.

#### Scenario: State has composite id

- GIVEN a list in state returned by the API
- WHEN state is written
- THEN `id` SHALL equal `<space_id>/<list_id>` for that list

#### Scenario: Auto-generated list_id preserved across plans

- GIVEN a list created without a practitioner-supplied `list_id`
- WHEN Terraform plans a subsequent operation with no config changes
- THEN `list_id` SHALL remain the value returned by the API on the initial read (UseStateForUnknown)

### Requirement: Import

The resource SHALL support Terraform import using an `id` value in the format `<space_id>/<list_id>` passed through to state via `ImportStatePassthroughID`. On refresh after import, the resource SHALL parse the composite `id` from state using `CompositeIDFromStrFw` to derive `space_id` and `list_id` for the read API call.

#### Scenario: Import then refresh

- GIVEN an import of id `my-space/my-list-id`
- WHEN the import and subsequent refresh complete
- THEN state SHALL hold `space_id = "my-space"`, `list_id = "my-list-id"`, and `id = "my-space/my-list-id"`

### Requirement: Lifecycle — force replacement

Changing `id`, `space_id`, `list_id`, or `type` SHALL require destroying and recreating the resource rather than an in-place update.

#### Scenario: Replace on immutable field change

- GIVEN an in-place plan change only to `type`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create) for the resource

### Requirement: Default space

When `space_id` is not set by the practitioner, the resource SHALL default `space_id` to `"default"` and SHALL include that value in all API calls and in the computed `id`.

#### Scenario: Default space_id

- GIVEN a resource configured without `space_id`
- WHEN the resource is created and state is read
- THEN `space_id` in state SHALL be `"default"`

### Requirement: Create — optional fields

When creating a list, the resource SHALL send `name`, `description`, and `type` in the request body. When `list_id` is known and set by the practitioner, the resource SHALL include it as the `id` field in the request body. When `meta` is known and set, the resource SHALL unmarshal and include it. When `version` is known and set, the resource SHALL include it in the create request.

#### Scenario: Create with practitioner-supplied list_id

- GIVEN `list_id` is set to a known string in the plan
- WHEN the create request is built
- THEN the API request body SHALL include that value as `id`

#### Scenario: Create without optional fields

- GIVEN `list_id`, `meta`, and `version` are null or unknown in the plan
- WHEN the create request is built
- THEN the API request body SHALL not include those optional fields

### Requirement: Update — mutable fields

The update request SHALL include `list_id` (as the API's `id` field), `name`, and `description` taken from the plan. When `version_id` is known in the plan, it SHALL be sent as `_version` in the update body. When `meta` is known, it SHALL be sent. When `version` is known, it SHALL be sent.

#### Scenario: Update sends version_id for optimistic concurrency

- GIVEN `version_id` is known in state before an update
- WHEN the update request is built
- THEN the request body SHALL include `_version` set to that version_id

### Requirement: Read — composite ID parsing

On read, the resource SHALL attempt to parse `id` from state using `CompositeIDFromStrFw`. When parsing succeeds, the resource SHALL use the parsed `space_id` and `list_id` values for the API call and SHALL update `space_id` in state to the parsed value.

#### Scenario: Read after import uses composite id

- GIVEN a state with `id` = `"default/my-list"`
- WHEN read runs
- THEN the read API call SHALL use `space_id = "default"` and list `id = "my-list"`

### Requirement: Delete

The resource SHALL derive `space_id` and `list_id` from state and call the delete API. The resource SHALL surface any error diagnostic from the delete call except when the API returns not found, which SHALL be treated as success.

#### Scenario: Delete uses state list_id

- GIVEN state with `list_id = "my-list"` and `space_id = "default"`
- WHEN delete runs
- THEN the delete API call SHALL use those values

### Requirement: Mapping — `meta` field

When the API response includes a non-nil `meta` object, the resource SHALL marshal it to a JSON string and store it as a normalized JSON value in state. When the API returns a nil `meta`, the resource SHALL store a null normalized JSON value in state. When the marshal operation fails, the resource SHALL surface an error diagnostic.

#### Scenario: Non-nil meta from API

- GIVEN the API returns a list with a non-nil `meta` map
- WHEN the provider maps the response to state
- THEN `meta` in state SHALL be the JSON-serialized form of that map

#### Scenario: Nil meta from API

- GIVEN the API returns a list with a nil `meta`
- WHEN the provider maps the response to state
- THEN `meta` in state SHALL be null

### Requirement: Mapping — timestamps and audit fields

The resource SHALL map `created_at`, `updated_at`, `created_by`, `updated_by`, `version`, `version_id`, `immutable`, and `tie_breaker_id` from the API response to state. Timestamps SHALL be formatted using RFC3339.

#### Scenario: Timestamps stored in RFC3339

- GIVEN the API returns `created_at` and `updated_at` as time values
- WHEN the provider maps the response to state
- THEN `created_at` and `updated_at` in state SHALL be RFC3339-formatted strings

### Requirement: Mapping — `type` validator

The `type` attribute SHALL only accept values from the set: `binary`, `boolean`, `byte`, `date`, `date_nanos`, `date_range`, `double`, `double_range`, `float`, `float_range`, `geo_point`, `geo_shape`, `half_float`, `integer`, `integer_range`, `ip`, `ip_range`, `keyword`, `long`, `long_range`, `shape`, `short`, `text`. Any other value SHALL be rejected at plan time by the schema validator.

#### Scenario: Invalid type rejected at plan

- GIVEN `type` is set to a string outside the allowed set
- WHEN Terraform validates the configuration
- THEN the provider SHALL return a validation diagnostic for `type`

## Traceability (implementation index)

| Area | Primary files |
|------|---------------|
| Schema | `internal/kibana/securitylist/schema.go` |
| Metadata / Configure / Import | `internal/kibana/securitylist/resource.go` |
| Create | `internal/kibana/securitylist/create.go` |
| Read | `internal/kibana/securitylist/read.go` |
| Update | `internal/kibana/securitylist/update.go` |
| Delete | `internal/kibana/securitylist/delete.go` |
| Model mapping | `internal/kibana/securitylist/models.go` |
| HTTP client helpers | `internal/clients/kibanaoapi/security_lists.go` |
| Composite id parsing | `internal/clients/api_client.go` (`CompositeID`, `CompositeIDFromStrFw`) |
