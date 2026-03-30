# Delta Spec: SLO Burn Rate Panel Support

Base spec: `openspec/specs/kibana-dashboard/spec.md`
Last requirement in base spec: REQ-025

---

## Schema additions

The following block is added to the panel object within the `panels` list (and within `sections[*].panels`):

```hcl
slo_burn_rate_config = <optional, object({
  # Required
  slo_id   = <required, string>
  duration = <required, string>  # format: [value][unit] where unit is m, h, or d — validated as ^\d+[mhd]$

  # Optional
  slo_instance_id = <optional, string>  # API default "*"; provider preserves null when not configured
  title           = <optional, string>
  description     = <optional, string>
  hide_title      = <optional, bool>
  hide_border     = <optional, bool>

  drilldowns = <optional, list(object({
    url             = <required, string>
    label           = <required, string>
    trigger         = <required, string>  # const "on_open_panel_menu"
    type            = <required, string>  # const "url_drilldown"
    encode_url      = <optional, bool>
    open_in_new_tab = <optional, bool>
  }))>
})> # only with type = "slo_burn_rate"; conflicts with all other config blocks
```

---

## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that `slo_burn_rate_config` is valid only for panels with `type = "slo_burn_rate"`, is mutually exclusive with all other panel configuration blocks, and that the `duration` attribute matches the pattern `^\d+[mhd]$`.

The existing REQ-006 text is extended. The sentence:

> Each panel SHALL declare at least one panel configuration block, panel configuration blocks SHALL be mutually exclusive, typed panel configuration blocks SHALL only be valid for their supported panel type, and `waffle_config` SHALL enforce its ES|QL-vs-non-ES|QL field consistency rules.

gains the following additions:

- `slo_burn_rate_config` SHALL be valid only for panels with `type = "slo_burn_rate"`.
- `slo_burn_rate_config` SHALL be mutually exclusive with all other panel configuration blocks.
- The `duration` attribute within `slo_burn_rate_config` SHALL match the pattern `^\d+[mhd]$` (a positive integer followed by `m`, `h`, or `d`); any other value SHALL be rejected at plan time.

#### Scenario: slo_burn_rate_config rejected for non-slo_burn_rate panel (ADDED)

- GIVEN a panel with `type = "lens"` and `slo_burn_rate_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: Invalid duration value rejected at plan time (ADDED)

- GIVEN a panel with `type = "slo_burn_rate"` and `slo_burn_rate_config.duration = "5x"`
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected at plan time with a diagnostic indicating the required format

---

### Requirement: Panels and `config_json` round-trip behavior (REQ-010)

`config_json` SHALL NOT be supported for `slo_burn_rate` panels; the `slo_burn_rate` panel type SHALL be managed exclusively through the typed `slo_burn_rate_config` block.

The existing REQ-010 text:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, including `slo_burn_rate`, or omitting all panel configuration blocks, SHALL return an error diagnostic. The `slo_burn_rate` panel type SHALL be managed exclusively through the typed `slo_burn_rate_config` block.

#### Scenario: config_json rejected for slo_burn_rate panel type (ADDED)

- GIVEN a panel with `type = "slo_burn_rate"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `slo_burn_rate`

---

## ADDED Requirements

### Requirement: SLO burn rate panel behavior (REQ-032)

For `type = "slo_burn_rate"` panels, the resource SHALL accept `slo_burn_rate_config` with the required fields `slo_id` and `duration`, and the optional fields `slo_instance_id`, `title`, `description`, `hide_title`, `hide_border`, and `drilldowns`.

The `duration` field SHALL be validated at plan time against the pattern `^\d+[mhd]$`. Any value that does not match SHALL be rejected before any dashboard API call.

On write (create and update), the resource SHALL map `slo_burn_rate_config` to the `config` object in the `slo-burn-rate-embeddable` API schema. The required fields `slo_id` and `duration` SHALL always be included in the API request. Optional fields SHALL be included only when they are set in Terraform state; absent optional fields SHALL NOT be sent to the API. When `drilldowns` is set, each drilldown object SHALL include all four required attributes (`url`, `label`, `trigger`, `type`) and the optional attributes (`encode_url`, `open_in_new_tab`) only when explicitly set.

On read, the resource SHALL repopulate `slo_burn_rate_config` from the API response. Fields that the API response omits SHALL not be forced into state. The `slo_instance_id` field SHALL use null-preservation: if the prior state value for `slo_instance_id` was null (i.e. the practitioner did not configure it) and the API returns `"*"`, the provider SHALL keep `slo_instance_id` as null in state rather than introducing the API sentinel. When `slo_instance_id` is explicitly configured to `"*"`, the provider SHALL round-trip it normally.

The `slo_burn_rate` panel type is a standalone embeddable, not a Lens visualization. It does not reference a Lens saved object, and its configuration is fully inline in the dashboard document. As a result, none of the Lens panel converters, Lens time-range injection, or Lens metric default normalization SHALL apply to `slo_burn_rate` panels.

#### Scenario: Creation of slo_burn_rate panel with required fields

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel with:
  - `type = "slo_burn_rate"`
  - `slo_burn_rate_config.slo_id = "my-slo-id"`
  - `slo_burn_rate_config.duration = "72h"`
- WHEN the resource is created
- THEN the provider SHALL send the mapped `config` object to the Kibana dashboard API with `slo_id` and `duration`
- AND the panel SHALL appear in state with both required fields populated
- AND `slo_instance_id` SHALL be null in state
- AND the provider SHALL NOT populate `config_json` for this panel in state

#### Scenario: slo_instance_id null-preservation after read-back

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel that does not set `slo_instance_id`
- WHEN the resource is created and then read back from Kibana
- AND Kibana returns `slo_instance_id = "*"` in the API response
- THEN the provider SHALL keep `slo_instance_id` as null in state
- AND a subsequent plan SHALL show no changes

#### Scenario: Creation of slo_burn_rate panel with slo_instance_id and drilldowns

- GIVEN a dashboard configuration containing an `slo_burn_rate` panel with:
  - `slo_burn_rate_config.slo_id = "my-slo-id"`
  - `slo_burn_rate_config.duration = "6d"`
  - `slo_burn_rate_config.slo_instance_id = "host-a"`
  - a `drilldowns` entry with `url = "https://example.com"`, `label = "View details"`, `trigger = "on_open_panel_menu"`, `type = "url_drilldown"`
- WHEN the resource is created and read back
- THEN all configured attributes SHALL be present in state and a subsequent plan SHALL show no changes
