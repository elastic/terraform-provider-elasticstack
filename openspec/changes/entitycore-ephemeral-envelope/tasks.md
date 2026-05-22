## 1. Spec

- [ ] 1.1 Keep the delta spec at `openspec/changes/entitycore-ephemeral-envelope/specs/entitycore-ephemeral-envelope/spec.md` aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate entitycore-ephemeral-envelope --type change` (or `make check-openspec`).
- [ ] 1.2 Confirm whether `providerschema` already exposes an ephemeral-namespaced Kibana connection block helper; if not, plan an addition in implementation task 2.2 below.
- [ ] 1.3 On completion of implementation, **sync** delta into `openspec/specs/entitycore-ephemeral-envelope/spec.md` or **archive** the change per project workflow.

## 2. Shared close-state and connection-snapshot infrastructure

- [ ] 2.1 Add `internal/entitycore/ephemeral_close_state.go` with:
  - A package-private reserved private-state key constant for the user close-state slot (e.g. `ephemeralUserStateKey = "entitycore.ephemeral.user_state"`).
  - `encodeUserCloseState[S any](s S) ([]byte, diag.Diagnostics)` — JSON marshal.
  - `decodeUserCloseState[S any](data []byte) (S, diag.Diagnostics)` — JSON unmarshal.
  - `mustBePlainGoCloseState[S any]()` — reflect-walk `S` (handling pointers, embedded structs, slices, maps, arrays) and `panic` with a message of the form `entitycore: ephemeral close state <type> has field <path> of plugin-framework type <PkgPath>/<Name>; Close state must be plain Go types only` when any field's `Type().PkgPath()` is `github.com/hashicorp/terraform-plugin-framework/types`.
- [ ] 2.2 Add `internal/entitycore/ephemeral_connection_snapshot.go` containing:
  - A package-private reserved private-state key constant for the connection slot (e.g. `ephemeralConnectionKey = "entitycore.ephemeral.connection"`).
  - `ephemeralConnectionSnapshot` struct using plain Go types (string, `[]string`, `map[string]string`, `*bool`) for every field of `config.ElasticsearchConnection` (lifted from `internal/elasticsearch/security/api_key/ephemeral_connection.go`).
  - `encodeElasticsearchConnection(ctx, types.List) (string, diag.Diagnostics)` and `decodeElasticsearchConnection(ctx, string) (types.List, diag.Diagnostics)` for the Elasticsearch flavor.
  - Kibana counterparts: `ephemeralKibanaConnectionSnapshot` + encode/decode pair, mirroring the `clientconfig.KibanaConnection` shape.
  - If `providerschema.GetKbEphemeralConnectionBlock()` does not exist, add it as a small companion change inside `internal/schema/ephemeral_connection.go` (or its Kibana counterpart) — pattern identical to the existing Elasticsearch helper.
- [ ] 2.3 Add `internal/entitycore/ephemeral_close_state_test.go`:
  - Positive: a struct with `string`, `int64`, `bool`, `*bool`, `[]string`, `map[string]string`, and an embedded plain-Go struct passes `mustBePlainGoCloseState`.
  - Negative: a struct with any `types.String`, `types.List`, `types.Object`, `types.Bool`, or `jsontypes.Normalized` field — including via embedded struct, slice element, map value, and pointer — panics with a message containing the offending field path and the offending package path.
  - JSON round-trip: encode → decode preserves all plain-Go field types.
- [ ] 2.4 Add `internal/entitycore/ephemeral_connection_snapshot_test.go` covering Elasticsearch and Kibana variants:
  - Round-trip preserves `endpoints`, `username`, `password`, `api_key`, `bearer_token`, `insecure`, `ca_file`, `ca_data`, `cert_file`, `cert_data`, `key_file`, `key_data`, `headers`, and the Elasticsearch-specific `es_client_authentication` field.
  - `insecure = false` survives round-trip (regression test for the bug fixed in the archived api_key change).
  - Null/absent connection encodes to an empty marker and decodes back to a null `types.List`.

## 3. Elasticsearch ephemeral envelope

- [ ] 3.1 Add `internal/entitycore/elasticsearch_ephemeral_envelope.go` with:
  - `ElasticsearchEphemeralModel` interface (`GetElasticsearchConnection() types.List`), satisfied by any struct embedding the existing `ElasticsearchConnectionField`.
  - `OpenRequest[T any]`, `OpenResult[T any, S any]`, `CloseRequest[S any]`, `CloseResponse` structured types.
  - `ElasticsearchEphemeralOpenFunc[T, S]` and `ElasticsearchEphemeralCloseFunc[S]` callback types.
  - `ElasticsearchEphemeralOptions[T, S]{Schema, Open, Close}` with `Close` documented as required.
  - `ElasticsearchEphemeralResource[T, S]` struct embedding `*ResourceBase`-equivalent state (it cannot embed `ResourceBase` directly because the Configure receiver type differs — `ephemeral.ConfigureRequest` vs `resource.ConfigureRequest`; factor a small internal helper or duplicate the ~10 LOC of Configure).
  - `NewElasticsearchEphemeralResource[T ElasticsearchEphemeralModel, S any](name string, opts ElasticsearchEphemeralOptions[T, S]) ephemeral.EphemeralResource`. The constructor: validates `opts.Schema`, `opts.Open`, and `opts.Close` are non-nil (panic with config error if not); calls `mustBePlainGoCloseState[S]()`; returns the constructed value.
  - Methods:
    - `Metadata` — `<provider_type_name>_elasticsearch_<name>`.
    - `Configure` — `clients.ConvertProviderDataToFactory`, error diagnostics on failure.
    - `Schema` — call the user schema factory, copy the `Blocks` map, inject `elasticsearch_connection` via `providerschema.GetEsEphemeralConnectionBlock()`.
    - `Open` — decode `req.Config` into `T`; resolve scoped client from `model.GetElasticsearchConnection()`; call `EnforceVersionRequirements(ctx, client, &model)`; invoke `opts.Open`; on success snapshot the connection (call `encodeElasticsearchConnection`) and the user close-state (call `encodeUserCloseState[S]`) into the reserved private-state slots; call `resp.Result.Set(ctx, &result.Model)`.
    - `Close` — load the connection snapshot and the user close-state from private; decode the connection back to `types.List`; resolve scoped client; unmarshal `S`; invoke `opts.Close` with `CloseRequest[S]{State: state}`.
  - Compile-time interface assertions for `ephemeral.EphemeralResource`, `ephemeral.EphemeralResourceWithConfigure`, `ephemeral.EphemeralResourceWithClose`.
- [ ] 3.2 Add `internal/entitycore/elasticsearch_ephemeral_envelope_test.go` covering:
  - Constructor panics when `S` contains a tfsdk type.
  - Constructor returns a value satisfying all three plugin-framework ephemeral interfaces.
  - `Schema` injects the `elasticsearch_connection` block alongside user-supplied attributes and blocks.
  - `Open` prelude:
    - decode error short-circuits before invoking the user callback.
    - client resolution error short-circuits.
    - version-requirement error short-circuits.
    - user callback diagnostics propagate to `resp.Diagnostics`.
    - on success, both private slots are populated and `Result` contains the user model.
  - `Close` prelude:
    - missing private slots returns cleanly without calling the user callback.
    - present private slots restore connection and call the user callback with the unmarshaled state.
    - user callback diagnostics propagate.
  - Use fake `clients.ProviderClientFactory` and stub callbacks (no real Elasticsearch).

## 4. Kibana ephemeral envelope

- [ ] 4.1 Add `internal/entitycore/kibana_ephemeral_envelope.go` mirroring section 3 but for Kibana: `KibanaEphemeralModel`, callback types parameterized on `*clients.KibanaScopedClient`, `NewKibanaEphemeralResource[T KibanaEphemeralModel, S any]`, `Metadata` composing `<provider_type_name>_kibana_<name>`, `Schema` injecting `kibana_connection`, Open prelude calling `client.GetKibanaClient(...)`.
- [ ] 4.2 Add `internal/entitycore/kibana_ephemeral_envelope_test.go` mirroring section 3.2 for the Kibana flavor.

## 5. Documentation

- [ ] 5.1 Update `internal/entitycore/doc.go` to:
  - Describe the two ephemeral envelopes alongside the existing resource and data source patterns.
  - Document the plain-Go close-state rule (`S` must not contain `terraform-plugin-framework/types`).
  - Document Open-on-plan as a known property of the Terraform ephemeral contract and direct resource authors to repeat the warning in their generated docs.
  - Include a minimal example mirroring the api_key migration.

## 6. api_key migration

- [ ] 6.1 Rewrite `internal/elasticsearch/security/api_key/ephemeral_resource.go` on top of `entitycore.NewElasticsearchEphemeralResource[ephemeralTfModel, ephemeralCloseState]`. Steps:
  - Add the embed: `ephemeralTfModel` embeds `entitycore.ElasticsearchConnectionField` (remove the explicit `ElasticsearchConnection types.List` field if it is present).
  - Define `ephemeralCloseState struct { KeyID string; InvalidateOnClose bool }` (plain Go).
  - Move the schema factory function so it returns `eschema.Schema` without the `elasticsearch_connection` block.
  - Implement the Open callback: branch on type, call existing `CreateAPIKey` / `CreateCrossClusterAPIKey`, populate `model`, return `(OpenResult[ephemeralTfModel, ephemeralCloseState]{Model: model, CloseState: {KeyID, InvalidateOnClose}}, diags)`.
  - Implement the Close callback: when `state.InvalidateOnClose && state.KeyID != ""`, call `elasticsearch.DeleteAPIKey(ctx, client, state.KeyID)`.
  - Replace `NewEphemeralResource()` with the entitycore constructor call.
- [ ] 6.2 Delete `internal/elasticsearch/security/api_key/ephemeral_connection.go` (functionality moved into entitycore).
- [ ] 6.3 Delete the now-unused helpers from `ephemeral_resource.go`: `ephemeralPrivateData`, `ephemeralPrivateState`, `saveEphemeralPrivateData`, `loadEphemeralPrivateData`, `elasticsearchConnectionFromPrivateJSON`, the `deleteAPIKeyFn` package-level var and `elasticsearchClientResolver` interface (replaced by entitycore's client resolution), and the `effectiveAPIKeyType` / `invalidateOnCloseValue` helpers (folded into the new Open callback).
- [ ] 6.4 Update `internal/elasticsearch/security/api_key/ephemeral_resource_test.go`:
  - Keep schema-validator tests (`TestEphemeralSchema*`) unchanged — they exercise the schema factory directly.
  - Replace `TestCloseAPIKeyIfRequested` with tests against the Close callback function in isolation (positive: `InvalidateOnClose=true` + `KeyID="abc"` invokes `DeleteAPIKey`; negative: `InvalidateOnClose=false` or empty `KeyID` does not).
  - Adjust `TestNewEphemeralResourceImplementsInterfaces` to assert against the envelope-returned type.
- [ ] 6.5 Confirm the existing acceptance tests pass unchanged:
  - `TestAccEphemeralResourceSecurityAPIKey`
  - `TestAccEphemeralResourceSecurityAPIKeyInvalidateOnClose`
  - `TestAccEphemeralResourceSecurityAPIKeyWithExpiration`
  - `TestAccEphemeralResourceSecurityAPIKeyCrossCluster`
  - `TestAccEphemeralResourceSecurityAPIKeyExplicitConnection`
  - Run targeted: `TF_ACC=1 go test -v -timeout 30m -run 'TestAccEphemeralResourceSecurityAPIKey' ./internal/elasticsearch/security/api_key/...`.

## 7. Validation

- [ ] 7.1 `make lint`.
- [ ] 7.2 `make build`.
- [ ] 7.3 `go test ./internal/entitycore/...` (unit tests for the new envelopes).
- [ ] 7.4 `go test ./internal/elasticsearch/security/api_key/...` (unit tests for the migrated resource).
- [ ] 7.5 Targeted acceptance tests against a live stack (per `dev-docs/high-level/testing.md`); if no stack is reachable locally, rely on PR CI.
- [ ] 7.6 `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate entitycore-ephemeral-envelope --type change`.
