## Context

All provider helper and resource code has been migrated to `go-elasticsearch` Typed API. The remaining `GetESClient()` usage is confined to acceptance-test files, where tests call raw `esapi` methods for:

1. **Preflight / setup** — creating indices, templates, policies, or roles before Terraform applies.
2. **Cleanup / destroy checks** — verifying that resources were actually deleted after `CheckDestroy` runs.
3. **State assertions** — querying Elasticsearch during `TestCheckFunc` to validate side effects.

These patterns are scattered across 25 test files. Each shares the same raw-client anti-pattern: get `*elasticsearch.Client`, build requests with `.With…` option funcs, manually read `*http.Response.Body`, and `json.Unmarshal` into ad-hoc structs or `map[string]interface{}`.

## Goals / Non-Goals

**Goals:**

- Replace every `GetESClient()` call in the listed test files with `GetESTypedClient()`.
- Convert raw `esapi` calls to strongly-typed `TypedClient` equivalents.
- Remove manual response-body reading and JSON decoding where typed structs can be used directly.
- Ensure every migrated test still compiles and passes (`go test` / acceptance tests).

**Non-Goals:**

- Do not modify any resource or helper source files (those are already migrated).
- Do not change test assertions, Terraform configs, or test step definitions.
- Do not alter `GetESClient()` or `ElasticsearchScopedClient` (the bridge already exists).
- Do not migrate `GetESClient()` calls outside the listed test files.

## Decisions

### 1. Keep test file scope focused — one file per reviewable unit

**Chosen:** Organize the migration task list by logical domain (transform, enrich, index, cluster, security, etc.) and migrate one file per top-level task.

**Rationale:** Acceptance tests are large and domain-specific. Reviewing them file-by-file keeps diffs readable and makes it easy to run targeted tests for verification.

### 2. Direct substitution of raw calls with typed equivalents

**Chosen:** For each raw call, find the equivalent `TypedClient` method and map parameters directly.

Examples:
- `esClient.Indices.Get(...)` → `typedClient.Indices.Get(ctx, indexName).Do(ctx)`
- `esClient.TransformGetTransform(...)` → `typedClient.Transform.GetTransform(ctx, transformId).Do(ctx)`
- `esClient.Security.GetRole(...)` → `typedClient.Security.GetRole(ctx, roleName).Do(ctx)`
- `esClient.EnrichGetPolicy(...)` → `typedClient.Enrich.GetPolicy(ctx, policyName).Do(ctx)`

**Rationale:** The typed API surface closely mirrors the raw API. Most migrations are mechanical substitutions.

### 3. Handle response extraction idiomatically

**Chosen:** Use the strongly-typed response structs returned by `TypedClient` methods. Where raw code did:

```go
res, err := esClient.Indices.Get(...)
// defer res.Body.Close()
// json.NewDecoder(res.Body).Decode(&body)
```

Replace with:

```go
resp, err := typedClient.Indices.Get(ctx, indexName).Do(ctx)
// resp is already a typed *indices.GetResponse or equivalent
```

**Rationale:** Eliminates `io.ReadAll`, `json.Unmarshal`, and ad-hoc struct definitions, reducing test code volume and error surface.

### 4. Preserve existing error-checking semantics

**Chosen:** Where raw code checked `res.StatusCode != 404` for existence checks, use typed API error handling or inspect the typed response. For example, a `Get` that returns a typed error on 404 can be checked with `elastic.IsNotFound(err)` or by checking the typed result.

**Rationale:** We must not change test behavior. The typed client may return errors differently (e.g., structured error types), so each migration must map the old status-code check to the equivalent typed check.

### 5. Use `context.Background()` or `t.Context()` where the raw call lacked explicit context

**Chosen:** Where raw calls passed `WithContext(ctx)`, preserve the same `ctx`. Where raw calls had no context, use `context.Background()` for the typed call.

**Rationale:** Typed API requires a context as the first argument. Using the same context semantics avoids unexpected side effects.

## Risks / Trade-offs

- **[Risk]** Some raw APIs may not yet have a perfect typed equivalent in the pinned `go-elasticsearch/v8` version.
  - **Mitigation:** For any missing typed endpoint, fall back to `GetESClient()` with an inline `TODO` comment and file a follow-up ticket. Review the generated `TypedClient` surface before starting implementation.
- **[Risk]** `kibana/streams/acc_test.go` performs raw HTTP requests against the ES query API (`/_query/view`) that may not be covered by the typed client.
  - **Mitigation:** If a raw HTTP call has no typed equivalent, leave it unchanged or migrate only the `GetESClient()` parts that do have equivalents.
- **[Risk]** Compilation errors in one test file may cascade to shared test helpers.
  - **Mitigation:** Run `make build` and `go test ./...` after every few files to catch type mismatches early.
- **[Risk]** Acceptance tests are slow; full acceptance suite execution may be impractical for every file.
  - **Mitigation:** Run targeted acceptance tests per package. The CI pipeline provides the definitive acceptance run.

## Migration Plan

1. Verify `GetESTypedClient()` is available on `ElasticsearchScopedClient` (from `typed-client-bootstrap`).
2. For each domain group, replace `GetESClient()` → `GetESTypedClient()` in the listed test files.
3. Replace raw API calls with typed equivalents and update response handling.
4. Run `go test ./internal/...` to verify compilation.
5. Run targeted acceptance tests for actively migrated packages where possible.
6. Final CI acceptance test run to confirm no regressions.
