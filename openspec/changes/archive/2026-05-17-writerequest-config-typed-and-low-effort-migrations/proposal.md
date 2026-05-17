## Why

`WriteRequest[T].Config` is typed as `tfsdk.Config` — a raw Terraform SDK type — forcing write callbacks to call `req.Config.Get(ctx, &model)` or `req.Config.GetAttribute(...)` directly. This leaks framework plumbing into business logic and is inconsistent with how `Plan` and `Prior` are already decoded into `T`. Four resources currently override the envelope's `Create`/`Update` method receivers instead of using pure write callbacks, partly because the current `Config` field doesn't offer the decoded model they need.

## What Changes

- **`WriteRequest[T].Config` changes from `tfsdk.Config` to `T`**: the envelope decodes the config into `T` before passing it to write callbacks, matching the pattern already used for `Plan` and `Prior`.
- **`index/template` Create and Update overrides removed**: migrated to pure `WriteFunc` callbacks that use `req.Config` as the decoded model for read-after-write seeding and the `allow_custom_routing` 8.x workaround.
- **`index/templateilmattachment` Create and Update overrides removed**: version gating migrated to `WithVersionRequirements` on the model; write logic moved into a `WriteFunc`.
- **`security/api_key` Update override removed**: `cross_cluster` vs regular key branching moved into a `WriteFunc` that inspects `req.Plan.Type`.
- **`transform` Create and Update overrides removed**: start/stop-on-enabled-change logic moved into a `WriteFunc` using `req.Prior` to detect the delta.
- **`security/user` write callback updated**: `req.Config.GetAttribute(..., "password_wo", ...)` replaced with direct struct field access `req.Config.PasswordWo`.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `entitycore-resource-envelope`: `WriteRequest[T].Config` field type changes from `tfsdk.Config` to `T`; the envelope decodes config before constructing `WriteRequest`.
- `elasticsearch-index-template`: Create/Update no longer override the envelope; write callbacks receive the decoded config model directly.
- `elasticsearch-index-template-ilm-attachment`: Create/Update no longer override the envelope; version requirement enforced via `WithVersionRequirements`.
- `elasticsearch-security-api-key`: Update no longer overrides the envelope; branching on key type handled inside the write callback.
- `elasticsearch-transform`: Create/Update no longer override the envelope; enabled-state delta and start/stop calls handled inside the write callback.

## Impact

- `internal/entitycore/resource_envelope.go` — `WriteRequest[T]`, `writeInvocation`, `runWrite`
- `internal/entitycore/resource_envelope_test.go` — test assertions on `req.Config`
- `internal/elasticsearch/index/template/` — `create.go`, `update.go`, `resource.go`
- `internal/elasticsearch/index/templateilmattachment/` — `create.go`, `update.go`, `resource.go`, model
- `internal/elasticsearch/security/api_key/` — `update.go`, `resource.go`
- `internal/elasticsearch/transform/` — `resource.go`
- `internal/elasticsearch/security/user/update.go` — `req.Config.PasswordWo` field access
