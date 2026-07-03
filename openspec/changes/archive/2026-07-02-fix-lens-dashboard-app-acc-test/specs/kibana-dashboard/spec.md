# Delta spec: Drop `lens-dashboard-app` acceptance test for Kibana 9.5

## Capability: `kibana-dashboard`

Kibana 9.5 introduced server-side validation of dashboard panel `type` values against an
allow-list. The panel type `lens-dashboard-app` is no longer accepted by Kibana 9.5+ PUT requests.
The acceptance test `TestAccResourceDashboardUnknownPanel_lensDashboardApp` relies on injecting a
`lens-dashboard-app` panel directly via the Kibana REST API; that injection now fails with HTTP 400
before the provider is involved.

The provider's unknown-panel preservation behavior is unchanged and correct: unrecognized panel
types are read from the Kibana API and stored verbatim in `config_json` per REQ-025. The failing
test exercised a code path that is already unit-tested with a stable arbitrary foreign type
(`custom_unknown_panel`).

## MODIFIED Requirements

### Requirement: Raw `config_json` panel behavior (REQ-025)

The provider SHALL preserve the unknown-panel round-trip behavior specified in REQ-025 on Kibana
9.5 and later. The acceptance test `TestAccResourceDashboardUnknownPanel_lensDashboardApp` and its
helper `replaceDashboardPanelWithLensDashboardApp` SHALL be removed from
`internal/kibana/dashboard/acc_unknown_panels_test.go` because the test fixture type
(`lens-dashboard-app`) is no longer accepted by the Kibana 9.5+ PUT API.

The provider SHALL continue to satisfy the unknown-panel preservation contract. The following unit
tests in `internal/kibana/dashboard/models_panels_test.go` SHALL remain as the primary test
coverage and are not modified by this change:

- `Test_unknownPanelRoundTrip`
- `Test_mapPanelsFromAPI` / `"unknown panel type preserves id, grid, and config"`
- `Test_panelsToAPI` / `"unknown panel type replays config_json"`

#### Scenario: config_json preserved for unrecognized panel types at read time

- GIVEN a panel with an unknown or unrecognized `type` value (e.g. `custom_unknown_panel`)
- WHEN the provider reads such a panel back from the Kibana API via `dashboardMapPanelsFromAPI`
- THEN the provider SHALL use the unknown-panel fallback and SHALL populate `config_json` in state with the verbatim API config
- AND SHALL NOT return an error diagnostic for the unrecognized panel type
- AND the round-trip through `dashboardPanelsToAPI` SHALL produce API JSON semantically identical to the original input

#### Scenario: Unknown panel type replays config_json on write

- GIVEN a panel model with an unknown `type` and a non-null `config_json`
- WHEN the provider serialises the panel to the API payload via `dashboardPanelsToAPI`
- THEN the API panel `config` SHALL contain the verbatim JSON from `config_json`
- AND the panel `type`, `id`, and `grid` SHALL be preserved unchanged
