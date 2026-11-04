## 1. Schema

- [ ] 1.1 Add a Kibana-specific `descKbHeaders` constant to `internal/schema/connection_constants.go`
      (reuse the existing `attrHeaders` key constant; do not reuse `descHeaders`, which is
      Elasticsearch-specific wording)
- [ ] 1.2 Add a `headers` entry (`kind: connAttrMap, sensitive: true, writeOnly: true`) to the `attrs`
      slice in `kbConnectionBlockSpec()` (`internal/schema/kb_connection_shared.go`)
- [ ] 1.3 Verify that generated provider, managed-resource, data-source, and ephemeral variants mark
      `headers` sensitive, that only the action variant marks it write-only, and that entity-local
      managed-resource/data-source values remain in state while ephemeral values follow Terraform's
      ephemeral-state semantics

## 2. Config plumbing

- [ ] 2.1 Add `Headers types.Map` to `KibanaConnection` in `internal/clients/config/provider.go`
- [ ] 2.2 Add `Headers map[string]string` to `kibanaoapi.Config` in `internal/clients/kibanaoapi/client.go`
- [ ] 2.3 In `buildKibanaOapiConfigFromFramework` (`internal/clients/config/kibana_oapi.go`), decode
      `kibConfig.Headers` into `config.Headers` (handling null/unknown as empty, mirroring the existing
      `endpoints`/`ca_certs` decoding pattern)
- [ ] 2.4 In the same function, after decoding, case-insensitively reject any configured header key
      equal to `kbn-xsrf` or `authorization`, returning a `fwdiags.Diagnostics` error that names the
      offending key before the scoped connection's first request; do not require the diagnostic to
      occur during planning for a new entity-local resource
- [ ] 2.5 Update `toFleetConfig()` (`internal/clients/config/kibana_oapi.go`) to explicitly clear
      `Headers` on the returned `fleetConfig` after the struct conversion, so Kibana-only headers are
      never forwarded to the Fleet client by default

## 3. Transport injection

- [ ] 3.1 In `transport.RoundTrip` (`internal/clients/kibanaoapi/client.go`), iterate `t.Headers` and
      `req.Header.Set(name, value)` for each entry, running before the existing `kbn-xsrf` and
      `Authorization` writes
- [ ] 3.2 Add a short code comment at the reserved-name check (task 2.4) and at the injection loop
      cross-referencing each other, so a future reader understands why ordering alone is not treated as
      the safety mechanism

## 4. Unit tests

- [ ] 4.1 Extend `internal/clients/kibanaoapi/transport_test.go` to assert a configured custom header
      reaches the outgoing request on both a GET/HEAD request and a mutation request, with and without
      provider-configured auth
- [ ] 4.2 Add a unit test asserting an empty/omitted `Headers` map sends no extra headers (no regression
      for existing callers)
- [ ] 4.3 Add config-level unit tests in `internal/clients/config/kibana_oapi_test.go` (new file, or
      extend if one already exists) asserting `headers` decodes into `kibanaoapi.Config.Headers`
      correctly, including the null/omitted case
- [ ] 4.4 Add a config-level unit test asserting a configured header named `kbn-xsrf` or `Authorization`
      (any case) is rejected with a diagnostics error identifying the key, and that no further
      processing silently overwrites it
- [ ] 4.5 Add a config-level unit test asserting `toFleetConfig()` does not carry `Headers` through when
      the source Kibana config has headers set

## 5. Acceptance test infrastructure

- [ ] 5.1 Check whether `internal/acctest` already has a reusable header-recording reverse-proxy test
      helper; if not, add one backed by `httptest.Server` that forwards requests to a target endpoint
      while recording the headers of each received request
- [ ] 5.2 Add an acceptance test that configures the provider-level `kibana { headers = {...} }` block
      pointed at the local proxy, sets `TF_ELASTICSTACK_PREFER_CONFIGURED_KIBANA_ENDPOINT=true` (or
      points `KIBANA_ENDPOINT` itself at the proxy) so a CI-configured real-cluster `KIBANA_ENDPOINT`
      cannot silently bypass the proxy, and asserts the proxy recorded at least one request carrying the
      configured header with the expected value
- [ ] 5.3 Add an acceptance test exercising the same proxy-recording assertion for a resource-level
      `kibana_connection { headers = {...} }` override on a resource/data source that already supports
      `kibana_connection`
- [ ] 5.4 Run both new acceptance tests against each Stack version in the provider's standard supported
      acceptance-test version matrix (no single pinned version), per the "no version gate" decision

## 6. Resource/data-source state-shape coverage

- [ ] 6.1 Following `TestAccDataSourceEnrichPolicyConnectionAPIKey` /
      `TestAccDataSourceEnrichPolicyConnectionAPIKeyEmptyHeaders`
      (`internal/elasticsearch/enrich/acc_test.go`), add acceptance coverage on a resource/data source
      that exposes a `kibana_connection` block in its own state, asserting `kibana_connection.0.headers.%`
      and specific key/value pairs across add → update → remove → empty-map steps

## 7. Documentation

- [ ] 7.1 Regenerate provider docs for the Kibana connection block (`make` doc-generation target per
      `dev-docs/high-level/documentation.md`) so the new `headers` attribute description appears
- [ ] 7.2 Add a note to the Kibana connection documentation that debug-level logging
      (`TF_LOG`/provider debug transport) already logs the full `Authorization` header unredacted today,
      and will log custom `headers` values the same way
- [ ] 7.3 Add a CHANGELOG entry describing the new `headers` attribute on the Kibana connection block

## 8. Validation

- [ ] 8.1 Run `make build`
- [ ] 8.2 Run `make check-lint`
- [ ] 8.3 Run the new and existing unit tests for `internal/clients/kibanaoapi` and
      `internal/clients/config`
- [ ] 8.4 Run the new acceptance tests against an available Elastic Stack per
      `dev-docs/high-level/testing.md`, across the supported version matrix called for in task 5.4
- [ ] 8.5 Confirm `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-connection-custom-headers --type change` passes
