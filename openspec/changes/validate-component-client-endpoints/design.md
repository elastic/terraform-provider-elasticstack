## Context

Provider configuration is intentionally component-optional, so the provider cannot require Elasticsearch, Kibana, and Fleet endpoints up front during provider configuration. The current provider wiring injects a `*clients.ProviderClientFactory`, which resolves typed `*clients.ElasticsearchScopedClient` and `*clients.KibanaScopedClient` values. Those scoped clients are adapted either from provider-level defaults stored on the internal `apiClient` or from entity-local connection blocks rebuilt through `config.NewFromFramework...` / `config.NewFromSDK...` helpers. The current scoped accessors mostly check only whether the underlying client pointer is `nil`, which means missing endpoint configuration is discovered later and with low-signal errors.

This is especially visible for Kibana-family access:
- the legacy Kibana client can be constructed with an empty address and will default to `http://localhost:5601`
- the Kibana OpenAPI and Fleet clients can be constructed with empty URLs and fail later when used
- `ElasticsearchScopedClient` and `KibanaScopedClient` do not currently retain resolved endpoint state for endpoint-specific accessor validation
- provider-default and entity-local connection resolution both allow endpoint-less configs because auth may still legitimately inherit from Elasticsearch or environment variables until a specific component is actually used

The change therefore needs a provider-internal design that preserves optional component configuration while letting entities fail early, at the point where they request a concrete component client from a typed scoped client, with an actionable configuration error.

## Goals / Non-Goals

**Goals:**
- Make typed scoped-client accessors the single user-facing enforcement point for missing component endpoints.
- Validate against the effective resolved endpoint values after provider config, entity-local connection blocks, and environment overrides have already been applied.
- Preserve existing provider behavior that allows users to configure only the components they use.
- Preserve existing Fleet endpoint inheritance from Kibana configuration, including the `kibana_connection`-derived Fleet path.
- Apply the same validation semantics to provider-default clients, entity-local scoped clients, and the acceptance-test client constructor.

**Non-Goals:**
- Adding authentication validation for username/password, API key, or bearer token fields.
- Changing provider schema requirements or making component endpoints mandatory at provider configure time.
- Redesigning config resolution order or the existing Kibana-to-Fleet endpoint inheritance behavior.
- Changing entity-level code paths beyond their existing factory and accessor usage.

## Decisions

Validate at typed scoped-client accessor entry, not during provider configuration.
`(*clients.ElasticsearchScopedClient).GetESClient()` and `(*clients.KibanaScopedClient).GetKibanaClient()`, `GetKibanaOapiClient()`, and `GetFleetClient()` are the first points where an entity asks for a concrete component client. Validating there preserves the current component-optional provider contract while replacing downstream transport failures with targeted configuration errors.

Alternative considered: validate in provider schema or provider `Configure`.
Rejected because it would force all components to be configured even for users who only manage Elasticsearch or only manage Kibana-backed resources.

Store resolved endpoint snapshots on the typed scoped clients, with provider-default metadata carried through the internal `apiClient`.
The scoped clients should retain the effective endpoint state needed by their accessors, separate from the constructed client objects. Because provider-default scoped clients are adapted from `apiClient`, the internal broad client should also retain the same resolved endpoint metadata so the adapters can populate the scoped accessors consistently.

Alternative considered: infer configuration state from the constructed clients.
Rejected because the client implementations are inconsistent. The legacy Kibana client masks missing configuration by defaulting to localhost, while the OpenAPI and Fleet clients do not provide a reliable, uniform signal for "configured but empty".

Populate endpoint snapshots from resolved `config.Client` values.
The snapshot should be taken after config resolution has already folded together provider input, entity-local overrides, and environment overrides. For Fleet, the stored value should reflect the already-resolved `cfg.Fleet` endpoint, which may have been inherited from the Kibana-derived config path. This keeps the accessor checks aligned with the provider's existing resolution semantics instead of reimplementing them.

Alternative considered: recompute Fleet fallback inside `GetFleetClient()`.
Rejected because it would duplicate config resolution rules in the accessor layer and create drift risk between config building and validation.

Keep endpoint validation separate from the `nil` client safety guard.
The accessors should use endpoint validation for user-facing configuration problems and retain the current `nil` checks as an internal safety net. In practice, missing endpoints should produce the new actionable errors, while unexpected construction gaps can still surface as internal "client not found" failures.

Alternative considered: replace the `nil` guards entirely.
Rejected because the `nil` checks still provide value for unexpected internal states and make the change less risky.

Unify constructor behavior enough that all `APIClient` creation paths carry validation metadata.
The normal provider constructors already flow through `newAPIClientFromConfig(...)`; the acceptance-test constructor currently builds the clients inline. The implementation should ensure both paths populate the same endpoint snapshot state so tests and production code see the same accessor behavior. Entity-local scoped-client builders in `ProviderClientFactory` should likewise populate endpoint metadata from their resolved `config.Client` inputs rather than relying on provider-default adapters only.

Alternative considered: leave `NewAcceptanceTestingClient()` as a special case.
Rejected because it would make accessor behavior differ between production and acceptance-test client setup, which weakens regression coverage for this change.

Limit the new validation to endpoint presence only.
The change should check only whether a component has an effective endpoint. If an endpoint is present, the accessor should not fail solely because auth fields are empty.

Alternative considered: validate endpoint and auth together.
Rejected because that broadens the behavior change beyond issue #355 and would interfere with deployments that rely on proxy-managed or anonymous auth behavior.

## Risks / Trade-offs

- [Risk] Eager client construction still happens before accessor validation, so some misconfigured clients may still be instantiated internally -> Mitigation: keep this change scoped to access-time validation and rely on the accessor boundary to prevent misleading runtime request errors.
- [Risk] Future config-builder changes could update endpoint resolution without updating the stored snapshot fields -> Mitigation: snapshot directly from the resolved `config.Client` values in the shared construction paths and add focused tests for each accessor.
- [Risk] Some entity paths may still wrap accessor failures (for example, version checks or higher-level helper diagnostics) -> Mitigation: make the accessor error strings self-contained and actionable so the composed diagnostic remains clear.

## Migration Plan

1. Extend the internal client construction path and typed scoped clients to retain the resolved endpoint state needed for Elasticsearch, Kibana, and Fleet validation.
2. Populate that state from resolved `config.Client` values in the shared provider constructor path, in entity-local scoped-client builder paths, and in the acceptance-test client path.
3. Update `GetESClient()`, `GetKibanaClient()`, `GetKibanaOapiClient()`, and `GetFleetClient()` to perform endpoint validation before returning the client.
4. Add focused unit coverage for missing-endpoint failures, entity-local override behavior, Fleet endpoint inheritance, and the legacy Kibana localhost fallback case.

## Open Questions

- None. The remaining work is implementation and focused test coverage.

