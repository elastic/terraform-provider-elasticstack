## Why

Practitioners cannot manage **APM source maps** as code today ([#677](https://github.com/elastic/terraform-provider-elasticstack/issues/677)). Kibana exposes `POST /api/apm/sourcemaps`, `GET /api/apm/sourcemaps`, and `DELETE /api/apm/sourcemaps/{id}` endpoints for uploading and deleting source maps used by APM for un-minifying JavaScript stack traces. The provider's generated `kbapi` already models `APMUIUploadSourceMapObject`, `APMUIUploadSourceMapsResponse`, and the endpoint handlers; no Terraform resource exists yet.

## What Changes

- Add a new resource `elasticstack_apm_source_map` that creates (uploads) and deletes APM source maps via the Kibana API.
- Support both JSON string source maps (`sourcemap_json`) and binary/encoded source maps (`sourcemap_binary`).
- Capture the Fleet artifact `id` returned by the upload response for state tracking and deletion.
- All write attributes are `RequireReplace` because the API has no update endpoint.

### Schema sketch

```hcl
resource "elasticstack_apm_source_map" "example" {
  id               = <computed, string>   # Fleet artifact ID returned by upload
  bundle_filepath  = <required, string>   # Absolute path of the final bundle in the web application
  service_name     = <required, string>   # Service name the source map applies to
  service_version  = <required, string>   # Service version the source map applies to
  sourcemap_json   = <optional, string>   # Source map as a JSON string; mutually exclusive with sourcemap_binary
  sourcemap_binary = <optional, string>   # Source map as a base64-encoded string; mutually exclusive with sourcemap_json
  space_id         = <optional, string>   # Kibana space ID; omit or set to "default" for the default space
  kibana_connection = <optional, block>   # Entity-local Kibana connection override
}
```

Exactly one of `sourcemap_json` or `sourcemap_binary` must be set. Both are write-only (sensitive content) and are not read back from the API; `id`, `bundle_filepath`, `service_name`, and `service_version` can be recovered during read/import from the source-map artifact metadata.

### CRUD semantics

| Operation | Behavior |
|-----------|----------|
| Create    | `POST /api/apm/sourcemaps` (multipart/form-data) via `kibanautil.BuildSpaceAwarePath`; captures `id` from `APMUIUploadSourceMapsResponse` |
| Read      | `GET /api/apm/sourcemaps` via `kibanautil.BuildSpaceAwarePath`; find artifact by state `id`; repopulate `bundle_filepath`, `service_name`, and `service_version` from the artifact body; remove from state if not found (404 equivalent) |
| Update    | Not supported — all write attributes have `RequireReplace`; Terraform destroys and recreates |
| Delete    | `DELETE /api/apm/sourcemaps/{id}` via `kibanautil.BuildSpaceAwarePath` |

## Capabilities

### New Capabilities

- `apm-source-map`: New resource `elasticstack_apm_source_map` for creating, reading, and deleting APM source maps via the Kibana APM source maps API.

### Modified Capabilities

_(none)_

## Impact

- **New spec**: `openspec/changes/apm-source-map/specs/apm-source-map/spec.md`
- **Implementation** (future): New package `internal/apm/source_map/` following the `internal/apm/agent_configuration/` pattern; register resource in provider; add acceptance tests; add docs.
