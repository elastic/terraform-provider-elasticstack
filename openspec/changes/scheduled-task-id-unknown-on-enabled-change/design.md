## Context

`elasticstack_kibana_alerting_rule` fails with "Provider produced inconsistent result after apply" on
`scheduled_task_id` when a rule created with `enabled = false` is flipped to `enabled = true` (and,
per the issue's "Expected behavior" section, the symmetric `true` → `false` transition is expected to
have the same bug). The provider's write path already distinguishes an enabled rule from a disabled
one across the update: `UpdateAlertingRule` sends the PUT, unmarshals the response's `enabled` field,
and only then calls the dedicated `_enable`/`_disable` endpoint
(`internal/clients/kibanaoapi/alerting_rule.go:173,233-255`). `scheduled_task_id` is `Computed` with
only `UseStateForUnknown()` (`internal/kibana/alertingrule/schema.go:150-156`), so on this transition
the plan keeps the prior (null) value, and the subsequent GET returns a populated value — the mismatch
Terraform rejects.

Whether Kibana's Task Manager behavior (task created only once enabled, possibly removed on disable)
holds across all supported versions is not established by any source consulted during research; this
is tracked as a non-blocking open question below and does not change the recommended fix, since
marking the field unknown on either transition direction is correct regardless of what Kibana's GET
response does afterward.

No human direction was provided on the issue beyond the bug report itself, so this design adopts the
research comment's recommendation (Approach 1) as-is.

## Goals / Non-Goals

**Goals:**
- Make `scheduled_task_id` plan as unknown whenever the rule's effective `enabled` value is changing
  between prior state and the new plan, in either direction, so apply no longer produces an
  inconsistent-result error on this transition.
- Keep config omission of `enabled` (which resolves to `booldefault.StaticBool(true)`,
  `schema.go:123-128`) from spuriously marking `scheduled_task_id` unknown when the effective value is
  not actually changing.
- Follow the existing in-resource precedent (`throttleSetUnknownIfActionsFrequencyConfigured()`) so the
  fix is minimal and idiomatic for this schema file.

**Non-Goals:**
- Changing `models.go`'s partial-response preserve logic for `scheduled_task_id` after a disable. It is
  not established whether Kibana's GET response omits or nulls the field once a rule's task is removed,
  and this fix is correct regardless (the field becomes unknown through the normal post-update read,
  not through the preserve-on-omission path, for the very apply that changes `enabled`). Any further
  staleness after a *subsequent* unrelated update is an independent, unverified question — see Open
  Questions.
- A resource-level `ModifyPlan` rewrite of this behavior.
- Any change to the `_enable`/`_disable` reconciliation flow in
  `internal/clients/kibanaoapi/alerting_rule.go` or to the PUT request contract — the request body has
  no `enabled` field; the response body's `enabled` field is already read by `reconcileRuleEnabled` to
  decide which endpoint to call, and that plumbing is unchanged by this fix.
- A Stack/Kibana minimum-version gate. This is pure plan-time bookkeeping with no new endpoint or
  request/response field, so no new version dependency is introduced.

## Decisions

### 1. Add a `StringSetUnknownIf` plan modifier to `scheduled_task_id`, predicate comparing Plan vs. State on `enabled`

Append a new plan modifier after the existing `UseStateForUnknown()` on `scheduled_task_id`
(`internal/kibana/alertingrule/schema.go:150-156`), built with the existing
`planmodifiers.StringSetUnknownIf` helper (`internal/utils/planmodifiers/set_unknown_if.go:31-62`) and
a new predicate in `internal/kibana/alertingrule/planmodifiers.go`, mirroring
`throttleSetUnknownIfActionsFrequencyConfigured()` (`planmodifiers.go:117-127`) and its wiring at
`schema.go:145-148`.

The predicate compares `req.Plan`'s `enabled` value against `req.State`'s `enabled` value — not
`req.Config` against `req.State`. `enabled` is `Optional` + `Computed` with
`booldefault.StaticBool(true)` (`schema.go:123-128`), so an omitted config value does not by itself
reflect the default-resolved effective value; only the plan (after defaults are applied) carries that.
Comparing plan vs. state means a config omission that keeps the effective value unchanged (e.g.
`enabled` already `true` in state, and the practitioner removes the line from config) is correctly
treated as a no-op, while an omission that *does* change the effective value (e.g. state has `false`,
config now omits `enabled` entirely so the default `true` takes effect) is correctly treated as a
transition.

When planned `enabled` is unknown (e.g. derived from another resource), the predicate treats that
conservatively as "assume it may be changing" and sets `scheduled_task_id` unknown, rather than
panicking or silently skipping the modifier — consistent with not being able to prove the value is
unchanged.

Why:
- Directly reuses an existing, already-tested helper and mirrors the one existing precedent in this
  same schema file, keeping the diff minimal and idiomatic.
- `StringSetUnknownIf` is already a no-op on create/destroy (`req.State.Raw.IsNull() ||
  req.Config.Raw.IsNull()` short-circuit in `set_unknown_if.go:53-55`), so no separate handling is
  needed for those cases — `scheduled_task_id` is already computed fresh on create, and plan modifiers
  don't run meaningfully on destroy.

Alternatives considered:

#### Resource-level `ModifyPlan` hook instead of an attribute-level modifier

Extend the resource's existing `ModifyPlan` (`internal/kibana/alertingrule/plan_modifier.go:36-87`,
which already does custom plan-modification work for this resource — marking the investigation-guide
checksum unknown on file-content drift) to also inspect `enabled` and set `scheduled_task_id` unknown
via `resp.Plan.SetAttribute`.

Rejected because it is heavier-handed for a single-attribute, single-condition fix, duplicates logic
the schema-level helper already encapsulates, and diverges from the more idiomatic, already-precedented
per-attribute pattern used for `throttle` in this same resource. It remains a reasonable direction if a
future change needs to react to more than one triggering condition for `scheduled_task_id` (e.g. a
`rule_type_id` interaction, which is unverified today — see Open Questions) and attribute-level
modifiers become awkward to compose.

## Risks / Trade-offs

- **The fix only triggers on `enabled` changing.** If Kibana's task lifecycle has other edge cases
  affecting `scheduled_task_id` (e.g. a `rule_type_id` change affecting scheduling) those are not
  covered by this predicate. No evidence of such a case was found; if one surfaces, a second predicate
  or an "OR" composition can be added later without restructuring this change.
- **Partial-response ambiguity on disable.** `models.go:129-135`'s preserve-on-omission logic can't
  distinguish a GET response that omits `scheduled_task_id` from one that sets it to JSON `null` (both
  decode to a nil `*string` per `alerting_rule_builders.go:56`). If Kibana's disable path omits the
  field rather than nulling it, a *subsequent* unrelated update could leave a stale known id in state
  rather than correctly nulling it. This plan modifier does not fix that path (the field is marked
  unknown during the disabling apply itself, so the disabling apply's own read is unaffected); it is
  listed as a non-blocking open question for a possible independent follow-up to `models.go`.

## Open questions

- Does disabling an enabled, previously-run rule actually remove the Task Manager task (and would
  Kibana's GET response then omit `scheduled_task_id`, and is that omission distinguishable from the
  field being explicitly null)? Neither the docs nor the OpenAPI spec state this; `models.go:129-135`
  can't tell "field omitted" from "field null" either way. This doesn't block the plan-modifier fix
  itself (unknowning the field on any `enabled` transition is correct regardless), but it does determine
  whether `models.go`'s read-side preserve logic also needs a follow-up to avoid leaving a stale task id
  after disable. (non-blocking for the recommended fix; relevant to a possible follow-up)
- The OpenAPI spec's example payload for the update-rule response shows `scheduled_task_id` as a UUID
  distinct from the rule `id`, while the issue reports the task id equal to the rule id on Kibana
  9.5.3 — likely just an illustrative example value rather than a real invariant, but not confirmed.
  (non-blocking)
- Is there any Kibana version where the rule read response omits `scheduled_task_id` even for enabled,
  running rules? Not found in docs, the spec, or the repo's existing version-gating logic (no version
  check exists today for `enabled`/`scheduled_task_id`), so the recommendation assumes none and
  introduces no new version gate. (non-blocking)
