## 1. Resource audit

- [ ] 1.1 Confirm both resources can use full envelope migration with real write callbacks
- [ ] 1.2 Record existing import and ID invariants that must remain unchanged
- [ ] 1.3 Confirm no hidden monitor-specific patterns leak into these smaller resources

## 2. Envelope migration

- [ ] 2.1 Migrate `internal/kibana/synthetics/parameter/` to `entitycore.NewKibanaResource`
- [ ] 2.2 Migrate `internal/kibana/synthetics/privatelocation/` to `entitycore.NewKibanaResource`
- [ ] 2.3 Add or adapt model methods required by the Kibana envelope without changing Terraform-visible behavior

## 3. Validation and regression coverage

- [ ] 3.1 Update unit tests for any model/helper changes introduced by the migration
- [ ] 3.2 Keep or extend acceptance coverage for CRUD and import behavior
- [ ] 3.3 Run targeted tests for the two resource packages
- [ ] 3.4 Run `make build`
- [ ] 3.5 Run `make check-openspec`
