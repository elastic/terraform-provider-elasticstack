## Why

The `elasticstack_elasticsearch_ingest_pipeline` resource produces drift ("Provider produced inconsistent result after apply") whenever a processor body contains a field that the go-elasticsearch typed client does not model — for example `override` on the rename processor (issue [#3002](https://github.com/elastic/terraform-provider-elasticstack/issues/3002)). The write path sends raw JSON and preserves every field, but the read path deserializes through `types.IngestPipeline`, silently dropping any field the typed schema omits before the JSON round-trip into state.

## What Changes

- Change `GetIngestPipeline` in `internal/clients/elasticsearch/ingest.go` to call the Get pipeline API via the typed client's raw transport (`.Perform(ctx)`) instead of `.Do(ctx)`, decoding the response into a model that stores each processor as an opaque `map[string]any`.
- Reintroduce/reuse `internal/models/ingest.go`'s `IngestPipeline` (with `Processors []map[string]any`, `OnFailure []map[string]any`, `Description *string`, `Metadata map[string]any` mapped via `_meta`) as the return type of `GetIngestPipeline`.
- Update `readIngestPipeline` in `internal/elasticsearch/ingest/pipeline_pf.go` to consume the new return type (rename the metadata field access and adjust any type signatures), keeping the existing `jsonListFromSlice` / state-building logic.
- Preserve current behaviour for 404s (return `nil`, no diagnostic) and for non-2xx responses (surface as Terraform error diagnostic).
- Remove the `ExpectError` from `internal/elasticsearch/ingest/issue_3002_acc_test.go` so the reproducer becomes a positive regression test asserting that `override = true` survives a refresh.

No schema, plan-modifier, write-path, or public API surface changes. No breaking changes for users.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `elasticsearch-ingest-pipeline`: Strengthen the Read (REQ-015–REQ-017) and JSON-mapping (REQ-020–REQ-022) requirements so that processor and `on_failure` objects refreshed from the API preserve **every** field the server returns, including fields not modeled by the go-elasticsearch typed client.

## Impact

- **Code**: `internal/clients/elasticsearch/ingest.go` (GET path), `internal/models/ingest.go` (resurrected/kept in use), `internal/elasticsearch/ingest/pipeline_pf.go` (consumer of the new return type), `internal/elasticsearch/ingest/issue_3002_acc_test.go` (flipped from negative to positive assertion).
- **APIs / dependencies**: No new dependencies. Uses `Perform(ctx)` already exposed by `go-elasticsearch/v8` `typedapi`.
- **Tests**: Existing ingest pipeline acceptance tests must continue to pass. The issue #3002 reproducer becomes the regression test. Consider adding a second case covering at least one other field absent from the typed model (e.g. a future-proofing assertion with an arbitrary unknown processor option) to lock in forward-compatibility.
- **Risk**: Low — read-path-only refactor. Main risks are (a) mis-handling non-2xx responses now that the typed error wrapping is gone, and (b) subtly different 404 detection. Both are covered by the existing acceptance suite and explicit task items.
- **Out of scope**: Filing an upstream issue against `elastic/elasticsearch-specification` to add `Override` to `RenameProcessor`. Worth doing separately, but not required — this change is the structural fix that protects against current and future spec gaps in any processor.
