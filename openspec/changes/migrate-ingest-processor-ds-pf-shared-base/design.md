## Context

The `internal/elasticsearch/ingest/` package contains 39 processor data sources. Every one follows the same boilerplate pattern: schema definition, `ReadContext` that populates a `models.ProcessorX`, marshals `{"<name>": processor}` JSON, hashes it for `id`, and sets the `json` attribute. There is no external API call. The provider already runs as a muxed provider (SDK + PF), so PF data sources can be registered alongside SDK resources.

This PR establishes the shared generic base and validates it with 4 processors chosen to exercise different aspects of the base:
- `drop` — minimal surface (common fields only)
- `append` — specific scalar + list fields with defaults
- `script` — PF validators (`ExactlyOneOf`) and a JSON blob attribute (`params`)
- `foreach` — a JSON string field (`processor`) parsed to `map[string]any`

## Goals / Non-Goals

**Goals:**
- Create a generic `processorDataSource[T ProcessorModel]` that owns `Read` entirely
- Extract `CommonProcessorModel` and `CommonProcessorSchemaAttributes()` for the 37 processors that use common fields
- Preserve byte-identical JSON output and hash-based ID for the 4 representatives
- Prove the pattern works in CI before scaling

**Non-Goals:**
- Migrating the remaining 35 processors
- Changing `geoip` or `user_agent` schemas (those come in PR #2)
- Moving `internal/models/ingest.go` structs yet (PR #2)
- Deleting old SDK files yet (PR #2)

## Decisions

### 1. Generic `processorDataSource[T ProcessorModel]`

**Decision:** A Go generic struct parameterized by a `ProcessorModel` interface owns `Read`. The generic `Read` decodes config, calls `model.MarshalBody()`, wraps as `{"name": body}`, marshals with `json.MarshalIndent(..., "", " ")`, hashes via `schemautil.StringToHash`, and sets state.

**Rationale:** Eliminates the identical `Read` function repeated 39 times. The interface is minimal and does not assume common fields exist.

### 2. `CommonProcessorModel` Embedded Struct

**Decision:** A `CommonProcessorModel` struct with `tfsdk` tags embeds into the 37 processors that use common fields. `CommonProcessorSchemaAttributes()` returns the common `schema.Attribute` map.

**Rationale:** Schema and model composition in PF is straightforward. Embedding reduces duplication without forcing processors that don't need common fields (none in this PR, but the design must accommodate it).

### 3. Local Inner Structs with `json` Tags

**Decision:** Lightweight local structs (mirroring the existing `models.ProcessorX` shapes) are constructed inside `MarshalBody`, marshaled to bytes, then `json.Unmarshal` to `map[string]any`.

**Rationale:** Reuses existing `omitempty` semantics. Clean compile-time field names. The struct is an implementation detail of `MarshalBody`.

### 4. `jsontypes.NormalizedType` for JSON Strings

**Decision:** `on_failure` and any JSON blob fields use `jsontypes.NormalizedType{}` as the element/attribute type.

**Rationale:** Gives JSON validation and semantic diff suppression in one type. Already used elsewhere in the provider.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| JSON output differs from SDK, causing `id` hash churn | Use identical `json.MarshalIndent` params and `json` tags; acceptance tests validate exact JSON match |
| `on_failure` parsing from `jsontypes.Normalized` fails | Parse with `json.Unmarshal` in `MarshalBody`; errors become `diag.Diagnostics` |
| Old SDK data sources left registered | Delete from `DataSourcesMap` in same PR that adds to `plugin_framework.go`. `make build` catches duplicates |

## Migration Plan

1. Create shared base files (`processor_datasource_base.go`, `processor_common.go`, `processor_models.go`)
2. Create 4 PF processor files
3. Wire in `provider/plugin_framework.go`; remove from `provider/provider.go`
4. `make build` + acceptance tests for the 4 processors

## Open Questions

- Should `ListAttribute` or `SetAttribute` be used for historically-set fields? The shared base is agnostic — each processor schema factory decides. `SetAttribute` + explicit sort in `MarshalBody` for fields that were `TypeSet` in SDK.
