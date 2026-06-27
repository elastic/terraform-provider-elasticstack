## Context

`elasticstack_fleet_agent_policy` sends `"id": ""` to the Fleet Create Agent Policy API when
`policy_id` is not set by the user. The `policy_id` attribute is Computed+Optional, so the
Framework puts it into the **unknown** state before the first apply. In that state,
`types.String.ValueStringPointer()` returns `&""` (a pointer to an empty string). Because the
generated `KibanaHTTPAPIsNewAgentPolicy.Id` field carries `json:"id,omitempty"`, Go's
`encoding/json` only omits the field for a `nil` pointer — a non-nil pointer to `""` is
serialised as `"id": ""`.

Before Kibana 9.3.6, that empty string was silently accepted and an ID was auto-generated.
Kibana 9.3.6 introduced strict validation and rejects it with HTTP 400.

The same `models.go` file also calls `ValueStringPointer()` for `DataOutputId`,
`DownloadSourceId`, `FleetServerHostId`, and `MonitoringOutputId`. These are nullable
(null, not unknown) at create time because they are Optional-only without Computed, so they
produce a `nil` pointer and are correctly omitted — no fix is needed for those fields right now.

## Decisions

### Decision 1: Nil-guard via `typeutils.OptionalString`

Replace `model.PolicyID.ValueStringPointer()` with `typeutils.OptionalString(model.PolicyID)`
in `toAPICreateModel`.

`typeutils.OptionalString` returns `nil` when the value is null, unknown, or an empty string.
Since `KibanaHTTPAPIsNewAgentPolicy.Id` is `*string \`json:"id,omitempty"\``, a `nil` pointer
causes `encoding/json` to omit the field entirely — Fleet then auto-generates a UUID.

**Why `OptionalString` rather than a nil-guard inline:** `OptionalString` is the established
pattern for this class of problem in the codebase (see `internal/utils/typeutils/tfsdk_primitives.go`).
Using it keeps the fix one line and consistent with other optional-string fields.

**Interaction with the validator:** an explicit `policy_id = ""` is rejected by the plan-time
validator (Decision 2) as a length-0 violation, so it never reaches `toAPICreateModel`.
`OptionalString` also treats `""` as "not set" and returns `nil`, which acts as defense in depth
should a future code path bypass the validator — but the user-visible behavior is the
plan-time error, not silent field omission.

### Decision 2: Plan-time `policy_id` validator

Add a custom `policyIDValidator` in a new file `internal/fleet/agentpolicy/validators.go`,
following the pattern of `internal/kibana/validators/is_iso8601_string.go` but as a struct
implementing `validator.String`.

The validator enforces the exact constraints from the Kibana 9.3.6 error message:

1. Length between 1 and 255 characters (inclusive).
2. Does not contain `/` (path separator).
3. Does not contain `..` (traversal sequence).
4. Is not one of the reserved keys: `__proto__`, `constructor`, `prototype`.

The validator returns early (no error) for null or unknown values — those are not
user-supplied IDs. An explicit empty string (`policy_id = ""`) is treated as a length-0
violation and produces a plan-time error (constraint 1).

The validator is attached to the `policy_id` attribute in `schema.go`:

```go
Validators: []validator.String{policyIDValidator{}},
```

**Why a custom struct rather than `stringvalidator.RegexMatches`:** Go RE2 does not support
lookaheads, so all four constraints cannot be expressed as a single regex. A custom struct
validator that checks each constraint explicitly produces a clearer error message and is easier
to maintain.

**Why at plan time rather than apply time:** Plan-time surfacing is strictly better UX — the
user sees the error before any infrastructure change occurs.

### Decision 3: No changes to `toAPIUpdateModel`

The update request body does not include an `id` field; it uses `policy_id` from state in the request path (`PUT /api/fleet/agent_policies/{agentPolicyId}`). This bug is specific to the create payload (`toAPICreateModel`) and does not require a corresponding change for updates.

## Open questions

- Should `DataOutputId`, `DownloadSourceId`, `FleetServerHostId`, and `MonitoringOutputId`
  (also using raw `ValueStringPointer()` at `models.go:387–390`) be audited? Those are null
  (not unknown) at create time so there is no current breakage, but a follow-up audit may be
  worth doing.
- Is a unit test for `toAPICreateModel` with `PolicyID = types.StringUnknown()` in scope for
  this PR? Recommended to include to prevent regression.
