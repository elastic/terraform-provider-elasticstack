## Why

`ElasticsearchScopedClient` exposes `ServerVersion()` and `ServerFlavor()` as public methods.
Eight callers in `internal/elasticsearch/` read them and most do raw `serverVersion.LessThan(...)`
comparisons with no serverless awareness — the same hazard that the archived
[`fix-serverless-version-gating`](../archive/2026-05-16-fix-serverless-version-gating) change had
to remediate after the fact for index templates. Every new Elasticsearch resource is one careless
`LessThan` away from rejecting valid serverless configuration.

The serverless-safe primitives exist (`EnforceMinVersion`, `entitycore.WithVersionRequirements`),
but they are easy to bypass. Forcing all version-gated decisions through serverless-aware APIs —
and removing the raw accessors from the public surface — closes the hazard at the type level. This
change is the Elasticsearch sibling to `kibana-scoped-client-unexport-version-accessors`. They
are split so each lands independently green; the ES side is the larger and more involved of the
two because `internal/elasticsearch/security/apikey` persists the cluster version into Terraform
private state and needs a state-shape migration.

## What Changes

- **NEW**: Add `IsServerless(ctx) (bool, fwdiag.Diagnostics)` on `ElasticsearchScopedClient`.
  Two real callers ask "am I on serverless?" for non-version reasons (`index/index/create`
  branches on flavor for stateful-only knobs; `versionutils.CheckIfNotServerless` is the
  acceptance-test skip predicate). Both currently do `flavor == clients.ServerlessFlavor`
  comparisons against the raw string; the predicate makes the question first-class.
- **NEW**: Add `EnforceVersionCheck(ctx, check func(*version.Version) bool) (bool, fwdiag.Diagnostics)`
  to `ElasticsearchScopedClient`. Mirrors the Kibana client; serverless short-circuits to `true`.
  Required for `apikey` plan-modifiers and any predicate-style gate (e.g. range checks).
- **BREAKING (internal)**: Remove `ElasticsearchScopedClient.ServerVersion(ctx)` and
  `ElasticsearchScopedClient.ServerFlavor(ctx)` from the public surface. They become package-private
  helpers used only by `EnforceMinVersion`, `EnforceVersionCheck`, and `IsServerless`.
- Migrate all eight `internal/elasticsearch/` callers off `ServerVersion`/`ServerFlavor`. Three
  patterns:
  - **Pure gate** (1 caller — `security/role/update` description and remote_indices fields):
    adopt `entitycore.WithVersionRequirements` on the role model. The Elasticsearch resource
    envelope already enforces requirements automatically during Create, Read, and Update.
  - **Feature toggle** (2 callers — `index/index/create` via `ServerFlavor`, `transform/write`
    via `isSettingAllowed`): replace with `IsServerless` (for the index resource) and
    `EnforceMinVersion` per setting (for the transform).
  - **Version-threaded / capability persistence** (4 callers — `security/apikey/{runtime_validation,
    models,resource,resource/schema}`): introduce an `apikeyCapabilities` struct (per-feature
    booleans, modelled on `agentpolicy.features`) resolved via `EnforceMinVersion` /
    `EnforceVersionCheck`. The capability bits replace the persisted cluster-version string in
    Terraform private state.
- **STATE MIGRATION**: `internal/elasticsearch/security/apikey/resource` persists `cluster-version`
  into Terraform private state and reads it back during `RequiresReplaceIf`. The persisted blob
  becomes a JSON `{ supportsUpdate: bool, supportsRoleDescriptors: bool, supportsRestriction: bool }`
  capability record. Read paths fall back to the legacy version string when present (parsing
  `clusterVersionPrivateData{Version: string}`), evaluating it against the same constants once,
  then rewriting the slot to the new shape on the next post-read.
- Update `internal/versionutils/testutils.go`:
  - Acceptance-test `fetchAcceptanceServerInfo` keeps needing the version itself (for
    `SkipIfUnsupported`'s "version below required minimum" path). It moves to a new
    `serverInfoForAcceptance(ctx)` helper exported from `internal/clients` that lives behind a
    build tag or test-only file (`*_acctest.go` / `acceptance_testing.go`) so production code
    cannot import it.
  - `CheckIfNotServerless` switches to `client.IsServerless(ctx)` and inverts the boolean.
- Update tests in `internal/clients/` (`elasticsearch_scoped_client_test.go`,
  `provider_client_factory_test.go`) to exercise the public surface only (`EnforceMinVersion`,
  `EnforceVersionCheck`, `IsServerless`).
- Update specs:
  - `provider-client-factory`: flip the Elasticsearch scoped-client contract from
    `ServerVersion`/`ServerFlavor` to `EnforceMinVersion`/`EnforceVersionCheck`/`IsServerless` +
    `WithVersionRequirements`.
  - `elasticsearch-client-pf-diagnostics`: update the `ElasticsearchScopedClient` requirement
    method enumeration to drop `ServerVersion`/`ServerFlavor` and add `EnforceVersionCheck` /
    `IsServerless`.

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- **`provider-client-factory`**: Replace the Elasticsearch-side scoped-client behavior. The
  contract names `EnforceMinVersion()`, `EnforceVersionCheck()`, `IsServerless()`, and
  `entitycore.WithVersionRequirements` enforcement as the only supported version- and
  flavor-gating surfaces, and forbids public `ServerVersion`/`ServerFlavor` accessors.
- **`elasticsearch-client-pf-diagnostics`**: Update the "ElasticsearchScopedClient methods return
  Plugin Framework diagnostics" requirement to enumerate the new public surface
  (`EnforceMinVersion`, `EnforceVersionCheck`, `IsServerless`) instead of the removed accessors.

## Impact

- `internal/clients/elasticsearch_scoped_client.go` — add `IsServerless`, add
  `EnforceVersionCheck`; remove `ServerVersion`, `ServerFlavor` (or unexport to
  `serverVersion`/`serverFlavor` package-private)
- `internal/clients/elasticsearch_scoped_client_test.go` — rewrite to test public surface
- `internal/clients/provider_client_factory_test.go` — update `ServerFlavor`-via-factory test
- `internal/clients/acceptance_testing.go` (or similar test-only file) — new
  `AcceptanceServerInfo(ctx) (*version.Version, bool, error)` exposing the version + serverless
  flag for the acceptance-test skip flow
- `internal/versionutils/testutils.go` — refactor `fetchAcceptanceServerInfo`,
  `CheckIfNotServerless` to use the new APIs
- `internal/elasticsearch/security/role/{update,models}.go` — add `WithVersionRequirements`;
  remove inline checks
- `internal/elasticsearch/security/apikey/runtime_validation.go` — capability lookup instead of
  version comparison
- `internal/elasticsearch/security/apikey/models.go` — capability flag drives
  `PopulateRoleDescriptorsDefaults` selection
- `internal/elasticsearch/security/apikey/resource/{resource,schema}.go` — capability struct
  persisted into private state; `RequiresReplaceIf` reads `supportsUpdate` boolean; backward
  compatibility for the legacy version-string blob
- `internal/elasticsearch/transform/write.go`, `transform/version_gating.go` —
  `isSettingAllowed` becomes a per-setting `EnforceMinVersion` call (or a precomputed feature
  struct)
- `internal/elasticsearch/index/index/create.go`, `index/models.go` — `ServerFlavor` replaced
  by `IsServerless`; `toPutIndexParams(serverFlavor string)` becomes
  `toPutIndexParams(isServerless bool)`
- `internal/elasticsearch/security/apikey/resource/acc_test.go` — `LessThan/GreaterThanOrEqual`
  range check migrates to `client.EnforceVersionCheck(ctx, func)` (or to an acceptance-test
  helper)
- `openspec/specs/provider-client-factory/spec.md` — update Elasticsearch scoped-client contract
- `openspec/specs/elasticsearch-client-pf-diagnostics/spec.md` — update
  `ElasticsearchScopedClient` method enumeration
