## Purpose

The `provider-fleet-connection` capability governs the provider's top-level `fleet {}` connection block:
how it authenticates to and locates Fleet, and the custom HTTP headers it may attach to outgoing Fleet API
requests independently of the Kibana connection.

## ADDED Requirements

### Requirement: Fleet connection block exposes a `headers` attribute

The `fleet {}` provider connection block SHALL expose an optional, sensitive map-of-string attribute named
`headers`. This attribute SHALL be available only on the provider-level `fleet {}` block; no entity-local
`fleet_connection` block exists. The provider SHALL trim leading and trailing whitespace from configured
header names and values before validation and use.

#### Scenario: Provider-level fleet block accepts headers

- **WHEN** a practitioner configures `provider "elasticstack" { fleet { headers = { "X-Custom" = "value" }
  } }`
- **THEN** Terraform SHALL accept the configuration without a schema validation error

#### Scenario: Omitted headers is a no-op

- **WHEN** a practitioner configures a `fleet {}` block without `headers`, or configures no `fleet {}` block
  at all
- **THEN** no additional request headers SHALL be sent to Fleet on behalf of this attribute, and existing
  configurations without `fleet.headers` SHALL continue to plan and apply with no diff attributable to this
  change

### Requirement: Invalid Fleet header sets are rejected before any Fleet request is sent, using the same reserved-name list as Kibana

The provider SHALL reject a `fleet.headers` key equal to `kbn-xsrf`, `authorization`, `host`,
`content-length`, `transfer-encoding`, or `trailer`, compared case-insensitively, using the same
reserved-name list and the same shared validation logic already enforced for the Kibana connection's
`headers` attribute. The provider SHALL also reject two configured `fleet.headers` keys that differ only by
case. Each rejection SHALL produce a diagnostic identifying the offending key (or both conflicting keys)
before the Fleet connection sends its first request.

#### Scenario: Custom Authorization header is rejected on the Fleet block

- **WHEN** a practitioner configures `fleet { headers = { "Authorization" = "custom" } }`
- **THEN** the provider SHALL report a diagnostic error identifying `Authorization` as a reserved header name
  before the Fleet connection sends its first request

#### Scenario: Custom kbn-xsrf header is rejected on the Fleet block regardless of case

- **WHEN** a practitioner configures `fleet { headers = { "KBN-XSRF" = "false" } }`
- **THEN** the provider SHALL report a diagnostic error identifying the key as a reserved header name before
  the Fleet connection sends its first request

#### Scenario: Transport-controlled framing header is rejected on the Fleet block

- **WHEN** a practitioner configures `fleet { headers }` with `Host`, `Content-Length`,
  `Transfer-Encoding`, or `Trailer`
- **THEN** the provider SHALL report a diagnostic error identifying the offending key as a reserved header
  name before the Fleet connection sends its first request

#### Scenario: Fleet header names that differ only by case are rejected

- **WHEN** a practitioner configures `fleet { headers = { "X-Proxy-Token" = "first", "x-proxy-token" =
  "second" } }`
- **THEN** the provider SHALL report a diagnostic error identifying both conflicting keys before the Fleet
  connection sends its first request

#### Scenario: Non-reserved Fleet headers are unaffected by the rejection rule

- **WHEN** a practitioner configures `fleet.headers` containing only uniquely named, non-reserved keys (for
  example `cf-access-client-id`)
- **THEN** the provider SHALL NOT report an invalid-header-set diagnostic for that configuration

### Requirement: Accepted Fleet headers are sent on every outgoing Fleet request

The effective Fleet connection SHALL send every accepted configured `fleet.headers` entry on every outgoing
Fleet API request, alongside the provider's own `kbn-xsrf` and `Authorization` headers for Fleet.

#### Scenario: Custom Fleet header reaches a Fleet request

- **GIVEN** a Fleet connection configured with `fleet { headers = { "cf-access-client-id" = "abc" } }`
- **WHEN** the provider sends a request to the Fleet API
- **THEN** the outgoing request SHALL include a `cf-access-client-id: abc` header

#### Scenario: Custom Fleet header reaches a request with no provider-configured Fleet auth

- **GIVEN** a Fleet connection configured with `fleet { headers = { "cf-access-client-id" = "abc" } }` and no
  Fleet-specific `username`, `password`, `api_key`, or `bearer_token`
- **WHEN** the provider sends a request to the Fleet API
- **THEN** the outgoing request SHALL include `cf-access-client-id: abc`

### Requirement: Fleet headers and Kibana headers are independent in both directions

Configuring `kibana.headers` SHALL NOT cause its values to appear on outgoing Fleet requests, and
configuring `fleet.headers` SHALL NOT cause its values to appear on outgoing Kibana requests. The provider
SHALL NOT infer or forward either connection's headers to the other based on matching, overlapping, or
diverging endpoints; forwarding occurs only when a value is explicitly configured on the corresponding
block.

#### Scenario: Kibana headers still do not reach Fleet once Fleet has its own headers attribute

- **GIVEN** a provider-level `kibana { headers = { "cf-access-token" = "xyz" } }` block and no `fleet.headers`
  configuration
- **WHEN** the provider sends a Fleet API request
- **THEN** the outgoing Fleet API request SHALL NOT include `cf-access-token`

#### Scenario: Fleet headers do not reach Kibana requests

- **GIVEN** a provider-level `fleet { headers = { "cf-access-client-id" = "abc" } }` block and no
  `kibana.headers` configuration
- **WHEN** the provider sends a Kibana API request
- **THEN** the outgoing Kibana API request SHALL NOT include `cf-access-client-id`

#### Scenario: Kibana and Fleet endpoints resolving to the same host still require independent configuration

- **GIVEN** `kibana.endpoints` and `fleet.endpoint` resolve to the same host, `kibana.headers` is configured,
  and `fleet.headers` is not configured
- **WHEN** the provider sends a Fleet API request
- **THEN** the outgoing Fleet API request SHALL NOT include any header from `kibana.headers`, because this
  capability does not implement same-host header forwarding

### Requirement: Accepted Fleet connection headers are not gated by a minimum Stack version

Accepted `fleet.headers` values SHALL NOT be gated by a minimum Kibana, Fleet, or Stack version. The behavior
only shapes outgoing HTTP requests before they reach a Fleet API operation.

#### Scenario: Accepted Fleet headers are honored on any supported Stack version

- **WHEN** a practitioner configures non-reserved `fleet.headers` against any Kibana/Fleet version within the
  provider's supported version range
- **THEN** the provider SHALL send the accepted configured headers without emitting a version-related
  diagnostic
