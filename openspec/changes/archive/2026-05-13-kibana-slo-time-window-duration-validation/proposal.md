## Why

The `elasticstack_kibana_slo` resource's `time_window.duration` attribute has two problems:

1. **Inaccurate documentation** — `internal/kibana/slo/descriptions/time_window.md` states *"Any duration greater than 1 day can be used: days, weeks, months, quarters, years"* for rolling windows. This is incorrect. The Kibana API only accepts `7d`, `30d`, and `90d` for rolling windows, and `1w` or `1M` for calendar-aligned windows.
2. **No client-side validation** — `duration` has no `Validators` attached. Users discover invalid values only at `terraform apply` time with an opaque HTTP 400 response, even though the sibling `type` attribute already uses `stringvalidator.OneOf`.

The maintainer (@tobio) confirmed (2026-05-11) that a type-conditional cross-field validator is the right fix: *"Yes, lets be explicit."*

## What Changes

- **Fix the description** in `internal/kibana/slo/descriptions/time_window.md` to accurately list the permitted `duration` values per window `type`.
- **Add `OneOfWhenDependentPathExpressionEquals`** to `internal/utils/validators/conditional.go` — a new `Condition` constructor that validates the current attribute's value against a restricted set only when a sibling attribute equals a specific value.
- **Add two conditional validators** to `duration` in `internal/kibana/slo/schema.go`: one for `type = "rolling"` (allowed: `7d`, `30d`, `90d`) and one for `type = "calendarAligned"` (allowed: `1w`, `1M`).
- **Add two acceptance test fixtures** and two new validation test steps in `TestAccResourceSloValidation` to exercise both conditional validators at plan time (no live Kibana required).
- **Regenerate `docs/resources/kibana_slo.md`** via `make generate`.

## Capabilities

### Modified Capabilities
- `elasticstack_kibana_slo.time_window.duration` — adds plan-time conditional validation; error message names the valid values for the configured `type`. Description corrected to reflect actual API constraints.

### New Capabilities
- `internal/utils/validators.OneOfWhenDependentPathExpressionEquals` — reusable conditional validator for "value must be one of X when sibling field equals Y" constraints.

## Impact

- `internal/kibana/slo/descriptions/time_window.md` — description correction
- `internal/utils/validators/conditional.go` — new `OneOfWhenDependentPathExpressionEquals` helper
- `internal/kibana/slo/schema.go` — two conditional validators on `duration`
- `internal/kibana/slo/testdata/TestAccResourceSloValidation/time_window_invalid_duration_rolling/test.tf` (new)
- `internal/kibana/slo/testdata/TestAccResourceSloValidation/time_window_invalid_duration_calendar/test.tf` (new)
- `internal/kibana/slo/acc_test.go` — two new plan-only validation test steps
- `docs/resources/kibana_slo.md` — regenerated
