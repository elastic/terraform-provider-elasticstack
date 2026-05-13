# `elasticstack_elasticsearch_ml_filter` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/ml/filter`

## Purpose

Define schema and behavior for the Elasticsearch ML filter resource: API usage, identity and import, connection, lifecycle (force-new on `filter_id`), create/read/update/delete flows, and mapping between Terraform state and the Elasticsearch Machine Learning Filters API — including description nullability and set-based items reconciliation.

## Schema

```hcl
resource "elasticstack_elasticsearch_ml_filter" "example" {
  id        = <computed, string>  # internal identifier: <cluster_uuid>/<filter_id>; UseStateForUnknown
  filter_id = <required, string>  # force new; 1–64 chars; lowercase alphanumeric, hyphens, underscores; must start and end with alphanumeric

  description = <optional, string>  # 1–4096 chars; empty string is invalid
  items       = <optional, set(string)>  # up to 10000 elements; wildcard * allowed at start or end

  # Resource-level Elasticsearch connection override (injected by entitycore)
  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    cert_data                = <optional, string>
    key_file                 = <optional, string>
    key_data                 = <optional, string>
    headers                  = <optional, map(string)>
  }
}
```

## Requirements

### Requirement: API — Create (REQ-001)

The resource SHALL call `PUT _ml/filters/<filter_id>` via the typed Elasticsearch client to create a filter.

The description SHALL be included in the request only when the plan value is known and non-empty. The items list SHALL be included only when the plan value is neither null nor unknown.

If the Elasticsearch API returns an error (including a 409 when the filter already exists), the resource SHALL surface the error and leave no state.

#### Scenario: Create with description and items
- GIVEN a plan with a valid `filter_id`, non-empty `description`, and a non-null `items` set
- WHEN Create is called
- THEN `PUT _ml/filters/<filter_id>` is called with the description and items in the request body
- AND the composite `id` is set in state as `<cluster_uuid>/<filter_id>`

#### Scenario: Create when filter already exists
- GIVEN a filter with the same `filter_id` already exists in Elasticsearch
- WHEN Create is called
- THEN the resource SHALL surface a "Failed to create ML filter" error
- AND no state is persisted

#### Scenario: Create with no items
- GIVEN a plan with `items` omitted (null)
- WHEN Create is called
- THEN `PUT _ml/filters/<filter_id>` is called without an items field

### Requirement: API — Read (REQ-002)

The resource SHALL call `GET _ml/filters/<filter_id>` to retrieve the current filter state.

When the response has HTTP status 404, or when the `filters` array in the response is empty, the resource SHALL signal "not found" to the framework (causing the resource to be removed from state). In all other API error cases, the resource SHALL surface the error.

#### Scenario: Read existing filter
- GIVEN a filter exists in Elasticsearch
- WHEN Read is called
- THEN the filter's `filter_id`, `description`, and `items` are mapped into state

#### Scenario: Read missing filter (404)
- GIVEN the filter does not exist (Elasticsearch returns 404)
- WHEN Read is called
- THEN the resource is removed from state with no error

#### Scenario: Read empty filters array
- GIVEN the API returns an empty `filters` array
- WHEN Read is called
- THEN the resource is removed from state with no error

### Requirement: API — Update (REQ-003)

The resource SHALL call `GET _ml/filters/<filter_id>` to fetch the current remote state, compute the item diff, then call `PUT _ml/filters/<filter_id>/_update` with only the changed fields.

Items are managed as a set diff: items in the plan but not on the server are sent as `add_items`; items on the server but not in the plan are sent as `remove_items`. Description is sent in the update request only when its plan value differs from the prior state value.

If the filter is not found during the update fetch (404 or empty array), the resource SHALL surface a "Filter not found" error rather than silently re-creating.

After a successful update, the resource SHALL read the filter back from the API and store the result in state.

#### Scenario: Update adds and removes items
- GIVEN a filter exists with items `["a", "b"]`
- AND the plan specifies items `["b", "c"]`
- WHEN Update is called
- THEN `add_items: ["c"]` and `remove_items: ["a"]` are sent to the update API

#### Scenario: Update description only
- GIVEN the plan has a new description but the same items
- WHEN Update is called
- THEN the description is sent and no item changes are sent

#### Scenario: Update when remote filter is missing
- GIVEN the filter was deleted out-of-band before the update
- AND the refresh step did not detect the deletion (e.g., plan was run with `-refresh=false`)
- WHEN Update is called
- THEN the resource SHALL surface a "Filter not found" error

### Requirement: API — Delete (REQ-004)

The resource SHALL call `DELETE _ml/filters/<filter_id>` to remove the filter.

A 404 response from Elasticsearch during delete SHALL be treated as a success (idempotent delete). Any other API error SHALL be surfaced as "Failed to delete ML filter".

Elasticsearch will reject deletion of a filter that is still referenced by one or more anomaly detection jobs via `custom_rules.scope`; such rejections SHALL be surfaced as errors.

#### Scenario: Delete existing filter
- GIVEN a filter exists in Elasticsearch
- WHEN Destroy is called
- THEN `DELETE _ml/filters/<filter_id>` is called and the resource is removed from state

#### Scenario: Delete when filter is already absent
- GIVEN the filter was deleted out-of-band before Terraform destroy
- WHEN Destroy is called
- THEN the delete is treated as a success (no error)

#### Scenario: Delete blocked by referenced job
- GIVEN one or more anomaly detection jobs reference this filter via `custom_rules.scope`
- WHEN Destroy is called
- THEN Elasticsearch rejects the delete and the resource SHALL surface a "Failed to delete ML filter" error

### Requirement: Identity and Import (REQ-005)

The Terraform resource `id` SHALL be a composite string in the format `<cluster_uuid>/<filter_id>`. The `id` attribute SHALL use `UseStateForUnknown` so it is not shown as unknown during plan.

`filter_id` is the ML API identifier and is used directly as the resource ID segment.

Import SHALL accept an import ID in the format `<cluster_uuid>/<filter_id>`. On import the resource SHALL set both `id` (the full composite string) and `filter_id` (the resource ID segment) in state. Import IDs that are not exactly two `/`-separated segments SHALL produce a "Wrong resource ID" error.

Importing a composite ID whose `filter_id` segment does not exist in Elasticsearch SHALL produce an error.

#### Scenario: Import valid composite ID
- GIVEN a valid import ID `<cluster_uuid>/<filter_id>`
- WHEN `terraform import` is run
- THEN `id` is set to the full composite string and `filter_id` is set to the resource ID segment
- AND Read is called and populates all other attributes from the API

#### Scenario: Import non-composite ID
- GIVEN an import ID that is not in `<uuid>/<filter_id>` format (no `/`, or more than one `/`)
- WHEN `terraform import` is run
- THEN the resource SHALL surface a "Wrong resource ID" error

### Requirement: Lifecycle — Filter ID Replacement (REQ-006)

The `filter_id` attribute SHALL carry `RequiresReplace`, so that any change to `filter_id` triggers destruction of the old filter and creation of a new one.

#### Scenario: Change filter_id
- GIVEN a filter exists with `filter_id = "old-id"`
- AND the configuration is changed to `filter_id = "new-id"`
- WHEN `terraform apply` is run
- THEN the old filter is deleted from Elasticsearch and a new filter with `filter_id = "new-id"` is created

### Requirement: Connection (REQ-007)

The resource SHALL use the `elasticsearch_connection` block, injected by the `entitycore.ElasticsearchResource` envelope, to resolve a scoped Elasticsearch client per resource instance. When the block is absent, the provider-level connection is used.

#### Scenario: Resource-level connection override
- GIVEN an `elasticsearch_connection` block is provided on the resource
- WHEN any CRUD operation is performed
- THEN the scoped connection is used instead of the provider-level connection

### Requirement: Validation — filter_id (REQ-008)

`filter_id` SHALL be validated at plan time:

- Length: 1–64 characters.
- Characters: lowercase alphanumeric, hyphens (`-`), and underscores (`_`).
- First and last character: must be lowercase alphanumeric.

Violations SHALL produce a plan-time error.

#### Scenario: Invalid filter_id — uppercase characters
- GIVEN a configuration with `filter_id = "INVALID_ID"`
- WHEN plan is run
- THEN the resource SHALL surface a validation error matching `lowercase|must contain`

### Requirement: Validation — description (REQ-009)

`description`, when set, SHALL be between 1 and 4096 characters. An empty string value (`""`) is invalid and SHALL produce a plan-time error.

#### Scenario: Description too long
- GIVEN a `description` of more than 4096 characters
- WHEN plan is run
- THEN the resource SHALL surface an "Invalid Attribute Value Length" error referencing the 4096 limit

#### Scenario: Empty description string
- GIVEN a configuration with `description = ""`
- WHEN plan is run
- THEN the resource SHALL surface an "Invalid Attribute Value Length" error indicating the length must be between 1 and 4096

### Requirement: Validation — items (REQ-010)

`items`, when set, SHALL contain at most 10000 elements.

#### Scenario: Items count at the limit
- GIVEN a configuration with exactly 10000 items
- WHEN plan is run
- THEN no validation error is produced

#### Scenario: Items count exceeds the limit
- GIVEN a configuration with more than 10000 items
- WHEN plan is run
- THEN the resource SHALL surface a validation error

### Requirement: Mapping — Description Nullability (REQ-011)

When the Elasticsearch API returns an empty string or a nil description, the resource SHALL map it to a null `description` in Terraform state (not an empty string).

#### Scenario: API returns empty description
- GIVEN the API returns `"description": ""`
- WHEN Read maps the response
- THEN `description` in state is null

#### Scenario: API returns nil description
- GIVEN the API returns no description field
- WHEN Read maps the response
- THEN `description` in state is null

### Requirement: Mapping — Items Set Nullability (REQ-012)

When the Elasticsearch API returns an empty items list, the mapping behavior SHALL depend on the prior Terraform state value:

- If the prior `items` value in state is null (i.e., `items` was omitted in configuration), the mapped state value SHALL remain null.
- If the prior `items` value in state is a non-null set (including an empty set), the mapped state value SHALL be an empty set.

This preserves the user's intent: an omitted `items` block stays absent; an explicit empty set stays as an empty set.

#### Scenario: Empty API items with null prior state
- GIVEN the API returns an empty items list
- AND the prior TF state has `items` as null
- WHEN Read maps the response
- THEN `items` in state is null

#### Scenario: Empty API items with non-null prior state
- GIVEN the API returns an empty items list
- AND the prior TF state has `items` as a non-null (possibly empty) set
- WHEN Read maps the response
- THEN `items` in state is an empty set

### Requirement: Drift Reconciliation (REQ-013)

On each plan/apply cycle, the resource SHALL read the current remote state and reconcile any out-of-band changes back to the desired configuration. Description and items changed outside Terraform SHALL be corrected on the next apply.

#### Scenario: Out-of-band description change
- GIVEN a filter is managed by Terraform with `description = "Original"`
- AND the description is changed out-of-band via the Elasticsearch API to `"Drifted"`
- WHEN `terraform apply` is run
- THEN the description is corrected to `"Original"` in Elasticsearch and in state
