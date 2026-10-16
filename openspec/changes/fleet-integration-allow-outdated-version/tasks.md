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

## 5. Read path

- [x] 5.1 No change needed to `readIntegration` (`internal/fleet/integration/read.go`) — confirm (do not implement) that it does not overwrite `AllowOutdatedVersion`, consistent with its existing treatment of `force`/`ignore_constraints`.

## 6. Unit tests

- [x] 6.1 Add a request-capture test around `InstallPackage` (`internal/clients/fleet/packages.go`) asserting the `allow_outdated_version` JSON key is absent from the request body when `InstallPackageOptions.AllowOutdatedVersion` is `nil`, and present with the correct boolean value when explicitly `false` or `true`.

## 7. Acceptance tests

- [x] 7.1 Extend `TestAccResourceIntegrationWithAllParameters` (`internal/fleet/integration/acc_test.go`) and its `testdata/TestAccResourceIntegrationWithAllParameters/` fixtures with a step where `allow_outdated_version` is omitted from config — assert `resource.TestCheckNoResourceAttr(..., "allow_outdated_version")` and that install/apply succeeds as today.
- [x] 7.2 Add a step (new testdata fixture pinning an older, non-latest version of a stable test package) with `allow_outdated_version = true` and `force`/`ignore_constraints` held `false` or omitted — assert `resource.TestCheckResourceAttr(..., "allow_outdated_version", "true")` and that the install succeeds, isolating this attribute's effect from its siblings.
- [x] 7.3 Add a plan-only/refresh check confirming no drift when `allow_outdated_version` is left unset across refresh.

## 8. Docs (follow-up, not required for this proposal)

- [x] 8.1 Add `allow_outdated_version` to `docs/resources/fleet_integration.md` and the resource's example.

## 9. Spec sync

- [x] 9.1 After implementation, update the schema block and REQ-011 of the main spec (`openspec/specs/fleet-integration/spec.md`) to list `allow_outdated_version` among the install options passed when configured, and add REQ-020/REQ-021 for the attribute itself, then sync the delta spec in `openspec/changes/fleet-integration-allow-outdated-version/specs/fleet-integration/spec.md` into the main spec (openspec-sync-specs skill) and archive the change (openspec-archive-change skill).
