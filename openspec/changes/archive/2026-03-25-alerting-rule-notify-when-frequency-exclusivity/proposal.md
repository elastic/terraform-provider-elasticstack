## Why

Kibana’s alerting APIs treat **rule-level** notification controls (`notify_when`, `throttle`) and **per-action** `frequency` as mutually exclusive: the action `frequency` documentation states these parameters cannot be used when `notify_when` or `throttle` are defined at the rule level. The provider today can still plan and send both, which risks API errors or ambiguous behavior. Separately, top-level `notify_when` is optional and computed with plan semantics that can leave values **unknown** while practitioners configure action `frequency` only—another source of friction and drift. This change addresses both problems with complementary mechanisms.

## What Changes

- **Plan modifier (top-level `notify_when`)**  
  When the planned value for top-level `notify_when` is **unknown** and **any** `actions` entry includes a `frequency` block (the nested object practitioners use for per-action notification behavior), the provider SHALL normalize the planned top-level `notify_when` to **null** so create/update does not send a rule-level `notify_when` in that situation. This targets unknown/computed plan edges without overriding an explicitly configured rule-level value.

- **Validation (mutual exclusivity)**  
  The provider SHALL reject configuration when **both** of the following hold:

  1. **Rule level:** the practitioner sets **either** top-level `notify_when` (known, non-null / non-empty as applicable) **or** top-level `throttle` (known, non-empty), mirroring the Kibana documentation that forbids rule-level `notify_when` **or** `throttle` together with action-level frequency parameters.

  2. **Action level:** **any** `actions` entry includes a `frequency` block.

  Validation applies only when **both** sides are present in configuration as above. It does **not** require rule-level `notify_when` on stacks before 8.6 where the API still requires it without action `frequency`; configurations that only set rule-level notification (no action `frequency`) remain valid. Configurations that only set action `frequency` (for example `TestAccResourceAlertingRule/frequency_create/rule.tf`, which omits top-level `notify_when` and uses `frequency` inside `actions`) remain valid.

- **BREAKING:** Practitioners who today set **both** rule-level `notify_when` or `throttle` **and** at least one action `frequency` block in the same resource will receive a plan-time error and must remove the redundant rule-level attributes or drop action `frequency` in favor of rule-level behavior, per Kibana rules.

- **Delta spec (this change):** REQ-041 (plan normalization for unknown top-level `notify_when` when action `frequency` exists) and REQ-042 (mutual exclusivity validation) in `openspec/changes/alerting-rule-notify-when-frequency-exclusivity/specs/kibana-alerting-rule/spec.md`.

- **Out of scope for this proposal artifact:** merging those requirements into canonical `openspec/specs/kibana-alerting-rule/spec.md` (sync or archive workflow).

## Capabilities

### New Capabilities

- _(none)_

### Modified Capabilities

- `kibana-alerting-rule`: Add requirements for (1) plan-time normalization of top-level `notify_when` when unknown and action `frequency` is configured, and (2) plan-time validation rejecting configurations that combine rule-level `notify_when` or `throttle` with any action `frequency` block, aligned with embedded Kibana-oriented documentation.

## Impact

- **Specs:** Delta under `openspec/changes/alerting-rule-notify-when-frequency-exclusivity/specs/kibana-alerting-rule/spec.md` (to be authored after this proposal).
- **Implementation (future):** `internal/kibana/alertingrule`—custom string plan modifier on `notify_when`, resource- or config-level validator crossing `notify_when`, `throttle`, and `actions`; unit tests for validation and modifier behavior; acceptance tests adjusted only if any existing case intentionally combined rule-level `notify_when`/`throttle` with action `frequency` (frequency-only fixtures such as `frequency_create` need no change for exclusivity).
