## Context

`entitycore.KibanaResource[T]` is the canonical envelope for Kibana-scoped resources in this provider. It handles: injecting the `kibana_connection` and `timeouts` schema blocks, resolving the scoped Kibana client, validating space IDs, executing the write→read-after-write→state-set cycle, and dispatching `Configure`/`Metadata`. Resources opt in by embedding `*entitycore.KibanaResource[T]` and supplying typed callbacks for Schema, Read, Delete, Create, and Update.

The three resources in scope still embed `*entitycore.ResourceBase` (the lower-level primitive) and implement raw `resource.Resource` CRUD methods. They represent the last ResourceBase-only resources in the `kibana/` and `fleet/` trees.

## Goals / Non-Goals

**Goals:**
- Migrate all three resources to `*entitycore.KibanaResource[T]`
- Remove the hand-rolled connection/client/space wiring from each resource
- Add `entitycore_contract_test.go` to each package
- Delete the dead `synthetics.ESAPIClient` interface

**Non-Goals:**
- Changing any externally visible schema attributes or behaviours
- Migrating data sources (out of scope)
- Migrating `fleet/agentpolicy`, `fleet/elastic_defend_integration_policy`, or `fleet/integration_policy` (higher complexity, separate work)

## Decisions

### Full envelope for all three (not the partial-override pattern)

The partial-override pattern (embed `KibanaResource[T]` but shadow `Create`/`Update` on the concrete type) exists for resources whose write path can't be cleanly expressed as a single callback — for example when a resource's create is fundamentally async and returns before the entity is readable. For the three resources here:

- `fleet/integration`: the `installInSpace` + `waitForFleetIntegrationInstalled` block fits cleanly inside a `KibanaWriteFunc`. The envelope's subsequent read-after-write is safe and consistent with analogous fleet resources (`fleet/customintegration`).
- `kibana/synthetics/monitor`: the `CreateMonitor`/`UpdateMonitor` response is sufficient to populate the model; read-after-write via the envelope is a minor addition already present in `synthetics/parameter` and `synthetics/privatelocation`.
- `kibana/slo`: the write path is self-contained; see the SLO-specific decision below.

Partial override is not used here because the additional round-trip is acceptable and the full envelope gives uniform contract-test coverage.

### fleet/integration: `IsUnscopedSpace` for nullable SpaceID

`integrationModel.SpaceID` is optional — `null` means "install in the default space". The envelope validates `GetSpaceID().ValueString() != ""` in the write path unless the model implements `entitycore.KibanaUnscopedSpace`. `integrationModel` implements `IsUnscopedSpace() bool` returning `true` when `SpaceID` is null or unknown, mirroring the pattern used by other space-optional fleet resources.

### kibana/slo: reconcile moved into write callbacks, not PostRead

`reconcileSloEnabledAfterWrite` calls dedicated Enable/Disable APIs when the plan's `enabled` value differs from the server's. This logic must run during Create and Update but not during a plain Read (where there is no desired-state intent). Placing it in the write callbacks avoids any dependency on a plan-aware `PostRead` extension (which is addressed in a separate envelope change). The SLO write callback sequence:

1. Call Create/Update SLO API → obtain `sloID`
2. Call `readSloAndPopulate` (intermediate read; create response only returns `id`, not `enabled`)
3. Compare `planEnabled` vs `serverEnabled`; if different call Enable/Disable API then re-read
4. Return `KibanaWriteResult[T]{Model: model}`

The envelope then calls the read callback for the canonical read-after-write. This means up to three Kibana API calls on a create where `enabled` needs reconciling, identical to the current implementation.

`readSloFromAPI` and `readAndPopulate` are both promoted from `*Resource` methods to package-level functions. `readSloFromAPI` is called from `readAndPopulate`, from `read.go`, and from the enabled-reconcile path inside the write callbacks — all three callers are non-method closures after migration. The promoted `readSloFromAPI` signature changes to take `resourceID, spaceID string` directly (removing the internal composite-ID parse) so the read callback can pass the envelope-provided values without reconstructing the composite ID. `readSloAndPopulate` reconstructs `resourceID`/`spaceID` from `model.ID` when calling through from write callbacks.

`resolveGroupBySupport` is called inside both write callbacks (before the Kibana API call), receiving the `client *KibanaScopedClient` that the envelope passes to the write function. `EnforceVersionRequirements` is NOT called explicitly in the write callbacks — the envelope invokes it automatically before the write function because `tfModel` implements `entitycore.WithVersionRequirements`.

### kibana/synthetics/monitor: remove `ESAPIClient` entirely

`synthetics.ESAPIClient` and `synthetics.GetKibanaOAPIClient(c ESAPIClient, dg)` exist in `synthetics/api_client.go`. The compile-time assertion `_ synthetics.ESAPIClient = newResource()` in `monitor/resource.go` is the only consumer — the function is never called from any CRUD path (all paths use `r.Client().GetKibanaClient(ctx, plan.KibanaConnection)` directly). After migration the concrete `Resource` type no longer embeds `ResourceBase` directly (it comes via `KibanaResource[T]`), and the dead interface has no value. Remove the interface, function, assertion, and `GetClient()` method.

### `KibanaResourceModel` methods on model types

Each model type needs four value-receiver methods. The composite-ID pattern (`{spaceID}:{resourceID}`) used by `synthetics/monitor` and `slo` means:

- `GetID()` returns the composite (stored in state, used by ImportState)
- `GetResourceID()` returns only the resource-UUID portion (used by envelope for read-after-write lookup and as the write identity)
- `GetSpaceID()` returns the space portion

`fleet/integration` does not use a composite ID — `GetResourceID()` returns `getPackageID(name, version)`.

## Risks / Trade-offs

- **Extra read-after-write for synthetics/monitor create/update** — current code populates state directly from the API response without a second read. The envelope always reads after write. This is one extra API call per create/update; consistent with `parameter` and `privatelocation`. Risk: negligible.
- **SLO intermediate read** — the reconcile path in write callbacks does an intermediate read before the envelope's read-after-write, totalling up to two reads on create. This matches current behaviour exactly. No change in semantics.
- **Acceptance test coverage** — the migrations are purely structural; no schema attribute changes. Existing acceptance tests exercise the full CRUD path and will detect any regression.

## Migration Plan

1. Migrate `fleet/integration` (smallest, clearest reference)
2. Migrate `kibana/synthetics/monitor` (remove ESAPIClient, composite ID)
3. Migrate `kibana/slo` (promote readAndPopulate, move reconcile)

Each resource is an independent commit. All three must pass `make build` and acceptance tests before merging.
