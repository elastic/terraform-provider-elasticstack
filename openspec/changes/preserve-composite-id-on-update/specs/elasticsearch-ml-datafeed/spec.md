## MODIFIED Requirements

### Requirement: Identity and import (REQ-007–REQ-009)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<datafeed_id>`. During create, the resource SHALL derive `id` by calling `r.client.ID(ctx, datafeedID)` to obtain the cluster UUID and `datafeed_id`, and SHALL set `id` in state after a successful API call. During update, the resource SHALL preserve the `id` already present in prior state unchanged and SHALL NOT call `r.client.ID` or otherwise recompute `id` from the currently connected cluster's UUID. The resource SHALL support import by accepting an `id` in the format `<cluster_uuid>/<datafeed_id>`, parsing it with `clients.CompositeIDFromStr`, and persisting both `id` and `datafeed_id` to state. When the import `id` format is invalid, the resource SHALL return an error diagnostic.

#### Scenario: Import with valid composite id

- GIVEN import with a valid `<cluster_uuid>/<datafeed_id>` id
- WHEN import completes
- THEN `id` and `datafeed_id` SHALL be stored in state and read SHALL populate all remaining attributes

#### Scenario: Import with invalid id format

- GIVEN import with an id that is not in `<cluster_uuid>/<datafeed_id>` format
- WHEN import runs
- THEN the resource SHALL return an error diagnostic

#### Scenario: Id preserved after update

- GIVEN an existing datafeed whose stored `id` carries a cluster UUID that no longer matches the UUID of the cluster the provider is currently connected to
- WHEN a non-id attribute of the datafeed is changed and applied
- THEN the update SHALL succeed
- AND `id` in the resulting state SHALL be unchanged from the value in prior state
