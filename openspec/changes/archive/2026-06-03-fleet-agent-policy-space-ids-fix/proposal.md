## Why

The `elasticstack_fleet_agent_policy` resource supports a `space_ids` attribute (available
since Kibana 9.1.0) that assigns the policy to a named Kibana space. However, after a
successful `terraform apply`, Terraform raises:

> Provider produced inconsistent result after apply: .space_ids: was
> cty.SetVal([]cty.Value{cty.StringVal("example_id")}), but now null.

The root cause is in `populateFromAPI` (`internal/fleet/agentpolicy/models.go`). The Fleet API may omit
`space_ids` from its response body; when the field is absent, `kbapi.AgentPolicy.SpaceIds` is
unmarshaled as `nil`. The current code interprets the absent field as "space_ids is null"
and overwrites the model with `types.SetNull`, contradicting the planned value and triggering
the consistency error.
triggering the consistency error.

The existing comment at `create.go:76` already acknowledges that "space_ids can be null in
the response even when specified in the request", but the follow-up GET still has the same
omission. The bug therefore persists on all three paths that call `populateFromAPI`: Create,
Read, and Update.

## What Changes

Apply the same null-preservation pattern already used for optional string fields
(`DataOutputId`, `FleetServerHostId`, `DownloadSourceId`) via the existing `preserveNullStr`
closure. In `populateFromAPI`, when the API returns `nil` for `SpaceIds`, keep the current
model value rather than overwriting it with null. This fixes the inconsistency error on all
three CRUD paths at once without any schema change.

The fix is a < 10 line change in `internal/fleet/agentpolicy/models.go` limited to the
`space_ids` block at lines 211–219.

## Capabilities

### Modified Capabilities

- `elasticstack-fleet-agent-policy`: The `space_ids` attribute now correctly retains its
  configured value after apply when the Fleet API omits the field from its response body
  (due to `omitempty` on the generated kbapi struct).

## Impact

- **Changed code**: `internal/fleet/agentpolicy/models.go` — 6–10 lines in the
  `populateFromAPI` method.
- **Tests**: The existing acceptance test `TestAccResourceAgentPolicyWithSpaceIDs` in
  `internal/fleet/agentpolicy/acc_test.go` exercises this path; a unit test for
  `populateFromAPI` should be added or updated to cover the API-omits-nil case.
- **No schema change**: `space_ids` attribute definition is unchanged.
- **Backward compatibility**: Additive fix only. Users who previously worked around the
  bug by omitting `space_ids` are unaffected; users who set `space_ids` now get correct
  behaviour.
