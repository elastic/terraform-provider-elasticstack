## 1. Read-side fix

- [ ] 1.1 In `internal/elasticsearch/security/user/read.go`, replace the unconditional
  `state.Metadata = jsontypes.NewNormalizedNull()` in the `else` branch of
  `if len(user.Metadata) > 0` (lines 69–71) with a conditional that preserves the incoming state
  value when it already holds an empty JSON object:

  ```go
  } else {
      if !isEmptyJSONObject(state.Metadata) {
          state.Metadata = jsontypes.NewNormalizedNull()
      }
  }
  ```

- [ ] 1.2 Add the `isEmptyJSONObject` helper function to the same file (or to a small `helpers.go`
  in the same package if one already exists):

  ```go
  func isEmptyJSONObject(v jsontypes.Normalized) bool {
      if v.IsNull() || v.IsUnknown() {
          return false
      }
      var m map[string]any
      return json.Unmarshal([]byte(v.ValueString()), &m) == nil && len(m) == 0
  }
  ```

  The `encoding/json` import is already present in `read.go`.

## 2. Unit test

- [ ] 2.1 In the `securityuser` package (or a `_test` file in `internal/elasticsearch/security/user/`),
  add a table-driven unit test for `isEmptyJSONObject` covering: `null` value, `"{}"` value,
  `"{\"k\":\"v\"}"` value, and unknown value. Assert that only `"{}"` returns `true`.

## 3. Acceptance test

- [ ] 3.1 In `internal/elasticsearch/security/user/acc_test.go`, add a test step to
  `TestAccResourceSecurityUser` (or a new `TestAccResourceSecurityUserEmptyMetadata`) that:
  - Creates a user with `metadata = jsonencode({})`.
  - Asserts the apply completes without a "Provider produced inconsistent result after apply" error.
  - Checks that `metadata` in state equals `"{}"` using `resource.TestCheckResourceAttr`.
  - Optionally: runs a second `terraform plan` step to confirm no perpetual diff.

## 4. Delta spec

- [ ] 4.1 Keep the delta spec at
  `openspec/changes/security-user-empty-metadata-inconsistency/specs/elasticsearch-security-user/spec.md`
  aligned with the implementation: it amends REQ-016/REQ-017 to document the null/empty-object
  equivalence invariant on the read side.

- [ ] 4.2 After merge: sync into `openspec/specs/elasticsearch-security-user/spec.md` or archive
  this change per project workflow; run `make check-openspec`.
