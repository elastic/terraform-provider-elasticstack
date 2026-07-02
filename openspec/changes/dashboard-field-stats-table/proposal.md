## Why

The `elasticstack_kibana_dashboard` resource cannot embed a `field_stats_table` panel — the Data Visualizer's field-statistics table, which enumerates fields on a data view or on the output of an ES|QL query with mini distribution charts and summary stats. Users who want to include field-statistics panels in code-managed dashboards must fall back to raw `config_json` blobs, losing typed validation and drift-safe planning. The Kibana Dashboard API defines a first-class `field_stats_table` panel type with a well-defined config union.

## What Changes

Add a new typed panel block `field_stats_table_config` to the `elasticstack_kibana_dashboard` resource with two mutually exclusive sub-blocks matching the API union:

- `by_dataview` — backed by a Kibana data view (`data_view_id`)
- `by_esql` — backed by an ES|QL query string (`query`)

Both branches share optional attributes `show_distributions`, `title`, `description`, `hide_title`, `hide_border`, and `time_range`, mirroring the API layout and the union patterns already established by `vis_config.by_value`/`by_reference` and `discover_session_config.by_value`/`by_reference`.

Branch keys match the API's `view_type` discriminator strings (`"dataview"` and `"esql"`) exactly. The `view_type` field is not exposed as a user-facing attribute — the branch identity already carries that information, and the model layer sets it on the wire.

## Capabilities

### New Capabilities

- `kibana-dashboard`: new requirement REQ-047 — `field_stats_table_config` typed panel block supporting `by_dataview` and `by_esql` branches.

### Modified Capabilities

- `kibana-dashboard` REQ-010 — extend the `config_json` rejection guard to include `field_stats_table` in the list of panel types that must be managed via their typed block, not via `config_json`.

## Impact

- `internal/kibana/dashboard/panel/fieldstatstable/` — new package with `schema.go`, `model.go`, `api.go`, `fromapi.go`.
- `internal/kibana/dashboard/registry.go` — register the new handler in `panelHandlers`.
- `internal/kibana/dashboard/schema.go` — add `field_stats_table_config` attribute to the panel schema.
- `openspec/changes/dashboard-field-stats-table/specs/kibana-dashboard/spec.md` — delta spec adding REQ-047 and updating REQ-010.
- Acceptance tests in `internal/kibana/dashboard/panel/fieldstatstable/acc_test.go` covering `by_dataview`, `by_esql`, validator rejection, and drift detection.
