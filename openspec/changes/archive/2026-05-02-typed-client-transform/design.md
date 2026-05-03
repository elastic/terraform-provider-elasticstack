## Context

`internal/clients/elasticsearch/transform.go` currently contains seven helper functions that perform Elasticsearch Transform API operations using the raw `esapi` client (`*elasticsearch.Client`). Each helper manually constructs HTTP requests via the `esapi` namespace, marshals request bodies into custom `models` types, and unmarshals JSON responses into parallel custom structs. The `go-elasticsearch/v8` library provides strongly-typed equivalents for all of these APIs via `elasticsearch.TypedClient`.

The typed client is already available through `ElasticsearchScopedClient.GetESTypedClient()` (introduced in `typed-client-bootstrap`). The scoped client caches the typed instance so callers do not pay the `ToTyped()` cost on every invocation.

The transform resource (`internal/elasticsearch/transform/transform.go`) is a Terraform Plugin SDK resource that consumes these helpers and performs schema-level state mapping. The resource supports version-gated capabilities such as `destination.aliases` (ES >= 8.8.0) that may not be fully represented in the current typed-client generated types.

## Goals / Non-Goals

**Goals:**
- Rewrite all functions in `internal/clients/elasticsearch/transform.go` to use the typed client APIs.
- Update the transform resource and tests to call the migrated helpers.
- Preserve exact Terraform-visible behavior: identical state mapping, identical error messages, identical not-found handling, identical start/stop semantics.
- Replace custom model types with typed-client equivalents where the shapes match one-to-one.

**Non-Goals:**
- Adding new Terraform resources or data sources.
- Changing schema definitions, validation, or plan modifiers.
- Modifying provider-level client construction or the scoped-client type itself.
- Removing support for version-gated fields that are not yet present in the typed-client specification.

## Decisions

**1. Use `client.GetESTypedClient()` to obtain the typed client.**
- **Rationale**: The scoped client already caches the typed client. Using the accessor keeps the change consistent with the bootstrap pattern and avoids per-call `ToTyped()` overhead.
- **Alternative considered**: Calling `client.GetESClient().ToTyped()` inline. Rejected because it would bypass the cached instance and add unnecessary allocations.

**2. Migrate Delete, Start, Stop, and Stats helpers to fully-typed `.Do()` calls.**
- **Rationale**: These APIs have simple request/response shapes and the typed client models them completely. `DeleteTransform`, `StartTransform`, `StopTransform`, and `GetTransformStats` can be expressed entirely through the typed API with no custom JSON handling.
- **Alternative considered**: Keeping raw `esapi` for these. Rejected because the typed API is a direct replacement with no gaps.

**3. Use `.Raw()` for Put and Update Transform request bodies.**
- **Rationale**: The typed `puttransform.Request.Dest` is `types.TransformDestination`, which does not include the `aliases` field that the resource supports (ES >= 8.8.0). Rather than dropping aliases support or constructing hybrid JSON, we keep the existing `models.Transform` body-building logic, marshal it to JSON, and pass the payload via the typed API's `.Raw()` method. This still moves the call off the raw `esapi` namespace and onto the typed client's transport, error handling, and parameter helpers (`Timeout()`, `DeferValidation()`).
- **Alternative considered**: Building `puttransform.Request` directly with typed structures. Rejected because it would silently drop `destination.aliases`, violating the zero-behavior-change requirement.

**4. Keep manual response decoding for Get Transform.**
- **Rationale**: The typed `gettransform.Response` uses `types.TransformSummary`, whose `Dest` field is `types.ReindexDestination` — again lacking `aliases`. To preserve read-back of aliases and all other fields, `GetTransform` will use the typed client's `.Perform()` to obtain the raw `*http.Response`, then decode into the existing `models.GetTransformResponse` structure exactly as today.
- **Alternative considered**: Using `.Do()` and mapping from `types.TransformSummary`. Rejected because it would lose `aliases` on read.

**5. Preserve `diag.Diagnostics` return type for SDK compatibility.**
- **Rationale**: The transform resource is a Plugin SDK resource. Helper signatures already return `diag.Diagnostics`. Keeping this return type avoids expanding the change into SDK-to-PF migration territory.
- **Alternative considered**: Migrating to `fwdiag.Diagnostics`. Rejected because that belongs to a separate PF migration effort.

**6. Retain custom models for request body construction and Get response decoding; remove response-only wrappers where possible.**
- **Rationale**: `models.Transform` and its nested types are still needed for JSON body construction (Put/Update) and response decoding (Get). `models.PutTransformParams`, `models.UpdateTransformParams`, and `models.TransformStats` can be removed because their shapes are superseded by typed-client parameters and `types.TransformStats`. `models.GetTransformResponse` is retained as the decode target for Get Transform.
- **Alternative considered**: Removing all custom models unconditionally. Rejected because the gap in typed-client types for aliases makes them still necessary.

## Risks / Trade-offs

- **[Risk]** The typed-client `types.TransformDestination` and `types.ReindexDestination` do not model `aliases`. For Put/Update we work around this with `.Raw()`; for Get we work around it with manual decode. If a future `go-elasticsearch` version adds aliases to these types, the workaround can be replaced with native typed structures.
  - **Mitigation**: Document the workaround in code comments and leave a TODO referencing the upstream gap.
- **[Risk]** `types.Settings.DocsPerSecond` is `*float32` while our model uses `*float64`. Converting between them could introduce precision loss.
  - **Mitigation**: Since we use `.Raw()` for Put/Update, we bypass `types.Settings` entirely for write paths. On read, `types.TransformSummary.Settings` is not consumed because we decode the raw response.
- **[Risk]** Typed-client error responses are returned as `*types.ElasticsearchError`. The current `diagutil.CheckError` expects an `*http.Response`. When using `.Do()`, the error is already parsed.
  - **Mitigation**: For helpers using `.Do()`, wrap the returned `error` directly into diagnostics. For helpers using `.Perform()`, continue using `diagutil.CheckError` on the raw response.
- **[Risk]** Manual JSON decode for Get Transform means we do not benefit from typed-client response validation.
  - **Mitigation**: The decode target (`models.GetTransformResponse`) has been stable for multiple provider releases. Acceptance tests will verify the mapping remains correct.

## Migration Plan

1. Update `PutTransform` to obtain the typed client, build the `models.Transform` body as today, marshal to JSON, and call `typedClient.Transform.PutTransform(name).Raw(body).Timeout(...).DeferValidation(...).Do(ctx)`.
2. Update `GetTransform` to use `typedClient.Transform.GetTransform().TransformId(name).Perform(ctx)`, then decode the response into `models.GetTransformResponse` and search for the matching transform.
3. Update `GetTransformStats` to use `typedClient.Transform.GetTransformStats(name).Do(ctx)`, then search the returned `[]types.TransformStats` for the matching ID.
4. Update `UpdateTransform` to use the typed client with `.Raw()` for the body, mirroring the Put approach.
5. Update `DeleteTransform` to use `typedClient.Transform.DeleteTransform(name).Force(true).Do(ctx)`.
6. Update `startTransform` to use `typedClient.Transform.StartTransform(name).Timeout(...).Do(ctx)`.
7. Update `stopTransform` to use `typedClient.Transform.StopTransform(name).Timeout(...).Do(ctx)`.
8. Update `internal/elasticsearch/transform/transform.go` to call the new helpers (signatures remain the same, so call sites should need zero or minimal changes).
9. Remove now-unused custom model types (`models.PutTransformParams`, `models.UpdateTransformParams`, `models.TransformStats`, `models.GetTransformStatsResponse`) from `internal/models/transform.go` if they are no longer referenced elsewhere.
10. Run `make build` and targeted acceptance tests for `elasticstack_elasticsearch_transform`.

## Open Questions

- None — the scoped client and typed client surface are already well understood.
