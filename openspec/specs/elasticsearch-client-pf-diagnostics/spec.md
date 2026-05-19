# elasticsearch-client-pf-diagnostics Specification

## Purpose
TBD - created by archiving change es-client-sdk-diag-to-pf-diag. Update Purpose after archive.
## Requirements
### Requirement: ES client functions return Plugin Framework diagnostics
All public functions in `internal/clients/elasticsearch/` that previously returned `github.com/hashicorp/terraform-plugin-sdk/v2/diag.Diagnostics` SHALL return `github.com/hashicorp/terraform-plugin-framework/diag.Diagnostics` instead. No function in this package SHALL import or use `terraform-plugin-sdk/v2/diag` for return values.

#### Scenario: Migrated client function returns PF diagnostics on error
- **GIVEN** a function such as `PutComponentTemplate`, `GetDataStream`, or `DeleteTransform` that previously returned SDK diagnostics
- **WHEN** the function encounters an error (e.g., network failure, API error, JSON marshal error)
- **THEN** the function SHALL return a `fwdiag.Diagnostics` value produced by `diagutil.FrameworkDiagFromError(err)` or equivalent PF helpers

#### Scenario: Migrated client function returns nil on success
- **GIVEN** a migrated function that previously returned `nil` on success as SDK diagnostics
- **WHEN** the operation completes without error
- **THEN** the function SHALL return `nil` (a valid zero value for `fwdiag.Diagnostics`) unchanged in behavior

### Requirement: Logstash client functions use PF HTTP error helpers directly
The `Put/Get/DeleteLogstashPipeline` functions in `internal/clients/elasticsearch/logstash.go` SHALL use `diagutil.CheckHTTPErrorFromFW` and return its result as `fwdiag.Diagnostics` directly. The previous `diagutil.SDKDiagsFromFramework(diagutil.CheckHTTPErrorFromFW(...))` round-trip SHALL be removed.

#### Scenario: Logstash HTTP error check uses PF helpers end-to-end
- **GIVEN** a Logstash pipeline operation that receives an HTTP error response
- **WHEN** the HTTP status indicates an error
- **THEN** the function SHALL return `fwdiag.Diagnostics` from `diagutil.CheckHTTPErrorFromFW` without wrapping it in `SDKDiagsFromFramework`

### Requirement: ElasticsearchScopedClient methods return Plugin Framework diagnostics
The methods `serverInfo`, `ClusterID`, `ID`, `ServerVersion`, `ServerFlavor`, and `EnforceMinVersion` on `ElasticsearchScopedClient` in `internal/clients/elasticsearch_scoped_client.go` SHALL return `fwdiag.Diagnostics` instead of `diag.Diagnostics` (SDK). No method on `ElasticsearchScopedClient` SHALL import `terraform-plugin-sdk/v2/diag` for method return values.

#### Scenario: ElasticsearchScopedClient.EnforceMinVersion returns PF diagnostics on error
- **GIVEN** a call to `EnforceMinVersion` where cluster info retrieval fails
- **WHEN** `serverInfo` returns an error (e.g., Elasticsearch unreachable)
- **THEN** `EnforceMinVersion` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic

#### Scenario: ElasticsearchScopedClient.EnforceMinVersion returns true for serverless
- **GIVEN** a call to `EnforceMinVersion` where the cluster reports serverless flavor
- **WHEN** flavor check succeeds with "serverless"
- **THEN** `EnforceMinVersion` SHALL return `(true, nil)` unchanged in behavior

### Requirement: KibanaScopedClient methods return Plugin Framework diagnostics
The methods `ServerVersion`, `ServerFlavor`, and `EnforceMinVersion` on `KibanaScopedClient` in `internal/clients/kibana_scoped_client.go` SHALL return `fwdiag.Diagnostics` instead of `diag.Diagnostics` (SDK). These methods SHALL consume PF diagnostics directly from `kibanaoapi.GetKibanaStatus` without bridging via `diagutil.FrameworkDiagsFromSDK`.

#### Scenario: KibanaScopedClient.EnforceMinVersion returns PF diagnostics on error
- **GIVEN** a call to `KibanaScopedClient.EnforceMinVersion` where Kibana is unreachable
- **WHEN** `GetKibanaOapiClient` or `GetKibanaStatus` fails
- **THEN** `EnforceMinVersion` SHALL return `(false, fwdiag.Diagnostics{...})` with a PF error diagnostic
- **AND** it SHALL NOT call `diagutil.FrameworkDiagsFromSDK`

#### Scenario: KibanaScopedClient.ServerVersion passes through PF diagnostics
- **GIVEN** a call to `KibanaScopedClient.ServerVersion`
- **WHEN** `kibanaoapi.GetKibanaStatus` returns PF diagnostics
- **THEN** `ServerVersion` SHALL return those diagnostics directly without wrapping

### Requirement: CompositeIDFromStr returns Plugin Framework diagnostics
`CompositeIDFromStr` in `internal/clients/api_client.go` SHALL return `(CompositeID, fwdiag.Diagnostics)` instead of `(CompositeID, sdkdiag.Diagnostics)`. The `CompositeIDFromStrFw` wrapper function SHALL be removed.

#### Scenario: CompositeIDFromStr returns PF error on malformed ID
- **GIVEN** a call to `CompositeIDFromStr` with an ID string that does not contain exactly one `/` separator
- **WHEN** the function parses the ID
- **THEN** the function SHALL return `(nil, fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(...)})` with a descriptive error

#### Scenario: CompositeIDFromStrFw is removed
- **GIVEN** the codebase after migration
- **WHEN** looking for `CompositeIDFromStrFw`
- **THEN** the function SHALL NOT exist; all former callers SHALL call `CompositeIDFromStr` directly

### Requirement: `minVersionClient` interface uses Plugin Framework diagnostics
The `minVersionClient` interface in `internal/entitycore/version_requirements.go` SHALL declare `EnforceMinVersion(ctx context.Context, minVersion *version.Version) (bool, fwdiag.Diagnostics)`, and `enforceVersionRequirements` SHALL append those diagnostics directly without wrapping via `diagutil.FrameworkDiagsFromSDK`.

#### Scenario: enforceVersionRequirements passes PF diagnostics directly
- **GIVEN** a call to `enforceVersionRequirements` where `EnforceMinVersion` returns a PF error diagnostic
- **WHEN** the diagnostics are collected
- **THEN** the caller SHALL receive `fwdiag.Diagnostics` directly without any SDK bridging

### Requirement: Callers drop FrameworkDiagsFromSDK wrappers
All call sites in `internal/elasticsearch/**` and other in-scope packages that previously wrapped migrated ES client or scoped-client function calls in `diagutil.FrameworkDiagsFromSDK(...)` SHALL call the functions directly and append the returned `fwdiag.Diagnostics` without wrapping.

#### Scenario: Caller appends PF diags directly
- **GIVEN** a caller that previously used `diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutComponentTemplate(...))...)`
- **WHEN** `PutComponentTemplate` is migrated to return `fwdiag.Diagnostics`
- **THEN** the caller SHALL use `diags.Append(elasticsearch.PutComponentTemplate(...)...)` with no intermediate bridging

### Requirement: Bridging helpers removed when unused
`diagutil.SDKErrorDiag`, `diagutil.FrameworkDiagsFromSDK`, and `diagutil.SDKDiagsFromFramework` in `internal/diagutil/translation.go` SHALL be removed. All callers of these functions SHALL be eliminated as part of this change.

#### Scenario: Helpers are deleted
- **GIVEN** that all ES client functions, scoped-client methods, kibanaoapi functions, and provider factory methods have been migrated
- **WHEN** inspecting `internal/diagutil/translation.go`
- **THEN** `SDKErrorDiag`, `FrameworkDiagsFromSDK`, and `SDKDiagsFromFramework` SHALL NOT exist in the file

#### Scenario: diagutil has no SDK diag import
- **GIVEN** the codebase after this change
- **WHEN** inspecting `internal/diagutil/translation.go`
- **THEN** it SHALL NOT import `github.com/hashicorp/terraform-plugin-sdk/v2/diag`

