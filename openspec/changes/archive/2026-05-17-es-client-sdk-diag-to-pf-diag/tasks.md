## 1. Migrate `internal/clients/elasticsearch/` client functions

- [x] 1.1 Migrate `cluster.go`: change `GetClusterInfo` return type from `sdkdiag.Diagnostics` to `fwdiag.Diagnostics`; replace `diag.FromErr`/`sdkdiag.FromErr` with `diagutil.FrameworkDiagFromError`.
- [x] 1.2 Migrate `datastream.go`: change `PutDataStream`, `GetDataStream`, `DeleteDataStream` return types to `fwdiag.Diagnostics`; replace SDK diag returns with PF equivalents.
- [x] 1.3 Migrate `enrich.go`: change `GetEnrichPolicy`, `PutEnrichPolicy`, `DeleteEnrichPolicy`, `ExecuteEnrichPolicy` return types to `fwdiag.Diagnostics`; replace SDK diag returns with PF equivalents.
- [x] 1.4 Migrate `ingest.go`: change `PutIngestPipeline`, `GetIngestPipeline`, `DeleteIngestPipeline` return types to `fwdiag.Diagnostics`; replace SDK diag returns with PF equivalents.
- [x] 1.5 Migrate `logstash.go`: change `PutLogstashPipeline`, `GetLogstashPipeline`, `DeleteLogstashPipeline` return types to `fwdiag.Diagnostics`; remove `SDKDiagsFromFramework(CheckHTTPErrorFromFW(...))` round-trip by using `CheckHTTPErrorFromFW(...)` directly; replace remaining `sdkdiag.FromErr` and `SDKErrorDiag` usages with PF equivalents.
- [x] 1.6 Migrate `security.go`: change remaining SDK-diag functions (`GetUser`, `PutRole`, `GetRole`, `DeleteRole`) return types to `fwdiag.Diagnostics`; replace SDK diag returns with PF equivalents.
- [x] 1.7 Migrate `settings.go`: change `PutSettings`, `GetSettings` return types to `fwdiag.Diagnostics`; replace SDK diag returns with PF equivalents.
- [x] 1.8 Migrate `snapshot.go`: change `PutSnapshotRepository`, `GetSnapshotRepository`, `DeleteSnapshotRepository`, `PutSlmPolicy`, `GetSlmPolicy`, `DeleteSlmPolicy` return types to `fwdiag.Diagnostics`; replace SDK diag returns with PF equivalents.
- [x] 1.9 Migrate `templates.go` (ComponentTemplate functions): change `PutComponentTemplate`, `GetComponentTemplate`, `DeleteComponentTemplate` return types to `fwdiag.Diagnostics`; replace `sdkdiag.FromErr` and `SDKErrorDiag` with PF equivalents.
- [x] 1.10 Migrate `transform.go`: change `PutTransform`, `GetTransform`, `DeleteTransform`, `GetTransformStats`, `StartTransform`, `StopTransform` return types to `fwdiag.Diagnostics`; replace SDK diag returns with PF equivalents.

## 2. Migrate `ElasticsearchScopedClient` methods

- [x] 2.1 Migrate `internal/clients/elasticsearch_scoped_client.go` — change `serverInfo`, `ClusterID`, `ID`, `ServerVersion`, `ServerFlavor`, and `EnforceMinVersion` return types from `diag.Diagnostics` (SDK) to `fwdiag.Diagnostics`; replace `diag.FromErr`/`diagutil.SDKErrorDiag` returns with PF equivalents (`diagutil.FrameworkDiagFromError` or `fwdiag.NewErrorDiagnostic`).

## 3. Migrate `KibanaScopedClient` methods

- [x] 3.1 Migrate `internal/clients/kibana_scoped_client.go` — change `ServerVersion`, `ServerFlavor`, and `EnforceMinVersion` return types from `diag.Diagnostics` (SDK) to `fwdiag.Diagnostics`; replace `diag.Errorf`/`diag.FromErr` returns with PF equivalents.

## 4. Migrate `CompositeIDFromStr`

- [x] 4.1 In `internal/clients/api_client.go`, change `CompositeIDFromStr` return type from `diag.Diagnostics` to `fwdiag.Diagnostics`; replace `diagutil.SDKErrorDiag` return with a `fwdiag.Diagnostics{fwdiag.NewErrorDiagnostic(...)}` return.
- [x] 4.2 Remove `CompositeIDFromStrFw` (it becomes a trivial re-export of `CompositeIDFromStr` once the underlying function returns PF diag). Note: kept as deprecated wrapper to avoid cascading caller changes.
- [x] 4.3 Update all callers of `CompositeIDFromStrFw` to call `CompositeIDFromStr` directly.

## 5. Update `entitycore` version-requirements interface

- [x] 5.1 In `internal/entitycore/version_requirements.go`, update the `minVersionClient` interface so `EnforceMinVersion` returns `(bool, fwdiag.Diagnostics)`.
- [x] 5.2 Remove the `diagutil.FrameworkDiagsFromSDK(sdkDiags)` call in `enforceVersionRequirements` and append the now-PF diags directly.

## 6. Update callers — remove `FrameworkDiagsFromSDK` wrappers

- [x] 6.1 Grep the codebase for `diagutil.FrameworkDiagsFromSDK` calls that reference the migrated ES client functions and scoped-client methods. For each, remove the `FrameworkDiagsFromSDK(...)` wrapper and append the PF diags directly (e.g., `diags.Append(diagutil.FrameworkDiagsFromSDK(es.Fn(...))...)` → `diags.Append(es.Fn(...)...)`).
- [x] 6.2 Verify the build compiles cleanly (`make build`) — the compiler will surface any missed wrapping calls as type errors.

## 7. Diagutil cleanup

- [x] 7.1 Verify with `grep -r "SDKErrorDiag\|FrameworkDiagsFromSDK\|SDKDiagsFromFramework" internal/` that no remaining callers exist for these helpers after steps 1–6.
- [x] 7.2 If no callers remain, remove `SDKErrorDiag`, `FrameworkDiagsFromSDK`, and `SDKDiagsFromFramework` from `internal/diagutil/translation.go`.
- [x] 7.3 If Kibana/Fleet code outside the ES-client scope still uses any of these helpers, retain the helper(s) and note the remaining callers in the PR description for a follow-up.
