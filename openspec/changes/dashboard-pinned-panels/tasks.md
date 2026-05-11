## 1. Schema reuse

- [x] 1.1 Extract the four `*_control_config` nested attribute schemas into a shared builder if not already shared between `panels[]` and standalone usage
- [x] 1.2 Add `pinned_panels` block list to the dashboard root schema reusing those builders, omitting `grid`
- [x] 1.3 Add description text to `internal/kibana/dashboard/descriptions/`

## 2. Validators

- [x] 2.1 Reuse the existing "exactly one `*_control_config` matching `type`" conditional validators from `panels[]`
- [x] 2.2 Ensure validators surface useful messages distinguishing pinned vs in-grid placement

## 3. Model and mapping

- [x] 3.1 Add `PinnedPanels` to `dashboardModel`
- [x] 3.2 Map `PinnedPanels` into the API request body on create and update; reuse the per-control write helpers
- [x] 3.3 Map API response `pinned_panels` back into state, preserving order and the unset-vs-empty distinction

## 4. Tests

- [x] 4.1 Unit tests for discriminator validation (matching/mismatching type, multiple blocks, no blocks)
- [x] 4.2 Unit test for unset-vs-empty preservation
- [x] 4.3 Acceptance test creating a dashboard with at least one pinned options-list control and one pinned range-slider control
- [x] 4.4 Run `make build`, `go vet ./...`, `go test ./internal/kibana/dashboard/...` (`TF_ACC=1` for acceptance)

## 5. Spec sync

- [x] 5.1 Run `make check-openspec`
