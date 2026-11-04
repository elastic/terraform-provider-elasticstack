## Why

Kibana is frequently placed behind an identity-aware proxy (Cloudflare Access, Google IAP) that
requires a custom HTTP header on every request (`cf-access-token: ...`, `Proxy-Authorization: ...`)
before Kibana itself is ever reached. The provider's `kibana` connection block has no way to attach
arbitrary headers to outgoing Kibana API requests, so these setups cannot authenticate the Terraform
provider at all ([#440](https://github.com/elastic/terraform-provider-elasticstack/issues/440), with
corroborating reports from `@hayk96` and `@hossamdash`). The Elasticsearch connection block already
solves this with a `headers` map attribute; Kibana has no equivalent.

## What Changes

- Add a `headers` map-of-string attribute to the shared Kibana connection block (`kbConnectionBlockSpec()`),
  mirroring the existing Elasticsearch `headers` attribute (`sensitive: true`, `writeOnly: true`). Because
  the Kibana connection block is defined once and shared across the provider-level `kibana` block and
  every entity-local `kibana_connection` block, the new attribute becomes available everywhere that
  block already appears — no per-entity schema changes required.
- Thread `headers` through `kibanaoapi.Config` and inject configured header values on every outgoing
  Kibana HTTP request via the existing custom `transport.RoundTrip`, alongside the current `kbn-xsrf`
  and `Authorization` header writes.
- Reject (case-insensitively) a configured header named `kbn-xsrf` or `authorization` with a clear
  provider error at config-build time, because the existing writes for those two headers are
  conditional (`kbn-xsrf` is only added for non-GET/HEAD requests; `Authorization` is only set when a
  provider auth method is configured) and cannot be relied on to always take precedence over a
  user-supplied value of the same name.
- Explicitly exclude `headers` from Fleet's client configuration (`toFleetConfig()`), since Fleet's
  endpoint can diverge from Kibana's and forwarding a Kibana-proxy credential to a different host by
  default would be a credential-leak risk. Fleet does not gain its own `headers` attribute in this
  change.
- Document the existing (unchanged) debug-log behavior: the debug transport already logs the full,
  unredacted `Authorization` header today, and will log custom headers the same way. This change adds
  documentation of that limitation; it does not add new redaction code.
- Ship as a normal, additive provider feature with no Kibana/Stack minimum-version gate (no source
  found that this is a version-dependent capability), validated by running the new acceptance coverage
  across the provider's standard supported-Stack-version test matrix rather than a single pinned
  version.

## Capabilities

### New Capabilities

<!-- None — this change extends an existing capability. -->

### Modified Capabilities

- `provider-kibana-connection`: Adds a `headers` attribute to the shared Kibana connection block
  (`kbConnectionBlockSpec()`), the config-build path that decodes it into `kibanaoapi.Config`, the
  reserved-header validation rule, the Kibana client's header injection in `transport.RoundTrip`, and
  the explicit exclusion of `headers` from the derived Fleet client configuration.

## Impact

- **Changed code**: `internal/schema/kb_connection_shared.go` (new `headers` attribute),
  `internal/clients/config/provider.go` (`KibanaConnection.Headers`),
  `internal/clients/config/kibana_oapi.go` (`buildKibanaOapiConfigFromFramework` decoding + reserved-header
  validation + `toFleetConfig()` exclusion), `internal/clients/kibanaoapi/client.go` (`Config.Headers` +
  `transport.RoundTrip` injection).
- **New tests**: unit tests in `internal/clients/kibanaoapi/transport_test.go` and
  `internal/clients/config/kibana_oapi_test.go`; acceptance tests using a local recording reverse-proxy
  stand-in (new acceptance-test helper, since Cloudflare Access/Google IAP are not available in CI) that
  exercise both the provider-level `kibana { headers = {...} }` block and a resource-level
  `kibana_connection` override, while also covering `KIBANA_ENDPOINT` env-var precedence over configured
  `endpoints` so the proxy test cannot silently pass without receiving traffic.
- **Docs**: Kibana connection block documentation gains a `headers` attribute description and a note
  about debug-log exposure of header values.
- **Out of scope**: Fleet's own `headers` attribute (noted as a non-blocking future follow-up), changes
  to the existing Elasticsearch `headers` attribute, a Beats-style `headers`/`proxy_headers` split, new
  debug-log redaction, any change to generated Kibana client code, and the ephemeral Kibana connection
  snapshot path in `internal/entitycore/ephemeral_connection_snapshot.go` (no Kibana ephemeral resource
  is registered today, so there is no reachable user-facing path to extend).
- **Backward compatibility**: purely additive — `headers` is `Optional` inside the existing
  `SizeAtMost(1)` connection block, so omitting it is a no-op for all existing configurations and state.
