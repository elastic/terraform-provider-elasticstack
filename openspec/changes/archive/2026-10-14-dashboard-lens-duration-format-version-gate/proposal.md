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
- Check affected Lens XY chart `y[]` metrics before dashboard create/update requests and require
  Elastic Stack 9.5.2 or later on stateful stacks.
- Return an actionable diagnostic identifying the affected configuration on older stacks, and
  propagate status-retrieval or version-parsing diagnostics while aborting the dashboard request.
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

- **Code**: Dashboard create/update validation, Lens XY chart configuration documentation, and
  focused tests.
- **Specs**: Delta under `openspec/changes/dashboard-lens-duration-format-version-gate/specs/
  kibana-dashboard/`; merges into `openspec/specs/kibana-dashboard/spec.md` on archive.
- **Dependencies**: No new dependencies.
