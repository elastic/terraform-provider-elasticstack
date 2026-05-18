## Why

Every resource and data source in the provider is already implemented on Plugin Framework (PF). The SDK v2 provider (`provider/provider.go`) registers zero resources and zero data sources — it exists solely as the left-hand side of a `tf6muxserver` that no longer serves anything. The mux adds complexity, the SDK provider schema duplicates the PF provider schema, and a layer of SDK diagnostic types (`github.com/hashicorp/terraform-plugin-sdk/v2/diag`) still pervades the kibanaoapi client layer and translation helpers. Removing the SDK provider and converting the remaining SDK diagnostics to PF diagnostics simplifies the codebase and eliminates a category of dependency that no longer serves a purpose.

## What Changes

- **Remove SDK provider** (`provider/provider.go`) and the `New()` function that constructs it.
- **Remove mux wiring** (`provider/factory.go`): collapse `ProtoV6ProviderServerFactory` to serve the PF provider directly via `tf6server` instead of `tf6muxserver` + `tf5to6server`.
- **Remove dead SDK configuration code** (`internal/clients/config/sdk.go`, `NewFromSDK`, `NewFromSDKResource`, `NewFromSDKKibanaResource`, and all `FromSDK` helpers in `internal/clients/config/`).
- **Remove dead SDK client factory methods** (`GetKibanaClientFromSDK`, `GetElasticsearchClientFromSDK`, `ConvertMetaToFactory` in `internal/clients/provider_client_factory.go`).
- **Remove dead SDK utility helpers** (`internal/utils/utils.go` SDK helpers, `internal/tfsdkutils/diffs.go`, `internal/elasticsearch/index/commons.go`, `internal/elasticsearch/index/template_sdk_shared.go`).
- **Convert kibanaoapi clients to PF diagnostics**: change `internal/clients/kibanaoapi/{status,security_role,connector,spaces}.go` from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`.
- **Update callers** (`internal/clients/kibana_scoped_client.go`, `internal/kibana/security_role/*.go`, and any other PF consumers of kibanaoapi functions) to consume PF diagnostics directly.
- **Remove dead translation helpers** from `internal/diagutil/translation.go`: `FrameworkDiagsFromSDK`, `SDKDiagsFromFramework`, `SDKErrorDiag`.
- **Update tests**: rewrite `provider_test.go` to validate PF provider only; remove mux test from `factory_test.go`; remove SDK-only connection schema validation from `connection_schema_test.go`; update kibanaoapi unit tests.
- **Preserve** `ConvertSettingsKeyToTFFieldKey` — it is used by a live PF data source (`internal/elasticsearch/index/indices/models.go`). Relocate it to a non-SDK package (e.g. `internal/utils/typeutils`).

## Capabilities

### New Capabilities
- `provider-pf-only`: The Terraform provider serves exclusively a Plugin Framework provider without SDK v2 mux fallback.
- `kibanaoapi-pf-diagnostics`: All public functions in `internal/clients/kibanaoapi/` return `github.com/hashicorp/terraform-plugin-framework/diag.Diagnostics` instead of SDK diagnostics.

### Modified Capabilities
- `provider-client-factory`: Remove the SDK `meta` injection requirement. The factory is injected only into PF `ProviderData` / `ResourceData`.
- `elasticsearch-client-pf-diagnostics`: Extend scope to cover removal of `internal/diagutil/translation.go` translation helpers (`FrameworkDiagsFromSDK`, `SDKDiagsFromFramework`, `SDKErrorDiag`) once no callers remain.

## Impact

- **Provider entry point** (`main.go`, `provider/factory.go`): direct PF server factory, no mux.
- **Client layer** (`internal/clients/config/*`, `internal/clients/api_client.go`, `internal/clients/provider_client_factory.go`): all SDK configuration and diagnostic paths removed.
- **Kibana OpenAPI client wrappers** (`internal/clients/kibanaoapi/*`): return type changes from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`.
- **PF resource implementations** (`internal/kibana/security_role/*.go`, connectors, spaces, etc.): remove `diagutil.FrameworkDiagsFromSDK()` round-trips.
- **Diagnostic utilities** (`internal/diagutil/translation.go`): three exported helpers removed.
- **Dead helper packages** (`internal/utils/utils.go` SDK parts, `internal/tfsdkutils/diffs.go`, `internal/elasticsearch/index/commons.go`, `internal/elasticsearch/index/template_sdk_shared.go`): removed.
- **Tests**: PF-only provider validation; framework-only connection schema checks; updated kibanaoapi unit tests.
- **No user-visible behavioral change**: all resources and data sources are already PF; provider schema is unchanged.
