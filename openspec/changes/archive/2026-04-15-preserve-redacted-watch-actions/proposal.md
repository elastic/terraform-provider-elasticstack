## Why

The Elasticsearch watch resource currently writes redacted Watcher action secrets such as `::es_redacted::` into Terraform state during read-after-write and refresh. That causes later updates that do not touch `actions` to resend the redacted placeholder back to Elasticsearch, which fails with a parse error and prevents normal watch maintenance.

## What Changes

- Preserve previously known `actions` values when the Watcher Get API returns redacted string leaves inside the actions JSON.
- Keep non-redacted action fields authoritative from the API so normal refresh behavior still applies.
- Add focused test coverage for watches whose actions include redacted secrets and are later updated through unrelated fields.
- Clarify the watch requirements so refresh behavior for redacted action secrets is explicitly specified.

## Capabilities

### New Capabilities

### Modified Capabilities

- `elasticsearch-watch`: refine read/state synchronization for `actions` when the Watcher API redacts nested secret values

## Impact

- `internal/elasticsearch/watcher/watch/models.go`
- `internal/elasticsearch/watcher/watch/read.go`
- `internal/elasticsearch/watcher/watch/acc_test.go`
- `internal/elasticsearch/watcher/watch/testdata/`
- `openspec/specs/elasticsearch-watch/spec.md`
