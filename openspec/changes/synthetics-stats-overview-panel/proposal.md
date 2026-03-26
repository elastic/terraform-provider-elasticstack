# Proposal: Synthetics Stats Overview Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage Synthetics stats overview panels as code today. These panels provide a high-level summary of Elastic Synthetics monitoring status — aggregating up/down counts, availability percentages, and alert state across monitors — and are essential for embedding synthetic monitoring health into operational dashboards alongside other observability signals.

Without Terraform support, teams that manage their dashboards as code must either configure Synthetics stats overview panels manually in the Kibana UI or omit them from their dashboard definitions entirely. This breaks reproducibility and prevents full infrastructure-as-code workflows for teams using Elastic Synthetics.

## What Changes

- **Add `synthetics_stats_overview_config` typed panel config block** for panels with `type = "synthetics_stats_overview"`. This block captures all fields from the Kibana `synthetics_stats_overview` embeddable API schema in a structured, ergonomic way.
- **Add new requirement REQ-033** defining the behavior, optional fields, `filters` sub-block, `drilldowns` handling, and read/write semantics of the `synthetics_stats_overview` panel type.
- **Update REQ-006** to include schema-level validation that `synthetics_stats_overview_config` is only valid when `type = "synthetics_stats_overview"` and is mutually exclusive with all other panel config blocks.
- **Update REQ-010** to document that `synthetics_stats_overview` must be managed through the typed `synthetics_stats_overview_config` block and is not in the `config_json`-supported set.

## Capabilities

### New Capabilities

- `kibana-dashboard`: practitioners can declare a `synthetics_stats_overview` panel with a fully optional typed config block, including display settings (`title`, `description`, `hide_title`, `hide_border`), URL drilldowns, and per-category Synthetics monitor filters (projects, tags, monitor IDs, locations, monitor types, statuses).
- All fields are optional; an empty or absent `synthetics_stats_overview_config` block produces a valid panel showing all monitors with no pre-filtering.
- Each filter category accepts a list of `{ label, value }` objects, providing human-readable labels alongside the machine-readable filter values.

### Modified Capabilities

- _(none)_

## Impact

- Specs: delta spec under `openspec/changes/synthetics-stats-overview-panel/specs/kibana-dashboard/spec.md`.
- Schema: `internal/kibana/dashboard/schema.go`.
- Models: `internal/kibana/dashboard/models_panels.go` and new `internal/kibana/dashboard/models_synthetics_stats_overview_panel.go`.
- Tests: new acceptance tests in `internal/kibana/dashboard/acc_test.go` (or a dedicated file) and unit tests alongside the converter.
- **Additive only**: no existing panel types or behaviors are changed.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
