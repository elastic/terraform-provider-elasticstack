## Context

`WriteRequest[T]` is the struct passed to every `WriteFunc[T]` callback in the Elasticsearch resource envelope. It carries `Plan T`, `Prior *T`, `Config tfsdk.Config`, and `WriteID string`. `Plan` and `Prior` are already decoded into the generic model type `T` by the envelope before the callback is invoked; `Config` is the exception — it is passed as the raw `tfsdk.Config` SDK type, requiring write callbacks to call `req.Config.Get(ctx, &model)` or `req.Config.GetAttribute(...)` themselves.

Four resources (`index/template`, `index/templateilmattachment`, `security/api_key` update-only, and `transform`) override the envelope's `Create` and/or `Update` method receivers rather than using pure `WriteFunc` callbacks. In three of those four cases the override exists partially or wholly because the resources need the config-as-model pattern that the current `Config tfsdk.Config` field doesn't provide naturally. The `security/api_key` Create override is explicitly out of scope — it is intentionally outside the envelope write path due to create-time sensitive field population.

## Goals / Non-Goals

**Goals:**
- Change `WriteRequest[T].Config` from `tfsdk.Config` to `T`, decoded by the envelope before the callback is invoked
- Migrate `index/template`, `index/templateilmattachment`, `security/api_key` (Update only), and `transform` (Create + Update) from method receiver overrides to pure `WriteFunc` callbacks
- Update the single existing consumer of `WriteRequest.Config` as `tfsdk.Config` (`security/user/update.go`) to use the decoded struct field instead

**Non-Goals:**
- Migrating `security/api_key` Create — the spec explicitly requires this to remain a concrete method receiver override
- Migrating `index/index` (Create, Update, Read overrides) — date-math concrete name identity is a separate, higher-complexity change
- Migrating `ml/jobstate`, `ml/datafeed_state`, or `transform` Create — only what is enumerated above

## Decisions

### Decision: Decode `Config` into `T` in the envelope, matching `Plan` and `Prior`

The envelope already performs `inv.plan.Get(ctx, &planModel)` and `inv.priorState.Get(ctx, &priorModel)` in `runWrite`. Adding a parallel `inv.config.Get(ctx, &configModel)` before constructing `WriteRequest` is consistent and eliminates the need for callbacks to operate on raw SDK types.

**Alternative considered: Keep `tfsdk.Config` and add a separate helper method.** Rejected — it still leaks the SDK type into the callback signature and doesn't solve the inconsistency.

**Write-only attributes:** `tfsdk.Config.Get(ctx, &T{})` does populate write-only fields (the framework only nullifies write-only values in *state*, not config). The sole write-only attribute in this provider is `security_user.password_wo`. After the change, `security/user/update.go` reads `req.Config.PasswordWo` directly — tested to confirm the field is populated.

### Decision: `templateilmattachment` version gate moves to `WithVersionRequirements` on the model

The resource currently calls `client.ServerVersion()` directly in Create/Update overrides with a comment noting "the envelope does not enforce version requirements for placeholder write callbacks". After this change, the resource will use real `WriteFunc` callbacks, and `WithVersionRequirements` is automatically enforced by the envelope for all write paths. Implementing `GetVersionRequirements()` on the model is the canonical pattern (already used by `componenttemplate` and `template`).

### Decision: `index/template` write callbacks use `req.Config` (decoded `T`) as the read-after-write seed

The template resource passes `config` (not `plan`) as the `priorForRead` argument to `readIndexTemplate` during Create/Update. This is needed because `plan` may carry `Unknown` placeholders in computed set elements that differ from the non-refreshed planning value. With `Config T` decoded by the envelope, `req.Config` is already the decoded config model — `priorForRead := req.Config` replaces the manual decode.

The `allow_custom_routing` 8.x workaround also lives in the Update override. It compares `prior.DataStream` against `config.DataStream` — both available as decoded `T` values in the callback (`req.Prior` and `req.Config` respectively).

### Decision: `security/api_key` Update moves to `WriteFunc`; Create stays as a method override

The Update override branches on `plan.Type` to call either `updateAPIKey` or `updateCrossClusterAPIKey`. These are already extracted as helper functions taking `planModel`. Wrapping them in a `WriteFunc` requires only removing the method receiver, accessing the client via `req.Plan.GetElasticsearchConnection()`, and passing `req.Plan` instead of a local `planModel`.

The Create override intentionally bypasses the envelope write path (per spec REQ-030) to set `api_key`, `encoded`, `key_id`, and `id` from the create response before the read-after-write step. This pattern cannot be expressed as a `WriteFunc` without adding new result fields to `WriteResult`. It stays as a method receiver.

### Decision: `transform` Create and Update move to a shared `WriteFunc`

The transform overrides call `createTransform` / `updateTransform` and manage start/stop via the `enabled` field delta. `req.Prior == nil` distinguishes Create from Update in a shared callback. The enabled-state delta is `req.Plan.Enabled != req.Prior.Enabled` (nil-guarded), matching what the current Update override does with `plan.Enabled` vs `state.Enabled`.

## Risks / Trade-offs

- **`Config` decode adds one allocation per write.** The decoded `T` is a small struct; this is negligible.
- **Decoding errors in config surface as write errors.** If `inv.config.Get(ctx, &configModel)` fails, `runWrite` returns diagnostics before invoking the callback. This matches existing plan-decode behaviour and is correct — a malformed config should never reach a write callback.
- **`api_key` Create remains a method override** alongside a `WriteFunc` for Update and a `Placeholder` for Create in `ElasticsearchResourceOptions`. The `PlaceholderElasticsearchWriteCallback` is already used for this pattern and the spec documents the intent explicitly.

## Migration Plan

1. Change `WriteRequest[T].Config` field type and `writeInvocation.config` field; update `runWrite` to decode config into `T`; update tests.
2. Update `security/user/update.go` to use `req.Config.PasswordWo` directly.
3. Migrate `index/template` Create and Update overrides to `WriteFunc` callbacks.
4. Migrate `index/templateilmattachment` Create and Update overrides: implement `WithVersionRequirements` on the model, write a `WriteFunc`.
5. Migrate `security/api_key` Update override to a `WriteFunc`; keep Create as a method receiver.
6. Migrate `transform` Create and Update overrides to a shared `WriteFunc`.
7. `make build` to confirm no compilation errors.

## Open Questions

_(none)_
