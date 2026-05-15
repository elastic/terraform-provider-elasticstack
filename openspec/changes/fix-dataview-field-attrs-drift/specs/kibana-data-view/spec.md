## MODIFIED Requirements

### Requirement: Lifecycle replacement fields (REQ-006)

**CHANGE**: Remove `data_view.field_attrs` from the replacement list.

Changes to `space_id`, `data_view.id`, or `data_view.allow_no_index` SHALL require resource replacement rather than an in-place update. Changes to `data_view.field_attrs` SHALL NOT trigger resource replacement; they are applied in place via the `UpdateFieldMetadata` API call (see REQ-016).

#### Scenario: Replace on immutable data view id

- GIVEN an existing managed data view
- WHEN `data_view.id` changes in configuration
- THEN Terraform SHALL plan replacement for the resource

#### Scenario: No replacement on field_attrs change

- GIVEN an existing managed data view with a known internal Kibana ID
- WHEN `data_view.field_attrs` changes in configuration (entries added, removed, or modified)
- THEN Terraform SHALL plan an in-place update rather than resource replacement
- AND the data view SHALL retain its internal Kibana ID after the update

### Requirement: Update request mapping and namespace reconciliation (REQ-009)

**CHANGE**: Document that `field_attrs` changes are applied via a separate `UpdateFieldMetadata` call, not via the main data view update body.

On update, the resource SHALL build a Data Views update request from Terraform state using `title`, `name`, `time_field_name`, `source_filters`, `runtime_field_map`, `field_formats`, and `allow_no_index` when those values are set. The Data Views update request SHALL NOT send `override`, `data_view.id`, `data_view.field_attrs`, or `data_view.namespaces`. After a successful Data Views update, the provider SHALL compare prior and planned `data_view.namespaces`; when membership changed, it SHALL call Kibana's Spaces object-sharing API with the computed `spaces_to_add` and `spaces_to_remove` sets for the managed data view id before writing final state. After the namespace reconciliation step, the provider SHALL apply any `field_attrs` delta via a separate `UpdateFieldMetadata` call (see REQ-016).

#### Scenario: Override is create-only

- GIVEN a managed data view whose configuration changes only `override`
- WHEN update runs
- THEN the update request SHALL NOT include `override`

#### Scenario: Namespace update happens in place

- GIVEN an existing managed data view and a plan that adds or removes entries from `data_view.namespaces`
- WHEN update runs successfully
- THEN the provider SHALL keep the same resource identity
- AND SHALL reconcile namespace additions and removals through the Spaces API instead of replacing the resource

#### Scenario: field_attrs update uses separate API call

- GIVEN a managed data view and a plan that changes one or more `field_attrs` entries
- WHEN update runs
- THEN the provider SHALL NOT include `field_attrs` in the main data view update body
- AND SHALL call the `UpdateFieldMetadata` endpoint with a delta payload covering changed and removed fields

## ADDED Requirements

### Requirement: field_attrs semantic equality via custom type (REQ-015)

The `field_attrs` attribute SHALL use a custom map type (`FieldAttrsType` / `FieldAttrsValue`) that implements `MapSemanticEquals`. The semantic equality check SHALL suppress the following categories of plan-time drift without requiring user intervention:

1. **Count-only server entries absent from config**: when the user's `field_attrs` is null or omits a field name entirely, Kibana may auto-populate `count` entries for that field. Such entries — identified by having a null `custom_label` and no user-configured counterpart — SHALL be considered semantically equal to the absent (null) state.
2. **Count growth for user-declared fields**: when a user declares a `field_attrs` entry with `custom_label` set but `count` null, subsequent growth of the server-side `count` for that field SHALL be considered semantically equal.

The custom type SHALL compare `custom_label` values strictly: any change to `custom_label` (addition, removal, or modification) is a real change and SHALL NOT be suppressed.

The `count` attribute SHALL remain `Optional` (not `Computed`). The provider SHALL persist `count` in state only when the user explicitly sets it in configuration.

#### Scenario: Server-generated count entries do not produce a diff

- GIVEN a `elasticstack_kibana_data_view` resource with no `field_attrs` in configuration
- AND Kibana has auto-populated `field_attrs` with `{ "host.hostname": { "count": 5 } }`
- WHEN `terraform plan` runs after apply
- THEN the plan SHALL show no changes

#### Scenario: Configured custom_label is compared strictly

- GIVEN a resource with `field_attrs = { "host.hostname" = { custom_label = "Host" } }` in configuration
- AND the prior state has `field_attrs = { "host.hostname" = { custom_label = "Host", count = 12 } }`
- WHEN `terraform plan` runs
- THEN the plan SHALL show no changes (count growth suppressed because count is null in config)

#### Scenario: custom_label removal is detected

- GIVEN a resource with prior state containing `field_attrs = { "host.hostname" = { custom_label = "Host" } }`
- AND the user removes `host.hostname` from `field_attrs` in configuration
- WHEN `terraform plan` runs
- THEN the plan SHALL show an update (the entry is managed and was removed from config)

#### Scenario: count-only entry in prior state with no custom_label is suppressed

- GIVEN a resource with prior state containing `field_attrs = { "host.hostname" = { count = 5 } }` (server-generated, no custom_label)
- AND the user's configuration has no `field_attrs` (or omits `host.hostname`)
- WHEN `terraform plan` runs
- THEN the plan SHALL show no changes

### Requirement: field_attrs write path via UpdateFieldMetadata (REQ-016)

When `field_attrs` changes between prior state and plan, the provider SHALL apply those changes by calling the Kibana `POST /api/data_views/data_view/{viewId}/fields` endpoint (`UpdateFieldMetadata` wrapper). This call SHALL be made after the main data view update and after namespace reconciliation, within the same Update operation.

The provider SHALL build the delta payload as follows:
- For each field present in the planned `field_attrs` that differs from the prior state: include its full `fieldAttrModel` values in the payload.
- For each field present in the prior state `field_attrs` that is absent from the planned `field_attrs`: include an entry in the payload to clear that field (exact clearing payload shape is an implementation detail).

The `UpdateFieldMetadata` API call SHALL use `kibanautil.SpaceAwarePathRequestEditor` to construct the space-aware URL path, ensuring correct routing for non-default Kibana spaces.

If `UpdateFieldMetadata` returns a transport error or unexpected HTTP status, the provider SHALL surface an error diagnostic and SHALL NOT write final state.

#### Scenario: field_attrs are written via UpdateFieldMetadata on update

- GIVEN a managed data view in space `observability` with a planned `field_attrs` change
- WHEN update runs
- THEN the provider SHALL call `UpdateFieldMetadata` with the space ID `observability` and the delta payload
- AND the main data view update body SHALL NOT contain `field_attrs`

#### Scenario: No UpdateFieldMetadata call when field_attrs unchanged

- GIVEN a managed data view with no `field_attrs` change between state and plan
- WHEN update runs
- THEN the provider SHALL NOT call `UpdateFieldMetadata`

#### Scenario: UpdateFieldMetadata error surfaces as diagnostic

- GIVEN a planned update that includes `field_attrs` changes
- WHEN `UpdateFieldMetadata` returns an error
- THEN the provider SHALL return an error diagnostic
- AND SHALL NOT write updated state
