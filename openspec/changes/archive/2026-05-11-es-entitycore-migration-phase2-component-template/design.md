## Context

`internal/elasticsearch/index/component_template.go` is the only remaining Plugin SDK v2 resource in the Elasticsearch index domain. It defines `ResourceComponentTemplate()` returning a `*schema.Resource` with custom Create/Update/Read/Delete functions. The resource shares some template-related helpers with the legacy SDK `index_template` via `internal/elasticsearch/index/template_sdk_shared.go` (e.g. alias set hashing, template expansion/flattening).

All other Elasticsearch resources have migrated to Plugin Framework and many already use the `entitycore.ElasticsearchResource[T]` envelope, which owns Schema (with `elasticsearch_connection` injection), Read prelude, Delete prelude, and callback-based Create/Update.

## Goals / Non-Goals

**Goals:**
- Re-implement `elasticstack_elasticsearch_component_template` on Plugin Framework using the entitycore envelope.
- Preserve every externally observable behavior: schema shape, attribute defaults, JSON validation, alias routing preservation, identity format, import passthrough, and CRUD API calls.
- Eliminate the last SDK resource in the index package so the provider can eventually drop SDK dependency from this domain.
- Keep shared helpers in `internal/elasticsearch/index/` if they are still consumed by SDK code; move resource-specific helpers into the new package.

**Non-Goals:**
- Changing the Terraform schema or adding new attributes.
- Changing acceptance test configurations or assertions.
- Removing `template_sdk_shared.go` while SDK resources elsewhere still need it.
- Adding concrete Create or Update overrides; this resource's PUT-on-write flow fits the envelope callback contract.

## Decisions

### D1. New sub-package `internal/elasticsearch/index/componenttemplate`

**Choice:** Create a dedicated package rather than adding files next to the legacy resource.

**Rationale:** The current file `component_template.go` mixes SDK idioms (ResourceData, schema.Schema, diag.Diagnostics from SDK) with the new PF patterns. A clean sub-package avoids churn in the parent package and follows the pattern used by `template/`, `ilm/`, and `templateilmattachment/`. It also makes the eventual deletion of the SDK file trivial.

**Alternatives considered:**
- *In-place replacement in `internal/elasticsearch/index/`.* Would require renaming the SDK file and adding `_sdk` suffixes, creating confusion during the transition. Rejected.

### D2. Create and Update use envelope callbacks

**Choice:** Both Create and Update use the same callback (PUT then read-back), fitting the envelope contract.

**Rationale:** The current SDK resource uses `resourceComponentTemplatePut` for both Create and Update. After PUT, it calls `resourceComponentTemplateRead` to refresh state. This is exactly the envelope's `writeFromPlan` flow. The alias routing preservation already lives in the read/flatten path, so no config-derived post-processing is needed.

**Alternatives considered:**
- *Keep Create/Update on concrete type with placeholder callbacks.* Unnecessary; the flow is a clean envelope fit.

### D3. Shared template helpers stay in parent package

**Choice:** Helpers consumed by other SDK code (e.g. alias hashing, `expandTemplate`, `flattenTemplateData`) remain in `internal/elasticsearch/index/`. Helpers only used by component template migrate with the resource.

**Rationale:** Minimizes risk for other SDK consumers. After all SDK resources are gone, a future cleanup change can move the remaining helpers.

**Alternatives considered:**
- *Move all helpers now.* Would force updating unrelated SDK resources in the same change. Rejected to keep scope narrow.

### D4. Schema factory strips `elasticsearch_connection`

**Choice:** The PF schema factory in the new package does NOT declare the `elasticsearch_connection` block. The envelope injects it.

**Rationale:** Matches every other envelope resource. Reduces duplication and ensures the connection block stays consistent with the provider-wide helper.

### D5. Model implements `ElasticsearchResourceModel`

**Choice:** The new `Data` struct declares `GetID()`, `GetResourceID()`, and `GetElasticsearchConnection()` value-receiver methods.

**Rationale:** Required by the envelope type constraint. `GetResourceID` returns `Name`, the natural write identity.

## Risks / Trade-offs

- **Risk:** Alias routing preservation behavior diverges subtly between SDK and PF flatten paths. **Mitigation:** Port the existing `extractAliasRoutingFromTemplateState` and `flattenTemplateData` logic verbatim; acceptance tests cover routing preservation.
- **Risk:** JSON diff suppression semantics differ between SDK `DiffSuppressFunc` and PF custom types. **Mitigation:** Use `jsontypes.Normalized` for `metadata`, `mappings`, and `settings`; it provides equivalent normalization. `template.alias.filter` also uses `jsontypes.Normalized`.
- **Risk:** SDK state upgrade is not applicable (this resource never had a PF predecessor with a different schema shape). **Mitigation:** No state upgrader needed; schema version starts at 1 or 0 as appropriate for a fresh PF resource.
- **Trade-off:** Acceptance tests for component template validate the final behavior but do not exercise SDK→PF upgrade paths. This is acceptable because the resource is newly created in PF; state shape differs from SDK and Terraform handles it via refresh.

## Migration Plan

1. Create `internal/elasticsearch/index/componenttemplate/` with:
   - `models.go` — PF `Data` struct + getters.
   - `schema.go` — PF schema factory without connection block.
   - `expand.go` / `flatten.go` — request/response mapping (port from SDK).
   - `read.go` — `readComponentTemplate` callback.
   - `delete.go` — `deleteComponentTemplate` callback.
   - `create.go` / `update.go` — `createComponentTemplate` / `updateComponentTemplate` callbacks.
   - `resource.go` — concrete type embedding `*entitycore.ElasticsearchResource[Data]` + `ImportState`.
2. Register the new resource in the provider.
3. Run `make build`, `make check-lint`, component template unit tests, and acceptance tests.
4. Delete `internal/elasticsearch/index/component_template.go` and `component_template_test.go`.
5. Re-run build and tests.

**Rollback:** Revert the provider registration and restore the SDK file; both implementations can coexist during development if needed.

## Open Questions

None.
