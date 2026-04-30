## Why

Issue [#2555](https://github.com/elastic/terraform-provider-elasticstack/issues/2555) flags duplicated Read/Delete preludes across the four Elasticsearch security resources (user, system_user, role, role_mapping): each handler repeats the same state-load → composite-ID parse → scoped-client resolution → API call. The data sources already have a generic envelope (`entitycore.NewElasticsearchDataSource[T]`); a sibling envelope for resources will collapse the boilerplate, give Read/Delete a single canonical implementation, and prevent diagnostic-handling drift between resources.

## What Changes

- Add `entitycore.NewElasticsearchResource[T]`: a generic constructor that returns a Plugin Framework resource owning Configure, Metadata, Schema (with `elasticsearch_connection` block injection), Read, and Delete.
- Add `entitycore.ElasticsearchResourceModel` type constraint requiring `GetID() types.String` and `GetElasticsearchConnection() types.List`.
- Concrete resources supply: a schema factory (without the connection block), a read callback `func(ctx, *clients.ElasticsearchScopedClient, resourceID string, state T) (T, bool, diag.Diagnostics)` (bool = found), and a delete callback `func(ctx, *clients.ElasticsearchScopedClient, resourceID string, state T) diag.Diagnostics`.
- Concrete resources continue to own Create and Update (their flows diverge — write-only passwords, server-version gating, JSON marshalling, post-write re-reads).
- Migrate four Elasticsearch security resources to the envelope: `internal/elasticsearch/security/{user,systemuser,role,rolemapping}`. Each Data struct gains `GetID()` and `GetElasticsearchConnection()` getters.
- system_user's delete callback is a no-op that returns nil diagnostics (system users aren't deletable). This is uniform with the required-callback contract.
- Schema factories are split out so the connection block is added by the envelope rather than declared in each resource's schema.

No external behavior, schema, generated docs, or acceptance test fixtures change. Existing resources that don't migrate (api_key, others outside security) are unaffected.

## Capabilities

### New Capabilities
- `entitycore-resource-envelope`: Generic resource envelope (sibling to `entitycore-datasource-envelope`) that owns Configure, Metadata, Schema-with-connection-injection, Read prelude with composite-ID parsing and not-found removal, and Delete prelude.

### Modified Capabilities
<!-- None. The four security resource specs describe externally-observable behavior, which is preserved verbatim. The provider-framework-entity-core spec describes the simple ResourceBase substrate, which is unchanged; the envelope is an additional, separately-specified capability. -->

## Impact

- New code: `internal/entitycore/resource_envelope.go` (and matching tests).
- Refactored code: `internal/elasticsearch/security/{user,systemuser,role,rolemapping}` — `resource.go`, `read.go`, `delete.go`, `schema.go`, `models.go`. `create.go` and `update.go` keep their existing logic (call sites for `r.Client()` continue to work via the embedded envelope).
- Specs: new `openspec/specs/entitycore-resource-envelope/spec.md` (created on archive of this change).
- No public API changes. No provider config changes. No Terraform schema changes.
- Acceptance tests for the four resources must continue to pass without modification.
