## Context

`elasticstack_fleet_integration`'s create/update flow (`writeIntegration` in `internal/fleet/integration/create.go`) installs a Fleet package and then polls until it reports installed:

```
writeIntegration
 ├─ 1. InstallPackage(..., installOptions.SpaceID)             scoped to space_id
 ├─ 2. waitForFleetIntegrationInstalled(..., "", false)         BUG: hard-coded default space
 ├─ 3. GetPackage(..., spaceID)                                 scoped (already correct)
 ├─ 4. installInSpace(...) → InstallKibanaAssets(..., spaceID)  scoped, 9.1.0+ only
 │        └─ waitForFleetIntegrationInstalled(..., spaceID, true)  scoped (already correct)
 └─ 5. read-after-write → readIntegration(..., readSpaceID)     scoped (entitycore)
```

Step 2 is the sole remaining defect: `waitForFleetIntegrationInstalled` calls `fleet.GetPackage(ctx, fleetClient, name, version, spaceID)`, and with `spaceID = ""` this resolves to the unscoped `/api/fleet/epm/packages/{name}/{version}` path (`kibanautil.BuildSpaceAwarePath` only inserts `/s/{space}` for a non-empty, non-"default" ID). A caller whose Elastic API key is scoped to a single custom space has no default-space read access, so every poll returns 403 and the resource create fails — even though `InstallPackage` in step 1 targeted the right space and would have succeeded.

Space context is currently threaded through the package as two loosely-coupled primitives: a `spaceID string` (where `""` is a magic sentinel for "default space") and a separate `spaceAware bool` (meaning "the server supports per-space install-status tracking", gated on `MinVersionSpaceAwareIntegration` = 9.1.0). Scattered `if spaceID == ""` / `spaceID != ""` checks appear in `space_helpers.go`, `create.go`, and `delete.go`. This is exactly what allowed the bug: it is possible, and in this case actually happened, for a caller to pass `"", false` and silently target the default space instead of propagating the real value.

## Goals / Non-Goals

**Goals:**

- Fix the 403: the post-install poll in `writeIntegration` must use the same space context (`space_id`, space-aware capability) as the install call that precedes it.
- Remove the structural cause, not just the symptom: replace the `spaceID string` + `spaceAware bool` pair with one resolved value type inside the `integration` package, so a future call site cannot express "default space" by accident.
- Prove the fix with an acceptance test that uses a genuinely space-restricted Elastic API key (not merely a differently-scoped superuser key), so an over-broad test credential cannot mask the bug.

**Non-Goals:**

- Changing `MinVersionSpaceAwareIntegration` semantics or gating. It continues to gate exactly what it gates today: cross-space Kibana-assets install/read via `installInSpace`.
- Pushing space state into the shared `fleet.Client`. That client is provider-level and shared across many resources and spaces; the resolved space scope stays local to the `integration` package, resolved fresh per CRUD entry point.
- Changing behavior of `readIntegration` or `deleteIntegration`. Both already pass the correct `spaceID` to their API calls; they are only refactored to use the new type for consistency.
- Any schema or public API changes.

## Deviation from the automated research comment

The issue's automated implementation-research comment recommends **Approach 1** ("minimal-unconditional-spaceaware-fix"): change the one call site to `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, waitSpaceID, waitSpaceID != "")` and stop there.

The issue's human comment history (two comments from `@tobio`, the triggering/trusted actor, posted after the research comment) explicitly overrides that recommendation: *"Direction: Option B — consolidate space handling into a single resolved `spaceScope`, which structurally fixes the `create.go:80` bug and removes the `""`-sentinel / `spaceAware bool` foot-gun."* This is a direct, explicit contradiction of the research comment's recommendation, not merely an adjacent discussion, so per the exclusive-scope-with-explicit-contradiction rule this proposal follows the human direction (Option B) rather than the research comment's Approach 1.

This proposal and its tasks are scoped to Option B: introduce `spaceScope` and thread it through create/read/delete in `internal/fleet/integration/`.

## Decisions

### Decision: Introduce a resolved `spaceScope` type

Add to `space_helpers.go`:

```go
// spaceScope is the resolved space context for one CRUD operation.
type spaceScope struct {
    id    string // "" == default space
    aware bool   // server supports per-space Kibana asset tracking (>= 9.1.0)
}

func resolveSpaceScope(ctx context.Context, client clients.MinVersionEnforceable, spaceID types.String, diags *diag.Diagnostics) spaceScope
```

`resolveSpaceScope` sets `id` to `spaceID.ValueString()` when known (else `""`), and `aware` to `false` when `id == ""`, otherwise to the result of `client.EnforceMinVersion(ctx, MinVersionSpaceAwareIntegration)` (appending any diagnostics to `diags`). This is the same logic `resolveSpaceAware`/`supportsSpaceAwareIntegration` implement today, just returned as one value instead of two.

`fleetPackageInstalled` changes from `(pkg, spaceID string, spaceAware bool)` to `(pkg, scope spaceScope)`. `waitForFleetIntegrationInstalled` changes from `(ctx, fleetClient, name, version, spaceID string, spaceAware bool)` to `(ctx, fleetClient, name, version string, scope spaceScope)`, and internally calls `fleet.GetPackage(ctx, fleetClient, name, version, scope.id)`.

**Alternatives considered** (superseding the research comment's Approach 1 vs. Approach 2 framing, per the human direction above):

- *Approach 1 (line fix only)*: `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, spaceID, spaceID != "")` at the one call site. Smallest diff, fixes the reported defect, but leaves the sentinel/bool pair — and therefore the foot-gun that produced this bug — in place for the next call added to the flow. Rejected per explicit human direction (Option B) in the issue comments.
- *Approach 2 (version-gated fix reusing `resolveSpaceAware` as-is)*: closer to today's convention in `read.go`/`delete.go` but keeps the two-primitive shape; does not address the structural issue. Also superseded by the Option B direction.

### Decision: Fix `writeIntegration` using the resolved scope

Resolve `scope := resolveSpaceScope(ctx, client, planModel.SpaceID, &diags)` once near the top of `writeIntegration`, before the install call. Replace the buggy `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, "", false)` with `waitForFleetIntegrationInstalled(ctx, fleetClient, name, version, scope)`. This makes the previous `"", false` construction impossible to accidentally reintroduce at that call site.

The subsequent `globallyInstalled := fleetPackageInstalled(pkg, "", false)` check (create.go:96) is intentionally a *different* scope — it is asking "is the package installed anywhere at all", not "is it installed in the target space" — so it becomes an explicit `fleetPackageInstalled(pkg, spaceScope{})` (the zero value, `id: "", aware: false`) rather than being folded into `scope`. The following `installedInTargetSpace := fleetPackageInstalled(pkg, spaceScope{id: scope.id, aware: true})` line preserves today's unconditional `spaceAware = true` for this specific check (unchanged behavior, matching create.go:97 today), keeping it decoupled from whether the server actually supports space-aware tracking. `installInSpace` receives `scope.id` and continues to call `supportsSpaceAwareIntegration` internally exactly as it does today (unchanged), since that call already correctly gates on the version check for the cross-space Kibana-assets path.

### Decision: Thread `spaceScope` through `read.go` and `delete.go` for consistency, no behavior change

`readIntegration` and `deleteIntegration` already call `resolveSpaceAware(ctx, client, model.SpaceID, &diags)` and separately receive `spaceID string` as a parameter (from the entitycore read/delete request). Both are updated to build a `spaceScope` via `resolveSpaceScope` and pass it to `fleetPackageInstalled`. `isInstalledInMultipleSpaces` (`delete.go`) and `deleteKibanaAssetsWithFallback` continue to take a plain `spaceID string` (they don't need `aware`), extracted from `scope.id`. This removes the duplicated sentinel/bool convention across the package without changing observable behavior in either function.

### Decision: Acceptance test with a genuinely space-restricted API key

Add `TestAccResourceIntegration_SpaceRestrictedKey` in `acc_test.go`, gated by `versionutils.SkipIfUnsupported(t, integration.MinVersionSpaceAwareIntegration, versionutils.FlavorAny)` (9.1.0+, consistent with the other space-aware tests in the file). The test config (new `testdata/TestAccResourceIntegration_SpaceRestrictedKey/` directory) builds, entirely in Terraform:

1. An `elasticstack_kibana_space` with a random space ID.
2. An `elasticstack_elasticsearch_security_api_key` whose `role_descriptors` embeds a Kibana application privilege scoped to only that space:
   ```json
   {
     "fleet_space_only": {
       "applications": [{
         "application": "kibana-.kibana",
         "privileges": ["feature_fleetv2.all", "feature_fleet.all"],
         "resources": ["space:${space_id}"]
       }]
     }
   }
   ```
   The exact privilege set needs empirical confirmation (see Open Questions / Risks) — if too narrow the key would 403 on the scoped path too, which would be a test bug, not a provider bug.
3. An `elasticstack_fleet_integration` with `space_id = <space_id>` and `kibana_connection { api_key = <encoded key> }` using the restricted key, installing a small fast package (`tcp`/`1.16.0`, matching the existing space tests).

**Positive assertion:** the integration installs successfully (pre-fix, this step 403s during the post-install wait); `testAccCheckIntegrationInstalledInSpace("tcp", "1.16.0", spaceID)` passes.

**Negative guard:** a `TestCheckFunc` builds a Fleet client from the encoded restricted key and calls `GetPackage(..., "")` (default space, empty space ID), asserting the response is 403/forbidden rather than success. This proves the key genuinely lacks default-space access, so an accidentally over-broad key cannot make the positive assertion pass without actually exercising the fix.

Cleanup is handled by Terraform destroy at the end of the test case (space + key + integration are all Terraform-managed).

## Open Questions

Copied verbatim from the issue's automated implementation-research comment:

- Does `GET /s/{space_id}/api/fleet/epm/packages/{name}/{version}` on pre-9.1 Kibana populate `InstalledKibanaSpaceId` for a plain single-space install? If not, Approach 1 (and existing lines 89-98) could time out instead of succeeding on old Kibana — would favor a version-aware fallback to the `Status`/`InstallStatus` check.
- Should the poll's space check and the lines 89-98 `globallyInstalled`/`installedInTargetSpace` check share one helper to avoid duplication?
- Is there an existing space-scoped-API-key test fixture, or does this need a new unit/acceptance test to reproduce the 403?

Note: this proposal's Option B design directly answers the second question above (yes — both now go through `fleetPackageInstalled` taking a `spaceScope`, with distinct scope values passed explicitly for the "globally installed" vs. "installed in target space" checks) and the third (yes — see the new `TestAccResourceIntegration_SpaceRestrictedKey` acceptance test above). The first question about pre-9.1 `GetPackage` response shape remains open; it is a pre-existing behavior of the code at create.go:96-97 today (unconditional `spaceAware=true` check for `installedInTargetSpace`), not something this change introduces, so it is tracked here rather than blocking this change.

## Risks / Trade-offs

- **Risk: pre-9.1 `GetPackage` response shape for `installedInTargetSpace`.** If a pre-9.1 Kibana's `GetPackage` response doesn't populate `InstalledKibanaSpaceId`/`AdditionalSpacesInstalledKibana` for a plain single-space install, the unconditional `aware: true` check at create.go:97 (preserved as-is in this refactor) could report "not installed in target space" incorrectly. This is a pre-existing risk, not one introduced by this change; not resolving it here.
- **Risk: privilege set for the space-restricted test key.** The Fleet EPM privileges needed for the space-restricted key must be confirmed empirically against a running 9.1+ stack; too narrow and the key 403s on the legitimately-scoped path too (a test bug), too broad and the negative guard could pass for the wrong reason. Mitigated by the explicit negative-guard assertion.
- **Trade-off: `read.go`/`delete.go` changes are refactor-only.** They carry no behavior change but do touch two files beyond the minimal one-line fix, increasing review surface. Justified by removing the sentinel/bool foot-gun package-wide, consistent with the explicit human direction to pursue Option B.
