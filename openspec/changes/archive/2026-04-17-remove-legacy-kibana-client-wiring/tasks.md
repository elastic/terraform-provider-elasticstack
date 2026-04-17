## 1. Preconditions (blocking)

- [x] 1.1 Confirm every Kibana/Fleet resource and data source that still imports `github.com/disaster37/go-kibana-rest` or calls `(*clients.KibanaScopedClient).GetKibanaClient()` / `synthetics.GetKibanaClientFromScopedClient` for CRUD has merged its kbapi migration (tracked OpenSpec changes such as `migrate-kibana-*-to-kbapi`, `finish-kibana-synthetics-parameter-kbapi`, and related items on the migration plan). **Note:** Resources that call `ProviderClientFactory.GetKibanaClient(ctx, ...)` (the factory method that returns a `*KibanaScopedClient`) are using the correct non-legacy API — this is a different method from the now-removed `(*KibanaScopedClient).GetKibanaClient()` which returned a `*kibana.Client`. Investigation confirmed no production resource calls `(*KibanaScopedClient).GetKibanaClient()` for CRUD.
- [x] 1.2 Run `go list` / `rg 'disaster37/go-kibana-rest'` on `internal/` and fix any remaining stragglers before deleting wiring; this change MUST NOT merge while production packages still require the legacy client.

## 2. Status and version wiring

- [x] 2.1 Add `internal/clients/kibanaoapi` helper(s) that call `generated/kbapi` `GetStatusWithResponse` (or equivalent) with the same HTTP client and request editors as other helpers, returning parsed `version.number` and optional `version.build_flavor` plus consistent error diagnostics on non-2xx or decode failures.
- [x] 2.2 Refactor `(*KibanaScopedClient).ServerVersion` and `ServerFlavor` in `internal/clients/kibana_scoped_client.go` to use the helper from 2.1 with `GetKibanaOapiClient()`, remove `GetKibanaClient()` and the legacy `kibana` field from the struct, and adjust `kibanaScopedClientFromAPIClient` accordingly.
- [x] 2.3 Refactor `(*APIClient).versionFromKibana` and `flavorFromKibana` in `internal/clients/api_client.go` to use the same helper path (via `kibanaoapi` on the API client’s OpenAPI client) instead of `GetKibanaClient().KibanaStatus.Get()`. Note: `versionFromKibana`/`flavorFromKibana` did not exist as standalone functions; the equivalent Kibana-only status path was in `kibana_scoped_client.go` which was fully migrated in 2.2. The `apiClient` legacy kibana fields and `buildKibanaClient` were removed in task 3.1.

## 3. Remove legacy client from shared types

- [x] 3.1 Remove `kibana` / legacy-related fields, `GetKibanaClient()`, and `buildKibanaClient` usage from `APIClient` construction paths (`NewAPIClientFromFramework`, SDK constructor, `NewAcceptanceTestingClient`, etc.) once no callers need them; retain debug transport behavior on the HTTP client shared with `kibanaoapi` where applicable.
- [x] 3.2 Update `internal/clients/provider_client_factory.go` (and any scoped rebuild helpers) so scoped clients are built without instantiating `kibana.NewClient`.
- [x] 3.3 Fix all compile breaks: replace remaining `GetKibanaClient()` usages in tests and resources with `GetKibanaOapiClient()` + `kibanaoapi` helpers or entity-specific kbapi paths.

## 4. Delete legacy-only helper surfaces

- [x] 4.1 Remove `GetKibanaClient` and `GetKibanaClientFromScopedClient` from `internal/kibana/synthetics/api_client.go` when zero references remain; if the file only contains OpenAPI helpers afterward, consider inlining `GetKibanaOAPIClient*` next to call sites or keeping a slim `api_client.go` without `go-kibana-rest` imports.
- [x] 4.2 Search for other packages whose sole purpose is re-exporting the legacy client from scoped wiring; delete or fold into `kibanaoapi` helpers.

## 5. Verification and cleanup

- [x] 5.1 Update or add unit tests for status parsing and for `ServerVersion` / `ServerFlavor` on `KibanaScopedClient` (including missing `build_flavor` and serverless flavor behavior).
- [x] 5.2 Run `make build` and targeted acceptance tests for version-gated Kibana resources per `dev-docs/high-level/testing.md`.
- [x] 5.3 Run `make check-openspec` (or `openspec validate` for this change) and address any spec or schema issues.
- [x] 5.4 If the module has no remaining `go-kibana-rest` imports, remove the dependency from `go.mod` / tidy; otherwise document the residual owners in `design.md` open questions and leave a follow-up task.
