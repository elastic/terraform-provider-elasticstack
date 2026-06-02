## Why

Three Kibana resources (`kibana_security_exception_item`, `kibana_security_detection_rule`, `kibana_alerting_rule`) still use `*entitycore.ResourceBase` and implement CRUD as methods on the resource struct, while the rest of the provider has migrated to `*entitycore.KibanaResource[T]`. This leaves duplicate boilerplate (manual `kibana_connection` schema injection, manual client resolution, manual read-after-write) and makes these resources harder to maintain consistently.

## What Changes

- `internal/kibana/security_exception_item`: migrate from `ResourceBase` to `KibanaResource[ExceptionItemModel]`
- `internal/kibana/security_detection_rule`: migrate from `ResourceBase` to `KibanaResource[Data]`
- `internal/kibana/alertingrule`: migrate from `ResourceBase` to `KibanaResource[alertingRuleModel]`
- Each model gains four interface methods: `GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`
- Each model implements `WithVersionRequirements` (`GetVersionRequirements()`); inline `EnforceMinVersion` calls and `clients.MinVersionEnforceable` parameters on model-conversion helpers are removed
- `alerting_rule` `features.go` (the `alertingRuleFeatures` struct, `alertingRuleFeaturesFromVersion`, `resolveAlertingRuleFeatures`) is deleted; `toAPIModel` loses its `features` parameter
- Manual `kibana_connection` block injection removed from each schema function (envelope handles it)
- CRUD method bodies converted to package-level callback functions (full envelope migration; no `PlaceholderKibanaWriteCallback`)
- `alerting_rule` `getRuleIDAndSpaceID()` helper removed (replaced by envelope's `resolveKibanaResourceIdentity`)
- `entitycore_contract_test.go` added to each package
- No user-visible schema or behaviour changes

## Capabilities

### New Capabilities

_(none — this is a pure internal refactor)_

### Modified Capabilities

_(none — no spec-level behaviour changes)_

## Impact

- **Code**: `internal/kibana/security_exception_item`, `internal/kibana/security_detection_rule`, `internal/kibana/alertingrule`
- **Tests**: existing acceptance tests unchanged; new unit contract tests added per package
- **Dependencies**: no new dependencies
- **APIs**: no API behaviour changes
- **Users**: no schema changes, no breaking changes
