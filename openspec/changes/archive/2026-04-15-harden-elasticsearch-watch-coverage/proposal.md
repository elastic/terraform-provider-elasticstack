## Why

The Elasticsearch watch resource currently lacks acceptance coverage for several migration-sensitive paths, including import, defaulted fields, and clearing an optional `transform`. The current read behavior also preserves a stale `transform` value after it has been removed, so the resource can keep incorrect state and mask regressions.

## What Changes

- Add acceptance coverage for watch import, defaulted `active`, defaulted `throttle_period_in_millis`, and removing a previously configured `transform`.
- Fix watch refresh behavior so `transform` is cleared from state when it is no longer configured and Elasticsearch no longer returns it.
- Tighten the watch requirements so default handling, import behavior, and `transform` state synchronization are explicitly specified.

## Capabilities

### New Capabilities

### Modified Capabilities

- `elasticsearch-watch`: clarify import/default expectations and correct refresh behavior for clearing an optional `transform`

## Impact

- `internal/elasticsearch/watcher/watch.go`
- `internal/elasticsearch/watcher/watch_test.go`
- `internal/elasticsearch/watcher/testdata/`
- `openspec/specs/elasticsearch-watch/spec.md`
