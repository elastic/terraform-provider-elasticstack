## MODIFIED Requirements

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` attribute representing the watch in the format `<cluster_uuid>/<watch_id>`. The resource SHALL compute `id` by combining the current cluster UUID with the configured `watch_id` value only during create. During update, the resource SHALL preserve the `id` already present in prior state unchanged and SHALL NOT recompute it from the currently connected cluster's UUID.

#### Scenario: ID set on create

- GIVEN a successful Put Watch API call during create
- WHEN create completes
- THEN the `id` in state SHALL be in the format `<cluster_uuid>/<watch_id>`

#### Scenario: Id preserved after update

- GIVEN an existing watch whose stored `id` carries a cluster UUID that no longer matches the UUID of the cluster the provider is currently connected to
- WHEN a non-id attribute of the watch is changed and applied
- THEN the update SHALL succeed
- AND `id` in the resulting state SHALL be unchanged from the value in prior state
