## Context

`elasticstack_fleet_integration` (`internal/fleet/integration/`) exposes `force`, `prerelease`, and `ignore_constraints` as install-time behavior toggles, but never reads, sets, or forwards the sibling request field `allow_outdated_version` on the Kibana "Install a package from the registry" API. The request body schema `Kibana_HTTP_APIs_install_package_from_registry_request` defines `allow_outdated_version` (boolean, default `false`): "When true, allow installing a version older than the latest available. Bypasses only the out-of-date version check." The generated Go client already has `AllowOutdatedVersion *bool` (json `allow_outdated_version,omitempty`) on `KibanaHTTPAPIsInstallPackageFromRegistryRequest`, next to `Force` and `IgnoreConstraints` — this is a provider-side wiring gap only.

`fleet.InstallPackageOptions` (`internal/clients/fleet/packages.go`) currently has two different wiring conventions for its existing bool fields:

- `Force bool` and `IgnoreConstraints bool`: plain, non-pointer fields, always addressed (`&opts.Force`, `&opts.IgnoreConstraints`) into the request body, so the JSON key is always sent (`true` or `false`) regardless of whether the user configured the attribute.
- `IgnoreMappingUpdateErrors *bool` and `SkipDataStreamRollover *bool`: pointer fields, set from the plan only when `typeutils.IsKnown(...)` is true (`create.go`), so the JSON key is sent only when the user explicitly configured the attribute and omitted otherwise.

`InstallPackage` passes `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` straight through as params on the request (`PostFleetEpmPackagesPkgnamePkgversionParams`), while `Force`/`IgnoreConstraints` are set on the JSON body (`PostFleetEpmPackagesPkgnamePkgversionJSONRequestBody`, a type alias for `KibanaHTTPAPIsInstallPackageFromRegistryRequest`). `allow_outdated_version` is a body field on that same struct, alongside `Force`/`IgnoreConstraints` — but this design wires it using the pointer-if-known convention, not the always-send convention of its body siblings (see Decision 1).

## Goals / Non-Goals

**Goals**

- Add `allow_outdated_version` to the resource schema, model, `InstallPackageOptions`, and the `InstallPackage` request body, covering both install operations' shared body type.
- Wire the attribute so an unset configuration never sends the JSON key, regardless of server version — no compatibility risk for users who don't touch the new attribute.
- Cover both create and update, since `writeIntegration` already backs both.

**Non-Goals**

- Regenerating `generated/kbapi` — the field already exists in the generated client.
- Adding a `MinVersion*` compatibility gate for `allow_outdated_version` in this change. The minimum supported server version is an open question (see below); because the chosen wiring never sends the field unless explicitly configured, this is non-blocking for v1.
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

This combines two existing conventions rather than copying one sibling exactly (declared like `force`, wired like `ignore_mapping_update_errors`) — a small, deliberate deviation. The alternative, mirroring `force`/`ignore_constraints`'s always-send wiring, would send `allow_outdated_version: false` on every install even when the user never mentioned the attribute, which is unnecessary exposure to unknown-field handling on servers that predate this field. Pointer-if-known wiring makes unset configurations behave identically to today on every server version, so there is no first-apply compatibility risk to reason about.

### Decision 2 — No version gate in this change

The minimum Kibana/Fleet version that supports `allow_outdated_version` is not documented in the OpenAPI spec (no `x-state`/"Added in ..." annotation, unlike some other fields in the same spec) and is absent from the public Elastic docs corpus. Because the wiring never sends the field unless explicitly configured, unset configurations are unaffected regardless of server version — this makes the open question non-blocking for this change. Implementation SHOULD still verify against a real Kibana instance at the provider's minimum-supported Stack version whether an explicit `true`/`false` is accepted before documenting the attribute as unconditionally available; if rejected, a `MinVersion` entry should be added to `integrationModel.GetVersionRequirements` (`internal/fleet/integration/models.go`) following the existing `MinVersionIgnoreMappingUpdateErrors`/`MinVersionSkipDataStreamRollover` pattern in `resource.go`. Any such future gate must trigger on "was the attribute explicitly configured at all" (an `IsKnown`-only check), not on `IsKnown && ValueBool()`, because the pointer-if-known wiring sends the JSON key for an explicit `false` just as much as for `true` — a copy-pasted `IsKnown && ValueBool()` gate would silently let an explicit `false` through ungated on an old server while gating `true`.

### Decision 3 — No read-path change

`readIntegration` (`internal/fleet/integration/read.go`) only overwrites `Version`, `ID`, and `SpaceID`; it otherwise preserves prior-state values for install-only, Optional (non-Computed) attributes. This is the same mechanism that already keeps `force`/`ignore_constraints` stable across refreshes, so `allow_outdated_version` needs no explicit read-back mapping.

## Open questions

- **(Non-blocking — implementation-time verification, does not change the recommended approach)** What is the minimum Kibana/Fleet server version that supports `allow_outdated_version` in this request body? The OpenAPI spec carries no "Added in ..." annotation for this field, and it is undocumented in the public Elastic docs corpus. With pointer-if-known wiring, this is lower urgency than it would be with always-send wiring — unset configs are unaffected regardless of server version. Before documenting `allow_outdated_version` as unconditionally available, confirm against a real Kibana instance at the provider's minimum-supported Stack version whether an explicit `true`/`false` is accepted; if rejected, add a `MinVersion` entry per Decision 2 above.
- Does `allow_outdated_version` apply to any uninstall/reinstall code path, or only to the install-from-registry request `writeIntegration` builds? Not separately investigated beyond confirming `writeIntegration` is shared by create and update; no uninstall-time use of `InstallPackageOptions` was found.
- Does `allow_outdated_version` interact with the resource's existing `prerelease` query parameter (e.g., does setting one implicitly affect version resolution for the other)? Not documented anywhere consulted (elastic-docs, OpenAPI spec); flagged as unresolved rather than asserted.

## Risks / Trade-offs

- **Mixed wiring conventions within `InstallPackageOptions`.** `allow_outdated_version` joins `IgnoreMappingUpdateErrors`/`SkipDataStreamRollover` on the pointer-if-known side while `Force`/`IgnoreConstraints` remain always-send. This is intentional (see Decision 1) and matches existing precedent in the same struct; it does mean a future reader of `InstallPackageOptions` must check each field's convention individually rather than assuming one struct-wide rule.
- **Version floor left unresolved.** Shipping without a `MinVersion*` gate means an explicit `allow_outdated_version` configured against a server that predates the field will surface whatever error Kibana itself returns for an unrecognized field, rather than a provider-level "Unsupported parameter for server version" diagnostic. Acceptable because the field is opt-in and the risk is isolated to users who explicitly set it.

## Test plan

- **Unit**: a request-capture test around `InstallPackage` (`internal/clients/fleet/packages.go`) with `InstallPackageOptions.AllowOutdatedVersion` unset (nil), explicitly `false`, and explicitly `true` — asserting the JSON key is absent in the nil case and present with the correct value otherwise. This directly tests the compatibility-critical omission behavior without depending on a live Kibana server's handling of unknown fields.
- **Acceptance**, extending `TestAccResourceIntegrationWithAllParameters` (`internal/fleet/integration/acc_test.go`):
  - A step with `allow_outdated_version` omitted — assert `resource.TestCheckNoResourceAttr`, mirroring the existing omitted-attribute assertions for `force`/`ignore_constraints`, and that install succeeds on the latest version as today.
  - A step installing an older, non-latest package version with `allow_outdated_version = true` (and `force`/`ignore_constraints` held `false`/omitted) — assert `resource.TestCheckResourceAttr(..., "true")` and that the install succeeds, isolating this attribute's effect from its siblings.
  - A plan-only/refresh step confirming no drift when the attribute is left unset across refresh.
