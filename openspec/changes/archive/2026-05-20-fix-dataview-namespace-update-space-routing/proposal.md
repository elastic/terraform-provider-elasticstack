## Why

`elasticstack_kibana_data_view` supports updating `data_view.namespaces` in-place via Kibana's `POST /api/spaces/_update_objects_spaces` endpoint. The implementation in `internal/clients/kibanaoapi/data_views_spaces.go` (introduced in v0.14.4, PR #2129) has two bugs that cause namespace updates to silently fail whenever the data view lives in a non-default Kibana space:

1. **Missing space context**: `UpdateDataViewNamespaces` calls `PostSpacesUpdateObjectsSpacesWithResponse` without a `SpaceAwarePathRequestEditor`, so the request resolves to `/api/spaces/_update_objects_spaces` (default-space context). Kibana returns HTTP 200 with a per-object 404 error in the body because the saved object cannot be found in the default space.
2. **Silent failure**: the provider only checks the HTTP status code (200). The per-object error embedded in the response body is never inspected, so the failed update is reported as success. Terraform writes "Update complete" and the data view is never actually shared to the target spaces.

The existing acceptance test (`TestAccResourceDataViewNamespaces`) passes because the data view's `space_id` is always `"default"`, which does not require the `/s/{spaceId}/` URL prefix.

## What Changes

- **API call routing** (`data_views_spaces.go`): add `spaceID string` parameter to `UpdateDataViewNamespaces`; pass `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as the request editor so the call hits the correct space-scoped URL.
- **Per-object error checking** (`data_views_spaces.go`): parse the `JSON200` response body from `PostSpacesUpdateObjectsSpacesWithResponse` and surface any per-object errors as Terraform error diagnostics.
- **Call site update** (`update.go`): pass `spaceID` to `UpdateDataViewNamespaces`.
- **Acceptance test coverage** (`acc_test.go` / test configs): add a test step (or a new test) that places the data view in a non-default space and exercises a namespace update, verifying that namespaces are actually written rather than silently dropped.
- **Requirements update** (`openspec/specs/kibana-data-view/spec.md`): tighten REQ-002 (error surfacing) and REQ-009 (namespace reconciliation) to explicitly cover space-aware routing and per-object error detection.

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- `kibana-data-view`:
  - **REQ-002** (API and client error surfacing): extend to require that per-object errors returned in the `_update_objects_spaces` HTTP 200 response body are surfaced as error diagnostics.
  - **REQ-009** (Update request mapping and namespace reconciliation): specify that the Spaces API call for namespace reconciliation SHALL use space-aware URL construction (via `SpaceAwarePathRequestEditor(spaceID)`) so the correct saved object is targeted regardless of which Kibana space the data view lives in.

## Impact

| File | Change |
|------|--------|
| `internal/clients/kibanaoapi/data_views_spaces.go` | Add `spaceID` parameter; add `SpaceAwarePathRequestEditor`; parse and surface per-object errors |
| `internal/kibana/dataview/update.go` | Pass `spaceID` to `UpdateDataViewNamespaces` |
| `internal/kibana/dataview/acc_test.go` | Add test step or new test covering non-default space namespace update |
| `internal/kibana/dataview/testdata/TestAccResourceDataViewNamespaces/` | Add or extend test configs to set a non-default `space_id` on the data view |
| `openspec/specs/kibana-data-view/spec.md` | Update REQ-002 and REQ-009 scenarios |
