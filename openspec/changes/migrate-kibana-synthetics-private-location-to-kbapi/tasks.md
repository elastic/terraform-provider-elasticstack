## 1. `kibanaoapi` private location helpers

- [ ] 1.1 Add a new source file under `internal/clients/kibanaoapi/` implementing `CreatePrivateLocation`, `GetPrivateLocation`, and `DeletePrivateLocation` (or equivalent names consistent with sibling helpers) using `client.API.PostPrivateLocationWithResponse`, `GetPrivateLocationWithResponse`, and `DeletePrivateLocationWithResponse`.
- [ ] 1.2 Apply a shared `kbapi.RequestEditorFn` that prefixes `/s/<space_id>` to the request path when the effective Kibana space is non-default (and treats `default` like the default space), matching existing legacy URL behavior.
- [ ] 1.3 Map HTTP failures to `diag.Diagnostics` using patterns from nearby `kibanaoapi` packages (status code, body snippet); treat GET 404 as a sentinel the resource layer can convert to state removal.

## 2. Wire model and schema away from `go-kibana-rest`

- [ ] 2.1 Replace `kbapi.PrivateLocationConfig`, `kbapi.PrivateLocation`, and `kbapi.SyntheticGeoConfig` usages in `internal/kibana/synthetics/privatelocation/schema.go` with `kbapi.PostPrivateLocationJSONBody` / `kbapi.SyntheticsGetPrivateLocation` (and nested geo structs) or small local DTOs that convert cleanly at the boundary.
- [ ] 2.2 Implement a single `privateLocationFromAPI(loc kbapi.SyntheticsGetPrivateLocation, spaceID string, kibanaConnection types.List) tfModelV0` (or equivalent) that reads `AdditionalProperties["tags"]` when needed.
- [ ] 2.3 Implement `privateLocationToCreateBody(m tfModelV0) kbapi.PostPrivateLocationJSONRequestBody` including optional `Tags` as `*[]string` and optional `Geo` with explicit float32 conversion.

## 3. CRUD migration

- [ ] 3.1 Update `create.go` to call `synthetics.GetKibanaOAPIClientFromScopedClient`, enforce existing `requiresSpaceIDMinVersion` + `EnforceMinVersion` logic unchanged, invoke the new helper, and set state from the normalized `SyntheticsGetPrivateLocation`.
- [ ] 3.2 Update `read.go` to use the OpenAPI client and helpers; preserve composite id parsing, `effectiveSpaceID`, 404 → `RemoveResource`, and version gating diagnostics.
- [ ] 3.3 Update `delete.go` similarly; preserve composite id handling and version gating.
- [ ] 3.4 Remove unused legacy client imports from the `privatelocation` package and ensure `resource.go` / `Update` messaging still matches REQ-006.

## 4. Tests and validation

- [ ] 4.1 Add unit tests for JSON → `SyntheticsGetPrivateLocation` → Terraform model mapping, including a fixture with `tags` and `geo` (and, if applicable, `AdditionalProperties` for tags).
- [ ] 4.2 Add a focused test that POST success `Body` normalization into `SyntheticsGetPrivateLocation` succeeds for a representative create response payload.
- [ ] 4.3 Run `make build` and `go test` for affected packages; run `TestSyntheticPrivateLocationResource` (and import / non-default space variants) against a stack meeting existing version gates documented in `dev-docs/high-level/testing.md`.

## 5. Spec hygiene

- [ ] 5.1 After implementation, run `make check-openspec` (or `openspec validate` as prescribed by repo docs) and resolve any validation issues for this change directory.
