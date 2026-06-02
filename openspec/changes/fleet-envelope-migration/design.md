## Context

The provider uses `entitycore.KibanaResource[T]` as the standard envelope for Kibana-backed resources. It owns Schema, Create, Read, Update, Delete, client resolution, version-requirement enforcement, and read-after-write. Resources only supply a model type + four callbacks.

Three Fleet resources (`fleet_server_host`, `fleet_output`, `fleet_custom_integration`) still use the older `*entitycore.ResourceBase` pattern, implementing CRUD as methods on the resource struct and manually injecting the `kibana_connection` schema block. This migration brings them in line with the rest of the provider.

Relevant reference: `internal/fleet/agentdownloadsource` and `internal/fleet/proxy` are already migrated and serve as the in-repo implementation template.

## Goals / Non-Goals

**Goals:**
- Replace `*entitycore.ResourceBase` with `*entitycore.KibanaResource[T]` in all three packages
- Add the four `KibanaResourceModel` interface methods to each model
- Remove manual `kibana_connection` block injection from each schema function
- Convert CRUD method bodies to package-level callback functions
- Add `entitycore_contract_test.go` to each package

**Non-Goals:**
- Changing the entitycore envelope itself
- Migrating any other Fleet or Kibana resources
- Changing user-visible schema, state format, or behaviour
- Addressing the `space_ids` semantics (preserved as-is)

## Decisions

### D1: `GetSpaceID()` for `space_ids`-plural resources

**serverhost** and **output** use `space_ids types.Set` (a Fleet membership list, 0–N spaces), while the envelope's `KibanaResourceModel` requires `GetSpaceID() types.String`.

**Decision**: `GetSpaceID()` returns the first element of `SpaceIDs`, or `""` if null/empty.

**Rationale**: The envelope loads the model from state before calling `readFunc`/`deleteFunc`, so `model.GetSpaceID()` on a state-loaded model equals what `GetOperationalSpaceFromState` computed from raw state. For Create, the model comes from the plan — again equivalent to `SpaceIDFromSet`. The envelope therefore replicates the existing space-routing logic without any envelope changes.

**Write callbacks and Update**: The Update callback receives `KibanaWriteRequest{Plan, Prior, ...}`. The callback uses `req.Prior.GetSpaceID()` as the operational space (where the resource currently exists), replacing the manual `GetOperationalSpaceFromState` call.

**Alternative considered**: Adding a new `KibanaFleetSpace` interface to the envelope. Rejected — the existing `GetSpaceID()` + `IsUnscopedSpace()` interfaces are sufficient.

### D2: `IsUnscopedSpace()` for Fleet resources

When `space_ids` is null/empty, `GetSpaceID()` returns `""`. The envelope rejects empty space IDs by default (non-Fleet Kibana resources require a space).

**Decision**: Both `serverhost` and `output` implement `KibanaUnscopedSpace` by returning `true` from `IsUnscopedSpace()`. `customintegration` uses `space_id` (singular string), so empty-string validation is still appropriate; it does NOT implement `KibanaUnscopedSpace`.

**Rationale**: `KibanaUnscopedSpace` already exists precisely for this pattern. Empty space = "default Fleet space" for Fleet resources.

### D3: `GetResourceID()` per resource

- `serverhost` → `HostID` (the API-assigned or user-provided host ID)
- `output` → `OutputID` (the API-assigned or user-provided output ID)
- `customintegration` → `types.StringValue(getPackageID(PackageName, PackageVersion))` when both are known, otherwise `ID`. The read callback ignores the `resourceID` parameter and uses `model.PackageName`/`model.PackageVersion` directly, so this value is only needed for the envelope's non-empty validation and read-after-write call.

### D4: Preserved wrapper-struct interfaces

The following interfaces are not owned by the envelope and stay on the outer resource struct unchanged:

| Resource | Preserved interfaces |
|---|---|
| serverhost | `ImportState` (via `*fleet.SpaceImporter`) |
| output | `ImportState` (via `*fleet.SpaceImporter`), `UpgradeState` |
| customintegration | `ModifyPlan` |

Go method promotion means these methods are visible on the wrapper struct without any changes.

### D5: serverhost Delete pre-condition

The Fleet API refuses to delete a server host marked `IsDefault=true`. The existing pre-delete clear-default logic stays inside the `deleteFunc` callback — it has access to the model and the Fleet client.

### D6: Full envelope callback migration — no placeholders

Three Kibana/Fleet resources currently embed `KibanaResource[T]` but pass `PlaceholderKibanaWriteCallback` for Create and/or Update, overriding those lifecycle methods directly on the wrapper struct (`agentdownloadsource`, `security_enable_rule`, `synthetics/privatelocation`).

**Decision**: The three resources in this change SHALL fully migrate to envelope callbacks. No `PlaceholderKibanaWriteCallback` usage. Create, Read, Update, and Delete are all supplied as `KibanaResourceOptions` callbacks.

**Rationale**: The historical reason for the placeholder pattern (Update needing the prior-state space via `GetOperationalSpaceFromState`) is solved by the envelope's `KibanaWriteRequest{Prior}` — Update callbacks read `req.Prior.GetSpaceID()` to recover the operational space. There is no remaining technical reason to bypass the envelope for these resources.

### D7: Version requirements via `GetVersionRequirements`

`customintegration` and `output` perform server-version gating via inline `EnforceMinVersion` calls inside CRUD methods and model helpers. The envelope already evaluates `WithVersionRequirements` (`GetVersionRequirements() ([]VersionRequirement, diag.Diagnostics)`) after client resolution and before callback dispatch.

**Decision**: Both `customintegration` and `output` models implement `WithVersionRequirements`. The inline `EnforceMinVersion` calls are removed.

| Resource | `GetVersionRequirements()` returns |
|---|---|
| serverhost | (does not implement; no version requirements) |
| customintegration | always: 8.2.0 (Fleet custom package endpoints) |
| output | conditional: 8.13.0 when `Type == "kafka"`; 8.10.0 when `ssl.verification_mode` is set |

**Consequence for `output`**: `assertKafkaSupport` and `assertSSLVerificationModeSupport` in `output/models.go` are deleted. `buildCommonNewOutput` and `buildCommonUpdateOutput` lose their `*clients.KibanaScopedClient` dependency for version checks (they may retain it for other reasons or drop it entirely).

**Consequence for `customintegration`**: `minVersionCustomPackageGet` checks in `create.go`, `read.go`, and `update.go` are deleted. The version requirement is enforced once by the envelope before each lifecycle call.

**Rationale**: Centralises version policy on the model where it is declarative, removes per-call boilerplate, and keeps behaviour identical (the envelope check fires at the same lifecycle point as the existing inline checks).

## Risks / Trade-offs

- **space_ids Set ordering** → `GetSpaceID()` uses the first element of the Set, which has no guaranteed order. This matches the existing `SpaceIDFromSet` / `GetOperationalSpaceFromState` behaviour (both also take the first element). No regression.
- **Read-after-write with empty space** → When `space_ids` is null (default space), the envelope calls `readFunc` with `spaceID=""`. The Fleet client already handles `""` as "default space" in all three resources. No regression.
- **Behaviour parity** → This is a pure structural refactor. The callbacks contain the same logic as the current method bodies. Risk is limited to transcription errors; existing acceptance tests catch these.
- **Version requirements firing point** → The envelope evaluates `GetVersionRequirements()` in Create, Update, and Read paths. Today, `customintegration` checks the min version in all three; `output` checks the version only when building Create/Update request bodies. The envelope is stricter (also fires in Read). In practice Read of a resource whose version is unsupported would have failed Create/Update first, so the stricter check has no behavioural impact on real workflows.

## Migration Plan

Each resource is independently migratable. Suggested order: `serverhost` → `output` → `customintegration` (increasing model complexity).

Per resource:
1. Add interface methods to model (`GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`; plus `IsUnscopedSpace` for serverhost/output)
2. Remove `kibana_connection` block from schema function
3. Extract CRUD method bodies into package-level callback functions
4. Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[T]` in resource struct and constructor
5. Add `entitycore_contract_test.go`
6. Run `make build` and existing acceptance tests

## Open Questions

_(none)_
