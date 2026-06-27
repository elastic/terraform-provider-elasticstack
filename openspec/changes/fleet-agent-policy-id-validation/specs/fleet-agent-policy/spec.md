## MODIFIED Requirements

### Requirement: Create behavior — omit id when unset (REQ-018 addendum)

When `policy_id` is not set in config (null or unknown), the resource SHALL omit the `id`
field from the Fleet Create Agent Policy POST body entirely, allowing Fleet to auto-generate
a UUID. The resource SHALL NOT send `"id": ""` to the API. An explicit `policy_id = ""` in
config is rejected at plan time by REQ-037 and therefore never reaches the create payload.

#### Scenario: policy_id unset — id omitted from create body

- GIVEN `policy_id` is not set in config
- WHEN create runs and the Fleet Create API is called
- THEN the POST body SHALL NOT contain an `"id"` field
- AND Fleet SHALL auto-assign a policy ID which is stored in `policy_id` state

#### Scenario: policy_id set — id sent in create body

- GIVEN `policy_id = "my-policy-id"` is set in config
- WHEN create runs and the Fleet Create API is called
- THEN the POST body SHALL contain `"id": "my-policy-id"`

### Requirement: policy_id plan-time validation (REQ-037)

The resource SHALL validate `policy_id` at plan time when an explicit value is provided.
A supplied `policy_id` value SHALL satisfy all of the following constraints:

1. Length is between 1 and 255 characters (inclusive).
2. Does not contain `/` (path separator).
3. Does not contain `..` (traversal sequence).
4. Is not equal to any of the reserved keys: `__proto__`, `constructor`, `prototype`.

When any constraint is violated, the resource SHALL surface a plan-time error diagnostic
naming the violated constraint. The validator SHALL NOT produce an error for null or unknown
values (those are treated as "not set" and the id field is omitted per REQ-018 addendum
above). An explicit empty string (`policy_id = ""`) SHALL be rejected as a length-0
violation of constraint 1.

#### Scenario: Valid explicit policy_id passes validation

- GIVEN `policy_id = "my-valid-policy"` is set in config
- WHEN a plan is generated
- THEN no validation errors SHALL be produced

#### Scenario: Explicit empty policy_id fails validation

- GIVEN `policy_id = ""` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the length constraint

#### Scenario: Unset policy_id passes validation

- GIVEN `policy_id` is not set in config (null or unknown)
- WHEN a plan is generated
- THEN the plan-time validator SHALL NOT produce an error
- AND the `id` field SHALL be omitted from the POST body (per REQ-018 addendum)

#### Scenario: policy_id with path separator fails validation

- GIVEN `policy_id = "my/policy"` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the `/` constraint

#### Scenario: policy_id with traversal sequence fails validation

- GIVEN `policy_id = "my..policy"` is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the `..` constraint

#### Scenario: policy_id exceeds maximum length fails validation

- GIVEN `policy_id` is set to a string of 256 or more characters
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the length constraint

#### Scenario: Reserved key policy_id fails validation

- GIVEN `policy_id = "__proto__"` (or `"constructor"` or `"prototype"`) is set in config
- WHEN a plan is generated
- THEN a plan-time error diagnostic SHALL be produced indicating the reserved-key constraint
