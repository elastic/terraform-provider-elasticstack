## Context

`KibanaPostReadFunc[T]` and `PostReadFunc[T]` (the Elasticsearch equivalent) are optional hooks in `KibanaResourceOptions[T]` / `ElasticsearchResourceOptions[T]` for logic that must run after a successful read that persisted state — both on the plain Read path and on the read-after-write path inside Create/Update. Their current signatures are:

```go
// Kibana
type KibanaPostReadFunc[T KibanaResourceModel] func(
    ctx context.Context,
    client *clients.KibanaScopedClient,
    model T,
    privateState any,
) diag.Diagnostics

// Elasticsearch
type PostReadFunc[T ElasticsearchResourceModel] func(
    ctx context.Context,
    client *clients.ElasticsearchScopedClient,
    model T,
    privateState any,
) diag.Diagnostics
```

Both call sites (plain Read and read-after-write) pass the freshly-read model and the private state. The hook can only inspect the model — it has no access to what existed before the read.

`kibana/dashboard` needs PostRead to run panel layout alignment: four functions that reconcile server-returned panels/sections/pinned-panels against a reference model. On the write path the reference is the plan; on the plain Read path the reference is the prior state. Both cases require a reference model the current signature cannot provide.

No existing Kibana caller passes a non-nil `PostRead` today. The sole Elasticsearch caller, `postReadPersistAPIKeyCapabilities` in `elasticsearch/security/apikey/resource/resource.go`, ignores the model entirely — the signature update is mechanical.

## Goals / Non-Goals

**Goals:**
- Replace both PostRead function types with struct-based callbacks matching the style of `KibanaWriteFunc[T]`
- Provide `Prior T` (reference model before the read) and `State T` (freshly-read model) in the request struct
- Return `(T, diag.Diagnostics)` so PostRead can modify the model before it is committed to state
- Apply the same contract change to both Kibana and Elasticsearch envelopes
- Migrate `kibana/dashboard` to `*entitycore.KibanaResource[DashboardModel]` using the new PostRead
- Update all PostRead tests in both envelope test files

**Non-Goals:**
- Changing any externally visible schema attributes or behaviours
- Migrating any other resource (dashboard is the only new PostRead consumer)

## Decisions

### Struct-based callback with `Prior T` and returned model

New types:

```go
// Kibana
type KibanaPostReadRequest[T KibanaResourceModel] struct {
    Client  *clients.KibanaScopedClient
    Prior   T   // plan on write path; prior state on plain Read path
    State   T   // freshly-read model from the read callback
    Private any
}

type KibanaPostReadFunc[T KibanaResourceModel] func(
    ctx context.Context,
    req KibanaPostReadRequest[T],
) (T, diag.Diagnostics)

// Elasticsearch — identical shape, different client type
type ElasticsearchPostReadRequest[T ElasticsearchResourceModel] struct {
    Client  *clients.ElasticsearchScopedClient
    Prior   T
    State   T
    Private any
}

type PostReadFunc[T ElasticsearchResourceModel] func(
    ctx context.Context,
    req ElasticsearchPostReadRequest[T],
) (T, diag.Diagnostics)
```

`Prior` is named for its semantic role: it carries what existed before this read. On the write path that is the plan (the desired state the practitioner requested, prior to the API call and read-after-write). On the plain Read path it is the prior persisted state (what was in state before this refresh). The same field name covers both callers without needing path-specific variants.

The struct form is consistent with `KibanaWriteRequest[T]` / `KibanaWriteResult[T]` and accommodates future fields without another signature break.

### State is set AFTER PostRead (order change)

Current call order for both envelopes:
1. Read callback → model
2. `resp.State.Set(model)`
3. PostRead(model, private) → diagnostics

New order:
1. Read callback → `req.State`
2. PostRead(req) → (model, diagnostics)
3. `resp.State.Set(model)`

PostRead now owns the final model committed to state. If PostRead returns error diagnostics, `resp.State.Set` is not called — consistent with the write callback contract.

The order change applies to both the plain Read path (`runKibanaRead` / `runElasticsearchRead`) and the write path (`runKibanaWrite` / `runElasticsearchWrite`).

### `Prior` population per path

**Write path** (`runKibanaWrite` / `runElasticsearchWrite`): `Prior` is set to the plan model decoded from `req.Plan`, identical to `KibanaWriteRequest.Plan`.

**Plain Read path** (`runKibanaRead` / `runElasticsearchRead`): `Prior` is set to the state model decoded from the incoming state before the read callback is invoked. The read callback receives this model to resolve the resource ID and space; the same value becomes `req.Prior`.

### `postReadPersistAPIKeyCapabilities` update

The sole Elasticsearch PostRead user ignores the model (`_ apikey.TfModel`). Updated signature:

```go
func postReadPersistAPIKeyCapabilities(
    ctx context.Context,
    req entitycore.ElasticsearchPostReadRequest[apikey.TfModel],
) (apikey.TfModel, diag.Diagnostics) {
    priv, ok := req.Private.(privateData)
    // ... same body
    return req.State, diags
}
```

### dashboard PostRead implements all four alignment functions

`kibana/dashboard` currently scatters alignment across three call sites:
- `read.go`: `alignDashboardStateFromPlanPanels`, `suppressReadTopLevelPanelsWhenPlanEmpty`, `alignDashboardStateFromPlanSections`, `alignDashboardStateFromPlanPinnedPanels` — called with prior state as reference
- `create.go`: same four functions — called with plan as reference
- `update.go`: same four functions — called with plan as reference

After migration, all three paths converge in `postReadDashboard`:

```go
func postReadDashboard(ctx context.Context, req entitycore.KibanaPostReadRequest[models.DashboardModel]) (models.DashboardModel, diag.Diagnostics) {
    alignDashboardStateFromPlanPanels(req.Prior.Panels, req.State.Panels)
    suppressReadTopLevelPanelsWhenPlanEmpty(req.Prior.Panels, &req.State)
    alignDashboardStateFromPlanSections(ctx, req.Prior.Sections, req.State.Sections)
    alignDashboardStateFromPlanPinnedPanels(ctx, req.Prior.PinnedPanels, req.State.PinnedPanels)
    return req.State, nil
}
```

- Write path: `req.Prior` = plan → aligns server panels against practitioner intent
- Plain Read path: `req.Prior` = prior state → stabilises drift from server-set panel defaults

The `read.go` callback becomes a pure API call: fetch dashboard, populate model, return. No alignment. The `create.go` and `update.go` callbacks set `model.ID` and `model.DashboardID` and return; no alignment. The envelope's read-after-write calls the read callback, then PostRead applies alignment.

### `DashboardModel` composite ID

`DashboardModel` field `ID` holds `<space_id>/<dashboard_id>`; `DashboardID` holds the naked UUID. Interface methods:
- `GetID()` → `m.ID`
- `GetResourceID()` → `m.DashboardID`
- `GetSpaceID()` → `m.SpaceID`
- `GetKibanaConnection()` → `m.KibanaConnection`

## Risks / Trade-offs

- **Order change could affect future PostRead users** — any future implementation that expected state to already be set when PostRead runs would need revision. Risk: low; there are no such users and the new contract is documented clearly.
- **ES PostRead is purely mechanical** — `postReadPersistAPIKeyCapabilities` ignores the model and only touches private state; the signature change does not affect its logic.
- **Extra read-after-write for dashboard create/update** — the envelope always reads after write. Current code also does this (via `r.read(ctx, client, planModel)`), so this is a no-change in API call count.

## Migration Plan

1. Change Kibana PostRead types and update `runKibanaRead`/`runKibanaWrite` in `kibana_resource_envelope.go`; update and extend `kibana_resource_envelope_test.go`
2. Change Elasticsearch PostRead types and update `runElasticsearchRead`/`runElasticsearchWrite` in `resource_envelope.go`; update `postReadPersistAPIKeyCapabilities`; update and extend `resource_envelope_test.go`
3. Migrate `internal/kibana/dashboard` using the new PostRead
4. Verify `make build` passes; run dashboard acceptance tests
