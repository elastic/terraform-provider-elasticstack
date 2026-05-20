# `elasticstack_elasticsearch_security_api_key` (ephemeral) — Schema and Functional Requirements

Resource implementation (new): `internal/elasticsearch/security/api_key/ephemeral_resource.go`

## Purpose

Define the Terraform schema and runtime behavior for the ephemeral resource variant of `elasticstack_elasticsearch_security_api_key`. The ephemeral resource creates Elasticsearch API keys during `Open()` and — if configured — invalidates them during `Close()`. Credentials are held only in memory and are never written to Terraform state.

## Schema

```hcl
ephemeral "elasticstack_elasticsearch_security_api_key" "example" {
  # Input attributes
  name                = <required, string>            # 1–1024 chars, Basic Latin printable, no leading/trailing whitespace
  type                = <optional, string>            # "rest" (default) or "cross_cluster"
  role_descriptors    = <optional, json string>       # REST keys only; role descriptor object
  expiration          = <optional, string>            # duration string (e.g. "7d"); strongly recommended when invalidate_on_close = false
  metadata            = <optional, json string>       # arbitrary metadata object
  invalidate_on_close = <optional, bool, default false> # true = call Invalidate API after apply completes

  # Cross-cluster access (type = "cross_cluster" only)
  access = <optional, object({
    search      = optional(list(object({
      names                    = list(string)
      field_security           = optional(json string)
      query                    = optional(json string)
      allow_restricted_indices = optional(bool)
    })))
    replication = optional(list(object({
      names = list(string)
    })))
  })>

  # Deprecated: resource-level Elasticsearch connection override
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

  # Result attributes (computed; in-memory only, never in state)
  # key_id               = <computed, string>
  # api_key              = <computed, sensitive string>
  # encoded              = <computed, sensitive string>
  # expiration_timestamp = <computed, int64>  # epoch-ms; 0 if no expiration
}
```

## ADDED Requirements

### Requirement: Ephemeral resource registration (REQ-EPH-001)

The provider SHALL register the ephemeral resource factory via an `EphemeralResources()` method on the `Provider` type in `provider/plugin_framework.go`. The factory SHALL return a new instance of the ephemeral resource type implementing `ephemeral.EphemeralResource`, `EphemeralResourceWithConfigure`, and `EphemeralResourceWithClose`.

#### Scenario: Provider exposes the ephemeral resource

- GIVEN the provider is initialized
- WHEN Terraform discovers ephemeral resources from the provider schema
- THEN `elasticstack_elasticsearch_security_api_key` SHALL appear as an available ephemeral resource

### Requirement: Open creates a new API key and returns credentials in the result (REQ-EPH-002)

`Open()` SHALL be called by Terraform on every `plan` and `apply`. During `Open()`, the provider SHALL call the Elasticsearch Create API key API (for `type = "rest"`) or the Create cross-cluster API key API (for `type = "cross_cluster"`), and SHALL return the resulting credentials (`key_id`, `api_key`, `encoded`, `expiration_timestamp`) in `OpenResponse.Result`. Credentials SHALL NOT be written to Terraform state.

#### Scenario: Successful Open for a REST API key

- GIVEN a valid Elasticsearch connection and `type = "rest"` (or unset)
- WHEN Terraform opens the ephemeral resource
- THEN the provider SHALL call `POST /_security/api_key`
- AND the result SHALL contain non-empty `key_id`, `api_key`, and `encoded`
- AND `expiration_timestamp` SHALL be `0` when no `expiration` is set

#### Scenario: Successful Open for a cross-cluster API key

- GIVEN `type = "cross_cluster"` and Elasticsearch >= 8.10.0
- WHEN Terraform opens the ephemeral resource
- THEN the provider SHALL call `POST /_security/cross_cluster/api_key`
- AND the result SHALL contain non-empty `key_id`, `api_key`, and `encoded`

#### Scenario: Credentials absent from state after apply

- GIVEN any successful Open operation
- WHEN Terraform completes the apply
- THEN neither `api_key` nor `encoded` SHALL appear in the Terraform state file

### Requirement: Close invalidates the key when `invalidate_on_close = true` (REQ-EPH-003)

`Close()` SHALL be called by Terraform after the apply completes. The provider SHALL:
- When `invalidate_on_close = false` (default): perform no action. The API key persists in Elasticsearch until it expires or is manually invalidated.
- When `invalidate_on_close = true`: call the Elasticsearch Invalidate API key API (`POST /_security/api_key/invalidate`) using the `key_id` from the `Open()` result. After invalidation, the key SHALL be immediately unusable.

#### Scenario: Close with invalidate_on_close = false

- GIVEN `invalidate_on_close = false` (or unset) in the ephemeral resource configuration
- WHEN Terraform calls Close after the apply completes
- THEN the provider SHALL NOT call the Invalidate API key API
- AND the API key SHALL remain valid in Elasticsearch

#### Scenario: Close with invalidate_on_close = true

- GIVEN `invalidate_on_close = true` in the ephemeral resource configuration
- WHEN Terraform calls Close after the apply completes
- THEN the provider SHALL call `POST /_security/api_key/invalidate` with the `key_id` from the result
- AND the API key SHALL be marked as invalidated in Elasticsearch

#### Scenario: Run interrupted before Close is called

- GIVEN `invalidate_on_close = true`
- WHEN Terraform is killed before `Close()` is invoked
- THEN the API key SHALL remain alive until it expires naturally (or indefinitely if no expiration was set)
- AND the provider SHALL document this as a known limitation

### Requirement: Renew is not implemented (REQ-EPH-004)

The provider SHALL NOT implement `EphemeralResourceWithRenew` for this resource. Elasticsearch API keys cannot be refreshed server-side; a new key is created each run.

#### Scenario: Renew capability

- GIVEN a Terraform operation that would invoke Renew
- WHEN the framework checks whether EphemeralResourceWithRenew is implemented
- THEN the assertion SHALL fail and Terraform SHALL not call Renew

### Requirement: Cross-cluster version gate (REQ-EPH-005)

When `type = "cross_cluster"`, `Open()` SHALL verify that the Elasticsearch server version is at least `8.10.0` before calling the Create cross-cluster API key API. If the version requirement is not met, the provider SHALL return an error diagnostic and SHALL NOT create a key.

#### Scenario: Cross-cluster API key on an older cluster

- GIVEN `type = "cross_cluster"` and an Elasticsearch cluster older than `8.10.0`
- WHEN Terraform opens the ephemeral resource
- THEN the provider SHALL return an error diagnostic indicating the minimum version requirement
- AND no API key SHALL be created

### Requirement: Input validation mirrors the managed resource (REQ-EPH-006)

The ephemeral resource schema SHALL enforce the same input validation rules as the managed resource:

- `name`: required; 1–1024 characters; only printable Basic Latin (ASCII) characters plus spaces; no leading or trailing whitespace.
- `type`: optional; accepted values are `"rest"` and `"cross_cluster"`; defaults to `"rest"`.
- `role_descriptors`: valid only when `type = "rest"`.
- `access`: valid only when `type = "cross_cluster"`.

The schema SHALL NOT include plan modifiers (`RequiresReplace`, `UseStateForUnknown`) that are specific to managed resources.

#### Scenario: Invalid name

- GIVEN `name` is empty, exceeds 1024 characters, or contains non-printable characters
- WHEN schema validation runs
- THEN the provider SHALL return a validation error

#### Scenario: role_descriptors set for cross_cluster key

- GIVEN `type = "cross_cluster"` and `role_descriptors` is set
- WHEN schema validation runs
- THEN the provider SHALL return a validation error

#### Scenario: access set for rest key

- GIVEN `type = "rest"` (or unset) and `access` is set
- WHEN schema validation runs
- THEN the provider SHALL return a validation error

### Requirement: `elasticsearch_connection` block is supported (REQ-EPH-007)

The ephemeral resource SHALL support the deprecated `elasticsearch_connection` block, consistent with the managed resource, so that practitioners using a non-default connection can migrate to the ephemeral variant without losing this capability.

#### Scenario: Ephemeral resource with explicit connection

- GIVEN `elasticsearch_connection` is configured on the ephemeral resource
- WHEN `Open()` or `Close()` performs Elasticsearch API calls
- THEN the provider SHALL use the resource-scoped client derived from that block

### Requirement: Expiration populates expiration_timestamp (REQ-EPH-008)

When `expiration` is set and the Elasticsearch create response includes an expiration value, the provider SHALL populate `expiration_timestamp` in the result with the epoch-millisecond value from the response. When the API key does not expire, `expiration_timestamp` SHALL be `0`.

#### Scenario: Key with expiration

- GIVEN `expiration = "7d"` in the ephemeral resource configuration
- WHEN Open creates the key and Elasticsearch returns an expiration value
- THEN `expiration_timestamp` in the result SHALL be a non-zero epoch-millisecond value

#### Scenario: Key without expiration

- GIVEN no `expiration` attribute is set
- WHEN Open creates the key
- THEN `expiration_timestamp` in the result SHALL be `0`

### Requirement: Documentation warns about key accumulation and footgun (REQ-EPH-009)

The provider documentation for this ephemeral resource SHALL include:

1. A **warning** that combining `invalidate_on_close = true` with a persistent secret store (e.g. Vault) results in the stored credential being immediately invalidated after the Terraform run, making it unusable.
2. A **warning** that each `terraform plan` and `terraform apply` creates a new API key, and that setting `expiration` is strongly recommended when `invalidate_on_close = false` to prevent unlimited key accumulation.
3. A **note** explaining that `Open()` is invoked during both `terraform plan` and `terraform apply`, not only during apply.
4. A **note** that if Terraform is killed mid-apply before `Close()` is called, the key remains alive even when `invalidate_on_close = true`.

#### Scenario: Documentation content check

- GIVEN the generated documentation for the ephemeral resource
- WHEN a practitioner reads the docs
- THEN all four listed warnings/notes SHALL be present and legible

### Requirement: Existing managed resource is unaffected (REQ-EPH-010)

The implementation of the ephemeral resource SHALL NOT modify any existing file in `internal/elasticsearch/security/api_key/` other than by adding new files. The managed resource `Resource` type, its schema, models, converters, validators, state upgraders, and acceptance tests SHALL remain unchanged.

#### Scenario: Managed resource behavior after change

- GIVEN an existing `elasticstack_elasticsearch_security_api_key` managed resource in state
- WHEN the provider version is upgraded to include the ephemeral resource
- THEN the managed resource CRUD operations SHALL behave identically to before the upgrade
