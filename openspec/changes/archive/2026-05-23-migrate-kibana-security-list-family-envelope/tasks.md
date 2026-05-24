## 1. Resource audit

- [x] 1.1 Confirm each resource schema already separates provider-managed `kibana_connection` handling cleanly enough for envelope injection, or capture the minimal schema reshaping needed
- [x] 1.2 Identify each resource's current state ID and import semantics and record any invariants that must remain unchanged
- [x] 1.3 Confirm all four resources can use full envelope migration (no placeholder write callbacks)

## 2. Envelope migration

- [x] 2.1 Migrate `internal/kibana/securitylist/` to `entitycore.NewKibanaResource`
- [x] 2.2 Migrate `internal/kibana/securitylistitem/` to `entitycore.NewKibanaResource`
- [x] 2.3 Migrate `internal/kibana/securityexceptionlist/` to `entitycore.NewKibanaResource`
- [x] 2.4 Migrate `internal/kibana/security_list_data_streams/` to `entitycore.NewKibanaResource`
- [x] 2.5 Add any shared model/helper adjustments needed so each resource satisfies the Kibana envelope contract without changing Terraform-visible behavior (no shared adjustments needed)

## 3. Validation and regression coverage

- [x] 3.1 Update or add unit tests for any model/identity helpers introduced by the migration
- [x] 3.2 Keep or extend acceptance coverage to prove import, CRUD, and not-found behavior remain unchanged for each migrated resource (existing `acc_test.go` in each package covers import, CRUD, and not-found; no new acceptance tests required)
- [x] 3.3 Run targeted tests for the four resource packages
- [x] 3.4 Run `make build` (`go build ./...` succeeded)
- [x] 3.5 Run `make check-openspec`
