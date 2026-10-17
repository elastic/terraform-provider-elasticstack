## Why

`elasticstack_fleet_integration` never reads, sets, or forwards the Kibana install-package request field `allow_outdated_version` on either install operation (`POST /api/fleet/epm/packages/{pkgName}` and `POST /api/fleet/epm/packages/{pkgName}/{pkgVersion}`). The field (boolean, default `false`) lets a caller bypass only the out-of-date-version check and install a version older than the latest available. The resource already exposes the sibling install-time behavior toggles `force`, `prerelease`, and `ignore_constraints` for its required `version` attribute; `allow_outdated_version` is the same kind of user-controllable install-behavior toggle and is the one missing from `fleet.InstallPackageOptions`, as flagged by the Kibana API gap scanner in [#5318](https://github.com/elastic/terraform-provider-elasticstack/issues/5318).

The generated Kibana client already carries this field — `KibanaHTTPAPIsInstallPackageFromRegistryRequest.AllowOutdatedVersion *bool` (`generated/kbapi/kibana.gen.go`), which backs the JSON body for both install operations — so this is a provider-side wiring gap only; no `kbapi` regeneration is required.

This closes the gap by adding `allow_outdated_version` to the resource schema and wiring it through the shared create/update path.

## What Changes

- Add `allow_outdated_version` to `integrationSchema` (`internal/fleet/integration/schema.go`) as a plain `Optional` `schema.BoolAttribute`, declared the same way as `force`/`ignore_constraints` (no `Computed`, no default).
- Add a matching `AllowOutdatedVersion types.Bool` field to `integrationModel` (`internal/fleet/integration/models.go`).
- Add `AllowOutdatedVersion *bool` to `fleet.InstallPackageOptions` (`internal/clients/fleet/packages.go`) and set it on the install request body in `InstallPackage`, following the **pointer-if-known** convention already used in the same struct for `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` — not the always-send convention used for `Force`/`IgnoreConstraints`. The JSON key is included only when the attribute is explicitly configured (`true` or `false`); it is omitted entirely when the attribute is left unset.
- In `writeIntegration` (`internal/fleet/integration/create.go`), read `AllowOutdatedVersion` with the same `typeutils.IsKnown` guard already used for `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover`, and pass `planModel.AllowOutdatedVersion.ValueBoolPointer()` into `InstallPackageOptions`. Because `writeIntegration` backs both `createIntegration` and `updateIntegration` (`internal/fleet/integration/update.go`), this single change covers create and update. No change to `readIntegration` is needed — it does not touch install-only, Optional (non-Computed) attributes, the same mechanism that already keeps `force`/`ignore_constraints` stable across refreshes.
- No new version gate: because the wiring never sends the field unless the user explicitly sets it, existing configurations are unaffected on any server version regardless of whether that server recognizes `allow_outdated_version`. The minimum supported Kibana/Fleet version for this field is an open, non-blocking question (see design.md) that only matters for users who explicitly configure the attribute, and does not block this change.

## Capabilities

### Modified Capabilities

- `fleet-integration`: Adds the Optional `allow_outdated_version` bool attribute to `elasticstack_fleet_integration`, wired pointer-if-known into the shared create/update install-options path, with no behavior change for configurations that never set it.

## Impact

- **Changed code**: `internal/fleet/integration/schema.go` (new attribute), `internal/fleet/integration/models.go` (new model field), `internal/fleet/integration/create.go` (`writeIntegration` install-options wiring), `internal/clients/fleet/packages.go` (`InstallPackageOptions.AllowOutdatedVersion`, `InstallPackage` request-body wiring).
- **New/updated acceptance tests**: `internal/fleet/integration/acc_test.go` and its `testdata/`, extending `TestAccResourceIntegrationWithAllParameters` with steps covering the attribute omitted, explicitly `false`, and explicitly `true` against an older package version.
- **New/updated unit tests**: a request-capture test around `InstallPackage` (`internal/clients/fleet/packages.go`) asserting the `allow_outdated_version` JSON key is absent when unset and present with the correct value when explicitly configured true/false.
- **Docs**: `docs/resources/fleet_integration.md` and its example will need the new attribute (not produced in this proposal-only change).
- **No `generated/kbapi` changes**: `AllowOutdatedVersion *bool` already exists on `KibanaHTTPAPIsInstallPackageFromRegistryRequest`; no regeneration is required.
- **Compatibility**: Purely additive and omission-by-default. The new attribute is `Optional` (not `Computed`), and the pointer-if-known wiring sends no `allow_outdated_version` key at all for configurations that never set it, so there is no behavior change and no unknown-field risk for existing users on any server version.
