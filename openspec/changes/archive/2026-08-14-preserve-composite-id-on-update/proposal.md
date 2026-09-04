## Why

`elasticstack_elasticsearch_security_role` (and three other resources with the identical shape) recompute their composite `id` (`<cluster_uuid>/<name>`) from the *live* cluster's UUID on every Update, even though the `id` attribute uses `UseStateForUnknown` and Terraform's plan therefore does not expect it to change. If the cluster UUID stored in state (set at Create) differs from the UUID the provider currently sees — for example after an Elastic Cloud deployment recreation behind the same endpoint — Update silently rewrites `id` and Terraform aborts the apply with `Error: Provider produced inconsistent result after apply`. The rewritten id *is* persisted, so a re-apply always succeeds, making this a one-shot failure that still breaks unattended pipelines mid-apply.

`elasticstack_elasticsearch_security_user` does not have this problem: its `id` has no `UseStateForUnknown` modifier, so Terraform plans `id -> (known after apply)` on every update and the recompute is contractually expected. `security_role`, `data_stream_lifecycle`, `watch`, and `ml_datafeed` added `UseStateForUnknown` to `id` (for `security_role`, in #2160) without also removing the unconditional `client.ID` recompute from their Create+Update write path, so all four make a promise (`id` won't change) that their write callback can break.

## What Changes

- In `writeRole` (`internal/elasticsearch/security/role/update.go`), only compute `id` via `client.ID` when `req.Prior == nil` (Create). On Update, carry `id` forward from `req.Prior.ID` instead of recomputing it.
- Apply the same guard to the other three resources confirmed to share this exact shape:
  - `writeDataStreamLifecycle` (`internal/elasticsearch/index/datastreamlifecycle/write.go`) — single Create+Update write function, same fix shape as `security_role`.
  - `updateWatch` (`internal/elasticsearch/watcher/watch/update.go`) — separate Create/Update functions; `updateWatch` always has non-nil prior state, so it drops the `client.ID` call entirely and reuses the incoming `id` from state.
  - `updateDatafeed` (`internal/elasticsearch/ml/datafeed/update.go`) — same as `watch`: separate Create/Update, `updateDatafeed` drops its `client.ID` call and reuses state.
- Extract the `req.Prior == nil` / carry-forward guard already used in `internal/elasticsearch/index/template/write.go:56-65` into a small shared helper (or otherwise mirror that precedent) so `security_role` and `data_stream_lifecycle` apply it consistently rather than re-deriving it independently.
- Add unit test coverage for all four resources exercising Update with a `req.Prior.ID` (or state) whose cluster-UUID prefix differs from what `client.ID` would currently compute, asserting the returned `id` is unchanged.
- Add (or extend) an acceptance test — likely on `security_role`, per the issue's reproduction — that imports a resource with a composite id carrying a deliberately "incorrect" cluster UUID and asserts that a subsequent config-driven update succeeds without an inconsistent-result error and without changing `id`.
- Update the delta specs for all four capabilities so the documented Identity requirement describes "computed once at Create, preserved from prior state at Update" instead of "computed at create or update."
- For `data_stream_lifecycle` only (where `name` can change in place), add an `id` plan modifier after `UseStateForUnknown` so a name change plans `id` as unknown and apply writes `<prior-uuid>/<new-name>` instead of failing with an inconsistent-result error. Role, watch, and datafeed identity keys are `RequiresReplace` and do not need this.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `elasticsearch-security-role`: Requirement: Identity (REQ-005–REQ-006) — `id` is computed only at Create; Update preserves the prior `id` unchanged.
- `elasticsearch-data-stream-lifecycle`: Requirement: Identity (REQ-005–REQ-006) — same correction, plus in-place `name` changes plan `id` as unknown and apply `<prior-uuid>/<new-name>`.
- `elasticsearch-watch`: Requirement: Identity (REQ-005–REQ-006) — same correction.
- `elasticsearch-ml-datafeed`: Requirement: Identity and import (REQ-007–REQ-009) — same correction to the "During create and update, the resource SHALL derive `id`" sentence.

## Impact

- `internal/elasticsearch/security/role/update.go` (and its unit tests)
- `internal/elasticsearch/index/datastreamlifecycle/write.go` (and its unit tests)
- `internal/elasticsearch/index/datastreamlifecycle/id_plan_modifier.go` (and its unit tests)
- `internal/elasticsearch/watcher/watch/update.go` (and its unit tests)
- `internal/elasticsearch/ml/datafeed/update.go` (and its unit tests)
- A new or extended shared helper mirroring `internal/elasticsearch/index/template/write.go:56-65`, likely in `internal/entitycore` or a small package shared by these callbacks
- Acceptance tests for at least `security_role` covering import-with-stale-uuid then update
- `openspec/specs/elasticsearch-security-role/spec.md`, `openspec/specs/elasticsearch-data-stream-lifecycle/spec.md`, `openspec/specs/elasticsearch-watch/spec.md`, `openspec/specs/elasticsearch-ml-datafeed/spec.md` (synced from this change's delta specs once implemented)

## Out of Scope

- Migrating or rewriting a stale cluster-UUID prefix during Read/refresh. Confirmed out of scope for this change (see Open Questions in `design.md`); the cluster-UUID prefix is cosmetic (never scopes the actual API call) and read-time migration would itself produce an unprompted `id` diff on plan/refresh.
- Any resource not confirmed to share this exact write/schema shape (`security_role`, `data_stream_lifecycle`, `watch`, `ml_datafeed` only).
- `elasticstack_elasticsearch_security_user`, which already plans `id` as unknown and has no `UseStateForUnknown` modifier to violate.
