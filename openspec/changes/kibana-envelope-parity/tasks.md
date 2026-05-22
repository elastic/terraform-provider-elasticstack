## 1. Envelope core: new types and constructor

- [ ] 1.1 Add `KibanaWriteRequest[T]`, `KibanaWriteResult[T]`, `KibanaWriteFunc[T]`, and `KibanaPostReadFunc[T]` types to `kibana_resource_envelope.go`
- [ ] 1.2 Add `KibanaResourceOptions[T]` struct with `Schema`, `Read`, `Delete`, `Create`, `Update`, and `PostRead` fields
- [ ] 1.3 Update `KibanaResource[T]` struct: replace `createFunc KibanaCreateFunc[T]` and `updateFunc KibanaUpdateFunc[T]` with `createFunc KibanaWriteFunc[T]`, `updateFunc KibanaWriteFunc[T]`, and `postReadFunc KibanaPostReadFunc[T]`
- [ ] 1.4 Update `NewKibanaResource` to accept `KibanaResourceOptions[T]` instead of positional callback parameters
- [ ] 1.5 Rename `PlaceholderKibanaWriteCallbacks[T]()` to `PlaceholderKibanaWriteCallback[T]()` and change return type from `(KibanaCreateFunc[T], KibanaUpdateFunc[T])` to `KibanaWriteFunc[T]`
- [ ] 1.6 Remove the `KibanaCreateFunc[T]` and `KibanaUpdateFunc[T]` type aliases entirely

## 2. Envelope core: runKibanaWrite and read-after-write

- [ ] 2.1 Add `kibanaWriteInvocation[T]` struct (plan, priorState, config, outState, privateState, isUpdate) mirroring `writeInvocation[T]` in the ES envelope
- [ ] 2.2 Implement `runKibanaWrite`: nil callback check → decode plan → decode prior (if update) → spaceID validation (incl. KibanaUnscopedSpace) → client resolution → version requirements → decode config → invoke write callback → resolve read identity from written model → call readFunc → not-found error or state.Set → PostRead
- [ ] 2.3 Update `Create` to delegate to `runKibanaWrite` (removing existing direct callback invocation)
- [ ] 2.4 Update `Update` to delegate to `runKibanaWrite` (removing existing direct callback invocation)
- [ ] 2.5 Update `Read` to invoke `postReadFunc` after a successful `resp.State.Set`, skipping on not-found / readFunc error / state-set error

## 3. Envelope tests: new coverage

- [ ] 3.1 Update existing test helpers: `testKibanaCreateFuncFound` and `testKibanaUpdateFuncFound` → single `testKibanaWriteFuncFound` matching `KibanaWriteFunc[T]` signature; update `defaultTestKibanaResourceOptions()` helper
- [ ] 3.2 Add `TestNewKibanaResource_Create_readAfterWriteHappyPath`: verify state is set from readFunc result, not write callback result
- [ ] 3.3 Add `TestNewKibanaResource_Update_readAfterWriteHappyPath`: same as above for Update path
- [ ] 3.4 Add `TestNewKibanaResource_Create_notFoundAfterWrite`: readFunc returns `found == false`; expect error diagnostic, state not mutated
- [ ] 3.5 Add `TestNewKibanaResource_Update_notFoundAfterWrite`: same for Update
- [ ] 3.6 Add `TestNewKibanaResource_Create_readFuncErrorAfterWrite`: readFunc returns errors; expect those diagnostics, state not mutated
- [ ] 3.7 Add `TestNewKibanaResource_Update_readFuncErrorAfterWrite`: same for Update
- [ ] 3.8 Add `TestNewKibanaResource_Create_callbackReceivesNilPriorAndConfig`: assert `req.Prior == nil` and `req.Config` decoded correctly for Create
- [ ] 3.9 Add `TestNewKibanaResource_Update_callbackReceivesNonNilPriorAndConfig`: assert `req.Prior != nil` and `req.Config` decoded correctly for Update
- [ ] 3.10 Add `TestNewKibanaResource_SingleWriteFuncServesCreateAndUpdate`: one `KibanaWriteFunc[T]` wired to both; verify `sawCreate` and `sawUpdate` via `req.Prior == nil` check
- [ ] 3.11 Add `TestNewKibanaResource_Read_invokesPostReadAfterSuccessfulStateSet`
- [ ] 3.12 Add `TestNewKibanaResource_Read_skipsPostReadWhenNotFound`
- [ ] 3.13 Add `TestNewKibanaResource_Read_skipsPostReadWhenReadFuncError`
- [ ] 3.14 Add `TestNewKibanaResource_Read_skipsPostReadWhenStateSetFails`
- [ ] 3.15 Add `TestNewKibanaResource_Read_postReadReceivesFrameworkPrivateHandle`
- [ ] 3.16 Add `TestNewKibanaResource_Create_invokesPostReadAfterReadAfterWrite`
- [ ] 3.17 Add `TestNewKibanaResource_Update_invokesPostReadAfterReadAfterWrite`
- [ ] 3.18 Update `TestNewKibanaResource_Create_placeholderCallbackError` to use new single-return `PlaceholderKibanaWriteCallback`

## 4. Migrate concrete resources

- [ ] 4.1 `fleet/proxy`: update `createProxy` and `updateProxy` signatures to `KibanaWriteFunc[proxyModel]`; update `resource.go` to use `KibanaResourceOptions`
- [ ] 4.2 `kibana/streams`: update `createStream` and `updateStream` signatures; update `resource.go`
- [ ] 4.3 `kibana/maintenance_window`: update `createMaintenanceWindow` signature and remove manual read-after-write (lines 54–68 of `create.go`); update `updateMaintenanceWindow` signature; update `resource.go`
- [ ] 4.4 `kibana/spaces`: update `createSpace` and `updateSpace` signatures; update `resource.go`
- [ ] 4.5 `kibana/security_role`: update `createRole` and `updateRole` signatures; update `resource.go`
- [ ] 4.6 `fleet/agentdownloadsource`: update `PlaceholderKibanaWriteCallback` call site in `resource.go`

## 5. Validate and verify

- [ ] 5.1 Run `make build` to verify the project compiles cleanly with no remaining references to the old types
- [ ] 5.2 Run unit tests: `go test ./internal/entitycore/...` — all existing and new tests pass
- [ ] 5.3 Run `make check-lint` to verify no lint regressions
