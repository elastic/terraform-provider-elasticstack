## MODIFIED Requirements

### Requirement: Auth and secrets not in v1 schema

The resource SHALL expose an optional `auth` nested attribute mapping to the Fleet API's `auth` request/response field: `username` (string), `password` (string, sensitive), `api_key` (string, sensitive), and `headers` (list of `{key, value}`, up to 100 entries). The resource SHALL expose an optional `ssl` nested attribute per the "Optional ssl attribute" requirement below, including `key`, which the provider SHALL submit through the Fleet `secrets.ssl.key` managed-secret field rather than the plain `ssl.key` field. The resource SHALL NOT expose the `{id}`-reference form of the Fleet download source `secrets.auth.*` or `secrets.ssl.key` fields (referencing a pre-existing Fleet secret store entry by ID, as opposed to submitting a raw value for Fleet to store) as Terraform schema attributes; that reference form remains managed outside this resource's schema.

#### Scenario: Auth attribute present in schema

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `auth` attribute exposing `username`, `password`, `api_key`, and `headers`

#### Scenario: No sensitive auth attributes

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be no Terraform attributes exposing the `{id}`-reference form of `secrets.auth.*` or `secrets.ssl.key` for this resource

#### Scenario: ssl object is exposed, with key routed through the Fleet secrets field

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `ssl` attribute exposing `certificate`, `certificate_authorities`, and `key`
- **AND** the provider SHALL submit a configured `key` value through the Fleet `secrets.ssl.key` field, not the plain `ssl.key` field

## ADDED Requirements

### Requirement: Optional ssl attribute

The resource SHALL expose an optional `ssl` single-nested-object attribute with sub-attributes `certificate_authorities` (list of string, optional), `certificate` (string, optional), and `key` (string, optional, sensitive). The provider SHALL map these to/from the Fleet API per the scenarios below.

#### Scenario: Create sends ssl fields

- **WHEN** the practitioner sets `ssl.certificate_authorities`, `ssl.certificate`, and/or `ssl.key` on a new resource
- **THEN** the provider SHALL include `certificate`/`certificate_authorities` in the create request body's plain `ssl` object
- **AND** the provider SHALL include a configured `key` in the create request body's `secrets.ssl.key` field, not the plain `ssl.key` field

#### Scenario: Update sends changed ssl fields

- **WHEN** the practitioner changes any `ssl` sub-attribute on an existing resource and applies
- **THEN** the provider SHALL include `certificate`/`certificate_authorities` in the update request body's plain `ssl` object
- **AND** the provider SHALL include a changed `key` in the update request body's `secrets.ssl.key` field, not the plain `ssl.key` field

#### Scenario: Read hydrates certificate and certificate_authorities from the API response

- **WHEN** the provider reads an existing download source whose stored `ssl` object has a non-null `certificate` and/or `certificate_authorities`
- **THEN** state SHALL contain a non-null `ssl` object with `certificate`/`certificate_authorities` populated from the response

#### Scenario: Refresh preserves key from prior state rather than the API response

- **WHEN** the provider refreshes an existing download source
- **THEN** state SHALL retain the prior Terraform state's `ssl.key` value, or null if none exists, regardless of what the API response's `ssl.key` or `secrets.ssl.key` fields contain

#### Scenario: Read-after-write preserves the applied key value

- **WHEN** the provider reads a download source after creating or updating it
- **THEN** state SHALL use the applied plan's `ssl.key` value, including null when removed, rather than the pre-write state or either API key field

#### Scenario: Import does not recover key material

- **WHEN** the provider imports an existing download source without a configured key in state
- **THEN** `ssl.key` SHALL be null regardless of either API key field

#### Scenario: Omitted ssl is null in state

- **WHEN** the practitioner does not set `ssl` and the API response's `ssl` object is absent or has no non-null fields, and no `key` is present in prior state
- **THEN** state SHALL contain a null `ssl` object

#### Scenario: key is marked sensitive

- **WHEN** the practitioner inspects the `ssl.key` attribute in the resource schema
- **THEN** it SHALL be marked `Sensitive`

#### Scenario: No verification_mode sub-attribute

- **WHEN** the practitioner inspects the `ssl` attribute in the resource schema
- **THEN** there SHALL be no `verification_mode` sub-attribute, since the Fleet agent download source API has none

### Requirement: ssl empty-value validation

The resource SHALL accept a configuration where `ssl` is set but has empty or omitted children, matching the existing `auth` attribute's permissiveness, rather than rejecting the ambiguity at plan time. The resource SHALL continue to reject a non-empty `certificate_authorities` list that contains an empty-string element.

#### Scenario: Empty ssl object accepted

- **WHEN** the practitioner configures `ssl = {}`
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a validation error

#### Scenario: Empty string certificate or key accepted

- **WHEN** the practitioner configures `ssl.certificate = ""` or `ssl.key = ""`
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a validation error

#### Scenario: Empty certificate_authorities list accepted

- **WHEN** the practitioner configures `ssl.certificate_authorities = []`
- **THEN** Terraform SHALL accept the configuration and SHALL NOT report a validation error

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

The `ssl` attribute SHALL be guarded by its own attribute-scoped minimum Kibana version requirement of 9.1.0, enforced only when `ssl` is configured, independent of and in addition to the resource's overall minimum version guard. Configurations that do not set `ssl` SHALL NOT be affected by this requirement.

#### Scenario: ssl configured against unsupported stack version

- **WHEN** the practitioner configures `ssl` with at least one sub-attribute set
- **AND** the connected Kibana is below version 9.1.0
- **THEN** the provider SHALL emit an attribute-scoped diagnostic on `ssl` that it is not supported on this version

#### Scenario: ssl omitted against old stack version is unaffected

- **WHEN** the practitioner does not configure `ssl`
- **AND** the connected Kibana is below version 9.1.0 but at or above the resource's overall minimum version
- **THEN** the provider SHALL NOT emit a version diagnostic related to `ssl`

### Requirement: ssl removal consistency with PUT semantics

Removing a previously-set `ssl` (in whole or in part) from configuration SHALL always be performed as an in-place update, never a forced replacement. Omitting `ssl` entirely from a PUT body leaves the prior value unchanged server-side, so the provider SHALL send an explicit `ssl: {}` to fully clear it; omitting a single child (`certificate`, `key`, or `certificate_authorities`) from an otherwise-populated `ssl`/`secrets.ssl` object on PUT clears only that child.

#### Scenario: Full removal sends an explicit empty ssl object

- **WHEN** the practitioner removes a previously-set `ssl` entirely from configuration and applies
- **THEN** the provider SHALL send an explicit `ssl: {}` (not an omitted `ssl` field) in the update request body
- **AND** the provider SHALL perform an in-place update, and a subsequent read SHALL show `ssl` as null

#### Scenario: Per-child removal clears only that child

- **WHEN** the practitioner removes a single `ssl` child (`certificate`, `key`, or `certificate_authorities`) from an otherwise-populated `ssl` configuration and applies
- **THEN** the provider SHALL omit only that child from the update request body's `ssl`/`secrets.ssl` object while still sending the other configured children
- **AND** the provider SHALL perform an in-place update, and a subsequent read SHALL show the removed child as null while the other children persist
