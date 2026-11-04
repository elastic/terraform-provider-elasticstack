## ADDED Requirements

### Requirement: Kibana connection block exposes a `headers` attribute

The shared Kibana connection block (`kbConnectionBlockSpec()`) SHALL define an `Optional`,
`Sensitive`, `WriteOnly` map-of-string attribute named `headers`, mirroring the Elasticsearch
connection block's `headers` attribute. Adding it to the shared block SHALL make it available in every
context the block is used — provider-level, entity-local, ephemeral, and action — without any
entity-specific schema change.

#### Scenario: Provider-level kibana block accepts headers

- **WHEN** a practitioner configures `provider "elasticstack" { kibana { headers = { "X-Custom" =
  "value" } } }`
- **THEN** Terraform SHALL accept the configuration without a schema validation error

#### Scenario: Entity-local kibana_connection accepts headers

- **WHEN** a practitioner configures `kibana_connection { headers = { "X-Custom" = "value" } }` on a
  resource or data source that already supports `kibana_connection`
- **THEN** Terraform SHALL accept the configuration without a schema validation error

#### Scenario: Omitted headers is a no-op

- **WHEN** a practitioner configures a `kibana` or `kibana_connection` block without `headers`
- **THEN** no additional request headers SHALL be sent, and existing configurations without `headers`
  SHALL continue to plan and apply with no diff attributable to this change

### Requirement: Reserved header names are rejected before any request is sent

The provider SHALL reject, case-insensitively, a configured Kibana connection `headers` key equal to
`kbn-xsrf` or `authorization`, with a diagnostic error identifying the offending key, before sending
any request. The provider's own writes for these two headers are conditional on request method and
configured auth method, so write order alone cannot guarantee a same-named custom header is always
overridden; rejecting the collision up front avoids relying on that ordering.

#### Scenario: Custom Authorization header is rejected

- **WHEN** a practitioner configures `headers = { "Authorization" = "custom" }` on a Kibana connection
  block
- **THEN** the provider SHALL report a diagnostic error identifying `Authorization` as a reserved header
  name before any Kibana API call is made

#### Scenario: Custom kbn-xsrf header is rejected regardless of case

- **WHEN** a practitioner configures `headers = { "KBN-XSRF" = "false" }` on a Kibana connection block
- **THEN** the provider SHALL report a diagnostic error identifying the key as a reserved header name
  before any Kibana API call is made

#### Scenario: Non-reserved headers are unaffected by the rejection rule

- **WHEN** a practitioner configures `headers` containing only non-reserved keys (for example
  `cf-access-token`)
- **THEN** the provider SHALL NOT report a reserved-header diagnostic for that configuration

### Requirement: Configured headers are sent on every outgoing Kibana request

When one or more `headers` entries are configured on the effective Kibana connection (provider-level or
entity-local), the Kibana HTTP client SHALL set each configured header, by name and value, on every
outgoing request to the Kibana API, in addition to the provider's own `kbn-xsrf` and `Authorization`
headers.

#### Scenario: Custom header reaches a GET request

- **GIVEN** a Kibana connection configured with `headers = { "cf-access-token" = "xyz" }`
- **WHEN** the provider sends a GET request to the Kibana API
- **THEN** the outgoing request SHALL include a `cf-access-token: xyz` header

#### Scenario: Custom header reaches a mutating request alongside kbn-xsrf

- **GIVEN** a Kibana connection configured with `headers = { "cf-access-token" = "xyz" }`
- **WHEN** the provider sends a POST, PUT, or DELETE request to the Kibana API
- **THEN** the outgoing request SHALL include both `cf-access-token: xyz` and `kbn-xsrf: true`

#### Scenario: Custom header reaches a request with no provider-configured auth

- **GIVEN** a Kibana connection configured with `headers = { "Proxy-Authorization" = "proxy-value" }`
  and no `username`, `password`, `api_key`, or `bearer_token`
- **WHEN** the provider sends a request to the Kibana API
- **THEN** the outgoing request SHALL include `Proxy-Authorization: proxy-value`
- **AND** the outgoing request SHALL NOT include an `Authorization` header

### Requirement: Fleet client excludes Kibana-only headers

The Fleet client configuration derived from the effective Kibana connection (`toFleetConfig()`) SHALL
NOT carry the Kibana connection's `headers` into the Fleet client, regardless of whether the Fleet
block itself configures any connection fields. Fleet does not gain its own `headers` attribute as part
of this requirement.

#### Scenario: Fleet requests omit Kibana's custom headers

- **GIVEN** a provider-level `kibana { headers = { "cf-access-token" = "xyz" } }` block and no
  `headers`-equivalent configuration on the `fleet` block
- **WHEN** the provider derives the Fleet client configuration from the Kibana connection
- **THEN** outgoing Fleet API requests SHALL NOT include `cf-access-token`

### Requirement: Kibana connection headers are not gated by a minimum Stack version

The `headers` attribute on the Kibana connection block SHALL NOT be subject to a `VersionRequirement`
or equivalent minimum-Kibana/Stack-version check, since it only shapes outgoing HTTP requests before
they reach any Kibana or Fleet API operation.

#### Scenario: Headers are honored on any supported Stack version

- **WHEN** a practitioner configures `headers` on a Kibana connection block against any Kibana version
  within the provider's supported version range
- **THEN** the provider SHALL send the configured headers without emitting a version-related diagnostic
