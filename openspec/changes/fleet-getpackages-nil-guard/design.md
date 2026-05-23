## Context

The Fleet integration data source (`internal/fleet/integration_ds/read.go`) calls `GetPackages` (`internal/clients/fleet/packages.go`) to retrieve the list of available packages. The generated kbapi function `ParseGetFleetEpmPackagesResponse` (in `generated/kbapi/kibana.gen.go`) only populates the `JSON200` field of the response struct when both the HTTP status is 200 **and** the `Content-Type` header contains the string `"json"`. If `Content-Type` is absent or does not contain `"json"`, the field stays `nil`. `GetPackages` asserts a 200 status code in its `switch` branch but never checks whether `resp.JSON200` is non-nil before dereferencing it. The result is a nil pointer dereference (SIGSEGV) that crashes the provider.

The same race exists in the prerelease-parameter compatibility retry path: if `retryResp.StatusCode() == http.StatusOK` is true but `retryResp.JSON200` is nil (for the same reason), the next line panics.

## Goals / Non-Goals

**Goals:**

- Prevent the provider from crashing when the Fleet EPM packages endpoint returns HTTP 200 with a non-JSON `Content-Type`.
- Return a descriptive `diag.Diagnostics` error so users receive actionable output instead of a SIGSEGV.

**Non-goals:**

- Changing any resource or data source schema.
- Modifying the generated kbapi client.
- Altering any other API call paths outside `GetPackages`.

## Decisions

- **Nil guard location**: immediately inside `case http.StatusOK` in `GetPackages`, before the `return resp.JSON200.Items` dereference. Likewise immediately inside the `if retryResp.StatusCode() == http.StatusOK` branch.
- **Error message**: `"Fleet returned HTTP 200 for the packages list endpoint but the response body could not be decoded as JSON. Verify the Kibana Fleet endpoint is reachable and returns a JSON Content-Type."` This is informative without leaking internal implementation details.
- **Error severity**: `diag.NewErrorDiagnostic` — a missing package list is always a hard failure; a warning would leave the data source in an undefined state.
- **No retry or fallback**: the condition is a misconfigured server (wrong Content-Type), not a transient error; retrying is not appropriate.

## Risks / Trade-offs

- **None significant**: this is a two-site, four-line change. The nil guard executes only in an anomalous code path that currently causes a crash; any production regression is strictly less bad than the status quo.

## Open Questions

- None.
