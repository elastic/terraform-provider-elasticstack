## Context

`elasticstack_kibana_data_view` reconciles `data_view.namespaces` in-place via Kibana's Spaces object-sharing API endpoint `POST /api/spaces/_update_objects_spaces`. The implementation (`internal/clients/kibanaoapi/data_views_spaces.go:UpdateDataViewNamespaces`) was introduced in v0.14.4 (PR #2129) and is called from `internal/kibana/dataview/update.go`.

Two defects exist:

**Defect 1 — Missing space context**: `PostSpacesUpdateObjectsSpacesWithResponse` is called with no request editor, so the HTTP request path resolves to `/api/spaces/_update_objects_spaces`. When the data view lives in a non-default space (e.g. `space_id = "test-space"`), the request reaches Kibana with no space prefix, Kibana looks for the saved object in the default space, does not find it, and returns HTTP 200 with a per-object error:

```json
{"objects":[{"id":"my-data-view","type":"index-pattern","error":{"statusCode":404,"error":"Not Found","message":"Saved object [index-pattern/my-data-view] not found"}}]}
```

The correct URL for a non-default space is `/s/test-space/api/spaces/_update_objects_spaces`. The `kibanautil.SpaceAwarePathRequestEditor` helper already implements this transformation for other endpoints (`UpdateFieldMetadata`, and endpoints in `dashboards.go`, `alerting_rule.go`, etc.).

**Defect 2 — Silent failure**: the provider only checks `resp.StatusCode() != http.StatusOK`. Because the endpoint always returns HTTP 200 (even when individual objects fail), per-object errors in `resp.JSON200.Objects[*].Error` are never inspected. Terraform records success and writes state, but the namespace change has not taken effect on the Kibana side.

The `spaceID` is already computed at the `Update` call site:

```go
viewID, spaceID := planModel.getViewIDAndSpaceID()          // update.go:59
kibanaoapi.UpdateDataViewNamespaces(ctx, oapiClient, viewID, oldNS, newNS)  // update.go:97 — spaceID missing
```

The fix is localized to two functions and their call site. No schema change is required.

## Goals / Non-Goals

**Goals:**
- Fix the silent failure by passing `spaceID` to `UpdateDataViewNamespaces` and using `SpaceAwarePathRequestEditor`.
- Surface per-object errors from the `_update_objects_spaces` response body as Terraform error diagnostics.
- Expand acceptance test coverage to include a non-default space namespace update.
- Ensure `TestAccResourceDataViewNamespaces` catches regressions for both default and non-default space scenarios.

**Non-Goals:**
- Changes to the data view schema.
- Changes to create-time namespace injection.
- Changes to read behavior.
- Changes to any other resource.

## Decisions

### Decision 1: Add `spaceID` parameter to `UpdateDataViewNamespaces` and apply `SpaceAwarePathRequestEditor`

The function signature changes from:

```go
func UpdateDataViewNamespaces(ctx context.Context, client *Client, dataViewID string, oldNamespaces []string, newNamespaces []string) diag.Diagnostics
```

to:

```go
func UpdateDataViewNamespaces(ctx context.Context, client *Client, spaceID string, dataViewID string, oldNamespaces []string, newNamespaces []string) diag.Diagnostics
```

The API call changes from:

```go
resp, err := client.API.PostSpacesUpdateObjectsSpacesWithResponse(ctx, reqBody)
```

to:

```go
resp, err := client.API.PostSpacesUpdateObjectsSpacesWithResponse(ctx, reqBody, kibanautil.SpaceAwarePathRequestEditor(spaceID))
```

This follows the exact pattern used by `UpdateFieldMetadata` in `data_views.go` and other kibanaoapi wrappers. `SpaceAwarePathRequestEditor` is a no-op when `spaceID` is `""` or `"default"`, so the default-space behavior is preserved without any conditional logic.

### Decision 2: Parse and surface per-object errors from the response body

After the HTTP status check, parse `resp.Body` and inspect the per-object results returned by Kibana. The generated type for this endpoint is currently `kbapi.PostSpacesUpdateObjectsSpacesResponse`, which exposes the raw `Body`/`HTTPResponse` rather than a typed `JSON200` field, so the implementation should not assume `resp.JSON200.Objects` exists.

In `internal/clients/kibanaoapi/data_views_spaces.go`, unmarshal `resp.Body` with `encoding/json` into a small local response struct that models the response shape needed here, for example an `Objects` slice whose elements contain `ID`, `Type`, and optional `Error` fields (`statusCode`, `error`, `message`). After decoding, iterate the parsed objects and collect any entries where `Error` is non-nil, surfacing each as a Terraform error diagnostic. If `json.Unmarshal` fails, or if the response is HTTP 200 but the body cannot be interpreted as the expected object-result payload, treat that as an unexpected response and return an error diagnostic.

If the OpenAPI spec/generator is updated later to emit a typed `JSON200` response for this endpoint, the implementation may use that generated type instead, but this change does not require generator/spec work.

### Decision 3: Pass `spaceID` from the `Update` call site

In `internal/kibana/dataview/update.go`, line 97 changes from:

```go
kibanaoapi.UpdateDataViewNamespaces(ctx, oapiClient, viewID, oldNS, newNS)
```

to:

```go
kibanaoapi.UpdateDataViewNamespaces(ctx, oapiClient, spaceID, viewID, oldNS, newNS)
```

`spaceID` is already in scope at that point (derived at line 59).

### Decision 4: Acceptance test coverage for non-default space

Extend `TestAccResourceDataViewNamespaces` (or add a parallel variant) to place the data view in a non-default space. Concretely, one of the existing test steps should set `space_id = var.space1` on the `elasticstack_kibana_data_view` resource and exercise an add-namespace update, then assert via `TestCheckResourceAttr` that the expected namespaces count is reflected in state AND verify via a follow-up API check or `terraform plan` empty-diff step that the namespace was actually applied server-side.

Because the existing test already creates `space1`, `space2`, `space3` as separate Kibana spaces, the test config change is confined to adding a `space_id` attribute on the data view and updating the initial `namespaces` list accordingly.

## Risks / Trade-offs

- **Signature change is internal**: `UpdateDataViewNamespaces` is not exported outside the provider; the single call site is `update.go`. No external callers affected.
- **Per-object error structure**: the generated client currently does not appear to model a typed JSON200 response body for this endpoint, so the implementer should plan to unmarshal `resp.Body` into a local struct (or update the generator/spec to emit a typed response) and verify the parsed field names against Kibana's raw API contract. The issue body shows the raw JSON shape, which is the authoritative contract.
- **`SpaceAwarePathRequestEditor` for spaces endpoint**: the endpoint path is `/api/spaces/_update_objects_spaces`. `BuildSpaceAwarePath` searches for `/api/` and prepends `/s/{spaceID}`, producing `/s/{spaceID}/api/spaces/_update_objects_spaces`. This is the correct Kibana URL per the issue's `curl` reproduction.
- **Acceptance test environment**: tests that create non-default spaces require the test Kibana user to have the `manage_spaces` privilege. Existing namespace tests already create spaces, so this prerequisite is met.

## Open Questions

*(none — the fix scope, root cause, and API contract are fully specified in the issue body, including working `curl` reproductions for both the broken and fixed paths)*
