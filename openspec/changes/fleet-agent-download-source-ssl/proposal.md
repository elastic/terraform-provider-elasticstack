## Why

`elasticstack_fleet_agent_download_source` never sets or reads the Kibana Fleet `ssl` field on
`POST /api/fleet/agent_download_sources` and `PUT /api/fleet/agent_download_sources/{sourceId}`.
Users who point an agent binary download source at a host with a self-signed or privately-issued
certificate (a common air-gapped/on-prem pattern) have no way to supply a trusted CA bundle or a
client certificate/key for mTLS through this resource today, and no opaque/passthrough attribute
covers it. The gap is purely additive: `ssl` is optional on both request bodies, so existing
configurations are unaffected. Elastic docs mark TLS support for agent binary download sources as
generally available since Elastic Stack 9.1.

This closes the gap flagged in #5243, and — for the `key` sub-attribute specifically — the
`secrets.ssl.key` part of the gap flagged in #5242 (the sibling `secrets.auth.*` part of #5242 is
already closed functionally by the shipped `auth` attribute, #5263: sending plain `auth.password`/
`auth.api_key` already causes Fleet to store them as managed secrets server-side, with no
`secrets.auth.*` request wiring needed).

## What Changes

- Add an optional `ssl` single-nested-object attribute to `elasticstack_fleet_agent_download_source`
  with `certificate_authorities` (list of string, max 10 entries), `certificate` (string), and `key`
  (string, sensitive) — the same nested shape `elasticstack_fleet_output`'s `ssl` block already uses,
  minus `verification_mode` (which the download-source API does not have).
- Map `certificate`/`certificate_authorities` to/from the generated `kbapi` plain `Ssl` struct already
  present on `PostFleetAgentDownloadSourcesJSONBody`, `PutFleetAgentDownloadSourcesSourceidJSONBody`,
  and the read response item. Map `key` to/from the sibling `Secrets.Ssl.Key` union field on the same
  request/response bodies instead of the plain `Ssl.Key` field: on write, send the raw string through
  `secrets.ssl.key` (the generated union's string arm) so Fleet stores it as a managed secret, the same
  outcome `auth.password`/`auth.api_key` already get; on read, `secrets.ssl.key` always comes back as
  `{id}` (never the plaintext value), so `key` is always carried forward from prior state rather than
  derived from the API response. No `generated/kbapi` regeneration required — both the plain `Ssl` and
  the `Secrets.Ssl.Key` union already exist on the generated client.
- Accept `ssl = {}` and empty/omitted children, matching the existing `auth` attribute's permissiveness
  (`auth = {}` is already valid): drop the object-level "at least one of `certificate`,
  `certificate_authorities`, `key` must be set" validator and the per-field non-empty-string/
  non-empty-list validators on `certificate`, `key`, and the `certificate_authorities` list itself.
  Keep only `certificate_authorities`'s `maxItems: 10` validator and its element-level non-empty-string
  validator (an explicitly-provided list still can't contain an empty-string entry).
- Gate `ssl` with a new attribute-scoped minimum-version requirement (9.1.0), leaving the resource's
  existing overall `8.13.0` floor and `auth`'s own `9.4.0` attribute-scoped gate untouched.
- Removal sends an explicit `ssl: {}` on update (never a nil/omitted `ssl`, since omitting leaves the
  prior value unchanged server-side); per-child removal is encoded as an omitted field within an
  otherwise-populated `ssl`/`secrets.ssl` object, which Fleet clears in place. No `RequiresReplace`
  plan modifier is needed for any `ssl` transition.

## Capabilities

### New Capabilities

<!-- None — this change extends an existing capability only. -->

### Modified Capabilities

- `fleet-agent-download-source`: adds the `ssl` attribute (certificate, certificate authorities,
  client key) to `elasticstack_fleet_agent_download_source`'s schema and CRUD mapping, with its own
  attribute-scoped 9.1.0 version gate and validators permitting ambiguous-empty input (matching
  `auth`). Wires `ssl.key` through the Fleet `secrets.ssl.key` managed-secret field instead of the
  plain `ssl.key` field. Narrows the existing "Auth and secrets not in v1 schema" requirement so it no
  longer excludes `ssl`, while continuing to exclude the `{id}`-reference form of `secrets.auth`/
  `secrets.ssl.key` (referencing a pre-existing Fleet secret by ID, as opposed to submitting a raw
  value for Fleet to store).

## Impact

- **Changed code**: `internal/fleet/agentdownloadsource/schema.go` (new `ssl` attribute + validators),
  `internal/fleet/agentdownloadsource/models.go` (new `ssl`-to-API mapping in
  `toAPICreateModel`/`toAPIUpdateModel`, routing `key` through `Secrets.Ssl.Key`; new attribute-scoped
  `GetVersionRequirements` entry), `internal/fleet/agentdownloadsource/read.go` (hydrate
  `certificate`/`certificate_authorities` from `item.Ssl`, and `key` by carrying the prior state value
  forward since the API never echoes it), a new `internal/fleet/agentdownloadsource/models_ssl.go`
  adapted from `internal/fleet/output/models_ssl.go`.
- **New tests**: `internal/fleet/agentdownloadsource/models_ssl_test.go` (unit), additions to
  `internal/fleet/agentdownloadsource/acc_test.go` modeled on
  `internal/fleet/proxy/acc_test.go`'s `TestAccResourceFleetProxy_TLS`.
- **Docs/examples**: regenerate `docs/resources/fleet_agent_download_source.md` and update
  `examples/resources/elasticstack_fleet_agent_download_source/resource.tf` to show the new `ssl`
  block.
- **No `generated/kbapi` changes**: the generated client already exposes the plain `Ssl` struct and the
  `Secrets.Ssl.Key` union on both request bodies and the response item.
- **Backward compatibility**: additive only. `ssl` is optional; omitting it from configuration
  produces no behavior change for existing users.
