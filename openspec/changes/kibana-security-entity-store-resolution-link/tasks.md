## 1. Spec

- [x] 1.1 Keep delta specs aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-security-entity-store-resolution-link --type change` (or `make check-openspec` after sync).
- [x] 1.2 Resolve open question on set-diff update atomicity (see `design.md` Q1); update delta spec with the chosen update strategy (set-diff or RequiresReplace).
- [x] 1.3 Verify minimum Stack version (9.1.0 vs possible 8.x backport); update `EnforceMinVersion` call and delta spec version gate accordingly.
- [ ] 1.4 On completion of implementation, **sync** delta specs into canonical `openspec/specs/` or **archive** the change per project workflow.

## 2. Resource — `elasticstack_kibana_security_entity_store_entity_link`

- [x] 2.1 Create package `internal/kibana/security_entity_store_entity_link/`. Add `resource.go` with factory function `NewEntityLinkResource()`, `Metadata`, `Configure`, and `ImportState` (parses `<space_id>/<target_id>`); add PF interface assertions.
- [x] 2.2 Add `schema.go` with the resource schema:
  - `id` (computed string, UseStateForUnknown) — composite `<space_id>/<target_id>`.
  - `space_id` (optional, computed string, default `"default"`, RequiresReplace).
  - `target_id` (required string, RequiresReplace).
  - `entity_ids` (required `schema.SetAttribute` of string, 1–1000 items; custom validator: target_id not in entity_ids).
  - `resolution_group_json` (computed `jsontypes.NormalizedType{}`).
  - `kibana_connection` block via `schema.GetKbFWConnectionBlock()`.
- [x] 2.3 Add `models.go` with `entityLinkModel` (Terraform PF model struct with `ID`, `SpaceID`, `TargetID`, `EntityIDs`, `ResolutionGroupJSON`, `KibanaConnection` fields).
- [x] 2.4 Add `create.go`: call `PostSecurityEntityStoreResolutionLink` with `{target_id, entity_ids}` + space request editor; on 200, call `read()` to populate final state; enforce `EnforceMinVersion("9.1.0")`.
- [x] 2.5 Add `read.go`: call `GetSecurityEntityStoreResolutionGroup` with `entity_id = target_id`; if 404, remove resource from state; parse raw JSON body, store in `resolution_group_json`; verify managed `entity_ids` are all present in the response (surface a warning diagnostic if any are missing — they may have been removed out-of-band).
- [x] 2.6 Add `update.go`: compute set-diff between plan `entity_ids` and current state `entity_ids`; call `PostSecurityEntityStoreResolutionLink` for added IDs; call `PostSecurityEntityStoreResolutionUnlink` for removed IDs; call `read()` to populate final state.
- [x] 2.7 Add `delete.go`: call `PostSecurityEntityStoreResolutionUnlink` with the managed `entity_ids`; treat 404 as already-deleted (no error).
- [x] 2.8 Register the resource in `provider/plugin_framework.go`.

## 3. Data Source — `elasticstack_kibana_security_entity_store_resolution_group`

- [ ] 3.1 Create package `internal/kibana/security_entity_store_resolution_group/`. Add `data_source.go` with factory `NewResolutionGroupDataSource()`, `Metadata`, and `Configure`; add PF interface assertions.
- [ ] 3.2 Add `schema.go` for data source schema:
  - `id` (computed string) — composite `<space_id>/<entity_id>`.
  - `space_id` (optional, computed string, default `"default"`).
  - `entity_id` (required string).
  - `resolution_group_json` (computed `jsontypes.NormalizedType{}`).
  - `kibana_connection` block via `schema.GetKbFWConnectionBlock()`.
- [ ] 3.3 Add `models.go` with `resolutionGroupModel` (PF model struct).
- [ ] 3.4 Add `read.go` for the data source Read: call `GetSecurityEntityStoreResolutionGroup` with `entity_id`; enforce `EnforceMinVersion("9.1.0")`; parse and store raw JSON body in `resolution_group_json`.
- [ ] 3.5 Register the data source in `provider/plugin_framework.go`.

## 4. Testing

- [ ] 4.1 Add acceptance test in `internal/kibana/security_entity_store_entity_link/acc_test.go`:
  - Skip if enterprise license unavailable (detect 403 from link API or use `acctest.SkipIfEnterpriseNotAvailable`-style helper).
  - **Step 1**: Create a link resource with 2 entity_ids; assert `id`, `target_id`, `entity_ids`, `resolution_group_json` in state.
  - **Step 2**: Add a third entity_id (update); assert state reflects new set.
  - **Step 3**: Remove one entity_id (update); assert state reflects reduced set.
  - **Step 4**: Import via `<space_id>/<target_id>`; assert state reconstructed correctly.
  - **Step 5**: Destroy; assert no residual links.
- [ ] 4.2 Add acceptance test for the data source in `internal/kibana/security_entity_store_resolution_group/acc_test.go`:
  - Depends on a linked resource from 4.1 (or creates its own).
  - Assert `resolution_group_json` is non-empty and contains the expected entity IDs.
- [ ] 4.3 Add acceptance test for schema validation:
  - Expect plan-time error when `entity_ids` contains `target_id` (self-link).
  - Expect plan-time error when `entity_ids` is empty.
- [ ] 4.4 Add acceptance test in a non-default space (set `space_id = "test-space"`), if the test environment supports it.
- [ ] 4.5 Add unit tests for:
  - `BuildSpaceAwarePath` integration (covered by existing tests; verify request editors are applied correctly in resource logic).
  - Set-diff logic in update (new/removed IDs, all-new, all-removed edge cases).
  - `id` composite construction and `ImportState` parsing.
  - Self-link validator (positive and negative cases).
