## Context

`elasticstack_fleet_integration` (`internal/fleet/integration/`) exposes `force`, `prerelease`, and `ignore_constraints` as install-time behavior toggles, but never reads, sets, or forwards the sibling request field `allow_outdated_version` on the Kibana "Install a package from the registry" API. The request body schema `Kibana_HTTP_APIs_install_package_from_registry_request` defines `allow_outdated_version` (boolean, default `false`): "When true, allow installing a version older than the latest available. Bypasses only the out-of-date version check." The generated Go client already has `AllowOutdatedVersion *bool` (json `allow_outdated_version,omitempty`) on `KibanaHTTPAPIsInstallPackageFromRegistryRequest`, next to `Force` and `IgnoreConstraints` — this is a provider-side wiring gap only.

`fleet.InstallPackageOptions` (`internal/clients/fleet/packages.go`) currently has two different wiring conventions for its existing bool fields:

- `Force bool` and `IgnoreConstraints bool`: plain, non-pointer fields, always addressed (`&opts.Force`, `&opts.IgnoreConstraints`) into the request body, so the JSON key is always sent (`true` or `false`) regardless of whether the user configured the attribute.
- `IgnoreMappingUpdateErrors *bool` and `SkipDataStreamRollover *bool`: pointer fields, set from the plan only when `typeutils.IsKnown(...)` is true (`create.go`), so the JSON key is sent only when the user explicitly configured the attribute and omitted otherwise.

`InstallPackage` passes `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` straight through as params on the request (`PostFleetEpmPackagesPkgnamePkgversionParams`), while `Force`/`IgnoreConstraints` are set on the JSON body (`PostFleetEpmPackagesPkgnamePkgversionJSONRequestBody`, a type alias for `KibanaHTTPAPIsInstallPackageFromRegistryRequest`). `allow_outdated_version` is a body field on that same struct, alongside `Force`/`IgnoreConstraints` — but this design wires it using the pointer-if-known convention, not the always-send convention of its body siblings (see Decision 1).

## Goals / Non-Goals

**Goals**

- Add `allow_outdated_version` to the resource schema, model, `InstallPackageOptions`, and the `InstallPackage` request body, covering both install operations' shared body type.
- Wire the attribute so an unset configuration never sends the JSON key, on any server version — no compatibility risk for users who don't touch the new attribute.
- Cover both create and update, since `writeIntegration` already backs both.

**Non-Goals**

- Regenerating `generated/kbapi` — the field already exists in the generated client.
- Changing the always-send wiring of `force`/`ignore_constraints`, or the pointer-if-known wiring of `ignore_mapping_update_errors`/`skip_data_stream_rollover` — both existing conventions are left as-is.
- Any change to the bulk-install package endpoint/schema (`KibanaHTTPAPIsBulkInstallPackagesFromRegistryRequest`), which also already carries an unused `AllowOutdatedVersion` field — out of scope for this resource-scoped issue.
- Implementation itself — this change produces only the OpenSpec proposal/design/tasks/delta-spec artifacts.

## Decisions

### Decision 1 — Plain Optional bool attribute, wired pointer-if-known

Declare `allow_outdated_version` in the schema and model the same way as `force`/`ignore_constraints` (plain `Optional` `schema.BoolAttribute`, no `Computed`, no default) — this matches the resource's existing attribute-declaration convention for install-time toggles tied to `version`.

Wire it through `InstallPackageOptions`/`InstallPackage` using the pointer-if-known convention already precedented by `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` in the same struct, not the always-send convention of `Force`/`IgnoreConstraints`:

- `InstallPackageOptions.AllowOutdatedVersion *bool`.
- In `writeIntegration`, set it only when `typeutils.IsKnown(planModel.AllowOutdatedVersion)` is true: `installOptions.AllowOutdatedVersion = planModel.AllowOutdatedVersion.ValueBoolPointer()`.
- In `InstallPackage`, set `body.AllowOutdatedVersion = opts.AllowOutdatedVersion` directly (already a `*bool`; no address-of needed, unlike `Force`/`IgnoreConstraints`).

This combines two existing conventions rather than copying one sibling exactly (declared like `force`, wired like `ignore_mapping_update_errors`) — a small, deliberate deviation. The alternative, mirroring `force`/`ignore_constraints`'s always-send wiring, would send `allow_outdated_version: false` on every install even when the user never mentioned the attribute, which is unnecessary exposure to unknown-field handling on servers that predate this field. Pointer-if-known wiring makes unset configurations behave identically to today on every server version.

### Decision 2 — Version gate: Kibana >= 9.6.0

The field is not annotated with a minimum version in the OpenAPI spec, but a live Kibana 9.5.5 returned HTTP 400 (`Additional properties are not allowed ('allow_outdated_version' was unexpected)`) when it was sent. The upstream change (elastic/kibana PR #280316, `backport:skip`) targets 9.6.0. The resource therefore adds `MinVersionAllowOutdatedVersion = 9.6.0` (`resource.go`) and a `GetVersionRequirements` entry (`models.go`), following the existing `MinVersionIgnoreMappingUpdateErrors`/`MinVersionSkipDataStreamRollover` pattern.

The gate triggers on "was the attribute explicitly configured at all" (`typeutils.IsKnown` only), not on `IsKnown && ValueBool()`, because the pointer-if-known wiring sends the JSON key for an explicit `false` just as much as for `true`; an `IsKnown && ValueBool()` gate would let an explicit `false` through ungated on an old server and get a 400. Unset configurations are never gated.

### Decision 3 — No read-path change

`readIntegration` (`internal/fleet/integration/read.go`) only overwrites `Version`, `ID`, and `SpaceID`; it otherwise preserves prior-state values for install-only, Optional (non-Computed) attributes. This is the same mechanism that already keeps `force`/`ignore_constraints` stable across refreshes, so `allow_outdated_version` needs no explicit read-back mapping.

## Open questions

- **(Resolved)** Minimum server version: 9.6.0, inferred from the live 9.5.5 rejection and the upstream PR targeting 9.6.0. See Decision 2.
- Does `allow_outdated_version` apply to any uninstall/reinstall code path, or only to the install-from-registry request `writeIntegration` builds? Not separately investigated beyond confirming `writeIntegration` is shared by create and update; no uninstall-time use of `InstallPackageOptions` was found.
- Does `allow_outdated_version` interact with the resource's existing `prerelease` query parameter (e.g., does setting one implicitly affect version resolution for the other)? Not documented anywhere consulted (elastic-docs, OpenAPI spec); flagged as unresolved rather than asserted.

## Risks / Trade-offs

- **Mixed wiring conventions within `InstallPackageOptions`.** `allow_outdated_version` joins `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` on the pointer-if-known side while `Force`/`IgnoreConstraints` remain always-send. This is intentional (see Decision 1) and matches existing precedent in the same struct; it does mean a future reader of `InstallPackageOptions` must check each field's convention individually rather than assuming one struct-wide rule.
- **Floor inferred from 9.5.5 rejection.** The 9.6.0 floor comes from Kibana 9.5.5 rejecting the field and the upstream PR targeting 9.6.0; releases between 9.5.5 and 9.6.0 (e.g. 9.6 pre-releases) are unverified. Acceptable: the gate errs on the side of a clear provider diagnostic, and the field is opt-in.

## Test plan

- **Unit**: a request-capture test around `InstallPackage` (`internal/clients/fleet/packages.go`) with `InstallPackageOptions.AllowOutdatedVersion` unset (nil), explicitly `false`, and explicitly `true` — asserting the JSON key is absent in the nil case and present with the correct value otherwise. This directly tests the compatibility-critical omission behavior without depending on a live Kibana server's handling of unknown fields. A model test asserts the version requirement is produced for null/unknown (none) and for explicit `true`/`false` (9.6.0).
- **Acceptance**, extending `TestAccResourceIntegrationWithAllParameters` (`internal/fleet/integration/acc_test.go`):
  - A step with `allow_outdated_version` omitted — assert `resource.TestCheckNoResourceAttr`, mirroring the existing omitted-attribute assertions for `force`/`ignore_constraints`, and that install succeeds on the latest version as today.
  - Steps gated by `SkipFunc` on `MinVersionAllowOutdatedVersion`: `allow_outdated_version = true` on a pinned older version, then explicit `false`, then the attribute removed — asserting `true`, `false`, and `TestCheckNoResourceAttr` respectively. These verify the key is accepted by Kibana >= 9.6.0 and round-trips; they do not by themselves prove the flag changes install behavior.
