## Context

The provider already exposes a family of conditional validators in `internal/utils/validators/conditional.go` that encode plan-time cross-attribute rules such as allowed, required, and forbidden combinations. The `elasticstack_kibana_space` resource needs a more nuanced rule than `ConflictsWith`: `disabled_features` must be allowed when `solution` is `classic`, allowed when `solution` is omitted, and not rejected while `solution` remains unknown/computed during validation. Existing `AllowedIfDependentPathEquals` behavior is too strict because it only accepts explicit string equality and treats null and unknown dependent values as failures.

## Goals / Non-Goals

**Goals:**
- Add an explicit conditional-validator path for "equals or null" behavior without overloading `nil` or changing existing caller signatures broadly.
- Make `AllowedIf...` validators treat an unknown dependent value as allowed so computed values do not cause premature validation failures.
- Reuse the new validator behavior in `elasticstack_kibana_space` so `disabled_features` is validated conditionally instead of with unconditional conflicts.
- Cover the validator and resource schema behavior with focused tests.

**Non-Goals:**
- Redesign all conditional validators around fully-typed Terraform values.
- Change the semantics of `ForbiddenIf...` or `RequiredIf...` validators beyond the existing behavior.
- Introduce resource-level `ValidateConfig` for `elasticstack_kibana_space`.

## Decisions

### Add a dedicated `AllowedIfDependentPathEqualsOrNull` helper
A new helper will wrap the existing `AllowedIfDependentPathOneOf` behavior and mark that a null dependent value is acceptable. This keeps call sites explicit and self-documenting, avoids using `nil` as an overloaded signal, and limits the scope of change to the `AllowedIf` path that needs the new behavior.

Alternative considered: overload `nil` in the existing `[]string` API to mean "allow null". Rejected because it is ambiguous, difficult to discover, and does not cleanly capture unknown/computed semantics.

Alternative considered: change the API to accept `types.String` or other Terraform typed values. Rejected for now because it is more invasive than needed for this change and would create unnecessary churn across existing validator call sites.

### Treat unknown dependent values as allowed for `AllowedIf...`
The internal dependent-value evaluation for `AllowedIf...` will consider an unknown dependent value to satisfy the condition. This matches Terraform planning behavior for computed attributes and avoids false-positive validation errors before the dependent value is known.

Alternative considered: add an `AllowUnknown` option. Rejected because unknown should always pass for `AllowedIf` validation and an extra option would add complexity without meaningful control.

### Update Kibana space schema to use conditional validation instead of reciprocal conflicts
The `disabled_features` attribute in `internal/kibana/spaces/resource_schema.go` will use the new helper against `solution`. The reciprocal `ConflictsWith` validators between `disabled_features` and `solution` will be removed so the resource accepts `solution = "classic"` and omitted `solution` while still rejecting concrete non-`classic` values through the conditional validator.

Alternative considered: implement the rule in resource-level `ValidateConfig`. Rejected because the conditional validator utilities can express the rule directly and keep the validation logic colocated with the schema.

## Risks / Trade-offs

- [Broader `AllowedIf` semantics for unknown dependents] → Mitigation: constrain the semantic change to treating unknown as allowed only for `AllowedIf...`, add targeted tests, and avoid changing `RequiredIf` / `ForbiddenIf` behavior.
- [New helper adds another API surface] → Mitigation: keep the helper narrowly scoped to the concrete requirement (`equals or null`) and document its intended use through tests and descriptive naming.
- [Resource behavior may diverge from current tests or assumptions] → Mitigation: update `elasticstack_kibana_space` schema tests or acceptance coverage to explicitly cover `classic`, unset, unknown, and non-`classic` cases.
