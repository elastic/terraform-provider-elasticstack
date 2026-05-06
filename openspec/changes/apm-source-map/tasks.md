## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate apm-source-map --type change` (or `make check-openspec` after sync).
- [ ] 1.2 On completion of implementation, **sync** delta into `openspec/specs/apm-source-map/spec.md` or **archive** the change per project workflow.

## 2. Implementation

- [ ] 2.1 Create package `internal/apm/source_map/` with `resource.go`, `schema.go`, `create.go`, `read.go`, `delete.go`, and `models.go` following the `internal/apm/agent_configuration/` structure.
- [ ] 2.2 Implement `schema.go`: define `elasticstack_apm_source_map` schema with `id` (computed), `bundle_filepath`, `service_name`, `service_version` (all required, `RequireReplace`), `sourcemap_json` (optional, sensitive, `RequireReplace`), `sourcemap_binary` (optional, sensitive, `RequireReplace`), `space_id` (optional, `RequireReplace`), and `kibana_connection` block. Add `ExactlyOneOf` validator enforcing exactly one of `sourcemap_json` / `sourcemap_binary` is set.
- [ ] 2.3 Implement `create.go`: read plan attributes; construct multipart form body with `bundle_filepath`, `service_name`, `service_version`, and the decoded source map content as the `sourcemap` file field; construct space-aware path via `kibanautil.BuildSpaceAwarePath(spaceID, "/api/apm/sourcemaps")`; call `UploadSourceMapWithBodyWithResponse`; capture `id` from `APMUIUploadSourceMapsResponse`; call `read.go` to populate state.
- [ ] 2.4 Implement `read.go`: read `space_id` from state to construct the space-aware path via `kibanautil.BuildSpaceAwarePath`; call `GetSourceMapsWithResponse`; paginate through all pages; locate artifact by `id` from state; if not found, remove resource from state; if found, set `id`, `bundle_filepath`, `service_name`, and `service_version` from the matching artifact body; preserve `space_id` from state (the API does not return space metadata); do not attempt to repopulate `sourcemap_json` or `sourcemap_binary` as the API does not return source map content.
- [ ] 2.5 Implement `delete.go`: construct space-aware path via `kibanautil.BuildSpaceAwarePath`; call `DeleteSourceMapWithResponse` with the state `id`.
- [ ] 2.6 Register `elasticstack_apm_source_map` resource in the provider's resource list (alongside other APM resources).
- [ ] 2.7 Add embedded descriptions / `MarkdownDescription` strings for all schema attributes.

## 3. Testing

- [ ] 3.1 Add acceptance test `TestAccResourceApmSourceMap_json` in `internal/apm/source_map/acc_test.go`: create a source map using `sourcemap_json`; assert `id` is set and non-empty; delete and confirm state removed.
- [ ] 3.2 Add acceptance test `TestAccResourceApmSourceMap_binary` in `internal/apm/source_map/acc_test.go`: create a source map using `sourcemap_binary` (base64-encoded minimal source map); assert `id` is set and non-empty.
- [ ] 3.3 Add acceptance test for import (`TestAccResourceApmSourceMap_import`): create a source map in a named space; import using the composite ID `"<space_id>/<fleet_artifact_id>"`; assert that `space_id`, `id`, `bundle_filepath`, `service_name`, and `service_version` are correctly set in state after import; also assert that a plain (no-slash) import ID correctly defaults to the default space.
- [ ] 3.4 Add acceptance test for space-aware operations (`TestAccResourceApmSourceMap_space`): create a source map with a non-default `space_id`; assert the resource is created and readable within that space using `GET /s/{space_id}/api/apm/sourcemaps`; confirm deletion removes the artifact from that space. Mirror the pattern in `internal/fleet/proxy/acc_test.go` for non-default-space CRUD verification.
- [ ] 3.5 Add acceptance test for `ExactlyOneOf` validation (`TestAccResourceApmSourceMap_validationNeitherSet` and `TestAccResourceApmSourceMap_validationBothSet`): use `ExpectError` to assert that applying a configuration with neither or both of `sourcemap_json`/`sourcemap_binary` returns a validation diagnostic (REQ-007).
- [ ] 3.6 Add acceptance test for `RequireReplace` semantics (`TestAccResourceApmSourceMap_requireReplace`): verify that changing `service_version` (or another write attribute) produces a `plancheck.ResourceActionReplace` action rather than an in-place update (REQ-008).
- [ ] 3.7 Add unit tests for the multipart form construction helper (source map content encoding, field names, boundary).
- [ ] 3.8 Add unit tests for the read pagination loop (mock: id found on page 1; id found on page N; id not found → resource removed from state).

## 4. Documentation

- [ ] 4.1 Generate provider docs via `make docs` (or equivalent) after implementation; verify the resource page renders `bundle_filepath`, `service_name`, `service_version`, `sourcemap_json`, `sourcemap_binary`, `space_id` with descriptions and marks `sourcemap_json` / `sourcemap_binary` as sensitive.
- [ ] 4.2 Add example Terraform configuration under `examples/resources/elasticstack_apm_source_map/` showing both `sourcemap_json` usage and a `space_id` usage example.

## 5. Verification

- [ ] 5.1 `make build` passes.
- [ ] 5.2 `make lint` passes.
- [ ] 5.3 `make check-openspec` passes with delta spec in this change.
