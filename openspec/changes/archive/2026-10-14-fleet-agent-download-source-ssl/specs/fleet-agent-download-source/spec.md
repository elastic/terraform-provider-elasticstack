## MODIFIED Requirements

### Requirement: Auth and secrets not in v1 schema

The initial implementation SHALL NOT expose the Fleet download source `auth` and `secrets` fields (API key, username/password, the `secrets.ssl.key` secret-reference variant) as Terraform schema attributes. Those fields remain managed outside this resource's schema for the first version. This exclusion no longer covers the plain `ssl` object (`certificate`, `certificate_authorities`, `key`), which is exposed per the "Optional ssl attribute" requirement below.

#### Scenario: No sensitive auth attributes

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be no Terraform attributes exposing `auth` or the `secrets.auth` / `secrets.ssl.key` material for this resource in v1

#### Scenario: Plain ssl object is exposed

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `ssl` attribute exposing `certificate`, `certificate_authorities`, and `key`

## ADDED Requirements

### Requirement: Optional ssl attribute

The resource SHALL expose an optional `ssl` single-nested-object attribute with the sub-attributes `certificate_authorities` (list of string, optional), `certificate` (string, optional), and `key` (string, optional, sensitive). The provider SHALL map `ssl` to and from the `ssl` object on the Fleet create, update, and read response bodies.

#### Scenario: Create sends ssl fields

- **WHEN** the practitioner sets `ssl.certificate_authorities`, `ssl.certificate`, and/or `ssl.key` on a new resource
- **THEN** the provider SHALL include the corresponding `ssl` object in the create request body

#### Scenario: Update sends changed ssl fields

- **WHEN** the practitioner changes any `ssl` sub-attribute on an existing resource and applies
- **THEN** the provider SHALL include the corresponding `ssl` object in the update request body

#### Scenario: Read hydrates ssl from the API response

- **WHEN** the provider reads an existing download source whose stored `ssl` object has at least one non-null field
- **THEN** state SHALL contain a non-null `ssl` object populated with the response's `certificate`, `certificate_authorities`, and `key` values

#### Scenario: Omitted ssl is null in state

- **WHEN** the practitioner does not set `ssl` and the API response's `ssl` object is absent or has no non-null fields
- **THEN** state SHALL contain a null `ssl` object

#### Scenario: key is marked sensitive

- **WHEN** the practitioner inspects the `ssl.key` attribute in the resource schema
- **THEN** it SHALL be marked `Sensitive`

#### Scenario: No verification_mode sub-attribute

- **WHEN** the practitioner inspects the `ssl` attribute in the resource schema
- **THEN** there SHALL be no `verification_mode` sub-attribute, since the Fleet agent download source API has none

### Requirement: ssl empty-value validation

The resource SHALL reject, at plan time, any configuration where `ssl` is set but ambiguously empty, rather than silently normalizing the ambiguity to null on a later read.

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

When removing a previously-set `ssl` from configuration would produce a plan/state inconsistency because the Fleet PUT API does not honor that removal in place (as determined by a pre-implementation probe against a live Kibana instance), the resource SHALL force replacement for that specific transition via a plan modifier, instead of issuing an in-place update that silently leaves the stale `ssl` value on the server while the plan shows it removed.

#### Scenario: In-place removal when the API honors it

- **GIVEN** the pre-implementation probe determines that omitting `ssl` (or a specific `ssl` child) from the PUT body clears it server-side
- **WHEN** the practitioner removes `ssl` (or that child) from configuration and applies
- **THEN** the provider SHALL perform an in-place update, and a subsequent read SHALL show the removed value as null

#### Scenario: Forced replacement when the API cannot honor in-place removal

- **GIVEN** the pre-implementation probe determines that neither omitting `ssl` nor sending `ssl = {}` clears a previously-set `ssl`, or that omitting a specific child does not clear that child
- **WHEN** the practitioner removes `ssl` (or that child) from configuration and applies
- **THEN** Terraform SHALL plan a replace operation for the resource
