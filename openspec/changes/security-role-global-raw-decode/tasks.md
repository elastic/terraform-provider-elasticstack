## 1. Fix GetRole read path to bypass typed Global decode

- [x] 1.1 In `internal/clients/elasticsearch/security_role.go`, replace `typedClient.Security.GetRole().Name(rolename).Do(ctx)` with a raw `GET /_security/role/<rolename>` request via `typedClient.Transport.Perform(req)`, following the pattern in `internal/clients/elasticsearch/index.go`
- [x] 1.2 Decode the response body as `map[string]json.RawMessage` to locate the per-role entry by name
- [x] 1.3 From the per-role raw entry, extract each field independently: decode all fields except `global` into the typed `types.Role` struct (or individual field decoders), and decode `global` as `json.RawMessage`
- [x] 1.4 Change `GetRole`'s return contract to carry the raw `global` JSON **alongside** the typed `*types.Role` (e.g. a result struct holding both). Do **not** assign `global` to `types.Role.Global` — that field is `map[string]map[string]map[string][]string` and cannot represent array-typed categories like `data_source: []`
- [x] 1.5 Handle non-2xx HTTP responses (read body, return error diagnostic) and body-read/decode errors
- [x] 1.6 Preserve the existing 404 → `(nil, nil)` behavior for not-found roles

## 2. Consume raw global in both fromAPIModel variants

- [x] 2.1 In `internal/elasticsearch/security/role/models.go` `fromAPIModel`, consume the raw `global` JSON returned by `GetRole` directly into `customtypes.JSONWithDefaultsValue` (via `populateGlobalPrivilegesDefaults`) instead of reading `role.Global` / marshaling it
- [x] 2.2 In `internal/elasticsearch/security/role/data_source.go` `fromAPIModel`, apply the same change so the data source read path stays in lockstep with the resource
- [x] 2.3 Verify `PutRole` (`internal/clients/elasticsearch/security_role.go`) is unchanged — it marshals `role.Global` to `map[string]json.RawMessage` and is unaffected for the shapes users configure today

## 3. Extend default-stripping for data_source

- [x] 3.1 In `internal/elasticsearch/security/role/global_defaults.go`, extend `populateGlobalPrivilegesDefaults` to strip `data_source` when it is an empty array (`[]`), mirroring the existing `role: {}` empty-object strip
- [x] 3.2 Generalize the strip to "strip server-injected empty `global` defaults" so future empty-array/empty-object categories don't cause perpetual diffs
- [x] 3.3 Add a unit test covering `data_source: []` stripping (and a mixed case where `data_source` is non-empty and is preserved)

## 4. Validation and build

- [x] 4.1 Run `make build` and confirm compilation succeeds
- [x] 4.2 Run `go vet ./internal/clients/elasticsearch/... ./internal/elasticsearch/security/role/...` and resolve any issues
- [x] 4.3 Run `go test ./internal/elasticsearch/security/role/...` (unit tests) and confirm they pass
- [ ] 4.4 Run `TestAccResourceSecurityRole` against a 9.5.0-SNAPSHOT stack and confirm the `global` state assertion holds (no `data_source`, no `role`)

## 5. Spec sync

- [x] 5.1 Verify `make check-openspec` passes after merging the delta spec into the main spec
