## REMOVED Requirements

### Requirement: `lens-dashboard-app` panel behavior (REQ-035)

Remove REQ-035 entirely. The `lens-dashboard-app` panel type was included in the Kibana Dashboard API spec by mistake and has been removed upstream. All configurations that used `type = "lens-dashboard-app"` must be migrated to `type = "vis"`. The existing unknown-panel fallback (`config_json`) handles read-time gracefully for any Kibana dashboards that still have `lens-dashboard-app` panels at the API level.

## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

The `lens_dashboard_app_config` block SHALL no longer be a recognized panel configuration block. The provider SHALL reject any configuration that includes `lens_dashboard_app_config` at plan time. The `panelTypeAliases` entry and schema registration for `"lens-dashboard-app"` SHALL be removed; attempting to set `type = "lens-dashboard-app"` SHALL be treated as an unsupported panel type.

#### Scenario: lens_dashboard_app_config rejected

- GIVEN a panel configuration that contains a `lens_dashboard_app_config` block
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected at plan time (the attribute no longer exists in the schema)

#### Scenario: type lens-dashboard-app no longer valid

- GIVEN a panel with `type = "lens-dashboard-app"` and no `lens_dashboard_app_config` block
- WHEN Terraform validates the resource schema
- THEN the provider SHALL treat the panel as an unsupported type and SHALL NOT attempt to use the removed `lensdashboardapp` handler

### Requirement: Raw `config_json` panel behavior (REQ-025)

The restriction on `config_json` for `lens-dashboard-app` panels is removed. `lens-dashboard-app` is no longer a recognized panel type with a typed block; the config_json type allowlist SHALL no longer include a rejection rule for `lens-dashboard-app`. All other `config_json` restrictions (e.g., for `options_list_control` and `synthetics_monitors`) remain in force.

#### Scenario: config_json no longer explicitly rejected for lens-dashboard-app

- GIVEN a panel with an unknown or unrecognized `type` value (including a Kibana-internal type such as `lens-dashboard-app`)
- WHEN the provider reads such a panel back from the Kibana API
- THEN the provider SHALL use the unknown-panel fallback and SHALL populate `config_json` in state
- AND SHALL NOT return an error diagnostic for the unrecognized panel type

### Requirement: Chart-level `time_range` null-preservation and inheritance from dashboard (REQ-040)

The chart-level `time_range` null-preservation rule applies only to typed Lens chart blocks reachable under `panels[].vis_config.by_value.<chart>_config` (for `type = "vis"`). The `panels[].lens_dashboard_app_config.by_value.<chart>_config` path SHALL no longer exist and SHALL NOT be referenced in this requirement. All other aspects of REQ-040 remain unchanged.

#### Scenario: time_range null-preservation applies only to vis_config path

- GIVEN a `vis` panel with a typed Lens chart block under `vis_config.by_value`
- AND the chart block has `time_range = null` in prior state
- AND the API-returned chart-level `time_range` equals the dashboard-level `time_range`
- WHEN the provider refreshes state
- THEN the provider SHALL preserve `time_range = null` in state for that chart block
- AND SHALL NOT produce a spurious diff for `time_range`
