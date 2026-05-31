## Why

`elasticstack_fleet_elastic_defend_integration_policy` only accepts a single agent policy via
`agent_policy_id` (Required string). The generic `elasticstack_fleet_integration_policy` already
supports attaching one integration policy to multiple agent policies via `agent_policy_ids`
(Optional list), with `agent_policy_id` changed to Optional. The dedicated Elastic Defend resource
lacks this parity, making it impossible to share one Elastic Defend policy across several agent
policies with the dedicated resource without resorting to `for_each` (which produces N independent
package policies rather than one shared policy, increasing object count and the surface for drift).

## What Changes

Add `agent_policy_ids` (Optional `list(string)`) to
`elasticstack_fleet_elastic_defend_integration_policy`, changing `agent_policy_id` from Required
to Optional, exactly mirroring the pattern already in production for
`elasticstack_fleet_integration_policy`.

- `agent_policy_id` — changes from Required to Optional (non-breaking: existing configs unchanged)
- `agent_policy_ids` — new Optional `list(string)`, `SizeAtLeast(1)`, `ConflictsWith(agent_policy_id)`;
  runtime-gated on Elastic Stack ≥ 8.15.0

No schema version bump or state upgrader is needed. The Plugin Framework treats the absent
`agent_policy_ids` key in existing state as null — valid for Optional-only attributes. This is
confirmed by commit `df995c0d` in this repo, which added `agent_policy_ids` to the live V1 schema
of `elasticstack_fleet_integration_policy` without a version bump.

The underlying kbapi already carries `PolicyIds *[]string` in both request and response structs;
only the Terraform layer needs updating.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `fleet-elastic-defend-integration-policy`: `agent_policy_id` becomes Optional; new
  `agent_policy_ids` list attribute added; both request phases (bootstrap and finalize) updated
  to populate `PolicyIds` (and `PolicyId` as first-element for compatibility); read logic updated
  to populate whichever field is in state.

## Impact

- Files changed: `schema.go`, `models.go`, `request.go`, `mapping.go` (all under
  `internal/fleet/elastic_defend_integration_policy/`); and the main spec at
  `openspec/specs/fleet-elastic-defend-integration-policy/spec.md`.
- No generated-client changes needed.
- No schema version bump or state upgrader.
- Fully non-breaking: existing configs using `agent_policy_id = "..."` continue to work.
- Acceptance tests are needed to cover the `agent_policy_ids` path and the version gate.
