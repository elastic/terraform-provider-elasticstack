## ADDED Requirements

### Requirement: GetPackages nil-safe response handling (REQ-019)

The `GetPackages` internal client function (`internal/clients/fleet/packages.go`) SHALL guard against a nil `JSON200` field on the Fleet EPM packages list response even when the HTTP status code is 200.

The generated Kibana API client (`ParseGetFleetEpmPackagesResponse`) only populates the `JSON200` field when **both** the HTTP status is 200 **and** the `Content-Type` response header contains the string `"json"`. If the `Content-Type` condition is not met the field remains nil. Dereferencing a nil `JSON200` causes an unrecoverable provider crash (SIGSEGV). Both the primary code path and the prerelease-parameter compatibility retry path are affected.

`GetPackages` SHALL:

1. After receiving an HTTP 200 response (primary path), check whether `resp.JSON200` is nil before accessing `resp.JSON200.Items`. When nil, the function SHALL return `nil, diag.Diagnostics` containing an `ErrorDiagnostic` with summary `"Unexpected Fleet response"` and a detail explaining that the Fleet endpoint returned HTTP 200 but the response body could not be decoded as JSON, and advising the operator to verify the Kibana Fleet endpoint is reachable and returns a JSON `Content-Type`.

2. After receiving an HTTP 200 response on the prerelease-parameter compatibility retry (the `retryResp` path), apply the same nil guard before accessing `retryResp.JSON200.Items`, returning the same form of error when nil.

In both cases the function SHALL NOT panic.

#### Scenario: HTTP 200 with non-JSON Content-Type (primary path)

- GIVEN the Kibana Fleet EPM packages list endpoint responds with HTTP 200
- AND the response `Content-Type` header does not contain `"json"`
- WHEN `GetPackages` processes the response
- THEN `resp.JSON200` SHALL be nil
- AND `GetPackages` SHALL return a non-nil `diag.Diagnostics` containing an error diagnostic
- AND the provider SHALL NOT crash with a nil pointer dereference

#### Scenario: HTTP 200 with non-JSON Content-Type (retry path)

- GIVEN the Kibana Fleet EPM packages list endpoint responds with HTTP 400 containing "prerelease" (triggering the compatibility retry)
- AND the retry response has HTTP 200 but a `Content-Type` header that does not contain `"json"`
- WHEN `GetPackages` processes the retry response
- THEN `retryResp.JSON200` SHALL be nil
- AND `GetPackages` SHALL return a non-nil `diag.Diagnostics` containing an error diagnostic
- AND the provider SHALL NOT crash with a nil pointer dereference

#### Scenario: HTTP 200 with JSON Content-Type (no change)

- GIVEN the Kibana Fleet EPM packages list endpoint responds with HTTP 200
- AND the response `Content-Type` header contains `"json"`
- WHEN `GetPackages` processes the response
- THEN `resp.JSON200` SHALL be non-nil
- AND `GetPackages` SHALL return the items list and nil diagnostics as before
