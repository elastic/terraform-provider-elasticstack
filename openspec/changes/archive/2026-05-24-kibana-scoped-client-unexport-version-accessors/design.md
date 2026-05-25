## Context

`internal/clients/kibana_scoped_client.go` exposes four version-related methods today:

- `ServerVersion(ctx) (*version.Version, diags)` — raw Kibana server version
- `ServerFlavor(ctx) (string, diags)` — raw build flavor (`"serverless"` | `"default"` | `""`)
- `EnforceMinVersion(ctx, min) (bool, diags)` — min-version check, serverless short-circuits to true
- `EnforceVersionCheck(ctx, check func(*version.Version) bool) (bool, diags)` — predicate
  check, serverless short-circuits to true

All four route through a single private `getServerStatusRaw`. The two safe methods exist precisely
so callers do not have to combine version comparisons with flavor checks themselves. Yet the
unsafe raw accessors remain on the public surface and are still the most commonly used: 12 callers
in `internal/kibana/` use `ServerVersion`, 0 use `ServerFlavor`. The
`fix-serverless-version-gating` change archived 2026-05-16 fixed two such broken comparisons in
Elasticsearch — the underlying type-level hazard remained.

The Kibana resource envelope (`internal/entitycore/kibana_resource_envelope.go`) already calls
`EnforceVersionRequirements(ctx, client, &model)` automatically during Create, Read, and Update
when the model implements `entitycore.WithVersionRequirements`. Roughly two thirds of the current
`ServerVersion` callers are pure gates that fit this interface directly.

## Goals / Non-Goals

**Goals:**

- Remove `ServerVersion` and `ServerFlavor` from `KibanaScopedClient`'s public method set.
- Migrate every `internal/kibana/` caller to a serverless-safe primitive.
- Preserve user-visible behavior on stateful clusters; improve behavior on serverless (gates
  that previously rejected valid configuration will now accept it).
- Adjust `provider-client-factory` spec so the Kibana scoped-client contract reflects the
  new surface.

**Non-Goals:**

- Touching `ElasticsearchScopedClient` (covered by
  `elasticsearch-scoped-client-unexport-version-accessors`).
- Changing the wire shape of any Kibana API request beyond what's incidentally needed when
  serverless is now reachable.
- Introducing a new lint rule. The compiler-level removal is the enforcement.

## Decisions

### Decision 1: Delete `ServerFlavor` rather than make it private

Flavor on the Kibana scoped client has zero production callers. The only references are two test
files in `internal/clients/`. Keeping a private `serverFlavor` accessor would tempt future
contributors to widen it again. We delete the method outright; flavor stays available inside
`getServerStatusRaw` as a local variable used by `EnforceMinVersion`/`EnforceVersionCheck`.

The companion ES change keeps a public `IsServerless(ctx) bool` because two real ES callers need
to ask "am I on serverless?" for non-version reasons. The Kibana client has no such caller, so no
equivalent helper is added.

**Alternative considered**: keep a public `IsServerless(ctx)` on the Kibana client for parity.
Rejected — adds API surface with no consumer.

### Decision 2: Capability struct for `alertingrule`, modelled on `agentpolicy.features`

The two `alertingrule` callers thread `serverVersion` into `plan.toAPIModel(ctx, serverVersion)`,
which uses it for version-specific field validation and shape decisions. Three options were
considered:

| Option | Pros | Cons |
|---|---|---|
| `EnforceMinVersion` per feature, inline inside `toAPIModel` | smallest diff | spreads version logic into the builder; `toAPIModel` needs access to the client |
| Capability struct resolved up front, passed in | mirrors `agentpolicy.features` precedent; builder stays pure | new type + new resolution function |
| Keep package-private `serverVersion` accessor for alertingrule only | smallest API surface change | leaves the hazard for the one resource that has it; undermines the refactor |

We pick the capability-struct option. `agentpolicy.resolveFeatures` is the established precedent
in this codebase (called from `internal/fleet/agentpolicy/resource.go`); reusing the shape gives
new contributors a familiar pattern to copy.

```go
type alertingRuleFeatures struct {
    SupportsX bool
    SupportsY bool
}

func resolveAlertingRuleFeatures(ctx context.Context, client *clients.KibanaScopedClient) (alertingRuleFeatures, diag.Diagnostics) { ... }
```

The exact field set is determined by what `toAPIModel` currently branches on — to be enumerated
during implementation by reading `alertingrule/models*.go` and identifying every comparison
against `serverVersion`. The number of `EnforceMinVersion` calls in `resolveAlertingRuleFeatures`
equals the number of distinct version thresholds in the existing builder.

### Decision 3: `WithVersionRequirements` for Pattern A, `EnforceMinVersion` boolean for Pattern B

Eight callers are pure gates (`if LessThan(min) { addError; return }`). These fit
`entitycore.WithVersionRequirements` cleanly: the envelope evaluates the requirement before the
write callback runs and emits a consistent "Unsupported server version" diagnostic — the same
message the existing inline code uses. Migration is mechanical: move the constant + error message
into a `GetVersionRequirements()` method on the model, delete the inline block.

Two callers (`synthetics/parameter/delete`, `agentbuilderagent/data_source`) use the version as
a feature toggle, not a gate (newer path vs older path; no error in either case). For these,
`EnforceMinVersion` already returns the boolean we want:

```go
useNewAPI, diags := client.EnforceMinVersion(ctx, minKibanaPerIDDeleteVersion)
```

Serverless gets `true` and selects the new path — which is the correct behavior, because
serverless Kibana exposes the newer per-ID API.

### Decision 4: `connectors/create` uses `WithVersionRequirements` with a conditional

`connectors/create` only enforces a minimum version when `apiModel.ConnectorID != ""`. The
`GetVersionRequirements()` method inspects the model and returns either an empty slice or a
single requirement, mirroring `componenttemplate.Data.GetVersionRequirements()` (which inspects
`data_stream_options` presence). No new infrastructure needed.

### Decision 5: Tests exercise the public surface only

Today's `kibana_scoped_client_test.go` directly tests `ServerVersion` and `ServerFlavor` against
HTTP fixtures. The new tests test the same fixtures through `EnforceMinVersion` and
`EnforceVersionCheck`:

- `TestKibanaScopedClient_EnforceMinVersion_MissingEndpoint`
- `TestKibanaScopedClient_EnforceMinVersion_StatefulBelowMin` (was version comparison test)
- `TestKibanaScopedClient_EnforceMinVersion_StatefulAtOrAboveMin`
- `TestKibanaScopedClient_EnforceMinVersion_ServerlessShortCircuits` (was the flavor test)
- `TestKibanaScopedClient_EnforceVersionCheck_*` analogues

`provider_client_factory_test.go::TestKibanaScopedClient_ServerFlavor_ViaFactory` becomes
`TestKibanaScopedClient_EnforceMinVersion_ViaFactory` and asserts the serverless short-circuit
through the factory-obtained client.

### Decision 6: Migration order inside the change

Land in this order so `make build` and `make test` stay green at every commit:

1. Add `WithVersionRequirements` to the eight Pattern A models (no caller changes yet — envelope
   now enforces in addition to the inline check; both produce the same error).
2. Delete the inline `ServerVersion` blocks in those eight callers; envelope is now the sole
   gate.
3. Migrate the two Pattern B callers.
4. Introduce `alertingRuleFeatures`, rewrite `toAPIModel`, migrate the two `alertingrule`
   callers.
5. Rewrite tests in `internal/clients/`.
6. Delete `ServerVersion` and `ServerFlavor` from `kibana_scoped_client.go`.
7. Update `provider-client-factory` spec.

## Risks / Trade-offs

- **Risk**: A caller we missed still imports `ServerVersion`/`ServerFlavor` and breaks the build.
  → Mitigation: `rg "\.ServerVersion\(|\.ServerFlavor\(" internal/kibana internal/clients`
  before the final delete; CI catches any miss because the methods will not exist.
- **Risk**: `WithVersionRequirements` evaluates on every Create/Read/Update, including refresh.
  Existing inline checks ran only at the call site they were placed in (often only Create or
  only Update). Behavior change: a resource created on a supported cluster, then refreshed
  against a downgraded one, will now error on refresh.
  → Mitigation: this is the same behavior change made by `fix-serverless-version-gating`'s
  read-time enforcement scenario and is already established as correct in the
  `elasticsearch-index-template` spec. Documented as a scenario in the per-spec deltas if any
  resource specs gain explicit scenarios; otherwise inherits envelope behavior.
- **Trade-off**: `alertingRuleFeatures` is a small new abstraction for a two-call-site case. The
  payoff is uniformity with `agentpolicy.features` and a clean serverless story for the
  alerting rule shape.
- **Risk**: Acceptance tests (`internal/kibana/.../acc_test.go`) may directly call
  `client.ServerVersion()` or `client.ServerFlavor()`.
  → Mitigation: `rg` across `internal/kibana/**/acc_test.go` during implementation; migrate any
  hits to `versionutils.CheckIfNotServerless` or `EnforceMinVersion`.
