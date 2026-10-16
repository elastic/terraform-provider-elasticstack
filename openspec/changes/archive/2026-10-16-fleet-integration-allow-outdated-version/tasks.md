## 1. Schema

- [x] 1.1 Add `allow_outdated_version` to `integrationSchema()` in `internal/fleet/integration/schema.go` as a plain `schema.BoolAttribute{Description: ..., Optional: true}`, declared next to `force`/`prerelease`/`ignore_constraints`. Do not add `Computed` or a default — same shape as `force`/`ignore_constraints`.
- [x] 1.2 Write the description to state that, when true, installing a version older than the latest available bypasses the out-of-date-version check.

## 2. Model

- [x] 2.1 Add `AllowOutdatedVersion types.Bool `tfsdk:"allow_outdated_version"`` to `integrationModel` in `internal/fleet/integration/models.go`, next to `IgnoreConstraints`.

## 3. Client options

- [x] 3.1 Add `AllowOutdatedVersion *bool` to `fleet.InstallPackageOptions` in `internal/clients/fleet/packages.go`, next to `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` (pointer-if-known fields), not next to the always-send `Force`/`IgnoreConstraints` plain-bool fields.
- [x] 3.2 In `InstallPackage`, set `body.AllowOutdatedVersion = opts.AllowOutdatedVersion` on `PostFleetEpmPackagesPkgnamePkgversionJSONRequestBody` (already a `*bool`, no address-of needed) alongside the existing `Force`/`IgnoreConstraints` body fields.

## 4. Request mapping

- [x] 4.1 In `writeIntegration` (`internal/fleet/integration/create.go`), read `AllowOutdatedVersion` with the same `typeutils.IsKnown` guard used for `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover`: when known, set `installOptions.AllowOutdatedVersion = planModel.AllowOutdatedVersion.ValueBoolPointer()`; when null/unknown, leave it `nil` so the JSON key is omitted. Do not add this to the `InstallPackageOptions{...}` struct literal that sets `Force`/`Prerelease`/`IgnoreConstraints` unconditionally — add it as a separate guarded assignment, consistent with how `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` are set.
- [x] 4.2 No change needed to `updateIntegration` (`internal/fleet/integration/update.go`) — it delegates to `writeIntegration`, so task 4.1 covers update automatically.

## 4A. Version gate

- [x] 4A.1 Add `MinVersionAllowOutdatedVersion = 9.6.0` to `internal/fleet/integration/resource.go` and a `GetVersionRequirements` entry in `models.go` triggering on `typeutils.IsKnown(m.AllowOutdatedVersion)` only (explicit `false` is gated too).
- [x] 4A.2 Mention the 9.6.0 requirement in the attribute description.

## 5. Read path

- [x] 5.1 No change needed to `readIntegration` (`internal/fleet/integration/read.go`) — confirm (do not implement) that it does not overwrite `AllowOutdatedVersion`, consistent with its existing treatment of `force`/`ignore_constraints`.

## 6. Unit tests

- [x] 6.1 Add a request-capture test around `InstallPackage` (`internal/clients/fleet/packages.go`) asserting the `allow_outdated_version` JSON key is absent from the request body when `InstallPackageOptions.AllowOutdatedVersion` is `nil`, and present with the correct boolean value when explicitly `false` or `true`.
- [x] 6.2 Add a model test asserting `GetVersionRequirements` returns no requirement for null/unknown `allow_outdated_version` and a 9.6.0 requirement for explicit `true` and `false`.

## 7. Acceptance tests

- [x] 7.1 Extend `TestAccResourceIntegrationWithAllParameters` (`internal/fleet/integration/acc_test.go`) and its `testdata/TestAccResourceIntegrationWithAllParameters/` fixtures with a step where `allow_outdated_version` is omitted from config — assert `resource.TestCheckNoResourceAttr(..., "allow_outdated_version")` and that install/apply succeeds as today.
- [x] 7.2 Add steps gated by `SkipFunc` on `integration.MinVersionAllowOutdatedVersion`, pinning an older version with `allow_outdated_version = true` (fixture `all_params_step6`), asserting `"true"`. This verifies the key is accepted and round-trips; it does not prove the flag changes install behavior.
- [x] 7.3 Add a gated step with explicit `allow_outdated_version = false` (fixture `all_params_step7`) asserting `"false"`, then a gated step removing the attribute asserting `resource.TestCheckNoResourceAttr(..., "allow_outdated_version")` (post-apply plan is empty, confirming no drift).

## 8. Docs (follow-up, not required for this proposal)

- [x] 8.1 Add `allow_outdated_version` to `docs/resources/fleet_integration.md` and the resource's example.

## 9. Spec sync

- [x] 9.1 After implementation, update the schema block and REQ-011 of the main spec (`openspec/specs/fleet-integration/spec.md`) to list `allow_outdated_version` among the install options passed when configured, and add REQ-020/REQ-021/REQ-022 for the attribute itself, then sync the delta spec in `openspec/changes/fleet-integration-allow-outdated-version/specs/fleet-integration/spec.md` into the main spec (openspec-sync-specs skill) and archive the change (openspec-archive-change skill).
