## Why

Flipping `enabled` on an `elasticstack_kibana_alerting_rule` from `false` to `true` (or `true` to
`false`) fails apply with "Provider produced inconsistent result after apply" on `scheduled_task_id`
([#5492](https://github.com/elastic/terraform-provider-elasticstack/issues/5492)). A rule created
disabled has no Task Manager task, so `scheduled_task_id` is `null` in state. Enabling it makes Kibana
schedule the rule and assign a task id; `scheduled_task_id`'s only plan modifier is
`UseStateForUnknown()` (`internal/kibana/alertingrule/schema.go:150-156`), so the plan keeps the prior
`null` value across the `enabled` transition, and the subsequent read returns a populated value —
exactly the mismatch Terraform rejects. The update itself succeeds against Kibana; a second
`terraform plan`/`apply` is clean because the refresh has already observed the new value. The bug is in
plan-time bookkeeping for a server-assigned attribute, not in the API call.

## What Changes

- Add a plan modifier to `scheduled_task_id` that resets the planned value to unknown whenever the
  rule's effective `enabled` value is changing between prior state and the new plan (either
  direction: disabled → enabled or enabled → disabled). Compare `enabled` using plan vs. state (not
  config vs. state), so omitting `enabled` from config — which resolves to its default via
  `booldefault.StaticBool(true)` — while the effective value is unchanged does not spuriously mark
  `scheduled_task_id` unknown.
- Reuse the existing `planmodifiers.StringSetUnknownIf` helper
  (`internal/utils/planmodifiers/set_unknown_if.go`) and the same composition pattern already used for
  `throttle` in this resource (`throttleSetUnknownIfActionsFrequencyConfigured()`,
  `internal/kibana/alertingrule/planmodifiers.go:117-127`, wired in at `schema.go:145-148`): append the
  new modifier after `UseStateForUnknown()` on `scheduled_task_id`.
- Extend acceptance coverage so the `false` → `true` transition (the exact failure in the issue) and
  the symmetric `true` → `false` transition are exercised end-to-end, including a case where `enabled`
  is enabled via config omission (default) rather than an explicit `true`, and a no-op case while
  `enabled` stays unchanged.
- Add a unit test for the new predicate alongside the existing pure-function tests in
  `internal/kibana/alertingrule/planmodifiers_test.go`.

## Capabilities

### New Capabilities

- None. This change modifies an existing capability.

### Modified Capabilities

- `kibana-alerting-rule`: Adds plan-time handling so `scheduled_task_id` is marked unknown whenever
  `enabled` changes between state and plan, fixing the inconsistent-result failure on enable/disable
  transitions. No change to the API request/response contract, the read-mapping persistence rules for
  a partial response (REQ-026), or any other requirement in this capability.

## Impact

- **Changed code**: `internal/kibana/alertingrule/schema.go` (new plan modifier appended to
  `scheduled_task_id`'s `PlanModifiers`), `internal/kibana/alertingrule/planmodifiers.go` (new
  predicate and modifier constructor, following the `throttle` precedent).
- **New tests**: unit test for the new predicate in
  `internal/kibana/alertingrule/planmodifiers_test.go`; acceptance test coverage extending
  `TestAccResourceAlertingRuleEnabledFalseOnCreate`
  (`internal/kibana/alertingrule/acc_test.go:560-587`) into multi-step enable/disable transitions.
- **Docs**: none — `scheduled_task_id` remains `Computed`-only with no schema-visible change.
- **Out of scope**: any change to `models.go`'s partial-response preserve logic for
  `scheduled_task_id` (REQ-026) after a disable, since it is not established whether Kibana's GET
  response omits or nulls the field once a rule's task is removed; this is tracked as an open question
  and a possible independent follow-up, not part of this fix. Also out of scope: a resource-level
  `ModifyPlan` rewrite of this logic (considered and rejected in favor of the smaller, more idiomatic
  attribute-level modifier — see `design.md`).
