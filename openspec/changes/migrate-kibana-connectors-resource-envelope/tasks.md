## 1. Resource audit

- [ ] 1.1 Confirm full envelope migration is feasible for `action_connector` with real write callbacks
- [ ] 1.2 Record Terraform-visible invariants for import, composite IDs, read-after-write refresh, and state upgrade behavior
- [ ] 1.3 Confirm how version-gated preconfigured connector ID validation should live after migration (model requirement vs callback-local check)

## 2. Envelope migration

- [ ] 2.1 Convert `internal/kibana/connectors/` resource wiring to embed `entitycore.NewKibanaResource`
- [ ] 2.2 Adapt the model to satisfy the Kibana envelope contract without changing the schema
- [ ] 2.3 Move create/update/read/delete orchestration into envelope callbacks while preserving current connector behavior
- [ ] 2.4 Preserve wrapper-level `ImportState` and `UpgradeState` behavior unchanged

## 3. Validation and regression coverage

- [ ] 3.1 Update or add unit tests for any new model/helper or version-gating behavior
- [ ] 3.2 Keep or extend acceptance coverage for CRUD, import, and upgrade-state-relevant behavior
- [ ] 3.3 Run targeted tests for `internal/kibana/connectors/...`
- [ ] 3.4 Run `make build`
- [ ] 3.5 Run `make check-openspec`
