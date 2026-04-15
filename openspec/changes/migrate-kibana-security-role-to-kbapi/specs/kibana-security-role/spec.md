## ADDED Requirements

### Requirement: kibanaoapi security role helpers (REQ-025)

The provider SHALL implement Kibana role management HTTP calls for `elasticstack_kibana_security_role` via functions in `internal/clients/kibanaoapi` that use `generated/kbapi` client methods for put, get, and delete by role name (`PutSecurityRoleNameWithResponse` or `PutSecurityRoleName`, `GetSecurityRoleNameWithResponse` or `GetSecurityRoleName`, `DeleteSecurityRoleNameWithResponse` or `DeleteSecurityRoleName`). Helpers SHALL decode successful JSON GET bodies into the same structural family used for PUT (`PutSecurityRoleNameJSONBody` or equivalent) before returning data to Terraform mapping code. Helpers SHALL map non-success HTTP status codes (other than not-found on read as defined in REQ-012–REQ-014) to Terraform diagnostics including response body context when available.

#### Scenario: Helper surfaces decode errors

- **GIVEN** a GET response whose body is not valid JSON for the role document
- **WHEN** the helper parses the response
- **THEN** the provider SHALL return an error diagnostic and SHALL NOT silently treat the role as absent

### Requirement: Privilege mapping parity (REQ-026)

Before this change is considered complete, the provider SHALL demonstrate that Elasticsearch index and remote index privileges, Kibana base and feature privileges, `metadata`, and `description` round-trip equivalently versus the pre-migration behavior for representative configurations covered by tests. At minimum, existing acceptance tests for `elasticstack_kibana_security_role` SHALL pass unchanged against a supported Stack, and unit tests SHALL assert stable JSON or struct equivalence for at least one index entry with `field_security`, one `remote_indices` entry, one `kibana` feature block, and one `kibana` base block.

#### Scenario: Acceptance suite unchanged

- **GIVEN** the existing `internal/kibana` acceptance tests for create, update, remote indices, and description
- **WHEN** the tests run against a cluster meeting documented version gates
- **THEN** they SHALL pass without relaxing assertions

## MODIFIED Requirements

### Requirement: Role Management APIs (REQ-001–REQ-003)

The resource SHALL create and update roles using the Kibana Create or update role HTTP API invoked through `internal/clients/kibanaoapi` on top of `generated/kbapi` (`PutSecurityRoleName` for `/api/security/roles/{name}`) ([docs](https://www.elastic.co/guide/en/kibana/current/role-management-specific-api-put.html)). The resource and data source SHALL read roles using the Kibana Get role HTTP API via the same helper layer (`GetSecurityRoleName`) ([docs](https://www.elastic.co/guide/en/kibana/current/role-management-specific-api-get.html)). The resource SHALL delete roles using the Kibana Delete role HTTP API via the same helper layer (`DeleteSecurityRoleName`) ([docs](https://www.elastic.co/guide/en/kibana/current/role-management-specific-api-delete.html)). The resource and data source SHALL NOT call `KibanaRoleManagement.CreateOrUpdate`, `KibanaRoleManagement.Get`, or `KibanaRoleManagement.Delete` for this entity once migration is complete. When a Kibana API call returns an error for create, update, read, or delete (other than role not found on read), the resource SHALL surface the error to Terraform diagnostics.

#### Scenario: API errors surfaced

- GIVEN a failing Kibana API response (other than role not found on read)
- WHEN the provider processes the response
- THEN diagnostics SHALL include the API error

### Requirement: Identity (REQ-004)

The resource SHALL expose a computed `id` equal to the role name. After a successful create or update, the resource SHALL set `id` to the configured role `name` (the path parameter used for PUT), which SHALL equal the role name persisted in Kibana after a successful response.

#### Scenario: Computed id after create

- GIVEN a successful create
- WHEN the provider commits state after write
- THEN `id` SHALL be set to the role `name` argument

### Requirement: Create and update behavior (REQ-010–REQ-011)

When creating a role, the resource SHALL set the `createOnly` query parameter to `true` on the PUT request to signal new-resource semantics. When updating an existing role, the resource SHALL set `createOnly` to `false` or omit it per Kibana API conventions so the role can be overwritten. When creating or updating a role, the resource SHALL build the API request body from all configured fields (`name`, `kibana`, `elasticsearch`, `metadata`, `description`) using `generated/kbapi` request types and submit it with the Create or update role API. After a successful API response, the resource SHALL set `id` and read the role back to populate state.

#### Scenario: Post-apply read

- GIVEN a successful create or update
- WHEN the provider refreshes state
- THEN it SHALL call the Get role API and populate state from the response

### Requirement: Read and refresh (REQ-012–REQ-014)

When refreshing state, the resource and data source SHALL use `id` (or `name` for the data source) as the role name to fetch. If the Get role HTTP response indicates the role does not exist (for example HTTP 404, or the documented empty success response if applicable), the resource SHALL remove itself from state (role not found) and the data source SHALL behave as today for a missing role (diagnostics or empty result per existing implementation). When a role is found, the resource SHALL set `name`, `elasticsearch`, `kibana`, `description`, and `metadata` in state from the decoded API response.

#### Scenario: Role removed in Kibana

- GIVEN refresh runs and the role no longer exists
- WHEN the Get role API returns a not-found result as defined above
- THEN the resource SHALL be removed from state

### Requirement: Delete (REQ-015)

When destroying, the resource SHALL use `id` as the role name and delete it via the Kibana Delete role HTTP API implemented through `internal/clients/kibanaoapi` and `generated/kbapi`.

#### Scenario: Destroy

- GIVEN destroy is requested
- WHEN delete runs
- THEN the provider SHALL call Delete role for the name stored in `id`
