## Context

The repository's four factory workflows (`research-factory`, `reproducer-factory`,
`change-factory`, `code-factory`) each have a pre-activation job that runs deterministic
repository-authored logic before the sandboxed agent job. The pre-activation job already holds
`issues: write` permission and uses `github.rest.*` API calls (via `actions/github-script`) to
remove the trigger label once the gate passes. The new phase-label step mirrors this existing
pattern exactly.

The existing shared helper `.github/workflows-src/lib/remove-trigger-label.js` removes a single
named label. A companion helper `set-phase-label.js` will add the new phase label and remove all
other `phase-*` labels from the issue using `github.rest.issues.addLabels`,
`github.rest.issues.listLabelsOnIssue`, and `github.rest.issues.removeLabel`.

## Goals / Non-Goals

**Goals:**

- Apply a persistent `phase-*` label as soon as a factory gate passes, so the issue's pipeline
  position is visible to anyone inspecting its labels.
- Clear stale `phase-*` labels from prior pipeline stages so exactly one phase label is present.
- Support both `issue-event` and `workflow_dispatch` intake modes where applicable — the issue is
  in the phase regardless of how the factory was triggered.
- Keep the implementation deterministic, inside the pre-activation job, and independent of agent
  success or failure.

**Non-Goals:**

- Removing the `phase-*` label when an issue is closed or its factory PR is merged — the label
  remains accurate for potential reopening and is not a source of noise per maintainer guidance.
- Retroactively labelling existing in-flight issues.
- Auto-provisioning the four `phase-*` labels in the GitHub repository.
- Modifying the `*-factory` trigger-label behavior (they continue as one-shot command triggers).
- Changes to `issue-classifier` triage labels (`needs-research`, `needs-reproduction`,
  `needs-spec`, `needs-human`).

## Decisions

### 1. Pre-activation step (Approach A)

Phase labels are pipeline infrastructure signals, not agent outputs. They must be set as soon as
the factory gate passes, in the same pre-activation job that removes the trigger label. This is
deterministic, immediate, and unaffected by agent behavior. Approach B (agent `add_labels`
safe-output) was evaluated and rejected: it requires agent completion to fire, does not cleanly
handle stale label removal, and adds fragility if the agent errors or omits the call.

### 2. Shared library `set-phase-label.js`

The implementation follows the pattern of `remove-trigger-label.js`: a shared JS module under
`.github/workflows-src/lib/` that each factory includes via `//include:`. The module provides a
single exported function `setPhaseLabel({ github, context, issueNumber, phaseLabelName })` that:
1. Calls `github.rest.issues.addLabels` to add the factory's `phase-*` label.
2. Calls `github.rest.issues.listLabelsOnIssue` and filters for labels matching the `phase-*`
   prefix that are **not** the newly set label.
3. Removes each stale `phase-*` label via `github.rest.issues.removeLabel`.

### 3. Step placement and `if:` guard

The `set_phase_label` step is placed immediately after `remove_trigger_label` in each factory's
pre-activation step list. The guard condition matches or widens the `remove_trigger_label` guard:

- **`change-factory`** (issue-event only): same guard as its existing `remove_trigger_label` step
  (`event_eligible == 'true' && actor_trusted == 'true' && duplicate_pr_found != 'true'`).
- **`research-factory`**, **`reproducer-factory`**, **`code-factory`** (issue-event + dispatch):
  the guard covers both intake modes, since the human direction confirmed that phase labels should
  be applied for dispatch-triggered runs too. The step runs when the normalized eligibility
  conditions pass for either mode.

### 4. Dispatch mode applies phase labels

For factories that support `workflow_dispatch`, the `set_phase_label` step fires for dispatch
runs as well as label-triggered issue-event runs. The trigger label removal is intentionally
skipped in dispatch mode (there is no label to remove), but the phase label should be set because
the issue is still at that pipeline stage.

### 5. Label provisioning is manual

The four `phase-*` labels must exist in the GitHub repository before the workflows can apply
them. They will be created manually by a maintainer. The change should note the required labels in
the relevant contributor documentation. Auto-provisioning is explicitly out of scope.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Phase label applied even if gate rejects | Guarded by the same eligibility `if:` condition as `remove_trigger_label`. |
| `phase-*` labels not yet created in repo | Document required labels; `addLabels` will return a 422 if absent, visible in workflow logs. |
| `listLabelsOnIssue` pagination | Default page is 100 labels; an issue is unlikely to have more. Implement with page size 100, log a warning if full page returned. |
| `removeLabel` 404 on already-absent label | Treat 404 as success (same pattern as `remove-trigger-label.js`). |

## Migration Plan

1. Add `set-phase-label.js` to `.github/workflows-src/lib/`.
2. Add `set_phase_label.inline.js` to each factory's `scripts/` directory.
3. Update each factory's `workflow.md.tmpl` to include the `set_phase_label` step after `remove_trigger_label`.
4. Recompile all four compiled `.lock.yml` files.
5. Create the four `phase-*` labels manually in the GitHub repository.
6. Note required labels in contributor documentation or the factory workflow source comments.

**Rollback**: Remove the `set_phase_label` step from each template, recompile. The four `phase-*`
labels can remain in the repository without impact (they will simply not be applied).

## Open Questions

The following questions were raised during research. Resolutions based on maintainer direction
(issue #2812 comments) are documented here for implementer reference.

- **Label provisioning**: How should the `phase-*` labels be created?  
  _Resolution_: Manual step by a maintainer. Note the required labels in the change documentation.

- **Dispatch mode**: Should phase labels be applied for `workflow_dispatch`-triggered factory runs?  
  _Resolution_: Yes. The issue is still in that phase regardless of how the factory is triggered.

- **Phase on closure**: Should `phase-*` labels be removed when an issue is closed?  
  _Resolution_: No. The label remains accurate if the issue is reopened.

- **Back-population**: Should existing in-flight issues be retroactively labelled?  
  _Resolution_: No. Future runs only.

- **Label colour scheme**: Should the four phase labels share a consistent colour family?  
  _Resolution_: Manually created; colour is at the maintainer's discretion and is not part of this change.
