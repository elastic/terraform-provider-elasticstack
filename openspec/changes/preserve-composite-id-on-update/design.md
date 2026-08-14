## Context

`elasticstack_elasticsearch_security_role`'s `id` attribute has `UseStateForUnknown()` (`internal/elasticsearch/security/role/schema.go:204-210`), which promises Terraform that `id` will not change on Update. However, `writeRole` — the single write callback shared by Create and Update — unconditionally recomputes `id` via `client.ID(ctx, roleID)`, which always resolves the *live* cluster's UUID (`internal/clients/elasticsearch_scoped_client.go:94-118`), never the `id` already present in prior state. When a deployment has been recreated behind the same endpoint (same config, new cluster UUID), Update rewrites `id` and Terraform aborts the apply with `Error: Provider produced inconsistent result after apply`. The bad id is nonetheless persisted, so the very next plan is clean and a re-apply always succeeds — a one-shot failure that still breaks unattended pipelines.

Three other resources share the identical shape (unconditional `client.ID` call inside a Create+Update write path, combined with `UseStateForUnknown` on `id`):

| Resource | Write callback | Schema |
|---|---|---|
| `security_role` | `internal/elasticsearch/security/role/update.go` (`writeRole`, shared Create+Update) | `internal/elasticsearch/security/role/schema.go:204-210` |
| `data_stream_lifecycle` | `internal/elasticsearch/index/datastreamlifecycle/write.go` (`writeDataStreamLifecycle`, shared Create+Update) | `internal/elasticsearch/index/datastreamlifecycle/schema.go:39-45` |
| `watch` | `internal/elasticsearch/watcher/watch/update.go` (`updateWatch`, separate from `createWatch`) | `internal/elasticsearch/watcher/watch/schema.go:42-48` |
| `ml_datafeed` | `internal/elasticsearch/ml/datafeed/update.go` (`updateDatafeed`, separate from `createDatafeed`) | `internal/elasticsearch/ml/datafeed/schema.go:51-57` |

`elasticstack_elasticsearch_security_user` has no plan modifier on `id` (`internal/elasticsearch/security/user/schema.go:41-44`), so it plans `id -> (known after apply)` on every update instead of crashing — it makes no promise the recompute could break, so it needs no change.

`internal/elasticsearch/index/template/write.go:56-65` already implements the correct pattern for a shared Create+Update callback:

```go
if req.Prior == nil {
    id, idDiags := client.ID(ctx, plan.Name.ValueString())
    ...
    priorForRead.ID = types.StringValue(id.String())
} else {
    priorForRead.ID = req.Prior.ID
}
```

`client.ID` → `ClusterID` → `Info()` has no side effect beyond returning a mutex-cached cluster-info response (`internal/clients/elasticsearch_scoped_client.go:94-135`), so it is safe to skip entirely on Update — there is no correctness reason to keep calling it.

## Goals / Non-Goals

**Goals:**
- Stop `id` from changing during Update for `security_role`, `data_stream_lifecycle`, `watch`, and `ml_datafeed`, so `UseStateForUnknown` on `id` is never violated regardless of whether the live cluster UUID differs from the one in state.
- Compute `id` only once, at Create, for all four resources; Update preserves whatever `id` is already in prior state.
- Reuse the existing `index/template` guard shape rather than inventing a new pattern, so the fix reads as "apply the proven precedent," not "add new machinery."

**Non-Goals:**
- Migrating a stale cluster-UUID prefix already in state. See Open Questions below — this is explicitly deferred.
- Changing what `id` looks like, how it's parsed on import, or how Read/Delete resolve it from a composite id. Only *when* it is computed changes.
- Touching `security_user` or any resource not confirmed in this run to share the write/schema shape above.

## Decisions

- **Guard shape**: for the two resources with a single write function shared by Create and Update (`security_role`, `data_stream_lifecycle`), branch on `req.Prior == nil`: compute `id` via `client.ID` only when `req.Prior == nil` (Create); otherwise set `data.ID = req.Prior.ID`. This mirrors `index/template/write.go:56-65` exactly.
- **Separate Create/Update resources**: for `watch` and `ml_datafeed`, whose `updateWatch`/`updateDatafeed` are only ever invoked for Update (Create is a distinct function, `createWatch`/`createDatafeed`), there is no `req.Prior == nil` branch needed — the `client.ID` call is simply removed from the Update path and the callback sets `data.ID = req.Prior.ID` (or leaves `plan.ID` as decoded from state, whichever the surrounding envelope expects) unconditionally.
- **Shared helper**: extract the branch into a small helper (e.g. `entitycore.PreserveIDOnUpdate` or a local helper per package if a cross-package helper doesn't fit the existing `entitycore.WriteRequest[T]` generic shape cleanly) so the four call sites read identically and future resources with this shape have an obvious pattern to copy. If a fully generic helper is awkward given each model's distinct ID field, it is acceptable to keep the guard inline per package as long as all four follow the identical two-branch (or single-branch) shape — the important invariant is behavioral consistency, not code deduplication for its own sake.
- **No Read-time change**: Read/refresh behavior is untouched by this change. A stale cluster-UUID prefix that predates this fix remains in state until the role/resource is otherwise recreated or explicitly migrated by a future, separately-scoped change.
- **Test strategy**: unit tests construct a `req.Prior` (or, for `watch`/`datafeed`, the equivalent prior-state input) whose `id` carries a cluster-UUID prefix that a mocked or real `client.ID` call would not currently reproduce, and assert the write callback's returned `id` equals the prior value unchanged, without asserting on whether `client.ID` was called. Additionally, add an acceptance test (recommended on `security_role`, matching the issue's own reproduction) that imports state with a deliberately "incorrect" cluster-UUID prefix and confirms a subsequent attribute-changing `terraform apply` succeeds without an inconsistent-result error and without an `id` diff.

## Open questions

- **Read-time stale-prefix migration** — downsides identified: (1) produces an unprompted `id` diff on `plan`/`refresh` with no config change, read as external drift; (2) adds a live `Info()` call to every Read for these resources, which today's read callbacks don't make; (3) the cluster-UUID prefix is purely cosmetic (never scopes the actual API call, always `name`-keyed against whatever cluster is connected), so rewriting it changes a value external tooling/`terraform_remote_state` consumers may treat as stable, without any functional justification. Since Approach C alone eliminates the crash, recommend treating this as an independent, explicitly-scoped fast-follow rather than bundling it — but `@tobio` should confirm.
- No existing acceptance/unit test simulates "cluster UUID changed between Create and Update" for any of the four resources — worth deciding how the fix's test should mock this (fake `Prior` in a unit test vs. a mocked `client.ID`).

**Resolution (from issue #4490 discussion, 2026-08-14):** `@tobio` confirmed read-time stale-prefix migration is out of scope for this change (tracked as a candidate fast-follow, not part of this proposal). For the test-plan question, `@tobio` directed: explore an acceptance test that imports with an "incorrect" cluster UUID and asserts a following update succeeds — see the Decisions section's test strategy above and `tasks.md`.

## Risks / Trade-offs

- [Risk] `watch` and `ml_datafeed`'s Update callbacks currently set `id` unconditionally from a fresh `client.ID` call; removing that call means Update no longer touches cluster-UUID resolution at all. Mitigation: `client.ID` has no side effect beyond a cached `Info()` read (confirmed above), so removing the call changes no observable behavior other than fixing the bug.
- [Risk] A shared generic helper across four packages with four different model types may not be worth the abstraction if each model's ID field access differs enough. Mitigation: the Decisions section explicitly permits per-package inline guards as a fallback — consistency of behavior matters more than a forced shared helper.
- [Risk] Existing delta/canonical specs for `data_stream_lifecycle` (and, to a lesser degree, `watch`/`ml_datafeed`) explicitly document today's buggy "create and update recompute id" behavior as correct. Mitigation: this change's delta specs correct that wording for all four capabilities so the spec and implementation stay in sync.
