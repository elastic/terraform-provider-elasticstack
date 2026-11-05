## ADDED Requirements

### Requirement: Kibana connection block exposes a `headers` attribute

The shared Kibana connection block SHALL expose an optional map-of-string attribute named `headers`.
Provider, managed-resource, data-source, and ephemeral variants SHALL mark the attribute sensitive,
while the action variant SHALL make it write-only. The attribute SHALL be available in every context
where the Kibana connection block is used without entity-specific schema changes. Entity-local
managed-resource and data-source header values SHALL remain in Terraform state as sensitive values;
action header values SHALL be excluded from state; and ephemeral header values SHALL follow
Terraform's ephemeral-state semantics.

#### Scenario: Provider-level kibana block accepts headers

- **WHEN** a practitioner configures `provider "elasticstack" { kibana { headers = { "X-Custom" =
  "value" } } }`
- **THEN** Terraform SHALL accept the configuration without a schema validation error

#### Scenario: Entity-local kibana_connection accepts headers

- **WHEN** a practitioner configures `kibana_connection { headers = { "X-Custom" = "value" } }` on a
  resource or data source that already supports `kibana_connection`
- **THEN** Terraform SHALL accept the configuration without a schema validation error
- **AND** the header value SHALL be retained in state as a sensitive value

#### Scenario: Action connection headers are write-only

- **WHEN** a practitioner configures `headers` on an action variant of the Kibana connection block
- **THEN** the header value SHALL be accepted as write-only input
- **AND** the header value SHALL NOT be retained in state

#### Scenario: Ephemeral connection headers follow ephemeral-state semantics

- **WHEN** an ephemeral variant of the Kibana connection block accepts `headers`
- **THEN** the header value SHALL be marked sensitive
- **AND** Terraform SHALL NOT persist the ephemeral value in state

#### Scenario: Omitted headers is a no-op

- **WHEN** a practitioner configures a `kibana` or `kibana_connection` block without `headers`
- **THEN** no additional request headers SHALL be sent, and existing configurations without `headers`
  SHALL continue to plan and apply with no diff attributable to this change

### Requirement: Invalid custom header sets are rejected before any request is sent

The provider SHALL reject reserved Kibana header names before the affected connection sends its first
request. A configured `headers` key equal to `kbn-xsrf`, `authorization`, `host`, `content-length`,
`transfer-encoding`, or `trailer`, compared case-insensitively, SHALL produce a diagnostic identifying
the offending key. The provider SHALL also reject two configured keys that differ only by case, with a
diagnostic identifying both conflicting keys. The diagnostic MAY occur during apply when a new
entity-local scoped connection is first built; this requirement does not guarantee plan-time rejection
in every context.

#### Scenario: Custom Authorization header is rejected

- **WHEN** a practitioner configures `headers = { "Authorization" = "custom" }` on a Kibana connection
  block
- **THEN** the provider SHALL report a diagnostic error identifying `Authorization` as a reserved header
  name before the affected connection sends its first request

#### Scenario: Custom kbn-xsrf header is rejected regardless of case

- **WHEN** a practitioner configures `headers = { "KBN-XSRF" = "false" }` on a Kibana connection block
- **THEN** the provider SHALL report a diagnostic error identifying the key as a reserved header name
  before the affected connection sends its first request

#### Scenario: Custom Host header is rejected rather than silently ignored

- **WHEN** a practitioner configures `headers = { "Host" = "kibana.example.com" }` on a Kibana
  connection block
- **THEN** the provider SHALL report a diagnostic error identifying `Host` as a reserved header name
  before the affected connection sends its first request

#### Scenario: Transport-controlled framing header is rejected

- **WHEN** a practitioner configures `headers` with `Content-Length`, `Transfer-Encoding`, or `Trailer`
  on a Kibana connection block
- **THEN** the provider SHALL report a diagnostic error identifying the offending key as a reserved
  header name before the affected connection sends its first request

#### Scenario: Header names that differ only by case are rejected

- **WHEN** a practitioner configures `headers = { "X-Proxy-Token" = "first", "x-proxy-token" =
  "second" }` on a Kibana connection block
- **THEN** the provider SHALL report a diagnostic error identifying both conflicting keys before the
  affected connection sends its first request

#### Scenario: Non-reserved headers are unaffected by the rejection rule

- **WHEN** a practitioner configures `headers` containing only uniquely named, non-reserved keys (for
  example `cf-access-token`)
- **THEN** the provider SHALL NOT report an invalid-header-set diagnostic for that configuration

### Requirement: Accepted configured headers are sent on every outgoing Kibana request

The effective Kibana connection SHALL send every accepted configured custom header on every outgoing Kibana
request. This behavior SHALL apply to provider-level and entity-local connections, alongside the
provider's own `kbn-xsrf` and `Authorization` headers.

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

### Requirement: Redirects cannot forward custom headers across origins

When one or more custom headers are configured, the Kibana client SHALL follow redirects only when
the destination has the same origin as the initial request. Origin equality SHALL require the same
scheme, host, and effective port. The client SHALL reject a cross-origin redirect, including an
HTTPS-to-HTTP redirect, before sending a request to the redirect destination.

#### Scenario: Same-origin redirect retains custom headers

- **GIVEN** a Kibana connection configured with `headers = { "cf-access-token" = "xyz" }`
- **WHEN** a Kibana endpoint redirects the request to another path on the same origin
- **THEN** the client SHALL follow the redirect
- **AND** the redirected request SHALL include `cf-access-token: xyz`

#### Scenario: Cross-origin redirect is rejected before credentials are forwarded

- **GIVEN** a Kibana connection configured with `headers = { "cf-access-token" = "xyz" }`
- **WHEN** a Kibana endpoint redirects the request to a different scheme, host, or effective port
- **THEN** the client SHALL return a redirect error
- **AND** the redirect destination SHALL NOT receive any request

#### Scenario: HTTPS-to-HTTP redirect is rejected

- **GIVEN** a Kibana connection configured with `headers = { "cf-access-token" = "xyz" }`
- **WHEN** a Kibana endpoint redirects an HTTPS request to an HTTP destination
- **THEN** the client SHALL return a redirect error
- **AND** the HTTP destination SHALL NOT receive any request

### Requirement: Fleet client excludes Kibana-only headers

The Fleet client SHALL exclude custom headers configured on the Kibana connection. This exclusion
SHALL apply regardless of whether the Fleet block configures other connection fields. Fleet does not
gain its own `headers` attribute as part of this requirement.

#### Scenario: Fleet requests omit Kibana's custom headers

- **GIVEN** a provider-level `kibana { headers = { "cf-access-token" = "xyz" } }` block and no
  `headers`-equivalent configuration on the `fleet` block
- **WHEN** the provider sends a Fleet API request
- **THEN** outgoing Fleet API requests SHALL NOT include `cf-access-token`

### Requirement: Accepted Kibana connection headers are not gated by a minimum Stack version

Accepted Kibana connection headers SHALL NOT be gated by a minimum Kibana or Stack version. The
behavior only shapes outgoing HTTP requests before they reach a Kibana API operation.

#### Scenario: Accepted headers are honored on any supported Stack version

- **WHEN** a practitioner configures non-reserved `headers` on a Kibana connection block against any
  Kibana version within the provider's supported version range
- **THEN** the provider SHALL send the accepted configured headers without emitting a version-related
  diagnostic
