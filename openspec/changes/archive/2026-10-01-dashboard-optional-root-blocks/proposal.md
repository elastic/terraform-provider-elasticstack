## Why

`elasticstack_kibana_dashboard` marks the root `time_range`, `refresh_interval`, and `query` blocks as required, but the published Kibana Dashboards API (`kbn-dashboard-data`) only requires `title` on create and update. The requirement comes from the Terraform schema, not from the OpenAPI spec. Practitioners cannot create a title-only dashboard even though Kibana accepts one.

Live probing (Kibana 9.4.2 and 9.6.0-SNAPSHOT) established that Kibana applies **no defaults** to these three blocks: a title-only POST or PUT succeeds and a GET returns the blocks as absent. A PUT is a full replace, so omitting a block on update clears it.

A hidden prerequisite exists: the hand-maintained `PUT /api/dashboards/{id}` operation in `generated/kbapi/dashboard-paths.json` lists `options, query, refresh_interval, time_range, title` as required, so the generated `PutDashboardsIdJSONBody` has non-pointer fields and `dashboardToAPIUpdateRequest` sends zero-value objects for nil blocks. Kibana rejects those with HTTP 400 on `query.language`.

## What Changes

- Make root `time_range`, `refresh_interval`, and `query` **optional** in `internal/kibana/dashboard/schema.go` (`timeRangeSingleNestedAttribute(..., false)` for `time_range`). Nested attributes stay required when a block is configured. No schema version bump and no state upgrader.
- Omit null blocks from both POST and PUT payloads.
- Change the PUT `required` list in `generated/kbapi/dashboard-paths.json` to `["title"]`, regenerate `generated/kbapi/kibana.gen.go` (generated diff in its own commit), and update `dashboardToAPIUpdateRequest` for the resulting pointer fields, including `Options`.
- Read side: plain `Optional` semantics with **no** null-preservation or default-detection logic. Blocks present in the API response populate state; absent blocks are null.
- Update `openspec/specs/kibana-dashboard/spec.md` (schema sketch, REQ-036, REQ-007, REQ-009) with scenarios for omitted blocks and for removal clearing a block.
- Correct the REQ-007 create-method sentence so it defers to REQ-003/REQ-003a (create uses `PUT` when a practitioner-supplied `dashboard_id` is set, otherwise `POST`), so the title-only payload-omission guarantee covers both create paths.
- Examples and docs: existing examples stay explicit; `panel_no_time_range.tf` drops its now-unneeded blocks; add a title-only example; document that removing a block clears it rather than resetting it to a default; regenerate `docs/resources/kibana_dashboard.md`.
- Tests: unit tests for omit-on-write; acceptance tests for title-only create, empty re-plan and import, add-then-remove of each block, and non-default values.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-dashboard`: root `time_range`, `refresh_interval`, and `query` become optional; create/update mapping omits null blocks; read-back requires no null-preservation for these blocks.

## Impact

- `internal/kibana/dashboard/schema.go`, `schema_time_range.go` (caller), `models.go` (`dashboardToAPIUpdateRequest`).
- `generated/kbapi/dashboard-paths.json`, `generated/kbapi/kibana.gen.go` (regenerated).
- `openspec/specs/kibana-dashboard/spec.md`.
- `examples/resources/elasticstack_kibana_dashboard/` and `docs/resources/kibana_dashboard.md`.
- Unit and acceptance tests under `internal/kibana/dashboard/`.
- Non-breaking for existing configurations and state: required to optional only widens the accepted configuration.

## Out of scope

- Broader OpenAPI overlay or path drift beyond the PUT `required` list.
- Dashboard list/search endpoint behavior.
- `query` wire-shape differences (`text` / `json` vs API `expression`).
- Migration strategy (none is needed).
- Kibana 9.5 behavior was not probed; the title-only acceptance test acts as the tripwire.
