## Context

`internal/elasticsearch/watcher/watch/` contains the `elasticstack_elasticsearch_watch` resource. It currently embeds `*entitycore.ResourceBase` and implements `Create`, `Update`, `Read`, `Delete`, `Schema`, and `ImportState` manually. The resource's `read` helper already encapsulates the GET-watch logic and is used by both `Create` and `Update` for read-after-write.

The entitycore envelope owns Read/Delete/Schema/Configure/Metadata and supports callback-based Create/Update. Concrete resources that fit the callback contract can delete their Create/Update/Read/Delete method bodies and delegate to the envelope.

## Goals / Non-Goals

**Goals:**
- Adopt the envelope for watch, eliminating duplicated Read/Delete preludes and manual connection-block declaration.
- Preserve all externally observable behavior: schema shape, identity/import, JSON mapping, defaults, and actions redaction.
- Keep the actions redaction logic exactly as-is; it lives in `fromAPIModel` and is already isolated from the CRUD prelude.

**Non-Goals:**
- Changing the Terraform schema or adding new attributes.
- Changing acceptance test configurations.
- Altering the redaction algorithm or its test coverage.

## Decisions

### D1. Create and Update use envelope callbacks

**Choice:** Extract both method bodies into `createWatch` and `updateWatch` callbacks.

**Rationale:** Both methods follow the same flow: decode plan → get client → compute id → build PutWatch model → call PutWatch API → return model with id set. The envelope then calls `readFunc` for read-after-write and persists the result. This is a clean fit.

**Alternatives considered:**
- *Keep Create/Update on concrete type with placeholders.* Unnecessary; there is no config-derived post-processing.

### D2. The existing `read` helper becomes `readWatch` callback

**Choice:** Rename/refactor the existing `read` method into a package-level function with the envelope callback signature.

**Rationale:** The method already does exactly what the envelope read callback needs: parse composite ID, get client, call GetWatch, map to `Data` via `fromAPIModel` (which handles redaction), and return `(Data, bool, diag.Diagnostics)`.

**Alternatives considered:**
- *Inline the read logic into a new callback.* Would duplicate code. Rejected.

### D3. Actions redaction stays in `fromAPIModel`

**Choice:** No changes to the redaction logic or `mergeActionsPreservingRedactedLeaves`.

**Rationale:** The redaction is a state-mapping concern, not a CRUD-prelude concern. It is called inside `fromAPIModel`, which is invoked by the read callback. The envelope migration does not change this path.

### D4. Schema factory strips `elasticsearch_connection`

**Choice:** Remove the block from `schema.go`; the envelope injects it.

**Rationale:** Consistent with every other envelope resource.

## Risks / Trade-offs

- **Risk:** The envelope's `writeFromPlan` calls `readFunc` after create/update with the `writtenModel` (the model returned by the callback). The current `Create`/`Update` call `r.read(ctx, data)` where `data` has the **plan's** `Actions` value. The envelope's `readFunc` receives `writtenModel` which also carries the plan's `Actions`. `fromAPIModel` uses the prior actions parameter for redaction. Since `writtenModel.Actions` is the plan value, redation behaves identically. **Mitigation:** Confirm in code that `writtenModel` carries plan-shaped `Actions`; acceptance tests cover redacted action round-trips.
- **Risk:** ImportState passthrough must remain opt-in. **Mitigation:** Keep `ImportState` as a method on the concrete type; the envelope does not implement it.

## Migration Plan

1. Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to `Data`.
2. Refactor existing `read` into `readWatch(ctx, client, resourceID, state Data) (Data, bool, diag.Diagnostics)`. Return `(_, false, nil)` when watch is nil.
3. Extract `createWatch` from existing `Create` body. Remove the manual `r.read` call; just return the model with `ID` set.
4. Extract `updateWatch` from existing `Update` body. Same pattern.
5. Extract `deleteWatch` from existing `Delete` body.
6. Update `resource.go`: embed `*entitycore.ElasticsearchResource[Data]` with all four callbacks. Strip connection block from schema.
7. Keep `ImportState` on the concrete type.
8. Run build, lint, and acceptance tests.

**Rollback:** Restore `*entitycore.ResourceBase` embedding and the old method bodies.

## Open Questions

None.
