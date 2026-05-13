## 1. Fix Time Window Description

- [ ] 1.1 Replace the body of `internal/kibana/slo/descriptions/time_window.md` with accurate per-type duration lists:
  - Rolling: `7d` (7 days), `30d` (30 days), `90d` (90 days)
  - Calendar aligned: `1w` (weekly), `1M` (monthly)
  - Remove the inaccurate "Any duration greater than 1 day can be used: days, weeks, months, quarters, years" sentence

## 2. Add `OneOfWhenDependentPathExpressionEquals` Helper

- [ ] 2.1 Add `OneOfWhenDependentPathExpressionEquals(dependentPathExpression path.Expression, dependentValue string, allowedValues []string) Condition` to `internal/utils/validators/conditional.go`
  - The validator is a no-op when the dependent field does not equal `dependentValue`
  - When the dependent field equals `dependentValue`, validates that the current attribute's value is one of `allowedValues`
  - Error summary: `"Invalid Attribute Value Match"`
  - Error message format: `"Attribute <path> must be one of [<values>] when type is \"<dependentValue>\", got: \"<actual>\""`
  - Add `"strings"` to the import block if not already present (needed for `strings.Join`)
- [ ] 2.2 Add unit tests for `OneOfWhenDependentPathExpressionEquals` covering:
  - Condition not met (dependent field has different value) → no error
  - Condition met, valid value → no error
  - Condition met, invalid value → error with correct message

## 3. Add Conditional Validators to `duration`

- [ ] 3.1 In `internal/kibana/slo/schema.go`, update the `"duration"` attribute in the `time_window` nested block (`schema.go:176`) to add two validators:
  ```go
  Validators: []validator.String{
      validators.OneOfWhenDependentPathExpressionEquals(
          path.MatchRelative().AtParent().AtName("type"),
          "rolling",
          []string{"7d", "30d", "90d"},
      ),
      validators.OneOfWhenDependentPathExpressionEquals(
          path.MatchRelative().AtParent().AtName("type"),
          "calendarAligned",
          []string{"1w", "1M"},
      ),
  },
  ```
- [ ] 3.2 Ensure required imports are present in `schema.go`: `github.com/hashicorp/terraform-plugin-framework/path` and the internal `validators` package

## 4. Add Acceptance Test Fixtures

- [ ] 4.1 Create `internal/kibana/slo/testdata/TestAccResourceSloValidation/time_window_invalid_duration_rolling/test.tf` with `duration = "4d"` and `type = "rolling"` (modelled after `time_window_invalid_type/test.tf`)
- [ ] 4.2 Create `internal/kibana/slo/testdata/TestAccResourceSloValidation/time_window_invalid_duration_calendar/test.tf` with `duration = "30d"` and `type = "calendarAligned"`

## 5. Add Acceptance Test Steps

- [ ] 5.1 In `internal/kibana/slo/acc_test.go`, add two steps to `TestAccResourceSloValidation` after the existing `time_window_invalid_type` step:
  ```go
  {
      ProtoV6ProviderFactories: acctest.Providers,
      ConfigDirectory:          acctest.NamedTestCaseDirectory("time_window_invalid_duration_rolling"),
      ConfigVariables: config.Variables{
          "name": config.StringVariable("tw-dur-rolling"),
      },
      ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*duration`),
  },
  {
      ProtoV6ProviderFactories: acctest.Providers,
      ConfigDirectory:          acctest.NamedTestCaseDirectory("time_window_invalid_duration_calendar"),
      ConfigVariables: config.Variables{
          "name": config.StringVariable("tw-dur-calendar"),
      },
      ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*duration`),
  },
  ```

## 6. Regenerate Documentation

- [ ] 6.1 Run `make generate` to regenerate `docs/resources/kibana_slo.md` from the updated description

## 7. Final Verification

- [ ] 7.1 Run `make build` to confirm the provider compiles without errors
- [ ] 7.2 Run `make check-lint` and fix any issues
- [ ] 7.3 Run the targeted `TestAccResourceSloValidation` acceptance test steps against a live stack (or verify plan-only steps pass without a stack)
- [ ] 7.4 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-slo-time-window-duration-validation --type change` and resolve any issues
- [ ] 7.5 Archive the change with `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec archive kibana-slo-time-window-duration-validation`
