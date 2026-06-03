## Context

`elasticstack_fleet_agent_policy` exposes a `space_ids` attribute backed by the Kibana
Fleet `AgentPolicy.SpaceIds *[]string` field. The generated kbapi struct tags this field
`json:"space_ids,omitempty"`, meaning the Fleet API response body omits it when the value
is the default (the default-space policy, or in many Fleet versions, always). The provider
code at `internal/fleet/agentpolicy/models.go:211–219` currently treats the omitted field
as "space_ids is null" and writes `types.SetNull` into the model, overwriting the planned
value and producing:

  Provider produced inconsistent result after apply: .space_ids: was
  cty.SetVal([]cty.Value{cty.StringVal("example_id")}), but now null.

The same pattern was already identified and solved for optional string fields
(`DataOutputId`, `FleetServerHostId`, `DownloadSourceId`) via the `preserveNullStr` closure
defined at `models.go:110–115`. That closure keeps the model null when the API value is nil
and the model is already null — preventing a null→value overwrite — but there is no
equivalent for the `SpaceIDs` set field.

## Goals / Non-Goals

**Goals:**
- Fix the "Provider produced inconsistent result after apply" error for `space_ids`.
- Apply the null-preservation pattern consistently for `space_ids` on all three paths that
  call `populateFromAPI` (Create, Read, Update).
- Add or adjust a unit test that covers the API-returns-nil case for `space_ids`.

**Non-Goals:**
- Adding a Kibana space existence validator for `space_ids` values.
- Changes to `GetOperationalSpaceFromState` or `SpaceIDFromSet`.
- Upstream Fleet API changes.
- Adding explicit error surfacing when a named space does not exist (follow-on, out of scope).

## Decisions

### Decision 1: Preserve configured value when API omits space_ids

In `populateFromAPI`, replace the current unconditional null-write for the API-omits-nil
case with a guard that retains the existing model value:

```go
if data.SpaceIds != nil && len(*data.SpaceIds) > 0 {
    spaceIDs, d := types.SetValueFrom(ctx, types.StringType, *data.SpaceIds)
    if d.HasError() {
        return d
    }
    model.SpaceIDs = spaceIDs
} else if !model.SpaceIDs.IsNull() && !model.SpaceIDs.IsUnknown() {
    // API omitted space_ids (omitempty) — retain configured value
} else {
    model.SpaceIDs = types.SetNull(types.StringType)
}
```

**Why:** This is the minimal change that is consistent with the codebase's established pattern
for other optional Fleet fields. It handles all scenarios:
- API returns ids → use API value (unchanged from before).
- API returns nil AND model has a configured value → keep model value.
- API returns nil AND model has null → write null (unchanged from before).

**Trade-off accepted:** If the Fleet API silently trims invalid space IDs, the provider
reports the user's value rather than the actual trimmed value. This is the same trade-off
already accepted for the optional string fields and is the correct position for a bug fix.

### Decision 2: No schema change

The `space_ids` attribute definition in the schema is correct; only the read-path
interpretation of the API response changes.

### Decision 3: Unit test addition

The existing acceptance test `TestAccResourceAgentPolicyWithSpaceIDs` tests the full
lifecycle but requires a real Kibana. Add a focused unit test for `populateFromAPI` that
covers the API-omits-nil case (nil SpaceIds with a configured model value) without a live
stack. Alternatively, update an existing unit test if one exercises `populateFromAPI`.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **API silently ignores an invalid space ID** | The API returns 200 with omitted space_ids; the provider retains the user's value. The user may not realise the space was ignored. A validator is a follow-on. |
| **Fleet GET behaviour varies by version** | The omitempty is in the generated struct; field omission is consistent across Fleet versions ≥ 9.1. |

## Open Questions

- Does the Fleet GET always omit `space_ids` (omitempty) or only in certain cases (e.g.,
  single-space policies)? This affects how targeted a unit test can be. (Non-blocking: the
  fix is correct regardless; a thorough acceptance test covers the full lifecycle.)
- Is the user's referenced space pre-existing? If not, the policy may be silently created in
  the default space, making a validator a useful follow-on. (Non-blocking for this fix.)
