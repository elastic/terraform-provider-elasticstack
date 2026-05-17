## 1. Implementation

- [ ] 1.1 In `internal/clients/kibanautil/spaces.go`, add `"strings"` to the import block and replace the body of `SpaceAwarePathRequestEditor` so that it inserts `/s/{spaceID}` immediately before the first `/api/` segment of `req.URL.Path`, falling back to prepend-at-root if no `/api/` anchor is found. `BuildSpaceAwarePath` is not changed.

  Replacement logic (illustrative):
  ```go
  func SpaceAwarePathRequestEditor(spaceID string) func(ctx context.Context, req *http.Request) error {
      return func(_ context.Context, req *http.Request) error {
          if spaceID == "" || spaceID == "default" {
              return nil
          }
          path := req.URL.Path
          if idx := strings.Index(path, "/api/"); idx != -1 {
              req.URL.Path = path[:idx] + "/s/" + spaceID + path[idx:]
          } else {
              req.URL.Path = "/s/" + spaceID + path
          }
          return nil
      }
  }
  ```

## 2. Testing

- [ ] 2.1 In `internal/clients/kibanautil/spaces_test.go`, extend `TestSpaceAwarePathRequestEditor` with cases covering base-path configurations:
  - Path `/kibana/api/alerting/rule/abc` with space `ops` → `/kibana/s/ops/api/alerting/rule/abc`
  - Path `/kibana/api/alerting/rule/abc` with empty space → `/kibana/api/alerting/rule/abc` (unchanged)
  - Path `/kibana/api/alerting/rule/abc` with space `default` → `/kibana/api/alerting/rule/abc` (unchanged)
  - Path `/api/alerting/rule/abc` with space `ops` → `/s/ops/api/alerting/rule/abc` (no regression)
  - Path `/nested/prefix/api/alerting/rule/abc` with space `ops` → `/nested/prefix/s/ops/api/alerting/rule/abc`
  - Path `/internal/observability/slos/_definitions` with space `ops` → `/s/ops/internal/observability/slos/_definitions` (fallback when `/api/` anchor is absent)

- [ ] 2.2 Run `go test ./internal/clients/kibanautil/...` and confirm all tests pass.

## 3. Validation

- [ ] 3.1 Run `make build` to ensure the change compiles.
- [ ] 3.2 Run `make check-lint` to ensure lint passes.
- [ ] 3.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-client-base-path-url-fix --type change` and confirm the change is valid.
