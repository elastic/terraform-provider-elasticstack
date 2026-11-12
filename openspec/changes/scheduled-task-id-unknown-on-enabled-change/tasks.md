## 1. Plan `scheduled_task_id` as unknown when `enabled` transitions

- [ ] 1.1 Add a predicate in `internal/kibana/alertingrule/planmodifiers.go` that compares the planned
      `enabled` value against the prior state `enabled` value (not config vs. state) and returns true
      when they differ, including when planned `enabled` is unknown. Extract it as a pure function
      taking the plan/state `enabled` values, following the existing
      `actionFrequencyNewlyIntroduced`/`throttleShouldResetForActionFrequency` split in the same file so
      the predicate is unit-testable without fabricating `tfsdk.Plan`/`tfsdk.State`.
- [ ] 1.2 Add a constructor (e.g. `scheduledTaskIDSetUnknownIfEnabledChanged()`) that wires the
      predicate through `planmodifiers.StringSetUnknownIf`, mirroring
      `throttleSetUnknownIfActionsFrequencyConfigured()`.
- [ ] 1.3 Append the new modifier after `stringplanmodifier.UseStateForUnknown()` in
      `scheduled_task_id`'s `PlanModifiers` in `internal/kibana/alertingrule/schema.go:150-156`.
- [ ] 1.4 Add a short comment on the attribute (matching the style of the `throttle` comment at
      `schema.go:139-144`) explaining why `scheduled_task_id` is unknowned on `enabled` transitions.

## 2. Unit test the predicate

- [ ] 2.1 In `internal/kibana/alertingrule/planmodifiers_test.go`, add cases for the new predicate:
      `enabled` unchanged (known false→false, known true→true) does not trigger; `enabled` changing in
      either direction (false→true and true→false) triggers; a planned `enabled` that is unknown
      triggers conservatively (does not panic or silently skip).
- [ ] 2.2 Add a modifier-adapter test with config `enabled` omitted, state `enabled = true`,
      default-resolved plan `enabled = true`, and a known `scheduled_task_id`. Assert that the
      modifier preserves the known task ID. Exercise the adapter's attribute reads, not only the
      pure helper, so reading config instead of plan makes this regression test fail.

## 3. Acceptance coverage for the enable/disable transition

- [ ] 3.1 Extend `TestAccResourceAlertingRuleEnabledFalseOnCreate`
      (`internal/kibana/alertingrule/acc_test.go:560-587`) into a multi-step test (or add a new test
      using the same `ConfigDirectory`-per-step pattern as `TestAccResourceAlertingRuleInconsistentParams`,
      `acc_test.go:620-651`) with these steps in order:
      1. "create" step (existing config): `enabled = "false"`, and assert
         `resource.TestCheckNoResourceAttr(..., "scheduled_task_id")`.
      2. "enable" step: new testdata config with `enabled = true`; assert `enabled = "true"` and
         `resource.TestCheckResourceAttrSet(..., "scheduled_task_id")`. This step reproduces the
         issue's failure today and must apply cleanly after the fix. Follow with a plan-only check
         (`PlanOnly: true`, or a repeat apply step with the same config) asserting no further diff.
      3. "enable-omitted-noop" step, applied while still enabled: config that drops the `enabled` line
         entirely (falls back to `booldefault.StaticBool(true)`) from the prior `true` state; assert no
         diff and that `scheduled_task_id` is not spuriously unknowned — the regression case for the
         plan-vs-config predicate choice.
      4. "disable again" step: config with `enabled = false`; use `ConfigPlanChecks.PreApply` with
         `plancheck.ExpectUnknownValue` to assert `scheduled_task_id` is unknown in the disabling plan;
         assert `enabled = "false"` after apply and that a following plan is diff-free.
- [ ] 3.2 Add a second test (or additional steps) covering enabling via config omission rather than an
      explicit `true`: "create" step with `enabled = false`, then an "enable-via-omission" step that
      drops the `enabled` line entirely from the `false` state (so the default `true` takes effect);
      assert `enabled = "true"` and `resource.TestCheckResourceAttrSet(..., "scheduled_task_id")`.
- [ ] 3.3 Add the new testdata `main.tf` files needed for the above steps under
      `internal/kibana/alertingrule/testdata/<TestName>/<step>/`, following the existing layout for
      this test's `create` step.

## 4. Validate

- [ ] 4.1 Run `go vet ./internal/kibana/alertingrule/...` and the package's unit tests
      (`go test ./internal/kibana/alertingrule/...`, excluding `TestAcc*`) to confirm the new predicate
      and its wiring compile and pass.
- [ ] 4.2 Run the new/extended acceptance tests (`TF_ACC=1 go test ./internal/kibana/alertingrule/...
      -run <TestName>`) against a Kibana instance to confirm the `enabled` transition no longer
      produces "Provider produced inconsistent result after apply", per the project's acceptance
      testing docs.
