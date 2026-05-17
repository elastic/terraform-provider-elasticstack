## Context

The provider uses `SpaceAwarePathRequestEditor` (in `internal/clients/kibanautil/spaces.go`) as an `oapi-codegen` request editor to inject the Kibana space segment into API paths. Every `kibanaoapi` and `fleet` client call site passes this editor via `kbapi.WithRequestEditorFn`.

The `kbapi` generated client (`generated/kbapi/kibana.gen.go`) constructs each operation URL by calling `serverURL.Parse(operationPath)`, where `serverURL` is the full Kibana URL including any base-path prefix (e.g. `https://host/kibana`). This means `req.URL.Path` already contains the base-path prefix when the request editor is invoked.

Current `SpaceAwarePathRequestEditor` delegates to `BuildSpaceAwarePath(spaceID, req.URL.Path)`, which prepends `/s/{spaceID}` to the entire path. With base-path `/kibana` the request path becomes `/s/{spaceID}/kibana/api/...` instead of the correct `/kibana/s/{spaceID}/api/...`.

## Goals

- Fix the incorrect URL path for all Kibana and Fleet resources when `server.basePath` is configured.
- Require no changes to any call site of `SpaceAwarePathRequestEditor`.
- Do not change `BuildSpaceAwarePath` (used by the two direct callers that pass raw API-relative paths).
- Maintain the existing correct behavior when no base path is configured.

## Non-Goals

- Changing the signature of `SpaceAwarePathRequestEditor` (no threading of base-path through all call sites).
- Fixing Elasticsearch endpoint URL handling (not affected).
- Supporting a base path that itself contains the literal string `/api/` (this is not a valid Kibana `server.basePath` pattern; Kibana's own documentation warns against it).

## Decisions

### Decision 1: Anchor-based injection — insert `/s/{spaceID}` before the first `/api/` segment

**Rationale:** The `/api/` segment is the stable anchor in every Kibana API path. A base-path prefix always appears before it; space and resource segments always appear after it. Splitting at `/api/` places the space segment in the correct position without knowing the actual base-path value.

**Behavior before / after with `server.basePath = /kibana`:**

| Stage | Path |
|-------|------|
| kbapi resolves path | `/kibana/api/alerting/rule/{id}` |
| current editor output | `/s/{spaceID}/kibana/api/alerting/rule/{id}` (WRONG) |
| fixed editor output | `/kibana/s/{spaceID}/api/alerting/rule/{id}` (CORRECT) |

**Behavior without base path (unchanged):**

| Stage | Path |
|-------|------|
| kbapi resolves path | `/api/alerting/rule/{id}` |
| fixed editor output | `/s/{spaceID}/api/alerting/rule/{id}` (CORRECT, same as today) |

**Fallback:** If no `/api/` segment is found (non-standard internal path), the editor falls back to prepending at root — same as the current behavior. This edge case is only reachable if a future `kbapi` update introduces paths that neither start with `/api/` nor contain `/api/` anywhere and are used with `SpaceAwarePathRequestEditor`. Such paths are currently non-existent in the generated client for this editor's call sites.

**Alternative (Approach B — explicit base-path threading) rejected:** Passing the parsed base-path string through all ~85 call sites and both client structs is an invasive refactor for a hypothetical edge case (a base path that contains `/api/`). Approach A is correct for all realistic Kibana deployments.

### Decision 2: No changes to `BuildSpaceAwarePath`

`BuildSpaceAwarePath` is called directly by `enrollment_tokens.go:68` and `synthetics_monitor.go:101` with raw API-relative paths (e.g. `/api/fleet/enrollment_api_keys?...`) that do not carry the base-path prefix. Those callers are unaffected by the bug.

## Risks / Trade-offs

- **Base-path containing `/api/`**: If a user configures `server.basePath = /api/something`, the anchor-based split would insert the space segment in the wrong place. This configuration is not supported by Kibana's own validation and is not seen in practice; documenting it as unsupported is sufficient.
- **Future `/internal/` paths**: If the generated `kbapi` client introduces paths that use `SpaceAwarePathRequestEditor` but do not contain `/api/` (e.g. paths rooted at `/internal/`), the fallback prepend-at-root behavior would be incorrect. Mitigation: verify at code-review time when updating the generated client.

## Open Questions

- Are there Kibana or Fleet API paths in the `kbapi` generated client that use `SpaceAwarePathRequestEditor` but do **not** contain `/api/` (e.g. `/internal/` paths)? One `/internal/` path (`/s/%s/internal/observability/slos/_definitions`) was found but it already has the space as a path parameter and does not use the editor. New `/internal/` paths that use the editor would require re-evaluation of the anchor approach.
- Can the issue reporter share `TF_LOG=trace` output showing the exact request URL? This would confirm the root-cause analysis, though the fix is unambiguous from code inspection alone.
