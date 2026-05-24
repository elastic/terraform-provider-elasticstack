## 1. Resource audit

- [ ] 1.1 Confirm full envelope migration is still feasible for `data_view` despite custom create reconciliation and update-time namespace/field metadata handling
- [ ] 1.2 Record Terraform-visible invariants for import, IDs, namespace handling, and field metadata refresh semantics
- [ ] 1.3 Identify any small reusable entitycore seam needed by this migration and confirm it is worth generalizing

## 2. Envelope migration

- [ ] 2.1 Convert `internal/kibana/dataview/` to embed `entitycore.NewKibanaResource`
- [ ] 2.2 Adapt the model to satisfy the Kibana envelope contract without changing the schema
- [ ] 2.3 Move create reconciliation logic into the envelope create callback
- [ ] 2.4 Move namespace update and field metadata delta logic into the envelope update callback while preserving current behavior
- [ ] 2.5 Preserve existing import-state initialization on the wrapper

## 3. Validation and regression coverage

- [ ] 3.1 Update or add unit tests for any new model/identity helper behavior
- [ ] 3.2 Keep or extend existing tests covering create reconciliation, namespace updates, field metadata deltas, and import behavior
- [ ] 3.3 Run targeted tests for `internal/kibana/dataview/...`
- [ ] 3.4 Run `make build`
- [ ] 3.5 Run `make check-openspec`
