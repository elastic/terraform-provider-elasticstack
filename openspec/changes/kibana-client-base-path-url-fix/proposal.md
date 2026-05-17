## Why

When Kibana is deployed behind a reverse proxy with `server.basePath` set (for example `/kibana` at `https://elk-cluster.org/kibana`), every Kibana and Fleet resource managed by this provider generates an invalid API URL and receives a 404 ([#2804](https://github.com/elastic/terraform-provider-elasticstack/issues/2804)).

**Concrete failure** with `space_id = "pd-core"` and `server.basePath = /kibana`:

```
POST /s/pd-core/kibana/api/alerting/rule/{id}   →  404
```

Expected:

```
POST /kibana/s/pd-core/api/alerting/rule/{id}
```

**Root cause**: `SpaceAwarePathRequestEditor` in `internal/clients/kibanautil/spaces.go` calls `BuildSpaceAwarePath` which unconditionally prepends `/s/{spaceID}` to the entire request path. The `kbapi` generated client resolves operation paths relative to the full Kibana server URL (including the base-path prefix), so `req.URL.Path` already contains the base-path prefix at the time the editor runs. Prepending `/s/{spaceID}` at the front of `/kibana/api/...` produces `/s/{spaceID}/kibana/api/...` rather than `/kibana/s/{spaceID}/api/...`.

This affects all ~85 non-test call sites of `SpaceAwarePathRequestEditor` across `internal/clients/kibanaoapi/` and `internal/clients/fleet/`.

## What Changes

Modify `SpaceAwarePathRequestEditor` in `internal/clients/kibanautil/spaces.go` to insert `/s/{spaceID}` immediately **before** the first `/api/` segment rather than at the start of the path. This correctly handles both the no-base-path case (path starts with `/api/`) and the base-path case (path starts with `/kibana/api/` or any other prefix).

`BuildSpaceAwarePath` (used by the two direct callers that pass raw API-relative paths: `enrollment_tokens.go:68` and `synthetics_monitor.go:101`) is left unchanged because those callers are not affected by the bug.

**Files changed:**

- `internal/clients/kibanautil/spaces.go` — logic change to `SpaceAwarePathRequestEditor`; add `strings` import
- `internal/clients/kibanautil/spaces_test.go` — add test cases covering base-path configurations

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `kibana-space-aware-url-construction`: Fix URL path construction in `SpaceAwarePathRequestEditor` so that Kibana's `server.basePath` prefix is preserved ahead of the injected space segment.

## Impact

- **Specs**: Delta under `openspec/changes/kibana-client-base-path-url-fix/specs/kibana-space-aware-url-construction/spec.md`.
- **Implementation** (future): Two-line logic change to `SpaceAwarePathRequestEditor` plus unit-test additions. No call-site changes required. All resources using `SpaceAwarePathRequestEditor` are fixed automatically.
