## Context

`internal/clients/elasticsearch/index.go` is the provider’s largest Elasticsearch helper file (~29KB). It contains hand-written request construction (`json.Marshal` into `bytes.NewReader`), raw `esapi.Response` decoding, and an extensive set of custom models (`models.Policy`, `models.IndexTemplate`, `models.ComponentTemplate`, `models.Index`, `models.DataStream`, `models.IngestPipeline`, etc.) that shadow types already present in the `go-elasticsearch/v8` typed API.

The typed client bridge (`GetESTypedClient()`) was added in the `typed-client-bootstrap` change, making it possible to incrementally migrate individual helper files. All other typed-client phases have been completed; the Index APIs are the remaining major surface.

## Goals / Non-Goals

**Goals:**
- Migrate every function in `index.go` to use `elasticsearch.TypedClient`.
- Replace custom models with typed API equivalents (`types.IndexTemplateItem`, `types.IndexState`, `types.IngestPipeline`, `types.Alias`, `types.IndexSettings`, `types.TypeMapping`, `types.DataStream`, `types.DataStreamLifecycle`, etc.) where feasible.
- Keep thin conversion helpers only where Terraform-specific semantics (e.g., `map[string]any` for raw settings, alias action builder) are required.
- Preserve all existing behavior: `flat_settings` reads, date-math name encoding, atomic alias updates, 404 → nil/empty state handling, and diagnostic formatting.

**Non-Goals:**
- Changing any Terraform resource schema, attribute, or validation.
- Adding new resources or data sources.
- Removing `GetESClient()` or breaking untyped consumers in other files.
- Changing provider-level client construction, authentication, or connection logic.

## Decisions

**1. Keep `flat_settings=true` for read paths**
- **Rationale**: The existing resource mapping logic (especially in `elasticstack_elasticsearch_index`) expects flat keys such as `"index.number_of_shards"`. The typed API supports `FlatSettings(true)` on `Indices.Get` and `Cluster.GetComponentTemplate`. The generated `types.IndexSettings` struct has a `default` fallback in its custom `UnmarshalJSON` that stores unknown flat keys in `IndexSettings map[string]json.RawMessage`, so no data is lost. Relying on this avoids rewriting the resource-side settings flattening/unflattening logic.
- **Alternative considered**: Switching to nested settings and re-mapping inside each resource. Rejected because it would touch far more resource code and risk introducing behavioral drift.

**2. Use typed request builders for write paths**
- **Rationale**: Instead of `json.Marshal` on custom models, construct typed structs (e.g., `types.IndexTemplate`, `types.IngestPipeline`) and pass them directly to the typed client’s `.Do()` methods. This gives compile-time safety and removes the need for parallel model maintenance.
- **Alternative considered**: Keeping custom models and only swapping the transport layer. Rejected because it defeats the purpose of the typed API—we would still be maintaining duplicate shapes.

**3. Migrate `index.go` function-by-function in logical groups**
- **Rationale**: The file covers ILM, component templates, index templates, index CRUD, aliases, settings/mappings, data streams, data stream lifecycle, and ingest pipelines. Migrating group-by-group keeps diffs reviewable and makes bisection easier if a regression occurs.
- **Order**: ILM → component templates → index templates → index CRUD → aliases → settings/mappings → data streams → data stream lifecycle → ingest pipelines.

**4. Return typed API errors directly**
- **Rationale**: The typed API returns `*types.ElasticsearchError` for non-2xx responses. We can map `Status == 404` to the existing “not found → nil” semantics and wrap all other errors into diagnostics using the existing `diagutil` helpers.
- **Alternative considered**: Converting every error back to raw `*esapi.Response` to reuse legacy error paths. Rejected because it adds unnecessary wrapping; the typed error already carries status, header, and body information.

**5. Retain `AliasAction` struct for atomic alias updates**
- **Rationale**: The typed API does not expose a high-level alias-action builder. The existing `[]AliasAction` → `map[string]any` builder is small, well-understood, and Terraform-specific (e.g., handling `IsWriteIndex`, `Filter`). We will keep it but switch the execution from `esClient.Indices.UpdateAliases` to the typed equivalent.
- **Alternative considered**: Deleting `AliasAction` and inlining the builder into each resource. Rejected because two resources (`index` and `alias`) share the same atomic-update logic.

## Risks / Trade-offs

- **[Risk]** Typed API request structs may omit empty fields differently than our custom `json.Marshal` calls, causing Elasticsearch to apply defaults we previously avoided.
  - **Mitigation**: Run the full acceptance test suite for affected resources and inspect the JSON diff (via transport logging or unit tests) for each migrated write path before moving to the next group.
- **[Risk]** `flat_settings` fallback into `map[string]json.RawMessage` may miss keys that the resource mapping expects to be present as strongly-typed struct fields.
  - **Mitigation**: Validate with existing data source tests that read settings (`TestAccIndicesDataSource_ReadsIndexSettings_BroadCoverage`, `TestAccIndicesDataSource_ReadsSlowlogSettings`, etc.). Add a helper that preferentially reads from typed fields and falls back to the raw map.
- **[Risk]** Partial migration could leave a resource calling both typed and untyped helpers.
  - **Mitigation**: Migrate all functions in `index.go` in a single commit/PR so that every consumer in the listed resource packages switches together.

## Migration Plan

This is an internal code migration with no deployment or rollout steps. The change is merged as a normal PR. Rollback is a standard git revert. Acceptance tests on the affected resources serve as the regression gate.

## Open Questions

- None at design time.
