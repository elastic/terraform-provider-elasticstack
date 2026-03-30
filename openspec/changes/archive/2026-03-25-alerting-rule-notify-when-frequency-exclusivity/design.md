## Context

`elasticstack_kibana_alerting_rule` maps to Kibana rule create/update payloads built in `internal/kibana/alertingrule/models.go` (`toAPIModel`, `convertActionsToAPI`). Rule-level `notify_when` is optional, computed, and uses `stringplanmodifier.UseStateForUnknown()`. Per-action notification is modeled as nested `actions { frequency { ... } }` (single nested block). Kibana documents that rule-level `notify_when` and `throttle` must not be combined with action-level `frequency` parameters (see embedded `descriptions/actions_frequency.md`). The provider does not yet enforce that split or normalize plans when `notify_when` stays unknown while practitioners configure only action `frequency`.

Plan-time validation for rule params lives in `ValidateConfig` (`validate.go`) using `resource.ValidateConfigRequest` and `alertingRuleModel`.

## Goals / Non-Goals

**Goals:**

- **REQ-041:** When configuration includes at least one `actions` entry with a **`frequency` block present** and the planned top-level `notify_when` is **unknown**, the custom plan modifier sets it to **null** before `UseStateForUnknown()` runs, so `toAPIModel` does not emit rule-level `notify_when` from an unknown placeholder; `UseStateForUnknown()` still fills **remaining** unknowns from state when the custom modifier did not apply.
- **REQ-042:** At validate/plan time, reject configuration when **any** action has a `frequency` block **and** the practitioner sets **either** a non-empty known top-level `notify_when` **or** a non-empty known top-level `throttle`.
- Keep **REQ-014** / **REQ-015** behavior: exclusivity validation only fires when **both** rule-level notification fields (as above) **and** action `frequency` are in play; frequency-only configs (e.g. `frequency_create`) and rule-level-only configs stay valid.

**Non-Goals:**

- Changing Kibana API semantics or adding new Terraform attributes.
- Automatically clearing rule-level `throttle` via plan modifiers (only `notify_when` is normalized per REQ-041).
- Solving all possible drift when API returns legacy rule-level `notify_when` after practitioners use only action `frequency` beyond the unknown-plan case.

## Decisions

| Topic | Decision | Alternatives considered |
|--------|-----------|-------------------------|
| REQ-042 placement | Extend existing **`ValidateConfig`** on `Resource` (same entry point as `params` validation): after `Get` into `alertingRuleModel`, run a small helper (e.g. `validateNotifyWhenThrottleFrequencyExclusivity`) that appends diagnostics on `notify_when`, `throttle`, or `actions` as appropriate. | **ConfigValidator** at provider level — heavier and loses resource-local paths; **apply-time only** — fails later and duplicates REQ-042 (“plan/validate time”). |
| “Frequency block present” | Treat as: Terraform config for that action has **`frequency` not null** (known object), consistent with “practitioner included the block.” Do not require inner fields to be fully known for exclusivity if the block is present—REQ-042 is about intent to use per-action frequency; partial blocks may still fail `AlsoRequires` on nested attributes separately. | Requiring fully populated `frequency` — weaker signal for exclusivity and harder to align with “includes a frequency block.” |
| Rule-level `notify_when` / `throttle` “set” | **`notify_when`:** known, not null, and **non-empty** string after trim (if applicable). **`throttle`:** known, not null, and **non-empty** string. Matches `toAPIModel`’s “omit when empty” behavior and proposal wording. | Treating unknown `notify_when` as conflicting — would break frequency-only configs and contradict REQ-042. |
| REQ-041 mechanism | Custom **`planmodifier.String`** on `notify_when`, listed **before** `stringplanmodifier.UseStateForUnknown()`: when the planned value is **unknown** and configuration (modifier request **config**) shows any action with non-null `frequency`, set planned `notify_when` to **null**; then `UseStateForUnknown()` runs and may copy state only for values that are **still** unknown. | **Resource `ModifyPlan`** — possible but scatters logic; **only validation** — does not fix unknown-plan emission without forcing practitioners to set `notify_when = null` explicitly; **custom after USFU** — rejected: must be **before** per project decision. |
| Modifier + validation interaction | Modifier **never** overwrites a **known** planned `notify_when`. Invalid combos (known rule-level `notify_when` or `throttle` + `frequency`) are **only** handled by **REQ-042** (error), not by clearing. | Single mechanism — validation alone cannot set planned null for unknowns. |

## Risks / Trade-offs

- **[Risk]** Modifier order wrong → overwrites state-backed `notify_when` or fails to null unknown. **→ Mitigation:** Integration/unit tests; keep the **custom REQ-041 modifier before** `UseStateForUnknown()`; document order in code comment.
- **[Risk]** **BREAKING** configs that mixed rule-level `notify_when`/`throttle` with `frequency`. **→ Mitigation:** Clear diagnostic text pointing to Kibana exclusivity; changelog / upgrade note when releasing.
- **[Risk]** Edge cases where config `frequency` is unknown at validate (e.g. dynamic block). **→ Mitigation:** REQ-042 applies when values are **known** per spec; if `frequency` presence is unknown, validation may skip conflict (document as limitation if observed).

## Migration / rollback

- No state upgrade: schema version unchanged.
- Rollback: revert provider version; practitioners who already removed conflicting attributes keep valid configs.

## Open Questions

- None blocking implementation; confirm modifier behavior on **import** / first plan with `frequency` only in a follow-up test if acceptance suite gaps appear.
