## Context

The `entitycore` package currently provides two data source patterns:

1. **Envelope generics** — `NewKibanaDataSource[T]` and `NewElasticsearchDataSource[T]` — which own `Configure`, `Metadata`, `Schema` (with connection block injection), and `Read` orchestration (config decode → scoped client resolution → optional version gate → callback invocation → state persistence). This pattern has been proven by three Agent Builder data sources (`agentbuilder_agent`, `agentbuilder_tool`, `agentbuilder_workflow`).

2. **Struct-based embedding** — `DataSourceBase` — which provides `Configure`, `Metadata`, and client factory access, but requires the concrete data source to implement `Schema` and `Read` manually. Nine data sources still use this or fully bespoke wiring.

Additionally, `agentbuilder_agent` currently satisfies the envelope contract through hand-rolled accessor methods (`GetKibanaConnection`, `GetVersionRequirements`) rather than embedding `KibanaConnectionField`, making it inconsistent with the other envelope data sources.

## Goals / Non-Goals

**Goals:**

- Migrate all remaining Plugin Framework data sources to the envelope pattern.
- Unify PF data source implementation so every one follows the same pattern.
- Remove `DataSourceBase` and all `entitycore_contract_test.go` files that assert its embedding, since it will have zero consumers.
- Update `agentbuilder_agent` to embed `entitycore.KibanaConnectionField` for consistency.
- Preserve every existing Terraform type name, schema attribute, and read behavior.
- Zero acceptance test breakage.

**Non-Goals:**

- Do not migrate SDK v2 data sources (ingest processors, cluster info, snapshot repository, etc.).
- Do not migrate resources to resource envelopes.
- Do not add `DataSourceVersionRequirement` support to any data source that doesn't already have static version gating (only `agentbuilder_agent` currently uses this).
- Do not change user-visible schemas or add/remove attributes.
- Do not introduce new envelope features (e.g., retry logic, caching, post-read hooks).

## Decisions

### 1. Use model embedding for all migrations

**Chosen:** All migrated data sources shall embed `entitycore.KibanaConnectionField` or `entitycore.ElasticsearchConnectionField` in their models.

**Rationale:** The envelope's doc.go already prescribes this as the canonical pattern. Embedding provides the `GetKibanaConnection()` / `GetElasticsearchConnection()` method automatically, eliminates hand-rolled accessors, and keeps models self-describing. For `agentbuilder_agent`, this replaces the current accessor-method approach with embedding.

**Alternative considered:** Keep existing explicit `KibanaConnection` / `ElasticsearchConnection` fields and add accessor methods. Rejected because it perpetuates inconsistency and the embedding pattern is already proven across `agentbuildertool` and `agentbuilderworkflow`.

### 2. Schema factory omits connection blocks

**Chosen:** Extract the current `Schema` method body into a package-level `func() dsschema.Schema` that returns entity attributes only, with no `kibana_connection` or `elasticsearch_connection` block.

**Rationale:** Matches the existing envelope contract. The envelope injects the connection block via `maps.Copy` in its `Schema` method. This avoids duplicate block definitions and ensures all connection blocks use the same schema definition from `providerschema`.

### 3. Read callback signature: `(ctx, client, model) (model, diag.Diagnostics)`

**Chosen:** Each data source's current `Read` method is stripped of envelope-owned orchestration and converted to a pure callback with this signature.

**Rationale:** The envelope owns config decode, client resolution, optional version enforcement, and state persistence. The callback owns only the entity-specific API call and model population. This is the exact contract already used by the Agent Builder POC.

### 4. Remove DataSourceBase as part of this change

**Chosen:** Once all `DataSourceBase` consumers are migrated, delete `internal/entitycore/data_source_base.go` and `data_source_base_test.go`.

**Rationale:** `DataSourceBase` exists solely for struct-based data sources. Once none remain, keeping it creates dead code and invites future inconsistency. If a future data source genuinely needs custom `Read` orchestration, it can still implement `datasource.DataSource` directly without `DataSourceBase`.

**Alternative considered:** Deprecate but keep `DataSourceBase`. Rejected because it has no callers and the envelope is strictly more powerful for the uniform case.

### 5. Fleet data sources use `NewKibanaDataSource` with `ComponentFleet`

**Chosen:** `fleet_integration`, `fleet_output`, and `enrollment_tokens` use `NewKibanaDataSource[T]` even though their Terraform type names use the `"fleet"` component.

**Rationale:** `Component` controls only the Terraform type name prefix (`elasticstack_fleet_*`). The envelope's `GetKibanaClient` resolves a Kibana scoped client, which is exactly what Fleet data sources need (they subsequently call `GetFleetClient()` on it). There is no `ComponentFleet` envelope variant because client resolution is driven by the connection type, not the component string.

### 6. Batch execution order

**Chosen:** Execute in three batches based on complexity and client backing:

1. **Batch 1 (Kibana-backed)**: `spaces`, `export_saved_objects` — simplest models, warm up on pattern.
2. **Batch 2 (Fleet-backed)**: `fleet_integration`, `fleet_output`, `enrollment_tokens` — moderate complexity, prove `ComponentFleet` + Kibana client combination.
3. **Batch 3 (ES-backed)**: `indices`, `index_template`, `enrich_policy`, `security_role_mapping` — heaviest schemas, includes one legacy not-found behavior (`index_template`).

**Rationale:** Risk increases with schema size and special-case behavior. Starting with smaller data sources validates the mechanical refactoring before tackling the complex ones. Each batch can be committed independently if CI passes.

## Risks / Trade-offs

| Risk | Mitigation |
|------|-----------|
| Schema block injection changes where `kibana_connection` / `elasticsearch_connection` is declared in source, but the final runtime schema should be identical. | Acceptance tests verify schema correctness. Unit tests in `entitycore` already assert block injection. Spot-check each migrated data source's schema response in a unit test if acceptance coverage is sparse. |
| `index_template` has legacy not-found behavior: when the template is missing, it sets an empty model with only `name` and `elasticsearch_connection`. In the envelope, the callback returns the model and the envelope calls `resp.State.Set`. Need to verify this preserves behavior. | Review the `index_template` callback carefully. The empty model still carries the connection block from config decode (embedded field), so `resp.State.Set` should produce the same state. Targeted acceptance test for missing template confirms this. |
| Large schema moves (`indices` has 50+ attributes, `index_template` has nested blocks with custom types) create noisy diffs that are hard to review. | Keep mechanical moves in dedicated commits separate from logic changes. Use `git diff --stat` to monitor line counts. |
| Removing `DataSourceBase` might break in-flight PRs that add new struct-based data sources. | This is a short window risk. Coordinate with team. Any future data source should use the envelope from the start. |
| `agentbuilder_agent` model change from accessors to embedding shifts the `KibanaConnection` field's position in the struct, which could affect struct tag-based reflection or test assertions. | `agentbuilder_agent` tests do not use reflection on the model for the connection field. Manual verification of unit and acceptance tests. |
| `enrich_policy` shares its `PolicyData` model between the data source and the resource. Embedding `ElasticsearchConnectionField` adds a field the resource doesn't need. | The resource uses `PolicyDataWithExecute`, which embeds `PolicyData`. The extra field is harmless (it's a Terraform framework type, not an API field). Alternatively, define a separate data-source-only model if the resource tests complain. |

## Migration Plan

1. **Batch 1 — Kibana-backed simple data sources**
   - Refactor `spaces` (embed `KibanaConnectionField`, schema factory, read callback, remove `entitycore_contract_test.go`).
   - Refactor `export_saved_objects` similarly.
   - Run targeted acceptance tests.

2. **Batch 2 — Fleet-backed data sources**
   - Refactor `fleet_integration`, `fleet_output`, `enrollment_tokens`.
   - Run targeted acceptance tests.

3. **Batch 3 — Elasticsearch-backed data sources**
   - Refactor `security_role_mapping`, `enrich_policy`, `indices`, `index_template`.
   - Pay special attention to `index_template` not-found handling.
   - Run targeted acceptance tests.

4. **Agent Builder agent model cleanup**
   - Update `agentbuilderagent/models.go` to embed `KibanaConnectionField`.
   - Remove `GetKibanaConnection` and `GetVersionRequirements` methods (the envelope's type constraint is now satisfied by embedding).
   - Run existing acceptance tests.

5. **Remove dead code**
   - Delete `internal/entitycore/data_source_base.go`.
   - Delete `internal/entitycore/data_source_base_test.go`.

6. **Final verification**
   - `go build` passes.
   - All acceptance tests for migrated data sources pass.
   - `make check-openspec` passes (no spec changes needed, but structural checks should succeed).

## Open Questions

- None. The pattern is proven and mechanical.
