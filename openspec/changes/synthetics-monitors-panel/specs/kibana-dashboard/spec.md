# Delta Spec: Synthetics Monitors Panel Support

Base spec: `openspec/specs/kibana-dashboard/spec.md`
Last requirement in base spec: REQ-025
This proposal uses: REQ-034

---

## Schema additions

The following block is added to the panel object within the `panels` list (and within `sections[*].panels`):

```hcl
synthetics_monitors_config = <optional, object({
  filters = <optional, object({
    projects      = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    tags          = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    monitor_ids   = <optional, list(object({ label = <required, string>, value = <required, string> }))> # max 5000 items (API constraint)
    locations     = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    monitor_types = <optional, list(object({ label = <required, string>, value = <required, string> }))>
    statuses      = <optional, list(object({ label = <required, string>, value = <required, string> }))>
  })>
})> # only with type = "synthetics_monitors"; conflicts with all other config blocks
```

Notes:

- The entire `synthetics_monitors_config` block is optional. A `synthetics_monitors` panel with no filtering requirements may be declared with no config block.
- The `filters` block within `synthetics_monitors_config` is optional.
- Each filter dimension (`projects`, `tags`, `monitor_ids`, `locations`, `monitor_types`, `statuses`) is an optional list of `{ label, value }` pairs.
- The filter structure is identical to that defined for `synthetics_stats_overview` (REQ-033). The implementation SHOULD share filter model types and converters between the two panel types.

---

## MODIFIED Requirements

### Requirement: Replacement fields and schema validation (REQ-006)

Schema validation SHALL enforce that `synthetics_monitors_config` is valid only for panels with `type = "synthetics_monitors"` and is mutually exclusive with all other panel configuration blocks and with `config_json`.

The existing REQ-006 text is extended with the following additions:

- `synthetics_monitors_config` SHALL be valid only for panels with `type = "synthetics_monitors"`.
- `synthetics_monitors_config` SHALL be mutually exclusive with all other panel configuration blocks and with `config_json`.

#### Scenario: synthetics_monitors_config rejected for non-synthetics_monitors panel (ADDED)

- GIVEN a panel with `type = "lens"` and `synthetics_monitors_config` set
- WHEN Terraform validates the resource schema
- THEN the configuration SHALL be rejected before any dashboard API call

#### Scenario: synthetics_monitors_config conflicts with other typed blocks (ADDED)

- GIVEN a panel entry with `type = "synthetics_monitors"` that sets both `synthetics_monitors_config` and any other typed config block (e.g. `markdown_config`)
- WHEN Terraform validates the resource schema
- THEN the provider SHALL return an error diagnostic indicating the conflicting blocks are mutually exclusive

---

### Requirement: Panels and `config_json` round-trip behavior (REQ-010)

`config_json` SHALL NOT be supported for `synthetics_monitors` panels; the `synthetics_monitors` panel type SHALL be managed exclusively through the typed `synthetics_monitors_config` block.

The existing REQ-010 text:

> On write, `config_json` SHALL be supported only for `markdown` and `lens` panel types; using `config_json` with any other panel type, or omitting all panel configuration blocks, SHALL return an error diagnostic.

is updated to additionally state:

> The `synthetics_monitors` panel type SHALL be managed exclusively through the typed `synthetics_monitors_config` block; using `config_json` with `type = "synthetics_monitors"` SHALL return an error diagnostic.

#### Scenario: config_json rejected for synthetics_monitors panel type (ADDED)

- GIVEN a panel with `type = "synthetics_monitors"` configured through `config_json`
- WHEN the provider builds the API request on create or update
- THEN it SHALL return an error diagnostic stating that `config_json` is not supported for `synthetics_monitors`

---

## ADDED Requirements

### Requirement: Synthetics monitors panel behavior (REQ-034)

For `type = "synthetics_monitors"` panels, the resource SHALL accept an optional `synthetics_monitors_config` block. The block, if present, may contain an optional `filters` nested block. Within `filters`, all six filter dimensions (`projects`, `tags`, `monitor_ids`, `locations`, `monitor_types`, `statuses`) are optional lists of `{ label, value }` objects.

The `synthetics_monitors` panel type is a standalone panel, not a Lens visualization. It does not reference a saved object, and its configuration is fully inline in the dashboard document. None of the Lens panel converters, Lens time-range injection, or Lens metric default normalization SHALL apply to `synthetics_monitors` panels.

**On write (create and update):**

When `synthetics_monitors_config` is set, the resource SHALL map the config block to the panel's `config` object in the API request. When the `filters` block is set, the resource SHALL include the `filters` sub-object with only the filter dimensions that are explicitly configured. Filter dimensions that are not set SHALL be omitted from the API request rather than sent as empty arrays. When `synthetics_monitors_config` is omitted entirely, the resource SHALL send an empty `config` object `{}` or omit `config` from the panel payload, consistent with how other all-optional panel config blocks are handled.

**On read:**

When Kibana returns a `synthetics_monitors` panel with an empty or absent `config` object, the provider SHALL keep `synthetics_monitors_config` null in state. When Kibana returns a present `config` with an empty or absent `filters` object, the provider SHALL keep the `filters` block null in state. When Kibana returns individual filter dimension arrays that are empty, the provider SHALL treat them as equivalent to omitted dimensions and SHALL NOT force empty lists into state.

The provider SHALL seed `synthetics_monitors_config` from prior state or plan on read-back, so that filter dimensions omitted by Kibana do not overwrite Terraform-authored values with null.

**Shared filter model:**

The filter structure used by `synthetics_monitors_config` (lists of `{ label, value }` pairs for each filter dimension) is identical to the filter structure used by `synthetics_stats_overview_config` (REQ-033). The implementation SHOULD share filter model types and converter functions between the two panel types to avoid duplication.

#### Scenario: Synthetics monitors panel with no config block

- GIVEN a dashboard configuration containing a `synthetics_monitors` panel with no `synthetics_monitors_config` block
- WHEN the resource is created
- THEN the provider SHALL send a valid API request for the panel without a populated `config` object
- AND state SHALL record `synthetics_monitors_config` as null
- AND a subsequent plan SHALL show no changes

#### Scenario: Synthetics monitors panel with filters

- GIVEN a dashboard configuration containing a `synthetics_monitors` panel with:
  - `type = "synthetics_monitors"`
  - `synthetics_monitors_config.filters.projects = [{ label = "My Project", value = "my-project" }]`
  - `synthetics_monitors_config.filters.statuses = [{ label = "Up", value = "up" }, { label = "Down", value = "down" }]`
- WHEN the resource is created
- THEN the provider SHALL send the mapped `config.filters` object to the Kibana dashboard API with the `projects` and `statuses` dimensions populated
- AND the panel SHALL appear in state with those filter dimensions populated
- AND omitted filter dimensions (`tags`, `monitor_ids`, `locations`, `monitor_types`) SHALL remain null in state

#### Scenario: Read-back null preservation when config is empty

- GIVEN a managed `synthetics_monitors` panel whose `synthetics_monitors_config` is null (no config block)
- WHEN Kibana returns the panel with an empty `config` object `{}`
- THEN the provider SHALL keep `synthetics_monitors_config` null in state
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: Read-back null preservation when filters is empty

- GIVEN a managed `synthetics_monitors` panel with `synthetics_monitors_config` set but `filters` omitted
- WHEN Kibana returns the panel with a `config` containing an empty `filters` object `{}`
- THEN the provider SHALL keep the `filters` block null in state
- AND SHALL NOT create a spurious diff on the next plan

#### Scenario: All filter dimensions set

- GIVEN a `synthetics_monitors` panel with all six filter dimensions configured in `filters`
- WHEN the resource is created and read back
- THEN all six filter dimensions SHALL be present in state
- AND a subsequent plan SHALL show no changes

#### Scenario: monitor_ids large list (API constraint documentation)

- GIVEN a `synthetics_monitors` panel with `filters.monitor_ids` containing more than 5000 items
- WHEN the provider sends the API request
- THEN the API MAY return an error; the provider SHALL surface that error as a diagnostic
- AND the provider SHALL NOT enforce a plan-time validator for the 5000-item limit (this is an API-side constraint)
