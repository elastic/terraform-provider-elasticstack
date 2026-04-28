## Context

Code Factory and Change Factory are AWF sources under `.github/workflows-src/`, compiled to `.github/workflows/`. They already use **`issues: types: [opened, labeled]`** and deterministic `qualify_trigger` / trust / duplicate gating. **OpenSpec verify (label)** (`.github/workflows-src/openspec-verify-label/workflow.md.tmpl`) already implements a two-step pattern in **`on.steps`**: verify, then **`Remove trigger label`** using `actions/github-script@v9` and `x-script-include: scripts/remove_trigger_label.inline.js`, which delegates to `.github/workflows-src/lib/remove-trigger-label.js` (today PR-specific and hard-coded to `verify-openspec`).

## Goals / Non-Goals

**Goals:**

- Leave the **existing `issues` triggers** unchanged.
- Add **`status-comment: true`** in `on:` per [status comments](https://github.github.com/gh-aw/reference/triggers/#status-comments-status-comment).
- Add a **remove factory label** pre-activation step **patterned on** openspec-verify-label (lines 24–30): same action, same `x-script-include` style, **maximal reuse** of `remove-trigger-label` by **generalizing** the library (label name + numeric issue target from `context.payload.issue`) and thin per-workflow inline scripts (or one shared include) that set outputs consistent with verify (`trigger_label_removed`, `trigger_label_removed_reason` or aligned names).
- Expose remove-step outputs on **`pre-activation` `jobs.pre-activation.outputs`** if downstream compiled YAML needs them (mirror verify workflow).

**Non-Goals:**

- Introducing `label_command` or changing eligibility rules in `factoryQualifyTriggerEvent` unless the new step’s `if:` conditions require trivial expression tweaks.
- Custom issue comments solely for workflow-run URLs (still covered by `status-comment`).

## Decisions

1. **`status-comment: true`** — unchanged from prior change direction; no duplicate run-link step in implementation `steps:`.

2. **Generalize `remove-trigger-label.js`** — add parameters (e.g. `labelName`, `issueNumber`) and implement removal via `github.rest.issues.removeLabel` for **issues** (PRs use the same API with `issue_number`). Keep 404-as-success behavior. **Update** `openspec-verify-label/scripts/remove_trigger_label.inline.js` to pass `verify-openspec` and PR number so verify behavior stays identical.

3. **When to remove the factory label** — Run the step only when the same logical gate as the **agent** would pass (eligible event, trusted actor, no duplicate linked PR), so untrusted or duplicate-suppressed runs do **not** strip the label. Express this with a composite `if:` on the step (and order steps so prerequisite outputs exist), or a single final pre-activation script—prefer matching existing step outputs (`qualify_trigger`, `check_actor_trust`, `check_duplicate_pr`) like the top-level agent `if:`.

4. **`on.permissions`** — Grant **`issues: write`** for pre-activation when the remove step runs (same class of permission as openspec-verify-label’s `on.permissions`).

## Risks / Trade-offs

- **[Risk] Skipped `check_duplicate_pr` leaves outputs unset** — composite `if:` for remove must handle skipped-step semantics. *Mitigation:* Match conditions to the duplicate step’s `if:` chain or consolidate gate outputs in `finalize_gate` and key remove off one output.
- **[Risk] Removing label on `issues.opened` with label present** — same as intentional one-shot: label is removed once the agent proceeds. *Mitigation:* Document for maintainers; align with product expectation.

## Migration Plan

1. Generalize `remove-trigger-label.js`; update verify inline script; add factory inline scripts and `on.steps` + permissions + `pre-activation` outputs.
2. `make workflow-generate`; run workflow lib tests.
3. Sync specs / archive change per repo process.

## Open Questions

- None blocking; exact output names for factory pre-activation can mirror verify or stay factory-prefixed as long as compiled AWF is valid.
