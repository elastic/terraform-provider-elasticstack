# Design: SLO Burn Rate Panel Support

## Context

The Kibana Dashboard API exposes SLO burn rate panels as panel type `slo_burn_rate` (embeddable id `slo-burn-rate-embeddable`). Unlike `lens` panels, which embed a Lens visualization specification as a nested saved-object reference, `slo_burn_rate` panels carry their entire configuration inline within the panel `config` object. There is no separate saved object for an SLO burn rate chart: the SLO reference, duration, drilldowns, and display settings all live in the dashboard document.

The API schema for `slo-burn-rate-embeddable` defines a `config` object with two required fields (`slo_id`, `duration`) and several optional fields (`slo_instance_id`, `drilldowns`, `title`, `description`, `hide_title`, `hide_border`).

The existing provider architecture for typed panel configs (e.g. `markdown_config`, `xy_chart_config`) maps naturally onto this shape. This design follows those precedents. The `drilldowns` shape is shared with `slo_overview` (REQ-030) and `slo_error_budget` (REQ-031); this proposal uses the same typed drilldown object structure for consistency, except that the SLO burn rate implementation hardcodes the only API-supported `trigger` and `type` values rather than exposing them as practitioner input.

## Goals

1. Allow practitioners to manage `slo_burn_rate` panels fully through Terraform.
2. Expose all API fields in the Terraform schema with idiomatic naming.
3. Validate `duration` at plan time using the format `^\d+[mhd]$` to catch malformed values before any API call.
4. Preserve `slo_burn_rate_config` on read-back without drift from API-injected defaults.
5. Handle `slo_instance_id` null-preservation correctly: when not configured, preserve null in state rather than defaulting to the API sentinel `"*"`.
6. Be consistent with the existing panel architecture so the implementation is straightforward for contributors familiar with the codebase.

## Non-Goals

- Supporting `slo_burn_rate` panels through `config_json` (see decision below).
- Managing SLO lifecycle (the `slo_id` is a reference attribute, not a managed dependency).
- Supporting any other SLO panel types (e.g. `slo_alerts_table`, `slo_summary`) in this change.
- Changing any existing panel type behavior.

## Decisions

### Terraform shape: typed `slo_burn_rate_config` block

The panel config is modeled as a typed block `slo_burn_rate_config` on the panel object, consistent with `markdown_config` and all typed config blocks. This gives practitioners named, documented attributes rather than an opaque JSON blob, and allows Terraform to validate `duration` format and required fields at plan time.

The block is `optional` and mutually exclusive with all other panel config blocks, enforced by the existing schema-level exclusivity rules (REQ-006).

```hcl
slo_burn_rate_config = {
  # Required
  slo_id   = string
  duration = string  # format: [value][unit] where unit is m, h, or d — e.g. "5m", "3h", "6d"

  # Optional
  slo_instance_id = string  # API default "*"; provider preserves null when not configured
  title           = string
  description     = string
  hide_title      = bool
  hide_border     = bool

  drilldowns = list(object({
    url              = string  # required
    label            = string  # required
    encode_url       = bool    # optional; API default true
    open_in_new_tab  = bool    # optional; API default true
  }))
}
```

### Required fields

The two fields required by the API schema are marked required in the Terraform schema:
- `slo_id` — the identifier of the SLO to display burn rate for.
- `duration` — the look-back window for the burn rate chart; validated as matching `^\d+[mhd]$`.

These cannot be inferred from defaults and must be explicit for the panel to function.

### `duration` format validation

`duration` is validated using a schema-level string validator matching the regex `^\d+[mhd]$`. This allows values such as `"5m"`, `"3h"`, `"6d"` while rejecting free-form strings that would produce invalid API requests. The validation is applied at plan time, before any API call.

Valid unit suffixes are:
- `m` — minutes
- `h` — hours
- `d` — days

### `slo_instance_id` null-preservation

The API accepts `slo_instance_id` as an optional field and defaults to `"*"` (all instances) when omitted. The provider SHALL preserve null in Terraform state when `slo_instance_id` is not configured. It SHALL NOT substitute the API default `"*"` into state for practitioners who did not explicitly set the field. When `slo_instance_id` is configured, the provider writes it to the API; when it is absent from configuration, the provider omits it from the API request, allowing Kibana to apply its own default.

On read-back, if the API returns `slo_instance_id = "*"` but the prior state value was null (i.e. the practitioner never configured it), the provider SHALL keep the field null in state to avoid introducing drift.

### `drilldowns` as list of typed objects

`drilldowns` is modeled as a list of typed objects. Practitioners configure `url`, `label`, `encode_url`, and `open_in_new_tab`. The API-required `trigger` and `type` fields are not exposed because the generated Kibana schema currently defines only one valid value for each: `trigger = "on_open_panel_menu"` and `type = "url_drilldown"`. The provider injects those constants on write and ignores them on read.

### `config_json` support for `slo_burn_rate`

`config_json` write support is **not** extended to `slo_burn_rate` in this change. Reasons:

1. The `slo_burn_rate` config object is simple and fully expressible through typed attributes; there is no complex nested structure that would motivate JSON escape hatches.
2. Keeping `slo_burn_rate` typed-only simplifies `duration` validation: the regex constraint can be enforced at the schema layer without inspecting raw JSON.
3. This is consistent with the approach taken for `slo_overview` (REQ-030) and `slo_error_budget` (REQ-031).

REQ-010 is updated to document that `slo_burn_rate` is explicitly not in the `config_json`-supported set.

## Risks and Trade-offs

| Risk | Mitigation |
|------|-----------|
| `slo_instance_id = "*"` vs null drift: if a practitioner configures `slo_instance_id = "*"` explicitly, Kibana may return `"*"` on read and the provider will correctly populate it; but if they omit `slo_instance_id`, the provider must suppress the API-returned `"*"` to avoid drift | Implement null-preservation for `slo_instance_id` by seeding from prior state on read: if prior state is null and API returns `"*"`, keep null in state |
| `drilldowns` with `encode_url` or `open_in_new_tab` omitted: Kibana may return these with their defaults on read-back, causing drift if the provider always populates them | Treat omitted optional drilldown booleans as null in state; only populate from API response if the value was present in the API payload |
| New API fields added to `slo-burn-rate-embeddable` config not yet in the typed schema cannot be managed | Practitioners cannot use those fields until a schema update; this is consistent with how other typed panel blocks handle forward-compatibility |

## Migration and State

This change is purely additive. No existing dashboard state is affected. There is no schema version change required for the dashboard resource. Practitioners who have existing `slo_burn_rate` panels managed via `config_json` (which is currently unsupported on write) will need to migrate to `slo_burn_rate_config`.

## Open Questions

1. **`slo_instance_id` sentinel value**: Should the provider expose `"*"` as a valid non-null value distinct from omitted, or always normalize `"*"` to null when the practitioner did not configure it? The current design normalizes based on prior state; confirm during acceptance testing that this does not produce a perpetual diff when a practitioner explicitly sets `slo_instance_id = "*"`.

2. **`drilldowns` ordering**: The API returns drilldowns as an array. If Kibana reorders drilldowns on read-back, plan noise will result. Confirm during acceptance testing whether drilldown ordering is stable across create/read cycles. If not, a set-based comparison or sorted normalization may be needed.

