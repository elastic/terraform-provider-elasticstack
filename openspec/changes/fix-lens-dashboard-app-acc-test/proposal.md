# Proposal: Drop `lens-dashboard-app` Acceptance Test Failing on Kibana 9.5

## Why

`TestAccResourceDashboardUnknownPanel_lensDashboardApp` fails on the 9.5.0-SNAPSHOT matrix.
Kibana 9.5 now validates dashboard panel `type` against a server-side allow-list that no longer
includes `lens-dashboard-app` (a legacy embeddable type for embedded dashboards). The test fixture
injects a `lens-dashboard-app` panel via the Kibana REST API and expects the provider to
round-trip it as an unknown panel; however, Kibana 9.5 rejects the PUT with HTTP 400, so the test
itself cannot exercise the preservation path.

The **provider's unknown-panel preservation behavior is correct**: unrecognized panel types are read
back from Kibana and stored in `config_json`, and the provider does not error on them. The test's
chosen fixture type is simply no longer a valid panel type on Kibana 9.5+, so the test can no longer
serve as an end-to-end exercise of the preservation contract using a real Kibana instance.

## What Changes

- **Remove** the `TestAccResourceDashboardUnknownPanel_lensDashboardApp` acceptance test from
  `internal/kibana/dashboard/acc_unknown_panels_test.go`.
- **Verify** that the remaining unit test coverage in
  `internal/kibana/dashboard/models_panels_test.go` adequately demonstrates the unknown-panel
  preservation contract without relying on a Kibana-accepted panel type.

The existing unit tests already cover the preservation contract with an arbitrary foreign type
(`custom_unknown_panel`):
- `Test_unknownPanelRoundTrip` — full fromAPI → toAPI round-trip asserting identical JSON output.
- `Test_mapPanelsFromAPI` case `"unknown panel type preserves id, grid, and config"` — fromAPI
  preserves all fields.
- `Test_panelsToAPI` case `"unknown panel type replays config_json"` — toAPI replays `config_json`
  verbatim.

No new unit tests need to be added because these cases already exercise the exact contract the
removed acceptance test was trying to prove at the integration level.

## Capabilities

### Modified Capabilities
- `kibana-dashboard` — test coverage: `TestAccResourceDashboardUnknownPanel_lensDashboardApp`
  acceptance test removed; unknown-panel preservation remains functionally correct and is covered
  by existing unit tests.

## Impact

- **Test-only change**: no provider logic, schema, or state is modified.
- **Removes a failing acceptance test** that cannot be fixed by updating the fixture (no stable
  Kibana-9.5-accepted panel type is unmodeled and suitable as a stable fallback fixture).
- **No new spec requirements**: the unknown-panel preservation contract is already captured in
  REQ-025 (`config_json preserved for unrecognized panel types at read time`).
- **No breaking change**: practitioners are unaffected.
