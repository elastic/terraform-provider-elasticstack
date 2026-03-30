## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that `slo_error_budget_config` is valid only for panels with `type = "slo_error_budget"`, is mutually exclusive with all other panel configuration blocks and with `config_json`, and that `drilldowns` entries use only the allowed constant values for `trigger` and `type`.

REQ-006 is extended to include:

- `slo_error_budget_config` SHALL be valid only for panels with `type = "slo_error_budget"`.
- `slo_error_budget_config` SHALL be mutually exclusive with all other panel configuration blocks and with `config_json`.
- The `trigger` attribute within each `drilldowns` entry SHALL be restricted to the value `"on_open_panel_menu"`; any other value SHALL be rejected at plan time.
- The `type` attribute within each `drilldowns` entry SHALL be restricted to the value `"url_drilldown"`; any other value SHALL be rejected at plan time.

#### Scenario: slo_error_budget_config rejected for non-slo_error_budget panel

- GIVEN a panel with `type = "lens"` and `slo_error_budget_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

### Requirement: Panels, sections, and `config_json` round-trip behavior (REQ-010)

`config_json` SHALL NOT be supported for `slo_error_budget` panels; the `slo_error_budget` panel type SHALL be managed exclusively through the typed `slo_error_budget_config` block.

The existing REQ-010 text:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to additionally state:

> The `slo_error_budget` panel type SHALL be managed exclusively through the typed `slo_error_budget_config` block; using `config_json` with `type = "slo_error_budget"` SHALL return an error diagnostic.

#### Scenario: config_json rejected for slo_error_budget panel type

- GIVEN a panel with `type = "slo_error_budget"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `slo_error_budget`

---

## ADDED Requirements

### Requirement: SLO error budget panel behavior (REQ-031)

For `type = "slo_error_budget"` panels, the resource SHALL accept a typed `slo_error_budget_config` block containing the fields of the `slo-error-budget-embeddable` API schema. `slo_id` SHALL be required. `slo_instance_id`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns` SHALL be optional. `slo_error_budget_config` SHALL be mutually exclusive with all other typed panel config blocks and with `config_json`.

On write, the provider SHALL map all configured fields from `slo_error_budget_config` into the Kibana dashboard panel API request for the `slo_error_budget` embeddable type.

On read, the provider SHALL repopulate `slo_error_budget_config` from the API response. For `slo_instance_id`, the provider SHALL preserve the prior Terraform state value when the prior value was null: if the practitioner did not configure `slo_instance_id`, the provider SHALL NOT write the API-returned default `"*"` into state. For `encode_url` and `open_in_new_tab` drilldown fields, the provider SHALL normalize the API default value of `true` so that practitioners who omit those fields do not observe spurious drift after apply.

`drilldowns` SHALL be represented as a list of typed objects. Each drilldown object SHALL contain required `url` (string), `label` (string), `trigger` (string, must be `"on_open_panel_menu"`), and `type` (string, must be `"url_drilldown"`), and optional `encode_url` (bool, default `true`) and `open_in_new_tab` (bool, default `true`). The `trigger` and `type` attributes SHALL be validated at schema level to accept only their documented constant values.

#### Scenario: Minimal slo_error_budget panel with only slo_id

- GIVEN a panel with `type = "slo_error_budget"` and `slo_error_budget_config { slo_id = "my-slo-id" }`
- WHEN create and subsequent read run
- THEN the provider SHALL send `slo_id = "my-slo-id"` in the API request and SHALL read it back into state without error

#### Scenario: slo_instance_id null preservation

- GIVEN a panel with `type = "slo_error_budget"` and `slo_error_budget_config` that omits `slo_instance_id`
- WHEN the dashboard is created and subsequently read back from Kibana
- THEN the provider SHALL keep `slo_instance_id` null in state even if Kibana returns `"*"` as the default value
- AND a subsequent plan SHALL show no changes for `slo_instance_id`

#### Scenario: drilldowns configuration

- GIVEN a panel with `type = "slo_error_budget"` and `slo_error_budget_config` containing a `drilldowns` block with `url`, `label`, `trigger = "on_open_panel_menu"`, and `type = "url_drilldown"`
- WHEN the dashboard is created and subsequently read back from Kibana
- THEN the provider SHALL round-trip all drilldown fields and SHALL apply default normalization for `encode_url` and `open_in_new_tab` so that omitting them in configuration does not produce drift
