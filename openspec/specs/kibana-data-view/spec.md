# `elasticstack_kibana_data_view` — Schema and Functional Requirements

Resource implementation: `internal/kibana/dataview`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_data_view` resource, including Kibana Data Views API usage, composite identity and import, provider-level Kibana OpenAPI client usage, replacement-vs-update behavior, and stable mapping between Terraform state and Kibana responses for nested data view fields.

## Schema

```hcl
resource "elasticstack_kibana_data_view" "example" {
  id       = <computed, string> # canonical state id: "<space_id>/<data_view_id>"; UseStateForUnknown
  space_id = <optional, computed, string> # default "default"; RequiresReplace
  override = <optional, computed, bool>   # default false; used on create requests only

  data_view = {
    title           = <required, string> # minimum length 1
    name            = <optional, computed, string> # UseStateForUnknown
    id              = <optional, computed, string> # saved object id; RequiresReplace; UseStateForUnknown
    time_field_name = <optional, computed, string> # UseStateForUnknown

    source_filters = <optional, list(string)>

    field_attrs = <optional, map(object({
      custom_label = <optional, string>
      count        = <optional, int64>
    }))> # RequiresReplace

    runtime_field_map = <optional, map(object({
      type          = <required, string>
      script_source = <required, string>
    }))>

    field_formats = <optional, map(object({
      id     = <required, string>
      params = <optional, computed, object({
        pattern                    = <optional, string>
        urltemplate                = <optional, string>
        labeltemplate              = <optional, string>
        input_format               = <optional, string>
        output_format              = <optional, string>
        output_precision           = <optional, int64>
        include_space_with_suffix  = <optional, bool>
        use_short_suffix           = <optional, bool>
        timezone                   = <optional, string>
        field_type                 = <optional, string>
        colors = <optional, list(object({
          range      = <optional, string>
          regex      = <optional, string>
          text       = <optional, string>
          background = <optional, string>
        }))>
        field_length      = <optional, int64>
        transform         = <optional, string>
        lookup_entries = <optional, list(object({
          key   = <required, string>
          value = <required, string>
        }))>
        unknown_key_value = <optional, string>
        type              = <optional, string>
        width             = <optional, int64>
        height            = <optional, int64>
      })>
    }))>

    allow_no_index = <optional, computed, bool> # default false; RequiresReplace
    namespaces     = <optional, list(string)>   # RequiresReplace
  }
}
```

Notes:

- The resource uses provider-level Kibana OpenAPI client configuration only; there is no resource-level Kibana connection override block.
- This resource does not declare a schema version, custom state upgrader, custom config validator, or custom plan modifier beyond the schema-level defaults and plan modifiers listed above.

## Requirements

### Requirement: Kibana Data Views APIs (REQ-001)

The resource SHALL manage data views through Kibana's Data Views HTTP APIs: create, get, update, and delete ([Kibana data views API docs](https://www.elastic.co/guide/en/kibana/current/data-views-api.html)).

#### Scenario: CRUD uses Data Views APIs

- GIVEN a managed Kibana data view
- WHEN create, read, update, or delete runs
- THEN the provider SHALL use the corresponding Kibana Data Views API operation

### Requirement: API and client error surfacing (REQ-002)

For create, read, update, and delete, when the provider cannot obtain the Kibana OpenAPI client, the operation SHALL return an error diagnostic. For read and update, transport errors and unexpected HTTP statuses SHALL be surfaced as error diagnostics. For create, transport errors and unexpected HTTP statuses SHALL be surfaced as error diagnostics unless the provider can deterministically reconcile a managed data view create under REQ-014. Delete SHALL also surface transport errors and unexpected HTTP statuses, except that delete not-found SHALL be treated as success.

#### Scenario: Missing Kibana OpenAPI client

- GIVEN the resource cannot obtain a Kibana OpenAPI client from provider configuration
- WHEN any CRUD operation runs
- THEN the operation SHALL fail with an error diagnostic

#### Scenario: Delete not found

- GIVEN a delete request for a data view that is already absent
- WHEN Kibana returns HTTP 404
- THEN the provider SHALL treat the delete as successful

#### Scenario: Create error without deterministic reconciliation

- GIVEN a create request that does not meet the managed reconciliation conditions in REQ-014
- WHEN Kibana returns a transport error or unexpected HTTP status for create
- THEN the provider SHALL surface an error diagnostic and SHALL NOT record Terraform state for the resource

### Requirement: Identity and canonical `id` (REQ-003)

The resource SHALL store a canonical computed `id` in the format `<space_id>/<data_view_id>`. After create, read, or update receives a data view from Kibana, the provider SHALL set `id` from the current `space_id` in state and the data view id returned by the API.

#### Scenario: Canonical composite id in state

- GIVEN a successful read of a data view in space `default` with Kibana id `logs-view`
- WHEN Terraform state is written
- THEN `id` SHALL be `default/logs-view`

### Requirement: Import format and imported state seeding (REQ-004)

The resource SHALL support Terraform import using an id of the form `<space_id>/<data_view_id>` with exactly one `/`. On successful import, it SHALL set `id` to the full import string, set `space_id` from the first segment, set `override` to `false`, and seed `data_view` as unknown so a subsequent read can populate it. If the import id is not in that format, the provider SHALL return an error diagnostic describing the required composite format.

#### Scenario: Valid import

- GIVEN an import id `observability/my-view`
- WHEN import runs
- THEN state SHALL hold `id = "observability/my-view"`, `space_id = "observability"`, `override = false`, and unknown `data_view`

#### Scenario: Invalid import id

- GIVEN an import id without exactly one `/`
- WHEN import runs
- THEN the provider SHALL return an error diagnostic for the required `<space_id>/<data_view_id>` format

### Requirement: Provider-level Kibana client only (REQ-005)

The resource SHALL use the provider's configured Kibana OpenAPI client for create, read, update, and delete. The resource SHALL NOT support a resource-local connection override in its schema or request path.

#### Scenario: Standard provider connection

- GIVEN the provider is configured with Kibana access
- WHEN the resource performs CRUD
- THEN all API operations SHALL use that provider-level Kibana OpenAPI client

### Requirement: Lifecycle replacement fields (REQ-006)

Changes to `space_id`, `data_view.id`, `data_view.field_attrs`, `data_view.allow_no_index`, or `data_view.namespaces` SHALL require resource replacement rather than an in-place update.

#### Scenario: Replace on immutable data view id

- GIVEN an existing managed data view
- WHEN `data_view.id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

### Requirement: Defaults, validators, and unknown preservation (REQ-007)

If `space_id` is omitted, the resource SHALL default it to `default`. If `override` is omitted, the resource SHALL default it to `false`. If `data_view.allow_no_index` is omitted, the resource SHALL default it to `false`. The resource SHALL reject an empty `data_view.title`. For `id`, `data_view.name`, `data_view.id`, and `data_view.time_field_name`, unknown planned values SHALL preserve prior state via `UseStateForUnknown`.

#### Scenario: Minimal configuration

- GIVEN configuration that sets only `data_view.title`
- WHEN Terraform plans the resource
- THEN `space_id` SHALL default to `default`, `override` SHALL default to `false`, and `data_view.allow_no_index` SHALL default to `false`

### Requirement: Create request mapping (REQ-008)

On create, the resource SHALL build a create request from Terraform state and send `override` only on the create request. When `data_view.namespaces` is configured, the provider SHALL append the current `space_id` to that list if it is not already present before sending the create request. The create request SHALL map `title`, `name`, `id`, `time_field_name`, `source_filters`, `field_attrs`, `runtime_field_map`, `field_formats`, and `allow_no_index` from the Terraform model when those values are set.

#### Scenario: Create injects current space into namespaces

- GIVEN `space_id = "default"` and `data_view.namespaces = ["backend", "o11y"]`
- WHEN create builds the API request
- THEN the request namespaces SHALL include `"backend"`, `"o11y"`, and `"default"`

### Requirement: Update request mapping (REQ-009)

On update, the resource SHALL build an update request from Terraform state using `title`, `name`, `time_field_name`, `source_filters`, `runtime_field_map`, `field_formats`, and `allow_no_index` when those values are set. The update request SHALL NOT send `override`, `data_view.id`, `data_view.field_attrs`, or `data_view.namespaces`.

#### Scenario: Override is create-only

- GIVEN a managed data view whose configuration changes only `override`
- WHEN update runs
- THEN the update request SHALL NOT include `override`

### Requirement: Read behavior and missing resource handling (REQ-010)

When refreshing state, the resource SHALL determine the target data view id and space by parsing the composite `id` when present; if `id` is not a valid composite id, it SHALL fall back to the bare `id` value plus `space_id` from state. If the get request returns not found, the resource SHALL remove itself from Terraform state. Otherwise it SHALL repopulate state from the API response.

#### Scenario: Read of deleted data view

- GIVEN a resource recorded in Terraform state
- WHEN read calls Kibana and receives not found
- THEN the provider SHALL remove the resource from state

### Requirement: State mapping for empty collections (REQ-011)

When mapping API responses back to Terraform state, empty `source_filters`, `field_attrs`, `runtime_field_map`, and `field_formats` returned by Kibana SHALL preserve a prior null value instead of forcing an empty list or map into state. If a field format entry has no `params`, the resource SHALL store `params` as a null object in state.

#### Scenario: Empty API collection preserves null

- GIVEN prior state where `data_view.source_filters` is null
- WHEN Kibana returns an empty `source_filters` collection
- THEN the provider SHALL keep `data_view.source_filters` null in state

### Requirement: Namespace state normalization (REQ-012)

The resource SHALL normalize `data_view.namespaces` to avoid spurious drift. If Kibana omits `namespaces` and the prior Terraform value is null, the resource SHALL keep it null. If Kibana omits `namespaces` and prior state had a value, the resource SHALL preserve the prior state value. If prior state is null and Kibana returns exactly the current `space_id` as the only namespace, the resource SHALL keep `namespaces` null. If the response contains the same namespace membership as prior state, allowing for Kibana to add the current `space_id` and to return namespaces sorted by name, the resource SHALL preserve the prior state ordering and value. Otherwise it SHALL store the response value.

#### Scenario: Kibana returns only current space

- GIVEN prior state where `data_view.namespaces` is null and `space_id` is `default`
- WHEN Kibana returns `["default"]`
- THEN the provider SHALL keep `data_view.namespaces` null in state

#### Scenario: Kibana adds current space to shared namespaces

- GIVEN prior state `data_view.namespaces = ["ns1", "ns2"]` and `space_id = "test"`
- WHEN Kibana returns `["test", "ns1", "ns2"]`
- THEN the provider SHALL preserve the prior state value `["ns1", "ns2"]`

### Requirement: Nested object mapping (REQ-013)

The resource SHALL map `field_attrs`, `runtime_field_map`, and `field_formats` between Terraform and Kibana as structured objects keyed by field name. For runtime fields, the provider SHALL map `type` and `script_source`. For field formats, it SHALL map the format `id` plus any configured `params` fields, including color rules, static lookup entries, URL parameters, duration parameters, truncate length, transform, timezone, and width/height values.

#### Scenario: Runtime field round-trip

- GIVEN a runtime field entry with `type = "keyword"` and a `script_source`
- WHEN create, update, and read reconcile the object
- THEN the provider SHALL map those values between Terraform state and Kibana's runtime field structure

### Requirement: Managed create reconciliation after an error response (REQ-014)

When a create request supplies an explicit `data_view.id`, the provider SHALL treat that identifier as the managed identity for create reconciliation. If Kibana persists the create request but returns an error or unexpected HTTP status to the provider, the provider SHALL perform a follow-up read of that same data view id in the target `space_id`. If the read succeeds, the provider SHALL populate Terraform state from the read result and complete create successfully. If the read fails or the data view is not found, the provider SHALL surface the original create failure and SHALL NOT write state.

#### Scenario: Managed create succeeds server-side but returns an error response

- GIVEN configuration sets an explicit `data_view.id` and target `space_id`
- AND Kibana persists the data view create request
- AND Kibana returns an error or unexpected HTTP status for the create call
- WHEN the provider handles the create result
- THEN the provider SHALL read the data view by that configured id in the same space
- AND SHALL populate Terraform state from the read result
- AND SHALL complete create without leaving the resource unmanaged

#### Scenario: Managed create error cannot be reconciled

- GIVEN configuration sets an explicit `data_view.id`
- AND Kibana returns an error or unexpected HTTP status for the create call
- AND a follow-up read by that id does not return the created data view
- WHEN the provider handles the create result
- THEN the provider SHALL surface the original create failure
- AND SHALL NOT write Terraform state for the resource

#### Scenario: Create without explicit managed id

- GIVEN configuration does not set `data_view.id`
- WHEN Kibana returns an error or unexpected HTTP status for the create call
- THEN the provider SHALL NOT attempt heuristic reconciliation by title or other mutable fields under REQ-014
- AND SHALL surface the create failure as an error diagnostic

## Traceability

| Area | Primary files |
|------|----------------|
| Schema | `internal/kibana/dataview/schema.go` |
| Metadata / Configure / Import | `internal/kibana/dataview/resource.go` |
| CRUD orchestration | `internal/kibana/dataview/create.go`, `internal/kibana/dataview/read.go`, `internal/kibana/dataview/update.go`, `internal/kibana/dataview/delete.go` |
| Model mapping / id parsing / namespace normalization | `internal/kibana/dataview/models.go` |
| API status handling | `internal/clients/kibanaoapi/data_views.go` |
| Composite id parsing | `internal/clients/api_client.go` |
