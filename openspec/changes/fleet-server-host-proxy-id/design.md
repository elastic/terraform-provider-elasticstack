## Context

`elasticstack_fleet_server_host` (`internal/fleet/serverhost`) has no `proxy_id` attribute. The Fleet API already supports it: the generated client types in `generated/kbapi/kibana.gen.go` carry `ProxyId *string \`json:"proxy_id,omitempty"\`` on `ServerHost` (read), `PostFleetFleetServerHostsJSONBody` (create), and `PutFleetFleetServerHostsItemidJSONBody` (update). This is schema/model wiring only; no client regeneration is required.

A near-identical feature already ships in `elasticstack_fleet_agent_download_source` (`internal/fleet/agentdownloadsource`), including a documented fix for an update-semantics edge case with `proxy_id` on PUT. This change mirrors that pattern rather than re-deriving it from scratch.

This design adopts the recommendation from the issue's implementation-research comment (2026-10-08, run `37718859097`) as its spine, per the Approach A analysis there.

## Goals / Non-Goals

**Goals:**
- Add `proxy_id` (Optional, string, no Computed/plan-modifiers) to the `fleet_server_host` schema, model, create body, update body, and `populateFromAPI`.
- Reuse the `agentdownloadsource.proxyIDForUpdate`-style unset-handling rule so clearing `proxy_id` in config reliably clears it server-side instead of silently leaving the prior value in place.
- Update the resource example to show `proxy_id` wired to `elasticstack_fleet_proxy`.
- Add acceptance test coverage mirroring `TestAccResourceFleetAgentDownloadSource_ProxyID`.

**Non-Goals:**
- Changes to `elasticstack_fleet_proxy` itself (already fully supports `proxy_id`).
- Regenerating `generated/kbapi`.
- `proxy_id` support on other Fleet sub-resources (outputs, agent policies, etc.) not named in the issue.

## Decisions

### Decision 1: Schema shape mirrors `agentdownloadsource`'s `proxy_id`

Add `proxy_id` as an optional string attribute with no computed behavior or plan modifiers. Its description will identify it as the Fleet proxy used by the server host, matching the sibling agent download source resource.

**Why:** Consistency with the only other resource in the provider that already models this exact Fleet API field. No reason to diverge.

### Decision 2: Model field and `populateFromAPI`

Add `ProxyID types.String \`tfsdk:"proxy_id"\`` to `serverHostModel` in `internal/fleet/serverhost/models.go`. In `populateFromAPI`, add `m.ProxyID = types.StringPointerValue(data.ProxyId)` alongside the existing field assignments.

### Decision 3: Create body — direct pass-through

In `toAPICreateModel`, set `body.ProxyId = m.ProxyID.ValueStringPointer()`. On create there is no "prior" value to reconcile against, so a direct pointer conversion is correct and matches `agentdownloadsource.toAPICreateModel`.

### Decision 4: Update body — `prior`-aware unset handling, reusing `proxyIDForUpdate`'s logic

`serverHostModel.toAPIUpdateModel` currently takes no `prior` parameter (unlike `agentdownloadsource.model.toAPIUpdateModel(ctx, prior model)`). Grow its signature to `toAPIUpdateModel(ctx context.Context, prior serverHostModel)` and compute `body.ProxyId` with the same three-way rule `agentdownloadsource.proxyIDForUpdate` already implements:

- Plan value known and non-empty → send as-is.
- Plan value unset/empty AND prior value was known and non-empty (i.e. the user is clearing a previously-set `proxy_id`) → send an explicit empty string `""`, not `nil`. The generated `json:"proxy_id,omitempty"` tag drops `nil` from the request body, and Fleet's PUT handler treats an omitted field as "leave unchanged" (per the behavior already confirmed for `agent_download_sources`); sending `""` is required to actually clear it.
- Plan value unset AND prior value was never set → send `nil` (field stays omitted, which is correct — there's nothing to clear).

Rather than duplicating the three-way branch, extract `agentdownloadsource`'s `proxyIDForUpdate(plan, prior types.String) *string` into a shared location (e.g. `internal/fleet` or `internal/utils/typeutils`) and call it from both resources, OR duplicate the small helper locally in `serverhost` with the same doc comment explaining the `omitempty` interaction. Either is acceptable; prefer extraction if it can be done without changing `agentdownloadsource`'s existing call sites or tests. This decision is left to implementation — see Open Questions.

`update.go`'s `updateServerHost` already loads `req.Prior` for space-ID resolution; pass `*req.Prior` (or the zero-value `serverHostModel{}` when `req.Prior` is nil, which cannot occur on a real Terraform update but mirrors defensive handling elsewhere) through to `toAPIUpdateModel`.

**Why:** Re-deriving the naive pass-through-only version would very likely reproduce the exact bug `agentdownloadsource` already hit and documented — unsetting `proxy_id` would silently fail to clear it server-side, causing state/reality drift or an "inconsistent result after apply" error. The fix is already written and tested; copying it is strictly lower-risk than re-deriving it.

### Decision 5: No new minimum-version gate

`fleet_server_host` already gates on `8.6.0` (`minVersionFleetServerHost` in `acc_test.go`). Per human direction on the issue, whether `proxy_id` needs its own, later minimum-version gate is left open for implementation — the acceptance test matrix will determine the right version constraints. If an acceptance test run against the minimum-supported stack shows `proxy_id` requires a newer version, a dedicated `entitycore.VersionRequirement` should be added at that point; this design does not pre-emptively add one.

## Open Questions

(Copied verbatim from the issue's implementation-research comment; see "Human direction" below for answers already given on the issue.)

- Does `PUT /fleet/fleet_server_hosts/{itemId}` actually exhibit the same "omitted proxy_id = leave unchanged" behavior confirmed for `agent_download_sources`? Should be verified against a live Kibana/Fleet instance before assuming the fix transfers unchanged.
- Should `examples/resources/elasticstack_fleet_server_host/resource.tf` be updated to show `proxy_id` wired to `elasticstack_fleet_proxy`, per the issue's HCL snippet?
- Is a minimum stack-version gate needed for `proxy_id` on server hosts, or has it been supported since the resource's existing `8.6.0` minimum?
- Should a new acceptance test mirror `TestAccResourceFleetAgentDownloadSource_ProxyID` (set / unset / update of `proxy_id`)?

## Human direction (from issue comments)

The issue author (`@tobio`) answered these open questions directly on the issue, and those answers take precedence over leaving them fully open:

- **Omitted-`proxy_id`-on-PUT behavior**: left open for implementation — must be verified against a live Kibana/Fleet instance before shipping (see Decision 4 and Non-Goals; this change documents the planned behavior but implementation must confirm it against a real stack).
- **Update the resource example**: yes — reflected in Decision-adjacent scope (see proposal.md "What Changes").
- **Minimum stack-version gate**: left open for implementation; the acceptance test matrix will determine the right version constraints (see Decision 5).
- **Mirror `TestAccResourceFleetAgentDownloadSource_ProxyID`**: yes — reflected in tasks.md.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Fleet's `PUT /fleet/fleet_server_hosts/{itemId}` omitted-field or empty-string behavior differs from `agent_download_sources` | Treat live acceptance verification as a release gate. If omission does not leave the value unchanged or `""` does not clear it successfully, revise Decision 4 and REQ-014/REQ-017 before implementing the verified endpoint-specific behavior. |
| Changing `toAPIUpdateModel`'s signature to take `prior` touches the only call site in `update.go` | Single call site, mechanical change, directly mirrors the already-existing `agentdownloadsource` pattern. |
