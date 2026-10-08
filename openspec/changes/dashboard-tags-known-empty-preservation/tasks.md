## 1. Fix read-path normalization

- [ ] 1.1 In `internal/kibana/dashboard/models.go` around lines 104-109, replace the unconditional `types.ListNull` fallback for `tags` with an intent-preserving check: when the API returns a nil or empty tags value and prior `m.Tags` is a known value (including known-empty `[]`) or already null, preserve it as-is; only fall back to `types.ListNull` when `m.Tags` is `Unknown` (e.g. import).
- [ ] 1.2 Verify the write path (`models.go` around lines 178-183 and 222-227) still sends `tags` correctly for non-empty and known-empty values — no change expected, but confirm.

## 2. Unit tests

- [ ] 2.1 Add or extend unit tests in `internal/kibana/dashboard/` covering:
  - API returns nil, prior state/plan `tags = []` (known-empty) → state `[]` (the bug scenario).
  - API returns `[]` (empty, non-nil), prior state/plan `tags = []` → state `[]`.
  - API returns nil, prior state/plan `tags` null (omitted in config) → state null.
  - API returns `["a", "b"]`, prior state/plan null or `[]` → state `["a", "b"]` (non-empty normal case).
  - Prior `m.Tags` is `Unknown` (import) and API returns nil/empty → state null.

## 3. Acceptance tests

- [ ] 3.1 Add a targeted acceptance test (or extend an existing dashboard test) that:
  - Creates a dashboard with explicit `tags = []`.
  - Plans and applies — expects no diff on re-plan.
  - Confirms `tags` in state is `[]`, not null.
- [ ] 3.2 Verify existing dashboard acceptance tests that omit `tags` entirely continue to pass unchanged (regression gate for the null-intent path).
- [ ] 3.3 Run the dashboard package's acceptance tests against a running Elastic Stack (see `dev-docs/high-level/testing.md`) and confirm no "inconsistent result" error for the `tags = []` case.

## 4. Spec sync

- [ ] 4.1 After implementation, run `make check-openspec` to confirm the delta spec is aligned with the main spec and no validation errors remain.
