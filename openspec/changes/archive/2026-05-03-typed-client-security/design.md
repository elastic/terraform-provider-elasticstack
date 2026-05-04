## Context

`internal/clients/elasticsearch/security.go` currently contains helper functions for all Elasticsearch Security API operations using the raw `esapi` client (`*elasticsearch.Client`). These helpers manually construct HTTP requests, marshal/unmarshal JSON into custom `models` types, and handle status-code-based error and not-found semantics. The `go-elasticsearch/v8` library provides strongly-typed equivalents via `elasticsearch.TypedClient`, available through `ElasticsearchScopedClient.GetESTypedClient()`.

The security surface is the largest in the provider-wide typed-client migration. It covers:

- **Users**: `PutUser`, `GetUser`, `DeleteUser`, `ChangeUserPassword`, `EnableUser`, `DisableUser`
- **Roles**: `PutRole`, `GetRole`, `DeleteRole`
- **Role mappings**: `PutRoleMapping`, `GetRoleMapping`, `DeleteRoleMapping`
- **API keys**: `CreateAPIKey`, `GetAPIKey`, `UpdateAPIKey`, `DeleteAPIKey`
- **Cross-cluster API keys**: `CreateCrossClusterAPIKey`, `UpdateCrossClusterAPIKey`
- **Acceptance-test helper**: `CreateESAccessToken` in `internal/acctest/security_helpers.go`

The custom models (`models.User`, `models.Role`, `models.RoleMapping`, `models.APIKey`, `models.APIKeyRoleDescriptor`, `models.CrossClusterAPIKey`, `models.UserPassword`) duplicate the shape of upstream typed-client types. The helpers also use shared `doSDKWrite`/`doFWWrite` wrappers that are tied to the raw `esapi` response type.

## Goals / Non-Goals

**Goals:**
- Rewrite all functions in `internal/clients/elasticsearch/security.go` to use typed `TypedClient.Security.*` APIs.
- Replace custom model types with typed-client equivalents (`types.User`, `types.Role`, `types.SecurityRoleMapping`, `types.ApiKey`, `types.RoleDescriptor`, `types.Access`, etc.) where possible.
- Update every resource, data source, and test that calls these helpers so the codebase compiles and all tests pass.
- Migrate `internal/acctest/security_helpers.go` (`CreateESAccessToken`) to the typed client.
- Preserve exact Terraform-visible behavior: identical state mapping, identical error messages where possible, identical not-found handling.

**Non-Goals:**
- Adding new Terraform resources or data sources.
- Changing schema definitions, validation, or plan modifiers.
- Modifying provider-level client construction or the scoped-client type itself.
- Removing `GetESClient()` or migrating other `internal/clients/elasticsearch/` files.

## Decisions

**1. Use `apiClient.GetESTypedClient()` inline in each helper rather than adding a cached typed-client accessor.**
- **Rationale**: `GetESTypedClient()` is already available from `typed-client-bootstrap`. Calling it per operation is cheap (cached via `sync.Once`) and keeps this change self-contained.
- **Alternative considered**: Cache the typed client locally in `security.go`. Rejected because the scoped client already caches it.

**2. Map typed-client response types directly into resource state rather than maintaining parallel custom models.**
- **Rationale**: The custom models duplicate upstream types. Once helpers return typed types, resources consume them directly, allowing model cleanup.
- **Alternative considered**: Keep custom models and add adapter functions. Rejected because it preserves duplication and adds boilerplate with no benefit.

**3. Preserve SDK-style `diag.Diagnostics` for SDK resources and `fwdiag.Diagnostics` for Plugin Framework resources.**
- **Rationale**: The helper signatures are consumed by both SDK (`user_data_source.go`, `role_data_source.go`) and PF (`user/`, `role/`, `rolemapping/`, `api_key/`, `systemuser/`) resources. Keeping existing diagnostic types avoids a broader PF migration.

**4. Keep not-found handling explicit and identical to today.**
- **Rationale**: The typed client returns errors that may wrap a 404. We will inspect the underlying `*http.Response` for `StatusCode == http.StatusNotFound` so resources continue to see the same nil+no-error behavior on missing resources.

**5. Retain `models.APIKeyRoleDescriptor` and related domain-specific types in resource models where they are embedded in JSON-with-defaults custom types.**
- **Rationale**: `api_key` resources use `customtypes.JSONWithDefaultsValue[map[string]models.APIKeyRoleDescriptor]` for `role_descriptors`. Changing this to `types.RoleDescriptor` would require updating the JSON-with-defaults generic type and its default-population logic, expanding the change surface. The helper layer can accept/return typed `types.RoleDescriptor` and the resource can convert at the boundary.
- **Alternative considered**: Replace `models.APIKeyRoleDescriptor` everywhere. Rejected because it would force changes to the `customtypes` package and the JSON defaults mechanism.

**6. Update `CreateESAccessToken` to use `typedapi.Security.GetToken().Do(ctx)`.**
- **Rationale**: The typed `gettoken.Request` and `gettoken.Response` types map cleanly to the current password-grant flow. This removes the last raw `esapi` call from `internal/acctest/security_helpers.go`.

## Risks / Trade-offs

- **[Risk]** Typed-client types may have slightly different JSON tags or omitempty behavior than our custom models, causing state-mapping mismatches. → **Mitigation**: Run the full acceptance test suite for affected security resources and compare TF state before/after.
- **[Risk]** `types.Role` uses `clusterprivilege.ClusterPrivilege` enum slices instead of raw `[]string`, and `Global` has a deeply nested `map[string]map[string]map[string][]string` shape that differs from our `map[string]any`. → **Mitigation**: Verify the role resource's `toAPIModel`/`fromAPIModel` conversions produce identical JSON payloads. Acceptance tests cover all role fields.
- **[Risk]** `types.SecurityRoleMapping.Rules` is `types.RoleMappingRule` (a typed union) rather than `map[string]any`. → **Mitigation**: The role mapping resource already serializes/rules via JSON strings. Converting `types.RoleMappingRule` back to JSON for state should yield the same normalized output.
- **[Risk]** `types.ApiKey` does not include `EncodedKey` or the raw `api_key` credential; these only exist on create responses. → **Mitigation**: Create helpers already return separate response structs (`*createapikey.Response`, `*createcrossclusterapikey.Response`). Read helpers return `*types.ApiKey`, and the resource preserves `api_key`/`encoded` from prior state, matching today's behavior.
- **[Risk]** Deleting custom models from `internal/models/models.go` may break imports elsewhere. → **Mitigation**: Verify all references are removed before deleting. Use `make build` to confirm.

## Migration Plan

1. Rewrite user helpers: `PutUser`, `GetUser`, `DeleteUser`, `EnableUser`, `DisableUser`, `ChangeUserPassword`.
2. Rewrite role helpers: `PutRole`, `GetRole`, `DeleteRole`.
3. Rewrite role mapping helpers: `PutRoleMapping`, `GetRoleMapping`, `DeleteRoleMapping`.
4. Rewrite API key helpers: `CreateAPIKey`, `GetAPIKey`, `UpdateAPIKey`, `DeleteAPIKey`.
5. Rewrite cross-cluster API key helpers: `CreateCrossClusterAPIKey`, `UpdateCrossClusterAPIKey`.
6. Rewrite `CreateESAccessToken` in `internal/acctest/security_helpers.go`.
7. Update all call sites in security resources and data sources to match new helper signatures and typed types.
8. Remove now-unused custom model types from `internal/models/models.go` (or verify they are unused elsewhere).
9. Run `make build` and targeted acceptance tests for affected resources.

## Open Questions

- None — the scoped client, typed client surface, and security resource patterns are already well understood.
