## Context

The `elasticstack_fleet_server_host` resource uses the Terraform Plugin Framework. Its `host_id` attribute is declared `Optional+Computed` so practitioners may omit it and let Fleet assign a UUID on create. The Plugin Framework only propagates prior-state values into an update plan automatically when a `UseStateForUnknown()` plan modifier is attached. Without it, a plan where the user has not configured `host_id` produces `null` in the plan value for `host_id`. The `Update` handler then reads this null as an empty string and constructs a malformed URL (`PUT /api/fleet/fleet_server_hosts/` with no ID), to which Kibana responds 404.

All other fleet resources in this provider that have an `Optional+Computed` ID attribute already carry the correct plan modifiers:

| Resource | ID attribute | `UseStateForUnknown` | `RequiresReplace` |
|---|---|---|---|
| `fleet_output` | `output_id` | ✅ | ✅ |
| `fleet_agent_policy` | `policy_id` | ✅ | ✅ |
| `fleet_agent_download_source` | `source_id` | ✅ | ✅ |
| `fleet_proxy` | `proxy_id` | ✅ | ✅ |
| `elastic_defend_integration_policy` | `policy_id` | ✅ | ✅ |
| **`fleet_server_host`** | **`host_id`** | ❌ | ❌ |

`fleet_server_host.host_id` is the only outlier.

A secondary gap in the acceptance test suite means this bug was not caught by CI: `TestAccResourceFleetServerHost_computedID` only performs a CREATE assertion; no UPDATE step exercises the `Update` handler with a computed `host_id`.

## Goals

- Fix the 404 on update by adding the two missing plan modifiers to `host_id`.
- Guard the fix with an acceptance test UPDATE step.

## Non-Goals

- Changes to other fleet resources (they are already correct).
- Backporting to maintenance branches (forward-only fix; the secondary delete-of-default bug was already fixed in v0.15.0).
- Fixing a similar issue in `elasticstack_fleet_output` for default outputs (separate issue).

## Decisions

### Decision 1: Approach A (schema plan modifiers) over Approach B (read host_id from state in Update)

Adding `UseStateForUnknown()` and `RequiresReplace()` to the schema is the idiomatic Plugin Framework fix and requires changes to only `schema.go`. Approach B — reading `host_id` from state inside `Update` — would also fix the 404 but would leave the schema inconsistency in place and diverge from the pattern used by every other fleet resource. Approach A is chosen.

### Decision 2: `RequiresReplace()` included alongside `UseStateForUnknown()`

`RequiresReplace()` ensures that if a user explicitly provides a new `host_id` value that differs from the one in state, the resource is destroyed and recreated rather than attempting an update against the wrong ID. This matches the pattern on all other fleet ID attributes. The behavior change is strictly better: previously, such a change would produce a likely-to-fail update; now it produces a safe destroy/recreate.

### Decision 3: UPDATE step added to existing test, not a new test function

The existing `TestAccResourceFleetServerHost_computedID` already creates a host with a computed `host_id`. Appending an UPDATE step (e.g. change `name` or `hosts`) to that same test function is the minimal addition that guards the fixed code path without duplicating the full create/destroy lifecycle. A new test function is not needed.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| `RequiresReplace()` is a behavior change for users who explicitly set `host_id` and then change it | The previous behavior was a broken update returning 404; destroy-and-recreate is strictly better. No migration note needed. |
| UPDATE step in acceptance test requires a real Fleet-enabled Kibana | The test already has this requirement; the new step inherits the same skip conditions. |

## Open Questions

- **Backport eligibility**: Should the fix be backported to v0.11.x–v0.14.x maintenance branches? The `Update` 404 is observable in those versions; the delete-of-default fix was already backported (v0.15.0). Maintainers may prefer a forward-only fix. No action required during implementation.
- **CHANGELOG entry wording**: The `RequiresReplace()` behavior change for explicit `host_id` changes is technically a breaking change for a small set of users (those who set `host_id` explicitly and then change it). Maintainers should decide whether to note it in the CHANGELOG or rely on the "broken → working" framing.
