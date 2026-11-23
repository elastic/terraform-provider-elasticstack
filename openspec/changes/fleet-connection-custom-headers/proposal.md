## Why

The `kibana { headers = {...} }` block (added in
[`kibana-connection-custom-headers`](../kibana-connection-custom-headers/)) lets the provider send
custom HTTP headers (e.g. `CF-Access-Client-Id`/`CF-Access-Client-Secret` for a Cloudflare
Access-protected Kibana) on outgoing Kibana API requests. Fleet resources (for example
`elasticstack_fleet_integration`) never receive them: `toFleetConfig()` in
`internal/clients/config/kibana_oapi.go` explicitly clears `Headers` when deriving the Fleet client
configuration from the effective Kibana configuration, because Fleet's endpoint can diverge from
Kibana's and forwarding a Kibana-proxy header to a different host by default would be a credential-leak
risk. Any user who puts Kibana and Fleet behind a header-authenticating reverse proxy gets working
Kibana resources and a failing Fleet resource
(`Error: Failed to parse response ... API returned success status but response body was nil or not
JSON`, because the proxy serves its HTML login page instead of forwarding to Fleet)
([#5510](https://github.com/elastic/terraform-provider-elasticstack/issues/5510)). The original design
for the Kibana `headers` attribute already anticipated this and listed a Fleet `headers` attribute as a
non-blocking follow-up.

## What Changes

- Add an optional map-of-string `headers` attribute to the provider's `fleet {}` connection block
  (`GetFleetFWConnectionBlock()` in `internal/schema/connection.go`), mirroring the exact contract
  already used by `kibana { headers }`: `Optional: true`, `Sensitive: true`, `ElementType:
  types.StringType`, no schema-level validators. The `fleet {}` block is hand-written (not generated
  from `kbConnectionBlockSpec()`/`connectionBlockSpec()`), and there is no entity-local
  `fleet_connection` block anywhere in the provider, so this is the only schema surface that needs the
  new attribute.
- Add `Headers types.Map` (`tfsdk:"headers"`) to `FleetConnection` in
  `internal/clients/config/provider.go`.
- Extract the reserved-header-name and case-insensitive-duplicate validation that today lives inline in
  `buildKibanaOapiConfigFromFramework` (`internal/clients/config/kibana_oapi.go`) into a shared helper,
  and reuse it, unchanged, for `fleet.headers`. Fleet's `headers` attribute is rejected by the exact
  same reserved-name list as Kibana's (`authorization`, `content-length`, `host`, `kbn-xsrf`, `trailer`,
  `transfer-encoding`) — per explicit human direction on the issue, Fleet does not get its own,
  possibly-different, reserved-name list.
- Decode and validate `fleet.headers` in `newFleetConfigFromFramework`
  (`internal/clients/config/fleet.go`), assigning the result to the Fleet client's `Config.Headers`
  after the existing `kibanaCfg.toFleetConfig()` call that clears any inherited Kibana headers. Fleet's
  own `headers` value is therefore independent of, and never influenced by, `kibana.headers`.
- Trim leading/trailing whitespace from accepted Fleet header names and values before use, matching the
  existing Kibana and Elasticsearch header behavior.
- No part of this change forwards `kibana.headers` to Fleet automatically, even when the Fleet endpoint
  resolves to the same host as Kibana. Per explicit human direction on the issue, this "same-host
  auto-forward" convenience default (research Approach B) is out of scope and is not implemented as
  part of, or alongside, this change. The existing `toFleetConfig()` exclusion of Kibana-only headers
  from Fleet is unchanged.
- No Fleet/Kibana/Stack minimum-version gate is added. Like the Kibana `headers` attribute, this is a
  client-side HTTP transport concern with no identified API-surface version dependency; the new
  acceptance coverage runs across the provider's standard supported-Stack-version test matrix rather
  than a single pinned version.

## Capabilities

### New Capabilities

- `provider-fleet-connection`: The Fleet connection block's own `headers` attribute — its schema
  contract, decoding, shared reserved-name/case-collision validation, delivery on outgoing Fleet
  requests, and independence from `kibana.headers` (no automatic inheritance in either direction).

### Modified Capabilities

- None. `provider-kibana-connection`'s existing requirements (including the Kibana-headers-excluded-
  from-Fleet exclusion in `toFleetConfig()`) are unchanged in behavior and are not amended by this
  change. Per the human direction on the issue, this ships as a fresh spec (a new capability) rather
  than a delta to `provider-kibana-connection`.

## Impact

- **Changed code**: `internal/schema/connection.go` (`GetFleetFWConnectionBlock`: new `headers`
  attribute), `internal/clients/config/provider.go` (`FleetConnection.Headers`),
  `internal/clients/config/kibana_oapi.go` (extract the reserved-name/case-collision validation into a
  shared, exported-within-package helper), `internal/clients/config/fleet.go`
  (`newFleetConfigFromFramework`: decode, validate, and assign `fleet.headers`).
- **New tests**: unit tests extending `internal/clients/config/fleet_test.go` (omitted/null/empty-map
  all leave `fleetCfg.Headers` nil; a configured value reaches `fleetCfg.Headers`; reserved-name
  rejection and case-insensitive duplicate rejection behave identically to the existing Kibana-header
  tests; `kibana.headers` and `fleet.headers` stay independent of each other); an acceptance test
  alongside `TestFleetConfiguration_excludesKibanaHeaders` (`provider/acc_test.go`) that asserts a
  configured `fleet.headers` value is actually observed on outgoing Fleet HTTP requests, using the same
  env-isolation pattern (`PreferConfiguredKibanaEndpointEnvVar`, explicit `FLEET_ENDPOINT` handling) as
  the existing exclusion test.
- **Docs**: the generated `fleet` provider-block documentation gains a `headers` attribute description.
- **Out of scope**: automatic same-host header forwarding from Kibana to Fleet (research Approach B,
  explicitly rejected by human direction); any change to the Kibana `headers` attribute, its validation,
  or its redirect protection; an entity-local `fleet_connection` block (none exists); any change to
  generated Fleet/Kibana client code or the shared `transport.RoundTrip` header-injection logic (it
  already injects whatever `Config.Headers` it is given, for both Kibana and Fleet clients, including
  the existing cross-origin-redirect guard that activates whenever `Config.Headers` is non-empty).
- **Backward compatibility**: purely additive. `fleet.headers` is `Optional` inside the existing
  `fleet {}` block (already capped at one entry by `listvalidator.SizeAtMost(1)`), so omitting it is a
  no-op for every existing configuration and leaves `fleetCfg.Headers` `nil`, exactly as today.
