## 1. Resource audit

- [x] 1.1 Confirm all three resources can use full envelope migration with real write callbacks
- [x] 1.2 Record current Terraform-visible invariants for IDs, read-after-write behavior, and delete semantics
- [x] 1.3 Confirm `install_prebuilt_rules` can retain `ModifyPlan` unchanged on the wrapper while moving CRUD into envelope callbacks

## 2. Envelope migration

- [x] 2.1 Migrate `internal/kibana/defaultdataview/` to `entitycore.NewKibanaResource`
- [x] 2.2 Migrate `internal/kibana/security_enable_rule/` to `entitycore.NewKibanaResource`
- [x] 2.3 Migrate `internal/kibana/prebuilt_rules/` to `entitycore.NewKibanaResource`
- [x] 2.4 Add or adapt model methods required by the Kibana envelope without changing the Terraform schema
- [x] 2.5 Preserve `install_prebuilt_rules` delete behavior as a no-op remote delete

## 3. Validation and regression coverage

- [x] 3.1 Update unit tests for any new model or helper behavior introduced by the migration
- [x] 3.2 Keep or extend acceptance coverage for CRUD and `ModifyPlan`-relevant behavior
- [x] 3.3 Run targeted tests for the three resource packages
- [x] 3.4 Run `make build`
- [x] 3.5 Run `make check-openspec`
