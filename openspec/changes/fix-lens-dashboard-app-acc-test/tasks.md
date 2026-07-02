# Tasks: Drop `lens-dashboard-app` Acceptance Test Failing on Kibana 9.5

## 1. Test removal

- [ ] 1.1 Remove the `TestAccResourceDashboardUnknownPanel_lensDashboardApp` function and the
       `replaceDashboardPanelWithLensDashboardApp` helper from
       `internal/kibana/dashboard/acc_unknown_panels_test.go`.
- [ ] 1.2 Remove any imports that become unused after the deletion (e.g. `bytes`, `io`, `net/http`,
       `context`, `encoding/json`, `fmt` if no longer referenced by other tests in the same file).
- [ ] 1.3 If `acc_unknown_panels_test.go` becomes empty after removal, delete the file entirely.

## 2. Verify existing unit coverage

- [ ] 2.1 Confirm that `Test_unknownPanelRoundTrip`, the `"unknown panel type preserves id, grid,
       and config"` case in `Test_mapPanelsFromAPI`, and the `"unknown panel type replays config_json"`
       case in `Test_panelsToAPI` are all present and passing in
       `internal/kibana/dashboard/models_panels_test.go` (run `go test ./internal/kibana/dashboard/...`
       without `TF_ACC`).

## 3. Spec sync

- [ ] 3.1 Keep the delta spec under `openspec/changes/fix-lens-dashboard-app-acc-test/specs/`
       aligned with `proposal.md` and `design.md`.
