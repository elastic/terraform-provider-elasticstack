## Why

The first ephemeral resource in this provider (`elasticstack_elasticsearch_security_api_key`, archived under `openspec/changes/archive/2026-05-20-elasticsearch-security-api-key-ephemeral/`) was hand-rolled in ~488 LOC plus a ~229-LOC connection round-trip helper. Roughly two thirds of that code is mechanical wiring that mirrors the existing `entitycore.NewElasticsearchResource[T]` / `NewElasticsearchDataSource[T]` envelopes — Configure, factory wiring, connection-block injection, scoped client resolution, version-requirement enforcement, model decode/encode boilerplate. The only genuinely new concerns for ephemerals are the Open/Close lifecycle, surviving the lack of state between Open and Close, and round-tripping the optional `elasticsearch_connection` block into Close.

Three real bugs were caught in the api_key PR review and CI loops, each of which would be eliminated structurally by a shared envelope:

1. **Connection round-trip via `tfsdk` JSON silently dropped data.** Serializing `types.String`/`types.List` through `encoding/json` produced empty objects; explicit `elasticsearch_connection` credentials were lost between Open and Close.
2. **Writing a default value to a non-computed optional attribute** in `Open()` violated the ephemeral contract and failed every shard-1 Matrix Acceptance Test (`Provider produced invalid ephemeral resource instance`).
3. **A validator hole** allowed an `access` block to pass schema validation when `type` was unset; this is a schema-shape concern that an envelope cannot fully solve but that becomes more obvious when the envelope owns the prelude.

Without a shared envelope, every future ephemeral resource (a likely roster: Kibana service-account tokens, Fleet enrollment tokens, delegate API keys via token API) has to rediscover and re-litigate these traps.

## What Changes

Add a new `entitycore` capability that provides two generic envelopes for Plugin Framework ephemeral resources — one Elasticsearch-flavored and one Kibana-flavored — mirroring the existing resource and data source envelopes. Migrate the archived `elasticstack_elasticsearch_security_api_key` ephemeral resource to the Elasticsearch variant in the same change to validate the abstraction against a real resource.

### Envelope surface

```go
// internal/entitycore/elasticsearch_ephemeral_envelope.go
// internal/entitycore/kibana_ephemeral_envelope.go

type ElasticsearchEphemeralModel interface {
    GetElasticsearchConnection() types.List
}
type KibanaEphemeralModel interface {
    GetKibanaConnection() types.List
}

type OpenRequest[T any] struct {
    Config T
}
type OpenResult[T any, S any] struct {
    Model      T
    CloseState S
}
type CloseRequest[S any] struct {
    State S
}
type CloseResponse struct{}

type ElasticsearchEphemeralOpenFunc[T ElasticsearchEphemeralModel, S any] func(
    ctx context.Context,
    client *clients.ElasticsearchScopedClient,
    req OpenRequest[T],
) (OpenResult[T, S], diag.Diagnostics)

type ElasticsearchEphemeralCloseFunc[S any] func(
    ctx context.Context,
    client *clients.ElasticsearchScopedClient,
    req CloseRequest[S],
) (CloseResponse, diag.Diagnostics)

type ElasticsearchEphemeralOptions[T ElasticsearchEphemeralModel, S any] struct {
    Schema func(context.Context) eschema.Schema
    Open   ElasticsearchEphemeralOpenFunc[T, S]
    Close  ElasticsearchEphemeralCloseFunc[S]  // required, non-nil
}

func NewElasticsearchEphemeralResource[T ElasticsearchEphemeralModel, S any](
    name string,
    opts ElasticsearchEphemeralOptions[T, S],
) ephemeral.EphemeralResource

// Kibana variant: same shape, swap ElasticsearchScopedClient for KibanaScopedClient,
// ElasticsearchEphemeralModel for KibanaEphemeralModel, and inject kibana_connection.
```

The constructor signature mirrors `NewElasticsearchResource[T]` / `NewKibanaResource[T]`: name suffix + options struct, no positional callback list. The request/response wrapper types (`OpenRequest`, `OpenResult`, `CloseRequest`, `CloseResponse`) mirror `WriteRequest[T]` / `WriteResult[T]` on the resource envelope for symmetry.

### What the envelope owns

- **Metadata**: composes `<provider_type_name>_elasticsearch_<name>` / `<provider_type_name>_kibana_<name>`.
- **Configure**: converts provider data via `clients.ConvertProviderDataToFactory`, identical to `ResourceBase.Configure`.
- **Schema**: injects an optional `elasticsearch_connection` / `kibana_connection` block, reusing the existing helpers (`providerschema.GetEsEphemeralConnectionBlock()` for the ephemeral schema namespace; the Kibana equivalent is added if it does not already exist).
- **Open prelude**: decode `Config` into `T` → resolve scoped client from `T.GetElasticsearchConnection()` (or Kibana equivalent) → enforce version requirements via `EnforceVersionRequirements` → invoke user `Open` callback → snapshot connection into a reserved envelope-owned private slot → JSON-marshal user `S` into a second reserved envelope-owned private slot → call `resp.Result.Set(ctx, model)`.
- **Close prelude**: load both private slots → restore connection to `types.List` via a typed snapshot struct (not `tfsdk`-type JSON round-trip) → resolve scoped client → unmarshal `S` → invoke user `Close` callback.
- **Plain-Go S enforcement**: at construction time, recursively reflect over `S` and `panic` with a precise error (`field path` + `type name`) if any field has a type whose `PkgPath` is `github.com/hashicorp/terraform-plugin-framework/types`. Embedded structs, slices, maps, and pointers are walked transitively. The check fires on first construction of the resource in any process — covered by the interface-implements unit test that each ephemeral resource already has.

### What the user provides

- Model `T` embedding `entitycore.ElasticsearchConnectionField` (or `KibanaConnectionField`).
- Close state `S` — plain Go types only (string, int, bool, slices, maps, embedded structs); enforced at construction.
- Schema factory `func(context.Context) eschema.Schema` returning the schema without a connection block (envelope injects it).
- Open callback: read `Config`, perform API calls, populate result attributes on the returned `Model`, return the typed `CloseState` needed for Close.
- Close callback: act on the typed `CloseState`. The envelope has already restored the connection-scoped client; the callback never sees `Private`.

### api_key migration (same change)

Replace `internal/elasticsearch/security/api_key/ephemeral_resource.go` and `ephemeral_connection.go` with a thin file built on `entitycore.NewElasticsearchEphemeralResource[ephemeralTfModel, ephemeralCloseState]`. Schema factory, validators, version-gating, and the existing acceptance tests are preserved unchanged. The `ephemeralConnectionSnapshot` / `encodeElasticsearchConnection` / `decodeElasticsearchConnection` helpers move into entitycore as internal implementation detail. The package-level `descriptions_ephemeral.go` and the docs template remain untouched.

Net effect on the api_key package: ~700 LOC of resource + connection code shrinks to ~150 LOC of schema + callbacks; round-trip bug class is eliminated structurally.

## Capabilities

### New Capabilities

- **`entitycore-ephemeral-envelope`**: Generic Elasticsearch and Kibana ephemeral resource envelopes that own Metadata, Configure, Schema (with connection block injection), Open prelude (decode → client → version checks → user callback → connection snapshot → typed close-state snapshot → Result.Set), and Close prelude (load private state → restore connection → resolve client → user callback). Enforces plain-Go close-state types at construction.

### Modified Capabilities

- _(none; the archived `elasticsearch-security-api-key-ephemeral` change has no synced spec under `openspec/specs/`, so its migration to the envelope is an implementation refactor preserved by the existing acceptance test suite.)_

## Impact

- **Specs**: New `openspec/specs/entitycore-ephemeral-envelope/spec.md` after sync/archive.
- **Implementation**:
  - New `internal/entitycore/elasticsearch_ephemeral_envelope.go`, `kibana_ephemeral_envelope.go`, and a shared `ephemeral_close_state.go` (reflect check + JSON snapshot helpers).
  - New connection-snapshot helpers internal to `entitycore` (lifted from the api_key package).
  - If the Kibana ephemeral connection block helper does not yet exist, add `providerschema.GetKbEphemeralConnectionBlock()` mirroring the existing Elasticsearch one.
  - Rewrite `internal/elasticsearch/security/api_key/ephemeral_resource.go` and delete `ephemeral_connection.go` (move into entitycore).
- **Tests**:
  - Unit tests for entitycore: connection snapshot round-trip, reflect-check positive (plain Go S passes) and negative (S with any `tfsdk` field panics at construction with a clear message), Open prelude version-gate, Close prelude with and without an explicit connection block.
  - The existing api_key ephemeral acceptance suite (`TestAccEphemeralResourceSecurityAPIKey*`) is the migration regression check; no changes expected to those test files beyond fixture adjustments only if the schema shape changes (it does not).
- **Docs**: Update `internal/entitycore/doc.go` to describe the ephemeral envelopes alongside the resource and data source ones; include the Open-on-plan known-property note so future ephemeral authors inherit the warning. The api_key ephemeral resource's generated docs are unchanged.
- **No breaking changes** to any existing resource, data source, or to the public surface of the api_key ephemeral resource. The migration is internal.
