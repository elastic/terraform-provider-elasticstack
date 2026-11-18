## MODIFIED Requirements

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<name>`. The resource SHALL compute `id` by calling `client.ID(ctx, name)`, which combines the current cluster UUID with the configured `name` value, only during create. During update, the resource SHALL NOT recompute `id` from the currently connected cluster's UUID. When the configured `name` is unchanged, the resource SHALL preserve the `id` already present in prior state unchanged, and `UseStateForUnknown` on `id` SHALL apply so the prior value is planned as known. When `name` changes in place, the planned `id` SHALL be unknown and apply SHALL write `<prior-uuid>/<new-name>`: the cluster UUID prefix from prior state is kept and the name segment is replaced with the new `name`.

#### Scenario: ID computed on create

- GIVEN a new resource with `name = "my-data-stream"`
- WHEN create completes successfully
- THEN `id` SHALL be set to `<cluster_uuid>/my-data-stream`

#### Scenario: Id preserved after update

- GIVEN an existing resource whose stored `id` carries a cluster UUID that no longer matches the UUID of the cluster the provider is currently connected to
- WHEN a non-id attribute (for example `data_retention`) is changed and applied
- THEN the update SHALL succeed
- AND `id` in the resulting state SHALL be unchanged from the value in prior state

#### Scenario: Name segment adopted on in-place name change

- GIVEN an existing resource whose stored `id` is `<stale_uuid>/old-name`
- WHEN `name` is changed to `new-name`
- THEN the plan SHALL treat `id` as unknown
- AND when applied, the update SHALL succeed
- AND `id` in the resulting state SHALL be `<stale_uuid>/new-name`

### Requirement: Create and update (REQ-013–REQ-015)

On create, the resource SHALL read the plan model, resolve the Elasticsearch client, compute the composite `id` from the live cluster UUID and configured `name`, convert the plan to a `models.LifecycleSettings` struct, and call `PutDataStreamLifecycle`. On update, the resource SHALL reuse the cluster UUID prefix from prior state's `id` rather than recomputing it from the currently connected cluster: if `name` is unchanged the prior `id` is preserved in full, and if `name` changes in place the name segment of `id` is replaced with the new `name`. The `expand_wildcards` value SHALL be forwarded as the `WithExpandWildcards` option on the Put request. After a successful Put, the resource SHALL perform a read and store the result in state. If any step (client resolution, id computation, API call, or read-back) returns an error, the resource SHALL return the error diagnostic and SHALL not finalize the state.

#### Scenario: Successful create

- GIVEN a valid plan with `name`, `data_retention`, and `enabled`
- WHEN create runs and Put succeeds
- THEN the resource SHALL call read and store the refreshed model in state
