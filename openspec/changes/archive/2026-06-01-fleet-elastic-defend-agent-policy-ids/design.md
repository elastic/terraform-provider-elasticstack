# Design: Elastic Defend Integration Policy — Multiple Agent Policy IDs

## Context

`elasticstack_fleet_elastic_defend_integration_policy` uses a two-phase create (bootstrap then
finalize) and talks to the Kibana Fleet package policy API. The kbapi-generated structs already
carry both `PolicyId` (deprecated) and `PolicyIds` on request and response:

- `PackagePolicyRequestTypedInputs.PolicyId` — annotated "Deprecated. Use policy_ids instead."
- `PackagePolicyRequestTypedInputs.PolicyIds *[]string` — primary multi-policy field
- `PackagePolicy.PolicyIds *[]string` — present in the read response

The generic resource (`internal/fleet/integration_policy/`) already implements the side-by-side
pattern. The reference implementation lives in:

- `internal/fleet/integration_policy/schema.go` — `attrAgentPolicyID` (Optional) + `attrAgentPolicyIDs` (Optional list)
- `internal/fleet/integration_policy/models.go:77–103` — read logic that populates only the field originally in state
- `internal/fleet/integration_policy/resource.go:43` — `MinVersionPolicyIDs = version.Must(version.NewVersion("8.15.0"))`
- `internal/fleet/integration_policy/capabilities.go` — `resolveIntegrationPolicyFeatures` with `EnforceMinVersion`

## Schema Changes

### `internal/fleet/elastic_defend_integration_policy/schema.go`

Change `agent_policy_id` from Required to Optional and add `agent_policy_ids`:

```go
"agent_policy_id": schema.StringAttribute{
    Description: "ID of the agent policy. Conflicts with agent_policy_ids.",
    Optional:    true,
    Validators: []validator.String{
        stringvalidator.ConflictsWith(path.Root("agent_policy_ids").Expression()),
    },
},
"agent_policy_ids": schema.ListAttribute{
    Description: "List of agent policy IDs. Requires Elastic Stack >= 8.15.0. " +
        "Conflicts with agent_policy_id.",
    ElementType: types.StringType,
    Optional:    true,
    Validators: []validator.List{
        listvalidator.ConflictsWith(path.Root("agent_policy_id").Expression()),
        listvalidator.SizeAtLeast(1),
    },
},
```

Required new imports: `listvalidator`, `path` (framework path package).

### `internal/fleet/elastic_defend_integration_policy/models.go`

Add `AgentPolicyIDs types.List` to `elasticDefendIntegrationPolicyModel`:

```go
AgentPolicyID  types.String `tfsdk:"agent_policy_id"`
AgentPolicyIDs types.List   `tfsdk:"agent_policy_ids"`
```

### `internal/fleet/elastic_defend_integration_policy/resource.go`

Add the `MinVersionPolicyIDs` constant and capability-check helper, mirroring
`internal/fleet/integration_policy/resource.go`:

```go
var (
    MinVersionPolicyIDs = version.Must(version.NewVersion("8.15.0"))
)
```

In `Create` and `Update`, before using `agent_policy_ids`, call `EnforceMinVersion` and return
early if the version gate fails.

### `internal/fleet/elastic_defend_integration_policy/request.go`

Update `buildBootstrapRequest` and `buildFinalizeRequest` to populate `PolicyIds` when
`agent_policy_ids` is set, and set `PolicyId` to the first element for compatibility with older
Kibana instances that may require it during the endpoint artifact-manifest bootstrap:

```go
// When agent_policy_ids is used:
if !model.AgentPolicyIDs.IsNull() && !model.AgentPolicyIDs.IsUnknown() {
    var ids []string
    _ = model.AgentPolicyIDs.ElementsAs(ctx, &ids, false)
    req.PolicyIds = &ids
    if len(ids) > 0 {
        req.PolicyId = &ids[0]  // compatibility: first element
    }
} else {
    req.PolicyId = model.AgentPolicyID.ValueStringPointer()
}
```

### `internal/fleet/elastic_defend_integration_policy/mapping.go`

Update `populateModelFromAPI` to mirror the generic resource read logic — populate only the
field that was originally configured in state:

```go
originallyUsedAgentPolicyID  := typeutils.IsKnown(model.AgentPolicyID)
originallyUsedAgentPolicyIDs := typeutils.IsKnown(model.AgentPolicyIDs)

if originallyUsedAgentPolicyID {
    model.AgentPolicyID = types.StringPointerValue(policy.PolicyId)
}
if originallyUsedAgentPolicyIDs {
    if policy.PolicyIds != nil {
        ids, d := types.ListValueFrom(ctx, types.StringType, *policy.PolicyIds)
        diags.Append(d...)
        model.AgentPolicyIDs = ids
    } else {
        model.AgentPolicyIDs = types.ListNull(types.StringType)
    }
}

if !originallyUsedAgentPolicyID && !originallyUsedAgentPolicyIDs {
    // Default: use singular field from API response
    model.AgentPolicyID = types.StringPointerValue(policy.PolicyId)
}
```

## Behavioural Contract

| Terraform config | Bootstrap `PolicyId` | Bootstrap `PolicyIds` | Finalize `PolicyId` | Finalize `PolicyIds` |
|---|---|---|---|---|
| `agent_policy_id = "abc"` | `"abc"` | nil | `"abc"` | nil |
| `agent_policy_ids = ["abc", "xyz"]` | `"abc"` (compat) | `["abc","xyz"]` | `"abc"` (compat) | `["abc","xyz"]` |

## Bootstrap-Phase Compatibility Note

The research comment identified an open question about whether the endpoint bootstrap endpoint
accepts `PolicyIds` without `PolicyId`. The safest approach — adopted here — is to always set
`PolicyId` to the first element of `PolicyIds` when the list form is used. This matches the
behavior of the kbapi deprecation notice ("Deprecated. Use policy_ids instead.") while maintaining
backward compatibility with older Kibana instances.

## No Schema Version Bump

`agent_policy_ids` is purely Optional (not Computed). The Plugin Framework treats an absent key
in existing state as null, which is valid for Optional-only attributes. Precedent: commit
`df995c0d` added `agent_policy_ids` to the live V1 schema of `elasticstack_fleet_integration_policy`
and changed `agent_policy_id` from Required to Optional — both without a version bump, and the
resource has been in production since. The defend resource follows the same logic.

## Open Questions

- **Bootstrap endpoint behavior**: Does the Kibana endpoint-package bootstrap endpoint
  (`ENDPOINT_INTEGRATION_CONFIG` input type) work correctly when `PolicyId` is omitted and only
  `PolicyIds` is set? Since `policy_id` is deprecated in the API, the "set both" approach
  (first element in `PolicyId`, full list in `PolicyIds`) is safest but needs a live API test to
  confirm the bootstrap step accepts `PolicyIds` at all for the endpoint package specifically.
- **AtLeastOneOf validator**: Should the schema enforce that at least one of `agent_policy_id` /
  `agent_policy_ids` must be set (matching the generic resource's behavior, which leaves this to
  documentation), or add an explicit `AtLeastOneOf` validator?
- **Space constraint documentation**: All agent policies in `agent_policy_ids` must reside in the
  same Kibana space as `space_ids`. Should the resource document this constraint only, or validate
  it at plan time (which would require API calls in `ModifyPlan`)?
- **CI stack version**: What is the minimum Elastic Stack version in CI for acceptance testing
  `agent_policy_ids` (requires ≥ 8.15.0)?
