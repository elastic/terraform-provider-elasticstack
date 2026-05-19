## 1. Investigate existing patterns

- [x] 1.1 Search `internal/clients/elasticsearch/` for any existing `.Perform(ctx)` usage or shared helper for decoding raw responses and surfacing non-2xx errors; reuse if present.
- [x] 1.2 Confirm `GetIngestPipeline` is only called from `internal/elasticsearch/ingest/pipeline_pf.go` (and any test mocks). Note callers that will need updating.
- [x] 1.3 Skim `internal/models/ingest.go` to confirm the existing `IngestPipeline` struct matches the Get pipeline API response shape (`description`, `processors`, `on_failure`, `_meta`). Adjust JSON tags only if the API uses different names.

## 2. Rework `GetIngestPipeline`

- [x] 2.1 Change the return type of `GetIngestPipeline` from `*types.IngestPipeline` to `*models.IngestPipeline`.
- [x] 2.2 Replace `typedClient.Ingest.GetPipeline().Id(name).Do(ctx)` with `typedClient.Ingest.GetPipeline().Id(name).Perform(ctx)`.
- [x] 2.3 In the response handler: defer-close the body, branch on `resp.StatusCode` — 404 returns `(nil, nil)`, 2xx decodes the body into `map[string]models.IngestPipeline` and returns the entry for `name`, any other status returns a diagnostic that includes status code and response body (matching the style of other client functions).
- [x] 2.4 Add a short code comment at the call site explaining why this endpoint uses `Perform` (typed decoding silently drops processor fields not modeled by the go-elasticsearch typed client — see issue #3002).
- [x] 2.5 Drop the now-unused `types` import if nothing else in the file references it.

## 3. Update the read path consumer

- [x] 3.1 In `readIngestPipeline` (`internal/elasticsearch/ingest/pipeline_pf.go`), rename `pipeline.Meta_` to `pipeline.Metadata` to match `models.IngestPipeline`.
- [x] 3.2 Verify the calls to `jsonListFromSlice(ctx, pipeline.Processors, ...)` and `jsonListFromSlice(ctx, pipeline.OnFailure, ...)` still compile with `[]map[string]any` element types (the helper is generic).
- [x] 3.3 Update any other field accesses on the returned pipeline (e.g. `pipeline.Description`) to match the model.

## 4. Flip the reproducer into a regression test

- [x] 4.1 In `internal/elasticsearch/ingest/issue_3002_acc_test.go`, remove the `ExpectError` line.
- [x] 4.2 Replace the godoc to describe the test as a regression for issue #3002 (no longer "reproduces" — it now "regresses against").
- [x] 4.3 Add explicit assertions that the rename processor element in state contains `override: true` (e.g. a JSON-aware check that decodes `processors[0]` and asserts the `override` key, or `resource.TestCheckResourceAttrWith` with a custom comparator).
- [x] 4.4 Add a `PlanOnly` follow-up step (or rely on the implicit post-apply refresh + plan diff) to assert no drift on a second plan.

## 5. Verify the spec change locally

- [x] 5.1 Run `make check-openspec` (or `openspec validate`) and confirm the delta validates against the existing `elasticsearch-ingest-pipeline` spec.
- [x] 5.2 Run `make build`.
- [x] 5.3 Run targeted acceptance tests: the existing ingest pipeline test suite plus `TestAccReproduceIssue3002`. Confirm all pass and that the previously-failing reproducer is now green.
- [x] 5.4 Spot-check that `make check-lint` is clean for the touched files.

## 6. Post-merge follow-up (optional, separate change)

- [ ] 6.1 File an upstream issue against `elastic/elasticsearch-specification` to add `Override` to `RenameProcessor`, referencing #3002. (Documented as out-of-scope here; tracked for completeness.)
