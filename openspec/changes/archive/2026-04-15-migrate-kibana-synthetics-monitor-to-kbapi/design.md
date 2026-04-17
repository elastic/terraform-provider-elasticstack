## Context

Today `internal/kibana/synthetics/monitor` calls `kibanaClient.KibanaSynthetics.Monitor.{Add,Get,Update,Delete}` on `github.com/disaster37/go-kibana-rest/v8`, using types from `github.com/disaster37/go-kibana-rest/v8/kbapi` (`SyntheticsMonitor`, `SyntheticsMonitorConfig`, `HTTPMonitorFields`, `JsonObject`, `APIError`, etc.). Mapping between Terraform models and those structs lives largely in `schema.go` (~1000+ lines) with table-driven tests in `schema_test.go`.

The repository already ships `generated/kbapi` (oapi-codegen) including `/api/synthetics/monitors` operations and rich models (`SyntheticsHttpMonitorFields`, `SyntheticsBrowserMonitorFields`, …). `internal/clients/kibanaoapi.Client` wraps `*kbapi.ClientWithResponses` and is the established pattern for newer resources (streams, dashboards, alerting, …).

## Goals / Non-Goals

**Goals:**

- Route all synthetics monitor **HTTP** CRUD through `kibanaoapi` helpers built on `ClientWithResponses` (`PostSyntheticMonitorsWithResponse`, `GetSyntheticMonitorWithResponse`, `PutSyntheticMonitorWithResponse`, `DeleteSyntheticMonitorWithResponse` or equivalent).
- Preserve every requirement in `openspec/specs/kibana-synthetics-monitor/spec.md` that affects Terraform users: schema shape, validation, composite ID, import, labels version gate, read mapping quirks, error semantics for 404 on read, and delete error surfacing.
- Centralize space header / base URL handling in `kibanaoapi` the same way as other resources (reuse request editor patterns from existing files).
- Map legacy `kbapi.JsonObject` fields to/from `json.RawMessage` or concrete generated pointer fields without changing normalized JSON string behavior in state.

**Non-Goals:**

- Changing the Terraform schema, attribute names, or import ID format.
- Migrating other synthetics resources (`private_location`, `parameter`, …) in the same change.
- Regenerating the OpenAPI bundle (unless a gap is found; assume current `generated/kbapi` is sufficient).
- Rewriting acceptance tests to new fixtures unless behavior forces it (prefer keeping `testdata` HCL).

## Decisions

1. **`kibanaoapi` owns HTTP + status decoding**  
   Rationale: Matches streams/dashboards; keeps `monitor` package focused on Terraform ↔ domain mapping. Helpers accept `context.Context`, space id, monitor id (where applicable), typed or raw bodies, and return typed structs plus `diag.Diagnostics` using `internal/diagutil` patterns.

2. **Generated union bodies (`PostSyntheticMonitorsJSONBody`, `PutSyntheticMonitorJSONBody`)**  
   The generated types wrap `json.RawMessage`. Decision: implement small constructors in `kibanaoapi` that accept the appropriate concrete generated struct (e.g. HTTP/TCP/ICMP/browser monitor + shared config) and marshal into the union field Kibana expects, mirroring the JSON shape the legacy client produced.  
   *Alternatives considered:* (a) keep an internal “wire DTO” struct tagged for JSON and bypass generated body types — rejected as second schema to maintain; (b) raw `json.Marshal` from `map[string]any` — rejected as losing compile-time checks.

3. **404 handling**  
   Map `GetSyntheticMonitorWithResponse` non-success responses to the same behavior as today (`errors.As` into legacy `APIError` with code 404). Use `kibanaoapi`/`diagutil` shared helpers for HTTP status if they exist; otherwise add a minimal status classifier in the new monitor helper file.

4. **Version / labels gate**  
   Keep `enforceVersionConstraints` on `*clients.KibanaScopedClient` as today to avoid scope creep; only the monitor HTTP path switches to OpenAPI.

5. **Incremental replacement**  
   Replace legacy types end-to-end in `monitor` in one PR series (not half-migrated public types). Internal package-private aliases are acceptable short-term inside `kibanaoapi` only.

## Risks / Trade-offs

- **[Risk] JSON shape drift** between legacy structs and generated models → **Mitigation:** golden JSON tests for one monitor per type (create/update body) and compare to captured legacy payloads or Kibana docs; run full acc tests.
- **[Risk] Opaque union marshal bugs** → **Mitigation:** isolate marshal in unit-tested `kibanaoapi` functions; fail fast with clear diagnostics on marshal error.
- **[Risk] Large diff in `schema.go`** → **Mitigation:** consider extracting `api_model_http.go`-style files only if it clarifies review; default to minimal file moves to reduce merge pain.
- **[Trade-off] Two client stacks** (`ScopedClient` for version, `kibanaoapi` for HTTP) until a later consolidation.

## Migration Plan

1. Land `kibanaoapi` monitor helpers + unit tests (no resource wiring).
2. Switch `create`/`read`/`update`/`delete` to helpers; fix compile in `schema.go` mappers.
3. Rewrite `schema_test.go` tables for new types; run `go test ./internal/kibana/synthetics/monitor/...`.
4. Run acceptance tests for all `TestAcc*` monitor tests against a versioned stack (including labels tests on ≥ 8.16).
5. Remove unused legacy imports from `monitor` package; optionally add a `go mod` tidy follow-up if nothing else uses legacy kbapi synthetics types.

Rollback: revert the branch; no state migration required.

## Open Questions

- Whether `GetSyntheticMonitor` response typing is fully adequate in `generated/kbapi` for all nested fields the resource reads, or whether some fields still arrive as raw JSON needing custom unmarshaling.
- Exact Kibana `kbn-xsrf` / space header requirements for these endpoints relative to other `kibanaoapi` calls (validate against `client.go` transport).
