## Why

All Elasticsearch resources and data sources have been migrated to the Plugin Framework (PF). However, the `internal/clients/elasticsearch` package still exposes functions returning `sdkdiag.Diagnostics` (from `terraform-plugin-sdk/v2/diag`), forcing every caller to wrap invocations in `diagutil.FrameworkDiagsFromSDK(...)`. Similarly, `ElasticsearchScopedClient.EnforceMinVersion`, `KibanaScopedClient.EnforceMinVersion`, and `CompositeIDFromStr` still return SDK diagnostics, causing additional bridging boilerplate throughout the codebase.

This is a pure type-layer cleanup: replace all SDK diagnostic returns in the affected client functions with Plugin Framework diagnostics and remove the bridging wrappers in callers.

## What Changes

- Migrate all 10 remaining files in `internal/clients/elasticsearch/` from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`.
- Migrate `ElasticsearchScopedClient` methods (`serverInfo`, `ClusterID`, `ID`, `ServerVersion`, `ServerFlavor`, `EnforceMinVersion`) in `internal/clients/elasticsearch_scoped_client.go` to return PF diagnostics.
- Migrate `KibanaScopedClient` methods (`ServerVersion`, `ServerFlavor`, `EnforceMinVersion`) in `internal/clients/kibana_scoped_client.go` to return PF diagnostics.
- Migrate `CompositeIDFromStr` in `internal/clients/api_client.go` to return `fwdiag.Diagnostics`, and remove the now-redundant `CompositeIDFromStrFw` wrapper.
- Update the `minVersionClient` interface in `internal/entitycore/version_requirements.go` to use PF diagnostics and remove the `FrameworkDiagsFromSDK` call.
- Update all callers in `internal/elasticsearch/**` and elsewhere to drop `diagutil.FrameworkDiagsFromSDK(...)` wrappers.
- Remove `diagutil.SDKErrorDiag`, `diagutil.FrameworkDiagsFromSDK`, and `diagutil.SDKDiagsFromFramework` once they have no remaining callers after the migration.

## Capabilities

### Modified Capabilities
- `elasticsearch-client-pf-diagnostics`: The diagnostic contract for the Elasticsearch client layer — requires all client functions to use `fwdiag.Diagnostics` exclusively.

## Impact

- **Specs**: delta spec at `openspec/changes/es-client-sdk-diag-to-pf-diag/specs/elasticsearch-client-pf-diagnostics/spec.md`
- **Client files**: `internal/clients/elasticsearch/` (10 files), `internal/clients/elasticsearch_scoped_client.go`, `internal/clients/kibana_scoped_client.go`, `internal/clients/api_client.go`
- **Entitycore**: `internal/entitycore/version_requirements.go`
- **Callers**: all files in `internal/elasticsearch/**` that currently call `diagutil.FrameworkDiagsFromSDK(...)`
- **Diagutil cleanup**: `internal/diagutil/translation.go` — removal of `SDKErrorDiag`, `FrameworkDiagsFromSDK`, `SDKDiagsFromFramework`
