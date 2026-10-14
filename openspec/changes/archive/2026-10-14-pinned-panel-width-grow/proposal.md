## Why

A pinned panel (filter dropdown) ignores the `width` and `grow` fields when deploying. As a result
the filter dropdowns end up having the wrong size (issue #5226). The Kibana Dashboards API's
pinned-control-group payload carries per-panel `width` (enum `small`/`medium`/`large`, default
`medium`) and `grow` (bool, default `false`), but `pinned_panels[]` currently exposes only `type`
and the four `*_config` blocks — matching the [documented provider
spec](https://registry.terraform.io/providers/elastic/elasticstack/latest/docs/resources/kibana_dashboard#nested-schema-for-pinned_panels)
— so these two values are silently discarded on apply instead of being forwarded to Kibana.

The generated Kibana API client already fully supports both fields (`Grow *bool` / `Width *enum` on
all four `ControlsGroupSchema*Control` structs). The gap is entirely on the provider side: the
Terraform schema, model, and the JSON-remap conversion path between the Terraform model and the API
payload all need to carry `width`/`grow`.

## What Changes

- Add optional `width` (`small`/`medium`/`large`) and `grow` (bool) attributes directly on each
  `pinned_panels[]` entry, alongside the existing `type` and `*_config` blocks.
- Thread `width`/`grow` through the pinned-control conversion path (`ToAPI`/`FromAPI`) for all four
  pinned control types (options list, range slider, ES|QL, time slider) so the values round-trip to
  and from the Kibana Dashboards API instead of being dropped.
- On read, follow the existing null-preservation convention: when there is no prior Terraform state
  (import or first read), populate `width`/`grow` directly from the API value; when a prior state
  exists, preserve a null/unknown configured value instead of overwriting it with the server's
  reported default (`medium`/`false`), so omitting the attributes from configuration produces no
  diff.
- No change to the four `*_config` block schemas themselves, to in-grid `panels[]` (which do not
  expose `width`/`grow` — unaffected), or to `number`/`percent`-style unrelated panel sizing fields.

## Capabilities

### New Capabilities

<!-- None — this constrains an existing capability. -->

### Modified Capabilities

- `kibana-dashboard`: Extend the dashboard-level pinned controls round-trip requirement (REQ-038) so
  `pinned_panels[]` entries also carry optional `width` and `grow` attributes that round-trip to and
  from the Kibana Dashboards API, instead of being silently discarded.

## Impact

- **Code**: `internal/kibana/dashboard/schema_control_configs.go` (schema), `internal/kibana/
  dashboard/models/panel.go` (model), `internal/kibana/dashboard/panelkit/control_pinned_handler.go`
  (conversion mechanism), `internal/kibana/dashboard/panel/{esqlcontrol,optionslist,rangeslider,
  timeslider}/pinned.go` (per-control-type wiring), plus unit and acceptance test updates.
- **Specs**: Delta under `openspec/changes/pinned-panel-width-grow/specs/kibana-dashboard/`; merges
  into `openspec/specs/kibana-dashboard/spec.md` (REQ-038) on archive.
- **Dependencies**: No new dependencies.
