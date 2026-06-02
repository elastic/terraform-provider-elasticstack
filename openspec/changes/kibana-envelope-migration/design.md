## Context

The provider uses `entitycore.KibanaResource[T]` as the standard envelope for Kibana-backed resources. It owns Schema, Create, Read, Update, Delete, client resolution, version-requirement enforcement, and read-after-write. Resources only supply a model type + four callbacks.

Three Kibana resources (`kibana_security_exception_item`, `kibana_security_detection_rule`, `kibana_alerting_rule`) still use the older `*entitycore.ResourceBase` pattern. All three already have `kibana_connection types.List` and `space_id types.String` (singular) in their models — the cleanest possible fit for the envelope's `KibanaResourceModel` interface.

Relevant reference: `internal/kibana/connectors`, `internal/kibana/securityexceptionlist`, `internal/kibana/securitylistitem` are already migrated and serve as in-repo templates.

## Goals / Non-Goals

**Goals:**
- Replace `*entitycore.ResourceBase` with `*entitycore.KibanaResource[T]` in all three packages
- Add the four `KibanaResourceModel` interface methods to each model
- Remove manual `kibana_connection` block injection from each schema function
- Convert CRUD method bodies to package-level callback functions
- Add `entitycore_contract_test.go` to each package

**Non-Goals:**
- Changing the entitycore envelope itself
- Migrating any other Kibana or Fleet resources
- Changing user-visible schema, state format, or behaviour

## Decisions

### D1: `GetResourceID()` per resource

The envelope uses `GetResourceID()` as the plan-safe write identity: the value that can be determined from config (before the API responds) and that Read/Update/Delete use to address the resource.

- **security_exception_item** → `ItemID` (the `item_id` field; user-provided or API-assigned)
- **security_detection_rule** → `RuleID` (the `rule_id` field; user-provided or API-assigned)
- **alerting_rule** → `RuleID` (the `rule_id` field; API-assigned on create)

### D2: alerting_rule composite ID and `resolveKibanaResourceIdentity`

The current resource stores a composite `<spaceID>/<ruleID>` string as the `ID` field (set after create and after import). The existing `getRuleIDAndSpaceID()` helper parses it, falling back to the raw `rule_id`/`space_id` fields if parsing fails.

The envelope's `resolveKibanaResourceIdentity` implements the same logic: it attempts to parse `GetID()` as a composite ID; on failure it falls back to `GetResourceID()` + `GetSpaceID()`. These are functionally equivalent.

**Decision**: After migration, `getRuleIDAndSpaceID()` is removed. The envelope handles identity resolution for Read, Update, and Delete automatically.

**Alternative considered**: Keep `getRuleIDAndSpaceID()` and call it inside callbacks. Rejected — it duplicates logic already in the envelope, and the envelope's version is tested independently.

### D3: security_detection_rule private `read` helper

The current code has a private `read` method called both from the `Create` CRUD method (for read-after-write) and from the `Read` CRUD method. After migration, this becomes the `readFunc` callback. The Create callback no longer calls it directly — the envelope performs read-after-write automatically after the create callback returns.

**Decision**: Promote the `read` helper to a package-level `readDetectionRule` callback function; remove the manual read-after-write call from the existing `Create` method body.

### D4: Preserved wrapper-struct interfaces

The following interfaces are not owned by the envelope and stay on the outer resource struct unchanged:

| Resource | Preserved interfaces |
|---|---|
| security_exception_item | `ValidateConfig` |
| security_detection_rule | `UpgradeState`, `ImportState` (passthrough) |
| alerting_rule | `ValidateConfig`, `UpgradeState`, `ImportState` (composite ID parser) |

Go method promotion means these methods are visible on the wrapper struct without changes. `validate.go` (442 lines) in alerting_rule is untouched.

### D5: `space_id` is singular — no special handling needed

All three resources use `space_id types.String` (singular), which maps directly to `GetSpaceID() types.String`. No `KibanaUnscopedSpace` implementation is needed. The envelope's standard space validation applies.

### D6: Full envelope callback migration — no placeholders

**Decision**: All three resources SHALL fully migrate to envelope callbacks. No `PlaceholderKibanaWriteCallback` usage. Create, Read, Update, and Delete are supplied as `KibanaResourceOptions` callbacks; the wrapper struct does NOT override `Create` or `Update`.

**Rationale**: The placeholder pattern (seen elsewhere in `agentdownloadsource`, `security_enable_rule`, `synthetics/privatelocation`) is a half-migration artefact. For these three resources there is no lifecycle wrinkle that prevents using the envelope's callback dispatch directly.

### D7: Version requirements via `GetVersionRequirements`

All three resources currently perform server-version gating via inline `EnforceMinVersion` calls inside model-conversion helpers, with the helpers receiving `clients.MinVersionEnforceable` parameters threaded through their call chains. `alertingrule` additionally has a bespoke `features` struct populated by `resolveAlertingRuleFeatures` and threaded into `toAPIModel`.

**Decision**: All three models implement `entitycore.WithVersionRequirements` by returning a list of `VersionRequirement`s whose membership is conditional on model fields. The envelope's `EnforceVersionRequirements` then handles the checks. All inline `EnforceMinVersion` calls, the `MinVersionEnforceable` parameters on model-conversion helpers, the `features` parameter on `toAPIModel`, the `features.go` file, and `resolveAlertingRuleFeatures` are deleted.

| Resource | `GetVersionRequirements()` returns |
|---|---|
| security_exception_item | 8.7.2 when `ExpireTime` is set |
| security_detection_rule | 8.16.0 when response actions are configured; 8.9.0 when alerts_filter is configured |
| alertingrule | 8.6.0 when `Frequency` set on any action; 8.6.0 when `NotifyWhen` is unset/empty (notify_when is required below 8.6); 8.9.0 when `AlertsFilter` set; 8.13.0 when `AlertDelay` set; 8.16.0 when `Flapping` set; 9.3.0 when `Flapping.Enabled` set |

**Behaviour parity for alertingrule**: The existing `toAPIModel(ctx, features)` already raises clear errors when an unsupported field is set (e.g. `"alert_delay is only supported for Kibana v8.13 or higher"`). It is not silent degradation. Migrating to `GetVersionRequirements()` is a pure refactor — the same error fires at a slightly earlier lifecycle point (before the callback rather than during API request construction).

**The `NotifyWhen` reverse case**: One alertingrule rule is structurally backwards: "if `notify_when` is unset, require 8.6+". Expressed as a version requirement: when `NotifyWhen` is null/empty, the model emits a requirement at `frequencyMinSupportedVersion` (8.6.0). The error message remains "notify_when is required until v8.6".

## Risks / Trade-offs

- **Composite ID parity** — The envelope's `resolveKibanaResourceIdentity` must reproduce the behaviour of `getRuleIDAndSpaceID()`. Both parse `<spaceID>/<ruleID>` via `clients.CompositeIDFromStr`. Verified: the implementations are equivalent.
- **Read-after-write for security_detection_rule** — Today Create manually calls `r.read(...)` after the API create call. After migration, the envelope calls `readFunc` automatically. The net behaviour is identical; the risk is transcription error when promoting the helper to a callback. Existing acceptance tests catch this.
- **Behaviour parity** — Pure structural refactor. Callbacks contain the same logic as the current method bodies. Existing acceptance tests are the primary safety net.
- **Version requirements firing point** — The envelope evaluates `GetVersionRequirements()` in Create, Update, AND Read paths. Today, version checks inside `toCreateRequest`/`toUpdateRequest`/`toAPIModel` fire only during Create and Update. Read with a stale-but-once-valid resource won't change behaviour — if it existed in state, Create/Update must have succeeded on a supporting server. Server downgrade is a corner case where the new strictness is arguably more correct.

## Migration Plan

Each resource is independently migratable. Suggested order: `security_exception_item` (simplest) → `security_detection_rule` → `alerting_rule` (most interfaces).

Per resource:
1. Add interface methods to model (`GetID`, `GetResourceID`, `GetSpaceID`, `GetKibanaConnection`)
2. Remove `kibana_connection` block from schema function
3. Extract CRUD method bodies into package-level callback functions
4. Swap `*entitycore.ResourceBase` for `*entitycore.KibanaResource[T]` in resource struct and constructor
5. Add `entitycore_contract_test.go`
6. Run `make build` and existing tests

For alerting_rule: additionally remove `getRuleIDAndSpaceID()` after callbacks are wired.

## Open Questions

_(none)_
