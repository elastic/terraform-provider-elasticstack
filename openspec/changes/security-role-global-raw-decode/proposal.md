## Why

`TestAccResourceSecurityRole` fails on Elasticsearch 9.5.0-SNAPSHOT because the server now returns `"data_source": []` as a top-level entry inside the `global` privilege object — an array where the go-elasticsearch typed client expects `map[string]map[string]map[string][]string`. The typed client's `Role.UnmarshalJSON` cannot decode this heterogeneous shape and raises a JSON unmarshal error, breaking every `GetRole` read that returns a role with `global` privileges on ES 9.5+.

The upstream spec bug (elasticsearch-specification#6377) is still open and no released go-elasticsearch fix exists, so bumping the client dependency is not a viable short-term fix. The provider's own schema field (`data.Global` is already `customtypes.JSONWithDefaultsValue[map[string]any]`) is permissive; only the **typed-client read path** in `internal/clients/elasticsearch/security_role.go` is too strict.

## What Changes

Replace the typed `Security.GetRole` call in `GetRole` with a raw HTTP transport call (`typedClient.Transport.Perform`) that decodes the API response as `map[string]json.RawMessage` and extracts the `global` field as opaque JSON. All other fields (applications, cluster, indices, etc.) are decoded using the existing typed structs; only `global` bypasses the typed decoder. PutRole and DeleteRole continue to use the typed API.

This approach mirrors the existing precedent in `internal/clients/elasticsearch/index.go` (date-math index handling) and is forward-compatible with future heterogeneous `global` categories without any additional provider changes.

Additionally, the write-path decode in `models.go` (`toAPIModel`) is loosened from `map[string]map[string]map[string][]string` to `map[string]any` so that future `global` categories with non-uniform shapes do not cause write-side failures either.

## Capabilities

### Modified Capabilities

- `elasticsearch-security-role`: modify REQ-001–REQ-003 — the `GetRole` implementation SHALL bypass the typed client's `Role.Global` field and fetch the `global` object via raw transport, decoding it as `json.RawMessage`. The existing requirement language for typed API usage is narrowed: PutRole and DeleteRole continue to use the typed API; GetRole uses the typed API for all fields except `global`, which is fetched via raw transport.
- `elasticsearch-security-role`: add REQ-039 — the write-path conversion in `toAPIModel` SHALL decode `global` as `map[string]any` (not `map[string]map[string]map[string][]string`) to accommodate heterogeneous per-category shapes returned or configured for ES 9.5+.

## Impact

- `internal/clients/elasticsearch/security_role.go` — replace `Security.GetRole().Do(ctx)` with a raw `/_security/role/<name>` GET via `typedClient.Transport.Perform`; decode non-`global` fields using typed decoders, and carry `global` as `json.RawMessage` through to the model/state mapping without decoding into `types.Role.Global`.
- `internal/elasticsearch/security/role/models.go:122` — change `var global map[string]map[string]map[string][]string` to `var global map[string]any` in `toAPIModel`.
- Acceptance test `TestAccResourceSecurityRole` — existing test covers the round-trip; no new test file required, but the test must pass on ES 9.5+ after this fix.
