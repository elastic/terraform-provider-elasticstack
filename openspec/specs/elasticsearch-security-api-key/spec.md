# `elasticstack_elasticsearch_security_api_key` — Schema and Functional Requirements

Resource implementation: `internal/elasticsearch/security/api_key`

## Purpose

Define the Terraform schema and runtime behavior for the `elasticstack_elasticsearch_security_api_key` resource, including regular and cross-cluster API keys, version-gated features, composite identity handling, state upgrades, and the documented key-rotation workflow.

## Schema

```hcl
resource "elasticstack_elasticsearch_security_api_key" "example" {
  name = <required, string>

  # Immutable: requires replacement when changed.
  type       = <optional+computed, string> # "rest" (default) or "cross_cluster"
  expiration = <optional, string>          # by default keys do not expire

  # REST API keys only.
  role_descriptors = <optional+computed, json string>

  # Cross-cluster API keys only.
  access = <optional, object({
    search = optional(list(object({
      names                    = list(string)
      field_security           = optional(json string)
      query                    = optional(json string)
      allow_restricted_indices = optional(bool)
    })))
    replication = optional(list(object({
      names = list(string)
    })))
  })>

  metadata = <optional+computed, json string>

  id                   = <computed, string> # <cluster_uuid>/<api_key_id>
  key_id               = <computed, string>
  expiration_timestamp = <computed, int64>  # 0 when Elasticsearch reports no expiration

  api_key = <computed, sensitive, string>
  encoded = <computed, sensitive, string>

  # Deprecated: resource-level Elasticsearch connection override.
  elasticsearch_connection {
    endpoints                = <optional, list(string)>
    username                 = <optional, string>
    password                 = <optional, string>
    api_key                  = <optional, string>
    bearer_token             = <optional, string>
    es_client_authentication = <optional, string>
    insecure                 = <optional, bool>
    ca_file                  = <optional, string>
    ca_data                  = <optional, string>
    cert_file                = <optional, string>
    cert_data                = <optional, string>
    key_file                 = <optional, string>
    key_data                 = <optional, string>
    headers                  = <optional, map(string)>
  }
}
```

## Requirements

### Requirement: API lifecycle uses Elasticsearch security APIs (REQ-001-REQ-006)

The resource SHALL create regular API keys with the Elasticsearch Create API key API, create cross-cluster API keys with the Create cross-cluster API key API, read keys with the Get API key API, update regular API keys with the Update API key API, update cross-cluster API keys with the Update cross-cluster API key API, and delete keys by invalidating them with the Invalidate API key API.

#### Scenario: Resource operation selects the matching Elasticsearch API

- GIVEN an API key resource operation runs
- WHEN the provider performs create, read, update, or delete
- THEN it SHALL call the Elasticsearch security API that matches the key type and lifecycle step

### Requirement: API and response errors surface to Terraform (REQ-007-REQ-008)

When Elasticsearch returns a non-success status for create, read, update, or delete, the resource SHALL surface that error through Terraform diagnostics. When the create API returns a nil response body, or when the Get API returns a response containing anything other than exactly one API key, the resource SHALL return an error diagnostic instead of silently continuing.

#### Scenario: Elasticsearch returns an unexpected response

- GIVEN an API call fails or returns an invalid success payload
- WHEN the provider processes the response
- THEN Terraform SHALL receive an error diagnostic describing the failure

### Requirement: Composite identity and stored identifiers (REQ-009-REQ-012)

The resource SHALL expose a computed `id` in the format `<cluster_uuid>/<api_key_id>` and a computed `key_id` containing the Elasticsearch API key identifier. After a successful create, the resource SHALL build `id` from the current cluster UUID and the created API key id. During read and delete, the resource SHALL parse the stored `id` to determine the API key id, and if the stored `id` does not match the composite format it SHALL return an error diagnostic.

#### Scenario: Stored id is malformed

- GIVEN state contains an `id` that is not `<cluster_uuid>/<resource identifier>`
- WHEN read or delete parses the id
- THEN the provider SHALL return an error diagnostic instead of calling Elasticsearch

### Requirement: Import is not supported (REQ-013)

The resource SHALL NOT support Terraform import.

#### Scenario: Import capability

- GIVEN a user attempts to import this resource
- WHEN Terraform checks the resource capabilities
- THEN the resource SHALL provide no import handler

### Requirement: Immutable arguments require replacement (REQ-014-REQ-016)

When `name`, `type`, or `expiration` changes, the resource SHALL require replacement rather than performing an in-place update.

#### Scenario: Immutable argument changes

- GIVEN a configuration change to `name`, `type`, or `expiration`
- WHEN Terraform plans the change
- THEN the resource SHALL plan a replacement

### Requirement: Client selection follows provider or resource connection settings (REQ-017-REQ-018)

The resource SHALL use the provider-configured Elasticsearch client by default. When the deprecated `elasticsearch_connection` block is configured on the resource, the resource SHALL create and use a client derived from that block for create, read, update, and delete operations on that instance.

#### Scenario: Resource-scoped Elasticsearch connection

- GIVEN `elasticsearch_connection` is set on the resource
- WHEN the resource performs any Elasticsearch API call
- THEN it SHALL use the resource-scoped client instead of the provider default client

### Requirement: Type-dependent validation and defaults (REQ-019-REQ-022)

The resource SHALL default `type` to `rest` when it is unset or unknown at plan time, and SHALL only accept `rest` or `cross_cluster` as configured values. The resource SHALL allow `role_descriptors` only when `type="rest"` and SHALL allow `access` only when `type="cross_cluster"`. The resource SHALL require `name` to be between 1 and 1024 characters and contain only printable Basic Latin characters plus spaces, with no leading or trailing whitespace.

#### Scenario: Attribute used with the wrong API key type

- GIVEN a configuration sets `role_descriptors` for `cross_cluster` or `access` for `rest`
- WHEN schema validation runs
- THEN the resource SHALL return a validation error for the mismatched attribute

### Requirement: Version-gated features and planning behavior (REQ-023-REQ-026)

When `type="cross_cluster"`, the resource SHALL verify on create that the Elasticsearch server version is at least `8.10.0`, and SHALL fail with an explicit unsupported-feature error when it is lower. When any role descriptor includes a `restriction`, the resource SHALL verify that Elasticsearch is at least `8.9.0`, and SHALL fail if restrictions are unsupported. The resource SHALL require replacement for changes to `metadata` or `role_descriptors` when the last cluster version recorded during refresh is lower than `8.4.0`, because update APIs are not supported there. During refresh against Elasticsearch `8.5.0` or newer, the resource SHALL read `role_descriptors` from the API response; on older versions it SHALL preserve the prior state value instead.

#### Scenario: Cross-cluster API keys on an older cluster

- GIVEN a configuration sets `type = "cross_cluster"`
- WHEN create runs against Elasticsearch older than `8.10.0`
- THEN the resource SHALL fail with an unsupported-feature error

### Requirement: JSON and access mapping (REQ-027-REQ-029)

When `metadata`, `role_descriptors`, `access.search[].field_security`, or `access.search[].query` are configured, the resource SHALL parse them as JSON before building the Elasticsearch request, and SHALL return an error diagnostic if JSON decoding fails. When `role_descriptors` is present, the resource SHALL apply the provider's role-descriptor defaults before sending the request. When building a cross-cluster API key request, the resource SHALL omit `access` unless at least one `search` or `replication` entry is populated.

#### Scenario: Invalid JSON input

- GIVEN one of the JSON-backed attributes contains invalid JSON
- WHEN create or update builds the Elasticsearch request
- THEN the resource SHALL fail before sending the API request

### Requirement: Create writes sensitive outputs and then refreshes state (REQ-030-REQ-033)

On create, the resource SHALL select the regular or cross-cluster create API according to `type`, submit the request payload derived from the Terraform plan, set `id`, `key_id`, `api_key`, and `encoded` from the create response, and then read the created key back from Elasticsearch to populate the remaining state fields.

#### Scenario: Successful create

- GIVEN Elasticsearch accepts the create request
- WHEN the provider finalizes state
- THEN state SHALL include the composite `id`, the computed `key_id`, the sensitive credentials, and refreshed non-sensitive fields from a follow-up read

### Requirement: Read refreshes state and preserves non-returned sensitive values (REQ-034-REQ-038)

During refresh, the resource SHALL read the API key identified by the stored composite `id`. If Elasticsearch returns HTTP 404, the resource SHALL remove the resource from Terraform state. When a key is returned, the resource SHALL set `key_id`, `name`, `expiration_timestamp`, and `metadata` from the API response. The resource SHALL preserve prior-state values for `api_key` and `encoded`, because the Get API does not return them. After a successful refresh, the resource SHALL save the Elasticsearch cluster version in private state for later planning decisions.

#### Scenario: API key no longer exists

- GIVEN refresh runs for a key that has been removed from Elasticsearch
- WHEN the Get API returns HTTP 404
- THEN the resource SHALL remove itself from Terraform state

### Requirement: Update only changes mutable API key fields (REQ-039-REQ-041)

During update, the resource SHALL call the regular or cross-cluster update API according to `type`, SHALL identify the target API key by `key_id`, and SHALL omit immutable fields such as `id`, `name`, and `expiration` from the update request payload. After a successful update, the resource SHALL read the API key again and write the refreshed state.

#### Scenario: Update request payload

- GIVEN Terraform updates a managed API key in place
- WHEN the provider builds the update request
- THEN it SHALL send only mutable fields and SHALL refresh state afterward

### Requirement: Cross-cluster access changes force deferred role descriptor planning (REQ-042)

When the resource manages a cross-cluster API key and the configured `access` value differs from the prior state's `access` value, the resource SHALL set `role_descriptors` to unknown during planning so Terraform does not compare incompatible access representations.

#### Scenario: Cross-cluster access changes

- GIVEN `type = "cross_cluster"` and `access` changes between prior state and new config
- WHEN planning evaluates `role_descriptors`
- THEN the planned `role_descriptors` value SHALL become unknown

### Requirement: Delete invalidates the API key and removes state (REQ-043)

On delete, the resource SHALL invalidate the API key id parsed from the stored composite `id` by calling the Elasticsearch Invalidate API key API, and SHALL remove the resource from Terraform state after a successful invalidation.

#### Scenario: Destroying an API key

- GIVEN Terraform destroys the resource
- WHEN delete runs
- THEN the provider SHALL invalidate the referenced API key and remove the resource from state

### Requirement: State upgrades preserve compatibility (REQ-044-REQ-045)

The resource SHALL support upgrading schema version `0` to `1` by converting `expiration = ""` to `expiration = null`. The resource SHALL support upgrading schema version `1` to `2` by setting `type = "rest"` in upgraded state.

#### Scenario: Upgrading legacy state

- GIVEN Terraform reads stored state from schema version `0` or `1`
- WHEN the provider upgrades the state
- THEN it SHALL apply the documented conversions before continuing

### Requirement: Documentation includes a rotation workflow example (REQ-046-REQ-047)

The resource documentation and example configuration SHALL include an API key rotation workflow that uses `time_rotating` to trigger replacement through the `name` attribute and `create_before_destroy` to keep a valid key available during rotation. The acceptance suite SHALL verify that changing the rotation trigger causes both the `time_rotating` id and the API key `key_id` to change.

#### Scenario: Rotation example is exercised by acceptance coverage

- GIVEN the rotation acceptance test changes the `time_rotating` trigger input
- WHEN Terraform reapplies the configuration
- THEN the rotation id and the managed API key id SHALL both change
