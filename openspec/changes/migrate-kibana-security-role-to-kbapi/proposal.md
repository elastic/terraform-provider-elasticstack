## Why

The `elasticstack_kibana_security_role` resource and data source still depend on `github.com/disaster37/go-kibana-rest` (`KibanaRoleManagement`) while the rest of the provider is standardizing on the generated Kibana OpenAPI client (`generated/kbapi`) plus thin `internal/clients/kibanaoapi` helpers. Migrating removes a duplicate HTTP stack, aligns error handling with other kbapi-backed entities, and makes future OpenAPI regeneration the single source of truth for request/response shapes.

## What Changes

- Add `internal/clients/kibanaoapi` helpers that wrap Kibana Security Role endpoints exposed in `generated/kbapi` (`GetSecurityRoleName`, `PutSecurityRoleName`, `DeleteSecurityRoleName`), including consistent decoding of JSON bodies, HTTP status handling (including not-found on read), and diagnostics on failure.
- Refactor `internal/kibana/role.go` (and the data source companion if separated) to build requests and interpret responses using `kbapi` types (for example `PutSecurityRoleNameJSONRequestBody`) instead of `kbapi.KibanaRole` from go-kibana-rest, while preserving all Terraform schema semantics.
- Preserve existing expand/flatten behavior for `elasticsearch`, `kibana`, `metadata`, and JSON `query` diff suppression; preserve version gates for `remote_indices` (≥ 8.10.0) and `description` (≥ 8.15.0).
- Add or extend verification so Elasticsearch and Kibana privilege mappings remain equivalent to today’s behavior (acceptance tests plus targeted parity checks where practical).
- Drop the security-role code path’s dependency on `KibanaRoleManagement` for create/read/update/delete once migration is complete (no **BREAKING** Terraform schema or identity changes intended).

## Capabilities

### New Capabilities

- None (all behavior remains under the existing `kibana-security-role` capability).

### Modified Capabilities

- `kibana-security-role`: Update normative requirements so role CRUD is specified in terms of the generated OpenAPI client and `kibanaoapi` helpers instead of `KibanaRoleManagement.*`, including create-only semantics via the `createOnly` query parameter and not-found handling for GET responses.

## Impact

- `internal/kibana/role.go`, related tests under `internal/kibana/`, and any shared helpers used only by this entity.
- New file(s) under `internal/clients/kibanaoapi/` for role operations.
- `go.mod` / imports: reduced use of go-kibana-rest for this flow once callers are switched.
- OpenSpec delta under `openspec/changes/migrate-kibana-security-role-to-kbapi/specs/kibana-security-role/spec.md` for archive-time merge into `openspec/specs/kibana-security-role/spec.md`.
