## 1. Typed client migration — user helpers

- [x] 1.1 Rewrite `PutUser` in `internal/clients/elasticsearch/security.go` to use typed `Security.PutUser` API
- [x] 1.2 Rewrite `GetUser` to use typed `Security.GetUser` API and return `*types.User`
- [x] 1.3 Rewrite `DeleteUser` to use typed `Security.DeleteUser` API
- [x] 1.4 Rewrite `EnableUser` to use typed `Security.EnableUser` API
- [x] 1.5 Rewrite `DisableUser` to use typed `Security.DisableUser` API
- [x] 1.6 Rewrite `ChangeUserPassword` to use typed `Security.ChangePassword` API
- [x] 1.7 Update `internal/elasticsearch/security/user/` resource files for new helper signatures and `*types.User`
- [x] 1.8 Update `internal/elasticsearch/security/user_data_source.go` for new `GetUser` signature and `*types.User`
- [x] 1.9 Update `internal/elasticsearch/security/systemuser/` resource files for new helper signatures and `*types.User`

## 2. Typed client migration — role helpers

- [x] 2.1 Rewrite `PutRole` in `internal/clients/elasticsearch/security.go` to use typed `Security.PutRole` API
- [x] 2.2 Rewrite `GetRole` to use typed `Security.GetRole` API and return `*types.Role`
- [x] 2.3 Rewrite `DeleteRole` to use typed `Security.DeleteRole` API
- [x] 2.4 Update `internal/elasticsearch/security/role/` resource files for new helper signatures and `*types.Role`
- [x] 2.5 Update `internal/elasticsearch/security/role_data_source.go` for new `GetRole` signature and `*types.Role`

## 3. Typed client migration — role mapping helpers

- [x] 3.1 Rewrite `PutRoleMapping` in `internal/clients/elasticsearch/security.go` to use typed `Security.PutRoleMapping` API
- [x] 3.2 Rewrite `GetRoleMapping` to use typed `Security.GetRoleMapping` API and return `*types.SecurityRoleMapping`
- [x] 3.3 Rewrite `DeleteRoleMapping` to use typed `Security.DeleteRoleMapping` API
- [x] 3.4 Update `internal/elasticsearch/security/rolemapping/` resource files for new helper signatures and `*types.SecurityRoleMapping`
- [x] 3.5 Update `internal/elasticsearch/security/rolemapping/data_source.go` for new `GetRoleMapping` signature

## 4. Typed client migration — API key helpers

- [x] 4.1 Rewrite `CreateAPIKey` to use typed `Security.CreateApiKey` API and return `*createapikey.Response`
- [x] 4.2 Rewrite `GetAPIKey` to use typed `Security.GetApiKey` API and return `*types.ApiKey`
- [x] 4.3 Rewrite `UpdateAPIKey` to use typed `Security.UpdateApiKey` API
- [x] 4.4 Rewrite `DeleteAPIKey` to use typed `Security.InvalidateApiKey` API
- [x] 4.5 Update `internal/elasticsearch/security/api_key/` resource files for new helper signatures and typed responses

## 5. Typed client migration — cross-cluster API key helpers

- [x] 5.1 Rewrite `CreateCrossClusterAPIKey` to use typed `Security.CreateCrossClusterApiKey` API and return `*createcrossclusterapikey.Response`
- [x] 5.2 Rewrite `UpdateCrossClusterAPIKey` to use typed `Security.UpdateCrossClusterApiKey` API
- [x] 5.3 Update `internal/elasticsearch/security/api_key/` resource files for new cross-cluster helper signatures

## 6. Typed client migration — acceptance test helper

- [x] 6.1 Rewrite `CreateESAccessToken` in `internal/acctest/security_helpers.go` to use typed `Security.GetToken` API
- [x] 6.2 Remove raw `esapi` import and manual JSON marshaling from `security_helpers.go`

## 7. Model cleanup and verification

- [x] 7.1 Verify `models.User` is no longer used outside `security.go`, then remove from `internal/models/models.go`
- [x] 7.2 Verify `models.UserPassword` is no longer used, then remove from `internal/models/models.go`
- [x] 7.3 Verify `models.Role` is no longer used, then remove from `internal/models/models.go`
- [x] 7.4 Verify `models.RoleMapping` is no longer used, then remove from `internal/models/models.go`
- [x] 7.5 Verify `models.APIKey` is no longer used, then remove from `internal/models/models.go`
- [x] 7.6 Verify `models.APIKeyCreateResponse` is no longer used, then remove from `internal/models/models.go`
- [x] 7.7 Verify `models.APIKeyResponse` is no longer used, then remove from `internal/models/models.go`
- [x] 7.8 Verify `models.CrossClusterAPIKey` is no longer used, then remove from `internal/models/models.go`
- [x] 7.9 Verify `models.CrossClusterAPIKeyCreateResponse` is no longer used, then remove from `internal/models/models.go`
- [x] 7.10 Run `go mod tidy` and `make build` to confirm compilation

## 8. Testing

- [x] 8.1 Run unit tests for `internal/elasticsearch/security/user`
- [x] 8.2 Run unit tests for `internal/elasticsearch/security/role`
- [x] 8.3 Run unit tests for `internal/elasticsearch/security/rolemapping`
- [x] 8.4 Run unit tests for `internal/elasticsearch/security/api_key`
- [x] 8.5 Run unit tests for `internal/elasticsearch/security/systemuser`
- [x] 8.6 Run acceptance tests for `elasticstack_elasticsearch_security_user`
- [x] 8.7 Run acceptance tests for `elasticstack_elasticsearch_security_role`
- [x] 8.8 Run acceptance tests for `elasticstack_elasticsearch_security_role_mapping`
- [x] 8.9 Run acceptance tests for `elasticstack_elasticsearch_security_api_key`
- [x] 8.10 Run acceptance tests for `elasticstack_elasticsearch_security_system_user`
- [x] 8.11 Run acceptance tests for user data source
- [x] 8.12 Run acceptance tests for role data source
- [x] 8.13 Run `make check-lint` and `make check-openspec`
