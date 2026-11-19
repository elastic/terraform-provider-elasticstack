## 1. Fix

- [x] 1.1 Add `inputsWhenUnconfigured` to `internal/fleet/integration_policy/models.go`: a known empty map stays an empty map, null and unknown become null
- [x] 1.2 Keep a known empty map in `populateInputsFromAPI` instead of setting null
- [x] 1.3 Use `inputsWhenUnconfigured` with the planned value in `create.go` and `update.go` and with the prior state value in `read.go`

## 2. Tests

- [x] 2.1 Unit tests `TestInputsWhenUnconfigured` (null, unknown, empty map) and `TestPopulateFromAPI_Inputs_KnownEmptyStaysEmpty`
- [x] 2.2 Acceptance test `TestAccResourceIntegrationPolicyEmptyInputs`: create without `inputs`, apply `inputs = {}`, apply a further update with `inputs = {}`; every step's post-apply plan is empty
- [x] 2.3 Acceptance test `TestAccResourceIntegrationPolicyEmptyInputsOnCreate`: create with `inputs = {}`
- [x] 2.4 Confirm both acceptance tests fail without the fix (`.inputs: was cty.MapValEmpty(...), but now null`) and pass with it
- [x] 2.5 Run the existing `integration_policy` acceptance suite and confirm no regressions

## 3. Validation

- [x] 3.1 Run unit tests and golangci-lint
- [x] 3.2 Run `openspec validate --all`
