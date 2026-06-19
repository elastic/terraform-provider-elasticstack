## Context

The Kibana Dashboard API has two create endpoints:

1. `POST /api/dashboards/dashboard` — server assigns a UUID; no `id` field in request body.
2. `PUT /api/dashboards/dashboard/{id}` — caller supplies `{id}`; upsert semantics; returns `201 Created` on new resource, `200 OK` on overwrite.

The generated kbapi client already exposes `PutDashboardsIdWithResponse`. The `kibanaoapi.UpdateDashboard` wrapper (lines 71–84 of `internal/clients/kibanaoapi/dashboards.go`) already accepts both `http.StatusOK` and `http.StatusCreated`, so PUT-for-create requires no kbapi layer changes.

## Goals / Non-Goals

**Goals:**
- Allow practitioners to set a stable `dashboard_id` at create time.
- Branch create logic on whether `dashboard_id` is known at plan time.
- Preserve full backward compatibility for configs that omit `dashboard_id`.

**Non-Goals:**
- Allow renaming `dashboard_id` in place — `RequiresReplace` handles that via destroy+create.
- Remove the POST path — backward compatibility requires preserving auto-assigned UUIDs.
- Data source changes — unaffected.
- Adding an explicit upsert function to `kibanaoapi` — `UpdateDashboard` already does the right thing.

## Decisions

- **Schema modifier**: `RequiresReplace()` on `dashboard_id` mirrors `space_id`. Rename-in-place is not supported; changing the id triggers destroy+create. This is the same behavior users expect from `space_id`.
- **API method selection at create time**: `planModel.DashboardID.IsKnown() && !planModel.DashboardID.IsNull()` determines whether to use PUT. This is consistent with how the entitycore framework propagates `WriteID` for known-at-plan-time values.
- **PUT response id extraction**: When `kibanaoapi.UpdateDashboard` returns `JSON201` (new dashboard created), `JSON201.Id` SHALL be used as `DashboardID`. When the API returns `JSON200` instead (pre-existing dashboard), `JSON200.Id` SHALL be used. This is an edge case — on a fresh create, the server SHALL always return `201` for a new id — but defensive handling prevents a nil-dereference if the server surprises us.
- **`dashboardToAPIUpdateRequest` vs `dashboardToAPICreateRequest`**: The two request body types (`PutDashboardsIdJSONRequestBody` vs `PostDashboardsJSONRequestBody`) differ only in generated name; the field shapes are the same. Reuse `dashboardToAPIUpdateRequest` for the PUT path in `createDashboard`.

## Open questions

- **API version gate**: Is `PUT /api/dashboards/dashboard/{id}` available from the same 9.4.x minimum as `POST`, or does it require a higher version? Should be confirmed against the dashboards-api-spec changelog before setting the version guard.
- **Import round-trip**: If a practitioner imports an existing dashboard with a human-readable ID while `dashboard_id` is set in config, does `GetDashboard` + composite-ID construction round-trip cleanly? No blocker expected, but worth a quick verification.
- **`JSON201.Id` identity**: When PUT returns `201`, is `JSON201.Id` always the caller-supplied `{id}` path param (no server-side remapping)? Should be asserted in the acceptance test.

## Risks / Trade-offs

- [Low] PUT upsert silently recreates an out-of-band-deleted dashboard on re-apply (generally desirable, but different from the POST path which errors on double-create). This is the intended behavior for upsert and should be documented in the attribute description.
- [Low] If the user supplies a `dashboard_id` that already exists in Kibana (and is not tracked in Terraform state), PUT will overwrite it silently. This is the standard upsert risk and is consistent with existing Kibana API semantics.
