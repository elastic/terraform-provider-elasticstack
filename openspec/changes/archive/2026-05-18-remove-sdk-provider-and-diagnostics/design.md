## Context

The provider currently serves two provider implementations through a mux:

```
main.go
  └─> ProtoV6ProviderServerFactory
        ├─> New("dev") ──────► SDK v2 Provider (zero resources, zero data sources)
        │                       Schema: elasticsearch, kibana, fleet blocks
        │                       Configure: clients.NewAPIClientFuncFromSDK
        │
        └─> NewFrameworkProvider("dev") ─────────► PF Provider
                                    All resources and data sources live here

        tf6muxserver.NewMuxServer([frameworkProvider, upgradedSdkProvider])
```

The SDK provider is a ghost. Its `DataSourcesMap` and `ResourcesMap` are empty. All production entities are PF. The only reason the SDK side exists is historical — it was the original provider before the PF migration. Keeping it means:
- Duplicate provider schema definitions (`internal/schema/connection.go` has both SDK `*schema.Schema` and PF `fwschema.Block` builders)
- `tf6muxserver` and `tf5to6server` dependencies
- SDK diagnostic types still present in `internal/clients/kibanaoapi/*` because those wrappers predate the PF migration
- Dead code in `internal/clients/config/sdk.go`, `internal/utils/utils.go`, etc.

## Goals / Non-Goals

**Goals:**
- Remove the SDK v2 provider implementation entirely.
- Remove mux wiring; serve PF provider directly.
- Convert all remaining `sdkdiag.Diagnostics` return types in `internal/clients/kibanaoapi/` to `fwdiag.Diagnostics`.
- Remove `diagutil` translation helpers that become unused.
- Remove all dead SDK-only configuration and utility code.
- Update tests to reflect a single PF provider.
- Preserve `ConvertSettingsKeyToTFFieldKey` by relocating it.

**Non-Goals:**
- Converting any resource or data source from SDK to PF (already complete).
- Changing provider schema, resource behavior, or user-visible configuration.
- Removing `terraform-plugin-sdk` test dependencies (`terraform-plugin-testing` still uses SDK types internally for test assertions).
- Refactoring the generated `kbapi` client layer.

## Decisions

### Decision: Serve PF provider directly, no mux

**Rationale**: There are zero SDK resources or data sources. The mux adds no value. Replacing `ProtoV6ProviderServerFactory` with a direct `providerserver.NewProtocol6` call eliminates two dependencies (`tf6muxserver`, `tf5to6server`) and simplifies `main.go`.

**Alternative considered**: Keep mux as a safety net in case SDK resources are needed. Rejected: the provider has been PF-only for some time; adding SDK resources would require intentional design work, not a mux fallback.

### Decision: Convert kibanaoapi wrappers in place, keep error-helpers in `diagutil`

**Rationale**: `internal/clients/kibanaoapi/status.go`, `security_role.go`, `connector.go`, and `spaces.go` all return `sdkdiag.Diagnostics`. Converting them to `fwdiag.Diagnostics` is mechanical — replace `sdkdiag.Diagnostics` with `fwdiag.Diagnostics`, replace `diag.FromErr(err)` with `diagutil.FrameworkDiagFromError(err)`, replace inline `sdkdiag.Diagnostic{}` with `fwdiag.NewErrorDiagnostic(...)`.

The `diagutil` package keeps `FrameworkDiagFromError`, `FwDiagsAsError`, and `ReportUnknownHTTPError` / `HandleStatusResponse` / `CheckHTTPErrorFromFW` (the PF-native HTTP helpers) because those are still used. Only `FrameworkDiagsFromSDK`, `SDKDiagsFromFramework`, and `SDKErrorDiag` are removed.

**Alternative considered**: Return Go `error` from kibanaoapi and let callers wrap. Rejected: the functions currently return rich diagnostics (summary + detail + severity); collapsing to `error` would lose that structure at call sites.

### Decision: Remove dead code aggressively, verify with `go build`

**Rationale**: Dead SDK code (`internal/clients/config/sdk.go`, `*FromSDK` functions, `AddConnectionSchema`, `ExpandIndividuallyDefinedSettings`, `DiffJSONSuppress`, `ExpandIndexAliases`, `FlattenIndexAliases`, `ExpandLifecycle`, `FlattenLifecycle`, `validateDataStreamOptionsVersion`) is unused outside tests. Removing it all in one change is safe because Go compilation will flag any missed references immediately.

### Decision: Preserve `ConvertSettingsKeyToTFFieldKey` in typeutils

**Rationale**: This function is used by `internal/elasticsearch/index/indices/models.go` (a live PF data source). It is pure string manipulation with no SDK dependency. Moving it to `internal/utils/typeutils` keeps it alive while allowing `internal/utils/utils.go` (which imports SDK types for dead functions) to be deleted.

### Decision: Update `connection_schema_test.go` to PF-only validation

**Rationale**: This test currently enumerates both SDK and PF entities and asserts their connection schemas match equivalent helpers. With no SDK entities, the test becomes a PF-only validator that checks every PF resource and data source has the correct `elasticsearch_connection` or `kibana_connection` block. The logic is preserved; the SDK branch is removed.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| A future PR unknowingly references removed SDK code | `go build` will fail immediately; CI lint catches it |
| kibanaoapi tests use `sdkdiag.Diagnostics` assertions and need updating | Convert test assertions to check `fwdiag.Diagnostics` severity/summary/detail |
| `provider_test.go` `TestProvider` calls `provider.New("dev").InternalValidate()` | Replace with PF provider validation (e.g., test that `NewFrameworkProvider` satisfies `fwprovider.Provider`) |
| External consumers of `provider.New()` (e.g., internal test helpers) | Search for all references to `provider.New` and `provider.NewFrameworkProvider`; update or remove |
| `internal/schema/connection.go` still has SDK schema builders (`GetEsConnectionSchema`, `GetKibanaConnectionSchema`, `GetFleetConnectionSchema`) | These are used only by `connection_schema_test.go` after this change. Evaluate whether to keep them for the test or inline test expectations. If kept, document they are test-only. |

## Migration Plan

This is a codebase-only change. No user migration is required. Deploy strategy:
1. Implement all code changes in a single branch.
2. Run `make build` to verify compilation.
3. Run `make check-lint` to verify lint.
4. Run kibanaoapi unit tests.
5. Run provider package tests (`go test ./provider/...`).
6. Acceptance tests for affected resources (security_role, spaces, connectors) to confirm PF diagnostics path end-to-end.
7. Merge.

Rollback: revert the commit. No state or migration concerns.

## Open Questions

- Should `internal/schema/connection.go` SDK schema builders be removed after `connection_schema_test.go` is rewritten, or retained as test fixtures?
- Are there any non-test references to `provider.New()` outside `provider/` and `main.go`?
