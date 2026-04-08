# Delta Spec: Pie Chart Structured Legend

Base spec: `openspec/specs/kibana-dashboard/spec.md`
Last requirement in base spec: REQ-036
This proposal modifies: REQ-023

---

## Schema changes

The `pie_chart_config` block in the panel object within `panels` (and within `sections[*].panels`) changes from:

```hcl
pie_chart_config = <optional, object({
  title                 = <optional, string>
  description           = <optional, string>
  dataset_json          = <optional, json string, normalized>
  query                 = <optional, object({ language = <optional, string>, query = <required, string> })>
  filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
  ignore_global_filters = <optional, computed, bool>
  sampling              = <optional, computed, float64>
  donut_hole            = <optional, string>
  label_position        = <optional, string>
  legend_json           = <optional, json string, normalized>
  metrics               = <required, list(object({ config = <required, json string with defaults> }))>
  group_by              = <optional, list(object({ config = <required, json string with defaults> }))>
})>
```

to:

```hcl
pie_chart_config = <optional, object({
  title                 = <optional, string>
  description           = <optional, string>
  dataset_json          = <optional, json string, normalized>
  query                 = <optional, object({ language = <optional, string>, query = <required, string> })>
  filters               = <optional, list(object({ filter_json = <required, json string, normalized> }))>
  ignore_global_filters = <optional, computed, bool>
  sampling              = <optional, computed, float64>
  donut_hole            = <optional, string>
  label_position        = <optional, string>
  legend = <optional, computed, object({
    nested               = <optional, bool>
    size                 = <required, string> # auto | s | m | l | xl
    truncate_after_lines = <optional, float64>
    visible              = <optional, string> # auto | visible | hidden; maps to API `visibility`
  })> # schema default when omitted (typical size/visibility auto); optional+computed for Terraform
  metrics               = <required, list(object({ config = <required, json string with defaults> }))>
  group_by              = <optional, list(object({ config = <required, json string with defaults> }))>
})>
```

Notes:

- `legend_json` is removed.
- `legend` uses the same Terraform-facing shape as treemap and mosaic legends.
- The Terraform attribute is `visible`, which maps to the API field `legend.visibility`.
- The `legend` block remains optional for pie charts; the Terraform schema marks it optional and computed with a default object so omitted practitioner configuration stays consistent with typical Kibana read-back.

---

## MODIFIED Requirements

### Requirement: Pie chart panel behavior (REQ-023)

For pie Lens panels, the resource SHALL require at least one `metrics` entry and MAY accept `group_by`. It SHALL select the non-ES|QL branch when `query` is present and the ES|QL branch otherwise. When Kibana omits `ignore_global_filters` or `sampling` on read, the provider SHALL treat their default values as `false` and `1.0` respectively. Pie metric and group-by semantic equality SHALL normalize the implementation's pie metric defaults and Lens group-by defaults.

`dataset_json` SHALL remain a normalized JSON string for the pie dataset object. The resource SHALL expose an optional structured **`legend`** block matching treemap and mosaic legends (attributes `nested`, required `size`, optional `truncate_after_lines`, optional `visible`). The Terraform attribute `legend.visible` SHALL map to the API field `legend.visibility`. When the `legend` block is absent from practitioner configuration, the provider SHALL still build a valid API pie legend by supplying the implementation default legend size `auto`. The Terraform schema SHALL use an optional computed **`legend`** with a default object (typically size and visibility `auto`) so plan-time defaults align with typical Kibana read-back when the block is omitted.

#### Scenario: Pie chart API defaults

- GIVEN a pie panel read from Kibana without explicit `ignore_global_filters` or `sampling`
- WHEN state is refreshed
- THEN the provider SHALL reconcile those fields as `false` and `1.0`

#### Scenario: Pie chart uses dataset_json

- GIVEN `pie_chart_config` with `dataset_json` set to a normalized JSON string for the pie dataset
- WHEN the provider builds the Lens attributes
- THEN it SHALL decode `dataset_json` into the API pie dataset shape

#### Scenario: Pie chart uses structured legend

- GIVEN `pie_chart_config.legend` with `size = "auto"` and `visible = "visible"`
- WHEN the provider builds the Lens attributes
- THEN it SHALL encode the pie legend using the API pie legend shape
- AND it SHALL map Terraform `visible` to API `visibility`

#### Scenario: Pie chart legend omitted

- GIVEN `pie_chart_config` with no `legend` block
- WHEN the provider builds the Lens attributes
- THEN it SHALL still produce a valid pie legend object for the API
- AND it SHALL use the implementation default legend size `auto`

#### Scenario: Pie chart read-back uses legend block

- GIVEN a managed pie chart whose API payload contains a legend object
- WHEN the provider refreshes state
- THEN it SHALL populate `pie_chart_config.legend`
- AND it SHALL NOT populate `pie_chart_config.legend_json`
