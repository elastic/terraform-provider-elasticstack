## 1. Spec

- [x] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `./node_modules/.bin/openspec validate alerting-rule-notify-when-frequency-exclusivity` (or `make check-openspec` after sync).
- [x] 1.2 On completion of implementation, **sync** delta into `openspec/specs/kibana-alerting-rule/spec.md` or **archive** the change per project workflow.

## 2. Implementation — REQ-042 (validation)

- [x] 2.1 Add `validateNotifyWhenThrottleFrequencyExclusivity` (or equivalent) in `internal/kibana/alertingrule`, called from `ValidateConfig` after successful `req.Config.Get` into `alertingRuleModel` (and before or after params validation—either is fine if diagnostics are consistent).
- [x] 2.2 Implement exclusivity: if **any** action has **`frequency` non-null** in config **and** (**non-empty known `notify_when`** OR **non-empty known `throttle`**), add attribute error(s) with a clear message referencing rule-level `notify_when` / `throttle` vs action `frequency` (aligned with embedded `actions_frequency.md`).
- [x] 2.3 Choose diagnostic path(s): e.g. root `notify_when` when that attribute is set, else `throttle`, else first `actions[i].frequency`—consistent and testable.

## 3. Implementation — REQ-041 (plan modifier)

- [x] 3.1 Implement a custom **`planmodifier.String`** that: reads **config** for `actions`; if any element has **`frequency` not null** and the **planned** `notify_when` is **unknown**, set planned value to **null**; if planned value is **known**, no-op.
- [x] 3.2 Register the modifier on **`notify_when`** in `schema.go` **before** `stringplanmodifier.UseStateForUnknown()` per `design.md`.
- [x] 3.3 Add a short comment at the schema or modifier explaining ordering and REQ-041.

## 4. Testing

- [x] 4.1 **Unit tests** in `validate_test.go` (or new file): REQ-042 cases — `notify_when` + `frequency`; `throttle` + `frequency`; frequency-only; rule-level only; no false positive when `frequency` absent.
- [x] 4.2 **Unit or integration tests** for the plan modifier: unknown `notify_when` + frequency in config → planned null; known `notify_when` unchanged (when not invalid under REQ-042).
- [x] 4.3 **Acceptance / fixtures:** grep testdata for configs that set **both** top-level `notify_when` or `throttle` **and** `actions { frequency { ... } }`; update those fixtures to remove the invalid combination **or** replace with a negative test expecting validation error, if the suite should assert BREAKING behavior.
- [x] 4.4 Confirm **`frequency_create`** and similar frequency-only fixtures remain unchanged (no rule-level `notify_when`/`throttle`).
- [x] 4.5 Run `make build` and targeted `go test` for `internal/kibana/alertingrule`.
- [x] 4.6 Run **`internal/kibana/alertingrule` acceptance tests** against a live Elastic Stack using the environment and command pattern in [`dev-docs/high-level/testing.md`](../../../dev-docs/high-level/testing.md): set **`TF_ACC=1`** and the **`ELASTICSEARCH_*` / `KIBANA_ENDPOINT`** variables (see “Required environment variables” and the example targeted run in that doc); verify the stack is reachable as described there; then run the package acceptance tests, for example  
  `go test -v -run 'TestAccResourceAlertingRule' ./internal/kibana/alertingrule`.
