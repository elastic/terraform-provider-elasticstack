## Why

The Kibana Dashboard API `kbn-dashboard-data` schema exposes a top-level `filters` array (discriminated union of `is`, `is_one_of`, `range`, `exists`, group, DSL, and spatial filter shapes) that lets dashboards persist saved filter pills shown above all panels. The Terraform resource maps `query` but not `filters`, so any saved filter present on a dashboard is dropped on read and absent on write — practitioners cannot manage one of Kibana's most-used dashboard primitives.

## What Changes

- Add an optional `filters` attribute at the dashboard root: `filters = list(object({ filter_json = string }))`.
- `filter_json` carries one normalized JSON object per saved filter, matching the API discriminated union shape (consistent with the existing per-panel `filter_json` pattern used in the typed chart blocks).
- Wire `filters` into create, update, and read paths; round-trip with semantic JSON equality so Kibana-injected defaults do not produce diffs.
- Preserve null/empty distinction: an unset `filters` attribute stays unset on read when Kibana returns no filters.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-dashboard`: add a new requirement covering dashboard-root `filters` round-trip.

## Impact

- `internal/kibana/dashboard/schema.go` — add `filters` block to the dashboard schema.
- `internal/kibana/dashboard/models.go` — extend `dashboardModel` with `Filters` and map to/from `Filters` on the API request/response.
- New unit tests for filter normalization and the unset/empty distinction.
- New acceptance test creating a dashboard with multiple saved filters and verifying round-trip.
- No effect on per-panel `filter_json` blocks (which remain unchanged).
