## Context

The entitycore resource envelope (`internal/entitycore/resource_envelope.go`) provides a generic `NewElasticsearchResource[T]` constructor that owns Schema (with connection-block injection), Read, Delete, Create, and Update. Concrete resources embed the returned `*ElasticsearchResource[T]` and may override any method or implement additional interfaces (ImportState, UpgradeState, ModifyPlan, ValidateConfig).

Both `index_template` and `index_lifecycle` currently embed `*entitycore.ResourceBase` directly and implement `Read`, `Delete`, `Create`, `Update`, `Schema`, `ImportState`, and `UpgradeState` manually. They also both declare the `elasticsearch_connection` block in their schema factories.

The security resources (`user`, `systemuser`, `role`, `rolemapping`) already migrated to the envelope under the `elasticsearch-resource-envelope` change, establishing the pattern: envelope owns Read/Delete/Schema/Configure/Metadata; concrete type keeps anything that diverges (Create/Update/ImportState/UpgradeState).

## Goals / Non-Goals

**Goals:**
- Adopt the envelope for both resources, eliminating duplicated Read/Delete preludes and manual connection-block declarations.
- Preserve all externally observable behavior: schema shape, identity/import, state upgrade, plan modification, validation, and CRUD semantics.
- Keep custom Create/Update logic where it diverges from a simple callback (index_template).
- Migrate index_lifecycle's full CRUD into envelope callbacks because it is a clean fit.

**Non-Goals:**
- Changing the Terraform schema or adding new attributes.
- Changing acceptance test configurations.
- Wrapping ModifyPlan or ValidateConfig in the envelope; those remain on the concrete type.
- Removing the existing `ResourceBase` simple base; non-envelope resources still need it.

## Decisions

### D1. index_template keeps Create/Update on concrete type with placeholder callbacks

**Choice:** The `Resource` struct overrides `Create` and `Update` (retaining the existing methods). The envelope constructor receives `PlaceholderElasticsearchWriteCallbacks`.

**Rationale:** `Create` and `Update` in index_template perform config-derived logic that the envelope callback signature cannot express:
- They read `req.Config` for alias reconciliation (plan unknowns vs config-shaped defaults).
- They call `ServerVersion` for feature gating (`ignore_missing_component_templates`, `data_stream_options`).
- Update applies the 8.x `allow_custom_routing` workaround using prior state and config.

Keeping these methods on the concrete type avoids forcing that complexity into the callback contract.

**Alternatives considered:**
- *Extract everything into callbacks and pass plan/config/state via closure or custom struct.* Would require changing the envelope signature or adding a side channel. Over-engineered for one resource.

### D2. index_lifecycle uses real envelope callbacks for Create/Update

**Choice:** Extract `Create` and `Update` into `createILM` and `updateILM` callbacks passed to the envelope.

**Rationale:** The flow is plan decode → version check → expand to `models.Policy` → PUT → compute id → read-back → set state. This fits the envelope's `writeFromPlan` exactly. No config-derived post-processing is needed after the read-back.

**Alternatives considered:**
- *Keep Create/Update on concrete type for symmetry with index_template.* Unnecessary; ilm has no special post-PUT reconciliation.

### D3. Read callbacks perform post-processing before returning

**Choice:** Both `readIndexTemplate` and `readILM` accept the prior-state model, fetch API data, perform any necessary reconciliation/normalization, copy `ID` and `ElasticsearchConnection` from the incoming model, and return the fully prepared model.

**Rationale:** The envelope's `Read` calls `resp.State.Set(ctx, &resultModel)` directly after the callback. There is no hooks for post-processing. For index_template, alias reconciliation and canonicalization must happen inside the read callback. The incoming `state T` parameter carries the prior state values needed for reconciliation.

**Alternatives considered:**
- *Override `Read` on the concrete type.* Would duplicate the envelope's Read prelude (state decode, composite ID parse, client resolution). Rejected because it defeats the purpose of the migration.

### D4. Schema factories strip `elasticsearch_connection`

**Choice:** Both schema factories remove the `elasticsearch_connection` block declaration; the envelope injects it.

**Rationale:** Matches every other envelope resource. Ensures consistency with the provider-wide helper.

### D5. Preserve ImportState as opt-in on concrete types

**Choice:** Both resources keep their own `ImportState` methods. The envelope does not implement `ResourceWithImportState`.

**Rationale:** Preserves the provider-wide convention that import is opt-in. Both resources currently support import and shall continue to do so.

## Risks / Trade-offs

- **Risk:** index_template's alias reconciliation inside the read callback may behave subtly differently than when it ran in the method body. **Mitigation:** The callback receives the exact same prior-state model; the reconciliation functions (`applyTemplateAliasReconciliationFromReference`, `canonicalizeTemplateAliasSetInModel`) are called with the same arguments. Acceptance tests cover alias routing drift.
- **Risk:** index_lifecycle's state upgrader interacts with the envelope's state handling. **Mitigation:** UpgradeState runs before Read; the envelope's Read sees already-upgraded state. No interaction risk.
- **Risk:** Concrete methods that override envelope methods could silently break if the envelope signature changes. **Mitigation:** Go compile-time checks (interface assertions) catch signature mismatches.
- **Trade-off:** index_template keeps duplicated Create/Update method declarations. Acceptable because the logic does not fit the callback contract.

## Migration Plan

1. **index_template:**
   - Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to `Model`.
   - Refactor `readIndexTemplate` to accept `Model` and return `Model` with reconciliation applied.
   - Extract `deleteIndexTemplate` callback.
   - Update `resource.go`: embed `*entitycore.ElasticsearchResource[Model]`, pass schema factory, read/delete callbacks, and placeholder create/update callbacks. Keep `UpgradeState`, `ModifyPlan`, `ValidateConfig`, `ImportState` on the concrete type.
   - Strip `elasticsearch_connection` from `resourceSchema()`.

2. **index_lifecycle:**
   - Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to `tfModel`.
   - Refactor existing `Read` into `readILM` callback that accepts prior state and returns populated model.
   - Extract `deleteILM`, `createILM`, `updateILM` callbacks from existing method bodies.
   - Update `resource.go`: embed `*entitycore.ElasticsearchResource[tfModel]` with all four callbacks. Keep `UpgradeState` and `ImportState` on the concrete type.
   - Strip `elasticsearch_connection` from the schema factory.

3. Run `make build`, `make check-lint`, and acceptance tests for both resources.

**Rollback:** Revert `resource.go` changes to restore `*entitycore.ResourceBase` embedding; the old method bodies remain as package-level functions and can be re-wired.

## Open Questions

None.
