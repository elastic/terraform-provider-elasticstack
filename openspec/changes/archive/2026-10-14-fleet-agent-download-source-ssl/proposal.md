## Why

`elasticstack_fleet_agent_download_source` never sets or reads the Kibana Fleet `ssl` field on
`POST /api/fleet/agent_download_sources` and `PUT /api/fleet/agent_download_sources/{sourceId}`.
Users who point an agent binary download source at a host with a self-signed or privately-issued
certificate (a common air-gapped/on-prem pattern) have no way to supply a trusted CA bundle or a
client certificate/key for mTLS through this resource today, and no opaque/passthrough attribute
covers it. The gap is purely additive: `ssl` is optional on both request bodies, so existing
configurations are unaffected. Elastic docs mark TLS support for agent binary download sources as
generally available since Elastic Stack 9.1.

This closes the gap flagged in #5243.

## What Changes

- Add an optional `ssl` single-nested-object attribute to `elasticstack_fleet_agent_download_source`
  with `certificate_authorities` (list of string, max 10 entries), `certificate` (string), and `key`
  (string, sensitive) — the same nested shape `elasticstack_fleet_output`'s `ssl` block already uses,
  minus `verification_mode` (which the download-source API does not have).
- Map `ssl` to/from the generated `kbapi` `Ssl` struct already present on
  `PostFleetAgentDownloadSourcesJSONBody`, `PutFleetAgentDownloadSourcesSourceidJSONBody`, and the
  read response item — no `generated/kbapi` regeneration required.
- Reject ambiguous empty `ssl` input with validators instead of normalizing it away: an object-level
  "at least one of `certificate`, `certificate_authorities`, `key` must be set" check, plus
  element/length validators on each child (`certificate`/`key` non-empty string,
  `certificate_authorities` 1-10 non-empty entries), so `ssl = {}` and empty-string/empty-list
  children fail at plan time rather than round-tripping to `null`.
- Gate `ssl` with a new attribute-scoped minimum-version requirement (9.1.0), leaving the resource's
  existing overall `8.13.0` floor untouched for configurations that do not set `ssl`.
- Add a `RequiresReplace` plan-modifier fallback on `ssl` (or on specific children, depending on the
  pre-implementation PUT-clearing probe outcome — see design.md) so that any transition the Fleet PUT
  API cannot actually honor in place (clearing a previously-set `ssl` or one of its children) is
  modeled as resource replacement rather than producing a plan/state inconsistency.

## Capabilities

### New Capabilities

<!-- None — this change extends an existing capability only. -->

### Modified Capabilities

- `fleet-agent-download-source`: adds the `ssl` attribute (certificate, certificate authorities,
  client key) to `elasticstack_fleet_agent_download_source`'s schema and CRUD mapping, with its own
  attribute-scoped 9.1.0 version gate, validators rejecting ambiguous-empty input, and a
  `RequiresReplace` fallback for PUT transitions the API cannot honor in place. Narrows the existing
  "Auth and secrets not in v1 schema" requirement so it no longer excludes `ssl`.

## Impact

- **Changed code**: `internal/fleet/agentdownloadsource/schema.go` (new `ssl` attribute + validators +
  plan modifier), `internal/fleet/agentdownloadsource/models.go` (new `ssl`-to-API mapping in
  `toAPICreateModel`/`toAPIUpdateModel`, new attribute-scoped `GetVersionRequirements` entry),
  `internal/fleet/agentdownloadsource/read.go` (hydrate `ssl` from `item.Ssl` on read), a new
  `internal/fleet/agentdownloadsource/models_ssl.go` adapted from
  `internal/fleet/output/models_ssl.go`.
- **New tests**: `internal/fleet/agentdownloadsource/models_ssl_test.go` (unit), additions to
  `internal/fleet/agentdownloadsource/acc_test.go` modeled on
  `internal/fleet/proxy/acc_test.go`'s `TestAccResourceFleetProxy_TLS`.
- **Docs/examples**: regenerate `docs/resources/fleet_agent_download_source.md` and update
  `examples/resources/elasticstack_fleet_agent_download_source/resource.tf` to show the new `ssl`
  block.
- **No `generated/kbapi` changes**: the generated client already exposes `Ssl` on both request bodies
  and the response item.
- **Backward compatibility**: additive only. `ssl` is optional; omitting it from configuration
  produces no behavior change for existing users.
- **Pre-implementation verification required**: a live-Kibana probe of PUT-clearing semantics for
  `ssl` (whole-object and per-child) must run before `toAPIUpdateModel` is finalized — see
  design.md "Open questions" and the first entries of tasks.md.
