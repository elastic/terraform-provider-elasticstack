## 1. Resource audit

- [x] 1.1 Confirm full envelope migration is still feasible for `data_view` despite custom create reconciliation and update-time namespace/field metadata handling
- [x] 1.2 Record Terraform-visible invariants for import, IDs, namespace handling, and field metadata refresh semantics
- [x] 1.3 Identify any small reusable entitycore seam needed by this migration and confirm it is worth generalizing

## 2. Envelope migration

- [x] 2.1 Convert `internal/kibana/dataview/` to embed `entitycore.NewKibanaResource`
- [x] 2.2 Adapt the model to satisfy the Kibana envelope contract without changing the schema
- [x] 2.3 Move create reconciliation logic into the envelope create callback
- [x] 2.4 Move namespace update and field metadata delta logic into the envelope update callback while preserving current behavior
- [x] 2.5 Preserve existing import-state initialization on the wrapper

## 3. Validation and regression coverage

- [x] 3.1 Update or add unit tests for any new model/identity helper behavior
- [x] 3.2 Keep or extend existing tests covering create reconciliation, namespace updates, field metadata deltas, and import behavior
- [x] 3.3 Run targeted tests for `internal/kibana/dataview/...`
- [x] 3.4 Run `make build`
- [x] 3.5 Run `make check-openspec`
