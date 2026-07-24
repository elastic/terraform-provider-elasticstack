## 1. Introduce `spaceScope` (`internal/fleet/integration/space_helpers.go`)

- [ ] 1.1 Add the `spaceScope` struct (`id string`, `aware bool`).
- [ ] 1.2 Add `resolveSpaceScope(ctx, client clients.MinVersionEnforceable, spaceID types.String, diags *diag.Diagnostics) spaceScope`, folding in the existing `supportsSpaceAwareIntegration` version-gate logic.
- [ ] 1.3 Change `fleetPackageInstalled` to take `(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo, scope spaceScope) bool` instead of `(pkg, spaceID string, spaceAware bool)`.
- [ ] 1.4 Remove or keep `resolveSpaceAware`/`supportsSpaceAwareIntegration` only as internals used by `resolveSpaceScope`; do not leave dead exported-equivalent duplicates.

## 2. Fix the defect in `writeIntegration` and rework `waitForFleetIntegrationInstalled` (`internal/fleet/integration/create.go`)

- [ ] 2.1 Change `waitForFleetIntegrationInstalled` signature to `(ctx, fleetClient, name, version string, scope spaceScope) error`; internally call `fleet.GetPackage(ctx, fleetClient, name, version, scope.id)` and `fleetPackageInstalled(pkg, scope)`.
- [ ] 2.2 In `writeIntegration`, resolve `scope := resolveSpaceScope(ctx, client, planModel.SpaceID, &diags)` once, before the install call; return early on `diags.HasError()`.
- [ ] 2.3 Replace the buggy call `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, "", false)` with `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, scope)`. This is the actual fix for issue #4282.
- [ ] 2.4 Replace `globallyInstalled := fleetPackageInstalled(pkg, "", false)` with `fleetPackageInstalled(pkg, spaceScope{})` (explicit default-space scope — unchanged semantics, no additional API call).
- [ ] 2.5 Replace `installedInTargetSpace := fleetPackageInstalled(pkg, spaceID, true)` with `fleetPackageInstalled(pkg, spaceScope{id: scope.id, aware: true})`, preserving today's unconditional `aware: true` for this specific check.
- [ ] 2.6 Update `installInSpace` to accept `scope.id` (or the full `scope`, whichever keeps the signature simplest) and pass it through to `waitForFleetIntegrationInstalled`; keep its internal `supportsSpaceAwareIntegration` re-check unchanged (it independently gates the cross-space Kibana-assets path).
- [ ] 2.7 Confirm `installOptions.SpaceID` (used for the `InstallPackage` call) and `scope.id` agree — both derive from `planModel.SpaceID` — and simplify to a single source of truth if straightforward.

## 3. Thread `spaceScope` through read and delete for consistency (no behavior change)

- [ ] 3.1 `internal/fleet/integration/read.go`: replace `spaceAware := resolveSpaceAware(ctx, client, model.SpaceID, &diags)` with `scope := resolveSpaceScope(ctx, client, model.SpaceID, &diags)`; update the `fleetPackageInstalled(pkg, spaceID, spaceAware)` call to `fleetPackageInstalled(pkg, scope)` (using `scope.id` in place of the existing `spaceID` parameter where they refer to the same value — confirm they're always equal, since `spaceID` is passed in separately from entitycore).
- [ ] 3.2 `internal/fleet/integration/delete.go`: same substitution in `deleteIntegration`; keep `isInstalledInMultipleSpaces(pkg, spaceID)` and `deleteKibanaAssetsWithFallback(ctx, fleetClient, name, version, spaceID, force)` taking a plain `spaceID string` (extract from `scope.id`), since they don't need `aware`.
- [ ] 3.3 Re-run a diff review to confirm no observable behavior changed in `read.go`/`delete.go` — this step is a pure refactor.

## 4. Acceptance test proving the fix (`internal/fleet/integration/acc_test.go`)

- [ ] 4.1 Add `TestAccResourceIntegration_SpaceRestrictedKey`, gated by `versionutils.SkipIfUnsupported(t, integration.MinVersionSpaceAwareIntegration, versionutils.FlavorAny)`.
- [ ] 4.2 Add `testdata/TestAccResourceIntegration_SpaceRestrictedKey/` config directory: `elasticstack_kibana_space` (random space id) + `elasticstack_elasticsearch_security_api_key` with `role_descriptors` embedding a Kibana application privilege scoped to `resources: ["space:${space_id}"]` with Fleet feature privileges + `elasticstack_fleet_integration` (`name = "tcp"`, `version = "1.16.0"`, `space_id = <space_id>`, `kibana_connection { api_key = <restricted key> }`).
- [ ] 4.3 Positive assertion: apply succeeds; `testAccCheckIntegrationInstalledInSpace("tcp", "1.16.0", spaceID)` passes.
- [ ] 4.4 Negative guard: a `resource.TestCheckFunc` that builds a Fleet client from the restricted key's encoded value and calls `fleet.GetPackage(ctx, client, "tcp", "1.16.0", "")` (default space), asserting the result is a 403/forbidden diagnostic and not a successful read.
- [ ] 4.5 Confirm the exact Fleet EPM privilege set empirically against a running 9.1+ stack (see design.md Risks) — adjust the `role_descriptors` privileges list if the restricted key 403s on the legitimately-scoped path too.

## 5. Verify

- [ ] 5.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fleet-integration-space-scoped-install-poll --type change` and fix any reported issues.
- [ ] 5.2 Run `make build`.
- [ ] 5.3 Run `go vet ./internal/fleet/integration/...`.
- [ ] 5.4 Bring up a 9.1+ Elastic Stack and run `make testacc TESTARGS='-run TestAccResourceIntegration_SpaceRestrictedKey'` (or the repo's docker-testacc equivalent); confirm it fails pre-fix (403 during post-install wait) and passes post-fix.
- [ ] 5.5 Re-run `TestAccResourceIntegration_MultiSpaceInstall`, `TestAccResourceIntegration_MultiSpaceDelete`, and `TestAccResourceIntegration_SpaceAwareDrift` to confirm no regression from the `spaceScope` refactor.
- [ ] 5.6 Run `make check-lint` on touched files.
