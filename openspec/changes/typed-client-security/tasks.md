## 1. Typed client migration — user helpers

- [ ] 1.1 Rewrite `PutUser` in `internal/clients/elasticsearch/security.go` to use typed `Security.PutUser` API
- [ ] 1.2 Rewrite `GetUser` to use typed `Security.GetUser` API and return `*types.User`
- [ ] 1.3 Rewrite `DeleteUser` to use typed `Security.DeleteUser` API
- [ ] 1.4 Rewrite `EnableUser` to use typed `Security.EnableUser` API
- [ ] 1.5 Rewrite `DisableUser` to use typed `Security.DisableUser` API
- [ ] 1.6 Rewrite `ChangeUserPassword` to use typed `Security.ChangePassword` API
- [ ] 1.7 Update `internal/elasticsearch/security/user/` resource files for new helper signatures and `*types.User`
- [ ] 1.8 Update `internal/elasticsearch/security/user_data_source.go` for new `GetUser` signature and `*types.User`
- [ ] 1.9 Update `internal/elasticsearch/security/systemuser/` resource files for new helper signatures and `*types.User`

## 2. Typed client migration — role helpers

- [ ] 2.1 Rewrite `PutRole` in `internal/clients/elasticsearch/security.go` to use typed `Security.PutRole` API
- [ ] 2.2 Rewrite `GetRole` to use typed `Security.GetRole` API and return `*types.Role`
- [ ] 2.3 Rewrite `DeleteRole` to use typed `Security.DeleteRole` API
- [ ] 2.4 Update `internal/elasticsearch/security/role/` resource files for new helper signatures and `*types.Role`
- [ ] 2.5 Update `internal/elasticsearch/security/role_data_source.go` for new `GetRole` signature and `*types.Role`

## 3. Typed client migration — role mapping helpers

- [ ] 3.1 Rewrite `PutRoleMapping` in `internal/clients/elasticsearch/security.go` to use typed `Security.PutRoleMapping` API
- [ ] 3.2 Rewrite `GetRoleMapping` to use typed `Security.GetRoleMapping` API and return `*types.SecurityRoleMapping`
- [ ] 3.3 Rewrite `DeleteRoleMapping` to use typed `Security.DeleteRoleMapping` API
- [ ] 3.4 Update `internal/elasticsearch/security/rolemapping/` resource files for new helper signatures and `*types.SecurityRoleMapping`
- [ ] 3.5 Update `internal/elasticsearch/security/rolemapping/data_source.go` for new `GetRoleMapping` signature

## 4. Typed client migration — API key helpers

- [ ] 4.1 Rewrite `CreateAPIKey` to use typed `Security.CreateApiKey` API and return `*createapikey.Response`
- [ ] 4.2 Rewrite `GetAPIKey` to use typed `Security.GetApiKey` API and return `*types.ApiKey`
- [ ] 4.3 Rewrite `UpdateAPIKey` to use typed `Security.UpdateApiKey` API
- [ ] 4.4 Rewrite `DeleteAPIKey` to use typed `Security.InvalidateApiKey` API
- [ ] 4.5 Update `internal/elasticsearch/security/api_key/` resource files for new helper signatures and typed responses

## 5. Typed client migration — cross-cluster API key helpers

- [ ] 5.1 Rewrite `CreateCrossClusterAPIKey` to use typed `Security.CreateCrossClusterApiKey` API and return `*createcrossclusterapikey.Response`
- [ ] 5.2 Rewrite `UpdateCrossClusterAPIKey` to use typed `Security.UpdateCrossClusterApiKey` API
- [ ] 5.3 Update `internal/elasticsearch/security/api_key/` resource files for new cross-cluster helper signatures

## 6. Typed client migration — acceptance test helper

- [ ] 6.1 Rewrite `CreateESAccessToken` in `internal/acctest/security_helpers.go` to use typed `Security.GetToken` API
- [ ] 6.2 Remove raw `esapi` import and manual JSON marshaling from `security_helpers.go`

## 7. Model cleanup and verification

- [ ] 7.1 Verify `models.User` is no longer used outside `security.go`, then remove from `internal/models/models.go`
- [ ] 7.2 Verify `models.UserPassword` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.3 Verify `models.Role` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.4 Verify `models.RoleMapping` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.5 Verify `models.APIKey` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.6 Verify `models.APIKeyCreateResponse` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.7 Verify `models.APIKeyResponse` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.8 Verify `models.CrossClusterAPIKey` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.9 Verify `models.CrossClusterAPIKeyCreateResponse` is no longer used, then remove from `internal/models/models.go`
- [ ] 7.10 Run `go mod tidy` and `make build` to confirm compilation

## 8. Testing

- [ ] 8.1 Run unit tests for `internal/elasticsearch/security/user`
- [ ] 8.2 Run unit tests for `internal/elasticsearch/security/role`
- [ ] 8.3 Run unit tests for `internal/elasticsearch/security/rolemapping`
- [ ] 8.4 Run unit tests for `internal/elasticsearch/security/api_key`
- [ ] 8.5 Run unit tests for `internal/elasticsearch/security/systemuser`
- [ ] 8.6 Run acceptance tests for `elasticstack_elasticsearch_security_user`
- [ ] 8.7 Run acceptance tests for `elasticstack_elasticsearch_security_role`
- [ ] 8.8 Run acceptance tests for `elasticstack_elasticsearch_security_role_mapping`
- [ ] 8.9 Run acceptance tests for `elasticstack_elasticsearch_security_api_key`
- [ ] 8.10 Run acceptance tests for `elasticstack_elasticsearch_security_system_user`
- [ ] 8.11 Run acceptance tests for user data source
- [ ] 8.12 Run acceptance tests for role data source
- [ ] 8.13 Run `make check-lint` and `make check-openspec`
