## 1. Fix GetRole read path to bypass typed Global decode

- [ ] 1.1 In `internal/clients/elasticsearch/security_role.go`, replace `typedClient.Security.GetRole().Name(rolename).Do(ctx)` with a raw `GET /_security/role/<rolename>` request via `typedClient.Transport.Perform(req)`, following the pattern in `internal/clients/elasticsearch/index.go`
- [ ] 1.2 Decode the response body as `map[string]json.RawMessage` to locate the per-role entry by name
- [ ] 1.3 From the per-role raw entry, extract each field independently: decode all fields except `global` using the typed `types.Role` partial struct (or individual field decoders), and decode `global` as `json.RawMessage`
- [ ] 1.4 Construct the result without decoding `global` into `types.Role.Global` (e.g., return typed role fields plus a separate `global` `json.RawMessage`, or introduce a small provider wrapper struct that carries `*types.Role` and `GlobalRaw`).
- [ ] 1.5 Handle non-2xx HTTP responses (read body, return error diagnostic) and body-read/decode errors
- [ ] 1.6 Preserve the existing 404 → `(nil, nil)` behavior for not-found roles

## 2. Loosen write-path global decode in models.go

- [ ] 2.1 In `internal/elasticsearch/security/role/models.go` at line 122, change `var global map[string]map[string]map[string][]string` to `var global map[string]any`
- [ ] 2.2 Verify the downstream marshal path (`role.Global = global`) is compatible — the typed PutRole builder already accepts `map[string]json.RawMessage` via the conversion at `security_role.go:46–50`, so `map[string]any` is safe as the intermediate type

## 3. Validation and build

- [ ] 3.1 Run `make build` and confirm compilation succeeds
- [ ] 3.2 Run `go vet ./internal/clients/elasticsearch/... ./internal/elasticsearch/security/role/...` and resolve any issues
- [ ] 3.3 Run `go test ./internal/elasticsearch/security/role/...` (unit tests) and confirm they pass

## 4. Spec sync

- [ ] 4.1 Verify `make check-openspec` passes after merging the delta spec into the main spec
