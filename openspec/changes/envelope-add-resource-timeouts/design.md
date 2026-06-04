## Context

The Plugin Framework provides two ergonomic `timeouts` schema shapes via [`terraform-plugin-framework-timeouts`](https://github.com/hashicorp/terraform-plugin-framework-timeouts):

- `timeouts.BlockAll(ctx)` / `timeouts.Block(ctx, Opts{...})` → HCL block syntax: `timeouts { create = "10m" }`
- `timeouts.AttributesAll(ctx)` / `timeouts.Attributes(ctx, Opts{...})` → HCL attribute syntax: `timeouts = { create = "10m" }`

The block style is the legacy SDKv2 idiom; the attribute style is the modern Plugin Framework convention. All four current Plugin Framework timeouts call-sites in this provider use the resource timeouts package; three use the attribute style and one (`ml/anomalydetectionjob`) still uses the block style.

The action envelope already implements timeouts at the framework layer via `ActionTimeoutsField`, the `WithActionTimeouts` interface, auto-injected schema in `genericAction.Schema`, and ctx-wrapping in the `Invoke` prelude. It uses `actiontimeouts.Block(ctx)` because at the time the actions package was added, the framework-timeouts library only offered block style for actions. That precedent does not transfer to resources: we have both styles available, and attribute style is preferred.

Sixty-six resources today use `entitycore.NewElasticsearchResource[T]` or `entitycore.NewKibanaResource[T]`. None of these envelope-backed resources offer `timeouts` to practitioners except the four migration targets listed in the proposal.

## Goals / Non-Goals

**Goals:**

- Single uniform `timeouts` attribute exposed by every entitycore-envelope-backed resource
- Attribute-style `timeouts.AttributesAll` schema injection
- Per-op ctx-wrap (Create, Read, Update, Delete) inside the envelope, transparent to callback authors
- Per-resource override of default durations via `Options.Timeouts`, with framework-wide defaults as a fallback
- Migration path for all four existing timeouts resources, with the one breaking change called out explicitly
- Symmetric treatment in `ElasticsearchResource[T]` and `KibanaResource[T]`

**Non-Goals:**

- Retry semantics (Plugin Framework does not provide a generic retry primitive — out of scope, see issue #780)
- Migrating actions to attribute-style timeouts
- Timeouts on data sources, ephemeral resources, or the `ResourceBase`/`DataSourceBase` shared embeddings outside the envelopes
- Per-callback retry/polling helpers
- Plumbing timeouts to in-resource sub-operations (resources free to read `ctx.Deadline()` and propagate as needed)

## Decisions

### Auto-inject via compile-time model constraint

The `ElasticsearchResourceModel` and `KibanaResourceModel` interfaces tighten to embed `WithResourceTimeouts`:

```go
type WithResourceTimeouts interface {
    GetTimeouts() timeouts.Value
}

type ElasticsearchResourceModel interface {
    GetID() types.String
    GetResourceID() types.String
    GetElasticsearchConnection() types.List
    WithResourceTimeouts          // ← added
}

type KibanaResourceModel interface {
    GetID() types.String
    GetResourceID() types.String
    GetSpaceID() types.String
    GetKibanaConnection() types.List
    WithResourceTimeouts          // ← added
}
```

Every resource model that uses the envelope must satisfy this contract — embedding `entitycore.ResourceTimeoutsField` is the one-line ergonomic fix:

```go
type myModel struct {
    entitycore.ElasticsearchConnectionField
    entitycore.ResourceTimeoutsField
    // ...
}
```

This is **compile-time enforced**: any resource that fails to embed will fail `make build`. There is no opt-out and no runtime check.

**Why this over runtime opt-in:** mirrors the existing action-envelope contract; gives a single uniform user-facing surface; avoids per-resource boilerplate beyond the one embed line; compile-time safety prevents silent "this resource forgot timeouts" bugs in 62-resource churn.

**Cost accepted:** 62 model files require a one-line embed addition. The work is mechanical, reviewable as a single commit, and can be done with `goimports`/grep + manual review.

### `timeouts.AttributesAll` injection in Schema

Both envelope `Schema` methods extend their existing pattern (which already injects the connection block):

```go
// Resource envelope Schema today (sketch)
schema := r.schemaFactory(ctx)
blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
maps.Copy(blocks, schema.Blocks)
blocks[blockElasticsearchConnection] = providerschema.GetEsFWConnectionBlock()
schema.Blocks = blocks
resp.Schema = schema

// After change
schema := r.schemaFactory(ctx)
blocks := make(map[string]rschema.Block, len(schema.Blocks)+1)
maps.Copy(blocks, schema.Blocks)
blocks[blockElasticsearchConnection] = providerschema.GetEsFWConnectionBlock()
schema.Blocks = blocks

attrs := make(map[string]rschema.Attribute, len(schema.Attributes)+1)
maps.Copy(attrs, schema.Attributes)
attrs[attrTimeouts] = timeouts.AttributesAll(ctx)
schema.Attributes = attrs

resp.Schema = schema
```

`timeouts.AttributesAll` produces an attribute with Create/Read/Update/Delete sub-attributes, each optional with no default. This is the modern Plugin Framework idiom and matches the migration target syntax in `fleet/customintegration`.

### Silent overwrite of a pre-existing `timeouts` attribute

The envelope copies the factory's attributes first, then assigns the `timeouts` key. If a factory unintentionally includes a `timeouts` *attribute*, the envelope's version wins silently. This mirrors how the action envelope handles its block injection and is documented in the envelope godoc; resources MUST NOT include a `timeouts` attribute in their schema factory output.

The silent overwrite covers `schema.Attributes` only — the envelope does not touch `schema.Blocks`. After this change there is no resource left in the codebase that exposes `timeouts` as a block: `ml/anomalydetectionjob` is being migrated to attribute style as part of task 9. Plugin Framework rejects schemas whose `Attributes` and `Blocks` maps share a key, so any future regression that re-introduces a `timeouts` block alongside the envelope-injected attribute will surface as a hard schema-validation failure at startup — caught immediately, no special handling needed.

We considered panicking on attribute collision, but discarded that approach to stay consistent with the action envelope's precedent and to make the 4-resource migration smoother (interim builds during a phased migration would otherwise need the schema removal and the envelope wiring to land in lock-step).

### Per-op default durations via `Options.Timeouts`

```go
type ResourceTimeouts struct {
    Create time.Duration
    Read   time.Duration
    Update time.Duration
    Delete time.Duration
}

const (
    DefaultResourceCreateTimeout = 20 * time.Minute
    DefaultResourceReadTimeout   =  5 * time.Minute
    DefaultResourceUpdateTimeout = 20 * time.Minute
    DefaultResourceDeleteTimeout = 20 * time.Minute
)

type ElasticsearchResourceOptions[T ElasticsearchResourceModel] struct {
    // ...existing fields...
    Timeouts ResourceTimeouts
}
```

Any zero-valued field falls back to the corresponding `DefaultResource<Op>Timeout` constant. A resource that wants 5 minutes for Create writes `Timeouts: entitycore.ResourceTimeouts{Create: 5*time.Minute}` and inherits framework defaults for the other three.

### Ctx-wrap inside the envelope, per op

Each of `Create`, `Read`, `Update`, `Delete` in both envelopes gains an early ctx-wrap after the model is decoded but before any client call:

```go
// Generic shape; specific op uses .Create / .Read / .Update / .Delete
defaultTimeout := opts.Timeouts.Create
if defaultTimeout <= 0 {
    defaultTimeout = DefaultResourceCreateTimeout
}
createTimeout, timeoutDiags := model.GetTimeouts().Create(ctx, defaultTimeout)
resp.Diagnostics.Append(timeoutDiags...)
if resp.Diagnostics.HasError() {
    return
}
ctx, cancel := context.WithTimeout(ctx, createTimeout)
defer cancel()
```

#### Ordering inside each operation

The wrap happens immediately after model decode and **before client resolution and version-requirement enforcement**, so every API-touching step runs under the timeout. The full ordering for each op becomes:

1. Decode model from `req.Plan`/`req.State`/`req.Config`
2. **Wrap ctx with `context.WithTimeout`** (this change)
3. Resolve scoped client via `GetElasticsearchClient` / `GetKibanaClient`
4. `EnforceVersionRequirements` (issues a Stack info API call)
5. Invoke the resource callback (create/read/update/delete write/read func)
6. Envelope read-after-write (Create/Update only)
7. PostRead hook (if configured)

Placing the wrap at step 2 means the version-check API call, the client connection probing, the resource callback, and any read-after-write fan-out all share the single deadline. If `EnforceVersionRequirements` takes too long, the timeout fires. This matches how the action envelope wraps ctx before its single Invoke callback.

For the Update path, the timeout is derived from the plan model (not prior state), matching how the action envelope reads its model from configuration. For Read and Delete, the timeout is derived from the state model.

#### Behavior on null/unknown `timeouts` (upgraded state)

The framework-timeouts library (`v0.7.0` `resource/timeouts/timeouts.go`) is graceful when a `timeouts` value is missing, null, or unknown:

```go
// getTimeout in terraform-plugin-framework-timeouts:
if !ok            { return defaultTimeout, nil }  // attribute key missing
if value.IsNull() || value.IsUnknown() {
    return defaultTimeout, nil                    // null/unknown → default
}
```

Practitioners upgrading from a provider version that did not expose `timeouts` will refresh state with `Timeouts` decoded as a null `timeouts.Value`. The framework returns the envelope-supplied default with no diagnostics, so Read/Refresh/Import after upgrade just works. Diagnostics are only produced for an explicitly-set value that fails `time.ParseDuration` — that is a practitioner config error and correctly aborts the operation.

### Migration mechanics for the four existing timeouts resources

The three attribute-style resources (`customintegration`, `jobstate`, `datafeed_state`) migrate cleanly:

```
BEFORE                                    AFTER
──────                                    ─────
models.go:                                models.go:
  Timeouts timeouts.Value `tfsdk:"timeouts"`   entitycore.ResourceTimeoutsField

schema.go:                                schema.go:
  "timeouts": timeouts.Attributes(...)         (deleted — envelope injects)

create.go / update.go:                    create.go / update.go:
  t, _ := plan.Timeouts.Create(ctx, 20m)       (deleted — envelope wraps)
  ctx, cancel := context.WithTimeout(...)
  defer cancel()

resource.go:                              resource.go:
  NewKibanaResource[T](..., opts)              NewKibanaResource[T](..., opts{
                                                 Timeouts: ResourceTimeouts{
                                                   Create: 20*time.Minute,
                                                   Update: 20*time.Minute,
                                                 },
                                               })
```

Schema effect: the three resources gain `read` and `delete` timeout sub-attributes that they previously did not advertise. This is additive — existing configurations remain valid — and the new sub-attributes have no behavioral effect for these resources beyond bounding their respective callbacks (which is desirable).

The block-style resource (`anomalydetectionjob`) migrates the same way but with a config-syntax break: practitioners writing `timeouts { delete = "10m" }` must rewrite to `timeouts = { delete = "10m" }`. This is the only breaking change in the proposal and is called out in CHANGELOG and resource docs.

### Default values for migration targets

Defaults are preserved from each resource's existing hard-coded durations:

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| `fleet/customintegration` | 20m | (framework default 5m) | 20m | (framework default 20m) |
| `ml/jobstate` | 5m | (framework default 5m) | 5m | (framework default 20m) |
| `ml/datafeed_state` | 5m | (framework default 5m) | 5m | (framework default 20m) |
| `ml/anomalydetectionjob` | (framework default 20m) | (framework default 5m) | (framework default 20m) | 20m |

## Risks / Trade-offs

### Compile-time breakage of 62 resources during model embed rollout

Tightening `ElasticsearchResourceModel`/`KibanaResourceModel` makes the build fail for every resource that hasn't yet embedded `ResourceTimeoutsField`. Mitigation: land the embed additions in the same PR as the envelope change. The mechanical edit is a single line per file and tractable to do in one commit.

Alternative considered: introduce `ResourceTimeoutsField` and the interface first, leave the model constraint loose, then tighten in a follow-up. Rejected because it leaves an indefinite window where new resources can be merged without the embed and silently miss out on timeouts.

### PF timeouts only matter if downstream code honors ctx

The Plugin Framework timeouts library wraps `context.WithTimeout` around the operation. Cancellation propagates only if the underlying HTTP client respects `ctx.Done()`. The Elasticsearch Go client and the generated Kibana client used here both pass ctx through their request paths, so timeouts will actually fire. We note this in the envelope godoc but accept it as the normal Plugin Framework contract.

### Read timeout interaction with `terraform refresh` / `terraform import`

The `Read` callback is invoked on `terraform refresh`, `terraform plan` (refresh phase), and `terraform import`. A too-tight default would surface as a transient failure on slow Kibana instances. The chosen 5-minute default mirrors common Plugin Framework defaults and is overridable per resource via `Options.Timeouts.Read`.

### Breaking change for `ml/anomalydetectionjob` block-style users

Only one resource is affected, and only users who explicitly configured `timeouts { delete = "10m" }`. Mitigation: CHANGELOG `BREAKING CHANGES` entry with copy-paste before/after HCL; resource docs include a migration note.

We considered keeping block style for this single resource as a one-off (i.e. providing a deprecated `timeouts.Block` alongside `timeouts.AttributesAll`), but this would permanently split the envelope contract and is not worth the dual code path.

### Provider documentation churn

`make docs-generate` will update all 66 entitycore-envelope resource doc pages (37 Elasticsearch + 29 Kibana). 62 of those gain a `timeouts` section for the first time; the 4 migration targets see their existing section either expand from C/U to C/R/U/D (attribute-style three) or switch from block to attribute syntax (anomaly detection job). The diff is large but mechanical and reviewable as a separate commit.

## Migration Plan

Single PR with the envelope change, all 62 mechanical embeds, all 4 resource migrations, regenerated docs, and the CHANGELOG entry. The work is large but coupled: any partial application either leaves the build broken (due to the model interface tightening) or leaves the 4 migrated resources with a schema collision against their own pre-migration `timeouts` attribute.

Suggested commit structure inside the PR:

1. Envelope additions: `resource_timeouts.go`, envelope edits, envelope test updates
2. Mechanical model embeds: 62 single-line additions
3. Migration of the 4 existing timeouts resources (3 additive + 1 breaking)
4. `make docs-generate` output
5. CHANGELOG entry
