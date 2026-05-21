## Context

`elasticstack_elasticsearch_ml_datafeed_state` manages the running state of an ML datafeed. It exposes two timestamps — `start` and `end` — that are accepted by Elasticsearch's `_start` API to bound the search interval. After Elasticsearch starts the datafeed, the *effective* search interval is reported on `running_state.search_interval.{start_ms,end_ms}`. The current implementation reads those values back and **overwrites the user-supplied `start`/`end`**, because the schema was designed when `start` was implicitly null (PR #1563) and the read-back path was the only sane way to populate it.

Elasticsearch does not preserve the user's requested `start`/`end` verbatim:

- For aggregated jobs, ES snaps `start` to the next bucket boundary (`bucket_span`). Issue #2353 example: requested `2025-07-13T02:23:23.935Z`, returned `2025-07-13T02:26:42.000Z` (15-minute bucket alignment).
- For non-aggregated datafeeds, ES snaps `start` forward to the timestamp of the first matching document. The reproducer in `issue_2353_acc_test.go` demonstrates this case (requested `00:07:30Z`, returned `00:10:00Z`).

In both cases the round-trip produces a plan/state mismatch and Terraform aborts with `Provider produced inconsistent result after apply`.

A partial fix already exists for the **null** case: REQ-018 + `SetUnknownIfStateHasChanges` marks `start` unknown when the `state` attribute itself transitions, allowing the read-back value to land in state without violating the framework contract. That mechanism explicitly bails out as soon as the user provides a known `start` value (`if typeutils.IsKnown(req.ConfigValue) { return false }`), which is precisely why #2353 remains broken.

## Goals / Non-Goals

**Goals:**

- Eliminate the "Provider produced inconsistent result after apply" error for any user-supplied `start` / `end` in `elasticstack_elasticsearch_ml_datafeed_state`.
- Preserve the user's declared intent verbatim in state for `start` / `end`.
- Continue to surface Elasticsearch's view of the *effective* search interval so users do not lose observability into how ES interpreted their request.
- Keep the import path working: users who import a datafeed without supplying `start` / `end` must still get a usable state and a clean re-plan.
- Closes #2353.

**Non-Goals:**

- No change to how the datafeed is *started* or *stopped*. The `_start` request still carries the user's `start` / `end`.
- No change to other ML resources (`ml_datafeed`, `ml_job_state`, `ml_anomaly_detection_job`).
- No change to the `state`, `force`, `datafeed_timeout`, `timeouts`, or `elasticsearch_connection` attributes.
- We do not attempt to "round-trip" the user-supplied `start` against the ES-effective start to warn the user when they differ. That's a follow-up if desired.

## Decisions

### Decision 1: Split user input from runtime observation into separate attributes

`start` and `end` become pure user inputs:

- `start`: `Optional` only (drop `Computed`). No plan modifiers.
- `end`: unchanged (already `Optional` only). No plan modifiers.

Two new attributes surface Elasticsearch's view:

- `effective_search_start`: `Computed`, RFC3339, never settable. Populated from `running_state.search_interval.start_ms`.
- `effective_search_end`: `Computed`, RFC3339, never settable. Populated from `running_state.search_interval.end_ms`. Set to null when `running_state.real_time_configured == true` or when the datafeed is `stopped`.

**Why split:** Mixing "intent" and "observed state" in the same attribute is the source of the bug. Decoupling them removes the inconsistency by construction — a `Computed` attribute is allowed to differ from any value, and an input-only attribute is required to round-trip unchanged. There is no middle-ground policy that satisfies the framework contract for the same attribute used both ways.

**Alternative considered — Plan A (preserve `start`, no new attribute):** Smaller diff, no schema growth, but loses the ES-effective value entirely. Users today can at least see ES's view via `terraform state show`; that would be silently dropped. Rejected because the runtime value is the most operationally useful thing in this resource — it tells you what ES is actually searching.

**Alternative considered — keep `start` Computed and use plan modifiers to "accept" whatever ES returns:** The framework does not allow a `Computed+Optional` attribute to disagree with a *known* config value at apply time. The only escape valves are `UseStateForUnknown` (already used) or marking the value unknown in the plan — and we can only mark it unknown when the config value is null. There is no clean mechanism to say "this attribute is user-settable but the provider may rewrite it on apply." Rejected as infeasible.

### Decision 2: Drop REQ-018 and `set_unknown_if_state_has_changes.go`

The plan modifier exists solely to keep the `start` attribute null-safe across `stopped`→`started` transitions. Once `start` is no longer `Computed`:

- When `start` is omitted from config, it is `null` in plan and `null` in state — no round-trip mismatch possible.
- When `start` is supplied, it is the literal config value in plan and state — no round-trip mismatch possible.

The modifier and its file are deleted. REQ-018 is removed from the spec.

### Decision 3: `effective_search_*` semantics when the datafeed is stopped

For `state = "stopped"`, ES does not report `running_state.search_interval`. The new attributes are set to `null` in state. This matches today's behaviour for `start`/`end` on the stopped path (`SetStartAndEndFromAPI` only writes them inside the `if datafeed.State(...) == StateStarted` branch).

### Decision 4: `effective_search_*` semantics during `updateAfterMissedTransition`

`updateAfterMissedTransition` is taken when a datafeed starts and stops too fast to be observed. In that case `running_state` will not contain `search_interval`. The new attributes are set to `null` in state (same null-as-unknown convention). The existing path that nulls `Start` on this branch is removed (no longer needed since `Start` is now driven entirely by config).

### Decision 5: Read-back path

`readMLDatafeedState` continues to call Get Datafeed Stats and now also populates `effective_search_start`/`effective_search_end` via the rewritten `SetStartAndEndFromAPI`. The function no longer touches `state.Start` or `state.End` — those round-trip from the previous state (or from config, in the create/update paths).

### Decision 6: Schema-shape change is acceptable

Existing state files store `start` as a string. Dropping `Computed` does not change the on-disk representation; it changes plan/apply semantics: the value will no longer be sourced from the read response. Users who have an existing state file where `start` holds the ES-effective value (e.g. `02:26:42Z`) will see a one-time plan showing `start` reverting to the value in their config (or to null if omitted). This is benign because:

- The actual ES datafeed state is not touched — only what we record locally changes.
- A subsequent apply with `state = "started"` and the same `start` would no-op (REQ-008 short-circuit: current state == desired state, skip the API call).

We document this in CHANGELOG and treat it as a behavioural fix rather than a backwards-incompatible break that requires a state migration.

## Risks / Trade-offs

- **[Risk] Existing state files contain `start = <ES-effective-value>` that doesn't match the user's config** → One-time plan diff after upgrade. Mitigation: clear CHANGELOG entry; the apply is a no-op against ES because `state` doesn't change.

- **[Risk] Users were relying on `terraform output` reading `start` to see what ES is actually searching** → They must switch to `effective_search_start`. Mitigation: keep both attributes; document the rename clearly in the resource description and CHANGELOG.

- **[Risk] Import path regression** → Imports never set `start`/`end` from config, so on import the user gets `start = null`, `end = null`, `effective_search_start = <ES value>`, `effective_search_end = <ES value>`. A subsequent plan with a config that omits `start`/`end` is clean. With a config that supplies a `start`, plan will show drift (user's value vs the imported null). This is correct behaviour — they're declaring intent that wasn't there before. Mitigation: extend the import acceptance test to cover both shapes.

- **[Risk] `end` semantics asymmetry** → `end` was already `Optional`-only (not Computed). We are not changing its surface, only stopping `SetStartAndEndFromAPI` from writing to it. Configs that set `end` will now also round-trip unchanged (a hidden fix, since `Optional`-only attributes already enforce round-trip and the previous behavior could have errored). Mitigation: add a test for explicit `end`.

- **[Trade-off] Two extra attributes inflate the resource surface** → Net win: they replace a buggy implicit behaviour with an explicit, observable one.

## Migration Plan

1. Land the change behind a normal release (no feature flag; this is a bug fix to a broken behaviour).
2. CHANGELOG entry under "Fixed" calling out #2353 and the new `effective_search_*` attributes. Note the one-time plan diff for existing state with explicit `start`.
3. No state-upgrader is required because the underlying type of `start` does not change.
4. Rollback: revert the PR. State files remain valid; users go back to the existing bug.

## Open Questions

- Naming: `effective_search_start` vs `current_search_start` vs `running_search_start`. Recommendation: `effective_search_start` — it most accurately describes "what ES is actually using" without implying real-time tracking.
- Should we add a warning diagnostic on create when the user's `start` differs significantly from `effective_search_start` (e.g. bucket alignment shifted by more than `bucket_span`)? Recommendation: defer to a follow-up; out of scope for the bug fix.
