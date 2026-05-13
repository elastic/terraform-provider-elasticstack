## Context

The `elasticstack_kibana_slo` resource's `time_window` block accepts two attributes: `type` (already validated with `stringvalidator.OneOf("rolling", "calendarAligned")`) and `duration` (currently unrestricted).

The Kibana SLO API (confirmed by `generated/kbapi/kibana.gen.go:38723-38730` and the issue reporter's testing against ES 8.12 / provider v0.11.6) restricts valid `duration` values per window type:

| `type` | Valid `duration` values |
|---|---|
| `rolling` | `7d`, `30d`, `90d` |
| `calendarAligned` | `1w`, `1M` |

Any other value produces an HTTP 400 at apply time with an opaque error body that does not identify the offending field.

## Goals / Non-Goals

**Goals:**
- Produce a Terraform plan-time diagnostic that names the valid `duration` values for the configured `type`.
- Fix the inaccurate `time_window.md` description.
- Add a reusable `OneOfWhenDependentPathExpressionEquals` helper to the existing conditional-validator framework.
- Cover the new validators with acceptance test fixtures that fire at plan time (no live Kibana needed).

**Non-Goals:**
- Accepting new valid duration values or requesting the API relax its constraints.
- Version-gating the validator by stack version.
- Changing any other SLO schema attributes.

## Decisions

### Type-conditional validators (Approach C) over union `OneOf` (Approach B)

**Decision:** Two `OneOfWhenDependentPathExpressionEquals` validators on `duration`, one per `type` value.

**Rationale:** A union `OneOf("7d", "30d", "90d", "1w", "1M")` (Approach B) silently permits `type = "rolling"` + `duration = "1w"`, since `1w` appears in the union. The conditional approach produces a diagnostic that names only the values valid for the configured `type` (e.g. *"must be one of [7d, 30d, 90d] when type is \"rolling\""*). The maintainer explicitly requested this: *"Yes, lets be explicit."*

### New `OneOfWhenDependentPathExpressionEquals` constructor

**Decision:** Add `OneOfWhenDependentPathExpressionEquals(dependentPathExpression, dependentValue, allowedValues)` to `internal/utils/validators/conditional.go`.

**Rationale:** The existing `Condition` struct already supports relative path expressions (used by `RequiredIfDependentPathExpressionOneOf`, `ForbiddenIfDependentPathExpressionOneOf`, `AllowedIfDependentPathExpressionOneOf`). The new constructor follows the same pattern with an `allowedValues` list for the *current* attribute (not the dependent one). It validates that `current ∈ allowedValues` when `dependent == dependentValue`, and is a no-op otherwise.

The `strings` package is already imported in `conditional.go` via adjacent helpers; if not, add it. `slices.Contains` (already used in the file) handles the membership check.

### Acceptance tests as plan-only steps

**Decision:** New `TestAccResourceSloValidation` steps use `ConfigDirectory` to load the new test fixtures and `ExpectError: regexp.MustCompile(`(?s)Invalid Attribute Value Match.*duration`)` to assert the plan-time error.

**Rationale:** Conditional validators fire during the Terraform plan phase, before any API call. The tests do not require a live Kibana stack and will pass in environments where only `make check-lint` is run (the plan step does not require real provider credentials when `ExpectError` is set and the error is in config validation).

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| **Kibana adds new valid rolling durations in a future version** | Validator must be updated when the API expands. Same maintenance cost as any enum validation; the list is small and stable since at least ES 8.9. |
| **Technically breaking for users who had invalid-duration configs silently passing plan** | Those configs always fail at apply. Surfacing the failure at plan is strictly an improvement; no valid use case is blocked. A CHANGELOG entry is appropriate. |
| **`strings` import** | Check `conditional.go` imports at implementation time; add `"strings"` if absent, or use `fmt.Sprintf` join directly. |

## Open Questions

- **Are the valid `rolling` duration values (`7d`, `30d`, `90d`) stable across all supported stack versions?** The generated kbapi comment and issue reporter's testing (v0.11.6 / ES 8.12) confirm these three values. If earlier or later stack versions allow additional rolling durations, the validator list may need broadening or version-gating. Non-blocking.
- **Should a CHANGELOG / release notes entry be added?** The description correction is a documentation bug fix; the new validator is technically breaking for anyone relying on an invalid-duration configuration silently passing plan. A note seems appropriate.
