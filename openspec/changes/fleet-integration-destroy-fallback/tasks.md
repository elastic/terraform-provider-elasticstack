## 1. Define the install-space error sentinel

- [x] 1.1 In `internal/fleet/integration/delete.go`, define a package-level constant `installSpaceDeleteRejectedMsg` set to "space where the package was installed" (a stable substring of the Fleet 9.5 400 message); when matching diagnostics, normalize whitespace/newlines before checking `Contains`.

## 2. Implement the DeleteKibanaAssets fallback in deleteIntegration

- [x] 2.1 After calling `fleet.DeleteKibanaAssets`, inspect the returned `diag.Diagnostics`. If any diagnostic's `Detail()` or `Summary()` contains `installSpaceDeleteRejectedMsg`, clear those diagnostics and proceed to the fallback branch. Otherwise surface the diagnostics and return.
- [x] 2.2 In the fallback branch: call `tflog.Debug(ctx, "DeleteKibanaAssets rejected by Fleet (install space); falling back to Uninstall", map[string]any{attrName: name, attrVersion: version, "space_id": spaceID})`.
- [x] 2.3 After the debug log, call `fleet.Uninstall(ctx, fleetClient, name, version, spaceID, force)` and append its diagnostics.
- [x] 2.4 Return the diagnostics from the fallback `Uninstall` call (not the original 400 diagnostics).
- [x] 2.5 **(added after review)** Gate 2.2–2.4 on `force == true`, since `fleet.Uninstall` removes the package globally across all spaces, not just the target space. When `force` is false, do NOT call `fleet.Uninstall`; instead return a distinct, actionable error diagnostic explaining that the target space is the package's install space, the package is also installed elsewhere, and the caller must destroy the other space's resource(s) first or set `force = true`.

## 3. Unit tests

- [x] 3.1 Add a unit test in `internal/fleet/integration/` (or `internal/clients/fleet/`) that mocks the Fleet API to return the 400 install-space response and verifies that the fallback logic falls back to `Uninstall` when `force = true` (i.e. `Uninstall` is called exactly once and the returned diagnostics are clean).
- [x] 3.2 Add a unit test that verifies a generic 400 (different message) is still surfaced as an error and does NOT call `Uninstall`, regardless of `force`.
- [x] 3.3 Add a unit test that verifies the happy path (HTTP 200 from `DeleteKibanaAssets`) still returns no diagnostics and does NOT call `Uninstall`.
- [x] 3.4 **(added after review)** Add a unit test that verifies: install-space 400 + `force = false` → `Uninstall` is NOT called, and the returned diagnostics contain an actionable error (distinct from the raw Fleet 400 body) directing the caller to destroy the other space's resource(s) or set `force = true`.

## 4. Validation and cleanup

- [x] 4.1 Run `make build` — confirm it compiles cleanly.
- [x] 4.2 Run `make check-lint` — fix any linting issues.
- [x] 4.3 Run `make check-openspec` — confirm this change validates cleanly.
- [x] 4.4 Run unit tests for the affected packages: `go test ./internal/fleet/integration/... ./internal/clients/fleet/...`.
