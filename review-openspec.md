# OpenSpec Verification: kibana-alerting-rule-discriminator-validation

## Status: PASS (with minor notes)

All requirements from the delta spec are implemented. Tests pass. The only incomplete task is an optional acceptance test.

---

## Requirements Verification

### REQ-018: Plan-time `params` validation — IMPLEMENTED ✓

| Sub-requirement | Status | Evidence |
|-----------------|--------|----------|
| Parse `params` as JSON; report "Invalid params JSON" on failure | ✓ | `ValidateConfig` in `validate.go:135-142` |
| Supported rule types: structural validation via default discriminator path | ✓ | `validateParamsViaDiscriminator` in `validate.go:184-268` |
| Supported rule types: override table for nested unions / Kibana mismatches | ✓ | `ruleTypeParamsOverrides` in `validate.go:165-182` (4 entries) |
| Unsupported rule types: pass-through (no structural checks) | ✓ | `validateParamsViaDiscriminator` returns `nil` for "unknown discriminator value" (`validate.go:209-210`) |
| Error summary: `Invalid params for rule_type_id "<id>"` | ✓ | `ValidateConfig` in `validate.go:149-152` |
| Apply-phase: no duplicate validation | ✓ | `toAPIModel` intentionally does not re-validate (unchanged) |

### REQ-051: Discriminator validation coverage guard — IMPLEMENTED ✓

- `TestDiscriminatorValidationCoversAllKbapiRuleTypes` in `validate_test.go:948-1019` parses `generated/kbapi/kibana.gen.go` with regex, extracts all 35 `ValueByDiscriminator()` cases, and asserts each dispatches successfully.
- Also asserts no override entry exists that is not in `ValueByDiscriminator()`.
- **Test result**: PASS (all 35 subtests pass).

### REQ-052: Params validation override table — IMPLEMENTED ✓

Override table `ruleTypeParamsOverrides` contains exactly 4 entries:

| Rule Type | Rationale | Evidence |
|-----------|-----------|----------|
| `logs.alert.document.count` | Multi-variant union (Params0 / Params1) | `validate.go:179-182` |
| `xpack.uptime.alerts.monitorStatus` | Generated struct + legacy fallback | `validate.go:177-178` |
| `.es-query` | Additional required `size` key | `validate.go:174-175`, `ruleTypeAdditionalRequiredParamsKeys:270` |
| `.index-threshold` | Post-decode `index` array check | `validate.go:172-173`, `validateRuleParamsPostDecode:312-318` |

- `apm.rules.anomaly` removed from overrides — confirmed it is not a valid kbapi discriminator. `apm.anomaly` validates via default path. ✓
- `xpack.uptime.alerts.tls` removed from overrides — kbapi uses `xpack.synthetics.alerts.tls` / `xpack.uptime.alerts.tlsCertificate`. ✓

---

## Scenario Coverage

| Scenario | Status | Test / Evidence |
|----------|--------|-----------------|
| Invalid JSON | ✓ Covered | `ValidateConfig` handles `json.Unmarshal` error → "Invalid params JSON" |
| Unsupported rule type passes structural checks | ✓ Covered | `TestValidateRuleParamsUnknownRuleTypeIsAllowsAnyKey` (`validate_test.go:83-91`) |
| Supported rule type with wrong shape | ✓ Covered | Multiple tests (`.index-threshold`, `.es-query`, `apm.anomaly` invalid fixtures) |
| Discriminator-known rule type validates via default path | ✓ Covered | `observability.rules.custom_threshold` valid fixture (`validate_test.go:688-704`) |
| Discriminator-known rule type rejects unknown params keys | ✓ Covered | `apm.anomaly` rejects `bogusKey` (`validate_test.go:718-729`) |
| Override takes precedence over default path | ✓ Covered | `logs.alert.document.count` multi-variant tests (`validate_test.go:554-571`) |
| APM anomaly uses kbapi rule type id | ✓ Covered | `apm.anomaly` fixtures replace old `apm.rules.anomaly` |
| New kbapi rule type without validation decision | ✓ Covered | `TestDiscriminatorValidationCoversAllKbapiRuleTypes` fails on new unmatched cases |

---

## Minor Notes (non-blocking)

1. **Stub body divergence from spec text** — `validateParamsViaDiscriminator` builds a stub containing only `rule_type_id` and `params`, omitting the `name`, `consumer`, `schedule.interval` fields that REQ-018 describes. This is functionally harmless because `AlertingRuleAPIBody.UnmarshalJSON` stores raw JSON and `ValueByDiscriminator()` only reads `rule_type_id` to dispatch. All 35 discriminator cases pass. Consider aligning the stub with the spec text for clarity if the `UnmarshalJSON` contract ever changes.

2. **Task 3.5 incomplete** — Acceptance test for `observability.rules.custom_threshold` was not added. The implementor scoped it as optional ("if CI Kibana supports it; otherwise document skip reason"). Since unit/fixture tests cover the validation behavior comprehensively, this is acceptable. If maintainers require an acceptance test, it can be added in a follow-up.

3. **`custom_threshold` required-keys risk** — The generated `KibanaHTTPAPIsObservabilityRulesCustomThresholdCreateRuleBodyAlerting` params type marks `searchConfiguration` as a non-pointer struct, so `computeRequiredKeys` treats it as required. If Kibana accepts configs without `searchConfiguration`, this could produce false positives. No practitioner reports exist. This is an OpenAPI/generation issue, not an implementation bug.

---

## Task Completion

- 14 of 15 subtasks complete.
- Remaining: 3.5 (optional acceptance test).
- `openspec validate kibana-alerting-rule-discriminator-validation --strict`: PASS
- `go test ./internal/kibana/alertingrule/...`: PASS
