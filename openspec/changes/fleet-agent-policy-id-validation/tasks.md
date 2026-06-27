## 1. Nil-guard fix

- [ ] 1.1 In `internal/fleet/agentpolicy/models.go`, in the `toAPICreateModel` function, change
  `Id: model.PolicyID.ValueStringPointer()` to `Id: typeutils.OptionalString(model.PolicyID)`.
  `typeutils` is already imported; no new import is required.

- [ ] 1.2 Verify that `typeutils` is already imported in `models.go`
  (it is, at `import` line). Confirm the generated `KibanaHTTPAPIsNewAgentPolicy.Id` field
  carries `json:"id,omitempty"` (it does, at `generated/kbapi/kibana.gen.go:53172`).

## 2. Plan-time policy_id validator

- [ ] 2.1 Create `internal/fleet/agentpolicy/validators.go` with a struct `policyIDValidator`
  implementing `validator.String` (from `github.com/hashicorp/terraform-plugin-framework/schema/validator`).

- [ ] 2.2 Implement `Description` and `MarkdownDescription` on `policyIDValidator` returning a
  human-readable description of the constraint.

- [ ] 2.3 Implement `ValidateString` on `policyIDValidator`:
  - Return immediately (no error) if `req.ConfigValue` is null, unknown, or empty string.
  - Return an error if the value is longer than 255 characters or has zero length.
  - Return an error if the value contains `/`.
  - Return an error if the value contains `..`.
  - Return an error if the value equals one of `__proto__`, `constructor`, `prototype`.
  - Each check should produce an error diagnostic naming the violated constraint.

- [ ] 2.4 In `internal/fleet/agentpolicy/schema.go`, add `Validators: []validator.String{policyIDValidator{}}`
  to the `policy_id` attribute.

## 3. Unit tests

- [ ] 3.1 In `internal/fleet/agentpolicy/` (or a `_test.go` file for `models.go`), add a unit
  test for `toAPICreateModel` covering the case where `PolicyID` is `types.StringUnknown()`.
  Assert that the returned body's `Id` field is `nil` (not `&""`).

- [ ] 3.2 Add unit tests for `policyIDValidator` covering:
  - Null value → no error.
  - Unknown value → no error.
  - Empty string → no error (handled by nil-guard, not validator).
  - Valid ID → no error.
  - Length 256 → error.
  - Contains `/` → error.
  - Contains `..` → error.
  - Equals `__proto__` → error.
  - Equals `constructor` → error.
  - Equals `prototype` → error.

## 4. Changelog entry

- [ ] 4.1 Add a changelog entry under the `Bug Fixes` section noting that
  `elasticstack_fleet_agent_policy` no longer sends `"id": ""` to Kibana on create when
  `policy_id` is not set, restoring compatibility with Kibana 9.3.6.
