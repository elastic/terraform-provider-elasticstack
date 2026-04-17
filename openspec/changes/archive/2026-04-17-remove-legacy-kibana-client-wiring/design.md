## Context

`internal/clients/kibana_scoped_client.go` and `internal/clients/api_client.go` still depend on `*kibana.Client` from `go-kibana-rest` for `KibanaStatus.Get()` when Elasticsearch is not configured (`APIClient`) and for all `KibanaScopedClient` version and flavor checks. The generated Kibana client already exposes `/api/status` as `GetStatus` / `GetStatusWithResponse` (`generated/kbapi`), and `internal/clients/kibanaoapi` is the established wrapper layer for auth, errors, and typed calls. Many resources still call `GetKibanaClient()` for CRUD; this change is intentionally sequenced **after** those migrations so wiring can drop the legacy pointer without leaving dead imports or broken packages.

## Goals / Non-Goals

**Goals:**

- Single HTTP stack for Kibana status reads used by `ServerVersion` / `ServerFlavor` on `KibanaScopedClient` and the Kibana-only branches of `APIClient`.
- Remove `*kibana.Client` fields and `GetKibanaClient()` from `APIClient` and `KibanaScopedClient` once no in-repo production code requires them.
- Delete `internal/kibana/synthetics/api_client.go` legacy accessors when no callers remain.
- Keep user-visible behavior for version constraints and serverless detection aligned with today (same version string source, same empty flavor when `build_flavor` is absent).

**Non-Goals:**

- Migrating individual resources in this change (covered by separate OpenSpec changes per entity).
- Changing Elasticsearch-backed `APIClient.ServerVersion` / `ServerFlavor` when an Elasticsearch client is configured (existing split behavior remains).
- Removing `libs/go-kibana-rest` from the repository entirely unless the module has zero references after wiring and resource migrations (may remain for examples or transitional tooling per repo policy).

## Decisions

1. **Status implementation path** — Implement a small helper in `internal/clients/kibanaoapi` (for example `GetKibanaStatus` or parse helpers used by `clients` package) that calls `ClientWithResponses.GetStatusWithResponse`, maps HTTP errors to Terraform diagnostics consistent with other `kibanaoapi` helpers, and unmarshals `version.number` and optional `version.build_flavor` from the response body. **Rationale:** Keeps raw OpenAPI usage out of duplicated logic and matches other resources. **Alternative considered:** Call `kbapi` directly from `kibana_scoped_client.go`; rejected to avoid scattering HTTP policy.

2. **Handling weakly typed `GetStatus` JSON** — The generator may surface `JSON200` as an opaque union. **Decision:** Unmarshal from `GetStatusResponse.Body` (or documented JSON200 path) into a minimal local DTO `{ Version struct { Number string; BuildFlavor *string } }` with `encoding/json` rather than relying on legacy `map[string]any`. **Rationale:** Stable parsing and explicit fields; easier tests. **Alternative:** Keep map-based parsing for parity with legacy; acceptable if generator makes typed decode impractical.

3. **Context threading** — Pass `context.Context` into status helpers (update `ServerVersion` / `ServerFlavor` on `KibanaScopedClient` to use ctx on the wire path where today `_ context.Context` is ignored for legacy). **Rationale:** Correct cancellation and tracing for acceptance tests and long applies.

4. **Synthetics helpers** — Remove `GetKibanaClient` / `GetKibanaClientFromScopedClient` only after the last synthetics (and any other) package stops importing them; consolidate callers on `GetKibanaOAPIClientFromScopedClient` or resource-local `kibanaoapi` usage.

## Risks / Trade-offs

- **[Risk] Subtle JSON or status-code differences** between legacy `KibanaStatus.Get()` and generated `/api/status` → **Mitigation:** Golden or fixture tests for status JSON; run targeted acceptance tests for version-gated resources (synthetics, spaces, etc.).
- **[Risk] Build breaks if prerequisite migrations slip** → **Mitigation:** Gate this change behind a checklist in `tasks.md`; CI compile catches stragglers.
- **[Risk] Duplicate status calls** if both `ServerVersion` and `ServerFlavor` fetch full status sequentially → **Mitigation:** Optional small cache on `KibanaScopedClient` for a single apply step (only if profiling shows pain; default is clarity over micro-optimization).

## Migration Plan

1. Complete all resource-level migrations that still require `GetKibanaClient()` for Kibana CRUD (tracked by their respective OpenSpec changes).
2. Land status helper on `kibanaoapi` and switch `KibanaScopedClient` / `APIClient` Kibana-only paths.
3. Remove legacy client construction from `buildKibanaClient` / factory / `APIClient` struct; fix compile errors by updating remaining call sites (should be none if step 1 is done).
4. Delete `internal/kibana/synthetics/api_client.go` if reduced to legacy-only functions; keep `GetKibanaOAPIClient*` in a remaining file if still needed.
5. Run `make build` and targeted acceptance tests per `dev-docs/high-level/testing.md`.
6. If `go list` shows no `go-kibana-rest` imports, trim `go.mod` / `replace` in a final task or follow-up change per scope.

## Open Questions

- Whether `GetStatusParams` (for example `getStatusSummary`) needs to be set for serverless or large payloads; default to generator defaults unless product requires a flag.
- Whether any SDK-only code paths still need a bare `*kibana.Client` for non-status operations after migrations; if yes, those resources block step 3.

## Residual `go-kibana-rest` Owners (post-task-5 state)

After completing tasks 1–5, `go-kibana-rest` cannot be removed from `go.mod` because
six production files still import it:

**`internal/clients/config/` (5 files: `client.go`, `env.go`, `sdk.go`, `framework.go`, `kibana.go`)**
These files carry `kibana.Config` (from `github.com/disaster37/go-kibana-rest/v8`) as:
- The type alias `type kibanaConfig kibana.Config` used throughout the config builders.
- The `Client.Kibana *kibana.Config` field, which is populated by every config builder
  (`NewFromEnv`, `NewFromSDK`, `NewFromSDKKibanaResource`, `NewFromFramework`, `NewFromFrameworkKibanaResource`).
- The field is used in `provider_client_factory.go` as a nil-check sentinel:
  `if cfg.Kibana == nil { return error }` — this is the only remaining runtime use
  of the legacy config type after `apiClient` no longer stores a `*kibana.Client`.

**`internal/kibana/synthetics/parameter/read.go`**
Imports `github.com/disaster37/go-kibana-rest/v8/kbapi` for `*kbapi.APIError` in an
`errors.As` check after `GetParameterWithResponse`. This is the OAPI-generated client,
not the legacy REST client, but `kbapi.APIError` is defined in the legacy library's
`kbapi` sub-package (not in `generated/kbapi`).

**Follow-up needed:**
1. Replace `kibana.Config` in `internal/clients/config/` with either:
   - A local struct containing the same fields (address, username, password, apiKey,
     bearerToken, insecure, caCerts), and update the nil-sentinel in
     `buildKibanaScopedClientFromConfig` to check `cfg.KibanaOapi` instead of `cfg.Kibana`.
2. Replace the `kbapi.APIError` check in `parameter/read.go` with a plain HTTP-status
   check on `getResult.StatusCode()` — the OAPI client does not wrap network errors in
   `kbapi.APIError`, so the `errors.As` is likely a no-op for that error type already.
3. After both changes, run `go mod tidy` — `go-kibana-rest` should drop from `go.mod`
   and the `replace` directive can be removed along with the `libs/go-kibana-rest` vendored tree.
