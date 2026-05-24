## Context

Both `fleet/proxy` and `fleet/agentdownloadsource` predate the `entitycore.KibanaResource` envelope and were hand-wired: they embed `*entitycore.ResourceBase` directly, implement all four CRUD operations as receiver methods on `Resource`, and each carry a standalone `assertVersionSupported()` function that is explicitly called in every operation. The `entitycore.KibanaResource` envelope (introduced later) handles exactly this pattern — schema injection, client resolution, version requirement dispatch, and CRUD delegation — for all comparable Kibana-backed resources.

The branch already converted `assertVersionSupported` from a receiver method to a package-level function in both packages as a preparatory step.

## Goals / Non-Goals

**Goals:**
- Replace `assertVersionSupported` with `WithVersionRequirements` on the model in both packages
- Wire both resources through `entitycore.NewKibanaResource` so the envelope owns shared CRUD behaviour
- Delete `version.go` in both packages
- Set both `source_id` and `id` from `SpaceImporter` on agentdownloadsource import

**Non-Goals:**
- Schema changes visible to practitioners
- Acceptance test changes (the external contract is identical)
- Changing any other Fleet resources

## Decisions

### Decision: `proxy` — full callback migration

All four CRUD operations become package-level callback functions matching the `KibanaCreateFunc / KibanaUpdateFunc / kibanaReadFunc / kibanaDeleteFunc` signatures. The concrete `Resource` struct embeds `*entitycore.KibanaResource[proxyModel]` and adds only `ImportState`.

**Why:** `proxyModel` maps cleanly to `KibanaResourceModel`. `GetID()` returns the composite `space/proxyID` — `resolveResourceIdentity` parses it on the primary path. `GetResourceID()` returns `ProxyID`, `GetSpaceID()` returns `SpaceID` (always non-empty; schema default is `"default"`). No space-identity complications.

### Decision: `agentdownloadsource` — partial callback migration (Read + Delete only)

`Create` and `Update` remain concrete receiver methods on `Resource` and override the envelope. `Read` and `Delete` become envelope callbacks. `PlaceholderKibanaWriteCallbacks` is passed for create/update.

**Why Create/Update stay manual:**

1. **Update needs space from STATE, not plan.** When `space_ids` changes (e.g. `["space-a"]` → `["space-b", "space-a"]`), the update must target the space where the resource currently exists (`space-a` from state), not the first entry of the new plan. The `KibanaResource.Update` envelope resolves identity from the plan model; there is no mechanism to substitute prior-state space. The update callback receives `prior T` but the `resourceID` and `spaceID` arguments are already resolved from plan.

2. **Create has a read-back-after-write pattern.** After `POST /api/fleet/agent_download_sources`, the resource reads back the created item to populate state fully. The `KibanaCreateFunc` contract returns the final model — the callback would need `client.GetFleetClient()` before making both calls, which is fine — but the space derivation uses `SpaceIDFromSet` on the plan's `space_ids` set, not `GetSpaceID()`, to preserve null-set semantics.

`Create` and `Update` call `entitycore.EnforceVersionRequirements(ctx, apiClient, &plan)` directly, which honours `WithVersionRequirements` on the model just as the envelope would.

### Decision: `GetSpaceID()` returns `"default"` for empty `space_ids`

`agentdownloadsource` uses `space_ids` (a `set(string)`) rather than a single `space_id` string. The `KibanaResourceModel` interface requires `GetSpaceID() types.String`. When `space_ids` is null or empty the correct API behaviour is to target the default Kibana space.

`BuildSpaceAwarePath` (in `internal/clients/kibanautil`) explicitly treats `""` and `"default"` identically — both leave the URL unchanged (no `/s/` prefix). Returning `"default"` therefore produces identical URLs and avoids the envelope's non-empty-space guard, without requiring `KibanaUnscopedSpace`.

```
GetSpaceID():
  if SpaceIDs null/unknown/empty → "default"
  else → first element of SpaceIDs
```

`types.Set.Elements()` is used (no context required), with a type-assertion to `types.String`.

**Alternatives considered:**
- Implement `KibanaUnscopedSpace.IsUnscopedSpace()` conditionally: works, but misrepresents the resource (it is space-scoped; empty just means the default space).
- Rename `KibanaUnscopedSpace` to a more general interface: correct semantics, but broader change with no other current need.

### Decision: `GetResourceID()` points to `SourceID`, not `ID`

For agentdownloadsource, `id` and `source_id` hold the same raw source ID (no composite). `SpaceImporter.ImportState` sets `source_id` but historically did not set `id`. `resolveResourceIdentity` tries composite parse on `GetID()` first; since the source ID has no `/`, parse fails and the fallback uses `GetResourceID()` = `SourceID`. `SourceID` is the field reliably populated after both normal read and import.

`SpaceImporter` is updated to set both fields (`path.Root("source_id")` and `path.Root("id")`), closing the state gap post-import. This doesn't change the identity resolution path (no composite) but makes state consistent immediately after import.

### Decision: Version enforcement dropped from Delete

The `KibanaResource` envelope does not call `EnforceVersionRequirements` in its Delete path. Enforcing a minimum server version on destroy is not operationally meaningful — if the stack was downgraded below the minimum between resource creation and deletion, the delete should still succeed. The `fleet-agent-download-source` spec is updated to narrow the version guard to Create, Read, and Update. `fleet-proxy` has no version requirement in its spec.

### Decision: Schema factory removes `kibana_connection` block

Both schema functions currently include `kibana_connection` in their `Blocks` map. The `KibanaResource.Schema` method injects this block automatically. The block must be removed from the factory function passed to `NewKibanaResource` to avoid duplicate-block panics.

## Risks / Trade-offs

- **State compatibility** — No schema or state shape changes; existing state files are fully compatible.
- **`GetSpaceID()` first-element non-determinism** — `types.Set` iteration order is deterministic within a single Terraform run but not guaranteed stable across runs. For the use cases (Read, Delete, Update-space resolution) the operational space must be consistent; since `space_ids` is `UseStateForUnknown` and the API reflects a single canonical space per resource, the first element reliably identifies the target space in practice.
- **Dropping version check on Delete** → No mitigation required; see decision rationale above.

## Open Questions

None.
