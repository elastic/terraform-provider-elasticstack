## Why

`TestAccResourceSecurityRole` fails on Elasticsearch 9.5.0-SNAPSHOT because the server now returns `"data_source": []` as a top-level entry inside the `global` privilege object — an array where the go-elasticsearch typed client expects `map[string]map[string]map[string][]string` (the type of `types.Role.Global`). The typed client's `Role.UnmarshalJSON` cannot decode this heterogeneous shape and raises a JSON unmarshal error, breaking every `GetRole` read that returns a role with `global` privileges on ES 9.5+.

The upstream spec bug (elasticsearch-specification#6377) is still open and no released go-elasticsearch fix exists, so bumping the client dependency is not a viable short-term fix. The provider's own schema field (`data.Global` is already `customtypes.JSONWithDefaultsValue[map[string]any]`) is permissive; only the **typed-client read path** in `internal/clients/elasticsearch/security_role.go` is too strict.

## What Changes

Replace the typed `Security.GetRole` call in `GetRole` with a raw HTTP transport call (`typedClient.Transport.Perform`) that decodes the API response as `map[string]json.RawMessage` and extracts the `global` field as opaque JSON. All other fields (applications, cluster, indices, etc.) are decoded using the existing typed structs; only `global` bypasses the typed decoder. The raw `global` JSON is returned to the model layer **out-of-band** (not via `types.Role.Global`, which cannot represent array-typed categories) and consumed directly by `fromAPIModel`.

The write path (`PutRole`/`DeleteRole`) is unchanged — it already marshals `global` through `map[string]json.RawMessage` and is unaffected by the upstream bug for the shapes users configure today.

This approach mirrors the existing precedent in `internal/clients/elasticsearch/index.go` (date-math index handling) and is forward-compatible with future heterogeneous `global` categories without any additional provider changes.

Additionally, the existing `populateGlobalPrivilegesDefaults` normalizer is extended to strip the new server-injected empty `data_source: []` default (mirroring the existing `role: {}` strip) so state matches user intent and `TestAccResourceSecurityRole`'s exact-match assertion holds.

## Capabilities

### Modified Capabilities

- `elasticsearch-security-role`: modify REQ-001–REQ-003 and the "Typed client implementation for security role" requirement — the `GetRole` implementation SHALL bypass the typed client's `Role.Global` field and fetch the `global` object via raw transport, decoding it as `json.RawMessage` and carrying it to the model layer out-of-band (not via `types.Role.Global`). Typed API usage is narrowed: PutRole and DeleteRole continue to use the typed API unchanged; GetRole uses the typed structs for all fields except `global`, which is fetched via raw transport.
- `elasticsearch-security-role`: modify the global-defaults normalization requirement — `populateGlobalPrivilegesDefaults` SHALL strip server-injected empty `global` defaults (including `role: {}` and the new `data_source: []`) so Terraform state matches user intent rather than the raw API blob.

## Impact

- `internal/clients/elasticsearch/security_role.go` — replace `Security.GetRole().Do(ctx)` with a raw `GET /_security/role/<name>` via `typedClient.Transport.Perform`; decode the response, returning the non-`global` fields as a `*types.Role` and the `global` field as a separate `json.RawMessage`. `GetRole`'s return contract changes to carry the raw `global` alongside the typed role.
- `internal/elasticsearch/security/role/models.go` (`fromAPIModel`) and `internal/elasticsearch/security/role/data_source.go` (`fromAPIModel`) — consume the raw `global` JSON directly into `customtypes.JSONWithDefaultsValue` instead of reading `role.Global`.
- `internal/elasticsearch/security/role/global_defaults.go` — extend `populateGlobalPrivilegesDefaults` to strip empty `data_source` (array) in addition to empty `role` (object).
- Acceptance test `TestAccResourceSecurityRole` — existing test covers the round-trip; no new test file required, but the test must pass on ES 9.5+ after this fix. A unit test for the extended default-stripping is added.
