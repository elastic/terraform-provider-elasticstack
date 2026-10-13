## Why

A Lens XY chart `y[]` metric whose `config_json.format` has `type = "duration"` and also sets
`decimals` and/or `compact` fails on Kibana Dashboard API versions **9.4.0–9.4.4** and
**9.5.0–9.5.1** with an HTTP 400 whose body is Kibana's multi-screen union-validation error (see
issue #5166). The same `format` keys are accepted when `format.type` is `number` or `percent`; the
Kibana duration-format schema simply did not accept `decimals`/`compact` until it was fixed in
**9.5.2**. The provider does not inject these properties itself — `PopulateLensMetricDefaults`
(`internal/kibana/dashboard/lenscommon/populate_lens_charts.go`) only defaults `decimals`/`compact`
for `number`/`percent` formats — so this only happens when practitioners set them explicitly in
their `config_json`. Today the provider forwards the payload to Kibana unchanged and surfaces
Kibana's raw, hard-to-read 400 error.

## What Changes

- Add a version constant for the Kibana release that fixed duration-format `decimals`/`compact`
  validation, alongside the existing `panelkit.MinKibanaAPISupport95` pattern.
- Add a pure helper that inspects a Lens metric's `config_json`-decoded map and reports whether its
  `format` uses `type/id == "duration"` together with a `decimals` or `compact` key (top-level or
  under `format.params`).
- Walk the dashboard plan model's panels (top-level `panels` and panels nested in `sections`) for XY
  chart layers before building the create/update API request. When any `y[]` metric trips the
  helper, call `client.EnforceMinVersion` against the new constant and fail with a diagnostic that
  names the offending panel/layer/metric path and the two unsupported properties, instead of letting
  the raw Kibana 400 reach the practitioner.
- No behavior change for `number`/`percent` formats, for duration formats without `decimals`/
  `compact`, for dashboards with no XY charts, or on Kibana >= 9.5.2 / Serverless.
- Document the minimum version on the `y[].config_json` attribute description for the XY chart data
  layer.

## Capabilities

### New Capabilities

<!-- None — this constrains an existing capability. -->

### Modified Capabilities

- `kibana-dashboard`: Add a minimum Elastic Stack version requirement for Lens XY chart `y[]`
  metrics whose `config_json.format` combines `type = "duration"` with `decimals` and/or `compact`;
  require a clear diagnostic (instead of the raw Kibana 400) when the connected stack is older.

## Impact

- **Code**: `internal/kibana/dashboard/panelkit/version_support.go` (new constant),
  `internal/kibana/dashboard/lenscommon/` (new detection helper, new tests),
  `internal/kibana/dashboard/create.go` and `internal/kibana/dashboard/update.go` (preflight check
  wired into the write path), `internal/kibana/dashboard/panel/lensxy/schema_attrs.go` (attribute
  description update).
- **Specs**: Delta under `openspec/changes/dashboard-lens-duration-format-version-gate/specs/
  kibana-dashboard/`; merges into `openspec/specs/kibana-dashboard/spec.md` on archive.
- **Dependencies**: Existing `clients.KibanaScopedClient.EnforceMinVersion` and `hashicorp/go-version`
  (already used throughout `internal/kibana/dashboard`); no new dependencies.
