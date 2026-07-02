# Design: Drop `lens-dashboard-app` Acceptance Test Failing on Kibana 9.5

## Context

`TestAccResourceDashboardUnknownPanel_lensDashboardApp` in
`internal/kibana/dashboard/acc_unknown_panels_test.go` is an acceptance test that verifies the
provider's unknown-panel preservation path. It does so by:

1. Creating a minimal dashboard via Terraform.
2. Replacing the dashboard's panels via a direct Kibana REST API PUT with a `lens-dashboard-app`
   typed panel (a panel type that the provider does not model, so it would be treated as unknown).
3. Running a `terraform plan` and asserting that the provider reads back the unknown panel and
   populates `config_json` without erroring.

On Kibana 9.5.0-SNAPSHOT, the PUT in step 2 is rejected with HTTP 400 because Kibana now validates
the panel `type` against an explicit allow-list that excludes `lens-dashboard-app`. The test fails
at the PreConfig step before the provider is involved at all.

## Goals

1. Remove the acceptance test that can no longer be executed against Kibana 9.5+.
2. Confirm that the unknown-panel preservation contract is already unit-tested with a stable
   arbitrary foreign type.
3. Produce an accurate delta spec entry documenting the test removal and the coverage that remains.

## Non-Goals

- Replacing the acceptance test with a new one using a different `type` value. Any such choice
  would be inherently unstable: Kibana either models the type (so it is not "unknown") or validates
  its internal config structure strictly, making it liable to break in the same way.
- Modifying the provider's read or write logic for unknown panels.
- Adding a version gate to run the acceptance test only on Kibana versions `< 9.5`. The correct fix
  is not to pin the test to older versions; the correct fix is to acknowledge that no stable
  accepted-but-unmodeled panel type exists and rely on unit coverage instead.
- Any changes to the existing `lens-dashboard-app` typed panel support added by the
  `lens-dashboard-app-panel` change (now archived).

## Decisions

### Remove the acceptance test; rely on existing unit tests

The unknown-panel preservation contract is exercised by three existing unit test cases in
`internal/kibana/dashboard/models_panels_test.go`:

| Test | Coverage |
|---|---|
| `Test_unknownPanelRoundTrip` | Full fromAPI → toAPI round-trip with an arbitrary foreign `type`. Asserts that the JSON output is semantically identical to the input. |
| `Test_mapPanelsFromAPI` / `"unknown panel type preserves id, grid, and config"` | The fromAPI path preserves `id`, `grid`, and all `config` fields for an unrecognized type. |
| `Test_panelsToAPI` / `"unknown panel type replays config_json"` | The toAPI path replays `config_json` verbatim for an unrecognized type. |

These tests use `custom_unknown_panel` as the foreign type, which is stable because it is simply
not a Kibana panel type at all and will never be added to any allow-list. The coverage they
provide is equivalent to what the removed acceptance test intended to prove at the integration
level: that the provider round-trips an unmodeled panel without modifying or erroring on it.

### No spec requirement changes needed

REQ-025 already captures the unknown-panel preservation contract:

> **Scenario: config_json preserved for unrecognized panel types at read time**
> - GIVEN a panel with an unknown or unrecognized `type` value (including a Kibana-internal type
>   such as `lens-dashboard-app`)
> - WHEN the provider reads such a panel back from the Kibana API
> - THEN the provider SHALL use the unknown-panel fallback and SHALL populate `config_json` in state
> - AND SHALL NOT return an error diagnostic for the unrecognized panel type

The delta spec for this change documents the removal of the acceptance test and cross-references
the unit test coverage, but does not add new requirements.

## Open Questions

None. The approach is straightforward and the coverage gap analysis is complete.
