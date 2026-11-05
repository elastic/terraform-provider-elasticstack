## Why

Kibana is frequently placed behind an identity-aware proxy (Cloudflare Access, Google IAP) that
requires a custom HTTP header on every request (`cf-access-token: ...`, `Proxy-Authorization: ...`)
before Kibana itself is ever reached. The provider's `kibana` connection block has no way to attach
custom headers to outgoing Kibana API requests, so these setups cannot authenticate the Terraform
provider at all ([#440](https://github.com/elastic/terraform-provider-elasticstack/issues/440), with
corroborating reports from `@hayk96` and `@hossamdash`).

## What Changes

- Add an optional map-of-string `headers` attribute everywhere the Kibana connection block is already
  available. Provider, managed-resource, data-source, and ephemeral values are sensitive;
  managed-resource and data-source values remain in state; action values are write-only; and ephemeral
  values follow Terraform's ephemeral-state semantics.
- Send each accepted custom header on every outgoing Kibana HTTP request alongside the
  provider-managed `kbn-xsrf` and `Authorization` headers.
- When custom headers are configured, allow redirects only when the destination has the same origin
  (scheme, host, and effective port) as the initial Kibana request. Reject cross-origin redirects,
  including HTTPS-to-HTTP downgrades, before custom credentials can be forwarded.
- Reject provider-managed or Go transport-controlled header names (`kbn-xsrf`, `authorization`,
  `host`, `content-length`, `transfer-encoding`, and `trailer`) case-insensitively with a clear
  provider error before the affected connection sends its first request. Also reject multiple
  configured keys that differ only by case, because HTTP header names are case-insensitive and
  accepting both would make the transmitted value nondeterministic.
- Exclude Kibana custom headers from Fleet requests, since Fleet's endpoint can diverge from Kibana's
  and forwarding a Kibana-proxy credential to a different host by default would be a credential-leak
  risk. Fleet does not gain its own `headers` attribute in this change.
- Document the existing (unchanged) debug-log behavior accurately: the debug formatter masks
  `Authorization`, but arbitrary custom proxy-token headers such as `cf-access-token` are not generally
  redacted. This change documents that limitation; it does not add new redaction code.
- Ship as a normal, additive provider feature with no Kibana/Stack minimum-version gate (no source
  found that this is a version-dependent capability), validated by running the new acceptance coverage
  across the provider's standard supported-Stack-version test matrix rather than a single pinned
  version.

## Capabilities

### New Capabilities

- None. This change extends an existing capability.

### Modified Capabilities

- `provider-kibana-connection`: Adds a `headers` attribute to the shared Kibana connection block
  (`kbConnectionBlockSpec()`), the config-build path that decodes it into `kibanaoapi.Config`, the
  reserved-header and case-insensitive-collision validation rules, same-origin redirect protection,
  the Kibana client's header injection in `transport.RoundTrip`, and the explicit exclusion of
  `headers` from the derived Fleet client configuration.

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
  `endpoints` so the proxy test cannot silently pass without receiving traffic; and redirect tests that
  prove same-origin redirects remain usable while cross-origin and HTTPS-to-HTTP redirects are rejected
  before the destination receives custom headers.
- **Docs**: Kibana connection block documentation gains a `headers` attribute description and a note
  that `Authorization` is masked by existing debug formatting while arbitrary custom header values are
  not generally redacted.
- **Out of scope**: Fleet's own `headers` attribute (noted as a non-blocking future follow-up), changes
  to the existing Elasticsearch `headers` attribute, a Beats-style `headers`/`proxy_headers` split, new
  debug-log redaction, any change to generated Kibana client code, and the ephemeral Kibana connection
  snapshot path in `internal/entitycore/ephemeral_connection_snapshot.go` (no Kibana ephemeral resource
  is registered today, so there is no reachable user-facing path to extend).
- **Backward compatibility**: purely additive — `headers` is `Optional` inside the existing
  `SizeAtMost(1)` connection block, so omitting it is a no-op for all existing configurations and state.
