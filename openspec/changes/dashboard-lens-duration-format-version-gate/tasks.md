## 1. Version constant

- [ ] 1.1 In `internal/kibana/dashboard/panelkit/version_support.go`, add
      `MinKibanaAPISupportDurationFormatParams = version.Must(version.NewVersion("9.5.2"))` next to
      `MinKibanaAPISupport95`, with a doc comment describing the Kibana 9.4.0–9.4.4 /
      9.5.0–9.5.1 rejection and the 9.5.2 fix (per issue #5166's empirical version matrix).

## 2. Detection helper

- [ ] 2.1 In `internal/kibana/dashboard/lenscommon/populate_lens_charts.go` (or a new file in the
      same package), add a pure function, e.g.
      `MetricFormatUsesUnsupportedDurationParams(model map[string]any) bool`, that:
      - Reads `model["format"]` as `map[string]any`; returns `false` if missing/not a map.
      - Treats the format as "duration" when `format["type"] == "duration"` or
        `format["id"] == "duration"` (mirroring the legacy `id` spelling tolerance already present
        in `PopulateLensMetricDefaults`'s `isNumberish` check).
      - Returns `true` when a duration format has a `decimals` or `compact` key present either
        directly on `format` or under `format["params"]` (mirroring the `params` vs. top-level
        duality `PopulateLensMetricDefaults` already handles for `number`/`percent`).
      - Returns `false` for `number`/`percent` formats and for duration formats with neither key.
- [ ] 2.2 Unit-test the helper directly over `map[string]any` fixtures: duration with `decimals`
      only, `compact` only, both, neither; `params`-nested variants of each; `number`/`percent`
      with the same keys (must not trip); missing `format`; `format: null`; legacy `id == "duration"`
      spelling.

## 3. Preflight wiring in the write path

- [ ] 3.1 Add a helper (e.g. in `internal/kibana/dashboard/` or `panelkit`) that walks a
      `models.DashboardModel`'s `Panels` and `Sections[].Panels`, finds XY chart data layers
      (`VisConfig.ByValue.XYChartConfig...Layers[].DataLayer`), decodes each `Y[].ConfigJSON`, and
      reports every metric that trips `lenscommon.MetricFormatUsesUnsupportedDurationParams`,
      along with enough path information (panel id/grid position, layer index, metric index) to
      build a diagnostic.
- [ ] 3.2 In `createDashboard` (`internal/kibana/dashboard/create.go`), before calling
      `dashboardToAPICreateRequest`/`putDashboard`, run the walk from 3.1. If it finds any
      offending metric, call
      `client.EnforceMinVersion(ctx, panelkit.MinKibanaAPISupportDurationFormatParams)`; if
      unsupported, append an error diagnostic naming the panel/layer/metric path and stating that
      `format.type "duration"` with `decimals`/`compact` requires Elastic Stack 9.5.2 or later, and
      return before issuing the API call.
- [ ] 3.3 Apply the same preflight in `updateDashboard` (`internal/kibana/dashboard/update.go`)
      before `putDashboard` is called from the update path, so `PUT` (true updates) gets the same
      early diagnostic. `putDashboard` is also called from `createDashboard` when the practitioner
      supplies `dashboard_id`; ensure that call site is covered too (either by running the same
      preflight before `createDashboard` calls `putDashboard`, or by sharing one helper called from
      both entry points).
- [ ] 3.4 Confirm no extra API call happens when no panel trips the helper (no new calls to
      `EnforceMinVersion`/`/api/status` for dashboards without the unsupported combination).

## 4. Documentation

- [ ] 4.1 Update the `y[].config_json` `MarkdownDescription` in
      `internal/kibana/dashboard/panel/lensxy/schema_attrs.go` to note that `format.type =
      "duration"` combined with `decimals` and/or `compact` requires Elastic Stack 9.5.2 or later.
- [ ] 4.2 Regenerate provider docs (`make docs-generate`) if the repository's non-code-change
      policy for this workflow allows it; otherwise leave as a follow-up task for the implementing
      change (this proposal only edits `openspec/changes/...`).

## 5. Tests

- [ ] 5.1 Unit tests per 2.2.
- [ ] 5.2 Unit test the preflight walk (3.1) with a stubbed/fake version-capable client: dashboard
      model with an offending metric inside a top-level panel, inside a `sections[].panels` panel,
      and with no offending metric at all (asserts `EnforceMinVersion` is not called in that case).
- [ ] 5.3 Acceptance test: reproduce the issue's panel (line layer, `y[]` metric with
      `format = {type="duration", from="s", to="ms", decimals=0, compact=true}`,
      `color = {type="auto"}`), gated with
      `versionutils.SkipIfUnsupported(t, panelkit.MinKibanaAPISupportDurationFormatParams,
      versionutils.FlavorAny)`, asserting apply succeeds and a follow-up plan shows no drift.
- [ ] 5.4 Acceptance test companion step for stacks below 9.5.2 asserting `ExpectError` with the
      new diagnostic's message, covering both a top-level panel and a panel nested inside a
      `sections[]` block (the path the original report hit).

## 6. Spec and validation

- [ ] 6.1 Keep the delta spec at
      `openspec/changes/dashboard-lens-duration-format-version-gate/specs/kibana-dashboard/spec.md`
      in sync with the behavior above.
- [ ] 6.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate
      dashboard-lens-duration-format-version-gate --type change` and fix any reported problems.
- [ ] 6.3 When implementing, run `make build` and the targeted unit/acceptance tests per
      [`dev-docs/high-level/testing.md`](../../../dev-docs/high-level/testing.md); run
      `make check-openspec` so the merged spec stays valid.
