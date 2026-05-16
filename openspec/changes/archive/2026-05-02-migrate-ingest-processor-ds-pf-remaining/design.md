## Context

PR #1 (`migrate-ingest-processor-ds-pf-shared-base`) established the shared generic base and validated it with 4 representative processors. This PR scales the pattern to the remaining 35 processors and performs cleanup. The shared base artifacts are:
- `ProcessorModel` interface
- `processorDataSource[T ProcessorModel]` generic struct
- `CommonProcessorModel` + `CommonProcessorSchemaAttributes()`
- `marshalAndHash()` helper

## Goals / Non-Goals

**Goals:**
- Migrate all remaining 35 processor data sources to Plugin Framework using the PR #1 pattern
- Add missing common processor fields to `geoip` and `user_agent` (non-breaking enhancement)
- Clean up old SDK data source files and test files
- Move processor model structs from `internal/models/ingest.go` to `internal/elasticsearch/ingest/processor_models.go`

**Non-Goals:**
- Changing the shared base architecture (established in PR #1)
- Migrating the `ingest_pipeline` resource
- Changing any processor's JSON output format or attribute names (except geoip/user_agent additions)

## Decisions

### 1. Scale the Established Pattern Without Changes

**Decision:** The remaining 35 processors use the exact same pattern as the 4 representatives. Each processor defines:
1. A schema factory returning `schema.Schema`
2. A model struct implementing `ProcessorModel`
3. A `MarshalBody()` converting model to `map[string]any`
4. A `NewDataSource()` constructor returning `datasource.DataSource`

**Rationale:** PR #1 validated the pattern with edge cases (validators, JSON blobs, common-only). Scaling is purely mechanical.

### 2. Geoip and User-Agent: Add Common Fields

**Decision:** Include `description`, `if`, `ignore_failure`, `on_failure`, `tag` in schemas and models for `geoip` and `user_agent`, making them consistent with all other processors.

**Rationale:** All Elasticsearch processors support these fields per official documentation. The SDK omission is a gap, not intentional design. Adding optional fields is non-breaking.

### 3. Move Models Local, Then Delete from `internal/models`

**Decision:** Copy processor model structs from `internal/models/ingest.go` to `internal/elasticsearch/ingest/processor_models.go` as part of migration. Delete from `internal/models/ingest.go` once all SDK references are gone.

**Rationale:** These structs have no consumers outside the `ingest` package. Co-location eliminates an unnecessary dependency on the shared `models` package.

### 4. Preserve Existing Tests, Remove Old Files

**Decision:** Keep existing acceptance tests as-is (they already use `ProtoV6ProviderFactories`). Delete old SDK implementation and test files after migration.

**Rationale:** Acceptance tests validate behavior via JSON output assertions. Since JSON output is preserved, tests pass unchanged. The test harness is muxed — PF data sources are exercised transparently.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| One processor in the bulk migration has an unanticipated edge case (complex `TypeSet` handling, pointer semantics, etc.) | Acceptance tests for each processor catch regressions. The PR #1 representatives covered the primary edge cases |
| Deleting `internal/models/ingest.go` structs while still referenced by SDK files | Only delete after all old SDK files are removed. `go build` catches dangling references |
| `on_failure` JSON parsing differs between SDK and PF | Use `jsontypes.NormalizedType` which enforces JSON validity and semantic equality. Parse identically in `MarshalBody` |

## Migration Plan

1. Create 35 new PF processor data source files using the established pattern
2. Migrate `geoip` and `user_agent` with common fields included
3. Register all new constructors in `provider/plugin_framework.go`
4. Remove all old SDK registrations from `provider/provider.go`
5. Delete old SDK implementation files and test files
6. Move processor structs to `processor_models.go`; delete from `internal/models/ingest.go`
7. `make build` + acceptance tests

## Open Questions

- Should `commons_test.go` be updated or deleted? Assess after migration — if the ingest pipeline resource tests still need `CheckResourceJSON`, keep it.
