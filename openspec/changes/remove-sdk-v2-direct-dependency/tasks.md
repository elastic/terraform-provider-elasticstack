## 1. Add internal logging helpers

- [x] 1.1 Create `internal/debugutils/logging.go` with `IsDebugOrHigher()` implementation
- [x] 1.2 Add `IsSensitiveInSchema() bool` shared helper to `internal/debugutils/logging.go`
- [x] 1.3 Run `go test ./internal/debugutils/...` to ensure new file compiles

## 2. Replace SDK v2 `logging.IsDebugOrHigher` imports

- [x] 2.1 Update `internal/clients/config/elasticsearch.go` to use `debugutils.IsDebugOrHigher()`
- [x] 2.2 Update `internal/clients/fleet/client.go` to use `debugutils.IsDebugOrHigher()`
- [x] 2.3 Update `internal/clients/kibanaoapi/client.go` to use `debugutils.IsDebugOrHigher()`
- [x] 2.4 Update `internal/fleet/integration_policy/schema.go` to use `debugutils.IsSensitiveInSchema()`
- [x] 2.5 Update `internal/fleet/integration_policy/schema_v2.go` to use `debugutils.IsSensitiveInSchema()`

## 3. Remove dead code

- [x] 3.1 Delete `internal/utils/typeutils/schema.go`
- [x] 3.2 Delete `internal/utils/typeutils/schema_test.go`
- [x] 3.3 Verify `go build ./...` still succeeds

## 4. Fix stray test import

- [x] 4.1 Update `internal/elasticsearch/index/templateilmattachment/acc_test.go` to import `github.com/hashicorp/terraform-plugin-testing/helper/acctest` instead of SDK v2 version
- [x] 4.2 Verify the test file compiles (`go test -c ./internal/elasticsearch/index/templateilmattachment/...`)

## 5. Clean up go.mod

- [ ] 5.1 Remove `github.com/hashicorp/terraform-plugin-sdk/v2` from the `require` block in `go.mod`
- [ ] 5.2 Run `go mod tidy`
- [ ] 5.3 Verify `go mod graph` no longer shows a direct edge to `terraform-plugin-sdk/v2`
- [ ] 5.4 Verify `go mod why github.com/hashicorp/terraform-plugin-sdk/v2` reports it as an indirect dependency only

## 6. Build and validate

- [ ] 6.1 Run `make build` to confirm the provider builds cleanly
- [ ] 6.2 Run `go vet ./...` to catch any static analysis issues
- [ ] 6.3 Run `go test ./internal/...` to confirm unit tests pass
