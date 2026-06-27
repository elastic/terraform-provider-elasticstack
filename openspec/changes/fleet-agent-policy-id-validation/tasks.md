## 1. Nil-guard fix

- [ ] 1.1 In `internal/fleet/agentpolicy/models.go`, in the `toAPICreateModel` function, change
  `Id: model.PolicyID.ValueStringPointer()` to `Id: typeutils.OptionalString(model.PolicyID)`.
  `typeutils` is already imported; no new import is required.

## 2. Plan-time policy_id validator

- [ ] 2.1 Create `internal/fleet/agentpolicy/validators.go` with a struct `policyIDValidator`
  implementing `validator.String` (from `github.com/hashicorp/terraform-plugin-framework/schema/validator`).

- [ ] 2.2 Implement `Description` and `MarkdownDescription` on `policyIDValidator` returning a
  human-readable description of the constraint.

- [ ] 2.3 Implement `ValidateString` on `policyIDValidator`:
  - Return immediately (no error) if `req.ConfigValue` is null or unknown.
  - Return an error if the value length is not between 1 and 255 characters (inclusive).
    Explicit empty string (`""`) is rejected here as a length-0 violation.
  - Return an error if the value contains `/`.
  - Return an error if the value contains `..`.
  - Return an error if the value contains any of the substrings `__proto__`, `constructor`,
    `prototype`. (Substring match, not equality — per Kibana's "must not contain" wording.)
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
  - Empty string → error (length-0 violates the 1–255 constraint).
  - Valid ID → no error.
  - Length 256 → error.
  - Contains `/` → error.
  - Contains `..` → error.
  - Bare `__proto__`, `constructor`, `prototype` → error.
  - Contains `__proto__`, `constructor`, or `prototype` as a substring (e.g.
    `"my-__proto__-policy"`) → error.

- [ ] 3.3 Add an acceptance test in `internal/fleet/agentpolicy/` (or extend an existing one)
  that creates an `elasticstack_fleet_agent_policy` resource without setting `policy_id`,
  applies against the running Kibana, and asserts the resource is created successfully with
  `policy_id` populated by Fleet and an empty plan on re-plan. This exercises the
  plan→apply→API round trip that the regression broke.

- [ ] 3.4 Add a `PlanOnly` (or equivalent) acceptance test step asserting that
  `policy_id = ""` and `policy_id = "bad/id"` each fail at plan time with the corresponding
  validator error, confirming the validator is wired into the schema correctly.

## 4. Changelog entry

- [ ] 4.1 Add a changelog entry under the `Bug Fixes` section noting that
  `elasticstack_fleet_agent_policy` no longer sends `"id": ""` to Kibana on create when
  `policy_id` is not set, restoring compatibility with Kibana 9.3.6.
