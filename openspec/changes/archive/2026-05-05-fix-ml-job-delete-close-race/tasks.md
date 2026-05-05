## 1. Client helper

- [x] 1.1 Add `WaitForMLJobClosed(ctx, apiClient, jobID)` to `internal/clients/elasticsearch/ml_job.go` using `asyncutils.WaitForStateTransition` and `GetMLJobStats`; treat a nil stats result (job not found) as settled

## 2. Delete fix

- [x] 2.1 Update `internal/elasticsearch/ml/anomalydetectionjob/delete.go` to call `WaitForMLJobClosed` after `CloseJob` (log warning and continue if wait fails), then call `DeleteJob`; if `DeleteJob` fails, retry once with `force=true` and surface the retry error as a diagnostic if that also fails

## 3. Spec update

- [x] 3.1 Apply the delta spec: update REQ-021–REQ-022 in `openspec/specs/elasticsearch-ml-anomaly-detection-job/spec.md` to require polling for `closed` state after `CloseJob` returns and before `DeleteJob` is called

## 4. Verification

- [x] 4.1 Run `make build` to confirm the project compiles
- [x] 4.2 Run the `TestAccResourceMLJobStateExplicitConnection` acceptance test locally (or confirm it passes in CI) to verify the fix
