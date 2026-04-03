# Proposal: ES|QL Control Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage ES|QL control panels on Kibana dashboards as code today. ES|QL control panels allow dashboard users to dynamically filter and parameterize queries using ES|QL variables — for example, selecting field values, applying time literals, or choosing function arguments that are injected into the dashboard's ES|QL queries at runtime. These controls are essential for interactive, self-service dashboards and are increasingly common in Elastic's observability and security solutions.

Without Terraform support, teams that otherwise manage their dashboards as code must either manually configure these controls in the Kibana UI or accept that their dashboard definitions are incomplete. This creates operational risk (configuration drift, loss of reproducibility) and prevents full infrastructure-as-code workflows for dashboard-heavy teams.

## What Changes

- **Add `esql_control_config` typed panel config block** for panels with `type = "esql_control"`. This block captures all fields from the `kbn-dashboard-panel-esql_control` API schema in a structured, ergonomic way.
- **Add new requirement REQ-026** defining the behavior, required fields, optional fields, and read/write semantics of the `esql_control` panel type.
- **Update REQ-010** to document that `config_json` write support is extended to the `esql_control` panel type, or explicitly document that `esql_control` must be managed through the typed `esql_control_config` block only.
- **Update REQ-006** to include schema-level validation that `esql_control_config` is only valid when `type = "esql_control"`, that `esql_control_config` is mutually exclusive with all other panel config blocks, and that `control_type` and `variable_type` are restricted to their documented enum values.

## Capabilities

After this change, practitioners will be able to:

- Declare `esql_control` panels on a dashboard with full control over the query, variable name, variable type, and control type.
- Statically enumerate `selected_options` for `STATIC_VALUES` controls or configure a `VALUES_FROM_QUERY` control backed by an ES|QL query.
- Optionally configure display settings such as `placeholder`, `hide_action_bar`, `hide_exclude`, `hide_exists`, and `hide_sort`.
- Restrict controls to single-select mode via `single_select`.
- Pre-populate `available_options` for query-backed controls.
- Import and plan-refresh existing `esql_control` panels without losing their configuration.

## Impact

- **Additive only**: no existing panel types or behaviors are changed.
- **Schema change**: adds a new optional `esql_control_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens the schema validation rules to cover the new panel type and config block.
- **REQ-010 update**: clarifies which panel types support `config_json` on write, with a decision to be recorded in design.md on whether `esql_control` is added to that set.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
