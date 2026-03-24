## 1. Schema and validation

- [ ] 1.1 Convert phase blocks in `schema.go` to `SingleNestedBlock`; set `schema.Schema.Version` to `1`.
- [ ] 1.2 Convert action helpers in `schema_actions.go` to `SingleNestedBlock`; remove `listBlockSingle` / list validators where obsolete.
- [ ] 1.3 For `forcemerge`, `searchable_snapshot`, `set_priority`, `wait_for_snapshot`, `downsample`: make former required attributes **optional**; add **`objectvalidator.AlsoRequires`** for each former required field on the parent action object / block.
- [ ] 1.4 Keep `elasticsearch_connection` as the shared list nested block from `internal/schema/connection.go`.

## 2. Model, flatten, expand

- [ ] 2.1 Update `attr_types.go`: action fields as object types (not `ListType` wrappers); phase top-level as object types.
- [ ] 2.2 Update `models.go`, `validate.go`, `policy.go`, `flatten.go`, `value_conv.go`, `model_expand.go` for `types.Object` phases and object action attributes; preserve `expand.go` contract via wrapping single objects as `[]any{m}` where needed.
- [ ] 2.3 Preserve toggle semantics for `readonly` / `freeze` / `unfollow` in `flatten.go` (`priorHasDeclaredToggle`).

## 3. State upgrade

- [ ] 3.1 Implement `ResourceWithUpgradeState` on the ILM resource; map `0 → migrateV0ToV1` JSON transform (key-aware unwrap only).
- [ ] 3.2 Add `state_upgrade.go` (or equivalent) and unit test `TestILMResourceUpgradeState` with v0- and v1-shaped fixtures.

## 4. Tests and docs

- [ ] 4.1 Update `internal/elasticsearch/index/ilm/acc_test.go` attribute paths for single-nested state (no `.0` for phase/action segments); do not edit `testdata/**/*.tf` HCL.
- [ ] 4.2 Regenerate `docs/resources/elasticsearch_index_lifecycle.md` (or project doc target).
- [ ] 4.3 Run `make build` and targeted ILM acceptance tests.

## 5. OpenSpec

- [ ] 5.1 Keep delta spec `openspec/changes/ilm-single-nested-and-action-validation/specs/elasticsearch-index-lifecycle/spec.md` aligned with implementation.
- [ ] 5.2 After merge decision: **sync** into `openspec/specs/elasticsearch-index-lifecycle/spec.md` or **archive** the change per project workflow; run `make check-openspec`.
