## Context

`elasticstack_elasticsearch_security_api_key` declares `role_descriptors` as
`Optional + Computed`. On first create (no prior state), the value is Unknown in the plan
because `stringplanmodifier.UseStateForUnknown()` only substitutes a prior-state value — and
there is none. Two call sites in the create path call `model.RoleDescriptors.Unmarshal(...)`
without first checking whether the value is Known:

- `tfModel.toAPICreateRequest()` in `models.go` (lines ~106-110)
- `Resource.validateRestrictionSupport()` in `create.go` (lines ~87-91)

`jsontypes.Normalized.Unmarshal` (the underlying method) explicitly returns an error when the
value is Unknown, producing the user-visible "json string value is unknown" diagnostic.

The workaround (`role_descriptors = jsonencode({})`) provides a known JSON string to both
failing `Unmarshal` call sites. This confirms the fix is purely a null/unknown guard — no API
behaviour change.

## Goals / Non-Goals

**Goals:**
- Make creating an API key without `role_descriptors` succeed on the first apply.
- Add regression coverage via an acceptance test.

**Non-Goals:**
- Changing the schema, plan modifiers, or semantics of `role_descriptors`.
- Handling `role_descriptors` in update paths differently (update has the same pattern;
  address it in the same guarded change).
- Modifying any other resource.

## Decisions

### D1. Guard pattern matches existing `Metadata` handling

**Choice:** Wrap both `Unmarshal` call sites with an `IsNull() || IsUnknown()` check (or use
`typeutils.IsKnown` which expresses the same thing) before attempting to parse.

**Rationale:** The `Metadata` field in the same `toAPICreateRequest()` already uses
`typeutils.IsKnown(model.Metadata)` as a precondition. Applying the same idiom to
`RoleDescriptors` keeps the two fields consistent and avoids introducing a new pattern.

**Alternatives considered:**
- *Return empty map on Unknown.* Semantically identical for the API call (absent = empty),
  but the nil/empty distinction matters for `len(roleDescriptors) > 0` guard downstream.
  A nil map from skipping the block satisfies that guard without injecting a spurious `{}`.
- *Set a default empty value via plan modifier.* Would change observable plan output and
  potentially affect drift detection. Rejected for scope.

### D2. `validateRestrictionSupport` skips when Unknown/Null

**Choice:** Return early from `validateRestrictionSupport` when
`model.RoleDescriptors.IsNull() || model.RoleDescriptors.IsUnknown()`.

**Rationale:** If there are no role descriptors, there can be no `restriction` blocks, so
validation is a no-op. Returning early is safer than attempting to unmarshal an unset value.

### D3. Acceptance test covers create-only (no update step)

**Choice:** A single-step test that creates an API key with only `name` set (no
`role_descriptors`, no `expiration`), asserts the resource was created, and checks that
`id`, `key_id`, `api_key`, and `encoded` are set in state.

**Rationale:** The regression is a create-time crash. A create-and-destroy test is sufficient
to prove the fix. Update behaviour for this scenario is covered by the existing update-capable
tests with `role_descriptors` set.

## Risks / Trade-offs

- **Risk:** Skipping `Unmarshal` when Unknown leaves `roleDescriptors` as nil, so
  `len(roleDescriptors) > 0` is false and `req.RoleDescriptors` stays nil. The Elasticsearch
  API accepts nil `role_descriptors` as "inherit caller privileges", which is the correct
  behaviour.
- **Trade-off:** The `update` path in `toUpdateAPIRequest()` has the same pattern. This
  change guards both create and update in a single commit for completeness.

## Open Questions

None.
