## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

REQ-006 SHALL be extended to include the following rules for the `slo_burn_rate_config` block:

- `slo_burn_rate_config` SHALL be valid only for panels with `type = "slo_burn_rate"`.
- `slo_burn_rate_config` SHALL be mutually exclusive with all other panel configuration blocks and with `config_json`.
- The `duration` attribute within `slo_burn_rate_config` SHALL match the pattern `^\d+[mhd]$` (a positive integer followed by `m`, `h`, or `d`); any other value SHALL be rejected at plan time.

#### Scenario: slo_burn_rate_config rejected for non-slo_burn_rate panel

- GIVEN a panel with `type = "lens"` and `slo_burn_rate_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: Invalid duration value rejected at plan time

- GIVEN a panel with `type = "slo_burn_rate"` and `slo_burn_rate_config.duration = "5x"`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating the required format

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

REQ-010 SHALL be updated so that `slo_burn_rate` is explicitly excluded from `config_json` support. The existing REQ-010 text:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to additionally state:

> The `slo_burn_rate` panel type SHALL remain outside the `config_json`-supported set; using `config_json` with `type = "slo_burn_rate"` SHALL return the standard diagnostic that `config_json` is supported only for `markdown` and `lens` panel types.

#### Scenario: config_json rejected for slo_burn_rate panel type

- GIVEN a panel with `type = "slo_burn_rate"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is supported only for `markdown` and `lens` panel types

## ADDED Requirements

### Requirement: SLO burn rate panel behavior (REQ-032)

The resource SHALL support `type = "slo_burn_rate"` panels through the typed `slo_burn_rate_config` block. The block requires `slo_id` and `duration`, and optionally accepts `slo_instance_id`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns`.

The `duration` field SHALL be validated at plan time against the pattern `^\d+[mhd]$`. Any value that does not match SHALL be rejected before any dashboard API call.

On write, the provider SHALL map `slo_burn_rate_config` to the `config` object in the `slo-burn-rate-embeddable` API schema. Optional fields SHALL be included only when set; absent optional fields SHALL NOT be sent to the API. When `drilldowns` is set, each drilldown object SHALL accept practitioner-configured `url` and `label`, SHALL hardcode `trigger = "on_open_panel_menu"` and `type = "url_drilldown"` in the API request, and SHALL include the optional attributes (`encode_url`, `open_in_new_tab`) only when explicitly set.

On read, the `slo_instance_id` field SHALL use null-preservation: if the prior state value was null and the API returns `"*"`, the provider SHALL keep `slo_instance_id` null in state rather than introducing the API sentinel. When `slo_instance_id` is explicitly configured to `"*"`, the provider SHALL round-trip it normally.

#### Scenario: Creation of slo_burn_rate panel with required fields

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel with `slo_burn_rate_config.slo_id = "my-slo-id"` and `slo_burn_rate_config.duration = "72h"`
- WHEN the resource is created
- THEN the provider SHALL send the mapped `config` object to the Kibana dashboard API with `slo_id` and `duration`
- AND `slo_instance_id` SHALL be null in state

#### Scenario: slo_instance_id null-preservation after read-back

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel that does not set `slo_instance_id`
- WHEN the resource is created and then read back from Kibana
- AND Kibana returns `slo_instance_id = "*"` in the API response
- THEN the provider SHALL keep `slo_instance_id` as null in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Creation of slo_burn_rate panel with slo_instance_id and drilldowns

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel with `slo_burn_rate_config.slo_instance_id = "host-a"` and a drilldown entry
- WHEN the resource is created and read back
- THEN all configured attributes SHALL be present in state and a subsequent plan SHALL show no changes
