## 1. Envelope: WriteRequest.Config field change

> **Note (task 1.2):** `writeInvocation.config` remains `tfsdk.Config` because Create/Update only receive the raw framework handle; `runWrite` decodes via `inv.config.Get(ctx, &configModel)` and passes the resulting `T` as `WriteRequest.Config`. The *typed* value is what write callbacks observe.

> **Note (decode order):** Terraform configuration is decoded after `requireReadFunc` (and after version requirements): if the read callback is nil or config decoding fails, diagnostics match the prior short-circuit behavior without invoking the write callback.

- [x] 1.1 Change `WriteRequest[T].Config` field type from `tfsdk.Config` to `T` in `internal/entitycore/resource_envelope.go`
- [x] 1.2 Change `writeInvocation.config` field type from `tfsdk.Config` to `T` in `internal/entitycore/resource_envelope.go` _(interpreted as above: carrier stays `tfsdk.Config`; `WriteRequest.Config` is `T`)_
- [x] 1.3 Update `runWrite` to decode config into `T` via `inv.config.Get(ctx, &configModel)` and pass `configModel` as `WriteRequest.Config`
- [x] 1.4 Update `resource_envelope_test.go` assertions that access `req.Config.Raw` / `req.Config.Schema` to use struct field access on the decoded model

## 2. security/user: Update write callback

- [x] 2.1 Replace `req.Config.GetAttribute(ctx, path.Root("password_wo"), &passwordWoFromConfig)` in `internal/elasticsearch/security/user/update.go` with direct struct field access `req.Config.PasswordWo` _(applied during Task 1 as a minimal compile fix when the field type changed)_

## 3. index/template: Migrate Create and Update overrides

- [x] 3.1 Remove `func (r *Resource) Create` and `func (r *Resource) Update` method receivers from `internal/elasticsearch/index/template/create.go` and `update.go`
- [x] 3.2 Create a `writeIndexTemplate` `WriteFunc[Model]` in `internal/elasticsearch/index/template/write.go` (or equivalent) that uses `req.Config` as the `priorForRead` seed and contains the `allow_custom_routing` 8.x workaround using `req.Prior` vs `req.Config`
- [x] 3.3 Update `internal/elasticsearch/index/template/resource.go` to wire `writeIndexTemplate` as both `Create` and `Update` in `ElasticsearchResourceOptions` (replacing the placeholders)

## 4. index/templateilmattachment: Migrate Create and Update overrides

- [x] 4.1 Implement `GetVersionRequirements()` on `tfModel` in `internal/elasticsearch/index/templateilmattachment/` returning a single requirement for ES ≥ 8.2.0, satisfying `entitycore.WithVersionRequirements`
- [x] 4.2 Remove `func (r *Resource) Create` and `func (r *Resource) Update` method receivers from `internal/elasticsearch/index/templateilmattachment/create.go` and `update.go`
- [x] 4.3 Create a `writeILMAttachmentCallback` `WriteFunc[tfModel]` that calls the existing `writeILMAttachment` pure function; use `req.Prior == nil` to pass the `isCreate bool` argument
- [x] 4.4 Update `internal/elasticsearch/index/templateilmattachment/resource.go` to wire the new callback as both `Create` and `Update` (replacing the placeholders)

## 5. security/api_key: Migrate Update override

- [ ] 5.1 Remove `func (r *Resource) Update` method receiver from `internal/elasticsearch/security/api_key/update.go`
- [ ] 5.2 Create a `writeAPIKey` `WriteFunc[tfModel]` that branches on `req.Plan.Type` to call `updateCrossClusterAPIKey` or `updateAPIKey`, then calls `readAPIKey` and returns the result model
- [ ] 5.3 Update `internal/elasticsearch/security/api_key/resource.go` to wire the new callback as `Update` in `ElasticsearchResourceOptions` (replacing the placeholder); `Create` remains as a method receiver override

## 6. transform: Migrate Create and Update overrides

- [ ] 6.1 Remove `func (r *transformResource) Create` and `func (r *transformResource) Update` method receivers from `internal/elasticsearch/transform/resource.go`
- [ ] 6.2 Create a `writeTransform` `WriteFunc[tfModel]` that uses `req.Prior == nil` to distinguish Create (Put Transform) from Update (Update Transform), and handles enabled-state delta start/stop by comparing `req.Plan.Enabled` vs `req.Prior.Enabled`
- [ ] 6.3 Update `internal/elasticsearch/transform/resource.go` to wire `writeTransform` as both `Create` and `Update` in `ElasticsearchResourceOptions` (replacing the placeholders)

## 7. Validation

- [ ] 7.1 Run `make build` and confirm no compilation errors
- [ ] 7.2 Run `go test ./internal/entitycore/...` to confirm envelope unit tests pass
- [ ] 7.3 Run `go test ./internal/elasticsearch/index/template/...` to confirm template unit tests pass (if any)
