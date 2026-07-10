## Why

Four APM service-map panel acceptance tests fail on the 9.5.0-SNAPSHOT CI matrix during import
verification because Kibana 9.5 now injects a server-side default `environment` field
(`"ENVIRONMENT_ALL"`) into the panel config when the practitioner did not configure it. After
import the stored state has 15 keys including `environment`, while the plan has 14 (no
`environment`), so `ImportStateVerify` reports a mismatch. The affected tests are:

- `TestAccDashboardPanelApmServiceMap_allFilters`
- `TestAccDashboardPanelApmServiceMap_noConfig`
- `TestAccDashboardPanelApmServiceMap_serviceGroupIdOnly`
- `TestAccDashboardPanelApmServiceMap_serviceNameOnly`

## What Changes

Apply a hybrid fix:

1. **Provider read-path suppression**: in `PopulateFromAPI` (and `apmServiceMapConfigFromAPIImport`
   for the import path), treat `environment = "ENVIRONMENT_ALL"` as the server default and suppress
   it back to `null` when the prior state did not configure `environment` explicitly. This is
   value/plan-aware (keyed off the prior state's `environment` field), so an explicit
   `environment = "ENVIRONMENT_ALL"` in the practitioner's config is honoured and
   round-trips as expected.

2. **Test tolerance**: update the four failing tests to tolerate the `environment` field. The
   primary fix is provider-side suppression; the test tolerance is a backstop for the import step
   where no prior state is available (import always falls back to the API value for initialization,
   so the post-suppression import must agree with the config's null `environment`).

## Capabilities

### Modified Capabilities

- `kibana-dashboard` — update the APM service-map panel read-path requirement (REQ-apm-env-default)
  to suppress the `ENVIRONMENT_ALL` server default from state when `environment` is unset in the
  practitioner's configuration.

## Impact

- `internal/kibana/dashboard/panel/apmservicemap/model.go` — extend
  `apmServiceMapPreserveNullIntentFromPrior` to suppress `environment` when prior `environment` is
  null/unknown and the API returns `"ENVIRONMENT_ALL"`; extend
  `apmServiceMapConfigFromAPIImport` to apply the same suppression on the import path.
- `internal/kibana/dashboard/panel/apmservicemap/acc_test.go` — update the four failing tests to
  tolerate `environment = "ENVIRONMENT_ALL"` if needed after the suppression fix.
- `internal/kibana/dashboard/panel/apmservicemap/model_test.go` — add unit tests covering the
  suppression logic for both the normal read path and the import path.
- `openspec/changes/apm-service-map-environment-default/specs/kibana-dashboard/spec.md` — delta
  spec documenting the new suppression behaviour.
