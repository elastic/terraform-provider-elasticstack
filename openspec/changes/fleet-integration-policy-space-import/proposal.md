## Why

`elasticstack_fleet_integration_policy` cannot be imported when the package policy lives in a
non-default Kibana space. `ImportState` currently calls
`resource.ImportStatePassthroughID`, which stores the raw import argument directly in
`policy_id`. The subsequent Read then queries
`GET /api/fleet/package_policies/{policyId}` (the default-space endpoint), which returns 404
for any policy that exists only inside a named space.

`elasticstack_fleet_agent_policy` already solved the identical problem: its `ImportState`
uses `clients.CompositeIDFromStrFw` to parse a `<space_id>/<policy_id>` composite ID, sets
`space_ids` to `[<space_id>]`, and the Read path picks up the correct space context via
`GetOperationalSpaceFromState`. This change brings `elasticstack_fleet_integration_policy`
into parity.

## What Changes

- Replace `resource.ImportStatePassthroughID` in
  `internal/fleet/integration_policy/resource.go` with a custom `ImportState` that parses
  the composite ID.
- When the import ID contains a `/` separator recognised by `clients.CompositeIDFromStrFw`,
  `policy_id` is set to the resource-ID segment and `space_ids` is set to a single-element
  set containing the space-ID segment.
- When the import ID is a plain string (no composite separator), the full string is placed in
  `policy_id` and `space_ids` is left unset — preserving existing behaviour for default-space
  imports.
- Update the OpenSpec `fleet-integration-policy` requirements spec to replace REQ-006 with
  the corrected import behaviour.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `fleet-integration-policy`: import now accepts `<space_id>/<policy_id>` as well as a
  plain `<policy_id>`. The user-visible contract for plain IDs is preserved unchanged.

## Impact

- Single file changed in the provider: `internal/fleet/integration_policy/resource.go`
  (`ImportState` method, ~15 lines).
- No schema changes, no state upgrade, no API client changes.
- Existing import invocations using a bare policy ID continue to work without modification.
- Acceptance test additions are needed to cover the composite-ID import path.
