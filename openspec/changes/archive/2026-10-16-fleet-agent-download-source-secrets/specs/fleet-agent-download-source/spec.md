# Delta spec: `fleet-agent-download-source` — `secrets` write-only attributes

Extends the main spec at `openspec/specs/fleet-agent-download-source/spec.md`.

## Schema refinements

Add `_wo` siblings to the `ssl` and `auth` blocks:

```hcl
resource "elasticstack_fleet_agent_download_source" "example" {
  # ... existing attributes unchanged ...

  ssl = {
    certificate = "..."
    key         = "..."    # existing plain attribute; now wired to secrets.ssl.key server-side
    key_wo      = "..."    # new, optional, string, write-only + sensitive; conflicts with key
  }

  auth = {
    username     = "svc-account" # requires either password form
    password     = "..."          # requires username; conflicts with password_wo and api_key family
    password_wo  = "..."          # new, optional, string, write-only + sensitive; requires username
    api_key_wo   = "..."          # new, optional, string, write-only + sensitive; conflicts with api_key and password family
  }
}
```

## MODIFIED Requirements

### Requirement: Auth and secrets not in v1 schema

The resource SHALL expose an optional `auth` nested attribute mapping to the Fleet API's `auth` request/response field: `username` (string), `password` (string, sensitive), `password_wo` (string, write-only, sensitive), `api_key` (string, sensitive), `api_key_wo` (string, write-only, sensitive), and `headers` (list of `{key, value}`, up to 100 entries). The resource SHALL expose an optional `key_wo` (string, write-only, sensitive) sibling on the `ssl` nested attribute alongside the existing plain `key`. Both `ssl.key` and `ssl.key_wo` SHALL map to `secrets.ssl.key`; they SHALL be mutually exclusive because they represent the same API value. The resource SHALL NOT expose the Fleet download source `secrets.auth`/`secrets.ssl.key` `{id}`-reference mechanism itself as a user-facing Terraform schema shape (e.g. no attribute through which a practitioner supplies or reads a raw secret id); the provider uses that mechanism internally to avoid resending unchanged secret literals, per the "Write-only secret reference reuse" requirement below.

#### Scenario: Auth attribute present in schema

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `auth` attribute exposing `username`, `password`, `password_wo`, `api_key`, `api_key_wo`, and `headers`

#### Scenario: No sensitive auth attributes

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be no Terraform attribute through which a practitioner directly reads or sets a Fleet secret-store id for `auth`

#### Scenario: Plain ssl object is exposed

- **WHEN** the practitioner inspects the resource schema
- **THEN** there SHALL be an `ssl` attribute exposing `certificate`, `certificate_authorities`, `key`, and `key_wo`

#### Scenario: SSL key forms are mutually exclusive

- **WHEN** the practitioner configures both `ssl.key` and `ssl.key_wo`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Write-only attributes are marked WriteOnly and Sensitive

- **WHEN** the practitioner inspects `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` in the resource schema
- **THEN** each SHALL be marked both `WriteOnly` and `Sensitive`

### Requirement: Auth credential-mode validation

The `auth` attribute SHALL be validated by a single object-level validator enforcing: `username` and a password form (`password` or `password_wo`) SHALL either both be configured or both be absent; `password` and `password_wo` are mutually exclusive; `api_key` and `api_key_wo` are mutually exclusive; any configured member of the password family (`password`, `password_wo`) conflicts with any configured member of the api_key family (`api_key`, `api_key_wo`). `headers` SHALL carry no dependency on any credential attribute, so an `auth` block containing only `headers` SHALL remain valid, and so SHALL an empty `auth = {}` block.

#### Scenario: Username without password rejected

- **WHEN** `auth = { username = "svc" }` is configured with neither `password` nor `password_wo`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Password without username rejected

- **WHEN** `auth = { password = "secret" }` or `auth = { password_wo = "secret" }` is configured with no `username`
- **THEN** Terraform SHALL report a validation error before any API call

#### Scenario: Username and password are both required

- **WHEN** `auth` configures `username` without either password form, or configures either password form without `username`
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

#### Scenario: SSL key_wo is the only SSL attribute

- **WHEN** `ssl = { key_wo = "secret" }` is configured
- **THEN** Terraform SHALL accept the configuration and retain the `ssl` object

## ADDED Requirements

### Requirement: Write-only secret attributes detect drift via writeonlyhash

The resource SHALL detect in-config edits and removals of `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` using the shared `internal/utils/writeonlyhash` helper in a `ModifyPlan` implementation, storing a bcrypt hash of each applied value in Terraform private state and comparing the configured value's hash against the stored hash on every plan. A previously stored hash paired with a now-null configured value SHALL force an update to clear the remote secret. The provider SHALL retain that hash until the clearing update succeeds, then remove it; re-adding the attribute after a successful removal SHALL be treated as a new literal write. The resource SHALL NOT expose or rely on a `_wo_version` companion attribute for any of the three write-only attributes.

#### Scenario: Unchanged write-only value produces no planned update

- **GIVEN** `ssl.key_wo` was set and applied in a prior apply
- **WHEN** the practitioner runs `terraform plan` with `ssl.key_wo` unchanged in configuration
- **THEN** the plan SHALL show no change attributable to `ssl.key_wo`

#### Scenario: Changed write-only value is detected without a version attribute

- **GIVEN** `auth.password_wo` was set and applied in a prior apply
- **WHEN** the practitioner changes `auth.password_wo`'s configured value and runs `terraform plan`
- **THEN** the plan SHALL show a planned update, without the practitioner having configured any `_wo_version`-style attribute

#### Scenario: Removing a write-only value plans a clearing update

- **GIVEN** a write-only attribute has a stored private-state hash
- **WHEN** the practitioner removes that attribute from configuration without changing other attributes
- **THEN** the plan SHALL show an update that clears the corresponding remote secret

#### Scenario: Re-adding a removed write-only value sends a literal

- **GIVEN** a write-only attribute was removed and its clearing update succeeded
- **WHEN** the practitioner configures that attribute again
- **THEN** the provider SHALL treat it as a new value and send its literal

#### Scenario: Deleting the resource clears stored hashes

- **GIVEN** a resource with one or more write-only attributes applied
- **WHEN** the practitioner destroys the resource
- **THEN** the private-state hash entries for `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` SHALL be cleared

### Requirement: Write-only secret reference reuse avoids clearing managed secrets

When writing a configured write-only attribute whose hash is unchanged since the last successful apply, the provider SHALL query the current resource by `source_id` immediately before the update and, if Fleet reports an existing secret id for that field, reference it by id rather than resending the literal value. When the hash differs, on the first write, or Fleet reports no existing secret id for that field, the provider SHALL send the literal value instead. The write path SHALL read write-only literals from configuration, not from the plan, where write-only values are null.

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

#### Scenario: Plain SSL key uses the secret API field

- **WHEN** the practitioner configures `ssl.key`
- **THEN** the provider SHALL send its literal under `secrets.ssl.key`, rather than the plain `ssl.key` API field

#### Scenario: Removed write-only value clears remote secret

- **GIVEN** a write-only value was previously applied and is now absent from configuration
- **WHEN** the provider performs the planned clearing update
- **THEN** the provider SHALL clear the corresponding remote secret and remove its private-state hash only after the update succeeds

### Requirement: Secret-bearing fields are never silently omitted from a write

The provider SHALL NOT omit a configured secret-bearing field's value from a create or update request while still sending its surrounding `ssl`/`auth` object, since Fleet treats that omission as a clear of the existing managed secret rather than a no-op. A deliberate removal of a previously configured write-only value is the exception: it SHALL be represented as an intentional clearing update, not mistaken for an unchanged null plan value.

#### Scenario: Configured secret values are never silently dropped from a write

- **WHEN** the provider builds any create or update request that includes the `ssl` or `auth` object
- **THEN** every configured write-only or plain secret-bearing field within that object SHALL be represented in the request as either a literal value or an id reference, never omitted

### Requirement: Read hydration protects secret values and configured plain keys

The provider SHALL never copy a value originating from a write-only attribute into its stateful plain sibling (`ssl.key`, `auth.password`, or `auth.api_key`). If the API redacts a configured plain secret on read, the provider SHALL preserve the configured plain value rather than replacing it with null. When `ssl.key_wo` is configured without another SSL child, the provider SHALL retain the parent `ssl` object in state. The provider SHALL use configuration/prior state as the source for these distinctions; an echoed value from Fleet alone SHALL NOT cause a write-only literal to be persisted.

#### Scenario: Write-only SSL key is never copied to plain key

- **GIVEN** `ssl.key_wo` was configured and Fleet echoes the stored value as `ssl.key` because secret storage is unavailable
- **WHEN** the provider reads the resource after write or during refresh
- **THEN** state SHALL keep `ssl.key_wo` write-only/null as appropriate and SHALL NOT populate the plain `ssl.key` with the write-only literal

#### Scenario: Redacted configured plain key is preserved

- **GIVEN** `ssl.key` was configured and the API redacts its value on read
- **WHEN** the provider reads the resource after write or during refresh
- **THEN** state SHALL preserve the configured `ssl.key` value instead of replacing it with null

#### Scenario: Key_wo-only SSL object remains present

- **GIVEN** `ssl.key_wo` is configured with no certificate or certificate authorities
- **WHEN** the provider reads the resource after write or during refresh
- **THEN** state SHALL retain a non-null `ssl` object without exposing the key in the plain `key` attribute

### Requirement: Import does not populate write-only secret attributes

On `terraform import`, the resource SHALL NOT populate `ssl.key_wo`, `auth.password_wo`, or `auth.api_key_wo` from any prior Terraform state or private-state hash, since none exists for a freshly imported resource.

#### Scenario: Imported resource has null write-only secret attributes

- **WHEN** the practitioner imports a download source with `ssl.key`/`auth.password`/`auth.api_key` configured server-side (via the plain wire fields or a managed secret)
- **THEN** state's `ssl.key_wo`, `auth.password_wo`, and `auth.api_key_wo` SHALL all be null after import
