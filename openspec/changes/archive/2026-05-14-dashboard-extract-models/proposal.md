## Why

The `internal/kibana/dashboard` package has grown to ~44,000 lines across ~150 source files. Every Terraform model struct with `tfsdk:` tags lives in `package dashboard`, coupled to API conversion logic, schema definition, validation, state alignment, and resource lifecycle code. This monolithic structure makes parallel development difficult, forces reviewers to load the entire package mental model, and complicates adding new panel types.

Before any structural refactoring of panel logic is possible, the data layer must be decoupled from the logic layer. Terraform Plugin Framework populates models via reflection on `tfsdk` struct tags, so all tagged structs must live in a single Go package. Extracting them into `dashboard/models` creates a clean foundation with zero import cycles, enabling strict package separation for panel handlers in subsequent changes.

## What Changes

Extract every Terraform model struct from `internal/kibana/dashboard` into a new `internal/kibana/dashboard/models` package. Rename unexported model types to exported (`panelModel` → `PanelModel`, `dashboardModel` → `DashboardModel`, etc.). Update all references across the dashboard package. No logic changes. No behavior changes. No schema changes.

### Scope

- New package: `internal/kibana/dashboard/models`
- Files to create: `panel.go`, `dashboard.go`, `slo_burn_rate.go`, `markdown.go`, `xy_chart.go`, `lens.go`, and all remaining `*ConfigModel` structs
- ~50 files in `dashboard/` updated with renamed type references
- Zero user-visible changes

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None.

## Impact

### Source files

- `internal/kibana/dashboard/models/` — new directory, all model structs moved here
- Every existing `models_*.go` in `dashboard/` — model structs removed, conversion logic temporarily updated to reference `models.PanelModel`, etc.
- `create.go`, `read.go`, `update.go`, `delete.go`, `resource.go` — `dashboardModel` → `models.DashboardModel`
- `schema.go`, `panel_config_validator.go`, `panel_config_defaults.go` — no edits were required: these layers define or validate schema shape and do not reference `tfsdk` model structs directly (contrast with conversion code in `models_*.go`).

### Tests

- All existing unit and acceptance tests continue to pass with zero behavioral differences.

### Examples

None.

### Dependencies and sequencing

- **No dependencies on other active changes.** Must land before `dashboard-panel-contract`, `dashboard-lens-contract`, and `dashboard-composite-panel-contract`.
- Zero risk of merge conflicts with unreleased panel additions because the change is purely mechanical renaming and file movement.
