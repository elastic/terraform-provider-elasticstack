## Why

`GetPackages` in `internal/clients/fleet/packages.go` panics with a nil pointer dereference (SIGSEGV) when the Fleet EPM list endpoint returns HTTP 200 but with a `Content-Type` header that does not contain `"json"`.

The generated kbapi parser (`ParseGetFleetEpmPackagesResponse`) only assigns `response.JSON200 = &dest` when **both** conditions hold:

1. `rsp.StatusCode == 200`
2. `strings.Contains(rsp.Header.Get("Content-Type"), "json")`

If condition 2 fails—for any reason—`JSON200` remains `nil`. The code then dereferences it unconditionally at line 175 (`return resp.JSON200.Items, nil`) and at line 187 (`return retryResp.JSON200.Items, nil`), causing the provider to crash. This crash surfaces as `panic: runtime error: invalid memory address or nil pointer dereference` in the Terraform logs and a `Plugin did not respond` error to the user.

## What Changes

Add defensive nil guards in `GetPackages` so that a nil `JSON200` field results in a descriptive `diag.Diagnostics` error rather than a panic.

**Specific locations:**

- `internal/clients/fleet/packages.go:175` — before `return resp.JSON200.Items, nil`
- `internal/clients/fleet/packages.go:187` — before `return retryResp.JSON200.Items, nil`

**Behaviour after fix:**

- If `resp.JSON200 == nil` after an HTTP 200 response, return a `diag.Diagnostics` error with a clear message ("Fleet returned HTTP 200 but response body could not be decoded as JSON; check the Content-Type header").
- If `retryResp.JSON200 == nil` after the compatibility retry HTTP 200 response, return the same kind of error.
- All other behaviour remains unchanged.

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- **`fleet-integration`**: `GetPackages` now returns a structured error instead of panicking when the Kibana Fleet EPM list endpoint responds with HTTP 200 but without a parseable JSON body.

## Impact

- **Users**: Provider no longer crashes with a SIGSEGV on `data "elasticstack_fleet_integration"` reads or any code path that calls `GetPackages`. Instead it surfaces a clear diagnostic error.
- **Code**: `internal/clients/fleet/packages.go` — two nil-guard additions (~4 lines each).
- **Maintenance**: No schema changes; no state changes; no test fixture changes.
