## Why

The Elasticsearch Watcher API returns `::es_redacted::` as a placeholder for any secret value
embedded in a watch definition — including `input.http.request.auth.basic.password`. After the
v0.14.4 migration to the Terraform Plugin Framework, the provider already handles this redaction for
`actions` (PR #2296 and follow-up), but the identical problem was explicitly deferred for `input`.

Users who configure an `elasticstack_elasticsearch_watch` resource with an HTTP input that uses
basic authentication and a sensitive Terraform variable now see:

```
Error: Provider produced inconsistent result after apply
.input: inconsistent values for sensitive attribute
```

This happens because `fromAPIModel` marshals the raw Get Watch API response — including the
redacted sentinel — directly into the `input` state attribute on read-after-write. The Terraform
Plugin Framework then detects the mismatch between the plan's sensitive password and the redacted
API value and raises the error. The workaround of `lifecycle { ignore_changes = [input] }` is
ineffective because the failure occurs during the post-apply state write, before lifecycle rules
run.

## What Changes

- Extend `fromAPIModel` in `internal/elasticsearch/watcher/watch/models.go` to accept a
  `priorInput jsontypes.Normalized` parameter and apply `mergePreserveRedactedLeaves` to the `input`
  field, mirroring the existing `actions` handling.
- Update the `readWatch` call in `internal/elasticsearch/watcher/watch/read.go` to pass
  `state.Input` as `priorInput`.
- Add unit tests in the `watch` package for the `input` redaction-preservation path, mirroring
  `actions_merge_test.go` coverage for HTTP basic auth and other nested paths.
- Add an acceptance test fixture for a watch that uses HTTP input basic auth, verifying that
  `terraform apply` succeeds and subsequent plans are empty.
- Update `openspec/specs/elasticsearch-watch/spec.md` to add an explicit `input` redaction
  requirement (REQ-030) and scenarios parallel to the existing `actions` scenarios (REQ-014–016,
  REQ-023–027).

## Capabilities

### Modified Capabilities

- `elasticsearch-watch`: extend read/state synchronization to preserve prior known Terraform values
  of any JSON type at nested `input` paths where the Watcher API returns the redacted string
  sentinel, mirroring the existing `actions` path behavior.

## Impact

- `internal/elasticsearch/watcher/watch/models.go`
- `internal/elasticsearch/watcher/watch/read.go`
- `internal/elasticsearch/watcher/watch/actions_merge_test.go` (or a new `input_merge_test.go`)
- `internal/elasticsearch/watcher/watch/acc_test.go`
- `internal/elasticsearch/watcher/watch/testdata/` (new fixture if needed)
- `openspec/specs/elasticsearch-watch/spec.md`
