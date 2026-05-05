## Context

The `anomalydetectionjob` delete function calls `CloseJob` then immediately calls `DeleteJob`. The Elasticsearch ML API uses optimistic concurrency control on its internal `.ml-config` index: `DeleteJob` reads the document's `seqNo` then performs a conditional write. If `CloseJob`'s final state transition commits to the primary shard between the read and write of `DeleteJob`, Elasticsearch returns HTTP 409 `version_conflict_engine_exception` (`required seqNo [N], current seqNo [N+1]`). The CI log shows this +1 pattern consistently, confirming the exact race.

The fix is to poll `GetMLJobStats` after `CloseJob` returns, waiting until the job's reported state is `closed` (or the job is gone), before calling `DeleteJob`. The codebase already has all the building blocks: `asyncutils.WaitForStateTransition` for bounded polling and `GetMLJobStats` for state inspection; `jobstate/state_utils.go` uses the same pattern to wait for open/close transitions.

## Goals / Non-Goals

**Goals:**
- Eliminate the `version_conflict_engine_exception` on `DeleteJob` when the job was open at destroy time.
- Reuse existing polling infrastructure (`asyncutils`, `GetMLJobStats`) without introducing new abstractions.

**Non-Goals:**
- Changes to the job state resource or datafeed resources.
- Timeout configuration for the new wait — the existing Terraform delete-operation context provides the bound.

## Decisions

**Polling location**: Add `WaitForMLJobClosed` to `internal/clients/elasticsearch/ml_job.go` (alongside `GetMLJobStats`, `OpenMLJob`, `CloseMLJob`) and call it from `anomalydetectionjob/delete.go`. This keeps the helper co-located with other ML job client functions and avoids duplicating the polling logic inline.

**Retry DeleteJob with force=true on first failure**: The polling wait eliminates the 409 race in the normal case, but if the wait times out (e.g. context nearing expiry) the job may still be open when DeleteJob is called. A single retry with `force=true` handles this edge case without masking legitimate errors — the retry error is still surfaced as a Terraform diagnostic if it also fails.

**Alternative considered — retry DeleteJob on 409 only**: Could catch only the 409 and retry. Rejected in favour of a broader retry on any first-failure, since other transient errors can also occur and `force=true` is safe when the intent is deletion.

**Alternative considered — inline poll in delete.go**: Could put the `asyncutils.WaitForStateTransition` call directly in `delete.go`. Rejected: the helper belongs in the client layer alongside `GetMLJobStats`. Keeping it there makes it reusable and keeps delete.go focused on orchestration.

**Context/timeout**: The context passed to `Delete()` already carries Terraform's delete timeout (default 20 minutes). No additional timeout parameter is needed.

**"Not found" as settled state**: If `GetMLJobStats` returns `nil` (job not found), the wait treats that as the job being gone and returns immediately. This handles edge cases where the close also deleted the job or it was externally removed.

## Risks / Trade-offs

[Slower delete for open jobs] → Poll adds latency (up to a few seconds, typically one 2s poll tick) when a job is open at destroy time. This is acceptable — it prevents a hard failure.

[Poll loop runs indefinitely if job stays in `opening` or `closing` state forever] → Mitigated by the Terraform delete context timeout (default 20 min). In practice, close transitions complete in seconds.

## Migration Plan

No migration needed — this is a bug fix with no API or schema changes. Existing Terraform state is unaffected.
