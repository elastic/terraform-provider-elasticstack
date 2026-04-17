## Why

The `elasticstack_kibana_synthetics_monitor` resource still depends on the legacy `go-kibana-rest` client and hand-maintained `kbapi` structs for monitor payloads, while the rest of the provider is standardizing on the OpenAPI-generated `generated/kbapi` client with thin `kibanaoapi` wrappers. Migrating removes duplicate type definitions, aligns error handling with other Kibana resources, and reduces drift between provider models and Kibana’s published OpenAPI schema.

## What Changes

- Add `internal/clients/kibanaoapi` helpers for synthetics monitor **create**, **read**, **update**, and **delete** using `generated/kbapi` (`PostSyntheticMonitors`, `GetSyntheticMonitor`, `PutSyntheticMonitor`, `DeleteSyntheticMonitor`) with space-aware request editors consistent with other `kibanaoapi` resources.
- Replace all uses of `github.com/disaster37/go-kibana-rest/v8/kbapi` **monitor request/response types** inside `internal/kibana/synthetics/monitor` (including `schema.go`, CRUD, and tests) with generated `kbapi` types and/or small internal DTOs where unions require explicit JSON handling.
- Keep **practitioner-visible behavior** unchanged: Plugin Framework schema, validators (exactly-one type, at-least-one location, enums), composite `id`, import format, `labels` version gating (≥ 8.16.0), read-side preservation rules (`locations`, `http.check`/`http.response`, null `params`), private-location mapping, and diagnostics text where the canonical spec requires it.
- Retain use of `*clients.KibanaScopedClient` where needed for **non-monitor** concerns already wired through it (for example `enforceVersionConstraints` / stack version checks), unless a follow-up change consolidates version discovery on the OpenAPI path.
- This change is **not** a Terraform schema or import-format **BREAKING** change; it is an internal client migration with a large code touch surface.

## Capabilities

### New Capabilities

- (none — behavior stays under the existing monitor capability)

### Modified Capabilities

- `kibana-synthetics-monitor`: Document that synthetics monitor HTTP traffic is implemented via `generated/kbapi` and `kibanaoapi` helpers; clarify write-path wording that today references legacy Go type names; add an implementation placement requirement for the new client layer.

## Impact

- **Primary**: `internal/kibana/synthetics/monitor` (`schema.go`, `create.go`, `read.go`, `update.go`, `delete.go`, `resource.go`, `schema_test.go`, `acc_test.go`, extensive `testdata/`).
- **New**: `internal/clients/kibanaoapi` (new file(s) for monitor operations, following patterns in `streams.go`, `dashboards.go`, etc.).
- **Shared**: `internal/kibana/synthetics/api_client.go` (may gain or reuse `GetKibanaOAPIClient` for monitor CRUD only; avoid widening scope to unrelated synthetics resources).
- **Tests**: Acceptance tests in `acc_test.go` must be re-run against a live stack; `schema_test.go` must be rewritten around generated types or public mapper tests.
- **External**: Removes monitor-specific dependency on legacy `kbapi` structs from `go-kibana-rest` for this resource (module cleanup may be deferred until no other package imports those types).
- **Risk**: Generated request bodies use `json.RawMessage` unions (`PostSyntheticMonitorsJSONBody`, `PutSyntheticMonitorJSONBody`); helpers must marshal discriminator + payload correctly or Kibana will reject requests.
