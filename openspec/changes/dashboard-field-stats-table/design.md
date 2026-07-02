## Context

The Kibana Dashboard API defines panel type `field_stats_table` (API type key: `KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable`). Its `config` field is `KibanaHTTPAPIsDataVisualizerFieldStats`, a discriminated union with `view_type` as the discriminator. Two branches exist:

- `KibanaHTTPAPIsDataVisualizerFieldStats0` (`view_type = "dataview"`): requires `data_view_id` (string).
- `KibanaHTTPAPIsDataVisualizerFieldStats1` (`view_type = "esql"`): requires `query.esql` (string).

Both branches share:
- `show_distributions` (bool, optional)
- `title` (string, optional)
- `description` (string, optional)
- `hide_title` (bool, optional)
- `time_range` (object `{ from, to, mode? }`, optional)
Relevant code locations:
- `generated/kbapi/kibana.gen.go`: `KibanaHTTPAPIsKbnDashboardPanelTypeFieldStatsTable` (line ~47051), `KibanaHTTPAPIsDataVisualizerFieldStats` (line ~40635).
- `internal/kibana/dashboard/schema.go` — add `field_stats_table_config`.
- `internal/kibana/dashboard/registry.go` — register the panel handler.
- Sibling union patterns: `internal/kibana/dashboard/panel/visconfig/`, `internal/kibana/dashboard/panel/discoversession/`.

## Goals / Non-Goals

**Goals:**
- Expose a fully typed `field_stats_table_config` block that maps cleanly to both API branches.
- Enforce exactly-one-of semantics on `by_dataview` / `by_esql` at plan time.
- Use panelkit passthroughs for `title`, `description`, `hide_title`, `hide_border`, and `time_range`.
- Apply REQ-009 null-preservation for `time_range` and other optional fields.
- Reject `config_json` on `type = "field_stats_table"` panels.

**Non-Goals:**
- Panel-level drilldowns — the API model does not expose them.
- Shared branch-scaffolding helpers with the P1 controls-esql restructure (deferred per issue decision).
- `data_view_id` existence validation at plan time (runtime Kibana errors surface as today).

## Decisions

**Branch naming**: `by_dataview` and `by_esql`, matching the API `view_type` discriminator strings (`"dataview"` / `"esql"`) exactly. Rationale per issue decision: `by_data_view` / `by_esql` was rejected (diverges from API discriminator); `by_field` / `by_esql` was rejected (misleading — the panel enumerates all fields, not a single field).

**Shared attributes inside each branch, not hoisted**: `show_distributions`, `title`, `description`, `hide_title`, `hide_border`, and `time_range` reside inside each branch block. This matches the API layout and the `vis_config` / `discover_session_config` patterns.

**`view_type` not user-facing**: The branch identity (`by_dataview` vs `by_esql`) determines the wire `view_type`. The provider sets it internally; practitioners do not write it in HCL.

**Handler package**: `internal/kibana/dashboard/panel/fieldstatstable/` following the existing naming convention.

**Exactly-one-of validator**: Applied at the `field_stats_table_config` object level, mirroring the discover session validator pattern.

**`config_json` rejected**: `field_stats_table` added to the list of panel types that must use their typed config block. Error diagnostic matches the pattern used for `discover_session`.

**Ships independently of P1 controls-esql (#4001)**: No state migration needed (panel doesn't exist yet), no blocking dependency. Union pattern (`by_X` / `by_esql`) is well-established.

## Risks / Trade-offs

- [Risk] The `query.esql` API nesting (`query.esql` not just `esql`) requires careful model mapping. Mitigation: implement ToAPI and FromAPI with explicit nesting; cover with unit tests.
- [Risk] `time_range` null-preservation on read requires care (REQ-009 semantics). Mitigation: apply the same null-preservation guard already used by other panels; unit-test the null-preservation path.
- [Low risk] `show_distributions` default behavior in Kibana is unspecified here. Mitigation: treat as optional bool with null-preservation; do not bake Kibana defaults into the TF schema.

## Open questions

None — all design decisions are resolved per the human direction in the issue comments.
