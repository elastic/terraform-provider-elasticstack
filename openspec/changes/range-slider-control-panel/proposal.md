# Proposal: Range Slider Control Panel Support for `elasticstack_kibana_dashboard`

## Why

Practitioners cannot manage range slider control panels on Kibana dashboards as code today. Range slider controls provide a numeric range filter — a min/max slider — tied to a data view field, allowing dashboard users to dynamically narrow query results to a specific numeric window (for example, filtering by response time, price, or age). Without Terraform support, teams that otherwise manage their dashboards as code must configure these controls manually in the Kibana UI or accept that their dashboard definitions are incomplete.

This creates operational risk (configuration drift, loss of reproducibility) and prevents full infrastructure-as-code workflows for dashboards that rely on range-based filtering controls.

## What Changes

- **Add `range_slider_control_config` typed panel config block** for panels with `type = "range_slider_control"`. This block captures all fields from the `range_slider_control` panel API schema in a structured, ergonomic way.
- **Add new requirement REQ-028** defining the behavior, required fields, optional fields, and read/write semantics of the `range_slider_control` panel type.
- **Update REQ-006** to include schema-level validation that `range_slider_control_config` is only valid when `type = "range_slider_control"`, that `range_slider_control_config` is mutually exclusive with all other panel config blocks, and that `value` when set must contain exactly 2 elements.
- **Update REQ-010** to document that `config_json` write support is not extended to the `range_slider_control` panel type; it must be managed through the typed `range_slider_control_config` block.

### Schema sketch (to merge into canonical `## Schema` on sync)

Add an optional typed config block at panel level:

```hcl
    range_slider_control_config = <optional, object({
      title                = <optional, string>
      data_view_id         = <required, string>
      field_name           = <required, string>
      use_global_filters   = <optional, bool>
      ignore_validations   = <optional, bool>
      value                = <optional, list(string)> # exactly 2 elements: [min, max]
      step                 = <optional, number>
    })> # only with type = "range_slider_control"; conflicts with all other config blocks
```

## Capabilities

After this change, practitioners will be able to:

- Declare `range_slider_control` panels on a dashboard with full control over which data view field drives the range filter.
- Optionally pre-populate the slider with an initial range (`value`) expressed as a `[min, max]` string pair.
- Configure slider sensitivity via `step`.
- Control global filter integration via `use_global_filters`.
- Suppress validation errors during intermediate states via `ignore_validations`.
- Import and plan-refresh existing `range_slider_control` panels without losing their configuration.

## Impact

- **Additive only**: no existing panel types or behaviors are changed.
- **Schema change**: adds a new optional `range_slider_control_config` block to the panel schema alongside existing typed config blocks.
- **REQ-006 update**: broadens the schema validation rules to cover the new panel type and config block, and adds the `value` list-length validator.
- **REQ-010 update**: clarifies that `range_slider_control` is not supported through `config_json` on write.
- **No state migration**: new block; existing dashboard state is unaffected.
- **No breaking change**: all existing dashboards remain valid.
