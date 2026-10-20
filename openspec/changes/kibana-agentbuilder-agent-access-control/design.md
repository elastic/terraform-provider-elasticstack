## Context

`elasticstack_kibana_agentbuilder_agent` (`internal/kibana/agentbuilderagent`) manages Kibana Agent Builder agents via `POST`/`GET`/`PUT`/`DELETE /api/agent_builder/agents{,/{id}}`. The generated client (`generated/kbapi`) already carries an `AccessControl *struct{ AccessMode ... }` field on both `PostAgentBuilderAgentsJSONBody` and `PutAgentBuilderAgentsIdJSONBody` (`kibana.gen.go` around `PostAgentBuilderAgentsJSONBodyAccessControlAccessMode`/`PutAgentBuilderAgentsIdJSONBodyAccessControlAccessMode`, enum values `private`/`public`/`shared`), but nothing in `internal/kibana/agentbuilderagent` reads or writes it.

`access_mode` is `required` inside the `access_control` object whenever the object is present (OpenAPI `required: [access_mode]`), and carries `x-state: Generally available; added in 9.4.0`. There is a genuine discrepancy between sources on the default when `access_control` is omitted entirely: the OpenAPI description text says agents default to `private`, while Elastic's Agent Builder product docs state Stack 9.4–9.5 defaults new custom agents to **Public** visibility (private on Serverless), with a planned change to align the two. This proposal does not resolve that discrepancy (see Open Questions); it is recorded so implementation does not hardcode either default into tests or docs.

Separately, the generated client also exposes a **dedicated** sub-resource, `GET`/`PUT /api/agent_builder/agents/{id}/access_control` (`GetAgentBuilderAgentsIdAccessControlWithResponse`, `PutAgentBuilderAgentsIdAccessControlWithResponse` in `kibana.gen.go`), documented as `x-state: Generally available; added in 9.5.0`. Its `GET` response documents `access_control.access_mode`. Whether the *main* agent `GET /agents/{id}` reliably returns `access_control` on 9.4.x servers is unconfirmed — the main-GET response is example-only in the spec (no typed response schema; the generated client's `GetAgent` wrapper decodes into `models.Agent`, an untyped JSON body), with no field-level introduction-version annotation on that example. Given that uncertainty, this design gates the entire attribute — write and read alike — at 9.5.0, where read-back is documented, rather than at 9.4.0 where only write would be confirmed.

The sub-resource `PUT` also manages a separate `entries` list (individual users each granted a `user`/`editor`/`manager` role) — that is a distinct grants-list feature, out of scope here (see Non-Goals); this change only reads `access_control.access_mode` from the sub-resource `GET` and writes `access_control.access_mode` via the main agent `POST`/`PUT` bodies (not via the sub-resource `PUT`).

## Goals / Non-Goals

**Goals:**
- Expose `access_control.access_mode` as a schema attribute on `elasticstack_kibana_agentbuilder_agent`, settable on create and update.
- Detect drift when an explicitly configured mode differs from the server via read-back; when the attribute is omitted or management is relinquished, refresh adopts the server's current mode without requiring a corrective update.
- Gate the whole feature at a single, clearly documented minimum version given the main-GET read-back uncertainty on 9.4.x.
- Treat removing `access_control` from config as relinquishing management of the mode (server keeps whatever mode it currently has), not as resetting it to a specific value.

**Non-Goals:**
- Managing the per-agent user/role `entries` grants list via the sub-resource `PUT`. The issue's gap fingerprint covers only `access_control.access_mode` on the create/update bodies.
- Supporting `access_control` writes in a narrower 9.4.0–9.4.x window where no confirmed read-back endpoint exists. If a future live-cluster probe confirms the main GET reliably returns `access_control` on 9.4.x, lowering the gate to 9.4.0 is a candidate follow-up, not part of this change.
- Data source parity (`elasticstack_kibana_agentbuilder_agent` data source exposing `access_control`). Not required to close the triggering gap; tracked as an open question below.
- Resolving the default-`access_mode`-when-omitted discrepancy between the OpenAPI text and Elastic's product docs. Acceptance tests for the omitted case must assert against whatever a live cluster actually returns, not a hardcoded literal.

## Decisions

### Decision 1: Schema shape — `SingleNestedAttribute`, Optional+Computed, one required nested string

```hcl
access_control = {                                  # optional, computed, single nested object
  access_mode = "public"                             # required (within the object); one of "public", "shared", "private"
}
```

Attribute-level: `access_control` is `Optional: true, Computed: true` with `objectplanmodifier.UseStateForUnknown()`. The nested `access_mode` is `Required: true` (required *within* the object, matching the API's `required: [access_mode]`) with `stringvalidator.OneOf("public", "shared", "private")`.

**Why Optional+Computed (not plain Optional):** the resource needs to capture the server-assigned mode (including whatever default a cluster actually applies) when the user never configures `access_control`, and needs read-back to detect out-of-band changes to an explicitly configured mode. When configuration omits the attribute, refresh adopts the server's current value rather than treating it as managed drift. A plain `Optional` attribute without `Computed` cannot hold a server-populated value that wasn't in the user's config. This mirrors the reasoning already used for `skill_ids`/avatar fields elsewhere in this same resource, and is the documented reason the dashboard resource's `access_control.access_mode` (`internal/kibana/dashboard/schema.go`) is a weaker precedent here: that field is plain `Optional` with `RequiresReplace`, appropriate for its own API where the field isn't required-when-present and the dashboard resource doesn't read it back. This resource's `access_mode` is required-when-present and updatable in place (the API's `PUT` accepts it), so neither of the dashboard's choices transfers directly.

### Decision 2: Keep the lifecycle baseline at 9.3.0; gate writes that send the field at 9.5.0

Add a new constant to `internal/kibana/agentbuilder/versions.go`, e.g. `MinAccessControlAPIVersion = semver.Must(semver.NewVersion("9.5.0"))`, documented alongside the existing `MinCoreAPIVersion` (9.3.0 — agents/tools) and `MinExtendedAPIVersion` (9.4.0 — skills/workflows/advanced config) rollout-timeline comment.

Keep `agentModel`'s existing embedded `agentVersionGate` as the lifecycle-wide 9.3.0 requirement. Do not add the 9.5.0 requirement to `GetVersionRequirements`: the shared envelope checks those requirements before refresh and deletion as well as writes, and a carried-forward `access_control` value must not prevent those operations against an older server.

Instead, check `MinAccessControlAPIVersion` immediately before a create or update that will send `access_control`: on create, when the planned value is known and non-null; on update, when the configuration value is known and non-null (Decision 3). Return a version-requirement diagnostic before the write API call if the server is older than 9.5.0. This is distinct from read-back's capability check: on an older server, refresh leaves `access_control` null and continues, and deletion remains subject only to the 9.3.0 lifecycle baseline.

### Decision 3: Update write path omits `access_control` based on config, not plan — relinquish, don't reset

`access_control` is Optional+Computed, so after a create/read cycle the **plan** value is never null once the server has assigned a mode (Computed carries the prior value forward via `UseStateForUnknown`). If `toAPIUpdateModel` checked the plan the way it checks `AvatarColor`/`Instructions` (`typeutils.IsKnown(model.AvatarColor)`), it would always resend `access_control` on every update — once set, never omittable — which forecloses the "stop managing this, let the server keep whatever it has" behavior called for in the proposal.

Instead, `updateAgent` (`internal/kibana/agentbuilderagent/update.go`) reads `req.Config.AccessControl` directly (available on `entitycore.KibanaWriteRequest[agentModel]` without any `entitycore` change — `KibanaWriteRequest` already carries both `Plan` and `Config`) and passes a `sendAccessControl bool` into the `toAPIUpdateModel` closure, the same way `supportsSkillIDs` is threaded in today. When `req.Config.AccessControl` is null (block absent from HCL), the update body omits `access_control` entirely, regardless of what the plan/state carry forward. When present in config, the plan's `access_control.access_mode` value is sent (plan and config agree on the value when both are non-null, since `access_mode` has no plan modifier of its own).

This resolves the research's "config vs. plan write-intent" open question: `KibanaWriteRequest.Config` is already available to `updateAgent` today (see `internal/entitycore/kibana_resource_envelope.go`), so no `entitycore` changes are needed — only `toAPIUpdateModel`'s signature and `updateAgent`'s call site change.

On **create**, there is no prior server state to relinquish, so `toAPICreateModel` follows the existing `typeutils.IsKnown(model.X)`-on-plan convention used for `AvatarColor`/`Instructions`/`Description`: when `access_control` is known (i.e., configured) on the plan, include it; otherwise omit it and let the server apply its own default.

### Decision 4: Read-back via the dedicated sub-resource, gated and sequenced after the main GET

Add `GetAgentAccessControl(ctx, client, spaceID, agentID) (*models.AgentAccessControl, diag.Diagnostics)` to `internal/clients/kibanaoapi/agentbuilder_agent.go`, next to `GetAgent`/`CreateAgent`/`UpdateAgent`/`DeleteAgent`, wrapping `client.API.GetAgentBuilderAgentsIdAccessControlWithResponse` the same way `GetAgent` wraps `GetAgentBuilderAgentsIdWithResponse`. Add a corresponding `models.AgentAccessControl{ AccessMode string }` (or reuse a shared small struct) in `internal/models/agent_builder.go`, which today has no `AccessControl` field on `models.Agent` at all.

In the resource read callback in `read.go` (registered by the resource in `resource.go` and reused by the envelope for refresh and post-write reads), after a successful (non-404) main `GET`:
1. Call `client.EnforceMinVersion(ctx, agentbuilder.MinAccessControlAPIVersion)` (the same capability-check mechanism `create.go`/`update.go` already use for `MinExtendedAPIVersion` — not a stored version field).
2. If the capability check fails (server below 9.5.0), set `access_control` null in state and continue, even if a prior state value was carried forward. This is not a user-facing error at read time; only a write that sends the attribute is rejected below 9.5.0 (Decision 2).
3. If the capability check passes, call the new `GetAgentAccessControl` wrapper and populate `access_control.access_mode` from its response.

The existing main-GET not-found (404 → remove from state) handling must run and return before step 1 is ever reached, so a deleted agent never triggers a second API call.

Do not change `readAgentDataSource`: data-source parity is out of scope, and its model does not contain `AccessControl`. The resource callback is the single read-back integration point; the envelope reuses it for refresh and post-write reads.

### Decision 5: No sub-resource `PUT`

Writes always go through the main `POST`/`PUT /agents{,/{id}}` bodies' `access_control` field (Decisions 1 and 3), never through `PUT /agents/{id}/access_control`. That sub-resource `PUT` is entries-list-shaped (requires at least one `entries` grant) and is reserved for a possible future change managing grants (see Non-Goals); mixing it into this change would couple an unrelated feature to this gap-closure.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Gating at 9.5.0 instead of 9.4.0 delays availability by one minor version | Deliberate given the main-GET read-back uncertainty on 9.4.x (see Context and Open Questions); revisit if a live-cluster probe confirms earlier read-back. |
| Two API calls per read (main GET + access-control GET) on 9.5.0+ | Only on supported versions; the second call is skipped entirely below 9.5.0. No change to existing call count on versions that don't support the field. |
| Default-mode-when-omitted discrepancy (OAS text vs. product docs) could make an acceptance test assert the wrong default | Tests must assert against the live cluster's actual response for the omitted case, not a hardcoded literal (see Non-Goals and Open Questions). |
| A write-specific 9.5.0 check could be applied inconsistently across create and update | Keep the check adjacent to the write operation and cover both configured writes in tests; do not place it in lifecycle-wide `GetVersionRequirements`. |

## Open questions

- **Blocking:** Default `access_mode` when omitted conflicts between the OAS text (`private`) and Elastic's product docs (`public` for Stack 9.4–9.5, `private` on Serverless). Needs confirmation against a live cluster before finalizing the "omitted" test assertions.
- Whether a 9.4.x server's main `GET /agents/{id}` reliably returns `access_control` (which would let the attribute be supported from 9.4.0 instead of 9.5.0) is unconfirmed — recommend a live-cluster probe before implementation.
- Is Agent Builder access control GA or tech-preview as of 9.4.0/9.5.0 by Elastic's own feature-maturity labeling (separate from the OAS `x-state` annotations used here)?
- Should the data source also expose `access_control.access_mode` for parity? Not required to close this gap.

Note: the research comment's fourth open question ("write-time intent: config vs. plan for update") is resolved in Decision 3 above via code inspection of `entitycore.KibanaWriteRequest` — `req.Config` is already available to `updateAgent` with no framework change required. It is omitted from the list above since it is no longer open; see Decision 3.

## Live-cluster probe results

Probed against a live Elastic Stack 9.5.5 (Kibana and Elasticsearch at the default `testing.md` endpoints):

- Default `access_mode` when `access_control` is omitted on create: **`public`** (matches Elastic's product docs for Stack 9.4-9.5, not the OAS description text). The create response and main `GET /agents/{id}` both include `access_control: {access_mode, entries: []}` on 9.5.5.
- `GET /agents/{id}/access_control` returns `{"access_control": {"access_mode": ..., "entries": []}, "permissions": {...}}`, so the access-control object is wrapped one level deep and needs its own response model (`AgentAccessControlResponse`).
- `PUT /agents/{id}` with `access_control.access_mode = "shared"` updates the mode and the sub-resource GET reflects it.
- **Not verified:** whether a 9.4.x server's main `GET` returns `access_control`. No 9.4.x cluster was available (only 9.5.5), so the 9.5.0 gate decision is unchanged and the first Open Question remains open. The "blocking" default-mode question is resolved for 9.5.x only (`public`); Serverless was not probed.

## Implementation notes

- `avatar_color` / `avatar_symbol` (Optional+Computed, previously without plan modifiers) gained a small `useStateForUnknownIncludingNull` plan modifier. Adding the computed `access_control` object made the framework mark these null computed attributes unknown on every plan after apply (non-empty plan on omitted `access_control`), because Terraform proposes a null `access_control` against a non-null prior value. The modifier carries the prior (including null) state value forward.
- The resource had no `read.go`; a dedicated `readAgent` callback was added (replacing `SimpleKibanaRead`) to perform the main GET followed by the gated access-control GET.
