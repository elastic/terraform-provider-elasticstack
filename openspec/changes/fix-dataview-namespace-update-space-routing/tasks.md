## 1. Fix `UpdateDataViewNamespaces` — space-aware routing

- [ ] 1.1 In `internal/clients/kibanaoapi/data_views_spaces.go`, add `spaceID string` as the third parameter (before `dataViewID`) to `UpdateDataViewNamespaces`. Update all imports: add `"github.com/elastic/terraform-provider-elasticstack/internal/clients/kibanautil"` if not already present.
- [ ] 1.2 Pass `kibanautil.SpaceAwarePathRequestEditor(spaceID)` as the final variadic argument to `PostSpacesUpdateObjectsSpacesWithResponse`, replacing the current call that passes no editor.

## 2. Fix `UpdateDataViewNamespaces` — per-object error surfacing

- [ ] 2.1 After the HTTP status check, verify that `resp.JSON200` is non-nil. If it is nil (unexpected for HTTP 200), add an error diagnostic and return.
- [ ] 2.2 Iterate `resp.JSON200.Objects`. For each element where the `Error` field is non-nil, add an error diagnostic that includes the object `Id`, `Type`, and the error `Message` (or `statusCode`/`error` sub-fields) from the per-object error struct. Return after appending all per-object errors so the caller sees every failed object in one apply.

## 3. Update call site in `update.go`

- [ ] 3.1 In `internal/kibana/dataview/update.go`, line ~97, update the call from:
  ```go
  kibanaoapi.UpdateDataViewNamespaces(ctx, oapiClient, viewID, oldNS, newNS)
  ```
  to:
  ```go
  kibanaoapi.UpdateDataViewNamespaces(ctx, oapiClient, spaceID, viewID, oldNS, newNS)
  ```
  `spaceID` is already in scope from line 59 (`viewID, spaceID := planModel.getViewIDAndSpaceID()`).

## 4. Acceptance test — non-default space namespace update

- [ ] 4.1 In `internal/kibana/dataview/testdata/TestAccResourceDataViewNamespaces/initial/data_view.tf`, add `space_id = var.space1` to the `elasticstack_kibana_data_view.ns_dv` resource so the data view lives in a non-default space. Adjust the initial `namespaces` list to include `var.space1` explicitly (the current `space_id`), `var.space2`, and `"default"` (or another target space).
- [ ] 4.2 Update the corresponding `add_space/data_view.tf`, `remove_space/data_view.tf`, and `add_remove_space/data_view.tf` configs to carry the same `space_id = var.space1` so the test steps remain consistent.
- [ ] 4.3 In `internal/kibana/dataview/acc_test.go`, add or update `TestAccResourceDataViewNamespaces` to assert the correct `namespaces.#` count after each step; also add a `terraform plan` empty-diff step after the final namespace update to confirm that Terraform sees no drift (meaning the namespaces were actually written and the read-back matches).

## 5. Build and lint

- [ ] 5.1 Run `make build` to confirm the provider compiles after the signature change and call-site update.
- [ ] 5.2 Run `make check-lint` to confirm no lint regressions.

## 6. Unit tests

- [ ] 6.1 Run `go test ./internal/clients/kibanaoapi/... ./internal/kibana/dataview/...` to confirm existing unit tests pass.

## 7. Requirements update

- [ ] 7.1 Update `openspec/specs/kibana-data-view/spec.md` to:
  - Extend **REQ-002** with a new scenario: "Namespace update per-object error surfaced" — when `_update_objects_spaces` returns HTTP 200 with a per-object error in the body, the provider SHALL surface an error diagnostic.
  - Extend **REQ-009** with an explicit requirement that the Spaces API call for namespace reconciliation SHALL use space-aware URL construction (the call SHALL include the resource's `space_id` in the URL path via `SpaceAwarePathRequestEditor`) so that the correct saved object is targeted in non-default spaces.

## 8. OpenSpec validation

- [ ] 8.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-dataview-namespace-update-space-routing --type change` and fix any reported issues until the command exits successfully.
