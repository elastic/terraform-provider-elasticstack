## Context

`internal/clients/elasticsearch_scoped_client.go` exposes five version/flavor-related methods today:

- `ServerVersion(ctx) (*version.Version, fwdiag.Diagnostics)`
- `ServerFlavor(ctx) (string, fwdiag.Diagnostics)`
- `EnforceMinVersion(ctx, min) (bool, fwdiag.Diagnostics)`
- (private) `serverInfo(ctx) (*info.Response, fwdiag.Diagnostics)` — cached cluster Info

Unlike the Kibana scoped client, there is no `EnforceVersionCheck` predicate variant, so callers
that need a non-min check fall back to `ServerVersion` + arbitrary comparisons. Eight
non-test references exist in `internal/elasticsearch/`:

| Caller | File | Pattern | Notes |
|---|---|---|---|
| pure gate | `security/role/update.go` | A | two conditional gates: description, remote_indices |
| feature toggle (flavor) | `index/index/create.go` | B | `toPutIndexParams(serverFlavor)` swaps stateful-only knobs |
| feature toggle (version) | `transform/version_gating.go` via `transform/write.go` | B | per-setting min version map |
| capability persistence | `security/apikey/runtime_validation.go` | C | `with_restriction` |
| capability persistence | `security/apikey/models.go` | C | `PopulateRoleDescriptorsDefaults` |
| capability persistence | `security/apikey/resource/resource.go` | C | persists version into Terraform private state |
| capability persistence | `security/apikey/resource/schema.go` | C | `RequiresReplaceIf` reads cached version |
| (acc test) | `security/apikey/resource/acc_test.go` | predicate | `LessThan || GreaterThanOrEqual` range check |
| (test util) | `versionutils/testutils.go` | both | acceptance-test skip plumbing |

The pattern C cluster of callers all live in `internal/elasticsearch/security/apikey/`. The
resource persists a version string into Terraform private state during read and consumes it from
plan-modifiers during plan computation. This is the most invasive migration in the change.

## Goals / Non-Goals

**Goals:**

- Remove `ServerVersion` and `ServerFlavor` from the public surface of `ElasticsearchScopedClient`.
- Expose `IsServerless(ctx)` and `EnforceVersionCheck(ctx, check)` as the only new public APIs.
- Migrate every `internal/elasticsearch/` caller, including the apikey private-state plumbing,
  to the new primitives.
- Preserve user-visible behavior on stateful clusters; improve behavior on serverless
  (today an apikey created against stateful and refreshed against serverless would store a
  serverless version string into private state with undefined behavior under `LessThan`).
- Keep acceptance-test skip plumbing working — but route it through a test-only export so
  production code cannot reach the raw version.

**Non-Goals:**

- Touching `KibanaScopedClient` (covered by `kibana-scoped-client-unexport-version-accessors`).
- Refactoring the apikey resource beyond what is needed to migrate the version-gated code paths.
- Auditing or changing version constants (`MinVersionWithUpdate`,
  `MinVersionReturningRoleDescriptors`, `MinVersionWithRestriction`, etc.).

## Decisions

### Decision 1: `IsServerless(ctx) (bool, fwdiag.Diagnostics)` over `ServerFlavor`

Two real callers ask the flavor question (`index/index/create`, `versionutils.CheckIfNotServerless`).
Both immediately compare against `clients.ServerlessFlavor`. A focused predicate eliminates the
need to either know the constant or handle the empty-string case.

`ServerFlavor` becomes a package-private `serverFlavor(ctx) (string, fwdiag.Diagnostics)` helper
used inside `IsServerless`, `EnforceMinVersion`, and `EnforceVersionCheck`. We could inline it
into `IsServerless` instead, but keeping it shared avoids three separate `info.Version.BuildFlavor`
reads.

**Alternative considered**: keep `ServerFlavor` public and add a documentation comment forbidding
its use for version comparison. Rejected — the type system, not a comment, should enforce intent.

### Decision 2: `EnforceVersionCheck(ctx, check func(*version.Version) bool)` mirrors the Kibana client

Same signature and short-circuit semantics as
`KibanaScopedClient.EnforceVersionCheck`. Required for:

- `apikey/resource/acc_test.go::checkAPIKeyVersionSupport` — range check
  (`LessThan(min) || GreaterThanOrEqual(other)`) — moves to `EnforceVersionCheck`.
- Plan-modifiers in `apikey/resource/schema.go` if they need predicate-style checks against
  cached capabilities — though after the private-state migration they will read booleans
  directly.

The `MinVersionClient` interface in `internal/entitycore/version_requirements.go` continues to
require only `EnforceMinVersion`. We do not add `EnforceVersionCheck` to that interface in this
change because no model-level `WithVersionRequirements` consumer needs predicate checks today;
keeping the interface narrow preserves existing envelope wiring.

### Decision 3: Capability struct for the apikey package, replacing the private-state version string

Today `internal/elasticsearch/security/apikey/resource/resource.go` writes
`{"Version": "8.x.y"}` into the `cluster-version` private-state slot during `postReadPersistClusterVersion`.
`schema.go::requiresReplaceIfUpdateNotSupported` reads it back during plan-modify and does
`ver.LessThan(apikey.MinVersionWithUpdate)`.

After this change, the persisted blob becomes:

```json
{
  "SupportsUpdate":          true,
  "SupportsRoleDescriptors": true,
  "SupportsRestriction":     true
}
```

— resolved via three `client.EnforceMinVersion(ctx, ...)` calls in
`resolveAPIKeyCapabilities(ctx, client) (apikeyCapabilities, fwdiag.Diagnostics)`. On serverless
all three bits are `true` (correct: serverless supports all current apikey features).

**Backward compatibility**: existing state may contain the legacy `{"Version": "8.x.y"}` blob.
`clusterVersionOfLastRead` is replaced by `apikeyCapabilitiesOfLastRead`, which:

1. Reads the private-state slot bytes.
2. Attempts `json.Unmarshal` into `apikeyCapabilities`. If `SupportsUpdate` is non-zero or any
   field is set, return as-is.
3. On failure, attempt `json.Unmarshal` into the legacy `{Version string}` shape. If it succeeds
   with a non-empty `Version`, parse it and synthesize an `apikeyCapabilities` from the same
   `EnforceMinVersion`-equivalent comparisons (`!parsedVer.LessThan(MinVersionWithUpdate)`, etc.).
3. Return the synthesized capabilities.
4. The next `postReadPersistClusterVersion` overwrites the slot with the new shape (renamed
   `postReadPersistCapabilities`).

The slot key remains `cluster-version` for state migration simplicity; we treat the bytes as
opaque and discriminate by JSON shape. Either shape is acceptable for the duration of one read
cycle.

**Alternative considered**: change the private-state key (`apikey-capabilities`) and drop legacy
data. Rejected — would force every existing apikey resource to do an extra refresh round-trip on
upgrade to populate the new slot before the next plan can succeed.

### Decision 4: Acceptance-test access via a separate exported helper

`internal/versionutils/testutils.go::fetchAcceptanceServerInfo` legitimately needs both the
server version and whether the cluster is serverless — to print "version below required minimum"
skip messages. We do not want a public production API for this.

Add a new file `internal/clients/acceptance_testing_version.go` (compiled always; this file lives
alongside the existing `NewAcceptanceTestingElasticsearchScopedClient`, which is itself a
test-only constructor):

```go
// AcceptanceServerInfo returns the connected Elasticsearch server version and a boolean
// indicating whether the cluster is serverless. It is exposed only for acceptance-test skip
// plumbing in internal/versionutils. Production code SHALL NOT call this.
func AcceptanceServerInfo(ctx context.Context, c *ElasticsearchScopedClient) (*version.Version, bool, fwdiag.Diagnostics) { ... }
```

The function is unexported-style by convention (documented "test-only") and lives in `internal/clients`
where Go's `internal/` visibility already restricts use to the module itself. We do not add a
build tag because some downstream tooling (e.g. coverage runs) compiles all files; the
documentation comment and a `// Deprecated: test-only` marker are sufficient.

**Alternative considered**: put the function in `internal/clients/acceptance_testing.go` behind
`//go:build acceptance`. Rejected — current acceptance tests do not use a build tag; introducing
one is out of scope.

### Decision 5: Transform per-setting gating becomes per-setting `EnforceMinVersion`

`internal/elasticsearch/transform/version_gating.go::isSettingAllowed` currently takes a
`*version.Version` parameter. Two reasonable rewrites:

- **Inline per-call**: change `isSettingAllowed` to take `(ctx, *clients.ElasticsearchScopedClient, settingName)`
  and call `EnforceMinVersion(ctx, settingsRequiredVersions[settingName])` internally. Minimal
  diff.
- **Capability struct**: resolve all four settings to booleans up front in
  `transform/write.go`, pass into `isSettingAllowed`.

We pick option 1 because the transform settings table is small (4 entries), the call sites are
already inside the write function, and the structure of `isSettingAllowed` (returns false +
logs a warning) does not benefit from precomputation.

### Decision 6: Index resource — `toPutIndexParams(isServerless bool)`

`internal/elasticsearch/index/index/models.go::toPutIndexParams` currently takes a
`serverFlavor string` parameter and compares against `clients.ServerlessFlavor`. The parameter
becomes a `bool`:

```go
func (model tfModel) toPutIndexParams(isServerless bool) models.PutIndexParams { ... }
```

`create.go` becomes:

```go
isServerless, isDiags := client.IsServerless(ctx)
if isDiags.HasError() { ... }
params := planModel.toPutIndexParams(isServerless)
```

The existing test `Test_tfModel_toPutIndexParams` (`models_test.go:503`) already iterates over
`isServerless := []bool{true, false}` — it directly maps to the new signature with no other
changes.

### Decision 7: Migration order inside the change

Land in this order so `make build` and `make test` stay green at every commit:

1. Add `IsServerless` and `EnforceVersionCheck` to `ElasticsearchScopedClient` (additive).
2. Migrate `security/role/update.go` to `WithVersionRequirements`.
3. Migrate `transform/write.go` + `version_gating.go` to per-setting `EnforceMinVersion`.
4. Migrate `index/index/create.go` to `IsServerless`; flip `toPutIndexParams` signature; update
   `Test_tfModel_toPutIndexParams`.
5. Introduce `apikeyCapabilities`, `resolveAPIKeyCapabilities`, capability-aware
   `apikeyCapabilitiesOfLastRead` (with legacy-blob fallback), `postReadPersistCapabilities`.
   Migrate `runtime_validation.go`, `models.go`, `resource.go`, `schema.go`. Acceptance test
   migration in `acc_test.go`.
6. Migrate `versionutils/testutils.go` to `IsServerless` + new `AcceptanceServerInfo` helper.
7. Rewrite `internal/clients/elasticsearch_scoped_client_test.go` and
   `provider_client_factory_test.go` to test only public surfaces.
8. Delete (or unexport to lowercase) `ServerVersion` and `ServerFlavor` on
   `ElasticsearchScopedClient`. Run `make build`; confirm clean.
9. Apply both spec deltas.

## Risks / Trade-offs

- **Risk**: apikey resources upgraded across this change have legacy private-state blobs. The
  shape-discriminator fallback in `apikeyCapabilitiesOfLastRead` handles this, but the path is
  narrow and easy to miss in tests.
  → Mitigation: add a dedicated unit test `TestApikeyCapabilitiesOfLastRead_LegacyVersionBlob`
  that constructs the legacy `{"Version":"8.15.0"}` bytes and asserts the right capability
  booleans come out. Add an integration-style test that round-trips through write→read with
  legacy bytes seeded.
- **Risk**: a missed caller still imports `ServerVersion` or `ServerFlavor`, breaking the build.
  → Mitigation: `rg "\.ServerVersion\(|\.ServerFlavor\(" internal/elasticsearch internal/clients internal/versionutils`
  before step 8. CI catches any miss because the methods will not exist after step 8.
- **Risk**: `versionutils.SkipIfUnsupported` is called by many acceptance tests; behavior must
  not regress.
  → Mitigation: `fetchAcceptanceServerInfo` and `checkSkip` keep their semantics. Only the
  underlying client call shape changes. Existing acceptance unit tests for `checkSkip` (if any)
  cover the matrix.
- **Trade-off**: `AcceptanceServerInfo` is a test-only public function in a production package.
  It is documented as such and lives next to `NewAcceptanceTestingElasticsearchScopedClient`,
  which has the same property. The alternative — a build tag — adds project-wide complexity for
  a single helper.
- **Risk**: adding `EnforceVersionCheck` to the ES client without adding it to `MinVersionClient`
  creates a small asymmetry: the envelope cannot evaluate predicate-style requirements declared
  on models.
  → Mitigation: documented as deliberate; no caller needs it. If a future change requires
  predicate-style requirements, extending the interface is mechanical.
- **Risk**: state-shape migration for apikey is irreversible — once the new blob is written,
  downgrading the provider would fail to parse it.
  → Mitigation: this is the same property all existing private-state migrations in this provider
  have. Documented in the release notes path produced by the implementation step. The legacy
  parser remains in place indefinitely so upgrades from any prior version work.
