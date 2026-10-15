## 1. Schema

- [x] 1.1 In `internal/kibana/dashboard/panel/lensxy/schema_attrs.go`, replace
      the `thresholds[].axis` validator with
      `stringvalidator.OneOf("x", "y", "y2")`.
- [x] 1.2 Replace the `referenceLineAxisDescription` constant with wording that
      documents `x`/`y`/`y2` (axis the reference line is drawn on; `y` is the
      primary Y axis, `y2` the secondary Y axis, `x` uses the chart's X axis, per Kibana's
      "values match the root axis configuration keys"; `x` rendering semantics
      were not exercised; omitted defaults to `y`). Remove the "not settable against
      Kibana 9.6 / omit this attribute / tracked in #4958" language.
- [x] 1.3 Confirm `axis` stays `Optional` (not `Computed`).

## 2. Read-back alignment

- [x] 2.1 In `internal/kibana/dashboard/panel/lensxy/alignment_panels.go`,
      extend the threshold loop in `alignXYLayerStateFromPlan` so that when
      `planThreshold.Axis` is known and equals `"y"` and
      `stateThreshold.Axis` is null (Kibana 9.4/9.5 stripping an explicit
      `"y"`), the state value is restored to `"y"` from the plan. Keep the
      existing `PreserveNullIfStateEquals(planThreshold.Axis,
      &stateThreshold.Axis, types.StringValue("y"))` call for the null-plan /
      injected-state direction (9.6+ omitted-axis default) — this is an
      additive change, not a replacement.
- [x] 2.2 Add a small helper (e.g. in `lenscommon`, following the naming
      pattern of `PreserveNullIfStateEquals`) only if the inline branch would
      duplicate logic needed elsewhere; otherwise keep it local to
      `alignment_panels.go` per the existing style of
      `preserveThresholdValueJSONIfStateIsPlanValue`.

- [x] 2.3 In `layer_conv.go`, merge the typed `axis` into the `value_json`
      operation object on write (error on conflicting `axis` or non-object
      `value_json`), keep non-`static_value` operation objects whole on read,
      and extend alignment for an `axis` carried inside `value_json`.

## 3. Fixture and test fixes

- [x] 3.1 Update `internal/kibana/dashboard/panel/lensxy/converter_test.go:48`:
      change `"axis":"left"` to `"axis":"y"` in the Y-metric `config_json`
      fixture (free-form JSON already using Kibana's post-9.4 values; this is
      a stale literal, not a schema change).
- [x] 3.2 In `internal/kibana/dashboard/panel/lensxy/alignment_panels_test.go`,
      replace
      `TestAlignXYLayerStateFromPlan_explicitLeftThresholdAxisIsNotTreatedAsOmitDefaultY`
      (asserts behavior for a value the new validator no longer accepts) with
      cases covering: explicit `"x"` pass-through, explicit `"y2"`
      pass-through (already covered by
      `TestAlignXYLayerStateFromPlan_explicitReferenceLineThresholdAxisIsNotOverridden`
      — confirm it still passes unchanged), explicit `"y"` with state `"y"`
      (9.6 echo, no-op), and explicit `"y"` with state null (9.4/9.5 strip,
      restored to `"y"`), and explicit `"y2"` with state null (must NOT be
      restored; the inconsistency is allowed to surface, proving the rule is
      limited to the literal default `"y"`).
- [x] 3.3 Add a schema validator unit test (no stack required) asserting `axis =
      "left"` (and `"right"`, `"bottom"`) produce a plan-time validation error
      naming the allowed values `x`, `y`, `y2`.

## 4. Acceptance coverage

- [x] 4.1 In `internal/kibana/dashboard/panel/lensxy/acc_test.go` and
      `testdata/TestAccResourceDashboardXYChart_layers_reference/`, add
      fixtures/steps for explicit `axis = "x"`, `axis = "y"`, and `axis =
      "y2"` on a reference-line threshold, each asserting the attribute value
      and an empty plan after apply.
- [x] 4.2 Keep the existing omitted-axis fixture/step to cover the Kibana 9.6+
      injection case.
- [ ] 4.3 Run the extended `TestAccResourceDashboardXYChart_layers_reference`
      against 9.4, 9.5 and 9.6+ stacks per
      `dev-docs/high-level/testing.md` — the explicit-`"y"` case is the one
      that distinguishes them (stripped on 9.4/9.5, echoed on 9.6+; see
      design.md). Record results in the PR.
- [x] 4.4 Add an import step for an explicit `axis = "y"` fixture and assert the
      documented per-stack state (`"y"` on 9.6+, null on 9.4/9.5).

## 5. Spec and docs sync

- [x] 5.1 Update `openspec/specs/kibana-dashboard/spec.md` REQ-011: replace
      the `thresholds[].axis` enum description (`bottom|left|right`,
      "reconciling the enum is tracked separately
      (https://github.com/elastic/terraform-provider-elasticstack/issues/4958)")
      with the `x|y|y2` enum and the bidirectional omit-default rule from
      design.md Decision 3. Add/update scenarios for explicit `x`, explicit
      `y2`, explicit `y` preserved on a stack that strips it, and rejection of
      `left`/`right`/`bottom`.
- [x] 5.2 Regenerate `docs/resources/kibana_dashboard.md` (`make
      docs-generate`) so both reference-line `axis` occurrences reflect the
      new `MarkdownDescription`.
- [ ] 5.3 Flag the enum change as breaking in the implementation PR title/body
      with the rationale that the removed values never applied successfully
      against any supported Kibana version. Do not add a changelog file:
      `CHANGELOG.md` `[Unreleased]` is generated from merged PR history
      (`dev-docs/high-level/contributing.md`).

## 6. Validation

- [x] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate
      xy-reference-line-axis-enum --type change` and fix any reported issues.
- [x] 6.2 During implementation (not part of this proposal run): `make
      build`; `go test ./internal/kibana/dashboard/panel/lensxy/...`; run the
      acceptance tests in Task 4 with `TF_ACC=1` against live 9.4/9.5 and 9.6+
      stacks per `dev-docs/high-level/testing.md`.
