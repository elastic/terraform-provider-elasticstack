## Why

Kibana 9.3.6 introduced strict validation on the agent policy `id` field, requiring it to be
1–255 characters and free of path separators, traversal sequences, and reserved object keys.

When `policy_id` is not set in Terraform config, `elasticstack_fleet_agent_policy` sends
`"id": ""` in the POST body. Before 9.3.6, Kibana silently ignored an empty string and
auto-generated a UUID. On 9.3.6, the empty string fails validation with HTTP 400:

```
id is not valid: must be 1–255 characters and must not contain path separators ("/"),
traversal sequences (".."), or reserved keys ("__proto__", "constructor", "prototype").
```

Root cause: `model.PolicyID.ValueStringPointer()` returns `&""` when `PolicyID` is in the
**unknown** state (Computed+Optional field with no prior state and no user value). Go's
`encoding/json` omits `nil` pointers with `omitempty`, but not `*""`, so `"id": ""` is always
serialised even when no ID was specified.

## What Changes

- **Nil-guard fix** (`internal/fleet/agentpolicy/models.go`): Change the `Id` field assignment
  from `model.PolicyID.ValueStringPointer()` to `typeutils.OptionalString(model.PolicyID)`.
  `OptionalString` returns `nil` for null, unknown, or empty-string values, so the `id` field
  is omitted from JSON when unset and Fleet auto-generates a UUID — restoring pre-9.3.6
  behaviour.

- **Plan-time validator** (`internal/fleet/agentpolicy/validators.go`): Add a custom
  `policyIDValidator` enforcing the exact constraints from the Kibana 9.3.6 error message.
  Attach it to the `policy_id` schema attribute so that users who supply an invalid explicit ID
  see a clear error at `terraform plan` rather than at `terraform apply`.

## Capabilities

### Modified Capabilities

- `fleet-agent-policy`: Fix nil-guard regression and add plan-time `policy_id` validation
  reflecting Kibana 9.3.6 ID constraints.

## Impact

- **Bug fix**: Restores compatibility with Kibana 9.3.6 for `elasticstack_fleet_agent_policy`
  resources that omit `policy_id`.
- **New validator**: Surfaces invalid explicit `policy_id` values at plan time with a clear
  error message.
- **Affected files**: `internal/fleet/agentpolicy/models.go`,
  `internal/fleet/agentpolicy/schema.go`, new
  `internal/fleet/agentpolicy/validators.go`.
- **No breaking changes**: A user who omitted `policy_id` will continue to get an
  auto-generated ID. A user who set a valid explicit `policy_id` is unaffected. Only an
  invalid explicit `policy_id` (which already triggered a runtime error) will now be caught
  earlier at plan time.
- **Backward compatibility**: Compatible with Kibana versions older than 9.3.6 — omitting
  `"id"` from the create body was already valid there.
