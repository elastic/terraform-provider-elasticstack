## Context

The archived wiring cleanup removed the last production uses of `*kibana.Client` for Kibana status reads, but the provider still carries `github.com/disaster37/go-kibana-rest/v8` in the module graph for two narrow reasons: the config builders still materialize `kibana.Config` in parallel with `kibanaoapi.Config`, and `internal/kibana/synthetics/parameter/read.go` still imports the legacy `kbapi` subpackage for an error-type assertion. Those residual edges keep the vendored `libs/go-kibana-rest` fork alive, force `go.mod` to retain a local replace, and leave canonical specs describing the removal as unfinished follow-up work.

This change is intentionally narrower than the existing `remove-deprecated-kibana-clients` proposal. It focuses only on retiring `go-kibana-rest` and leaves any standalone `generated/slo` cleanup to its own scope.

## Goals / Non-Goals

**Goals:**

- Remove all first-party imports of `github.com/disaster37/go-kibana-rest/v8` and `github.com/disaster37/go-kibana-rest/v8/kbapi`.
- Delete `config.Client.Kibana` and make `config.Client.KibanaOapi` the only Kibana connection surface used by provider wiring.
- Preserve existing user-visible behavior for provider-level and scoped `kibana_connection` resolution, including acceptance-test inputs and synthetics parameter 404 handling.
- Remove the root `go.mod` dependency, delete the vendored `libs/go-kibana-rest` fork, and align docs/specs with the final state.

**Non-Goals:**

- Removing `generated/slo` or changing SLO-specific provider wiring.
- Reworking the Kibana authentication model beyond translating existing field semantics onto `kibanaoapi.Config`.
- Changing the behavior of Kibana `/api/status` reads, version gating, or `kibana_connection` fallback rules.

## Decisions

1. **Use `kibanaoapi.Config` as the sole Kibana connection contract.**  
   The provider already constructs the generated Kibana client and Fleet client from `kibanaoapi.Config`, so the cleanup should delete `Client.Kibana` rather than introduce a new local compatibility type. This removes the duplicate config surface instead of just renaming it.

   **Alternative considered:** Keep a local struct with legacy field names (`Address`, `ApiKey`, `DisableVerifySSL`, `CAs`) and translate it later. Rejected because it would preserve the same duplication and keep tests/helpers anchored to a deprecated shape.

2. **Translate legacy field expectations at the call sites that still read them.**  
   The main drift is naming, not semantics: `Address -> URL`, `ApiKey -> APIKey`, `DisableVerifySSL -> Insecure`, and `CAs -> CACerts`. The config builders should stop generating legacy values, and the small number of remaining consumers such as `provider/provider_test.go` should switch to the OAPI names directly.

   **Alternative considered:** Add compatibility accessors on `config.Client` so test code can keep using legacy names. Rejected because it would reintroduce a legacy contract after the module is removed.

3. **Treat synthetics parameter 404 handling as a generated-client response concern.**  
   `GetParameterWithResponse` already returns a typed response object with HTTP status access, so the legacy `errors.As(..., *kbapi.APIError)` branch should be replaced with response-status handling. This preserves the existing state-removal semantics for 404 without depending on a legacy error type that does not belong to the generated client stack.

   **Alternative considered:** Create a small wrapper that mimics the legacy `APIError` shape. Rejected because it adds new compatibility code solely to preserve a dependency being removed.

4. **Update canonical requirements to describe the post-cleanup state, not the migration gap.**  
   The canonical `provider-kibana-connection` and `provider-client-factory` specs currently leave room for residual `go-kibana-rest` owners. This change should remove that caveat and add a module-level capability documenting that the dependency and vendored fork are gone.

## Risks / Trade-offs

- **[Risk] Field-name translation misses a test helper or config path** -> **Mitigation:** Search for `cfg.Kibana`, `Address`, `ApiKey`, `DisableVerifySSL`, and `CAs` uses before removing the field, then run targeted config/provider tests plus `go test ./...` if practical.
- **[Risk] Synthetics read behavior changes for transport or unexpected non-200 responses** -> **Mitigation:** Keep transport-error handling separate from HTTP-status handling and preserve the current 404-only state removal behavior.
- **[Risk] Overlap with the broader `remove-deprecated-kibana-clients` change causes duplicate requirement text** -> **Mitigation:** Keep this change explicitly scoped to `go-kibana-rest`, and treat `generated/slo` as out of scope in proposal, design, and spec deltas.
- **[Risk] Repo docs/specs drift from implementation order** -> **Mitigation:** Update OpenSpec deltas and contributor docs in the same change so the final state is internally consistent.

## Migration Plan

1. Remove `Client.Kibana` and legacy `kibanaConfig` construction from `internal/clients/config`, leaving `KibanaOapi` as the only Kibana config surface.
2. Update `internal/clients/provider_client_factory.go` and any remaining tests/helpers to validate and consume `cfg.KibanaOapi` directly.
3. Replace the synthetics parameter read `kbapi.APIError` assertion with response-based 404 handling.
4. Remove the `go-kibana-rest` `require` and `replace` directives from `go.mod`, delete `libs/go-kibana-rest`, and run `go mod tidy`.
5. Update docs and OpenSpec deltas, then verify with build/tests, repository search, and OpenSpec validation.

## Open Questions

- Whether any non-obvious test-only or tool-only code paths still assume `config.Client.Kibana` exists outside the already identified provider acceptance helpers.
- Whether repository policy wants all textual references to `go-kibana-rest` removed from active specs immediately, while archived change text naturally remains historical.
