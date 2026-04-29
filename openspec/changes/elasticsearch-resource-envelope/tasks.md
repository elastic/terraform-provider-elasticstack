## 1. Envelope implementation

- [ ] 1.1 Add `internal/entitycore/resource_envelope.go` defining `ElasticsearchResourceModel` constraint (`GetID() types.String` + `GetElasticsearchConnection() types.List`), exported `ElasticsearchResource[T]` struct embedding `*ResourceBase`, and the constructor `NewElasticsearchResource[T](component Component, name string, schemaFactory func() rschema.Schema, readFunc, deleteFunc) *ElasticsearchResource[T]`.
- [ ] 1.2 Implement `Schema` on the envelope: copy the user-supplied blocks map and inject `elasticsearch_connection` via `providerschema.GetEsFWConnectionBlock()`, mirroring the data source envelope.
- [ ] 1.3 Implement `Read` on the envelope: state.Get into `T`, parse composite ID via `clients.CompositeIDFromStrFw`, resolve scoped client via `Client().GetElasticsearchClient(ctx, model.GetElasticsearchConnection())`, invoke `readFunc`, then either `state.Set` (found=true) or `state.RemoveResource` (found=false). Short-circuit on every diagnostic gate.
- [ ] 1.4 Implement `Delete` on the envelope: same prelude as Read, invoke `deleteFunc`, append diagnostics.
- [ ] 1.5 Implement default `ImportState` on the envelope as `resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)`.
- [ ] 1.6 Add interface assertions: `var _ resource.Resource = (*ElasticsearchResource[…])(nil)` etc., to fail compile if any method is missing.
- [ ] 1.7 Update `internal/entitycore/doc.go` to document the resource envelope alongside the data source envelope (mention the model constraint, callback signatures, and the default ImportState).
- [ ] 1.8 Add `internal/entitycore/resource_envelope_test.go` covering: constructor returns valid resource, Metadata type-name composition, Schema injects connection block, Read happy path, Read not-found removes state, Read short-circuits on each diagnostic gate, Delete happy path, Delete short-circuits on each diagnostic gate, default ImportState passthrough.

## 2. Migrate `elasticsearch_security_user`

- [ ] 2.1 Add value-receiver methods `GetID() types.String` and `GetElasticsearchConnection() types.List` on `securityuser.Data`.
- [ ] 2.2 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in `userResource`. Update `newUserResource()` to call `NewElasticsearchResource[Data]` with the schema factory, `readUser`, and `deleteUser` callbacks.
- [ ] 2.3 Convert the existing `Read` body into a package-level `readUser(ctx, client, resourceID, state Data) (Data, bool, diag.Diagnostics)`. Remove the `Read` method from the resource type.
- [ ] 2.4 Convert the existing `Delete` body into a package-level `deleteUser(ctx, client, resourceID, state Data) diag.Diagnostics`. Remove the `Delete` method from the resource type.
- [ ] 2.5 Drop the `elasticsearch_connection` block from the schema factory; the envelope injects it.
- [ ] 2.6 Remove the explicit `ImportState` method (envelope provides the same passthrough).
- [ ] 2.7 Verify Create and Update still compile and use `r.Client()` correctly via the embedded envelope.
- [ ] 2.8 Run `go test ./internal/elasticsearch/security/user/...` and the user acceptance tests; confirm no behavior change.

## 3. Migrate `elasticsearch_security_system_user`

- [ ] 3.1 Add `GetID()` and `GetElasticsearchConnection()` getters on `systemuser.Data`.
- [ ] 3.2 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in `systemUserResource`.
- [ ] 3.3 Convert `Read` body into `readSystemUser(ctx, client, resourceID, state Data) (Data, bool, diag.Diagnostics)`. The not-found path SHALL include `user == nil || !user.IsSystemUser()` returning `(_, false, nil)`.
- [ ] 3.4 Convert `Delete` body into `deleteSystemUser(ctx, _ *clients.ElasticsearchScopedClient, resourceID string, _ Data) diag.Diagnostics`. Body logs the existing tflog warning and returns nil. Add a one-line comment explaining why no API call is made.
- [ ] 3.5 Strip the `elasticsearch_connection` block from the schema factory.
- [ ] 3.6 Run `go test ./internal/elasticsearch/security/systemuser/...`.

## 4. Migrate `elasticsearch_security_role`

- [ ] 4.1 Add `GetID()` and `GetElasticsearchConnection()` getters on `role.Data`.
- [ ] 4.2 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in `roleResource`.
- [ ] 4.3 Refactor the existing `read(ctx, data Data) (*Data, diag.Diagnostics)` into `readRole(ctx, client, resourceID, state Data) (Data, bool, diag.Diagnostics)`. Return `(_, false, nil)` for the not-found branch. Update `update.go`'s post-write re-read site to call this new function (it currently calls `r.read(ctx, data)`); a thin local helper bridging the new signature is acceptable.
- [ ] 4.4 Convert `Delete` body into `deleteRole(ctx, client, resourceID, state Data) diag.Diagnostics`.
- [ ] 4.5 Strip the `elasticsearch_connection` block from the schema factory.
- [ ] 4.6 Preserve the `UpgradeState` method on the concrete resource (envelope does not provide it).
- [ ] 4.7 Run `go test ./internal/elasticsearch/security/role/...` plus the role acceptance test.

## 5. Migrate `elasticsearch_security_role_mapping`

- [ ] 5.1 Add `GetID()` and `GetElasticsearchConnection()` getters on `rolemapping.Data`.
- [ ] 5.2 Replace `*entitycore.ResourceBase` with `*entitycore.ElasticsearchResource[Data]` in `roleMappingResource`.
- [ ] 5.3 Convert `Read` body into `readRoleMappingResource(ctx, client, resourceID, state Data) (Data, bool, diag.Diagnostics)` that wraps the existing `readRoleMapping` helper. Map `nil` return to `(_, false, nil)`.
- [ ] 5.4 Convert `Delete` body into `deleteRoleMapping(ctx, client, resourceID, state Data) diag.Diagnostics`.
- [ ] 5.5 Strip the `elasticsearch_connection` block from the schema factory.
- [ ] 5.6 Run `go test ./internal/elasticsearch/security/rolemapping/...`.

## 6. Verification

- [ ] 6.1 `make build` passes.
- [ ] 6.2 `make check-lint` passes (golangci-lint + revive + openspec validate).
- [ ] 6.3 `make check-openspec` passes.
- [ ] 6.4 Acceptance test sweep against the running stack: `user`, `systemuser`, `role`, `rolemapping` (per `dev-docs/high-level/testing.md`).
- [ ] 6.5 Generated docs unchanged: confirm `terraform-docs` / `tfplugindocs` produces no diff for the four security resources.
- [ ] 6.6 Ensure `openspec validate elasticsearch-resource-envelope --strict` passes.
