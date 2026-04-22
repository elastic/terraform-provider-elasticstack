## Why

The private location resource is one of the remaining Synthetics entities that still depends on `go-kibana-rest` / `kbapi.PrivateLocation*` types for CRUD. Moving it to the generated OpenAPI client (`generated/kbapi` via `kibanaoapi`) aligns it with other migrated Kibana resources, reduces dual-client drift, and makes wire types and responses traceable to the published Kibana OpenAPI spec.

## What Changes

- Add focused `kibanaoapi` helpers for Synthetics private locations (create, get, delete) that call `kbapi.ClientWithResponses` and apply the same space-scoped URL behavior the legacy client used for non-default spaces.
- Refactor `internal/kibana/synthetics/privatelocation` so schema/model mapping and CRUD use `generated/kbapi` request/response models instead of `github.com/disaster37/go-kibana-rest/v8/kbapi` private location structs.
- Preserve existing `space_id` version gating (`MinVersionSpaceID` / `requiresSpaceIDMinVersion`), composite import id handling, `RequiresReplace` / no in-place update semantics, and `kibana_connection` scoping.
- Add validation (unit tests and/or targeted checks) that private location payloads round-trip through the generated read model (`SyntheticsGetPrivateLocation`), including fields that may deserialize into `AdditionalProperties` when the OpenAPI struct is incomplete (notably `tags` if absent from first-class fields).

## Capabilities

### New Capabilities

_None — normative behavior stays under the existing `kibana-synthetics-private-location` capability; this change updates how the provider satisfies those requirements._

### Modified Capabilities

- `kibana-synthetics-private-location`: Replace legacy-client obligations with OpenAPI/`kbapi` obligations for private location CRUD while keeping observable Terraform behavior (import id format, replacement rules, 404 handling, `space_id` gating, connection scoping).

## Impact

- **Code:** `internal/kibana/synthetics/privatelocation` (`schema.go`, `create.go`, `read.go`, `delete.go`, tests), new or extended helpers under `internal/clients/kibanaoapi/`, and `internal/kibana/synthetics/api_client.go` usage patterns (switch from `GetKibanaClient` to `GetKibanaOAPIClient` for this resource).
- **Dependencies:** Less use of `go-kibana-rest` types in this package; continued use of `github.com/elastic/terraform-provider-elasticstack/generated/kbapi` types.
- **Specs:** Delta updates under `openspec/specs/kibana-synthetics-private-location` when the change is archived or synced; acceptance tests remain the regression surface for end-to-end behavior.
