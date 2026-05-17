## ADDED Requirements

### Requirement: Space segment injection with base-path support (REQ-001)

`SpaceAwarePathRequestEditor` SHALL insert `/s/{spaceID}` into the request URL path immediately before the first `/api/` segment. When `space_id` is empty or `"default"`, the request URL path SHALL remain unchanged.

#### Scenario: No base path — non-default space

- GIVEN a Kibana URL with no base-path prefix (e.g. `https://kibana.example.com`)
- AND `space_id = "ops"`
- AND `req.URL.Path = "/api/alerting/rule/{id}"`
- WHEN `SpaceAwarePathRequestEditor` is invoked
- THEN `req.URL.Path` SHALL become `"/s/ops/api/alerting/rule/{id}"`

#### Scenario: With base path — non-default space

- GIVEN a Kibana URL with a base-path prefix (e.g. `https://elk-cluster.org/kibana`, `server.basePath = /kibana`)
- AND `space_id = "ops"`
- AND `req.URL.Path = "/kibana/api/alerting/rule/{id}"`
- WHEN `SpaceAwarePathRequestEditor` is invoked
- THEN `req.URL.Path` SHALL become `"/kibana/s/ops/api/alerting/rule/{id}"`

#### Scenario: No `/api/` anchor — fallback prepend-at-root

- GIVEN `space_id = "ops"`
- AND `req.URL.Path = "/internal/observability/slos/_definitions"`
- WHEN `SpaceAwarePathRequestEditor` is invoked
- THEN `req.URL.Path` SHALL become `"/s/ops/internal/observability/slos/_definitions"`

#### Scenario: Empty space — path unchanged

- GIVEN any `req.URL.Path`
- AND `space_id = ""`
- WHEN `SpaceAwarePathRequestEditor` is invoked
- THEN `req.URL.Path` SHALL remain unchanged

#### Scenario: Default space — path unchanged

- GIVEN any `req.URL.Path`
- AND `space_id = "default"`
- WHEN `SpaceAwarePathRequestEditor` is invoked
- THEN `req.URL.Path` SHALL remain unchanged

### Requirement: Compatibility with direct BuildSpaceAwarePath callers (REQ-002)

`BuildSpaceAwarePath(spaceID, basePath string) string` SHALL remain unchanged. Direct callers that pass raw API-relative paths (such as `enrollment_tokens.go` and `synthetics_monitor.go`) are unaffected by `server.basePath` because they concatenate paths directly with `client.URL` without going through the `kbapi` URL resolver.

#### Scenario: Direct caller with non-default space

- GIVEN `spaceID = "ops"` and `basePath = "/api/fleet/enrollment_api_keys"`
- WHEN `BuildSpaceAwarePath` is called
- THEN the result SHALL be `"/s/ops/api/fleet/enrollment_api_keys"`

#### Scenario: Direct caller with default space

- GIVEN `spaceID = "default"` and `basePath = "/api/fleet/enrollment_api_keys"`
- WHEN `BuildSpaceAwarePath` is called
- THEN the result SHALL be `"/api/fleet/enrollment_api_keys"`

### Requirement: No call-site changes required (REQ-003)

The fix to `SpaceAwarePathRequestEditor` SHALL require no signature or call-site changes for any of its existing callers across `internal/clients/kibanaoapi/` and `internal/clients/fleet/`.

#### Scenario: Existing call site passes through correctly

- GIVEN an existing call site that invokes `SpaceAwarePathRequestEditor(spaceID)` and registers it as a `kbapi` request editor
- WHEN the `kbapi` client resolves an operation path relative to a Kibana URL that includes a base-path prefix
- THEN the resulting request URL path SHALL correctly place the space segment before the `/api/` anchor
- AND no change to the call site is required
