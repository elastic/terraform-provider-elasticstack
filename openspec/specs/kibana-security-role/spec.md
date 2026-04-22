# `elasticstack_kibana_security_role` — Schema and Functional Requirements

Resource implementation: `internal/kibana`
Data source implementation: `internal/kibana`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_kibana_security_role` resource and data source: Kibana Role Management API usage, identity, import, lifecycle, connection, compatibility gates, and mapping of Elasticsearch and Kibana privilege blocks to and from API state.

## Schema

```hcl
resource "elasticstack_kibana_security_role" "example" {
  name        = <required, string>         # force new
  description = <optional, string>         # requires Kibana >= 8.15.0 when set
  metadata    = <optional, computed, json string>

  elasticsearch {                          # required, set (max 1)
    cluster        = <optional, set(string)>
    run_as         = <optional, set(string)>

    indices {                              # optional, set
      names      = <required, set(string)>
      privileges = <required, set(string)>
      query      = <optional, string>      # JSON string, diff-suppressed
      field_security {                     # optional, list (max 1)
        grant  = <optional, set(string)>
        except = <optional, set(string)>
      }
    }

    remote_indices {                       # optional, set; requires Kibana >= 8.10.0 when non-empty
      clusters   = <required, set(string)>
      names      = <required, set(string)>
      privileges = <required, set(string)>
      query      = <optional, string>      # JSON string, diff-suppressed
      field_security {                     # optional, list (max 1)
        grant  = <optional, set(string)>
        except = <optional, set(string)>
      }
    }
  }

  kibana {                                 # optional, set
    spaces = <required, set(string)>
    base   = <optional, set(string)>       # must be ["all"] or ["read"]; mutually exclusive with feature
    feature {                              # optional, set; mutually exclusive with base
      name       = <required, string>
      privileges = <required, set(string)>
    }
  }
}

data "elasticstack_kibana_security_role" "example" {
  name        = <required, string>
  description = <optional, computed, string>
  metadata    = <optional, computed, json string>

  elasticsearch {                          # computed, set
    cluster        = <computed, set(string)>
    run_as         = <computed, set(string)>

    indices {                              # computed, set
      names          = <computed, set(string)>
      privileges     = <computed, set(string)>
      query          = <computed, string>
      field_security {                     # computed, list
        grant  = <computed, set(string)>
        except = <computed, set(string)>
      }
    }

    remote_indices {                       # optional, set
      clusters   = <required, set(string)>
      names      = <required, set(string)>
      privileges = <required, set(string)>
      query      = <optional, string>      # JSON string, diff-suppressed
      field_security {                     # optional, list (max 1)
        grant  = <optional, set(string)>
        except = <optional, set(string)>
      }
    }
  }

  kibana {                                 # computed, set
    spaces  = <computed, set(string)>
    base    = <computed, set(string)>
    feature {                              # computed, set
      name       = <computed, string>
      privileges = <computed, set(string)>
    }
  }
}
```

## Requirements

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

### Requirement: Import (REQ-005)

The resource SHALL support import via `schema.ImportStatePassthroughContext`, accepting an `id` equal to the role name and persisting it directly to state.

#### Scenario: Import by name

- GIVEN a role that exists in Kibana
- WHEN imported with the role name as the `id`
- THEN state SHALL be populated from the Kibana API

### Requirement: Name change lifecycle (REQ-006)

When the `name` argument changes, the resource SHALL require replacement (destroy then recreate), not an in-place update.

#### Scenario: Renaming a role

- GIVEN a configuration change to `name`
- WHEN Terraform plans the change
- THEN the resource SHALL be replaced

### Requirement: Connection (REQ-007)

The resource and data source SHALL use the provider's configured Kibana client by default for all API calls. When `kibana_connection` is configured on the resource or data source, the entity SHALL resolve an effective scoped client from that block and SHALL use the scoped Kibana client for the API call.

#### Scenario: Provider-level Kibana client

- **WHEN** `kibana_connection` is not configured on the resource or data source
- **THEN** the provider-level Kibana client SHALL be used

#### Scenario: Scoped Kibana connection

- **WHEN** `kibana_connection` is configured on the resource or data source
- **THEN** the scoped Kibana client derived from that block SHALL be used

### Requirement: Version compatibility — description (REQ-008)

When `description` is configured (non-empty), the resource SHALL verify the Kibana server version is at least 8.15.0, and if it is lower the resource SHALL fail with an error.

#### Scenario: Description on older Kibana

- GIVEN `description` is set and Kibana is below 8.15.0
- WHEN create or update runs
- THEN the provider SHALL fail with an error indicating the minimum supported version

### Requirement: Version compatibility — remote_indices (REQ-009)

When `elasticsearch.remote_indices` is configured with one or more entries, the resource SHALL verify the Kibana server version is at least 8.10.0, and if it is lower the resource SHALL fail with an error.

#### Scenario: Remote indices on older Kibana

- GIVEN one or more `remote_indices` entries and Kibana is below 8.10.0
- WHEN create or update runs
- THEN the provider SHALL fail with an error indicating the minimum supported version

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

### Requirement: Metadata mapping (REQ-016)

When `metadata` is configured, the resource SHALL parse it as JSON before sending it in the API request; if parsing fails, the resource SHALL return an error and SHALL not call the API. When the API response contains `metadata`, the resource SHALL serialize it to a JSON string and store it in state; when absent or nil, the resource SHALL not update the `metadata` state attribute.

#### Scenario: Invalid metadata JSON

- GIVEN `metadata` contains invalid JSON
- WHEN create or update runs
- THEN the provider SHALL return an error and SHALL NOT call the Create or update role API

### Requirement: Kibana privilege block mapping (REQ-017–REQ-018)

When `kibana` blocks are configured, the resource SHALL send each block as a `KibanaRoleKibana` entry with `base`, `feature`, and `spaces` fields. Within a single `kibana` block, `base` and `feature` SHALL be mutually exclusive: if both are non-empty, the resource SHALL return an error diagnostic. If neither `base` nor `feature` is non-empty, the resource SHALL return an error diagnostic requiring at least one to be set.

#### Scenario: Both base and feature set

- GIVEN a `kibana` block with both `base` and `feature` configured
- WHEN create or update runs
- THEN the provider SHALL error and SHALL NOT call the API

### Requirement: Elasticsearch index privilege mapping (REQ-019–REQ-020)

When `elasticsearch.indices` entries are configured, the resource SHALL map `names`, `privileges`, `query`, and `field_security` (with `grant` and `except`) to the corresponding Kibana API index privilege fields. When `query` is an empty string, the resource SHALL omit `query` from the API payload for that index entry. When `field_security` has at least one entry, the resource SHALL send `grant` and `except` as arrays in the `field_security` map.

#### Scenario: Query omitted when empty

- GIVEN an `indices` entry with no `query` value
- WHEN the API payload is built
- THEN the `query` field SHALL be omitted from that entry

### Requirement: Elasticsearch remote index privilege mapping (REQ-021)

When `elasticsearch.remote_indices` entries are configured, the resource SHALL map `clusters`, `names`, `privileges`, `query`, and `field_security` to the corresponding API fields. When `query` is an empty string, the resource SHALL omit `query` from the API payload for that remote index entry.

#### Scenario: Remote index mapping

- GIVEN `remote_indices` entries are configured
- WHEN the API payload is built
- THEN `clusters`, `names`, `privileges`, and optional `query` and `field_security` SHALL be populated correctly

### Requirement: Read state — elasticsearch block (REQ-022–REQ-023)

When the API response contains an `elasticsearch` object, the resource SHALL set the `elasticsearch` block in state including `cluster`, `indices`, `remote_indices`, and `run_as`. When the API `cluster` list is empty or when `run_as` is empty, those fields SHALL be omitted from the flattened state (not stored as empty lists). When no `elasticsearch` object is present in the response, the resource SHALL store an empty list for the `elasticsearch` attribute.

#### Scenario: Empty cluster and run_as omitted

- GIVEN the API returns an empty `cluster` list
- WHEN state is updated from the API
- THEN `cluster` SHALL be omitted from the flattened elasticsearch block

### Requirement: Read state — kibana block (REQ-024)

When the API response contains `kibana` privilege entries, the resource SHALL flatten each entry into a `kibana` block with `base`, `feature`, and `spaces`. When the API returns no `kibana` privileges, the resource SHALL store an empty list for the `kibana` attribute.

#### Scenario: Kibana privileges flattened

- GIVEN the API returns kibana privilege entries
- WHEN read populates state
- THEN each entry SHALL appear as a `kibana` block with `base`, `feature`, and `spaces`

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
