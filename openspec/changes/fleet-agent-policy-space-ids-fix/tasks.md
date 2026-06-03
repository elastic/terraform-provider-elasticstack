## 1. Fix populateFromAPI

- [x] 1.1 In `internal/fleet/agentpolicy/models.go`, locate the `space_ids` block at lines 211–219
- [x] 1.2 Replace the `else` branch that unconditionally writes `types.SetNull` with a guard
      that retains the existing model value when the API returns nil and the model is non-null,
      following the pattern from Decision 1 in `design.md`
- [x] 1.3 Verify that the change covers all three callers of `populateFromAPI`: `create.go`,
      `read.go`, and `update.go` (no separate changes required — they share `populateFromAPI`)

## 2. Tests

- [x] 2.1 Add or update a unit test in `internal/fleet/agentpolicy/` that calls `populateFromAPI`
      with a `nil` `SpaceIds` field while the model already has a non-null, non-unknown `SpaceIDs`
      set — and asserts the model value is unchanged after the call
- [x] 2.2 Add a complementary unit test case: `nil` `SpaceIds` and model `SpaceIDs` is null →
      assert `SpaceIDs` is still null after the call
- [ ] 2.3 Verify the existing acceptance test `TestAccResourceAgentPolicyWithSpaceIDs` in
      `internal/fleet/agentpolicy/acc_test.go` passes (if it previously failed due to the bug,
      it should now pass without modification)

## 3. Validation and cleanup

- [x] 3.1 Run `make build` — fix any build errors
- [x] 3.2 Run `make check-lint` — fix any linter issues
- [x] 3.3 Run `make check-openspec` — confirm change validates
- [x] 3.4 Run the targeted unit tests: `go test ./internal/fleet/agentpolicy/...`
