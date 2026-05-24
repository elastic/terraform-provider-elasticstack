## Why

`KibanaScopedClient` exposes `ServerVersion()` and `ServerFlavor()` as public methods on the typed
client surface. Twelve callers in `internal/kibana/` reach for `ServerVersion()` and do raw
`serverVersion.LessThan(...)` comparisons with no serverless awareness — the exact pattern that the
recently-archived [`fix-serverless-version-gating`](../archive/2026-05-16-fix-serverless-version-gating)
change had to remediate post-hoc. Every new Kibana resource is one careless `LessThan` away from
the same bug.

The serverless-safe primitives (`EnforceMinVersion`, `EnforceVersionCheck`,
`entitycore.WithVersionRequirements`) already exist on the client. Forcing version-gated decisions
through those primitives — and removing the unsafe escape hatch from the public surface — closes the
hazard at the type level: a future caller cannot accidentally break serverless because the API will
not compile.

## What Changes

- **BREAKING (internal)**: Remove `KibanaScopedClient.ServerVersion(ctx)` and
  `KibanaScopedClient.ServerFlavor(ctx)` from the public surface. The underlying
  `getServerStatusRaw` plumbing stays package-private. Flavor is no longer consumable outside
  `internal/clients`.
- Migrate all 12 Kibana callers off `ServerVersion`. Three patterns:
  - **Pure gate** (8 callers — `security_enable_rule {read,update,delete}`, `prebuilt_rules
    {read,update}`, `slo {create,update}`, `connectors/create`): adopt
    `entitycore.WithVersionRequirements` on the entity model. The Kibana resource envelope
    already enforces requirements automatically during Create, Read, and Update.
  - **Feature toggle** (2 callers — `synthetics/parameter/delete`,
    `agentbuilderagent/data_source`): replace `LessThan` with `client.EnforceMinVersion`, using
    the returned boolean to select the API path. Serverless naturally selects the newer path.
  - **Version-threaded builder** (2 callers — `alertingrule/{create,update}`): introduce a
    `kibanaCapabilities` struct (per-feature booleans, à la `agentpolicy.features`) resolved up
    front via `EnforceMinVersion`, then passed into `toAPIModel`. `toAPIModel` no longer takes a
    `*version.Version`.
- Update `internal/clients/kibana_scoped_client_test.go` and the `ServerFlavor` factory test in
  `provider_client_factory_test.go` to exercise the public surface (`EnforceMinVersion`,
  `EnforceVersionCheck`) rather than the raw accessors.
- Update the `provider-client-factory` spec to flip the Kibana scoped-client contract from
  "exposes `ServerVersion()`/`ServerFlavor()`" to "exposes `EnforceMinVersion()`,
  `EnforceVersionCheck()`, and supports `entitycore.WithVersionRequirements`".

The companion change `elasticsearch-scoped-client-unexport-version-accessors` mirrors this work
on `ElasticsearchScopedClient`. They are split so each lands independently green.

## Capabilities

### New Capabilities

*(none)*

### Modified Capabilities

- **`provider-client-factory`**: Replace the "Scoped client supports version gating" scenario on
  the Kibana scoped client contract. The new contract names `EnforceMinVersion()`,
  `EnforceVersionCheck()`, and `entitycore.WithVersionRequirements` enforcement as the only
  supported version-gating surfaces, and explicitly forbids public `ServerVersion`/`ServerFlavor`
  accessors on the Kibana scoped client.
- **`elasticsearch-client-pf-diagnostics`**: The "KibanaScopedClient methods return Plugin
  Framework diagnostics" requirement is updated so that its method enumeration covers
  `EnforceMinVersion` and `EnforceVersionCheck` (not the removed `ServerVersion` and
  `ServerFlavor`). The scenario about `ServerVersion` passing PF diagnostics through is replaced
  by an equivalent scenario for `EnforceVersionCheck`.

## Impact

- `internal/clients/kibana_scoped_client.go` — remove `ServerVersion`, `ServerFlavor`
- `internal/clients/kibana_scoped_client_test.go` — rewrite to test `EnforceMinVersion` /
  `EnforceVersionCheck` instead of raw accessors
- `internal/clients/provider_client_factory_test.go` — replace
  `TestKibanaScopedClient_ServerFlavor_ViaFactory` with an `EnforceMinVersion`/serverless
  short-circuit equivalent
- `internal/kibana/security_enable_rule/{read,update,delete}.go` — drop `ServerVersion` call;
  add `WithVersionRequirements` on the model
- `internal/kibana/prebuilt_rules/{read,update}.go` — same
- `internal/kibana/slo/{create,update}.go` — same (two requirements: prevent_initial_backfill
  and data_view_id)
- `internal/kibana/connectors/create.go` — same (conditional on `connector_id`)
- `internal/kibana/synthetics/parameter/delete.go` — replace `LessThan` with `EnforceMinVersion`
- `internal/kibana/agentbuilderagent/data_source.go` — same, drive `supportsAdvancedConfig`
- `internal/kibana/alertingrule/{create,update}.go` — introduce `kibanaCapabilities` struct,
  pass to `toAPIModel`
- `internal/kibana/alertingrule/models*.go` — update `toAPIModel` signature, internal callers
- `openspec/specs/provider-client-factory/spec.md` — update Kibana scoped-client contract
