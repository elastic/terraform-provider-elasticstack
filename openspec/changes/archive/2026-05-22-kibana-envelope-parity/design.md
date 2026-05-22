## Context

The Elasticsearch resource envelope (`entitycore.ElasticsearchResource[T]`) enforces read-after-write and provides a unified `WriteFunc[T]` type with a structured `WriteRequest[T]` (carrying `Plan`, `Prior *T`, `Config`, and `WriteID`). The Kibana envelope predates these features and still uses separate `KibanaCreateFunc[T]` / `KibanaUpdateFunc[T]` with positional arguments. Several Kibana resources compensate by doing their own ad-hoc reads inside callbacks.

This design brings the Kibana envelope to the same contract as the Elasticsearch envelope, with the Kibana-specific addition of `SpaceID` in the write request.

## Goals / Non-Goals

**Goals:**
- Enforce read-after-write in the Kibana envelope for all Create and Update paths
- Unify the write callback type to `KibanaWriteFunc[T]` (single type for both Create and Update)
- Pass decoded `Config T` to write callbacks for write-only attribute support
- Add optional `PostRead KibanaPostReadFunc[T]` hook
- Introduce `KibanaResourceOptions[T]` constructor options struct
- Migrate all 6 existing concrete Kibana resources to the new API in a single change
- Simplify `maintenance_window` by removing the manual read-after-write it currently does inside its create callback

**Non-Goals:**
- `UpdateNotSupportedKibanaWriteCallback` helper — not needed yet
- `WithKibanaReadIdentity` override interface — no current resource needs it; can be added later
- Changes to the `KibanaResourceModel` interface itself
- Changes to `kibanaReadFunc[T]` or `kibanaDeleteFunc[T]` signatures

## Decisions

### D1: Reuse `runWrite` pattern, add Kibana-specific `runKibanaWrite`

The ES envelope centralises all Create/Update logic in `runWrite`. We introduce an analogous `runKibanaWrite` on `KibanaResource[T]`. `Create` and `Update` delegate to it, passing a `kibanaWriteInvocation[T]` struct (matching `writeInvocation[T]` in ES).

**Alternative considered:** Shared generic helper across both envelopes. Rejected — the Kibana envelope has distinct concerns (spaceID, KibanaScopedClient) that would require awkward parameterisation of a shared function.

### D2: Read identity from written model via `GetResourceID()` + `GetSpaceID()`

After a write, the envelope resolves the read identity as:
- `readResourceID = writtenModel.GetResourceID().ValueString()`
- `readSpaceID = writtenModel.GetSpaceID().ValueString()`

This matches how the ES envelope uses `resolveElasticsearchReadResourceID(written.Model, writeKey)` — using the written model as the source of truth. Write callbacks that deal with API-assigned IDs (e.g., `maintenance_window`) must set the ID field on the model before returning; they already do this.

**Alternative considered:** `WithKibanaReadIdentity` interface for overriding read identity. Deferred — no current resource needs it. When needed, it follows the same pattern as `WithReadResourceID` in the ES envelope.

### D3: `KibanaWriteRequest[T]` includes `SpaceID` field

Unlike the ES `WriteRequest[T]` (which has no space concept), the Kibana equivalent carries `SpaceID string`. This lets callbacks access the space without re-deriving it from the plan, and makes the envelope's spaceID validation transparent.

```go
type KibanaWriteRequest[T KibanaResourceModel] struct {
    Plan    T
    Prior   *T      // nil for Create, non-nil for Update
    Config  T
    WriteID string  // GetResourceID() from plan
    SpaceID string  // GetSpaceID() from plan
}
```

### D4: `PlaceholderKibanaWriteCallbacks` returns a single `KibanaWriteFunc[T]`

Currently returns a pair. After this change, returns one function (matching `PlaceholderElasticsearchWriteCallback`). The only caller (`fleet/agentdownloadsource`) decomposes the pair — the call site becomes:

```go
// before
createFn, updateFn := entitycore.PlaceholderKibanaWriteCallbacks[model]()

// after
placeholder := entitycore.PlaceholderKibanaWriteCallback[model]()
// used in opts.Create and opts.Update
```

**Rename:** `PlaceholderKibanaWriteCallbacks` → `PlaceholderKibanaWriteCallback` (singular, matching the ES naming).

### D5: spaceID validation remains in `runKibanaWrite`

The current `Create` validates `spaceID` (non-empty, non-unknown, or `KibanaUnscopedSpace` bypass). This logic moves into `runKibanaWrite` so both Create and Update paths are covered. The `Update` path currently lacks spaceID validation — this is now enforced uniformly.

### D6: `PostRead` runs after every successful read including read-after-write

Same semantics as the ES envelope: invoked after `resp.State.Set` succeeds, skipped when not-found, on readFunc errors, or state-set errors. This covers both the standalone `Read` path and the embedded read in `runKibanaWrite`.

## Risks / Trade-offs

**[Risk] maintenance_window create simplification may miss edge cases** → The current `createMaintenanceWindow` fetches the full server state before returning. After simplification, the envelope calls `readMaintenanceWindow`. Both call the same GET endpoint with the same UUID. Risk is minimal; the behaviour is identical.

**[Risk] spaces resource: GetSpaceID() returns "default" hardcoded** → After write, envelope calls `readSpaceResource(ctx, client, spaceID, "default", writtenModel)`. Since `readSpaceResource` ignores its `spaceID` argument (uses `resourceID` to fetch the space), this is correct. Verified by inspection.

**[Risk] All 6 callers change in one PR** → This is a breaking internal API change with no external consumers. All callers are in this repo and will be migrated atomically. CI will catch any missed sites.

## Migration Plan

1. Update `kibana_resource_envelope.go`:
   - Add new types (`KibanaWriteRequest`, `KibanaWriteResult`, `KibanaWriteFunc`, `KibanaPostReadFunc`, `KibanaResourceOptions`)
   - Rename `PlaceholderKibanaWriteCallbacks` → `PlaceholderKibanaWriteCallback` and change return type
   - Update `KibanaResource[T]` struct fields
   - Update `NewKibanaResource` to accept `KibanaResourceOptions[T]`
   - Implement `runKibanaWrite`
   - Update `Create`, `Update`, `Read`

2. Migrate 6 concrete resources (each is a mechanical signature update):
   - `fleet/proxy` — update create/update callback signatures
   - `kibana/streams` — update create/update callback signatures
   - `kibana/maintenance_window` — update signatures, remove manual read-after-write from `create.go`
   - `kibana/spaces` — update create/update callback signatures
   - `kibana/security_role` — update create/update callback signatures
   - `fleet/agentdownloadsource` — update `PlaceholderKibanaWriteCallback` call site

3. Update `kibana_resource_envelope_test.go` with new test cases for all new behaviors.

4. Update delta spec at `openspec/changes/kibana-envelope-parity/specs/entitycore-kibana-resource-envelope/spec.md`.

There is no rollback concern — this is an internal library change with no state format changes. All changes are compile-time breaking, so CI enforces completeness.
