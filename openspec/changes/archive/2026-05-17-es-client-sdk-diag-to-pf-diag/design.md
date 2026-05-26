## Context

The `internal/clients/elasticsearch` package contains ~35 functions across 10 files that still return `sdkdiag.Diagnostics`. Because all ES Terraform resources have already been migrated to the Plugin Framework, every caller must wrap each call in `diagutil.FrameworkDiagsFromSDK(...)` to bridge the type mismatch. A similar bridging pattern exists for `ElasticsearchScopedClient.EnforceMinVersion`, `KibanaScopedClient.EnforceMinVersion`, and `CompositeIDFromStr` in `internal/clients/api_client.go`.

The already-migrated files (`ilm.go`, `watch.go`, `ml_job.go`, `script.go`, `inference.go`, `alias.go`) demonstrate the established pattern. This change extends that pattern to all remaining client functions and to the scoped-client methods.

**Files to migrate in `internal/clients/elasticsearch/`:**

| File | SDK-diag functions |
|---|---|
| `cluster.go` | `GetClusterInfo` (1) |
| `datastream.go` | `Put/Get/DeleteDataStream` (3) |
| `enrich.go` | `Get/Put/Delete/ExecuteEnrichPolicy` (4) |
| `ingest.go` | `Put/Get/DeleteIngestPipeline` (3) |
| `logstash.go` | `Put/Get/DeleteLogstashPipeline` (3) |
| `security.go` | `GetUser`, `Put/Get/DeleteRole` (4 remaining) |
| `settings.go` | `Put/GetSettings` (2) |
| `snapshot.go` | `Put/Get/Delete SnapshotRepository + Put/Get/DeleteSlm` (6) |
| `templates.go` | `Put/Get/DeleteComponentTemplate` (3 remaining) |
| `transform.go` | `Put/Get/Delete Transform + GetTransformStats + start/stopTransform` (6) |

**Additional in-scope functions:**

- `ElasticsearchScopedClient.serverInfo`, `ClusterID`, `ID`, `ServerVersion`, `ServerFlavor`, `EnforceMinVersion` in `internal/clients/elasticsearch_scoped_client.go`
- `KibanaScopedClient.ServerVersion`, `ServerFlavor`, `EnforceMinVersion` in `internal/clients/kibana_scoped_client.go`
- `CompositeIDFromStr` in `internal/clients/api_client.go`; `CompositeIDFromStrFw` becomes redundant and is removed

## Goals / Non-Goals

**Goals:**

- Remove all `sdkdiag.Diagnostics` return types from the affected functions.
- Eliminate all `diagutil.FrameworkDiagsFromSDK(...)` call sites in callers of migrated functions.
- Clean up `diagutil.SDKErrorDiag`, `diagutil.FrameworkDiagsFromSDK`, and `diagutil.SDKDiagsFromFramework` once no callers remain.
- Leave no intermediate mixed state: all ES client functions should use PF diag after a single PR.

**Non-Goals:**

- `internal/clients/kibanaoapi/` functions returning SDK diag (Kibana API client, separate concern).
- `internal/kibana/` and `internal/fleet/` callers of `FrameworkDiagsFromSDK` for Kibana/Fleet version gates — only the `entitycore/version_requirements.go` path (which is driven by ES/Kibana `EnforceMinVersion`) is in scope.
- Any behavior changes to resources; this is a pure type-layer cleanup.

## Decisions

- **Single comprehensive PR**: The entire migration is mechanically straightforward — no logic changes, no API shape changes. A single PR eliminates intermediate mixed state and enables immediate cleanup of bridging helpers. This is the approach recommended by the research analysis.
- **`CompositeIDFromStrFw` removal**: Once `CompositeIDFromStr` itself returns `fwdiag.Diagnostics`, `CompositeIDFromStrFw` is redundant. All callers should be updated to call `CompositeIDFromStr` directly, and the wrapper removed.
- **`diagutil` cleanup**: `SDKErrorDiag`, `FrameworkDiagsFromSDK`, and `SDKDiagsFromFramework` serve as bridges between SDK and PF diag types. Once no ES/ES-scoped-client callers remain, these helpers should be removed in the same PR. If Kibana/Fleet code still uses them (for non-`EnforceMinVersion` paths), keep them and note which callers remain.
- **`logstash.go` simplification**: Currently uses `diagutil.SDKDiagsFromFramework(diagutil.CheckHTTPErrorFromFW(...))` — a pointless SDK→PF→SDK round-trip. After migration, use `diagutil.CheckHTTPErrorFromFW(...)` directly and return `fwdiag.Diagnostics`.
- **`ElasticsearchScopedClient` private methods**: `serverInfo`, `ClusterID`, `ID`, `ServerVersion`, `ServerFlavor` all return SDK diag; migrating them is required to migrate `EnforceMinVersion`.

## Risks / Trade-offs

- **[Risk] Large diff with many files touched** — Mitigation: the transformation is mechanical; PR description should mark it explicitly as a global find-replace so reviewers can audit it efficiently rather than performing line-by-line logic analysis.
- **[Risk] `diagutil` cleanup has residual callers** — Mitigation: verify with grep before deletion; if Kibana/Fleet code outside the defined scope still uses `FrameworkDiagsFromSDK`, keep the helper and document remaining callers in the PR description.
- **[Risk] KibanaScopedClient callers miss update** — Mitigation: grep for all `FrameworkDiagsFromSDK` call sites after client migration; compiler will also catch type mismatches.

## Migration Pattern

```go
// Before (SDK diag)
func PutComponentTemplate(...) sdkdiag.Diagnostics {
    _, err = typedClient.Cluster.PutComponentTemplate(...).Do(ctx)
    if err != nil {
        return sdkdiag.FromErr(err)
    }
    return nil
}
// Caller
diags.Append(diagutil.FrameworkDiagsFromSDK(elasticsearch.PutComponentTemplate(...))...)

// After (PF diag)
func PutComponentTemplate(...) fwdiag.Diagnostics {
    _, err = typedClient.Cluster.PutComponentTemplate(...).Do(ctx)
    if err != nil {
        return diagutil.FrameworkDiagFromError(err)
    }
    return nil
}
// Caller
diags.Append(elasticsearch.PutComponentTemplate(...)...)
```

For `logstash.go` the `SDKDiagsFromFramework(CheckHTTPErrorFromFW(...))` round-trip simplifies to:

```go
// Before
if d := diagutil.SDKDiagsFromFramework(diagutil.CheckHTTPErrorFromFW(res, "...")); d.HasError() {
    return d
}
// After
if d := diagutil.CheckHTTPErrorFromFW(res, "..."); d.HasError() {
    return d
}
```

## Open Questions

- **`diagutil.SDKErrorDiag` removal**: Once the ES client functions are migrated, no ES-layer code needs `SDKErrorDiag`. Is it safe to remove in the same PR, or should it wait until all Kibana/Fleet SDK diag callers are also migrated? Verify by grepping for remaining callers — if none exist outside the ES scope after migration, remove it. If Kibana/Fleet callers remain, retain the helper and note it.
- **`EnforceMinVersion` is adjacent but out of scope?**: `ElasticsearchScopedClient.EnforceMinVersion` and `KibanaScopedClient.EnforceMinVersion` — now **in scope** per contributor direction (see issue #3058 comment by @tobio).
- **`CompositeIDFromStr` in `api_client.go`**: Also **in scope** per contributor direction.
