# Delta spec: `fleet-agent-download-source` — `secrets` write-only attributes

Extends the main spec at `openspec/specs/fleet-agent-download-source/spec.md`.

## Schema refinements

Add `_wo` siblings to the `ssl` and `auth` blocks:

```hcl
resource "elasticstack_fleet_agent_download_source" "example" {
  # ... existing attributes unchanged ...

  ssl = {
    certificate = "..."
    key         = "..."    # existing plain attribute, unchanged; now wired to secrets.ssl.key server-side
    key_wo      = "..."    # new, optional, string, write-only + sensitive; conflicts with key is NOT enforced (either may be used)
  }

  auth = {
    username     = "svc-account"
    password     = "..."   # existing; requires username or; conflicts with password_wo
    password_wo  = "..."   # new, optional, string, write-only + sensitive; requires username; conflicts with password/api_key family
    api_key_wo   = "..."   # new, optional, string, write-only + sensitive; conflicts with password family and api_key
  }
}
```

## MODIFIED Requirements

### Requirement: Auth and secrets not in v1 schema

The resource SHALL expose an optional `auth` nested attribute mapping to the Fleet API's `auth` request/response field: `username` (string), `password` (string, sensitive), `password_wo` (string, write-only, sensitive), `api_key` (string, sensitive), `api_key_wo` (string, write-only, sensitive), and `headers` (list of `{key, value}`, up to 100 entries). The resource SHALL expose an optional `key_wo` (string, write-only, sensitive) sibling on the `ssl` nested attribute alongside the existing plain `key`. The resource SHALL NOT expose the Fleet download source `secrets.auth`/`secrets.ssl.key` `{id}`-reference mechanism itself as a user-facing Terraform schema shape (e.g. no attribute through which a practitioner supplies or reads a raw secret id); the provider uses that mechanism internally to avoid resending unchanged secret literals, per the "Write-only secret reference reuse" requirement below.

#### Scenario: Auth attribute present in schema

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `auth` attribute exposing `username`, `password`, `password_wo`, `api_key`, `api_key_wo`, and `headers`

#### Scenario: No sensitive auth attributes

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be no Terraform attribute through which a practitioner directly reads or sets a Fleet secret-store id for `auth`

#### Scenario: Plain ssl object is exposed

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `ssl` attribute exposing `certificate`, `certificate_authorities`, `key`, and `key_wo`

#### Scenario: Write-only attributes are marked WriteOnly and Sensitive

- **WHEN** the practitioner inspects `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` in the resource schema
- **THEN** each SHALL be marked both `WriteOnly` and `Sensitive`

### Requirement: Auth credential-mode validation

The `auth` attribute SHALL be validated by a single object-level validator enforcing: a configured `username` requires `password` or `password_wo` (at least one); `password` and `password_wo` are mutually exclusive; `api_key` and `api_key_wo` are mutually exclusive; any configured member of the password family (`password`, `password_wo`) conflicts with any configured member of the api_key family (`api_key`, `api_key_wo`). `headers` SHALL carry no dependency on any credential attribute, so an `auth` block containing only `headers` SHALL remain valid, and so SHALL an empty `auth = {}` block.

#### Scenario: Username without password rejected

- **WHEN** `auth = { username = "svc" }` is configured with neither `password` nor `password_wo`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Password without username rejected

- **WHEN** `auth = { password = "secret" }` or `auth = { password_wo = "secret" }` is configured with no `username`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Password and api_key together rejected

- **WHEN** `auth` sets both a password-family attribute (`password` or `password_wo`) and an api_key-family attribute (`api_key` or `api_key_wo`)
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Headers-only auth block accepted

- **WHEN** `auth = { headers = [{ key = "X-A", value = "1" }] }` is configured with no credentials
- **THEN** Terraform SHALL accept the configuration and send the request with only `headers` populated under `auth`

#### Scenario: Empty auth object accepted

- **WHEN** `auth = {}` is configured
- **THEN** Terraform SHALL accept the configuration

#### Scenario: Username satisfied by write-only password

- **WHEN** `auth = { username = "svc", password_wo = "secret" }` is configured
- **THEN** Terraform SHALL accept the configuration

#### Scenario: password and password_wo together rejected

- **WHEN** `auth` sets both `password` and `password_wo`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: api_key and api_key_wo together rejected

- **WHEN** `auth` sets both `api_key` and `api_key_wo`
- **THEN** Terraform SHALL report a validation error before any API call

## ADDED Requirements

### Requirement: Write-only secret attributes detect drift via writeonlyhash

The resource SHALL detect in-config edits to `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` using the shared `internal/utils/writeonlyhash` helper in a `ModifyPlan` implementation, storing a bcrypt hash of each applied value in Terraform private state and comparing the configured value's hash against the stored hash on every plan. The resource SHALL NOT expose or rely on a `_wo_version` companion attribute for any of the three write-only attributes.

#### Scenario: Unchanged write-only value produces no planned update

- **GIVEN** `ssl.key_wo` was set and applied in a prior apply
- **WHEN** the practitioner runs `terraform plan` with `ssl.key_wo` unchanged in configuration
- **THEN** the plan SHALL show no change attributable to `ssl.key_wo`

#### Scenario: Changed write-only value is detected without a version attribute

- **GIVEN** `auth.password_wo` was set and applied in a prior apply
- **WHEN** the practitioner changes `auth.password_wo`'s configured value and runs `terraform plan`
- **THEN** the plan SHALL show a planned update, without the practitioner having configured any `_wo_version`-style attribute

#### Scenario: Deleting the resource clears stored hashes

- **GIVEN** a resource with one or more write-only attributes applied
- **WHEN** the practitioner destroys the resource
- **THEN** the private-state hash entries for `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` SHALL be cleared

### Requirement: Write-only secret reference reuse avoids clearing managed secrets

When writing a configured write-only attribute whose hash is unchanged since the last successful apply, the provider SHALL query the current resource by `source_id` immediately before the update and, if Fleet reports an existing secret id for that field, reference it by id rather than resending the literal value. When the hash differs, or Fleet reports no existing secret id for that field, the provider SHALL send the literal value instead.

#### Scenario: Unchanged write-only value is referenced by id, not resent

- **GIVEN** `ssl.key_wo` was set, applied, and Fleet currently holds a managed secret id for it
- **WHEN** the practitioner updates an unrelated attribute (e.g. `name`) without changing `ssl.key_wo`
- **THEN** the update request SHALL reference the existing secret by id for `secrets.ssl.key` rather than resending the literal value

#### Scenario: Changed write-only value is resent as a literal

- **GIVEN** `auth.api_key_wo` was set and applied
- **WHEN** the practitioner changes `auth.api_key_wo`'s configured value and applies
- **THEN** the update request SHALL send the new literal value under `secrets.auth.api_key`, not an id reference

#### Scenario: Secret-storage-unavailable fallback resends the literal

- **GIVEN** a configured write-only attribute whose value Fleet could not store as a managed secret (storage gating unmet server-side)
- **WHEN** the provider writes an update that would otherwise reference the secret by id
- **THEN** the provider SHALL resend the literal value instead of attempting an id reference

### Requirement: Secret-bearing fields are never silently omitted from a write

The provider SHALL NOT omit a configured secret-bearing field's value from a create or update request while still sending its surrounding `ssl`/`auth` object, since Fleet treats that omission as a clear of the existing managed secret rather than a no-op.

#### Scenario: No value is ever silently dropped from a write

- **WHEN** the provider builds any create or update request that includes the `ssl` or `auth` object
- **THEN** every configured write-only or plain secret-bearing field within that object SHALL be represented in the request as either a literal value or an id reference, never omitted

### Requirement: Import does not populate write-only secret attributes

On `terraform import`, the resource SHALL NOT populate `ssl.key_wo`, `auth.password_wo`, or `auth.api_key_wo` from any prior Terraform state or private-state hash, since none exists for a freshly imported resource.

#### Scenario: Imported resource has null write-only secret attributes

- **WHEN** the practitioner imports a download source with `ssl.key`/`auth.password`/`auth.api_key` configured server-side (via the plain wire fields or a managed secret)
- **THEN** state's `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` SHALL all be null after import
