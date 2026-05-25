## 1. Config layer: bidirectional Kibana ↔ Fleet inheritance

- [ ] 1.1 In `internal/clients/config/kibana_oapi.go` (or a new helper invoked from `framework.go`), add a Fleet-block field-level fallback step that runs after the kibana block overlay and before the `KIBANA_*` environment overrides. Inherit `URL`, `Username`, `Password`, `APIKey`, `BearerToken`, `CACerts`, and `Insecure` field-by-field from `cfg.Fleet[0]` only when the corresponding kibana-derived field is still unset.
- [ ] 1.2 Ensure the new fallback applies only on the provider-level path (`newKibanaOapiConfigFromFramework` in `framework.go` `NewFromFramework`) and **not** on the resource-level path (`NewFromFrameworkKibanaResource`). Add a regression assertion that `NewFromFrameworkKibanaResource` does not consult `cfg.Fleet`.
- [ ] 1.3 Add unit tests covering: (a) `fleet { endpoint }` only → kibana_oapi.URL is the fleet endpoint; (b) `kibana { endpoints }` with `fleet { api_key }` → kibana_oapi uses the kibana URL and the fleet API key; (c) `KIBANA_ENDPOINT` env override interacts correctly with the fleet-block fallback per existing `withEnvironmentOverrideUnlessConfigured` semantics; (d) resource-level `kibana_connection` ignores the fleet provider block entirely.

## 2. Factory: move endpoint validation into resolution methods

- [ ] 2.1 In `internal/clients/provider_client_factory.go`, change `GetElasticsearchClient` to perform the endpoint-presence check currently in `(*ElasticsearchScopedClient).GetESClient`. Return error diagnostics with the existing message `elasticsearch client is not configured: set elasticsearch.endpoints, elasticsearch_connection.endpoints, or ELASTICSEARCH_ENDPOINTS` when no non-empty endpoint is configured. Do not return a scoped client on this failure path.
- [ ] 2.2 In `GetKibanaClient`, perform the new "at least one of kibana or fleet endpoint" check after the scoped client is built. Return error diagnostics with a combined message naming all five user-facing configuration paths (`kibana.endpoints`, `kibana_connection.endpoints`, `KIBANA_ENDPOINT`, `fleet.endpoint`, `FLEET_ENDPOINT`) when neither endpoint is configured.
- [ ] 2.3 Verify the resource-level `kibana_connection` path in `GetKibanaClient` continues to surface its existing diagnostics from `buildKibanaScopedClientFromConfig` and `config.NewFromFrameworkKibanaResource` without regression.
- [ ] 2.4 Verify the resource-level `elasticsearch_connection` path in `GetElasticsearchClient` continues to surface its existing diagnostics (multiple blocks, build errors) and that the new factory-level endpoint check runs on the resulting scoped client.

## 3. Scoped clients: convert accessors to single-return getters

- [ ] 3.1 Change `(*ElasticsearchScopedClient).GetESClient()` to return `*elasticsearch.TypedClient` (no diagnostics). Drop the endpoint-presence and nil-client checks from the accessor body — the factory now guarantees those conditions.
- [ ] 3.2 Update `(*ElasticsearchScopedClient).serverInfo` and any other internal callers of `GetESClient` in `elasticsearch_scoped_client.go` to use the new signature.
- [ ] 3.3 Change `(*KibanaScopedClient).GetKibanaOapiClient()` to return `*kibanaoapi.Client` (no diagnostics). Drop the endpoint-presence and nil-client checks.
- [ ] 3.4 Change `(*KibanaScopedClient).GetFleetClient()` to return `*fleet.Client` (no diagnostics). Drop the endpoint-presence and nil-client checks.
- [ ] 3.5 Update `(*KibanaScopedClient).getServerStatusRaw` and any other internal callers in `kibana_scoped_client.go` to use the new `GetKibanaOapiClient` signature.

## 4. Update scoped-client and factory unit tests

- [ ] 4.1 Rewrite `internal/clients/elasticsearch_scoped_client_test.go` tests that constructed an `ElasticsearchScopedClient` directly with no endpoint and expected error diagnostics from `GetESClient`. Move those expectations to factory-level tests against `ProviderClientFactory.GetElasticsearchClient`. Retain test coverage that `GetESClient` returns the typed client when the scoped client was built via the factory.
- [ ] 4.2 Rewrite `internal/clients/kibana_scoped_client_test.go` tests that constructed a `KibanaScopedClient` directly with no endpoint and expected error diagnostics from `GetKibanaOapiClient` or `GetFleetClient`. Move those expectations to `ProviderClientFactory.GetKibanaClient`. Retain endpoint-inheritance coverage for the existing Fleet-from-Kibana path.
- [ ] 4.3 Extend `internal/clients/provider_client_factory_test.go` with: (a) `GetElasticsearchClient` returns a missing-endpoint error when no ES endpoint is configured; (b) `GetKibanaClient` returns a missing-endpoint error when neither Kibana nor Fleet endpoint is configured; (c) `GetKibanaClient` succeeds when only the Fleet endpoint is configured at provider level and `GetKibanaOapiClient` on the resulting scoped client targets that endpoint; (d) successful factory call → accessor returns a non-nil typed client.

## 5. Sweep consumer call sites — Elasticsearch sinks

- [ ] 5.1 Update all `(...).GetESClient()` call sites in `internal/clients/elasticsearch/**` to use the single-return signature. Delete the immediate `if diags.HasError()` block that follows the call and any local `d` / `diags` variable used only for that check. Verify each file compiles.
- [ ] 5.2 Sweep `internal/elasticsearch/**` resources and data sources for any remaining `GetESClient()` callers and apply the same simplification.
- [ ] 5.3 Sweep `internal/entitycore/**`, `internal/clients/**` (outside the scoped client files), and `internal/acctest/**` for `GetESClient()` callers.

## 6. Sweep consumer call sites — Kibana and Fleet

- [ ] 6.1 Update all `(...).GetKibanaOapiClient()` call sites in `internal/kibana/**`, `internal/apm/**`, and `internal/entitycore/**` to use the single-return signature; drop the trailing diagnostic check.
- [ ] 6.2 Update all `(...).GetFleetClient()` call sites in `internal/fleet/**` and `internal/acctest/**` to use the single-return signature; drop the trailing diagnostic check.
- [ ] 6.3 Sweep `internal/clients/fleet/**`, `internal/clients/kibanaoapi/**`, and any other package that holds `KibanaScopedClient` callers for the same simplification.

## 7. Verify external surface

- [ ] 7.1 Confirm `xpprovider/xpprovider.go` and any other consumer of `ProviderClientFactory` public methods still compile — the factory method signatures are unchanged.
- [ ] 7.2 Confirm `NewAcceptanceTestingElasticsearchScopedClient` and `NewAcceptanceTestingKibanaScopedClient` still work; if any acceptance-test helper unwraps an inner client via the changed accessors, update it.

## 8. Build, test, and acceptance verification

- [ ] 8.1 Run `make build` and resolve any compile errors from missed call sites.
- [ ] 8.2 Run the full `go test ./...` unit suite. Resolve any remaining test failures stemming from the moved validation.
- [ ] 8.3 Run a targeted acceptance test for one Elasticsearch resource, one Kibana resource, and one Fleet resource against a local stack. Confirm Create/Read/Update/Delete still succeed end-to-end.
- [ ] 8.4 Run `make check-openspec` to validate the change artifacts.

## 9. Documentation and CHANGELOG

- [ ] 9.1 Add a CHANGELOG entry under `[Unreleased]` summarising the new "provider configured with only `fleet { ... }` block can now serve Kibana resources" behavior and any earlier-failure user-visible behavior for unconfigured providers.
- [ ] 9.2 If `dev-docs/high-level/*.md` references the scoped accessor signatures, update those references.
