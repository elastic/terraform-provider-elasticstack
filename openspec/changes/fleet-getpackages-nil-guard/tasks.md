## 1. Add nil guard on primary response

- [ ] 1.1 In `internal/clients/fleet/packages.go`, inside `case http.StatusOK` of `GetPackages`, add a nil check before `return resp.JSON200.Items, nil`:

  ```go
  case http.StatusOK:
      if resp.JSON200 == nil {
          return nil, diag.Diagnostics{
              diag.NewErrorDiagnostic(
                  "Unexpected Fleet response",
                  "Fleet returned HTTP 200 for the packages list endpoint but the response body could not be decoded as JSON. Verify the Kibana Fleet endpoint is reachable and returns a JSON Content-Type.",
              ),
          }
      }
      return resp.JSON200.Items, nil
  ```

## 2. Add nil guard on retry response

- [ ] 2.1 In `internal/clients/fleet/packages.go`, inside the `if retryResp.StatusCode() == http.StatusOK` branch of `GetPackages`, add a nil check before `return retryResp.JSON200.Items, nil`:

  ```go
  if retryResp.StatusCode() == http.StatusOK {
      if retryResp.JSON200 == nil {
          return nil, diag.Diagnostics{
              diag.NewErrorDiagnostic(
                  "Unexpected Fleet response",
                  "Fleet returned HTTP 200 for the packages list endpoint but the response body could not be decoded as JSON. Verify the Kibana Fleet endpoint is reachable and returns a JSON Content-Type.",
              ),
          }
      }
      return retryResp.JSON200.Items, nil
  }
  ```

## 3. OpenSpec

- [ ] 3.1 Ensure the delta spec at `openspec/changes/fleet-getpackages-nil-guard/specs/fleet-integration/spec.md` is aligned with the implementation.
- [ ] 3.2 After merge decision: sync into `openspec/specs/fleet-integration/spec.md` or archive the change per project workflow; run `make check-openspec`.
