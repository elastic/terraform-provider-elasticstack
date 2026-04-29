## Context

The provider's `DataSourceBase` (in `internal/entitycore`) currently provides `Configure` and `Metadata` wiring, but every data source still rewrites its entire `Read()` method. Agent Builder data sources (`agentbuilderagent`, `agentbuildertool`, `agentbuilderworkflow`) repeat a full pipeline: config decode, scoped Kibana client resolution, version enforcement, composite ID resolution, space ID fallback, API fetch, and state set. This is the duplication identified in #2511.

Existing struct-based data sources (`spaces`, `enrollment_tokens`, `index_template`) all follow the same shape but vary in client type (Kibana vs Elasticsearch). The provider already has typed scoped clients (`KibanaScopedClient`, `ElasticsearchScopedClient`) and factory methods (`GetKibanaClient`, `GetElasticsearchClient`). The gap is at the orchestration layer.

## Goals / Non-Goals

**Goals:**
- Eliminate Read() orchestration boilerplate for simple Kibana and Elasticsearch data sources.
- Allow concrete packages to provide only schema + entity logic (no connection field, no Read method).
- Keep the change fully opt-in; existing struct-based data sources continue to work.
- Migrate the three Agent Builder data sources from #2511 as reference implementations.

**Non-Goals:**
- Changing resource patterns (Create/Read/Update/Delete). Resources remain struct-based.
- Migrating SDK-based data sources.
- Changing any Terraform schema or observable behavior.
- Complex data sources with conditional early returns or custom state manipulation (e.g., `agentbuilder_agent` with its tool-dependency graph) are not forced into the envelope; they stay struct-based or use local helpers.

## Decisions

### 1. Envelope with interface-constrained models and embeddable connection helpers

**Decision:** The generic constructor uses a type constraint `T KibanaDataSourceModel` (and `T ElasticsearchDataSourceModel`) where the model provides a `GetKibanaConnection() types.List` method. Two embeddable structs—`KibanaConnectionField` and `ElasticsearchConnectionField`—supply the connection block field and its getter so concrete models need only embed the helper.

**Rationale:** Go does not allow embedding a **type parameter** itself in a struct (golang/go#49030):

```go
type envelope[T any] struct {
    T   // ❌ invalid: embedded field type cannot be a type parameter
    KibanaConnection types.List
}
```

Because this is a hard language limitation, the framework's support for promoted embedded-struct fields cannot be used with generic envelopes. Instead, the concrete model `T` must itself contain the connection block field. An interface constraint keeps the constructor's contract explicit, and embeddable helper structs (`KibanaConnectionField`, `ElasticsearchConnectionField`) minimize boilerplate—concrete models embed the helper rather than declaring the field and method manually.

**Alternative considered:** Anonymous field embed `T` in a non-generic struct. Rejected because Go forbids embedding type parameters.

**Alternative considered:** Named field `Body T` in the envelope. Rejected because it would require the schema to namespace the concrete attributes under `body`, which is a breaking schema change.

**Alternative considered:** Reflection-based field injection or two-pass decode. Rejected because it is fragile, harder to test, and couples the envelope to Terraform framework internals.

### 2. Schema injection at construction time

**Decision:** The generic constructor accepts a schema factory `func() datasource.Schema`, calls it internally to obtain a fresh schema value, defensively clones the `Blocks` map, and injects `kibana_connection` (or `elasticsearch_connection`) before returning.

**Rationale:** The concrete package defines its schema without connection blocks and passes a factory function. The base obtains an isolated schema copy, clones the `Blocks` map to avoid mutating the caller's shared state, and adds the connection block. A function parameter eliminates the shared-mutation footgun because the constructor controls when the schema is built and owns the resulting instance. Defensive cloning of the `Blocks` map inside the constructor provides defense-in-depth regardless of what the schema function does.

**Alternative considered:** Pass `datasource.Schema` as a value parameter. Rejected because `Schema.Blocks` is a `map[string]datasource.Block` (reference type). Mutating the map in the constructor would surprise callers who reuse the same schema value across multiple data sources or hold it in package-level variables.

**Alternative considered:** Require the concrete package to include the connection block in its schema manually. Rejected because it defeats the purpose—the concrete package should not need to think about connections at all.

### 3. Two constructors: Kibana and Elasticsearch

**Decision:** Provide `NewKibanaDataSource[T]()` and `NewElasticsearchDataSource[T]()` as separate top-level functions.

**Rationale:** The connection block type, scoped client type, and factory method differ. A single generic constructor would need an extra type parameter for the client type and a way to dispatch the factory call, which adds complexity without reducing code. Two constructors are explicit and zero-cost.

**Alternative considered:** Single `NewDataSource[T, C any](...)` with a constraint that maps `C` to the right factory. Rejected as over-abstracted for two cases.

### 4. Pure read function signature

**Decision:** `readFunc func(context.Context, *clients.KibanaScopedClient, T) (T, diag.Diagnostics)`

**Rationale:** The function receives the scoped client and the decoded model, performs entity work, and returns the populated model. Diagnostics accumulate naturally. State setting is owned by the base.

**Alternative considered:** Callback with access to `*datasource.ReadResponse` for direct diagnostic append. Rejected because it leaks the Terraform contract into the entity logic, re-creating the coupling we're trying to eliminate.

### 5. No version enforcement in the generic base

**Decision:** `EnforceMinVersion` calls stay in the concrete read function, not the generic base.

**Rationale:** Not all Kibana data sources need version gating. Adding it to the base would require a way to opt out or pass a version parameter, which complicates the simple cases. The Agent Builder datasources can share a domain-local version helper.

## Risks / Trade-offs

- **[Risk]** Go may never lift the restriction on embedding type parameters in structs, which would permanently block the originally-envisioned anonymous-field pattern. → **Mitigation:** The shipped design explicitly chooses an interface-constraint + embeddable helper approach that works within the current language. The helper structs (`KibanaConnectionField`, `ElasticsearchConnectionField`) are a stable contract even if Go generics evolve.
- **[Risk]** Schema injection could mutate shared map state if the concrete package holds a reference to the `Blocks` map. → **Mitigation:** The constructor accepts a factory function (callers pass `getDataSourceSchema`, not a `schema.Schema` value) and defensively clones the `Blocks` map before injecting the connection block. The resulting schema instance is fully owned by the generic base.
- **[Risk]** Data sources with unusual `Read` patterns (no state set, conditional early return, manual state removal) can't use the envelope. → **Mitigation:** The envelope is opt-in. Complex data sources stay struct-based. We'll document the envelope's limitations.
- **[Risk]** Future framework versions may change `Config.Get()` reflection behavior. → **Mitigation:** The envelope code is centralized; a fix is one place. Existing struct-based data sources are unaffected.

## Migration Plan

1. Add `NewKibanaDataSource[T]()` and `NewElasticsearchDataSource[T]()` to `internal/entitycore`.
2. Add connection block helpers for datasource schema types.
3. Migrate `agentbuilderworkflow` data source (simplest of the three—no conditional workflow fetch).
4. Migrate `agentbuildertool` data source (medium—conditional workflow fetch).
5. Migrate `agentbuilderagent` data source if it fits; otherwise leave it struct-based with a local helper.
6. Acceptance tests for migrated data sources must pass unchanged.
7. Document the pattern for future data sources.

## Open Questions

- Should we also migrate `kibana/spaces` and `fleet/enrollmenttokens` (simple cases) as part of this change, or keep the scope to Agent Builder only?
