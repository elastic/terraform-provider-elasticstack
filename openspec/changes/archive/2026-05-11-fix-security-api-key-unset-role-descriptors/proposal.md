## Why

`elasticstack_elasticsearch_security_api_key` crashes during `terraform apply` with a
"Normalized JSON Unmarshal Error / json string value is unknown" when `role_descriptors` is
not set in the configuration.

The attribute is documented as optional. Omitting it should succeed; the Elasticsearch API
treats an absent `role_descriptors` as inheriting the calling credential's permissions.
Setting `role_descriptors = jsonencode({})` is a known workaround, confirming the failure is
a provider-side handling problem, not an API constraint.

## Root Cause

`role_descriptors` is declared `Optional + Computed`. On the first `apply` there is no prior
state, so `stringplanmodifier.UseStateForUnknown()` leaves the value as Unknown in the plan.
Two call sites then attempt to unmarshal that Unknown value as JSON:

1. `toAPICreateRequest()` in `models.go` — calls `model.RoleDescriptors.Unmarshal(...)` unconditionally.
2. `validateRestrictionSupport()` in `create.go` — also calls `model.RoleDescriptors.Unmarshal(...)` unconditionally.

`jsontypes.Normalized.Unmarshal` returns an error when the value is Unknown, producing the
"json string value is unknown" diagnostic.

## What Changes

- Guard both `Unmarshal` call sites with an `IsNull() || IsUnknown()` check, consistent with
  how `Metadata` is already guarded via `typeutils.IsKnown`.
- Add an acceptance test (`TestAccResourceSecurityAPIKeyNoRoleDescriptors`) that creates an
  API key with only `name` set (no `role_descriptors`, no `expiration`) to cover this path.

## Capabilities

### Modified Capabilities
- `elasticsearch-security-api-key`: add requirement that `role_descriptors` absent or null is
  a valid input that must not produce an error during create or update.

## Impact

- Two targeted guards in `internal/elasticsearch/security/api_key/models.go` and
  `internal/elasticsearch/security/api_key/create.go`.
- One new acceptance test function and one new testdata directory in
  `internal/elasticsearch/security/api_key/`.
- No schema changes; no changes to other resources or packages.
