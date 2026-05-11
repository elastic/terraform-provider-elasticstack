## 1. Schema

- [x] 1.1 Add `filters` block list to the dashboard root schema in `internal/kibana/dashboard/schema.go` with `filter_json` (required string) and the existing JSON normalization plan modifier
- [x] 1.2 Document the attribute (descriptions file under `internal/kibana/dashboard/descriptions/`)

## 2. Model and mapping

- [x] 2.1 Add `Filters` to `dashboardModel` in `models.go`
- [x] 2.2 Map `Filters` into the API request body on create and update
- [x] 2.3 Map API response `filters` back into state on read, preserving order and the unset-vs-empty distinction per REQ-009

## 3. Tests

- [x] 3.1 Unit tests in `models_dashboard_root_filters_test.go` for null-vs-empty preservation, normalization, and order
- [x] 3.2 Acceptance test creating a dashboard with multiple filters, verifying read-back diff is empty
- [x] 3.3 Run `make build`, `go vet ./...`, `go test ./internal/kibana/dashboard/...` (`TF_ACC=1` for acceptance)

## 4. Spec sync

- [x] 4.1 Run `make check-openspec`
