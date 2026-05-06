## Context

`NewElasticsearchResource[T]` in `entitycore` already owns the full CRUD prelude for Elasticsearch-backed resources (decode model, validate write identity, resolve scoped client, invoke callback, persist state). Kibana resources have an equivalent but distinct prelude: every lifecycle method independently resolves the `KibanaScopedClient` from the `kibana_connection` block and extracts `space_id` before reaching entity-specific logic. There is no shared abstraction for this.

Two structural differences make a direct port non-trivial:

1. **Space ID is explicit, not implicit.** For Elasticsearch resources, `compID.ClusterID` identifies the ES cluster and is absorbed entirely by client resolution — callbacks never see it. For Kibana resources, `space_id` is an explicit argument to every API call and must be passed to all callbacks.

2. **Resources split into two ID management patterns.** Some resources use a user-specified name as the primary key (streams: `name`). Others receive a UUID from the API on creation and have no plan-time write identity (maintenance_window). The Elasticsearch envelope assumes a plan-time write identity (`GetResourceID()`) for all operations; Kibana Create cannot make that assumption for UUID resources.

## Goals / Non-Goals

**Goals:**
- Provide `NewKibanaResource[T]` in `entitycore` that owns Metadata, Schema (with `kibana_connection` injection), Configure, Create, Read, Update, and Delete for Kibana-backed resources.
- Support both user-ID (streams) and API-UUID (maintenance_window) resource patterns with one interface and one constructor.
- Include `PlaceholderKibanaWriteCallbacks[T]` for parity with the Elasticsearch envelope.
- Migrate `streams` and `maintenance_window` as proof-of-concept consumers.

**Non-Goals:**
- Migrating all Kibana resources (only POC resources in this change).
- Handling Fleet-backed resources (separate client type, different pattern).
- Adding version-checking (`EnforceMinVersion`) as an envelope concern (remains in callbacks for now).

## Decisions

### Decision: `GetSpaceID()` on the model interface

**Chosen:** `KibanaResourceModel` requires `GetSpaceID() types.String` alongside `GetID()`, `GetResourceID()`, and `GetKibanaConnection()`.

**Rationale:** The space is a first-class identity dimension for every Kibana resource. Making it explicit on the interface lets the envelope validate it and pass it to callbacks without requiring the envelope to understand composite ID internals. It also describes what a Kibana resource fundamentally *is* — an entity that lives in a space.

**Alternative considered:** Parse space from composite ID only. Rejected because not all Kibana resources use composite IDs as their primary state format (maintenance_window uses a plain UUID), so the envelope cannot rely on a universal composite ID structure.

---

### Decision: Composite-ID-or-fallback for Read/Update/Delete resource identity

**Chosen:** For Read, Update, and Delete, the envelope first attempts `CompositeIDFromStr(model.GetID())` (or its `Fw` variant) to parse the state ID as a composite. Both functions return error diagnostics on parse failure. Crucially, the envelope discards those diagnostics — treating the failure purely as a "not composite" signal — and falls back to `model.GetResourceID()` and `model.GetSpaceID()`. This matches the pattern already used by `getMaintenanceWindowIDAndSpaceID()` and `securitylist.Read`, which discard the parse error with `_`.

**Rationale:** This handles all current resource patterns uniformly:
- Streams: `ID = "<space>/<name>"` → composite parse succeeds → `name` + `space`.
- Maintenance window (normal): `ID = "uuid"` → parse fails → `GetResourceID()` + `GetSpaceID()`.
- Maintenance window (import): `ID = "<space>/<uuid>"` → composite parse succeeds → `uuid` + `space`.

The fallback also centralises the logic that was previously scattered across `getMaintenanceWindowIDAndSpaceID()`, `securitylist.Read`, and similar resource-local helpers.

**Alternative considered:** Require all Kibana resources to use composite IDs. Rejected because retrofitting maintenance_window (and similar UUID-based resources) to change their state ID format would be a breaking change to existing state.

---

### Decision: Create callback does not receive a `resourceID` argument

**Chosen:** `KibanaCreateFunc[T]` is `func(ctx, *KibanaScopedClient, spaceID string, plan T) (T, diag.Diagnostics)` — no `resourceID`.

**Rationale:** During Create, API-UUID resources have no write identity yet; the UUID is returned by the server. Passing an empty or unknown string for `resourceID` is misleading. For user-ID resources (streams), the callback can call `plan.GetResourceID()` directly — it has the full plan model. The spaceID is always present in the plan and is the only identity dimension the envelope can reliably validate at Create time.

**Alternative considered:** Same signature for Create and Update (`resourceID, spaceID, model`), with envelope skipping `resourceID` validation only for Create. Rejected because passing an empty string to Create callbacks conflates "not yet assigned" with "empty string", and gives user-ID resource callbacks an unexpected non-empty `resourceID` alongside the model (redundant — they already have it via `plan.GetResourceID()`).

---

### Decision: Update callback receives both plan and prior state

**Chosen:** `KibanaUpdateFunc[T]` is `func(ctx, *KibanaScopedClient, resourceID, spaceID string, plan T, prior T) (T, diag.Diagnostics)` — plan and prior state as separate arguments.

**Rationale:** Kibana APIs frequently use PATCH semantics where the body must be derived from the difference between desired and current state. Providing prior state in the callback signature means resources that need it never have to override the envelope's Update. This diverges intentionally from `ElasticsearchUpdateFunc[T]` which only receives the plan; Kibana's API surface makes this addition consistently valuable.

**Alternative considered:** Pass only the plan (ES-identical). Rejected because Update overrides for PATCH-style resources would immediately be needed, defeating the purpose of the envelope.

---

### Decision: `KibanaScopedClient` passed to callbacks, not the inner oapi client

**Chosen:** All callbacks receive `*clients.KibanaScopedClient`. Callbacks call `client.GetKibanaOapiClient()` themselves when needed.

**Rationale:** Keeps callback signatures consistent with the data source envelope pattern and leaves room for callbacks that need `GetFleetClient()` or `EnforceMinVersion()`. The inner client unwrap is one additional line; eliminating it can be a follow-up refactor once a pattern emerges.

---

### Decision: Two POC migrations (streams + maintenance_window)

**Chosen:** `streams` demonstrates the user-ID pattern; `maintenance_window` demonstrates the API-UUID pattern.

**Rationale:** Both represent real resources with distinctly shaped Create paths. Migrating both validates that the interface and callback types cover both patterns without requiring per-resource overrides.

## Risks / Trade-offs

- **Callback signature asymmetry (Create vs Update).** Create callbacks have a simpler signature than Update/Read/Delete callbacks. This is intentional but adds cognitive overhead when authoring new resources that need to implement all four. Mitigated by clear documentation and type names (`KibanaCreateFunc` vs `KibanaUpdateFunc`).

- **No write-identity validation for Create.** User-ID resources (streams) rely on their own callbacks to validate that `plan.GetResourceID()` is non-empty; the envelope does not enforce this at Create time. A misconfigured resource could reach the API with an empty name. Mitigated by the fact that the API will reject empty-name calls, and future work can add an optional `KibanaKeyedResourceModel` extended interface that adds Create-time validation.

- **Composite-ID fallback is order-sensitive.** The composite parse always wins over `GetResourceID()` + `GetSpaceID()`. If a future resource stores a plain ID that coincidentally matches the composite format (contains a `/`), the envelope would misinterpret it. Mitigated by the fact that Kibana native IDs (UUIDs, names) do not contain `/`.

## Migration Plan

No data migration required. The envelope is additive — no existing resource interfaces change. POC resources (streams, maintenance_window) are refactored internally; their Terraform schema and state format are unchanged.

Implementation sequence: envelope core → unit tests → streams migration → maintenance_window migration.
