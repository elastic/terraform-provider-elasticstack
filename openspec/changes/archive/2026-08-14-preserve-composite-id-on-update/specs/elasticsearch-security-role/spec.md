## MODIFIED Requirements

### Requirement: Identity (REQ-005–REQ-006)

The resource SHALL expose a computed `id` representing a composite identifier in the format `<cluster_uuid>/<role_name>`. The resource SHALL compute `id` from the current cluster UUID and the configured role name only during create. During update, the resource SHALL preserve the `id` already present in prior state unchanged and SHALL NOT recompute it from the currently connected cluster's UUID, even if that UUID differs from the one encoded in the stored `id`. The `id` attribute's `UseStateForUnknown` plan modifier SHALL therefore never be violated by an apply.

#### Scenario: Computed id after apply

- GIVEN a successful create
- WHEN state is written
- THEN `id` SHALL equal `<cluster_uuid>/<role_name>` for the target cluster and configured name

#### Scenario: Id preserved after update

- GIVEN an existing role whose stored `id` carries a cluster UUID that no longer matches the UUID of the cluster the provider is currently connected to
- WHEN a non-id attribute of the role is changed and applied
- THEN the update SHALL succeed
- AND `id` in the resulting state SHALL be unchanged from the value in prior state
