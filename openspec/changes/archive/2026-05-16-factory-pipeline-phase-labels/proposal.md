## Why

The `*-factory` trigger labels (`research-factory`, `reproducer-factory`, `change-factory`,
`code-factory`) act as one-shot command triggers: a factory workflow fires when the label is applied
and then removes the label. This leaves no persistent record of which pipeline stage an issue has
reached, making it impossible to filter or sort issues by pipeline phase at a glance.

## What Changes

- Add a new step in the pre-activation job of each of the four factory workflows that **sets** the
  factory's associated phase label and **removes all other `phase-*` labels** from the issue,
  so exactly one phase label is present at any time.
- Add a shared helper library `set-phase-label.js` under `.github/workflows-src/lib/` that
  encapsulates the GitHub Issues API calls for adding the phase label and removing stale `phase-*`
  labels.
- Add per-factory inline scripts `set_phase_label.inline.js` in each factory's `scripts/` directory
  that invoke the shared helper with the factory-specific phase label name.
- The step fires under the same eligibility guard as `remove_trigger_label` — but also covers
  `workflow_dispatch`-triggered runs for factories that support dispatch, since the issue remains in
  that pipeline phase regardless of how the factory is invoked.
- Document that the four `phase-*` labels (`phase-research`, `phase-reproduction`,
  `phase-specification`, `phase-coding`) must be created manually in the GitHub repository by a
  maintainer before the workflows can apply them.

| Phase label | Factory that sets it |
|---|---|
| `phase-research` | `research-factory` |
| `phase-reproduction` | `reproducer-factory` |
| `phase-specification` | `change-factory` |
| `phase-coding` | `code-factory` |

## Capabilities

### New Capabilities

- `ci-factory-pipeline-phase-labels`: Requirements for the shared phase-label helper and the new
  `set_phase_label` step in all four factory workflow pre-activation jobs.

### Modified Capabilities

- `ci-research-factory-issue-intake`: Gains `set_phase_label` step that applies `phase-research`.
- `ci-reproducer-factory-issue-intake`: Gains `set_phase_label` step that applies `phase-reproduction`.
- `ci-change-factory-issue-intake`: Gains `set_phase_label` step that applies `phase-specification`.
- `ci-code-factory-issue-intake`: Gains `set_phase_label` step that applies `phase-coding`.

## Impact

- **Workflow sources**: New shared lib, four new inline scripts, and four updated `workflow.md.tmpl`
  files; all four compiled `.lock.yml` files must be regenerated.
- **Label provisioning**: The four `phase-*` labels do not exist today and must be created manually
  in the GitHub repository before the workflows go live. A note in each factory's source or in the
  repository's contributor docs should record which labels are required.
- **No backward compatibility concern**: Existing in-flight issues will not be retroactively
  labelled; the change only affects future factory runs.
- **Phase labels persist on closure**: When an issue is closed, the phase label is not removed; it
  remains accurate if the issue is later reopened.
