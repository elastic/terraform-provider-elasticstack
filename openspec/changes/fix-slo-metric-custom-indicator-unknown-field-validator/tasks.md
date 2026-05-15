## 1. Validator Fix

- [ ] 1.1 In `internal/utils/validators/conditional.go`, locate `RequiredIfDependentPathExpressionOneOf.validateValue` (≈ line 541). After `var diags diag.Diagnostics`, add:
  ```go
  if val == nil || val.IsUnknown() {
      return diags
  }
  ```
- [ ] 1.2 In the same file, locate `RequiredIfDependentPathOneOf.validateValue` (≈ line 373). After `var diags diag.Diagnostics`, add the same guard:
  ```go
  if val == nil || val.IsUnknown() {
      return diags
  }
  ```

## 2. Unit Test Fix

- [ ] 2.1 In `internal/utils/validators/conditional_test.go`, find `TestRequiredIfDependentPathExpressionOneOf` (≈ line 762). Locate the test case named `"invalid - current unknown, dependent matches required value"` (≈ line 799) and change `expectedError: true` to `expectedError: false`.
- [ ] 2.2 In `TestRequiredIfDependentPathOneOf`, locate the existing test case for `"invalid - current unknown, dependent matches required value"` and update it instead of adding a duplicate case:
  - change `expectedError: true` to `expectedError: false`
  - rename the case to reflect the valid behavior, for example `"valid - current unknown, dependent matches required value"`
- [ ] 2.3 Run `go test ./internal/utils/validators/...` to confirm all tests pass.

## 3. Requirements Update

- [ ] 3.1 Apply the delta spec in `openspec/changes/fix-slo-metric-custom-indicator-unknown-field-validator/specs/kibana-slo/spec.md` to `openspec/specs/kibana-slo/spec.md`: add or update the requirement for `metric_custom_indicator.metrics.field` and `histogram_custom_indicator.{good,total}.from`/`to` validator deferral behavior.

## 4. Validation

- [ ] 4.1 Run `make build` to confirm the change compiles.
- [ ] 4.2 Run `go test ./internal/utils/validators/...` to confirm all validator unit tests pass.
- [ ] 4.3 Run `make check-lint` to confirm lint passes.
- [ ] 4.4 If a Kibana stack is available, run the SLO acceptance tests: `TF_ACC=1 go test -v -run TestAccResourceKibanaSlo ./internal/kibana/slo/... -timeout 30m`
