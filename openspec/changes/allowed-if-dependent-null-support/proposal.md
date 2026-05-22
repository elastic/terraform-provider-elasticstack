## Why

The `elasticstack_kibana_space` resource needs conditional validation that allows `disabled_features` when `solution` is `classic`, when `solution` is unset, and when `solution` is still unknown/computed during validation. The current conditional validator helpers are too rigid for that case, which forces resource-specific workarounds or broader validators like `ConflictsWith` that reject valid configurations.

## What Changes

- Extend the conditional validator utilities in `internal/utils/validators/conditional.go` so `AllowedIf...` validation treats an unknown dependent value as allowed.
- Add explicit support for allowing a null/unset dependent value in `AllowedIf` validation via a dedicated helper for the equals case.
- Update the Kibana space resource schema to use the conditional validator helper instead of unconditional conflict validation for the `solution` / `disabled_features` relationship.
- Add tests covering null, unknown, matching, and non-matching dependent values for the validator utility and for the `elasticstack_kibana_space` schema behavior.

## Capabilities

### New Capabilities
- `conditional-validators-null-support`: Allow `AllowedIf` conditional validators to explicitly accept a null dependent value while always permitting unknown dependent values.

### Modified Capabilities
- `kibana-space`: Change validation requirements for the `solution` and `disabled_features` attributes so `disabled_features` is allowed when `solution` is `classic`, unset, or unknown, and rejected only for non-`classic` concrete values.

## Impact

- Affected code: `internal/utils/validators/conditional.go`, `internal/utils/validators/conditional_test.go`, `internal/kibana/spaces/resource_schema.go`, and associated resource tests.
- Affected behavior: plan-time validation for any schema that adopts the new helper, and specifically `elasticstack_kibana_space` configuration validation.
- No external API or dependency changes are required.
