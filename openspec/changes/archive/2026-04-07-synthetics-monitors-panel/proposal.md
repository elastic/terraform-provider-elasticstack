# Proposal: Synthetics Monitors Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage Synthetics monitors panels on Kibana dashboards as code today. These panels display a filterable table of Elastic Synthetics monitors and their current status, making them essential for synthetic monitoring dashboards that give operations teams real-time visibility into monitor health across projects, locations, tags, and monitor types.

Without Terraform support, teams that otherwise manage their dashboards as code must configure these panels manually in the Kibana UI, creating configuration drift, reducing reproducibility, and preventing full infrastructure-as-code workflows for synthetic monitoring dashboards.

## What Changes

- **Add `synthetics_monitors_config` typed panel config block** for panels with `type = "synthetics_monitors"`. This block is entirely optional — the only field it contains is a `filters` nested block, and all filter sub-fields are themselves optional.
- **Add new requirement REQ-034** defining the behavior, optional fields, and read/write semantics of the `synthetics_monitors` panel type, including filter structure, read-back when config or filters are omitted, and the relationship to the shared filter model established by REQ-033.
- **Update REQ-006** to include schema-level validation that `synthetics_monitors_config` is only valid when `type = "synthetics_monitors"` and that `synthetics_monitors_config` is mutually exclusive with all other panel config blocks.
- **Update REQ-010** to document that `config_json` write is not extended to `synthetics_monitors` panels; they must be managed through the typed `synthetics_monitors_config` block.

## Capabilities

After this change, practitioners will be able to:

- Declare `synthetics_monitors` panels on a dashboard to show a live table of Elastic Synthetics monitors and their current status.
- Optionally filter the monitors table by project, tag, monitor ID, location, monitor type, and status using the `filters` nested block.
- Each filter dimension accepts a list of `{ label, value }` pairs, consistent with the filter model used by the `synthetics_stats_overview` panel (REQ-033).
- Omit the `synthetics_monitors_config` block entirely when no filtering is required, and rely on read-back null preservation to avoid spurious plan diffs.
- Import and plan-refresh existing `synthetics_monitors` panels without losing their configuration.

## Impact

- **Additive only**: no existing panel types or behaviors are changed.
- **Schema change**: adds a new optional `synthetics_monitors_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens schema validation rules to cover the new panel type and config block.
- **REQ-010 update**: documents that `synthetics_monitors` is not added to the `config_json` write-support set; managed exclusively through the typed block.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
- **Files affected**: `internal/kibana/dashboard/schema.go`, `internal/kibana/dashboard/models_panels.go`, new `internal/kibana/dashboard/models_synthetics_monitors_panel.go`, and acceptance tests in `internal/kibana/dashboard/acc_test.go`.
- **Implementation note**: the filter model is structurally identical to that of the `synthetics_stats_overview` panel (REQ-033). The implementation may share filter model types and converter functions between the two panel types.
