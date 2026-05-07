## Context

`internal/elasticsearch/index/templateilmattachment/` contains the `elasticstack_elasticsearch_index_template_ilm_attachment` resource. It currently embeds `*entitycore.ResourceBase` and implements `Create`, `Update`, `Read`, `Delete`, `Schema`, and `ImportState` manually. The resource targets a `@custom` component template derived from the configured `index_template` name, reads its existing content, merges or removes the `index.lifecycle.name` setting, and writes the template back via the Put Component Template API.

The entitycore envelope owns Read/Delete/Schema/Configure/Metadata and supports callback-based Create/Update. Concrete resources may override Create/Update when their lifecycle diverges from the simple callback contract.

## Goals / Non-Goals

**Goals:**
- Adopt the envelope for the ILM attachment resource, eliminating duplicated Read/Delete preludes and manual connection-block declaration.
- Preserve all externally observable behavior: schema shape, identity/import, version gating, content preservation, and delete semantics.
- Keep Create/Update on the concrete type because they perform non-trivial config-derived logic.

**Non-Goals:**
- Changing the Terraform schema or adding new attributes.
- Changing acceptance test configurations.
- Refactoring the ILM merge/remove helpers (`mergeILMSetting`, `removeILMSetting`); they are already well-factored.

## Decisions

### D1. Create and Update stay on concrete type with placeholder callbacks

**Choice:** The `Resource` struct overrides `Create` and `Update`. The envelope constructor receives `PlaceholderElasticsearchWriteCallbacks`.

**Rationale:** Create and Update both:
- Call `ServerVersion` to enforce the `>= 8.2.0` version gate.
- Read the existing `@custom` component template to preserve content.
- Warn when the existing template has a `version` field or a pre-existing ILM setting.
- Merge the new ILM setting into the template's settings map.
- Call Put Component Template.
- Read back to verify.

This flow reads from the plan, the remote API, and prior state (for warnings), which does not fit the simple callback contract.

**Alternatives considered:**
- *Extract into callbacks by passing extra context via closures.* Would make the callback harder to reason about and test. Rejected.

### D2. Read and Delete use envelope callbacks

**Choice:** Extract `Read` and `Delete` into `readILMAttachment` and `deleteILMAttachment` callbacks.

**Rationale:** Both follow the standard pattern: decode state, parse composite ID, get client, call API, handle 404. The read callback also derives `index_template` from the component template name during import. This is straightforward.

### D3. Read callback derives index_template for import

**Choice:** The read callback preserves the existing import derivation logic: when `IndexTemplate` is unknown, strip the `@custom` suffix from the component template name.

**Rationale:** Import is passthrough on `id`; the subsequent read must derive `index_template` so Terraform can decode the state. The existing logic in `Read` handles this and should be preserved inside the callback.

### D4. Schema factory strips `elasticsearch_connection`

**Choice:** Remove the block from `getSchema()`; the envelope injects it.

**Rationale:** Consistent with every other envelope resource.

## Risks / Trade-offs

- **Risk:** The delete callback removes the ILM setting via Put Component Template, not Delete Component Template. The envelope's generic delete prelude might look unusual when the callback does a PUT. **Mitigation:** The callback is explicit and well-documented. Acceptance tests verify delete behavior.
- **Risk:** The read callback returns `found == false` when the component template exists but has no ILM setting, causing the envelope to remove the resource from state. This is the existing behavior and is preserved. **Mitigation:** No change; acceptance tests cover this path.

## Migration Plan

1. Add `GetID()`, `GetResourceID()`, `GetElasticsearchConnection()` to `tfModel`.
2. Refactor existing `Read` into `readILMAttachment` callback with envelope signature. Preserve `index_template` derivation for import.
3. Refactor existing `Delete` into `deleteILMAttachment` callback.
4. Update `resource.go`: embed `*entitycore.ElasticsearchResource[tfModel]`, pass schema factory, read/delete callbacks, and placeholder create/update callbacks. Keep `ImportState`.
5. Strip connection block from schema.
6. Run build, lint, and acceptance tests.

**Rollback:** Restore `*entitycore.ResourceBase` embedding and the old method bodies.

## Open Questions

None.
