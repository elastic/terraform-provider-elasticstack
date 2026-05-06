## Context

The Kibana APM source map API exposes three operations on `/api/apm/sourcemaps`:

- `POST` — upload via multipart/form-data; returns `APMUIUploadSourceMapsResponse` (Fleet artifact with `id`, `identifier`, checksums, timestamps, `relative_url`).
- `GET` — list all source maps as `APMUISourceMapsResponse` (paginated list of Fleet artifacts).
- `DELETE /{id}` — delete by Fleet artifact `id`.

The generated Kibana client (`generated/kbapi/kibana.gen.go`) provides:
- `UploadSourceMapWithBodyWithResponse` / `UploadSourceMapMultipartRequestBody = APMUIUploadSourceMapObject`
- `GetSourceMapsWithResponse` / `GetSourceMapsParams`
- `DeleteSourceMapWithResponse` / `DeleteSourceMapParams`

Request model (`APMUIUploadSourceMapObject`):
- `bundle_filepath` string (required) — absolute path of the final bundle in the web application
- `service_name` string (required) — service name
- `service_version` string (required) — service version
- `sourcemap` openapi_types.File (required) — multipart file upload; must be valid source map per the [spec](https://tc39.es/ecma426/)

Upload response model (`APMUIUploadSourceMapsResponse`):
- `id` *string — Fleet artifact identifier (used for state tracking and deletion)
- `identifier` *string — human-readable identifier (`{service_name}-{service_version}`)
- `body` *string, `compressionAlgorithm` *string, `created` *string, `decodedSha256` *string, `decodedSize` *float32, `encodedSha256` *string, `encodedSize` *float32, `encryptionAlgorithm` *string, `packageName` *string, `relativeUrl` *string, `type` *string

The API does not expose an update endpoint; any change to source map content or metadata requires deleting and re-uploading.

## Goals

- Expose `elasticstack_apm_source_map` as a new Plugin Framework resource.
- Support both JSON string input (`sourcemap_json`) and base64-encoded binary input (`sourcemap_binary`) for the source map content.
- Capture the Fleet artifact `id` for state tracking across plan/apply cycles.
- Keep Read robust: scan the paginated list and remove from state if the artifact is no longer found.

## Non-Goals

- Exposing read-only artifact metadata checksums and encoding fields (`identifier`, `encodedSha256`, `relativeUrl`, `compressionAlgorithm`, etc.) in state — only `id`, `bundle_filepath`, `service_name`, and `service_version` are surfaced since those are sufficient for lifecycle management and importability.
- Supporting concurrent management of multiple source maps for the same service/version combination (the API permits it; Terraform state would be distinct resources).

## Decisions

| Topic | Decision |
|-------|----------|
| Source map input | Exactly one of `sourcemap_json` or `sourcemap_binary` required; validated at config time via `ExactlyOneOf`. Both use `sensitive = true` to keep content out of logs. |
| Binary encoding | `sourcemap_binary` accepts base64-encoded content (standard encoding); the implementation decodes it before constructing the multipart form body. This avoids embedding raw binary in Terraform config/state. |
| RequireReplace | All write attributes (`bundle_filepath`, `service_name`, `service_version`, `sourcemap_json`, `sourcemap_binary`, `space_id`) use `RequireReplace` plan modifier — no update path. |
| Space awareness | An optional `space_id` attribute is supported. All API paths (`POST /api/apm/sourcemaps`, `GET /api/apm/sourcemaps`, `DELETE /api/apm/sourcemaps/{id}`) are constructed via `kibanautil.BuildSpaceAwarePath(spaceID, basePath)`. When `space_id` is empty or `"default"`, the path is unchanged (default space). |
| Read loop | `GET /api/apm/sourcemaps` returns a paginated list. The resource reads `space_id` from state to build the space-aware path, then iterates all pages (using `page`/`perPage` parameters) until the artifact whose `id` matches state is found or all pages exhausted. If found, `id`, `bundle_filepath`, `service_name`, and `service_version` are refreshed from the artifact body; `space_id` is preserved from state because the API does not return space metadata. If not found, the resource removes itself from state. |
| Import | Support import via a space-aware identifier. Accept `<space_id>/<id>` and populate both `space_id` and `id` in state; also accept bare `<id>` as an alias for the default space so existing/default-space imports remain simple. Passthrough import on `id` alone is not sufficient for non-default spaces because subsequent reads are explicitly space-aware. |
| Kibana client | Follow `internal/apm/agent_configuration` pattern: optional `kibana_connection` block; resolve `*clients.KibanaScopedClient` from provider; use `GetKibanaOapiClient()` for all operations; API version `2023-10-31`. |
| Package location | `internal/apm/source_map/` — new sub-package alongside `agent_configuration`. |

## Risks / Trade-offs

- **Multipart upload complexity**: The generated client's `UploadSourceMapWithBody` takes an `io.Reader` for the multipart body. The implementation must construct the multipart form manually (or use a helper) to combine `bundle_filepath`, `service_name`, `service_version`, and the `sourcemap` file field. This is the same pattern used elsewhere in multipart Kibana clients; no unique risk.
- **Read pagination**: A deployment with many source maps requires iterating pages. The implementation uses a page size of 100 and stops when the id is found or the returned page is smaller than `perPage` (last page). For typical deployments this is a single request.
- **`sourcemap_binary` content drift**: Since `sourcemap_binary` is write-only/sensitive, Terraform cannot detect drift in the uploaded source map content (the API doesn't return the original content). This is expected behavior for sensitive upload resources and should be documented.

## Open Questions

- None. The API model is fully defined in the generated client. Assumptions about `sourcemap_binary` as base64 are documented and consistent with how other providers handle binary file uploads in Terraform.
