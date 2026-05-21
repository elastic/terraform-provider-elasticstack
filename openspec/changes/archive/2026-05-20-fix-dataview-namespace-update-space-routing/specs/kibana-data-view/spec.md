## MODIFIED Requirements

### Requirement: API and client error surfacing (REQ-002)

**CHANGE**: Add per-object error surfacing for the Spaces `_update_objects_spaces` endpoint, which returns HTTP 200 even when individual objects fail.

For create, read, update, and delete, when the provider cannot obtain the Kibana OpenAPI client, the operation SHALL return an error diagnostic. For read and update, transport errors and unexpected HTTP statuses SHALL be surfaced as error diagnostics. For create, transport errors and unexpected HTTP statuses SHALL be surfaced as error diagnostics unless the provider can deterministically reconcile a managed data view create under REQ-014. Delete SHALL also surface transport errors and unexpected HTTP statuses, except that delete not-found SHALL be treated as success.

When the Spaces object-sharing API (`_update_objects_spaces`) returns HTTP 200, the provider SHALL additionally parse the response body and inspect per-object results. Any entry in the response where the object-level `error` field is non-nil SHALL be surfaced as an error diagnostic that includes the object id, object type, and the error details. The provider SHALL NOT record a successful update in Terraform state if any per-object error is present.

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

#### Scenario: Namespace update per-object error surfaced

- GIVEN a namespace reconciliation call where `_update_objects_spaces` returns HTTP 200
- AND the response body contains a per-object error for the managed data view (e.g. `{"statusCode":404,"error":"Not Found","message":"Saved object [index-pattern/my-view] not found"}`)
- WHEN the provider handles the response
- THEN the provider SHALL surface an error diagnostic that includes the object id, object type, and the error details
- AND SHALL NOT record a successful update in Terraform state

### Requirement: Update request mapping and namespace reconciliation (REQ-009)

**CHANGE**: Specify that the Spaces API call for namespace reconciliation uses space-aware URL construction so the correct saved object is targeted regardless of which Kibana space the data view lives in.

On update, the resource SHALL build a Data Views update request from Terraform state using `title`, `name`, `time_field_name`, `source_filters`, `runtime_field_map`, `field_formats`, and `allow_no_index` when those values are set. The Data Views update request SHALL NOT send `override`, `data_view.id`, `data_view.field_attrs`, or `data_view.namespaces`. The Data Views update request SHALL always send `source_filters`, `field_formats`, and `runtime_field_map` — defaulting null planned values to an empty collection — so Kibana clears any previously-stored values when the user removes them from configuration. After a successful Data Views update, the provider SHALL compare prior and planned `data_view.namespaces`; when membership changed, it SHALL call Kibana's Spaces object-sharing API with the computed `spaces_to_add` and `spaces_to_remove` sets for the managed data view id before writing final state. The Spaces API call SHALL use space-aware URL construction by passing the resource's `space_id` to `SpaceAwarePathRequestEditor`, which inserts `/s/{spaceID}` before the `/api/` segment when `spaceID` is neither empty nor `"default"`. When either prior or planned `data_view.namespaces` is null or empty, the provider SHALL substitute the resource's own `space_id` for that side of the diff so removing an explicit namespaces list keeps the data view in its own space rather than detaching it from every space. After the namespace reconciliation step, the provider SHALL apply any `field_attrs` delta via a separate `UpdateFieldMetadata` call (see REQ-016).

#### Scenario: Override is create-only

- GIVEN a managed data view whose configuration changes only `override`
- WHEN update runs
- THEN the update request SHALL NOT include `override`

#### Scenario: Namespace update uses space-aware URL for non-default space

- GIVEN an existing managed data view in space `test-space`
- AND a plan that adds `target-space` to `data_view.namespaces`
- WHEN update runs and calls the Spaces object-sharing API
- THEN the HTTP request path SHALL be `/s/test-space/api/spaces/_update_objects_spaces`
- AND the data view SHALL be accessible in `target-space` after the update completes

#### Scenario: Namespace update uses default path for default space

- GIVEN an existing managed data view in the `default` space
- AND a plan that changes `data_view.namespaces`
- WHEN update runs and calls the Spaces object-sharing API
- THEN the HTTP request path SHALL be `/api/spaces/_update_objects_spaces` (no `/s/default/` prefix)

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

#### Scenario: Removed collection fields are cleared in place

- GIVEN a managed data view with stored `source_filters`, `field_formats`, or `runtime_field_map`
- WHEN the plan removes those attributes (planned value becomes null)
- THEN the update request SHALL send each removed collection as an explicit empty value so Kibana clears the prior server-side data
- AND the resulting state SHALL match the planned null value without triggering a "Provider produced inconsistent result after apply" error

#### Scenario: Namespaces removed from configuration retains current space

- GIVEN an existing managed data view shared into multiple namespaces and a plan that removes `data_view.namespaces` (planned value becomes null)
- WHEN update runs
- THEN the provider SHALL substitute `[space_id]` for the planned namespaces when computing the spaces diff
- AND SHALL only remove the data view from namespaces other than its own `space_id`
- AND the data view SHALL remain accessible in its own space after the update completes
