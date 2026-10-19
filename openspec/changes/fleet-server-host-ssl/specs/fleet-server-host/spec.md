## ADDED Requirements

### Requirement: Optional ssl attribute

The resource SHALL expose an optional `ssl` single-nested-object attribute with sub-attributes `certificate` (string, optional), `key` (string, optional, sensitive), `certificate_authorities` (list of string, optional, 1-10 non-empty entries), and `client_auth` (string, optional, one of `"none"`, `"optional"`, `"required"`). The provider SHALL map `certificate`, `key`, `certificate_authorities`, and `client_auth` to and from the `ssl` object on the Fleet create, update, and read response bodies.

#### Scenario: Create sends ssl fields

- **WHEN** the practitioner sets `ssl.certificate`, `ssl.key`, `ssl.certificate_authorities`, and/or `ssl.client_auth` on a new resource
- **THEN** the provider SHALL include the corresponding fields in the create request body's `ssl` object
- **AND** any `ssl` sub-attribute left unset SHALL be omitted from the request body's `ssl` object rather than sent as an empty or zero value

#### Scenario: Update sends changed ssl fields

- **WHEN** the practitioner changes any `ssl` sub-attribute on an existing resource and applies
- **THEN** the provider SHALL include the corresponding fields in the update request body's `ssl` object

#### Scenario: Read hydrates certificate, certificate_authorities, and client_auth from the API response

- **WHEN** the provider reads an existing server host whose stored `ssl` object has a non-null `certificate`, `certificate_authorities`, and/or `client_auth`
- **THEN** state SHALL contain a non-null `ssl` object with those fields populated from the response

#### Scenario: Omitted ssl is null in state

- **WHEN** the practitioner does not set `ssl` and the API response's `ssl` object is absent or has no non-null fields
- **THEN** state SHALL contain a null `ssl` object

#### Scenario: key is marked sensitive

- **WHEN** the practitioner inspects the `ssl.key` attribute in the resource schema
- **THEN** it SHALL be marked `Sensitive`

#### Scenario: Invalid client_auth value rejected

- **WHEN** the practitioner configures `ssl.client_auth` to a value other than `"none"`, `"optional"`, or `"required"`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: client_auth omitted is not defaulted by the provider

- **WHEN** the practitioner configures `ssl` without setting `client_auth`
- **THEN** the provider SHALL omit `client_auth` from the request body's `ssl` object rather than sending a provider-chosen default
- **AND** an apply followed by refresh SHALL complete without an inconsistent-result error or a subsequent diff while configuration continues to omit `client_auth`

#### Scenario: client_auth default returned by Kibana remains plan-consistent

- **GIVEN** a pre-implementation create probe confirms Kibana returns a non-null `client_auth` when it is omitted
- **WHEN** the practitioner applies configuration that omits `client_auth` and then refreshes
- **THEN** the provider SHALL represent the returned value using a plan-consistent schema and hydration strategy
- **AND** the provider SHALL NOT send a computed/defaulted value in place of the omitted field in the initial create request
- **AND** the apply and refresh SHALL produce no inconsistent-result error or subsequent diff

### Requirement: No provider-enforced ssl default or pairing constraint

The provider SHALL NOT enforce a default value for `client_auth`, nor a "certificate and key must be set together" constraint, since the Kibana API declares neither at the schema level.

#### Scenario: certificate without key is accepted

- **WHEN** the practitioner configures `ssl.certificate` without `ssl.key`
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a validation error

#### Scenario: key without certificate is accepted

- **WHEN** the practitioner configures `ssl.key` without `ssl.certificate`
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a validation error

### Requirement: ssl empty-value validation

The resource SHALL reject, at plan time, any configuration where `ssl` is set but ambiguously empty (all sub-attributes null), rather than silently normalizing the ambiguity to null on a later read.

#### Scenario: Empty ssl object rejected

- **WHEN** the practitioner configures `ssl = {}`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Empty string certificate or key rejected

- **WHEN** the practitioner configures `ssl.certificate = ""` or `ssl.key = ""`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Empty certificate_authorities list rejected

- **WHEN** the practitioner configures `ssl.certificate_authorities = []`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Empty-string CA element rejected

- **WHEN** the practitioner configures `ssl.certificate_authorities` containing one or more empty-string elements
- **THEN** Terraform SHALL report a validation error before any API call

### Requirement: ssl certificate_authorities size limit

The `certificate_authorities` sub-attribute SHALL accept at most 10 entries, matching the Fleet API's `maxItems: 10` constraint.

#### Scenario: Exactly 10 CAs accepted

- **WHEN** the practitioner configures `ssl.certificate_authorities` with exactly 10 non-empty entries
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a size validation error

#### Scenario: 11 CAs rejected

- **WHEN** the practitioner configures `ssl.certificate_authorities` with 11 non-empty entries
- **THEN** Terraform SHALL report a validation error before any API call

### Requirement: ssl minimum Kibana version guard

The `ssl` attribute SHALL be guarded by its own attribute-scoped minimum Kibana version requirement of 9.1.0, enforced only when `ssl` is configured, independent of the resource's other attributes. Configurations that do not set `ssl` SHALL NOT be affected by this requirement, on any stack version. `serverHostModel` SHALL implement `GetVersionRequirements` to enforce this.

#### Scenario: ssl create or update against unsupported stack version

- **WHEN** the practitioner configures `ssl` with at least one sub-attribute set
- **AND** the connected Kibana is below version 9.1.0
- **AND** the practitioner applies a create or update
- **THEN** the provider SHALL emit an attribute-scoped diagnostic on `ssl` that it is not supported on this version before sending a Fleet mutation request

#### Scenario: ssl omitted against old stack version is unaffected

- **WHEN** the practitioner does not configure `ssl`
- **AND** the connected Kibana is below version 9.1.0
- **THEN** the provider SHALL NOT emit a version diagnostic related to `ssl`
- **AND** the apply SHALL proceed as it did before this change

### Requirement: Upgrade migration for pre-existing server-side ssl

The provider SHALL document that users upgrading with an existing server-side `ssl` object must add its settings to Terraform configuration as part of the upgrade. The provider's read hydration SHALL NOT be described as preserving unmanaged remote `ssl` when configuration omits it; an unchanged configuration that omits such an object may plan its removal or resource replacement, according to the tested PUT semantics.

#### Scenario: Unchanged configuration with no pre-existing ssl

- **GIVEN** a resource managed by the previous provider version has no server-side `ssl` object
- **WHEN** the provider is upgraded and the unchanged configuration still omits `ssl`
- **THEN** refresh and planning SHALL produce no diff attributable to `ssl`

#### Scenario: Unchanged configuration with pre-existing unmanaged ssl

- **GIVEN** a resource managed by the previous provider version has a server-side `ssl` object that its configuration omits
- **WHEN** the provider is upgraded and the unchanged configuration still omits `ssl`
- **THEN** planning SHALL surface the removal or replacement implied by the tested API semantics
- **AND** the resource documentation SHALL direct the practitioner to add the existing SSL settings to configuration to adopt them

### Requirement: Protect deferred SSL fields on update

The provider SHALL NOT overwrite non-null `agent_*` or `es_*` fields already stored in a host's `ssl` object when those fields are not exposed by the Terraform schema. Before a PUT that would replace the `ssl` object, the provider SHALL either preserve these fields or reject the update with an actionable diagnostic before sending the mutation.

#### Scenario: Update cannot erase pre-existing unexposed SSL fields

- **GIVEN** a host's stored `ssl` object contains one or more non-null `agent_*` or `es_*` fields
- **WHEN** an update would replace the stored `ssl` object but the provider cannot preserve those fields
- **THEN** the provider SHALL return an actionable diagnostic before sending the PUT
- **AND** the server-side `agent_*` and `es_*` values SHALL remain unchanged
- **AND** an acceptance test SHALL seed these fields and verify no mutating request is sent

### Requirement: ssl removal consistency with PUT semantics

Implementation SHALL independently probe omission, `{}`, and explicit JSON `ssl: null` against fresh populated hosts before choosing removal behavior. It SHALL use a supported in-place clearing form or force replacement if none work.

#### Scenario: In-place removal when the API honors it

- **GIVEN** the pre-implementation probe determines that at least one of omitting `ssl`, sending `ssl = {}`, or sending explicit JSON `ssl: null` from the PUT body clears it server-side
- **WHEN** the practitioner removes `ssl` (or that child) from configuration and applies
- **THEN** the provider SHALL perform an in-place update, and a subsequent read SHALL show the removed value as null
- **AND** when explicit JSON null is the clearing form, the serialized PUT body SHALL contain `"ssl": null`

#### Scenario: Forced replacement when the API cannot honor in-place removal

- **GIVEN** the pre-implementation probe determines that omitting `ssl`, sending `ssl = {}`, and sending explicit JSON `ssl: null` all fail to clear a previously-set `ssl`, or that omitting a specific child does not clear that child
- **WHEN** the practitioner removes `ssl` (or that child) from configuration and applies
- **THEN** Terraform SHALL plan a replace operation for the resource

### Requirement: Explicit-null SSL clearing request

If explicit JSON `ssl: null` is the only probed form that clears the object, implementation SHALL use a request path that serializes `"ssl": null`. A nil generated `Ssl` pointer that serializes by omission does not meet this requirement.

#### Scenario: Explicit-null clearing payload is serialized

- **GIVEN** the pre-implementation probe determines explicit JSON null is the only in-place clearing form
- **WHEN** the provider sends the clearing PUT
- **THEN** the serialized request body SHALL contain `"ssl": null`
- **AND** a serialized-payload assertion SHALL verify the explicit null

### Requirement: ssl.key hydration on read is probe-determined

The provider's handling of `ssl.key` when populating state after create, update, and refresh SHALL depend on whether the Fleet API echoes the configured `key` value back verbatim on `GET`/`POST`/`PUT` responses, or redacts/omits it (as determined by a pre-implementation probe against a live Kibana instance).

#### Scenario: API echoes key verbatim

- **GIVEN** the pre-implementation probe determines the Fleet API returns `ssl.key` unredacted
- **WHEN** the provider populates state after create, update, or refresh
- **THEN** `ssl.key` SHALL be hydrated from the API response, consistent with `certificate` and `certificate_authorities`

#### Scenario: API redacts or omits key

- **GIVEN** the pre-implementation probe determines the Fleet API redacts or omits `ssl.key` on response bodies
- **WHEN** the provider populates state after create, update, or refresh
- **THEN** the provider SHALL preserve the applied plan's `ssl.key` value (or the prior state's value on refresh) instead of overwriting it with the redacted/absent API response
- **AND** on import without a configured `key` in state, `ssl.key` SHALL be null regardless of the API response
