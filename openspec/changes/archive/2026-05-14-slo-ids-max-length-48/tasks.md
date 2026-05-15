## 1. Update SLO ID validation

- [x] 1.1 Change `internal/kibana/slo/schema.go` so `slo_id` allows up to 48 characters and update the attribute description to document the new limit.
- [x] 1.2 Review nearby SLO schema comments or docs for stale 36-character wording and align them with the new limit.

## 2. Add regression coverage

- [x] 2.1 Add or update an acceptance test fixture under `internal/kibana/slo/testdata/` that configures a valid 48-character `slo_id`.
- [x] 2.2 Update `internal/kibana/slo/acc_test.go` assertions so the acceptance suite verifies the 48-character `slo_id` is accepted and stored in state.

## 3. Verify requirements alignment

- [x] 3.1 Confirm the implementation and tests satisfy the `kibana-slo` delta spec for 48-character `slo_id` acceptance and rejection only above the new limit.
- [x] 3.2 Run the relevant OpenSpec and Go test commands for the touched SLO resource paths.