## Context

`elasticstack_kibana_security_enable_rule` embeds `entitycore.KibanaResource[enableRuleModel]` but bypasses it on the write path. The wrapper struct defines its own `Create` and `Update` methods (in `create.go` and `update.go`) that shadow the envelope's promoted methods, invoking a shared `upsert` helper directly. As a result those operations do not flow through the envelope's client resolution, version-enforcement, or read-after-write paths. The resource also passes `PlaceholderKibanaWriteCallback` for both `Create` and `Update` in `KibanaResourceOptions`, so the envelope callbacks are never reached.

In addition, `deleteSecurityEnableRule` calls `entitycore.EnforceVersionRequirements` directly. Because `KibanaResource` inherits the delete path from `baseResourceEnvelope`, which already calls `EnforceVersionRequirements` before invoking the delete callback, this is a redundant double-check.

See `proposal.md – Why` for motivation.

## Goals / Non-Goals

**Goals:**
- Route Create and Update through the envelope by replacing the wrapper-struct method overrides with a `writeSecurityEnableRule` callback that satisfies `KibanaWriteFunc[enableRuleModel]`.
- Wire `writeSecurityEnableRule` into `KibanaResourceOptions` for both `Create` and `Update`, replacing `PlaceholderKibanaWriteCallback`.
- Remove the redundant `EnforceVersionRequirements` call from `deleteSecurityEnableRule`.

**Non-Goals:**
- Changing the Terraform schema, resource ID format, or any user-visible API semantics.
- Modifying acceptance tests or unit tests (no schema or behavior changes).
- Migrating other resources; scope is `security_enable_rule` only.

## Decisions

### Single shared write callback for Create and Update

**Decision:** Implement one `writeSecurityEnableRule(ctx, client, req KibanaWriteRequest[enableRuleModel])` function used for both Create and Update, matching the `KibanaWriteFunc[enableRuleModel]` signature.

**Rationale:** The current `upsert` helper already handles both operations identically — there is no meaningful difference between Create and Update for this resource (it's an idempotent enable-by-tag call). A single callback avoids duplication and matches the pattern used by resources like `security_exception_item` and `security_detection_rule` when their create/update logic is symmetric.

**Alternative considered:** Two separate callbacks (`createSecurityEnableRule`, `updateSecurityEnableRule`) as done by `security_detection_rule`. Rejected because the enable-rule operation is idempotent and identical for both, making two callbacks unnecessary indirection.

### Set `SkipReadAfterWrite: true` in the write result

**Decision:** Return `KibanaWriteResult[enableRuleModel]{Model: model, SkipReadAfterWrite: true}`.

**Rationale:** The existing write path sets state directly from the plan model without a round-trip read. The `security_enable_rule` resource has no server-assigned fields that would differ from what was written — `ID` is computed client-side from `spaceID/key:value` and `AllRulesEnabled` is set to `true` unconditionally. Invoking a read-after-write would add an unnecessary API call. The existing `read.go` callback exists for the Terraform `Read` plan path, not for write confirmation.

**Alternative considered:** `SkipReadAfterWrite: false` (default, performs read-after-write). Rejected because the read callback queries `EnabledRulesByTag` to verify rule states, which is a separate API call with no state fields that differ from the written model.

### Remove redundant `EnforceVersionRequirements` from `deleteSecurityEnableRule`

**Decision:** Delete the `entitycore.EnforceVersionRequirements` call (and its associated early-return guard) from `deleteSecurityEnableRule`.

**Rationale:** `baseResourceEnvelope.Delete` (in `base_envelope.go:179`) calls `EnforceVersionRequirements` before invoking the registered delete callback. The `KibanaResource` envelope inherits this path, so the version check is guaranteed to run before `deleteSecurityEnableRule` is called. The duplicate call in the callback is dead code that adds noise and an extra version API round-trip.

## Risks / Trade-offs

**[Risk] Envelope read-after-write path is exercised for the first time on this resource.**
→ Mitigation: `SkipReadAfterWrite: true` is set explicitly, so the envelope skips the read-after-write flow and sets state directly from the callback model — identical to the current wrapper-method behavior.

**[Risk] Version requirement no longer enforced during delete if envelope base changes.**
→ Mitigation: Enforcing version requirements on delete is the envelope's documented responsibility (see `base_envelope.go`). Removing the duplicate from the callback follows the same pattern as other migrated resources. If the envelope base changes, it affects all resources uniformly.

**[Risk] `DisableOnDestroy` null-defaulting logic must be preserved.**
→ Mitigation: The null-check (`if model.DisableOnDestroy.IsNull() { model.DisableOnDestroy = types.BoolValue(true) }`) moves into `writeSecurityEnableRule` exactly as it appeared in `upsert`. No behavior change.

## File Changes

| File | Change |
|------|--------|
| `internal/kibana/security_enable_rule/create.go` | Delete the `Create` wrapper method (the entire file). |
| `internal/kibana/security_enable_rule/update.go` | Delete the `Update` wrapper method and `upsert` helper. Replace with `writeSecurityEnableRule` callback. |
| `internal/kibana/security_enable_rule/delete.go` | Remove `entitycore.EnforceVersionRequirements` call and its guard. |
| `internal/kibana/security_enable_rule/resource.go` | Replace `PlaceholderKibanaWriteCallback` with `writeSecurityEnableRule` for both `Create` and `Update`. |
