## Why

`elasticstack_fleet_integration` scopes the install call (`fleet.InstallPackage`) to `space_id` via `installOptions.SpaceID`, but the post-install poll that immediately follows it in `writeIntegration` (`internal/fleet/integration/create.go:80`) hard-codes `spaceID = ""` and `spaceAware = false`. The poll therefore always calls `fleet.GetPackage(ctx, fleetClient, name, version, "")`, which resolves to the unscoped `/api/fleet/epm/packages/{name}/{version}` endpoint regardless of the configured `space_id`. An API key that is scoped only to a custom Kibana space has no access to the default space and gets HTTP 403 on every poll, so `terraform apply` fails even though the install itself targeted the correct space (issue [#4282](https://github.com/elastic/terraform-provider-elasticstack/issues/4282)).

Only this first poll is affected — the later poll inside `installInSpace` (`create.go:149`) already passes `spaceID, true` correctly, and the read/delete paths are already space-scoped.

## What Changes

Per the issue's implementation-research comment (adopted as the authoritative recommendation), consolidate the loosely-coupled `spaceID string` / `spaceAware bool` pair into a single resolved `spaceScope` value inside `internal/fleet/integration/`, and use it everywhere a space context is threaded through create/read/delete. This structurally removes the `"", false` call-site pattern that produced the bug, rather than only patching the one call.

- Add a `spaceScope` type (`id string`, `aware bool`) and a `resolveSpaceScope` constructor in `space_helpers.go` that folds in the existing `MinVersionSpaceAwareIntegration` (9.1.0) capability check.
- Change `fleetPackageInstalled` and `waitForFleetIntegrationInstalled` to take a `spaceScope` instead of separate `(spaceID string, spaceAware bool)` parameters.
- In `writeIntegration` (`create.go`), resolve `scope` once from `planModel.SpaceID` and pass it to the first post-install wait call (replacing the hard-coded `"", false`), to the subsequent `GetPackage`/`fleetPackageInstalled` calls, and into `installInSpace`.
- Thread the same `spaceScope` type through `readIntegration` (`read.go`) and `deleteIntegration` (`delete.go`), replacing their existing `resolveSpaceAware` + separate `spaceID` parameter usage. Behavior of read and delete is unchanged; this is a refactor to remove the duplicated sentinel/bool convention, not a functional change to those paths.
- Add an acceptance test that reproduces the 403 with a space-restricted Elastic API key (built entirely in Terraform: a `elasticstack_kibana_space` plus an `elasticstack_elasticsearch_security_api_key` whose `role_descriptors` grant Fleet privileges scoped to only that space via a Kibana application-privilege resource entry), asserting both that the space-scoped install now succeeds and that the same restricted key has no default-space access (so an over-broad test key cannot mask the bug).

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `fleet-integration`: Strengthen the create/update install-options requirement (REQ-011) so the post-install wait that immediately follows the regular install API call is explicitly required to use the same space context (`space_id` and space-aware capability) as the install call itself, not the default space, regardless of whether a subsequent cross-space Kibana-assets install is later needed.

## Impact

- **Code**: `internal/fleet/integration/space_helpers.go` (new `spaceScope` type and resolver), `internal/fleet/integration/create.go` (fix + `installInSpace`/`waitForFleetIntegrationInstalled` signatures), `internal/fleet/integration/read.go`, `internal/fleet/integration/delete.go` (both switched to `spaceScope` for consistency; no behavior change).
- **Tests**: New `TestAccResourceIntegration_SpaceRestrictedKey` acceptance test in `internal/fleet/integration/acc_test.go` plus a new `testdata/TestAccResourceIntegration_SpaceRestrictedKey/` config directory. Existing space-aware acceptance tests (`TestAccResourceIntegration_MultiSpaceInstall`, `_MultiSpaceDelete`, `_SpaceAwareDrift`) must continue to pass unchanged after the `spaceScope` refactor.
- **APIs / dependencies**: No new dependencies, no schema changes, no new public attributes. Space restriction in the test is expressed using the existing `elasticstack_elasticsearch_security_api_key` `role_descriptors` field (raw ES role-descriptor JSON) — no first-class "restrict to space" field is added.
- **Risk**: Low-to-moderate. The one-line fix itself is low risk; the `spaceScope` refactor touches all three CRUD entry points but is intended to be behavior-preserving for read/delete. The main open risk is whether pre-9.1 Kibana's `GetPackage` response populates enough information for `fleetPackageInstalled` to observe a plain single-space install as "installed" when `aware` is derived unconditionally from `id != ""` — see Open Questions in `design.md`.
- **Out of scope**: Changes to `MinVersionSpaceAwareIntegration` semantics or its use in `installInSpace`'s cross-space Kibana-assets path (that gate is for the genuinely 9.1+-only multi-space asset feature). Other Fleet resources (e.g. `elasticstack_fleet_integration_policy`). The `/api/status` version-enforcement endpoint, which is not space-scoped anywhere in the provider and is unrelated to this defect.
