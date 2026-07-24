## 1. Introduce `spaceScope` (`internal/fleet/integration/space_helpers.go`)

- [x] 1.1 Add the `spaceScope` struct (`id string`, `aware bool`).
- [x] 1.2 Add `resolveSpaceScope(ctx, client clients.MinVersionEnforceable, spaceID types.String, diags *diag.Diagnostics) spaceScope`, folding in the existing `supportsSpaceAwareIntegration` version-gate logic.
- [x] 1.3 Change `fleetPackageInstalled` to take `(pkg *kbapi.KibanaHTTPAPIsGetPackageInfo, scope spaceScope) bool` instead of `(pkg, spaceID string, spaceAware bool)`.
- [x] 1.4 Remove or keep `resolveSpaceAware`/`supportsSpaceAwareIntegration` only as internals used by `resolveSpaceScope`; do not leave dead exported-equivalent duplicates.
- [x] 1.5 Add table-driven unit tests for `fleetPackageInstalled` spaceScope semantics (global vs strict target-space detection).

## 2. Fix the defect in `writeIntegration` and rework `waitForFleetIntegrationInstalled` (`internal/fleet/integration/create.go`)

- [x] 2.1 Change `waitForFleetIntegrationInstalled` signature to `(ctx, fleetClient, name, version string, scope spaceScope) error`; internally call `fleet.GetPackage(ctx, fleetClient, name, version, scope.id)` and `fleetPackageInstalled(pkg, scope)`.
- [x] 2.2 In `writeIntegration`, resolve `scope := resolveSpaceScope(ctx, client, planModel.SpaceID, &diags)` once, before the install call; return early on `diags.HasError()`.
- [x] 2.3 Replace the buggy call `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, "", false)` with `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, spaceScope{id: scope.id, aware: false})`. Scope the get-package API path to `scope.id` (fix for issue #4282) while keeping global install detection for this first wait.
- [x] 2.4 Replace `globallyInstalled := fleetPackageInstalled(pkg, "", false)` with `fleetPackageInstalled(pkg, spaceScope{})` (explicit default-space scope — unchanged semantics, no additional API call).
- [x] 2.5 Replace `installedInTargetSpace := fleetPackageInstalled(pkg, spaceID, true)` with `fleetPackageInstalled(pkg, spaceScope{id: scope.id, aware: true})`, preserving today's unconditional `aware: true` for this specific check.
- [x] 2.6 Update `installInSpace` to accept `scope.id` (or the full `scope`, whichever keeps the signature simplest) and pass it through to `waitForFleetIntegrationInstalled`; keep its internal `supportsSpaceAwareIntegration` re-check unchanged (it independently gates the cross-space Kibana-assets path).
- [x] 2.7 Confirm `installOptions.SpaceID` (used for the `InstallPackage` call) and `scope.id` agree — both derive from `planModel.SpaceID` — and simplify to a single source of truth if straightforward.

## 3. Thread `spaceScope` through read and delete for consistency (no behavior change)

- [x] 3.1 `internal/fleet/integration/read.go`: replace `spaceAware := resolveSpaceAware(ctx, client, model.SpaceID, &diags)` with `scope := resolveSpaceScope(ctx, client, model.SpaceID, &diags)`; update the `fleetPackageInstalled(pkg, spaceID, spaceAware)` call to `fleetPackageInstalled(pkg, scope)` (using `scope.id` in place of the existing `spaceID` parameter where they refer to the same value — confirm they're always equal, since `spaceID` is passed in separately from entitycore).
- [x] 3.2 `internal/fleet/integration/delete.go`: same substitution in `deleteIntegration`; keep `isInstalledInMultipleSpaces(pkg, spaceID)` and `deleteKibanaAssetsWithFallback(ctx, fleetClient, name, version, spaceID, force)` taking a plain `spaceID string` (extract from `scope.id`), since they don't need `aware`.
- [x] 3.3 Re-run a diff review to confirm no observable behavior changed in `read.go`/`delete.go` — this step is a pure refactor.

## 4. Acceptance test proving the fix (`internal/fleet/integration/issue_4282_acc_test.go`)

- [x] 4.1 Adapt the existing `TestAccReproduceIssue4282` (already on this branch from PR #4300), gated by `versionutils.SkipIfUnsupported(t, integration.MinVersionSpaceAwareIntegration, versionutils.FlavorAny)` — do not add a duplicate `TestAccResourceIntegration_SpaceRestrictedKey`.
- [x] 4.2 Reuse the existing `testdata/TestAccReproduceIssue4282/` fixture: `elasticstack_kibana_space` (random space id) + `elasticstack_kibana_security_role`/`elasticstack_elasticsearch_security_user` scoped only to that space + `elasticstack_fleet_integration` (`name = "tcp"`, `version = "1.16.0"`, `space_id = <space_id>`, restricted `kibana_connection`).
- [x] 4.3 Positive assertion: apply succeeds; `testAccCheckIntegrationInstalledInSpace(spaceID)` passes (admin client, tcp/1.16.0 fixture) and `testAccCheckFleetGetPackageTargetSpaceAllowed` confirms the restricted client sees the package installed in the target space (remove the pre-fix `ExpectError` 403 expectation).
- [x] 4.4 Negative guard: restricted-credentials checks that target-space `GetPackage` succeeds and default-space `GetPackage` returns HTTP 403 (not merely generic forbidden wording).
- [x] 4.5 Confirm the existing role privilege set (`fleet`/`fleetv2` all, scoped to the test space) works against a running 9.1+ stack — adjust only if the restricted user 403s on the legitimately-scoped path too. CI: run [30099588569](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/30099588569) job [89502403349](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/30099588569/job/89502403349) (9.1.10 shard 0) — `TestAccReproduceIssue4282` PASS (16.03s).

## 5. Verify

- [x] 5.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fleet-integration-space-scoped-install-poll --type change` and fix any reported issues.
- [x] 5.2 Run `make build`.
- [x] 5.3 Run `go vet ./internal/fleet/integration/...`.
- [x] 5.4 Bring up a 9.1+ Elastic Stack and run `make testacc TESTARGS='-run TestAccReproduceIssue4282'` (or the repo's docker-testacc equivalent); confirm it fails pre-fix (403 during post-install wait) and passes post-fix. CI: run [30099588569](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/30099588569) job [89502403349](https://github.com/elastic/terraform-provider-elasticstack/actions/runs/30099588569/job/89502403349) (9.1.10 shard 0) — `TestAccReproduceIssue4282` PASS (16.03s).
- [x] 5.5 Re-run `TestAccResourceIntegration_MultiSpaceInstall`, `TestAccResourceIntegration_MultiSpaceDelete`, and `TestAccResourceIntegration_SpaceAwareDrift` to confirm no regression from the `spaceScope` refactor. CI: same job — `TestAccResourceIntegration_MultiSpaceInstall` PASS (17.73s), `TestAccResourceIntegration_MultiSpaceDelete` PASS (18.06s), `TestAccResourceIntegration_SpaceAwareDrift` PASS (15.12s).
- [x] 5.6 Run `make check-lint` on touched files.
