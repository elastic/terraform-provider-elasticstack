## Schema coverage report: elasticstack_kibana_alerting_rule

### Scope
- **Schema**: `internal/kibana/alertingrule/schema.go` — unchanged by this change; no schema modifications.
- **Acceptance tests**: `internal/kibana/alertingrule/acc_test.go` + testdata configs
- **Validation unit tests**: `internal/kibana/alertingrule/validate_test.go`
- **Change focus**: Plan-time `params` validation now routes through `kbapi.AlertingRuleAPIBody.ValueByDiscriminator()` for 35 rule types, replacing a 12-entry hand-maintained map.

### Summary
The resource schema was **not modified** in this change — only the plan-time `params` validation path was refactored. Therefore, this coverage review focuses on whether existing acceptance tests exercise the **new validation paths** for rule types that moved from pass-through to discriminator-validated, and whether there are acceptance-test-level gaps for high-risk rule types.

---

### 1) Rule types with no acceptance test coverage (validation-related gap)

These rule types are now validated by the discriminator path but have **no acceptance tests** that create a real rule via Terraform. They are only covered by unit/fixture tests in `validate_test.go`.

| Rule type | Coverage in unit tests | Acceptance test gap |
|-----------|----------------------|---------------------|
| `observability.rules.custom_threshold` | ✓ valid params, ✓ unknown key via `AdditionalProperties` (doesn't reject) | **No acceptance test** — was explicitly deferred as task 3.5 |
| `monitoring_alert_cluster_health` | ✓ valid params | **No acceptance test** |
| `apm.anomaly` | ✓ valid params, ✓ rejects unknown key | **No acceptance test** |
| `apm.error_rate` | ✓ valid params | **No acceptance test** |
| `apm.transaction_duration` | ✓ valid params | **No acceptance test** |
| `apm.transaction_error_rate` | ✓ valid params | **No acceptance test** |
| `slo.rules.burnRate` | ✓ valid params, ✓ rejects unknown key | **No acceptance test** |
| `transform_health` | ✓ valid params, ✓ rejects unknown key | **No acceptance test** |
| `metrics.alert.threshold` | ✓ valid params (via fixtures from cluster mgmt) | **No acceptance test** |

**Gap significance**: MEDIUM-HIGH. The discriminator validation path exercises reflection on `Params`, strict re-decode, and required-keys heuristics. While unit tests cover the validation logic, acceptance tests would catch issues that only appear when Kibana actually accepts/rejects the payload (e.g., `AdditionalProperties` semantics, Kibana-side defaults that differ from generated types). `observability.rules.custom_threshold` is the highest priority because it was the motivating issue (#940).

---

### 2) Rule types with acceptance test coverage

| Rule type | Test function | Notes |
|-----------|--------------|-------|
| `.index-threshold` | `TestAccResourceAlertingRule`, `TestAccResourceAlertingRuleParamsLifecycle`, `TestAccResourceAlertingRuleEnabledFalseOnCreate`, `TestAccResourceAlertingRuleFlapping`, `TestAccResourceAlertingRuleFlappingEnabled`, `TestAccResourceAlertingRuleThrottle`, `TestAccResourceAlertingRuleFrequencyExclusivity`, `TestAccResourceAlertingRuleAutoRuleID`, `TestAccResourceAlertingRuleFromSDK` | Extensive coverage; also params lifecycle with add/remove aggType |
| `.es-query` | `TestAccResourceAlertingRuleAlertDelay`, `TestAccResourceAlertingRuleEsqlTermField`, `TestAccResourceAlertingRuleInconsistentParams` | Good coverage including ESQL variant and inconsistent params |
| `logs.alert.document.count` | `TestAccResourceAlertingRule` (alerts_filter_create/update steps) | Covered via alerts filter testing |

---

### 3) Schema attributes with no/poor coverage (unrelated to this change)

These are pre-existing gaps in acceptance test coverage for the alerting rule resource schema, not introduced by this change:

- `flapping.enabled`: Only tested at `true` in `TestAccResourceAlertingRuleFlappingEnabled`. No test verifies omission or `false` on create.
- `actions.alerts_filter.kql` without `timeframe`: Tested in `alerts_filter_update`, but `timeframe` is also present. No standalone `kql`-only test.
- `actions.alerts_filter.timeframe.days/timezone/hours_start/hours_end`: Only tested together as a complete block.
- `last_execution_status`, `last_execution_date`, `scheduled_task_id`: Computed-only; only `scheduled_task_id` is asserted as `Set`.

---

### 4) Specific validation-path risks from this change

#### Risk A: Discriminator path on real Kibana payloads
The unit tests mock params as `map[string]any` and call `validateRuleParams` directly. They do not test the full `ValidateConfig` → `validateRuleParams` → `validateParamsViaDiscriminator` path with real Terraform config parsing. The acceptance tests that exist (`.index-threshold`, `.es-query`, `logs.alert.document.count`) exercise the full path, but only for override-table rule types (`.index-threshold`, `.es-query`, `logs.alert.document.count`).

**No acceptance test exercises the default discriminator path** (i.e., a rule type NOT in the override table being created via Terraform).

#### Risk B: `AdditionalProperties` types silently accepting unknown keys
Types like `observability.rules.custom_threshold`, `metrics.alert.threshold`, and monitoring rules have `AdditionalProperties` maps in the generated kbapi types. `DisallowUnknownFields()` does NOT reject unknown keys for these types. The unit test for `custom_threshold` documents this (`"custom threshold accepts unknown key via additionalProperties"`), but practitioners may be surprised. This is expected behavior per the design doc, but there is no acceptance test verifying that Kibana actually accepts such payloads.

#### Risk C: Required-keys heuristic false positives
The required-keys heuristic (`computeRequiredKeys`) marshals a zero-value struct to discover non-omitempty fields. For `observability.rules.custom_threshold`, the generated type marks `searchConfiguration` as required (non-pointer struct). If Kibana accepts configs without it, this would cause false positives. An acceptance test would validate whether Kibana's runtime behavior matches.

---

### Suggested next steps (smallest diffs first)

1. **Acceptance test for `observability.rules.custom_threshold`** (task 3.5): Add a minimal `TestAccResourceAlertingRuleCustomThreshold` that creates a rule with valid params, then either:
   - Adds an `ExpectError` step with an invalid/unknown param to test rejection (if the type's `AdditionalProperties` doesn't swallow it), or
   - Just verifies successful create + update if strict rejection isn't expected.
   
   If the local CI Kibana doesn't expose this rule type, add a `SkipFunc` that checks the Kibana version/features.

2. **Acceptance test for `apm.anomaly`**: A minimal create/destroy test using the valid fixture from `validate_test.go`. This validates the default discriminator path end-to-end for a non-override rule type.

3. **Add a negative validation acceptance test**: Use `ExpectError` to verify that Terraform plan fails when `params` contains a genuinely unknown key for a discriminator-validated rule type that does NOT use `AdditionalProperties` (e.g., `apm.anomaly` or `slo.rules.burnRate`). This would be the first acceptance test asserting the new stricter validation at the Terraform level.

---

### Verdict

- **Schema coverage**: No change — schema untouched.
- **Validation path coverage**: The unit test suite (`validate_test.go`) is comprehensive for the refactored logic. The **critical gap** is the lack of any acceptance test exercising the **default discriminator path** (non-override rule types) end-to-end through Terraform's `ValidateConfig`. Task 3.5 (`observability.rules.custom_threshold` acceptance test) is the most impactful follow-up.
