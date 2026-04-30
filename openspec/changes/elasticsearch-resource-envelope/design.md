## Context

`internal/entitycore` provides shared Plugin Framework wiring for resources and data sources. Today it has two layers:

- **Simple base** (`resource_base.go`, `data_source_base.go`): shares Configure / Metadata / Client only. Concrete entities own Schema, Read, Delete, etc.
- **Data source envelope** (`data_source_envelope.go`): a generic constructor `NewElasticsearchDataSource[T]` / `NewKibanaDataSource[T]` that adds connection-block schema injection, config decode, scoped-client resolution, and state persistence on top of the simple base. Concrete data sources only supply a schema factory and a pure read function.

There is no resource analogue of the envelope. Issue #2555 found that the four Elasticsearch security resources duplicate the same Read/Delete prelude:

```go
// Repeated across user, systemuser, role, rolemapping (Read and Delete each)
var data Data
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
if resp.Diagnostics.HasError() { return }

compID, diags := clients.CompositeIDFromStrFw(data.ID.ValueString())
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() { return }

client, diags := r.Client().GetElasticsearchClient(ctx, data.ElasticsearchConnection)
resp.Diagnostics.Append(diags...)
if resp.Diagnostics.HasError() { return }
```

Create and Update flows in these resources diverge significantly (write-only passwords, server-version gating, JSON marshalling, post-write re-reads), so wrapping them is out of scope for this change.

## Goals / Non-Goals

**Goals:**

- Provide `entitycore.NewElasticsearchResource[T]` that owns Schema (with connection-block injection), Configure, Metadata, Read prelude, and Delete prelude.
- Mirror the data source envelope's shape: type-parameter constraint, schema factory, pure callbacks.
- Migrate the four Elasticsearch security resources to use it. No external behavior change.
- Preserve existing acceptance tests verbatim.
- Preserve the provider-wide convention that import support is opt-in: the envelope does NOT implement ImportState; resources that need it implement it themselves.

**Non-Goals:**

- Wrapping Create or Update.
- Adding a Kibana sibling (`NewKibanaResource[T]`) — deferred until we have at least one Kibana resource demanding it.
- Migrating non-security resources (e.g., `api_key`, ingest processors, Fleet resources). Those can adopt the envelope incrementally under the same spec without re-proposal.
- Changing what `*entitycore.ResourceBase` does. The simple base remains as-is for entities that don't fit the envelope.

## Decisions

### D1. Envelope owns Read + Delete + Schema; concrete owns Create + Update + ImportState

**Choice:** Wrap Read, Delete, Schema (with connection-block injection), Configure, and Metadata. Concrete resources keep Create, Update, and ImportState.

**Rationale:** Read and Delete have a uniform prelude across the four resources; Create and Update don't. ImportState is a one-liner (`resource.ImportStatePassthroughID`) that almost every Plugin Framework resource implements identically, but making it the envelope's default would force import support on every envelope consumer. The provider convention is opt-in: resources declare `ResourceWithImportState` explicitly. Removing ImportState from the envelope preserves that convention and lets future resources use the envelope regardless of whether they support import. Schema injection is included because connection-block declarations are also duplicated and the data source envelope already does this — symmetry keeps the entitycore API coherent.

**Alternatives considered:**

- *Read + Delete only, leave Schema alone.* Smaller blast radius but leaves the connection-block duplication in place and breaks symmetry with the data source envelope.
- *Full CRUD wrapper.* Would force Create/Update into a single shape. The four resources each have meaningfully different update flows (write-only passwords, version gating, post-write re-read). Premature.

### D2. Model interface requires `GetID()` and `GetElasticsearchConnection()`

**Choice:** Define `ElasticsearchResourceModel` as

```go
type ElasticsearchResourceModel interface {
    GetID() types.String
    GetElasticsearchConnection() types.List
}
```

Each concrete `Data` struct adds value-receiver methods:

```go
func (d Data) GetID() types.String { return d.ID }
func (d Data) GetElasticsearchConnection() types.List { return d.ElasticsearchConnection }
```

**Rationale:** The envelope needs both fields and access through an interface lets the type parameter `T` stay concrete. Mirrors `ElasticsearchDataSourceModel`'s `GetElasticsearchConnection()` convention. Embedded helper structs (analogous to `ElasticsearchConnectionField`) aren't needed because the four `Data` types already declare both fields explicitly.

**Alternatives considered:**

- *Read attributes via `state.GetAttribute(path.Root("id"), …)` instead of an interface method.* Avoids changing every Data struct but bypasses the type system and has worse compile-time guarantees.
- *Provide an embeddable `ElasticsearchResourceFields` helper.* Premature; only revisit if more migrations adopt it.

### D3. Read callback returns `(T, bool, diag.Diagnostics)`

**Choice:**

```go
type elasticsearchReadFunc[T ElasticsearchResourceModel] func(
    ctx context.Context,
    client *clients.ElasticsearchScopedClient,
    resourceID string,
    state T,
) (T, bool, diag.Diagnostics)
```

The bool signals whether the entity exists. When `false` the envelope calls `resp.State.RemoveResource(ctx)`; when `true` it calls `resp.State.Set(ctx, model)`. Errors short-circuit before either branch.

**Rationale:** The four resources all use the same pattern: API call → if nil/missing, log and remove from state; otherwise populate and persist. systemuser additionally checks `!user.IsSystemUser()`, which fits inside the `found` decision cleanly.

**Alternatives considered:**

- *Return `(*T, diag.Diagnostics)` with nil signalling not-found.* Awkward with non-pointer Data structs and value-typed `T`.
- *A separate `notFoundError` sentinel diagnostic.* Indirect and easy to misuse.

### D4. resourceID is a parsed string, not the full `*CompositeID`

**Choice:** The envelope parses `clients.CompositeIDFromStrFw(model.GetID().ValueString())` itself and passes only `compID.ResourceID` to the callbacks.

**Rationale:** Every current handler uses only `compID.ResourceID`. Pre-parsing keeps the callback surface narrow. If a future resource needs the full composite ID, broaden the signature later.

### D5. Delete callback is required; system_user supplies a no-op

**Choice:** `deleteFunc` is a non-nilable required parameter to the constructor. `systemuser.deleteSystemUser` is

```go
func deleteSystemUser(ctx context.Context, _ *clients.ElasticsearchScopedClient, resourceID string, _ Data) diag.Diagnostics {
    tflog.Warn(ctx, fmt.Sprintf(`System user '%s' is not deletable, just removing from state`, resourceID))
    return nil
}
```

**Rationale:** A nilable callback hides intent at the envelope layer. Making it required forces each resource to express the delete behavior explicitly. The "no API call" case is a one-line function — clearer than a magic nil.

### D6. ImportState is NOT in the envelope; concrete resources opt in

**Choice:** The envelope does NOT implement `ImportState`. Concrete resources that need import support implement it themselves with the standard one-liner:
```go
func (r *resourceType) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Rationale:** In Go, embedding a struct that implements `ResourceWithImportState` makes the outer type permanently implement that interface. There's no way to un-implement an interface. If the envelope provided a default ImportState, any resource using it would be forced to support import — even resources like `api_key` that intentionally do not. Keeping ImportState out of the envelope preserves the provider's opt-in convention and matches every other Plugin Framework resource in the codebase.

### D7. Connection-block injection mirrors the data source envelope

**Choice:** The schema factory returns `rschema.Schema` without an `elasticsearch_connection` block. The envelope's `Schema` method copies the user-provided blocks map and adds `elasticsearch_connection` via `providerschema.GetEsFWConnectionBlock()`.

**Rationale:** Identical to the data source envelope's pattern in `data_source_envelope.go:191-198`. Concrete resources stop carrying their own copy of the connection block.

### D8. Envelope reuses `ResourceBase` rather than re-implementing Configure/Metadata

**Choice:** `ElasticsearchResource[T]` embeds `*ResourceBase` and uses its existing `Configure`, `Metadata`, and `Client()` methods. The envelope adds Schema, Read, Delete, and ImportState only.

**Rationale:** Keeps the simple-base requirements (in `provider-framework-entity-core` spec) the single source of truth for Configure/Metadata behavior. The data source envelope today re-implements these — that's tolerable but slightly duplicative; resources have a chance to do better. Embedding also means the struct remains a `ResourceWithConfigure` automatically.

## Risks / Trade-offs

- **Risk:** Method promotion ambiguity — a concrete resource that defines its own `Read` would silently override the envelope's. **Mitigation:** Each concrete resource is `type fooResource struct { *entitycore.ElasticsearchResource[Data] }` and declares only Create / Update. Reviewers (and the assertion `var _ resource.Resource = …`) catch any accidental Read/Delete overrides.

- **Risk:** Generic type parameter `T` makes call sites verbose. **Mitigation:** Consistent naming (`type userResource struct { *entitycore.ElasticsearchResource[Data] }`) keeps the verbosity local to the constructor file. The data source envelope already proved this readable.

- **Risk:** systemuser's no-op delete looks like dead code. **Mitigation:** Leave a one-line comment in `deleteSystemUser` explaining why it's a no-op (system users aren't deletable).

- **Risk:** No envelope default for ImportState means slightly more boilerplate for resources that do support import. **Mitigation:** The boilerplate is a single standardized line; every other PF resource in the provider does exactly this. The consistency gain outweighs the three extra lines per resource.

- **Trade-off:** Update flows aren't wrapped. Each resource still decodes plan, gets the client, reads state where needed, and writes state. Acceptable for now — wrapping later is a refinement, not a rewrite.

## Migration Plan

1. Add `internal/entitycore/resource_envelope.go` with the generic constructor and supporting types. Add tests mirroring `data_source_envelope_test.go`.
2. For each of the four resources (`user`, `systemuser`, `role`, `rolemapping`), in independent commits:
   - Add `GetID()` and `GetElasticsearchConnection()` to `Data`.
   - Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in the resource struct.
   - Move `read` body into a package-level `readXxx` function with the new callback signature; delete the `Read` method.
   - Move `delete` body into a package-level `deleteXxx` function; delete the `Delete` method.
   - Strip the `elasticsearch_connection` block from the schema factory; the envelope injects it.
   - Add `ImportState` passthrough on `id` to the concrete resource type (opt-in, same as every other PF resource in the provider).
3. Update `internal/entitycore/doc.go` to document the resource envelope alongside the data source envelope.
4. Run `make check-lint`, `make build`, `make check-openspec`, and the security acceptance tests.

**Rollback:** Each resource migration is an independent commit; reverting any single one leaves the others functional. The envelope itself can be removed if no callers remain.

## Open Questions

None. All decisions are settled per the explore session.
