## 1. Fix read-path normalization

- [x] 1.1 In `internal/kibana/dashboard/models.go` around line 52, replace the `StringishPointerValue` call for `description` with an intent-preserving check: if the API returns `""` and prior `m.Description` is null, set `m.Description = types.StringNull()`; otherwise set it from the API value as before.
- [x] 1.2 Verify that the write path (lines ~153–157 and ~212–215 in `models.go`) still sends `description` correctly for non-empty values and omits it (or sends `nil`) when null — no change expected here, but confirm.

## 2. Unit tests

- [x] 2.1 Add or extend unit tests in `internal/kibana/dashboard/` (or a `_test.go` file in that package) covering:
  - API returns `""`, prior state null → state null (the bug scenario).
  - API returns `""`, prior state `""` → state `""` (explicit empty preserved).
  - API returns `"My dashboard"`, prior state null → state `"My dashboard"` (non-empty normal case).
  - API returns nil (pointer nil), prior state null → state null (8.x / omitted-field case).

## 3. Acceptance tests

- [x] 3.1 Verify the approximately 14 existing acceptance tests that omit `description` pass without config changes after the fix (regression gate).
- [x] 3.2 Add one targeted acceptance test (or extend an existing one) that:
  - Creates a dashboard with explicit `description = ""`.
  - Plans and applies — expects no diff on re-plan.
  - Confirms `description` in state is `""`.
- [x] 3.3 Run: `TF_ACC=1 go test -v -run '^TestAccResourceDashboardMetricChartMinimalConfig$' ./internal/kibana/dashboard/panel/lensmetric/...` against a 9.5 stack and confirm no "inconsistent result" error.

## 4. Spec sync

- [x] 4.1 After implementation is merged, run `make check-openspec` to confirm the delta spec is aligned with the main spec and no validation errors remain.
