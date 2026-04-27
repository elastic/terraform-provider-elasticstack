# Design: Fleet Integration Policy — Space-Aware Import

## Context

`elasticstack_fleet_integration_policy` maps to the Kibana Fleet package-policy API. All
Fleet package-policy endpoints support an optional Kibana-space prefix:

- Default space: `GET /api/fleet/package_policies/{packagePolicyId}`
- Named space:   `GET /s/{spaceId}/api/fleet/package_policies/{packagePolicyId}`

The `Read` operation already derives its space context from `space_ids` in state via
`fleetutils.GetOperationalSpaceFromState`. The problem is that `ImportState` never sets
`space_ids`; it only sets `policy_id` via `ImportStatePassthroughID`. After import the Read
queries the default-space endpoint, returning 404 for any policy created in a named space.

`elasticstack_fleet_agent_policy` (`internal/fleet/agentpolicy/resource.go`) resolves the
same problem with this pattern:

```go
func (r *agentPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    var spaceID string
    var policyID string

    compID, diags := clients.CompositeIDFromStrFw(req.ID)
    if diags.HasError() {
        policyID = req.ID
    } else {
        spaceID = compID.ClusterID
        policyID = compID.ResourceID
    }

    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), policyID)...)

    if spaceID != "" {
        resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_ids"), []string{spaceID})...)
    }
}
```

`clients.CompositeIDFromStrFw` splits on the `/` separator used by the provider's existing
composite-ID convention (`<clusterID>/<resourceID>` — here repurposed as
`<spaceID>/<policyID>`). It returns an error when no `/` is found, so the plain-ID path
falls through to the `policyID = req.ID` assignment, which preserves backward compatibility.

## Changes Required

### `internal/fleet/integration_policy/resource.go`

Replace:

```go
func (r *integrationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("policy_id"), req, resp)
}
```

With:

```go
func (r *integrationPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    var spaceID string
    var policyID string

    compID, diags := clients.CompositeIDFromStrFw(req.ID)
    if diags.HasError() {
        policyID = req.ID
    } else {
        spaceID = compID.ClusterID
        policyID = compID.ResourceID
    }

    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policy_id"), policyID)...)

    if spaceID != "" {
        resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("space_ids"), []string{spaceID})...)
    }
}
```

Required additional import: `"github.com/elastic/terraform-provider-elasticstack/internal/clients"`

The existing `"github.com/hashicorp/terraform-plugin-framework/path"` import is already
present. The `resource` package import is already present.

### Spec update

`openspec/specs/fleet-integration-policy/spec.md` REQ-006 must be replaced with the
corrected import behaviour (composite-ID or plain-ID).

### Acceptance test

`internal/fleet/integration_policy/acc_test.go` must gain a test step (or new test
function) that exercises:

1. Create a policy in a named space.
2. Remove it from state.
3. `terraform import` with `<space_id>/<policy_id>`.
4. Assert that `policy_id` and `space_ids` are populated correctly after the subsequent
   refresh.

## Behavioural Contract

| Import ID format | `policy_id` | `space_ids` |
|---|---|---|
| `"my-space/abc-123"` | `"abc-123"` | `["my-space"]` |
| `"abc-123"` (no `/`) | `"abc-123"` | not set (null) |

## Assumptions & Open Questions

- **UUID-like policy IDs do not contain `/`**: Fleet generates UUIDs for policy IDs. A
  plain UUID will never be mis-parsed as a composite ID. ✓ safe assumption.
- **Single-space import only**: The composite-ID format carries at most one space. If a
  policy belongs to multiple spaces, the user should import with any one space and then
  reconcile `space_ids` in configuration. This matches the agent-policy precedent.
- **Space availability at import time**: The named space must exist in Kibana when the
  import is run. If it does not, Fleet returns 404 and Terraform surfaces an appropriate
  error through the normal read path.
