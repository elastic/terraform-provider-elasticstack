## Context

The `elasticstack_elasticsearch_ingest_pipeline` resource intentionally treats each processor and `on_failure` entry as opaque JSON in its schema, so users can configure processor fields the provider has no first-class knowledge of. The write path honours this: `PutIngestPipeline` (`internal/clients/elasticsearch/ingest.go:41`) marshals a `map[string]any` body and sends it with `Ingest.PutPipeline(name).Raw(...)`, so every configured field reaches Elasticsearch.

The read path breaks the contract. `GetIngestPipeline` (`internal/clients/elasticsearch/ingest.go:53`) calls `Ingest.GetPipeline().Id(name).Do(ctx)`, which decodes the response into `types.IngestPipeline` and its `[]ProcessorContainer` processors. Each `ProcessorContainer` has a typed struct per processor kind (e.g. `types.RenameProcessor`), and any field that struct doesn't model is dropped on the floor before `readIngestPipeline` (`internal/elasticsearch/ingest/pipeline_pf.go`) re-marshals the value into a `jsontypes.Normalized` for state.

Issue #3002 is the concrete instance: `types.RenameProcessor` in `go-elasticsearch/v8@v8.19.6` (and at least the surrounding 8.19.x line) has no `Override` field, so a configuration with `override = true` round-trips to `override` absent in state, and Terraform rejects the apply with "Provider produced inconsistent result after apply".

This is structural — `RenameProcessor.Override` is one missing field today, but the same drift will appear for any future field the typed client lags behind on, in any processor type, and for processors nested under `on_failure`. The repo already keeps `internal/models/ingest.go`'s `IngestPipeline` (with `[]map[string]any` processors) for exactly this kind of opacity, but it is currently unused on the GET path.

## Goals / Non-Goals

**Goals:**

- Make a refresh of an ingest pipeline preserve every field the Elasticsearch Get pipeline API returns for each processor and `on_failure` handler, regardless of whether the field is modeled by the go-elasticsearch typed client.
- Fix issue #3002 with the minimal possible change to the read path; do not touch the schema, plan modifiers, or write path.
- Turn the existing #3002 reproducer (`internal/elasticsearch/ingest/issue_3002_acc_test.go`) into a positive regression test.

**Non-Goals:**

- Patching `elastic/elasticsearch-specification` to add `Override` to `RenameProcessor`. Worth doing upstream, but irrelevant to whether the provider preserves arbitrary fields — we want forward-compatibility anyway.
- Reworking the write path or any processor data sources (`processor_*_pf_data_source.go`). They build JSON locally and are unaffected by the typed-decode bug.
- Introducing a new client abstraction. We continue to use the `go-elasticsearch/v8` typed client's request builder; only the response decoding changes.

## Decisions

### Decision: Use the typed client's `.Perform(ctx)` for GET, decode into `models.IngestPipeline`

Switch `GetIngestPipeline` from `.Do(ctx)` (typed response) to `.Perform(ctx)` (raw `*http.Response`), then `json.Decode` the body into `map[string]models.IngestPipeline` where `models.IngestPipeline` holds `Processors []map[string]any`, `OnFailure []map[string]any`, `Description *string`, and `Metadata map[string]any` (mapped via the existing `_meta` JSON tag). The function's signature changes from returning `*types.IngestPipeline` to `*models.IngestPipeline`.

**Alternatives considered:**

- *Drop to `esapi.IngestGetPipelineRequest`* — bypasses the typed request builder for no benefit; we lose the consistent style used elsewhere in this client package and gain nothing the typed client doesn't already give us via `Perform`.
- *Decode into `map[string]json.RawMessage` and lazily re-marshal* — adds a second indirection in the read path without any benefit; we always want the parsed shape for `description`, `_meta`, etc.
- *Patch upstream and bump go-elasticsearch* — fixes only `override` on rename. Doesn't solve the general drift problem, and our users would still hit it whenever the upstream spec lags behind a new processor field.

### Decision: Reuse the existing `models.IngestPipeline` rather than introduce a new type

`models.IngestPipeline` already has exactly the shape we need (it predates the typed-client migration). Reusing it keeps a single canonical representation of an opaque pipeline body and avoids introducing a parallel struct in the clients package.

Consumer impact: in `readIngestPipeline`, accesses to `pipeline.Meta_` (the typed client's field) become `pipeline.Metadata` (the model field). The `Processors` and `OnFailure` element types change from typed `ProcessorContainer` to `map[string]any`; `jsonListFromSlice` is already generic (`[T any]`) and continues to work without changes.

### Decision: 404 detection and non-2xx handling live in the GET function

`.Perform(ctx)` does not return a typed not-found error and does not raise on non-2xx. We must:

1. Treat `resp.StatusCode == http.StatusNotFound` as "not found" — return `(nil, nil)` so `readIngestPipeline` can drop the resource from state, mirroring today's `IsNotFoundElasticsearchError(err)` branch.
2. Treat any other non-2xx status as an error — decode the body into the typed client's standard error envelope (or fall back to including the status + body in a diagnostic) and return it. We should reuse an existing helper if one exists in `internal/clients/elasticsearch/` to keep error messages consistent with `PutIngestPipeline` / `DeleteIngestPipeline`; otherwise add a small local helper.
3. Always read-and-close the response body before returning, including in the 404 and error paths.

### Decision: Convert the reproducer into a positive regression test

Today `TestAccReproduceIssue3002` asserts an `ExpectError` matching `Provider produced inconsistent result after apply.*override`. After the fix, the apply succeeds, so the `ExpectError` is removed and the test asserts `override` is present in state (e.g. via `resource.TestCheckResourceAttr` on a processor element, or a JSON-aware check that inspects the rendered processor body).

## Risks / Trade-offs

- **Risk: silently regressing non-404 error reporting.** Today users see `diagutil.FrameworkDiagFromError(err)` for any non-2xx that isn't 404. Switching to `Perform` means we synthesise that error ourselves; a bug here could swallow real failures.
  - *Mitigation:* explicit task to add a unit-level check (or, if the existing acceptance suite already covers e.g. malformed pipeline name returning 400, rely on that). Mirror the error message style of the existing typed-error path.
- **Risk: 404 detection diverges from `IsNotFoundElasticsearchError`.** The helper accepts any error shape; we need to check `StatusCode == 404` directly.
  - *Mitigation:* assert on real Elasticsearch behaviour in the existing "pipeline not found on refresh" scenario, which is already covered by acceptance tests.
- **Trade-off: we give up the typed-client conveniences (typed response, automatic error wrapping) on this single endpoint.** That's exactly the point — the typed response is the bug — but it means contributors touching `GetIngestPipeline` need to be aware that this endpoint is special. Document this with a short comment at the call site explaining why we use `Perform`.
- **Trade-off: any test that mocks `GetIngestPipeline` by constructing a `types.IngestPipeline` directly will need updating.** A quick search of the codebase should confirm scope; the function is internal and likely only called from `readIngestPipeline`.
