## Why

Three Fleet resources (`fleet_server_host`, `fleet_output`, `fleet_custom_integration`) still use `*entitycore.ResourceBase` and implement CRUD as methods on the resource struct, while the rest of the provider has migrated to `*entitycore.KibanaResource[T]`. This leaves duplicate boilerplate (manual `kibana_connection` schema injection, manual client resolution, manual read-after-write) and makes these resources harder to maintain consistently.

## What Changes

- `internal/fleet/serverhost`: migrate from `ResourceBase` to `KibanaResource[serverHostModel]`
- `internal/fleet/output`: migrate from `ResourceBase` to `KibanaResource[outputModel]`
- `internal/fleet/customintegration`: migrate from `ResourceBase` to `KibanaResource[customIntegrationModel]`
- Each model gains four interface methods: `GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`
- `serverhost` and `output` models additionally implement `KibanaUnscopedSpace` (`IsUnscopedSpace() bool`) to suppress space-validation for Fleet's optional `space_ids` field
- `output` and `customintegration` models implement `WithVersionRequirements` (`GetVersionRequirements()`); inline `EnforceMinVersion` calls and `assertKafkaSupport`/`assertSSLVerificationModeSupport` helpers are removed
- Manual `kibana_connection` block injection removed from each schema function (envelope handles it)
- CRUD method bodies converted to package-level callback functions (full envelope migration; no `PlaceholderKibanaWriteCallback`)
- `entitycore_contract_test.go` added to each package
- No user-visible schema or behaviour changes

## Capabilities

### New Capabilities

_(none — this is a pure internal refactor)_

### Modified Capabilities

_(none — no spec-level behaviour changes)_

## Impact

- **Code**: `internal/fleet/serverhost`, `internal/fleet/output`, `internal/fleet/customintegration`
- **Tests**: existing acceptance tests unchanged; new unit contract tests added per package
- **Dependencies**: no new dependencies
- **APIs**: no API behaviour changes
- **Users**: no schema changes, no breaking changes
