## Why

`elasticstack_fleet_integration_policy` fails every apply with `Provider produced inconsistent result after apply` when `inputs` is configured as an empty map:

```
Error: Provider produced inconsistent result after apply
...produced an unexpected new value: .inputs: was cty.MapValEmpty(cty.Object(...)), but now null.
```

When `inputs` is not configured with any input, create, update and read set it to null in state, and `populateInputsFromAPI` turns a known empty map into null. That is correct when `inputs` is omitted (null), but not when the configuration sets `inputs = {}`: Terraform requires the value after apply to equal the planned empty map. A configuration that states "no inputs" explicitly, for example one generated from a map of policies where some policies have no inputs, or a policy imported without inputs and then managed with `inputs = {}`, can therefore never be applied. Because the failed apply stores null, the next plan shows `+ inputs = {}` again and the error repeats.

## What Changes

- Keep a known empty `inputs` map as an empty map in state after create, update and read, instead of replacing it with null. Null (omitted) and unknown values still result in null, as before.
- `populateInputsFromAPI` no longer replaces a known empty map with null.

Not breaking: no schema change. Only configurations with `inputs = {}` behave differently; they now apply cleanly instead of failing, and their state holds an empty map instead of null.

## Capabilities

### New Capabilities

_None — this fixes the state handling of an existing attribute._

### Modified Capabilities

- `fleet-integration-policy`: REQ-013, REQ-015 and REQ-017 distinguish an omitted `inputs` (null in state) from an explicitly empty `inputs` map (empty map in state).

## Impact

- `internal/fleet/integration_policy/models.go`: `inputsWhenUnconfigured` helper; `populateInputsFromAPI` keeps a known empty map.
- `internal/fleet/integration_policy/create.go`, `update.go`, `read.go`: use `inputsWhenUnconfigured` with the planned (create, update) or prior state (read) value instead of always setting null.
- Tests: unit tests for `inputsWhenUnconfigured` and for `populateFromAPI` with an empty map; acceptance tests `TestAccResourceIntegrationPolicyEmptyInputs` (policy created without inputs, then `inputs = {}`, then a further update) and `TestAccResourceIntegrationPolicyEmptyInputsOnCreate`.
