# Delta spec: `fleet-agent-download-source` — `auth` attribute

Extends the main spec at `openspec/specs/fleet-agent-download-source/spec.md`.

## Schema refinements

Add an `auth` block to the canonical schema:

```hcl
resource "elasticstack_fleet_agent_download_source" "example" {
  # ... existing attributes unchanged ...

  auth = {                               # optional, single nested object
    username = "svc-account"             # optional, string; requires password
    password = "..."                     # optional, string, sensitive; requires username; conflicts with api_key
    api_key  = "..."                     # optional, string, sensitive; conflicts with password
    headers = [                          # optional, list of { key, value }; max 100 entries
      { key = "X-Custom-Header", value = "value" },
    ]
  }
}
```

## MODIFIED Requirements

### Requirement: Auth and secrets not in v1 schema

The resource SHALL expose an optional `auth` nested attribute mapping to the Fleet API's `auth` request/response field: `username` (string), `password` (string, sensitive), `api_key` (string, sensitive), and `headers` (list of `{key, value}`, up to 100 entries). The resource SHALL NOT expose the Fleet download source `secrets.auth` reference mechanism (referencing an existing Fleet secret store entry by ID) or the top-level `ssl` object (CA, client cert, client key material) as Terraform schema attributes; those remain managed outside this resource's schema.

#### Scenario: Auth attribute present in schema

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `auth` attribute exposing `username`, `password`, `api_key`, and `headers`

#### Scenario: No sensitive auth attributes

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be no Terraform attributes exposing the `secrets.auth` `{id}` reference mechanism or the `ssl` object (CA, client cert, client key material) for this resource

## ADDED Requirements

### Requirement: Auth credential-mode validation

The `auth` attribute's `username` and `password` SHALL be set together (each requires the other), and `password` SHALL conflict with `api_key` (mutually exclusive). `headers` SHALL carry no dependency on `username`, `password`, or `api_key`, so an `auth` block containing only `headers` SHALL be valid, and so SHALL an empty `auth = {}` block.

#### Scenario: Username without password rejected

- **WHEN** `auth = { username = "svc" }` is configured with no `password`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Password without username rejected

- **WHEN** `auth = { password = "secret" }` is configured with no `username`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Password and api_key together rejected

- **WHEN** `auth` sets both `password` and `api_key`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Headers-only auth block accepted

- **WHEN** `auth = { headers = [{ key = "X-A", value = "1" }] }` is configured with no credentials
- **THEN** Terraform SHALL accept the configuration and send the request with only `headers` populated under `auth`

#### Scenario: Empty auth object accepted

- **WHEN** `auth = {}` is configured
- **THEN** Terraform SHALL accept the configuration

### Requirement: Auth headers size limit

The `headers` list within `auth` SHALL accept at most 100 entries.

#### Scenario: Header count at limit

- **WHEN** `auth.headers` is configured with exactly 100 entries
- **THEN** Terraform SHALL accept the configuration

#### Scenario: Header count over limit

- **WHEN** `auth.headers` is configured with 101 entries
- **THEN** Terraform SHALL report a validation error before any API call

### Requirement: Auth mapped on create and update

When `auth` is present in configuration, the create and update requests SHALL include the corresponding `auth` object with `username`, `password`, `api_key`, and `headers` populated from configuration. When `auth` is absent from configuration on create, the create request SHALL omit `auth`.

#### Scenario: Create includes auth

- **WHEN** the practitioner creates a resource with `auth` configured
- **THEN** the create request body SHALL include an `auth` object matching the configured fields

#### Scenario: Create omits auth when unset

- **WHEN** the practitioner creates a resource without configuring `auth`
- **THEN** the create request body SHALL NOT include an `auth` key

### Requirement: Auth removal clears credentials on update

When the planned configuration omits `auth` but the prior state had `auth` configured, the update request SHALL send an explicit empty `auth` object (`"auth":{}`) rather than omitting the key, so Fleet clears the existing credentials and headers instead of merging. See scenarios below for the omit-when-never-configured and full-replacement-on-change cases.

#### Scenario: Removing the auth block clears it server-side

- **GIVEN** a resource with `auth.username`/`auth.password` configured and applied
- **WHEN** the practitioner removes the `auth` block entirely and applies
- **THEN** the update request body SHALL include `"auth":{}`
- **AND** a subsequent read SHALL show no credentials or headers configured

#### Scenario: Omitting auth from an update that never configured it is a no-op for auth

- **GIVEN** a resource created without `auth`
- **WHEN** the practitioner updates an unrelated attribute (e.g. `name`) without ever configuring `auth`
- **THEN** the update request body SHALL NOT include an `auth` key

#### Scenario: Setting a new credential replaces the full credential set

- **GIVEN** a resource with `auth.username`/`auth.password` configured
- **WHEN** the practitioner changes configuration to `auth.api_key` only and applies
- **THEN** the update request body SHALL include an `auth` object containing only `api_key` (and any configured `headers`), and the prior `username`/`password` SHALL no longer be present server-side

### Requirement: Auth password and api_key are never echoed on read

Kibana never returns `auth.password` or `auth.api_key` in any response; it stores the value in its secret store and returns only an opaque, unmapped reference. When the API omits either field, the resource SHALL preserve it from a fallback instead of writing null. `username`/`headers` always come from the API directly. See scenarios for the per-path fallback source.

#### Scenario: Password survives the post-create read

- **WHEN** the practitioner creates a resource with `auth.password` configured
- **THEN** the mandatory read-after-create SHALL populate state's `auth.password` with the configured value, not null, even though the create response omits it

#### Scenario: Rotated password survives the post-update read

- **GIVEN** a resource with `auth.password` already set
- **WHEN** the practitioner changes `auth.password` to a new value and applies
- **THEN** the mandatory read-after-update SHALL populate state's `auth.password` with the newly configured value, not the prior value and not null

#### Scenario: Refresh preserves the previously known password

- **GIVEN** state already has `auth.password` populated from a prior apply
- **WHEN** Terraform performs a plain refresh (no configuration change)
- **THEN** state's `auth.password` SHALL remain the previously known value, since the API read omits it

#### Scenario: Username and headers always come from the API

- **WHEN** the resource is read after create, update, or during refresh
- **THEN** state's `auth.username` and `auth.headers` SHALL reflect the API response directly, not a preserved/fallback value

### Requirement: Auth headers null vs empty list distinction

When the API response includes `auth.headers` as an empty list, state SHALL reflect an empty list, not null. When the API response omits `auth.headers` entirely (or omits `auth` entirely), state SHALL reflect null for `headers` (and for the whole `auth` object, respectively).

#### Scenario: Configured empty headers list round-trips as empty, not null

- **WHEN** the practitioner configures `auth = { headers = [] }`
- **THEN** state's `auth.headers` SHALL be an empty list after apply, not null

### Requirement: Auth minimum Kibana version

When `auth` is configured (non-null), the provider SHALL enforce a minimum Kibana version of 9.4.0 for the `auth` attribute, independent of the resource's overall minimum version. When `auth` is not configured, the resource's existing overall minimum version requirement applies unchanged.

#### Scenario: Auth rejected below minimum version

- **WHEN** the connected Kibana is below 9.4.0
- **AND** the practitioner configures `auth` on this resource
- **THEN** the provider SHALL emit a diagnostic that `auth` is not supported on this version

#### Scenario: Resource without auth still works below the auth minimum version

- **WHEN** the connected Kibana is at or above the resource's overall minimum version but below 9.4.0
- **AND** the practitioner's configuration omits `auth`
- **THEN** the provider SHALL NOT emit an `auth`-related version diagnostic

### Requirement: Import does not populate auth secrets

On `terraform import`, the resource SHALL NOT populate `auth.password` or `auth.api_key` from any prior Terraform state, since none exists for a freshly imported resource. The subsequent read cycle SHALL leave those two fields null whenever the API omits them (which it always does), regardless of whether the imported download source was originally created through Terraform or the Kibana UI. `auth.username` and `auth.headers` SHALL populate normally from the API response on import.

#### Scenario: Imported resource has null auth secrets

- **WHEN** the practitioner imports a download source that has `auth.password` configured server-side
- **THEN** state's `auth.password` SHALL be null after import
- **AND** state's `auth.username` and `auth.headers` (if configured server-side) SHALL be populated from the API response
