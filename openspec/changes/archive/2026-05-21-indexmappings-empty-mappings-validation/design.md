## Context

The `elasticstack_elasticsearch_index_mappings` resource uses a custom `index.MappingsType{}` for its `mappings` attribute. This type normalizes JSON and provides semantic equality but does not validate that the JSON represents a non-empty object. The separate `StringIsJSONObject` validator (used by other resources) checks that a string is a JSON object but also permits `{}`.

The resource's read logic treats `{}` in state as "no prior mask" and stores the full API response, which means an empty `mappings` declaration silently captures the entire index mapping. This is semantically broken for a resource whose sole purpose is to declare a subset.

## Goals / Non-Goals

**Goals:**
- Reject `mappings = jsonencode({})` at plan time for the `indexmappings` resource.
- Add unit tests for the parameterised validator.
- Add an acceptance test verifying the rejection.

**Non-Goals:**
- Changing validation behavior on any other resource or attribute (the `index` resource's `mappings` remains unchanged).
- Adding a generic "at least N keys" parameter — `NonEmpty` is sufficient.
- Changing the read/intersection logic — the existing read behavior is fine once `{}` can't reach it.

## Decisions

### Parameter shape: `NonEmpty bool` on `StringIsJSONObject`

- **Decision**: Add a `NonEmpty bool` field to `StringIsJSONObject` struct.
- **Rationale**: The zero-value (`false`) preserves backward compatibility across all 18 existing call sites without touching them. `NonEmpty` communicates intent more clearly than `MinKeys int` for a boolean on/off constraint. As discussed, if we ever need validation of specific sub-keys, that would be a new validator entirely.
- **Alternative considered**: `MinKeys int` — rejected as overkill; no foreseeable use case for "at least 2 keys".

### Add unit tests in `validation_test.go`

- **Decision**: Add a table-driven unit test for `StringIsJSONObject.ValidateString`.
- **Rationale**: `validation_test.go` currently contains only the copyright header; adding coverage while the validator is being modified is the lowest-friction moment. Covers zero-value (pass `{}`) and `NonEmpty: true` (reject `{}`), plus existing non-object rejections.

### Acceptance test: plan-only with `ExpectError`

- **Decision**: Add a new test step in `indexmappings/acc_test.go` that attempts to create the resource with `mappings = jsonencode({})` and expects a validation error.
- **Rationale**: A `PlanOnly` or single-step test with `ExpectError` is sufficient to verify the validator is wired into the schema. No index needs to exist since validation runs before API calls.

## Risks / Trade-offs

- **[Risk]** Adding `NonEmpty` to `StringIsJSONObject` could be re-used incorrectly on Optional fields where `{}` is semantically meaningful (e.g. `ilm.metadata`).
  → **Mitigation**: No existing call sites are modified. Any future use of `NonEmpty: true` requires an intentional code review.
- **[Risk]** Error message from the validator is generic and does not suggest valid top-level keys.
  → **Mitigation**: The Terraform framework path prefix (`mappings: `) gives the user enough context. Generic validator = correct separation of concerns.
