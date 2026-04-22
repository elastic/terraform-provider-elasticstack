# Proposal: SLO Overview Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage SLO overview panels on Kibana dashboards as code today. These panels are essential for SLO-driven dashboards that display service reliability status, allowing teams to monitor SLO health at a glance in both single-SLO and group-overview modes. Without Terraform support, teams that otherwise manage their dashboards as code must configure these panels manually in the Kibana UI, creating configuration drift, reducing reproducibility, and preventing full infrastructure-as-code workflows for observability dashboards.

The SLO overview panel is especially important in reliability engineering and observability contexts: single-SLO mode surfaces status for one specific SLO with optional instance scoping and drilldown links, while groups mode provides a combined view of multiple SLOs grouped by status, tags, indicator type, or index.

## What Changes

- **Add `slo_overview_config` typed panel config block** for panels with `type = "slo_overview"`. This block provides two mutually exclusive nested config blocks: `single` (maps to `slo-single-overview-embeddable`) and `groups` (maps to `slo-group-overview-embeddable`), discriminated by the `overview_mode` field in the API payload.
- **Add new requirement REQ-030** defining the behavior, required fields, optional fields, and read/write semantics of the `slo_overview` panel type, including mode selection, field validation, drilldowns handling, and group filter semantics.
- **Update REQ-006** to include schema-level validation that `slo_overview_config` is only valid when `type = "slo_overview"`, that `slo_overview_config` is mutually exclusive with all other panel config blocks, that exactly one of `single` or `groups` must be set, that `single.slo_id` is required when `single` is configured, and that `groups.group_filters.group_by` is restricted to its documented enum values.
- **Update REQ-010** to document that `config_json` write is not extended to `slo_overview` panels; they must be managed through the typed `slo_overview_config` block.

## Capabilities

After this change, practitioners will be able to:

- Declare `slo_overview` panels in single mode, referencing a specific SLO by `slo_id` with optional instance scoping via `slo_instance_id` (defaulting to `"*"` for all instances) and optional remote cluster targeting via `remote_name`.
- Declare `slo_overview` panels in groups mode, optionally configuring `group_filters` to control grouping via `group_by` enum, targeted `groups` list, KQL filter via `kql_query`, and complex AS-code filters via `filters_json`.
- Configure URL drilldowns on either mode via a typed `drilldowns` list with `url`, `label`, `trigger`, and `type` required fields, and optional `encode_url` and `open_in_new_tab`.
- Set per-panel display options (`title`, `description`, `hide_title`, `hide_border`) in either mode.
- Import and plan-refresh existing `slo_overview` panels without losing their configuration.

## Impact

- **Additive only**: no existing panel types or behaviors are changed.
- **Schema change**: adds a new optional `slo_overview_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens schema validation rules to cover the new panel type and config block, including mutual exclusion of `single` and `groups` sub-blocks and the `single.slo_id` required field constraint.
- **REQ-010 update**: documents that `slo_overview` is not added to the `config_json` write-support set; managed exclusively through the typed block.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
- **Files affected**: `internal/kibana/dashboard/schema.go`, `internal/kibana/dashboard/models_panels.go`, new `internal/kibana/dashboard/models_slo_overview_panel.go`, and acceptance tests in `internal/kibana/dashboard/acc_test.go`.
