## Context

In Kibana 9.5.0-SNAPSHOT the Dashboard API began injecting `environment = "ENVIRONMENT_ALL"` into
APM service-map panel configs when the field is not supplied by the practitioner. The Terraform
provider's existing `apmServiceMapPreserveNullIntentFromPrior` function already suppresses fields
that were null in prior state — this mechanism covers the normal refresh cycle. However, the
**import path** (`apmServiceMapConfigFromAPIImport`) initialises the model directly from the API
without applying any suppression, so after an import the state can contain `environment =
"ENVIRONMENT_ALL"` even when no prior-state null intent exists. `ImportStateVerify` then compares
the post-import state with the pre-import plan-computed state, which has `environment = null`, and
reports a mismatch.

Relevant code locations:
- `internal/kibana/dashboard/panel/apmservicemap/model.go` — `PopulateFromAPI`,
  `apmServiceMapConfigFromAPIImport`, `apmServiceMapPreserveNullIntentFromPrior`.
- `internal/kibana/dashboard/panel/apmservicemap/model_test.go` — existing unit tests.
- `internal/kibana/dashboard/panel/apmservicemap/acc_test.go` — four failing acceptance tests.

## Goals / Non-Goals

**Goals:**
- Stop the four failing acceptance tests by making the import of a panel with no explicit
  `environment` produce state that matches the plan (`environment = null`).
- Preserve existing round-trip correctness when `environment` is explicitly set (any value,
  including `"ENVIRONMENT_ALL"`).
- Apply the same suppression pattern already established by `apmServiceMapPreserveNullIntentFromPrior`
  so the fix is consistent with the existing code style.

**Non-Goals:**
- Changing the Terraform schema for the `environment` attribute.
- Suppressing any other field injected by the server.
- Changing the API request payload (write path) in any way.
- Adding per-server-version branching; the fix is purely value-based.

## Decisions

**Suppression key**: `"ENVIRONMENT_ALL"` is the well-known server default string. The suppression
checks `value == "ENVIRONMENT_ALL"` and `prior.Environment` is null/unknown. This is value-based,
not version-based, so it also works on future stack versions if the default persists.

**Import path**: `apmServiceMapConfigFromAPIImport` must apply the same suppression. The import
path has no prior state, so the condition simplifies to: if the API returns `environment ==
"ENVIRONMENT_ALL"` and the field is not meaningful without an explicit user choice, set it to null
in state. The rationale is that `"ENVIRONMENT_ALL"` is a sentinel meaning "no environment filter"
— it is functionally equivalent to the field being absent, so storing it in state on import
causes spurious drift against configs that omit `environment`.

**Explicit `environment = "ENVIRONMENT_ALL"`**: when the practitioner explicitly sets
`environment = "ENVIRONMENT_ALL"` in their config, the prior state will have a known, non-null
`Environment` value. `apmServiceMapPreserveNullIntentFromPrior` already checks
`typeutils.IsKnown(prior.Environment)` before nulling the field, so an explicit value is
preserved. No special casing is needed.

**Test updates**: after the suppression fix, all four tests should pass. The test update is a
backstop in case any edge path still surfaces `environment` post-import; the tests should set
`environment = "ENVIRONMENT_ALL"` in the config for those scenarios or use
`ImportStateVerifyIgnore` scoped to `environment` — whichever is least invasive.

**Unit-test coverage**: add unit tests in `model_test.go` for:
- Normal read path: prior state `environment = null`, API returns `"ENVIRONMENT_ALL"` → state has
  `environment = null`.
- Normal read path: prior state `environment = "production"`, API returns `"ENVIRONMENT_ALL"` → state
  has `environment = "ENVIRONMENT_ALL"`.
- Import path: API returns `"ENVIRONMENT_ALL"` → state has `environment = null`.
- Import path: API returns `"production"` → state has `environment = "production"`.

## Risks / Trade-offs

- [Low risk] If a future Kibana version changes the default from `"ENVIRONMENT_ALL"` to another
  sentinel, the suppression won't fire. Mitigation: the fix is data-driven — changing the constant
  is a one-line change. Because the code already uses `apmServiceMapPreserveNullIntentFromPrior`,
  adding a constant is consistent with the surrounding style.
- [Low risk] Import has no prior plan to distinguish an omitted `environment` from an explicitly configured `environment = "ENVIRONMENT_ALL"`. This change chooses to suppress `"ENVIRONMENT_ALL"` to null on import to avoid spurious diffs against configurations that omit `environment`; practitioners who want to pin `"ENVIRONMENT_ALL"` explicitly can run `terraform apply` after import to converge state to the configured value.

## Open questions

None — the issue body fully specifies the root cause and the selected fix (Hybrid option).
