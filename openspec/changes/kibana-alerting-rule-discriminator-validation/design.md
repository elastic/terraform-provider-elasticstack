## Context

`internal/kibana/alertingrule/validate.go` maintains `ruleTypeParamsSpecs`, a hand-curated map from `rule_type_id` to Go struct targets used for plan-time `params` validation. Each spec decodes practitioner JSON with `json.Decoder.DisallowUnknownFields()` and applies a required-keys heuristic derived from zero-value marshaling.

The generated client (`generated/kbapi/kibana.gen.go`) already models 35 alerting rule create bodies as a discriminated union on `AlertingRuleAPIBody`. oapi-codegen emits:

- `Discriminator()` — reads `rule_type_id` from union JSON
- `ValueByDiscriminator()` — switch dispatch to typed `AsKibanaHTTPAPIs*CreateRuleBodyAlerting()` variants
- 35 `From*` methods that set `RuleTypeId` constants

The provider uses the generic alerting rule API (`PostAlertingRuleId`) and does not need per-type create endpoints. Validation is the only consumer of the typed bodies today.

OpenSpec REQ-018 defines supported vs unsupported rule types. Currently only 12 IDs are "supported"; 25 kbapi-known types pass through with JSON-syntax-only checks.

## Goals / Non-Goals

**Goals:**

- Default params validation for every `rule_type_id` handled by `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()`, without maintaining a parallel 35-entry map.
- Preserve strictness: unknown top-level params keys MUST be rejected for discriminator-known types (via strict re-decode of `Params`).
- Preserve existing override behavior where OpenAPI diverges from Kibana or where params are nested unions.
- Fix `apm.rules.anomaly` → `apm.anomaly` ID mismatch as part of migration to discriminator dispatch.
- Add regression tests that fail if `ValueByDiscriminator()` gains cases not covered by validation.

**Non-Goals:**

- A `elasticstack_kibana_alerting_rule_type` data source (separate change).
- Terraform schema changes (still flat `params` JSON string + `rule_type_id` string).
- Validating nested union internals beyond what strict decode of generated types provides (unless covered by an override multi-variant try).
- Security detection rules (`elasticstack_kibana_security_detection_rule`) — different API union.
- Codegen of `validate.go` from `kibana.gen.go` (runtime approach chosen over build-time generation for this change).

## Decisions

### 1. Default path: stub body + `ValueByDiscriminator()` + strict `Params` re-decode

**Decision:** Build a minimal stub rule JSON `{ rule_type_id, params, name, consumer, schedule }`, unmarshal into `kbapi.AlertingRuleAPIBody`, call `ValueByDiscriminator()`, reflect the `Params` field from the returned typed struct, marshal params back to JSON, and decode into the params type with `DisallowUnknownFields()`.

**Rationale:** `ValueByDiscriminator()` is the canonical rule-type dispatch table already maintained by oapi-codegen. Reflecting `From*`/`As*` methods would not expose `rule_type_id` and would duplicate the switch.

**Alternatives considered:**

| Alternative | Why not |
|-------------|---------|
| Expand hand-maintained `ruleTypeParamsSpecs` to 35 entries | High maintenance; already drifted (`apm.rules.anomaly`) |
| Reflect on all `From*` methods | Cannot derive rule IDs; fragile |
| `ValueByDiscriminator()` alone without strict re-decode | Standard `json.Unmarshal` silently drops unknown keys — weaker than today |
| Build-time codegen from switch | Valid but adds tooling; runtime is sufficient for 35 types |

### 2. Override table for exceptions only

**Decision:** Keep `ruleTypeParamsOverrides` (rename from `ruleTypeParamsSpecs`) keyed by `rule_type_id` for types needing special handling. Overrides replace the default discriminator path entirely for that ID.

**Initial override set (carry forward from current code):**

| `rule_type_id` | Reason |
|----------------|--------|
| `logs.alert.document.count` | Params union: try `Params0` and `Params1` with `DisallowUnknownFields` |
| `xpack.uptime.alerts.monitorStatus` | Legacy `legacyMonitorStatusParams` fallback + Uptime struct (REQ-043/044) |
| `.es-query` | Kibana requires `size` though OpenAPI marks optional (`ruleTypeAdditionalRequiredParamsKeys`) |
| `.index-threshold` | Post-decode: `index` must be array of strings (`validateRuleParamsPostDecode`) |

Remove incorrect entries:

- `apm.rules.anomaly` — delete; covered by discriminator as `apm.anomaly`
- `xpack.uptime.alerts.tls` — delete unless confirmed live on clusters; kbapi uses `xpack.synthetics.alerts.tls` and `xpack.uptime.alerts.tlsCertificate`

**Rationale:** Nested unions and Kibana/OpenAPI mismatches cannot be solved generically without per-type logic.

### 3. Unsupported types: unchanged semantics

**Decision:** If `ValueByDiscriminator()` returns `unknown discriminator value`, validation SHALL return no structural errors (pass-through), matching current REQ-018 unsupported behavior.

**Rationale:** Custom or future Kibana rule types not yet in OpenAPI must remain usable until kbapi regen.

### 4. Required-keys heuristic on default path

**Decision:** After strict params decode succeeds, apply the existing `computeRequiredKeys` + `missingRequiredKeys` heuristic on the decoded params struct, with per-type `additionalRequiredKeys` / `additionalAllowedKeys` patches only in the override table (or a tiny patch map keyed by rule type).

**Rationale:** Preserves today's "missing required field" diagnostics for supported types; strict decode alone does not enforce presence of non-pointer zero-value fields that OpenAPI marks required.

### 5. Coverage guard test

**Decision:** Add `TestDiscriminatorValidationCoversAllKbapiRuleTypes` that parses the generated `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` switch (for example by parsing `generated/kbapi/kibana.gen.go` with `go/parser`) and asserts each discriminator value is either handled by the default path or listed in overrides.

**Rationale:** Prevents kbapi regen from adding types that silently fall back to pass-through without an explicit override decision.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Stricter validation breaks existing Terraform configs with extra params keys on previously pass-through types | Document in changelog; only affects types moving from pass-through to validated; valid Kibana payloads should pass |
| Union params (`custom_threshold` criteria metrics) decode loosely via `json.RawMessage` | Accept for v1; tighten in overrides if field reports emerge |
| `computeRequiredKeys` false positives on optional non-pointer OpenAPI fields | Retain `additionalRequiredKeys` / `additionalAllowedKeys` patch map; migrate patches into override entries |
| Reflection on `Params` field fails for unexpected return types from discriminator | Unit test each discriminator case; handle pointer vs value returns |
| Performance: stub marshal + discriminator + re-decode per validation | Negligible at plan time; validation runs once per resource in `ValidateConfig` |

## Migration Plan

1. Implement default discriminator validator alongside existing map.
2. Switch `validateRuleParams` to: check overrides first, else default path.
3. Remove redundant entries from override table (types fully covered by default path).
4. Run existing `validate_test.go` fixtures; add fixtures for `observability.rules.custom_threshold`, one monitoring rule, `apm.anomaly`.
5. No state migration; schema version unchanged.
6. Rollback: revert to hand map if critical false positives — overrides can temporarily disable default path per type.

## Open Questions

1. **`xpack.uptime.alerts.tls`**: Confirm whether any deployed clusters still use this ID vs `xpack.synthetics.alerts.tls` / `xpack.uptime.alerts.tlsCertificate` before removing override.
2. **Acceptance test for `observability.rules.custom_threshold`**: Does the default CI Kibana stack expose this rule type with `consumer = "logs"`? If not, rely on unit/fixture tests only.
3. **Nested union strictness**: Should `.es-query` params union variants get explicit multi-try in overrides (beyond container-level decode), or is container decode sufficient given existing acc tests?
