## Why

The provider-wide typed-client migration is moving through its phases. Security is the largest API surface — covering users, roles, role mappings, API keys, and cross-cluster API keys. The `internal/clients/elasticsearch/security.go` helper file currently uses the raw `esapi` client with hand-structured request bodies, manual JSON marshaling, and ad-hoc response decoding. Migrating this to the `go-elasticsearch` Typed API (`elasticsearch.TypedClient` via `GetESTypedClient()`) eliminates runtime errors from untyped JSON, enables IDE-driven refactoring, and brings the security helpers in line with the rest of the migrated codebase.

## What Changes

- Migrate `internal/clients/elasticsearch/security.go` from raw `esapi` calls to typed `TypedClient.Security.*` equivalents.
- Migrate security resources and data sources (user, role, role mapping, API key, system user) to call typed helpers instead of raw-client helpers.
- Migrate `internal/acctest/security_helpers.go` (`CreateESAccessToken`) from raw `esapi` to typed client.
- Remove manual `json.Marshal`/`json.Unmarshal` boilerplate where typed request/response structs are available.
- No Terraform resource schemas, provider configuration, or user-visible behavior changes.

Files affected:
- `internal/clients/elasticsearch/security.go`
- `internal/acctest/security_helpers.go`
- `internal/elasticsearch/security/user/*`
- `internal/elasticsearch/security/role/*`
- `internal/elasticsearch/security/rolemapping/*`
- `internal/elasticsearch/security/api_key/*`
- `internal/elasticsearch/security/systemuser/*`
- `internal/elasticsearch/security/user_data_source.go`
- `internal/elasticsearch/security/role_data_source.go`

## Capabilities

### New Capabilities
_(none — this is an internal refactoring with no new user-visible capabilities)_

### Modified Capabilities
_(none — no spec-level requirements change; all resource schemas, validation, and behavior remain identical)_

## Impact

- **Code**: Security helpers (`security.go`), security resources, data sources, and acceptance-test helpers.
- **APIs**: No Terraform resource or data source behavior changes.
- **Dependencies**: Relies on existing `go-elasticsearch/v8` `ToTyped()` already exposed via `ElasticsearchScopedClient.GetESTypedClient()`.
- **Build / CI**: Compilation and acceptance tests are affected; no new dependencies introduced.
