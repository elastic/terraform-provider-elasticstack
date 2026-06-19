## Why

`elasticstack_kibana_dashboard` always creates dashboards via `POST /api/dashboards/dashboard`, which auto-assigns a UUID. The `dashboard_id` attribute is `Computed`-only with `UseStateForUnknown`, so practitioners cannot supply a stable, human-readable identifier at create time. The Kibana API already supports `PUT /api/dashboards/dashboard/{id}` with upsert semantics for create-with-specific-id; the provider just doesn't surface it.

Existing saved-object workflows often depend on stable IDs (e.g. `my-service-overview`) for cross-resource drilldown links, external URLs, and embedded iframes. When IDs are auto-generated UUIDs, these references break on every resource recreation.

## What Changes

- Change `dashboard_id` from `Computed / UseStateForUnknown` to `Optional + Computed / RequiresReplace`.
- In `createDashboard()`, branch on whether `dashboard_id` is known and non-null at plan time:
  - **Known at plan time** → call `PUT /api/dashboards/dashboard/{id}` (via existing `kibanaoapi.UpdateDashboard`) and extract the returned id from `JSON201.Id`.
  - **Unknown or null** → continue calling `POST /api/dashboards/dashboard` as today.
- `RequiresReplace()` on `dashboard_id` mirrors the existing `space_id` behavior; changing the id destroys and recreates.

```hcl
resource "elasticstack_kibana_dashboard" "overview" {
  space_id     = "default"
  dashboard_id = "my-team-overview"  # optional; when set, drives PUT on create
  title        = "Team Overview"
  description  = "High-level metrics"
  # ... other fields
}
```

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `kibana-dashboard`: extend REQ-003 (Composite identity and computed ids) to allow `dashboard_id` to be practitioner-supplied (Optional + Computed) and drive PUT-based upsert on create when provided.

## Impact

- `internal/kibana/dashboard/schema.go` — change `dashboard_id` attribute from `Computed / UseStateForUnknown` to `Optional + Computed / RequiresReplace`.
- `internal/kibana/dashboard/create.go` — branch in `createDashboard()` on plan-time `DashboardID`: call `kibanaoapi.UpdateDashboard` (PUT) when known and non-null, `kibanaoapi.CreateDashboard` (POST) otherwise.
- `internal/kibana/dashboard/models/` — ensure model marshalling handles `dashboard_id` as optional input to `dashboardToAPIUpdateRequest`.
- Acceptance tests — add test cases for both the auto-generated-id path and the user-supplied-id path.
- No changes to the kbapi client layer (`internal/clients/kibanaoapi/dashboards.go`): `UpdateDashboard` already accepts both `200 OK` and `201 Created` responses correctly.
- No data source changes: the data source looks up by ID and is unaffected.
