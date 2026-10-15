## Why

`elasticstack_kibana_dashboard`'s XY reference-line `thresholds[].axis` attribute
(`internal/kibana/dashboard/panel/lensxy/schema_attrs.go`) validates
`stringvalidator.OneOf("bottom", "left", "right")`. Kibana's Dashboard API rejects
all three values with HTTP 400 (`Invalid input: expected "y"`) and instead accepts
`x`, `y`, and `y2`. This is not version-specific to 9.6: `x|y|y2` has been the enum
in every Kibana release checked from 9.4.0 onward (9.4.0, 9.4.1, 9.5.0, the 9.4/9.5
branches, and `main`), and the 9.4.0 and 9.5.5 stacks rejected the old values live.
Kibana 9.3 still used `bottom|left|right` in its Lens config-builder schema, but the
provider's `elasticstack_kibana_dashboard` Dashboard API support starts at 9.4 (whether
9.3 exposes an HTTP endpoint for that schema was not verified), so the provider
schema's `bottom|left|right` enum has never been usable against any stack the resource
supports — practitioners who set `axis` following the documented schema get a failed
apply.

This was discovered while implementing #4902 / PR #4948, which added read-back
normalization for Kibana 9.6's server-injected `thresholds[].axis = "y"` default
(`lenscommon.PreserveNullIfStateEquals` in
`internal/kibana/dashboard/panel/lensxy/alignment_panels.go:208`) but explicitly
left the schema/validator enum rewrite as a non-goal, tracked here. The current
`MarkdownDescription` already tells practitioners to omit the attribute entirely
as a workaround; that workaround remains valid after this change (omitting `axis`
still works), but practitioners should also be able to set it explicitly.

## What Changes

- Replace the `thresholds[].axis` validator enum `bottom|left|right` with
  `x|y|y2`, matching what Kibana's Dashboard API actually accepts (and what the
  generated `kbapi` client already models,
  `KibanaHTTPAPIsVisXyReferenceLineLayerESQLStaticThresholdsAxis`). This is a
  **breaking change** to the Terraform schema contract: no `StateUpgrader` is
  needed because no persisted state could legally contain the removed values —
  every Kibana version with the Dashboard API has always rejected them at apply
  time.
- Rewrite `referenceLineAxisDescription` to document the new enum and drop the
  "omit this attribute" workaround language and the back-link to this issue.
- Extend the existing omit-default alignment in
  `alignXYLayerStateFromPlan`/`alignment_panels.go` so an explicit `axis = "y"`
  configured by the practitioner round-trips on Kibana 9.4 and 9.5, which strip an
  explicit `"y"` from their read-back (only Kibana 9.6+ echoes it). Today's
  `PreserveNullIfStateEquals(planThreshold.Axis, &stateThreshold.Axis,
  types.StringValue("y"))` only restores a **null** plan value when state reads
  back `"y"`; it does not yet restore a **known** plan value of `"y"` when state
  reads back null/absent (the 9.4/9.5 shape). Add that missing direction so both
  stacks avoid "Provider produced inconsistent result after apply".
- Update the one test fixture that encodes the old Y-metric `config_json` axis
  value (`converter_test.go:48`, `"axis":"left"` → `"axis":"y"` — Y-metric
  `config_json` is free-form JSON already using Kibana's `y|y2` values in
  practice; this is a stale fixture, not a schema change).
- Add acceptance coverage for explicit `axis = "x"`, `"y"`, and `"y2"` on the
  reference-line threshold, alongside the existing omitted-axis coverage, and a
  plan-time negative test asserting `axis = "left"` (and `bottom`/`right`) is
  rejected by the validator with the new allowed-values message.
- Sync `openspec/specs/kibana-dashboard/spec.md` (REQ-011, and the XY
  reference-line scenarios added by the `lens-kibana-9-6-readback-defaults`
  change) to describe the `x|y|y2` enum and the bidirectional omit-default rule,
  removing the "reconciling the enum is tracked separately" language that points
  back at this issue.
- No version gating: the accepted enum is identical on every Kibana version that
  exposes the Dashboard API (9.4+). The only version-dependent behavior is
  read-back normalization (omitted-axis injection on 9.6+, explicit-`"y"`
  stripping on 9.4/9.5), which is already handled by inspecting the API response
  rather than branching on the stack version.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `kibana-dashboard`: change the XY reference-line `thresholds[].axis` contract
  from the unusable `bottom|left|right` enum to the Kibana-accepted `x|y|y2`
  enum, and extend REQ-011's omit-default normalization to cover the Kibana
  9.4/9.5 case where an explicit `axis = "y"` is dropped from the read-back.

## Impact

- `internal/kibana/dashboard/panel/lensxy/schema_attrs.go` — replace the
  `axis` validator and `MarkdownDescription`.
- `internal/kibana/dashboard/panel/lensxy/alignment_panels.go` — add the
  missing plan-known/state-null restoration direction for `thresholds[].axis`.
- `internal/kibana/dashboard/panel/lensxy/converter_test.go` — fix the stale
  `"axis":"left"` fixture.
- `internal/kibana/dashboard/panel/lensxy/alignment_panels_test.go` — replace
  `TestAlignXYLayerStateFromPlan_explicitLeftThresholdAxisIsNotTreatedAsOmitDefaultY`
  with cases covering `x`/`y`/`y2` pass-through and the new 9.4/9.5 restoration
  direction.
- `internal/kibana/dashboard/panel/lensxy/acc_test.go` and
  `internal/kibana/dashboard/panel/lensxy/testdata/` — new fixtures for
  explicit `axis = "x" | "y" | "y2"` and a validator-rejection test for the
  removed values.
- `openspec/specs/kibana-dashboard/spec.md` — update REQ-011's XY
  reference-line `thresholds[].axis` language to the new enum and
  bidirectional omit-default rule.
- `docs/resources/kibana_dashboard.md` — regenerate (`make docs-generate`) to
  reflect the new `MarkdownDescription`.
- Flag the enum change as breaking in the implementation PR title/body (the
  `CHANGELOG.md` `[Unreleased]` section is generated from merged PR history, so no
  manual changelog file is added), with rationale that the removed values
  (`bottom`, `left`, `right`) never applied successfully against any supported
  Kibana version.
