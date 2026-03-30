# Proposal: SLO Burn Rate Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage SLO burn rate panels on Kibana dashboards as code today. These panels visualize how quickly an SLO's error budget is being consumed over a configurable look-back window — for example, how fast burn is accumulating over the last 72 hours. Because the panel type is not yet supported by the provider, teams that otherwise manage their dashboards as code must either manually configure burn rate panels in the Kibana UI or accept that their dashboard definitions are incomplete. This creates operational risk (configuration drift, loss of reproducibility) and prevents full infrastructure-as-code workflows for SLO-driven observability dashboards.

## What Changes

- **Add `slo_burn_rate_config` typed panel config block** for panels with `type = "slo_burn_rate"`. This block captures all fields from the `slo-burn-rate-embeddable` API schema in a structured, ergonomic way.
- **Add new requirement REQ-032** defining the behavior, required fields, optional fields, `duration` format validation, `slo_instance_id` default handling, and read/write semantics of the `slo_burn_rate` panel type.
- **Update REQ-006** to document that `slo_burn_rate_config` is only valid when `type = "slo_burn_rate"`, that it is mutually exclusive with all other panel config blocks, and that `duration` is validated at plan time against the format `^\d+[mhd]$`.
- **Update REQ-010** to document that `config_json` write support is not extended to the `slo_burn_rate` panel type; `slo_burn_rate` panels must be managed exclusively through the typed `slo_burn_rate_config` block.

## Capabilities

After this change, practitioners will be able to:

- Declare `slo_burn_rate` panels on a dashboard with full control over the target SLO and look-back duration.
- Specify the SLO by `slo_id` and, when the SLO uses `group_by`, target a specific instance via `slo_instance_id`.
- Configure the burn rate chart look-back window using `duration` in a `[value][unit]` format (e.g. `"5m"`, `"3h"`, `"6d"`).
- Configure optional URL drilldowns via a typed `drilldowns` list that carries practitioner-supplied `url` and `label`, while the provider supplies the fixed API-required `trigger` and `type` values plus optional `encode_url` and `open_in_new_tab`.
- Optionally set panel-level display attributes (`title`, `description`, `hide_title`, `hide_border`).
- Import and plan-refresh existing `slo_burn_rate` panels without losing their configuration.

## Impact

- **Additive only**: no existing panel types or behaviors are changed.
- **Schema change**: adds a new optional `slo_burn_rate_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens the schema validation rules to cover the new panel type, config block, and `duration` format constraint.
- **REQ-010 update**: clarifies that `slo_burn_rate` is explicitly not in the `config_json`-supported set.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
