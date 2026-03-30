# `elasticstack_kibana_security_list_data_streams` — Schema and Functional Requirements

Resource implementation: `internal/kibana/security_list_data_streams`

## Purpose

Create and manage the `.lists` and `.items` Elasticsearch data streams required by Kibana Security for value lists and exceptions. On create, the resource calls Kibana's list index API to initialise those data streams in the configured space, then verifies both exist. On read, the resource checks whether both data streams are present and removes itself from state if either is missing. On delete, the resource calls the delete list index API to remove them. Import is supported using the space ID as the import identifier.

## Schema

```hcl
resource "elasticstack_kibana_security_list_data_streams" "example" {
  # Identity
  id       = <computed, string>           # equals space_id; UseStateForUnknown
  space_id = <optional, computed, string> # default "default"; RequiresReplace

  # Status (computed from API)
  list_index      = <computed, bool> # true when the .lists data stream exists
  list_item_index = <computed, bool> # true when the .items data stream exists
}
```

## Requirements

### Requirement: Create list data streams

On create, the resource SHALL call Kibana's create list index API (`CreateListIndex`) for the configured `space_id`. A 200 response SHALL be treated as success. A 409 (conflict) response SHALL also be treated as success because the data streams already exist. Any other status SHALL produce an error diagnostic.

#### Scenario: Successful create

- GIVEN `space_id = "default"` and neither data stream exists yet
- WHEN create runs
- THEN the resource SHALL call `CreateListIndex` for `"default"` and SHALL proceed to verify the data streams

#### Scenario: Conflict treated as success

- GIVEN the data streams already exist and the API returns 409
- WHEN create runs
- THEN the resource SHALL treat the response as success and SHALL NOT surface an error diagnostic

#### Scenario: Other error on create

- GIVEN the create API returns a status other than 200 or 409
- WHEN create runs
- THEN the resource SHALL surface an error diagnostic

### Requirement: Post-create verification

After a successful create API call, the resource SHALL read the data stream status with `ReadListIndex` and SHALL fail with an error diagnostic if either `list_index` or `list_item_index` is `false`.

#### Scenario: Verification fails after create

- GIVEN the create API succeeds but `ReadListIndex` returns `list_index = false`
- WHEN create runs
- THEN the resource SHALL fail with `Failed to verify list data streams`

### Requirement: Read — remove from state if data streams absent

On read (refresh), the resource SHALL call `ReadListIndex`. If either `list_index` or `list_item_index` is `false` (including a 404 response from the API), the resource SHALL remove itself from Terraform state. If both are `true`, the resource SHALL update state with the current values.

#### Scenario: Both data streams present

- GIVEN `ReadListIndex` returns `list_index = true` and `list_item_index = true`
- WHEN read runs
- THEN state SHALL be updated with both fields set to `true`

#### Scenario: Data stream missing

- GIVEN `ReadListIndex` returns `list_index = false` or `list_item_index = false`
- WHEN read runs
- THEN the resource SHALL remove itself from state

#### Scenario: API 404 on read

- GIVEN the read list index API returns 404
- WHEN read runs
- THEN the resource SHALL remove itself from state (not an error)

### Requirement: Delete list data streams

On delete, the resource SHALL call Kibana's delete list index API (`DeleteListIndex`) for `space_id`. A 200 response SHALL be treated as success. A 404 response SHALL also be treated as success (idempotent delete). Any other status SHALL produce an error diagnostic.

#### Scenario: Successful delete

- GIVEN an existing resource in state with `space_id = "default"`
- WHEN destroy runs
- THEN the resource SHALL call `DeleteListIndex` for `"default"`

#### Scenario: Not-found on delete treated as success

- GIVEN the delete API returns 404
- WHEN destroy runs
- THEN the resource SHALL complete successfully without an error diagnostic

### Requirement: Identity

The resource SHALL set `id` equal to `space_id`. The `id` attribute SHALL be preserved across reads using `UseStateForUnknown`.

#### Scenario: id equals space_id

- GIVEN `space_id = "my-space"`
- WHEN create succeeds
- THEN `id` SHALL equal `"my-space"`

### Requirement: Import

The resource SHALL support Terraform import using the space ID as the import identifier. On import, the `id` SHALL be passed through directly and `space_id` SHALL be derived from `id` during the subsequent read.

#### Scenario: Import by space_id

- GIVEN an import with id `"default"`
- WHEN import and the subsequent read run
- THEN state SHALL have `id = "default"` and `space_id = "default"`

### Requirement: Read — derive space_id from id during import

When `space_id` is not yet known in state (for example immediately after import) but `id` is known, the resource SHALL derive `space_id` from `id` before calling `ReadListIndex`.

#### Scenario: space_id derived from id on import read

- GIVEN state has `id = "my-space"` but `space_id` is unknown
- WHEN read runs
- THEN the resource SHALL use `"my-space"` as the space for the `ReadListIndex` call

### Requirement: Lifecycle — force replacement on space_id

Changing `space_id` SHALL require destroying and recreating the resource.

#### Scenario: Replace on space_id change

- GIVEN an existing resource and a plan that changes `space_id`
- WHEN Terraform evaluates the plan
- THEN the plan SHALL indicate replace (destroy/create)

### Requirement: Provider configuration and Kibana client

On every CRUD operation, the resource SHALL use the provider's configured Kibana OAPI client. If the provider data cannot be converted to a valid API client, the resource SHALL return a configuration error diagnostic.

#### Scenario: Unconfigured provider

- GIVEN the provider has not supplied a usable API client
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with a provider configuration error

### Requirement: Update is a no-op

Because `space_id` (the only configurable attribute) has a `RequiresReplace` plan modifier, an in-place update SHALL NOT occur in practice. If the framework calls update, the resource SHALL write the plan directly to state without calling any API.

#### Scenario: Update passes plan to state

- GIVEN the framework calls update
- WHEN update runs
- THEN the resource SHALL set state from the plan without making any API call
