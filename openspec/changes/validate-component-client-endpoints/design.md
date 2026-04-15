## Context

Provider configuration is intentionally component-optional, so the provider cannot require Elasticsearch, Kibana, and Fleet endpoints up front during provider configuration. The current `*clients.APIClient` accessors only check whether the underlying client pointer is `nil`, which means missing endpoint configuration is discovered later and with low-signal errors.

This is especially visible for Kibana-family access:
- the legacy Kibana client can be constructed with an empty address and will default to `http://localhost:5601`
- the Kibana OpenAPI and Fleet clients can be constructed with empty URLs and fail later when used
- `APIClient` currently retains `kibanaConfig` for SLO auth, but it does not retain resolved Elasticsearch or Fleet endpoint state for validation

The change therefore needs a provider-internal design that preserves optional component configuration while letting entities fail early, at the point where they request a component client, with an actionable configuration error.

## Goals / Non-Goals

**Goals:**
- Make `*clients.APIClient` accessors the single enforcement point for missing component endpoints.
- Validate against the effective resolved endpoint values after provider config and environment overrides have already been applied.
- Preserve existing provider behavior that allows users to configure only the components they use.
- Preserve existing Fleet endpoint inheritance from Kibana configuration.
- Apply the same validation semantics to the normal provider constructors and the acceptance-test client constructor.

**Non-Goals:**
- Adding authentication validation for username/password, API key, or bearer token fields.
- Changing provider schema requirements or making component endpoints mandatory at provider configure time.
- Redesigning config resolution order or the existing Kibana-to-Fleet endpoint inheritance behavior.
- Changing entity-level code paths beyond their existing accessor usage.

## Decisions

Validate at accessor entry, not during provider configuration or client construction.
The accessors are the first point where the provider knows which component a resource or data source is trying to use. Validating there preserves the current component-optional provider contract while replacing downstream transport failures with targeted configuration errors.

Alternative considered: validate in provider schema or provider `Configure`.
Rejected because it would force all components to be configured even for users who only manage Elasticsearch or only manage Kibana-backed resources.

Store resolved endpoint snapshots on `APIClient`.
`APIClient` should retain the effective endpoint state needed by the accessors, separate from the constructed client objects. The minimal design is to store resolved Elasticsearch endpoint presence, resolved Kibana endpoint, and resolved Fleet endpoint alongside the existing clients and `kibanaConfig`.

Alternative considered: infer configuration state from the constructed clients.
Rejected because the client implementations are inconsistent. The legacy Kibana client masks missing configuration by defaulting to localhost, while the OpenAPI and Fleet clients do not provide a reliable, uniform signal for "configured but empty".

Populate endpoint snapshots from resolved `config.Client` values.
The snapshot should be taken after config resolution has already folded together provider input and environment overrides. For Fleet, the stored value should reflect the already-resolved `cfg.Fleet` endpoint, which may have been inherited from the Kibana-derived config path. This keeps the accessor checks aligned with the provider's existing resolution semantics instead of reimplementing them.

Alternative considered: recompute Fleet fallback inside `GetFleetClient()`.
Rejected because it would duplicate config resolution rules in the accessor layer and create drift risk between config building and validation.

Keep endpoint validation separate from the `nil` client safety guard.
The accessors should use endpoint validation for user-facing configuration problems and retain the current `nil` checks as an internal safety net. In practice, missing endpoints should produce the new actionable errors, while unexpected construction gaps can still surface as internal "client not found" failures.

Alternative considered: replace the `nil` guards entirely.
Rejected because the `nil` checks still provide value for unexpected internal states and make the change less risky.

Unify constructor behavior enough that all `APIClient` creation paths carry validation metadata.
The normal provider constructors already flow through `newAPIClientFromConfig(...)`; the acceptance-test constructor currently builds the clients inline. The implementation should ensure both paths populate the same endpoint snapshot state so tests and production code see the same accessor behavior.

Alternative considered: leave `NewAcceptanceTestingClient()` as a special case.
Rejected because it would make accessor behavior differ between production and acceptance-test client setup, which weakens regression coverage for this change.

Limit the new validation to endpoint presence only.
The change should check only whether a component has an effective endpoint. If an endpoint is present, the accessor should not fail solely because auth fields are empty.

Alternative considered: validate endpoint and auth together.
Rejected because that broadens the behavior change beyond issue #355 and would interfere with deployments that rely on proxy-managed or anonymous auth behavior.

## Risks / Trade-offs

- [Risk] Eager client construction still happens before accessor validation, so some misconfigured clients may still be instantiated internally -> Mitigation: keep this change scoped to access-time validation and rely on the accessor boundary to prevent misleading runtime request errors.
- [Risk] Future config-builder changes could update endpoint resolution without updating the stored snapshot fields -> Mitigation: snapshot directly from the resolved `config.Client` values in the shared construction path and add focused tests for each accessor.
- [Risk] Wrapper diagnostics such as `unable to get kibana client` will still add their own summary around the new errors -> Mitigation: make the accessor error strings self-contained and actionable so the composed diagnostic remains clear.

## Migration Plan

1. Extend `APIClient` to retain the resolved endpoint state needed for Elasticsearch, Kibana, and Fleet validation.
2. Populate that state from resolved `config.Client` values in the shared constructor path and in the acceptance-test client path.
3. Update `GetESClient()`, `GetKibanaClient()`, `GetKibanaOapiClient()`, `GetSloClient()`, and `GetFleetClient()` to perform endpoint validation before returning the client.
4. Add focused unit coverage for missing-endpoint failures, Fleet endpoint inheritance, and the legacy Kibana localhost fallback case.

## Open Questions

- None. The remaining work is implementation and focused test coverage.

