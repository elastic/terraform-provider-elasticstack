# typed-client-security Specification

## Purpose
TBD - created by archiving change typed-client-security. Update Purpose after archive.
## Requirements
### Requirement: Security helpers use typed client
`PutUser`, `GetUser`, `DeleteUser`, `EnableUser`, `DisableUser`, and `ChangeUserPassword` in `internal/clients/elasticsearch/security.go` SHALL use the typed `Security.PutUser`, `Security.GetUser`, `Security.DeleteUser`, `Security.EnableUser`, `Security.DisableUser`, and `Security.ChangePassword` APIs. `GetUser` SHALL return `*types.User` and SHALL return `nil` with no error when the user is not found.

#### Scenario: Create or update user
- **WHEN** `PutUser` is called with a valid user definition
- **THEN** it calls the typed `Security.PutUser` API
- **AND** returns no error diagnostics on success

#### Scenario: Read existing user
- **WHEN** `GetUser` is called for an existing user
- **THEN** it returns `*types.User` and no error diagnostics

#### Scenario: Read missing user
- **GIVEN** the requested user does not exist
- **WHEN** `GetUser` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete user
- **WHEN** `DeleteUser` is called with a valid username
- **THEN** it calls the typed `Security.DeleteUser` API
- **AND** returns no error diagnostics on success

#### Scenario: Enable user
- **WHEN** `EnableUser` is called with a valid username
- **THEN** it calls the typed `Security.EnableUser` API
- **AND** returns no error diagnostics on success

#### Scenario: Disable user
- **WHEN** `DisableUser` is called with a valid username
- **THEN** it calls the typed `Security.DisableUser` API
- **AND** returns no error diagnostics on success

#### Scenario: Change user password
- **WHEN** `ChangeUserPassword` is called with a valid username and password payload
- **THEN** it calls the typed `Security.ChangePassword` API
- **AND** returns no error diagnostics on success

### Requirement: Role helpers use typed client
`PutRole`, `GetRole`, and `DeleteRole` in `internal/clients/elasticsearch/security.go` SHALL use the typed `Security.PutRole`, `Security.GetRole`, and `Security.DeleteRole` APIs. `GetRole` SHALL return `*types.Role` and SHALL return `nil` with no error when the role is not found.

#### Scenario: Create or update role
- **WHEN** `PutRole` is called with a valid role definition
- **THEN** it calls the typed `Security.PutRole` API
- **AND** returns no error diagnostics on success

#### Scenario: Read existing role
- **WHEN** `GetRole` is called for an existing role
- **THEN** it returns `*types.Role` and no error diagnostics

#### Scenario: Read missing role
- **GIVEN** the requested role does not exist
- **WHEN** `GetRole` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete role
- **WHEN** `DeleteRole` is called with a valid role name
- **THEN** it calls the typed `Security.DeleteRole` API
- **AND** returns no error diagnostics on success

### Requirement: Role mapping helpers use typed client
`PutRoleMapping`, `GetRoleMapping`, and `DeleteRoleMapping` in `internal/clients/elasticsearch/security.go` SHALL use the typed `Security.PutRoleMapping`, `Security.GetRoleMapping`, and `Security.DeleteRoleMapping` APIs. `GetRoleMapping` SHALL return `*types.SecurityRoleMapping` and SHALL return `nil` with no error when the role mapping is not found.

#### Scenario: Create or update role mapping
- **WHEN** `PutRoleMapping` is called with a valid role mapping definition
- **THEN** it calls the typed `Security.PutRoleMapping` API
- **AND** returns no error diagnostics on success

#### Scenario: Read existing role mapping
- **WHEN** `GetRoleMapping` is called for an existing role mapping
- **THEN** it returns `*types.SecurityRoleMapping` and no error diagnostics

#### Scenario: Read missing role mapping
- **GIVEN** the requested role mapping does not exist
- **WHEN** `GetRoleMapping` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Delete role mapping
- **WHEN** `DeleteRoleMapping` is called with a valid role mapping name
- **THEN** it calls the typed `Security.DeleteRoleMapping` API
- **AND** returns no error diagnostics on success

### Requirement: API key helpers use typed client
`CreateAPIKey`, `GetAPIKey`, `UpdateAPIKey`, and `DeleteAPIKey` in `internal/clients/elasticsearch/security.go` SHALL use the typed `Security.CreateApiKey`, `Security.GetApiKey`, `Security.UpdateApiKey`, and `Security.InvalidateApiKey` APIs. `GetAPIKey` SHALL return `*types.ApiKey` and SHALL return `nil` with no error when the API key is not found.

#### Scenario: Create API key
- **WHEN** `CreateAPIKey` is called with a valid API key definition
- **THEN** it calls the typed `Security.CreateApiKey` API
- **AND** returns the create response and no error diagnostics on success

#### Scenario: Read existing API key
- **WHEN** `GetAPIKey` is called for an existing API key
- **THEN** it returns `*types.ApiKey` and no error diagnostics

#### Scenario: Read missing API key
- **GIVEN** the requested API key does not exist
- **WHEN** `GetAPIKey` is called
- **THEN** it returns `nil` and no error diagnostics

#### Scenario: Update API key
- **WHEN** `UpdateAPIKey` is called with a valid API key definition
- **THEN** it calls the typed `Security.UpdateApiKey` API
- **AND** returns no error diagnostics on success

#### Scenario: Delete API key
- **WHEN** `DeleteAPIKey` is called with a valid API key id
- **THEN** it calls the typed `Security.InvalidateApiKey` API
- **AND** returns no error diagnostics on success

### Requirement: Cross-cluster API key helpers use typed client
`CreateCrossClusterAPIKey` and `UpdateCrossClusterAPIKey` in `internal/clients/elasticsearch/security.go` SHALL use the typed `Security.CreateCrossClusterApiKey` and `Security.UpdateCrossClusterApiKey` APIs.

#### Scenario: Create cross-cluster API key
- **WHEN** `CreateCrossClusterAPIKey` is called with a valid definition
- **THEN** it calls the typed `Security.CreateCrossClusterApiKey` API
- **AND** returns the create response and no error diagnostics on success

#### Scenario: Update cross-cluster API key
- **WHEN** `UpdateCrossClusterAPIKey` is called with a valid definition
- **THEN** it calls the typed `Security.UpdateCrossClusterApiKey` API
- **AND** returns no error diagnostics on success

### Requirement: Acceptance test helper uses typed client
`CreateESAccessToken` in `internal/acctest/security_helpers.go` SHALL use the typed `Security.GetToken` API instead of the raw `esapi` client.

#### Scenario: Create access token for tests
- **WHEN** `CreateESAccessToken` is called
- **THEN** it calls the typed `Security.GetToken` API with the password grant type
- **AND** returns the access token string on success

### Requirement: Custom security model types are removed
The custom model types `User`, `UserPassword`, `Role`, `RoleMapping`, `APIKey`, `APIKeyCreateResponse`, `APIKeyResponse`, `CrossClusterAPIKey`, and `CrossClusterAPIKeyCreateResponse` in `internal/models/models.go` SHALL be removed once all callers have been migrated to typed client equivalents. No remaining code SHALL reference these types after the migration.

#### Scenario: Build succeeds after model removal
- **GIVEN** all callers have been updated to use typed client types
- **WHEN** the custom models are removed from `internal/models/models.go`
- **THEN** `make build` completes successfully

