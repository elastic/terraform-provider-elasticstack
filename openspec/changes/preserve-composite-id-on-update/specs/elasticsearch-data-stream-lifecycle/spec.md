## MODIFIED Requirements

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<name>`. The resource SHALL compute `id` by calling `client.ID(ctx, name)`, which combines the current cluster UUID with the configured `name` value, only during create. During update, the resource SHALL preserve the `id` already present in prior state unchanged and SHALL NOT recompute it from the currently connected cluster's UUID. The `id` attribute SHALL use the `UseStateForUnknown` plan modifier so that it is preserved across plan/apply cycles once set, and update SHALL never violate that promise.

#### Scenario: ID computed on create

- GIVEN a new resource with `name = "my-data-stream"`
- WHEN create completes successfully
- THEN `id` SHALL be set to `<cluster_uuid>/my-data-stream`

#### Scenario: Id preserved after update

- GIVEN an existing resource whose stored `id` carries a cluster UUID that no longer matches the UUID of the cluster the provider is currently connected to
- WHEN a non-id attribute (for example `data_retention`) is changed and applied
- THEN the update SHALL succeed
- AND `id` in the resulting state SHALL be unchanged from the value in prior state
