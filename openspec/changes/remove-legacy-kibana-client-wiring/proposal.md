## Why

After per-resource migrations off `github.com/disaster37/go-kibana-rest` (`libs/go-kibana-rest`), the provider still constructs and threads a legacy `*kibana.Client` through `*clients.APIClient`, `*clients.KibanaScopedClient`, and small helper packages solely to read Kibana status (version and flavor) and to satisfy historical `GetKibanaClient()` call sites. That duplicates HTTP stacks, keeps the forked module in the dependency graph, and blocks fully retiring the legacy client. This change finishes the wiring layer by using `generated/kbapi` status calls for the same facts and removing the legacy surface from shared provider types.

**Prerequisite:** This work MUST land only after the outstanding Kibana resource and data source migrations that still call `GetKibanaClient()` / `synthetics.GetKibanaClientFromScopedClient` (or otherwise depend on the legacy client for CRUD) are complete. Attempting it earlier would break compilation or runtime behavior for unmigrated entities.

## What Changes

- Replace `KibanaScopedClient.ServerVersion` / `ServerFlavor` (and any equivalent paths in `APIClient` that read Kibana-only status via `kibClient.KibanaStatus.Get()`) with calls through the existing OpenAPI stack (`generated/kbapi` `GetStatus` / `GetStatusWithResponse`, typically via `internal/clients/kibanaoapi` helpers) using the same scoped credentials and base URL as today’s `kibanaoapi` client.
- Remove the legacy `*kibana.Client` field and `GetKibanaClient()` from `*clients.KibanaScopedClient` and `*clients.APIClient`, and stop building `kibana.NewClient` in provider wiring once nothing in-tree requires it for status or CRUD.
- Remove helper surfaces that exist only to obtain the legacy client from a scoped client, including `internal/kibana/synthetics/api_client.go` (`GetKibanaClient`, `GetKibanaClientFromScopedClient`) **after** all call sites are migrated to `GetKibanaOapiClient` / `GetKibanaOAPIClientFromScopedClient` (or direct `kibanaoapi` usage).
- Update `ProviderClientFactory` / acceptance test helpers / any tests that assumed a legacy Kibana client was always present on `APIClient` or `KibanaScopedClient`.
- Optional follow-on (same change if trivial, otherwise noted in tasks): trim `go.mod` / `replace` directives for `go-kibana-rest` if no remaining imports exist anywhere in the module (including tests and `libs/go-kibana-rest` vendoring policy per repo conventions).

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `provider-kibana-connection`: Update the Framework/SDK scoped-client requirements so the factory-built `*clients.KibanaScopedClient` no longer includes or rebuilds a legacy Kibana REST client; clarify that version and flavor checks remain scoped to the effective Kibana connection but are satisfied via the OpenAPI client path.
- `provider-client-factory`: Update the typed Kibana-scoped client contract so it no longer lists a legacy Kibana client as a required surface; keep OpenAPI, SLO, Fleet, auth helpers, and version/flavor behavior required for covered entities.

## Impact

- **Code:** `internal/clients/kibana_scoped_client.go`, `internal/clients/api_client.go`, `internal/clients/provider_client_factory.go` (and related constructors), tests under `internal/clients/`, `internal/kibana/synthetics/api_client.go` (delete when unused), any remaining `GetKibanaClient()` usages across `internal/kibana/**`, `internal/fleet/**`, and SDK resources.
- **Dependencies:** Reduced or eliminated direct use of `github.com/disaster37/go-kibana-rest/v8` from provider wiring; `generated/kbapi` becomes the single HTTP path for Kibana `/api/status` in these flows.
- **Risk:** Status JSON parsing may differ from the legacy client’s `map[string]any` shape; implementation must preserve current semantics for `version.number` and `version.build_flavor` (including absent `build_flavor` on older Kibanas) and preserve scoped-connection behavior required by existing specs.
