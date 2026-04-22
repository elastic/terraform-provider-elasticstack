## Context

`internal/kibana/synthetics/parameter` already uses `synthetics.GetKibanaOAPIClientFromScopedClient` and `kibanaoapi.Client` for create, read, and update. `delete.go` alone still calls `synthetics.GetKibanaClientFromScopedClient` and the legacy `go-kibana-rest` synthetics `Parameter.Delete` API. The canonical spec (`openspec/specs/kibana-synthetics-parameter/spec.md`) still documents that split.

## Goals / Non-Goals

**Goals:**

- Use the generated `kbapi` client (`DeleteParameterWithResponse` on `kibanaoapi.Client`, same type as read/create/update) for delete so all CRUD shares one HTTP stack and response parsing patterns.
- Keep create/update request construction identical in spirit: build the Go request DTO, `json.Marshal` it, and call `PostParametersWithBodyWithResponse` / `PutParameterWithBodyWithResponse` with `application/json` to work around oapi-codegen oneOf limitations.
- Preserve post-create and post-update `readState` / `GetParameterWithResponse` behavior and all `share_across_spaces` / `namespaces` mapping rules.

**Non-Goals:**

- Regenerating or editing the OpenAPI spec or `generated/kbapi` beyond what already exists.
- Replacing manual JSON marshalling with generated `PostParameters` / `PutParameter` helpers until upstream oneOf support is fixed.
- Migrating other synthetics resources (monitor, private location) off the legacy client.

## Decisions

1. **Delete implementation** — Switch `delete.go` to `GetKibanaOAPIClientFromScopedClient` and invoke `kibanaClient.API.DeleteParameterWithResponse(ctx, resourceID)` (after the same composite-id normalization used elsewhere). **Rationale:** Matches read/create/update, removes the second client factory path, and uses the same base URL and auth as other parameter endpoints. **Alternative considered:** Keep legacy delete “because it already works”; rejected because it blocks client consolidation and duplicates diagnostics behavior.

2. **Error handling** — Mirror `read.go` patterns where applicable: treat transport errors as diagnostics; if the API returns an unexpected status, surface it clearly (consistent with other `WithResponse` usages in this package). **Rationale:** Align delete with the OpenAPI client error model already used for get/post/put.

3. **Spec / notes cleanup** — When syncing to main spec after implementation, update schema notes and REQ-001 / REQ-002 / REQ-005 to describe a single OpenAPI client; add an explicit requirement for the oneOf JSON workaround so it is preserved during future refactors.

## Risks / Trade-offs

- **[Risk] Subtle behavioral difference between legacy delete and `kbapi` delete** (status codes, default space headers) **→ Mitigation:** Acceptance tests for the resource should continue to pass; compare `DeleteParameter` path in generated code to the legacy route; manually verify against a real Kibana if CI stack differs.

- **[Risk] `DeleteParameterWithResponse` returns a body or status we do not handle** **→ Mitigation:** Follow existing resource patterns for `WithResponse` calls; treat 2xx as success; map non-success to diagnostics with response details when the client exposes them.

## Migration Plan

1. Implement delete via `kbapi`; run `make build` and targeted acceptance tests for `TestSyntheticParameterResource` when a stack is available.
2. Run `make check-openspec` / lint as required by the repo.
3. After code review, use the OpenSpec sync workflow to merge delta requirements into `openspec/specs/kibana-synthetics-parameter/spec.md` (separate from this proposal folder).

## Open Questions

- None blocking: confirm whether any enterprise Kibana build omits `DeleteParameter` from the bundled OpenAPI (unlikely if GET/POST/PUT are already used from the same client).
