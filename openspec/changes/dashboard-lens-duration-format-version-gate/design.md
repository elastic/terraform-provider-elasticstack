## Context

`elasticstack_kibana_dashboard` lets practitioners set a Lens XY chart `y[]` metric's
`config_json.format` directly (`internal/kibana/dashboard/panel/lensxy/schema_attrs.go`). When
`format.type == "duration"` and the practitioner also sets `decimals` and/or `compact`, Kibana
versions **9.4.0–9.4.4** and **9.5.0–9.5.1** reject the panel with an HTTP 400 whose body is a deep
`types that failed validation` union error — unreadable without already knowing the cause. The fix
shipped in Kibana **9.5.2** (confirmed empirically against 9.4.0, 9.4.3, 9.4.4, 9.5.0–9.5.5, and
9.6.0-SNAPSHOT). `number`/`percent` formats are unaffected on every version tested; `color` is not a
trigger on its own. The provider has no existing runtime version gating in
`internal/kibana/dashboard` — `panelkit.MinKibanaAPISupport95` is referenced only by doc comments on
sibling `version_support.go` files today, not by any create/update code path — so this change
introduces the first live preflight check in that package.

## Goals / Non-Goals

**Goals:**

- Fail fast with a diagnostic that names the panel/layer/metric path and the unsupported
  property combination, instead of surfacing Kibana's raw union-validation 400.
- Only pay the cost of a version check (`EnforceMinVersion` issues a status call) when a dashboard
  actually contains the unsupported combination; leave every other dashboard write unaffected.
- Reuse the existing `clients.KibanaScopedClient.EnforceMinVersion` pattern already used elsewhere
  in the provider.
- Keep the detection helper generic over `map[string]any` so it is reusable if other Lens chart
  families (datatable `metrics[]`, `metric` chart) turn out to share the same Kibana schema
  constraint.

**Non-Goals:**

- Changing provider-side defaulting. `PopulateLensMetricDefaults` never injects `decimals`/`compact`
  for `duration` formats today and continues not to.
- Gating datatable `metrics[]` or the standalone `metric`/`legacy_metric` chart formats in this
  change. Those chart families were not empirically tested against the same Kibana schema
  constraint (see Open Questions) and are left for a follow-up once confirmed.
- Surfacing this at `ModifyPlan` time. That would need a scoped client during planning, which the
  resource does not currently have available in its plan-modifier path; it is a possible follow-up,
  not required for this change.

## Decisions

1. **Version constant: `9.5.2` (not `9.5.2-SNAPSHOT`)**
   Add `MinKibanaAPISupportDurationFormatParams` next to `MinKibanaAPISupport95` in
   `internal/kibana/dashboard/panelkit/version_support.go`. Use the released `9.5.2` version string:
   9.5.2 is the first released version with the fix, and the provider's version comparisons treat a
   bare release as satisfying any constraint phrased against an earlier pre-release of the same
   version. This matches how `panelkit.MinKibanaAPISupport95` already uses a `-SNAPSHOT` suffix only
   because the 9.5.0 fix landed before the 9.5.0 GA release; here the fix landed mid-line at 9.5.2,
   so no snapshot suffix is needed or wanted.

2. **Detection helper lives in `lenscommon`, not `lensxy`**
   Add `lenscommon.MetricFormatUsesUnsupportedDurationParams(model map[string]any) bool` (or
   equivalent name) next to `PopulateLensMetricDefaults`. It inspects `model["format"]` for
   `type`/`id == "duration"` (mirroring the legacy `id` spelling `PopulateLensMetricDefaults`
   already tolerates for `number`/`percent`) and reports true when a `decimals` or `compact` key
   exists either directly on `format` or under `format["params"]`. Keeping it a pure function over
   `map[string]any` (not over the typed `DataLayerModel`) lets `lensxy`, and any other Lens chart
   family that later turns out to share the constraint, call it with their own decoded
   `config_json`.

3. **Where to call it: preflight in `create.go`/`update.go`, not inside `layer_conv.go`**
   `internal/kibana/dashboard/panel/lensxy/layer_conv.go` converts between the plan model and the
   Kibana API request shape and has no access to a scoped client, so it cannot call
   `EnforceMinVersion`. Instead, `createDashboard` (`internal/kibana/dashboard/create.go`) and its
   update equivalent (`internal/kibana/dashboard/update.go`) — which already receive
   `*clients.KibanaScopedClient` — walk `req.Plan.Panels` and `req.Plan.Sections[].Panels` before
   building the API request, decode each XY data layer's `y[].ConfigJSON`, and run the detection
   helper. Only when at least one metric trips it do they call
   `client.EnforceMinVersion(ctx, panelkit.MinKibanaAPISupportDurationFormatParams)`; otherwise no
   extra API call happens.

4. **Panel traversal scope: `panels` and `sections[].panels` only**
   `models.PinnedPanelModel` (`internal/kibana/dashboard/models/panel.go`) only has control-type
   config fields (`time_slider_control_config`, `esql_control_config`,
   `options_list_control_config`, `range_slider_control_config`) — it has no `vis_config`, so an XY
   chart cannot be a pinned panel today. The walk therefore only needs top-level `panels` and
   `sections[].panels`; pinned panels are structurally excluded rather than skipped by a runtime
   check. If a future change adds chart support to pinned panels, that walk will need extending.

5. **Diagnostic message**
   On failure, add an error diagnostic naming the panel (grid position or `id`), the layer index,
   the metric index, and stating: `format.type "duration" with "decimals"/"compact" requires Elastic
   Stack 9.5.2 or later; remove these properties or upgrade Kibana.` This mirrors the style of
   existing version-gate diagnostics in the provider (e.g. Fleet `space_ids`, Streams).

6. **Alternatives considered**
   - *Let Kibana's error pass through unchanged*: rejected — the issue report shows this is an
     unreadable multi-screen union-validation error that gives practitioners no actionable signal.
   - *Validate in the schema (`stringvalidator`/custom plan-time validator) without a client call*:
     rejected — Terraform validators cannot know the connected Kibana version, and the constraint is
     stack-version-dependent, not a static schema rule.
   - *Gate unconditionally on every dashboard apply regardless of whether the combination is
     present*: rejected — would add an unconditional extra API round trip (`EnforceMinVersion` calls
     `/api/status`) to every dashboard write, including the overwhelming majority that never use
     `duration` formats.

## Risks / Trade-offs

- **[Risk]** Practitioners legitimately running Kibana 9.4.0–9.4.4 or 9.5.0–9.5.1 who already rely on
  this combination (e.g. via `config_json` authored outside Terraform, then imported) will now see
  their apply fail where it previously got a raw 400 from Kibana. **Mitigation**: the new diagnostic
  is strictly more actionable than today's failure — both fail today, but the new message says
  exactly what to remove or which version to upgrade to; no previously-succeeding apply starts
  failing.
- **[Risk]** The helper's schema detection (`format.type`/`id == "duration"` plus `decimals`/
  `compact` at top level or under `params`) could miss a shape that Kibana's actual duration schema
  also rejects. **Mitigation**: unit tests enumerate the shapes observed in the issue report and in
  `PopulateLensMetricDefaults`'s existing `number`/`percent` handling of the same `params` vs.
  top-level duality; the helper is a narrow, independently testable function so gaps are easy to
  patch later without touching the write path.
- **[Trade-off]** The preflight decode duplicates `config_json` parsing that `layer_conv.go` will
  perform again when building the API request. This is accepted: the alternative (threading the
  scoped client into `layer_conv.go`'s stateless conversion functions) is a larger structural change
  for a check that only runs when the unsupported combination is actually present.

## Migration Plan

- No state migration. Existing dashboards that already use duration `decimals`/`compact` and
  successfully applied in the past only did so against Kibana >= 9.5.2 (older versions always
  rejected this combination, per the reproduction in issue #5166), so no previously-successful
  apply is affected.
- After this change, applying the same configuration against Kibana 9.4.0–9.4.4 or 9.5.0–9.5.1 fails
  with the new diagnostic instead of Kibana's raw 400. Practitioners on those versions must either
  remove `decimals`/`compact` from the affected `format` blocks or upgrade Kibana to 9.5.2+.

## Open Questions

- Whether `metric`/`legacy_metric` chart `format` and datatable `metrics[]` share the same Kibana
  duration-format schema constraint is unconfirmed; the detection helper is generic enough to apply
  to them once verified, but this change scopes enforcement to XY chart `y[]` metrics only.
- Serverless behavior for this combination was not verified against a real Serverless Kibana;
  `EnforceMinVersion` already always returns `true` for Serverless, so this change does not block
  Serverless regardless. If Serverless turns out to reject the combination, that needs a separate
  decision (this constant does not help there, since the gate would need to apply regardless of
  reported version).
- Whether to also surface this at `ModifyPlan` time (better UX, catches the problem before apply)
  is left as a follow-up: it needs a scoped client during planning, which is a larger change than
  this proposal's scope.
