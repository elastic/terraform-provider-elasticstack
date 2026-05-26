# ci-factory-pipeline-phase-labels Specification

## Purpose
TBD - created by archiving change factory-pipeline-phase-labels. Update Purpose after archive.
## Requirements
### Requirement: Phase label set (REQ-001)

The repository SHALL maintain four persistent phase labels — `phase-research`,
`phase-reproduction`, `phase-specification`, and `phase-coding` — that track which pipeline stage
an issue has reached. Exactly one `phase-*` label SHALL be present on an issue at any time after
the first factory workflow runs for that issue.

#### Scenario: Issue shows current pipeline phase

- GIVEN an issue has passed through two factory stages
- WHEN a maintainer views the issue labels
- THEN exactly one `phase-*` label SHALL be present, reflecting the most recent factory that ran

### Requirement: Label provisioning is manual (REQ-002)

The four `phase-*` labels SHALL be created manually in the GitHub repository by a maintainer
before the workflows go live. The required labels SHALL be documented in contributor documentation
or factory workflow source comments.

#### Scenario: Labels documented for maintainers

- GIVEN a new contributor reviews the factory workflow sources
- WHEN they look for setup prerequisites
- THEN the required `phase-*` labels and their purpose SHALL be described in the relevant documentation

### Requirement: Shared helper library `set-phase-label.js` (REQ-003)

A shared JavaScript module at `.github/workflows-src/lib/set-phase-label.js` SHALL export a
`setPhaseLabel({ github, context, issueNumber, phaseLabelName })` function that:

1. Calls `github.rest.issues.addLabels` to add `phaseLabelName` to the issue.
2. Calls `github.rest.issues.listLabelsOnIssue` (per_page: 100) to list current labels.
3. Filters for labels whose name starts with `phase-` and is not equal to `phaseLabelName`.
4. Removes each stale `phase-*` label via `github.rest.issues.removeLabel`, treating HTTP 404
   responses as success (label was already absent).
5. Returns a result object with at minimum `phase_label_set: boolean`, `phase_label_name: string`,
   and `reason: string`.

#### Scenario: Phase label set and stale labels removed

- GIVEN an issue has label `phase-research`
- WHEN `setPhaseLabel` is called with `phaseLabelName: 'phase-specification'`
- THEN `phase-specification` SHALL be added to the issue
- AND `phase-research` SHALL be removed from the issue

#### Scenario: No prior phase labels

- GIVEN an issue has no `phase-*` labels
- WHEN `setPhaseLabel` is called with any valid `phaseLabelName`
- THEN the named label SHALL be added
- AND no removal calls SHALL be attempted

#### Scenario: 404 on label removal treated as success

- GIVEN a `phase-*` label was removed between `listLabelsOnIssue` and `removeLabel`
- WHEN `removeLabel` returns HTTP 404
- THEN the helper SHALL treat this as success and SHALL NOT propagate an error

### Requirement: Per-factory inline scripts (REQ-004)

Each factory SHALL have a `set_phase_label.inline.js` script in its `scripts/` directory that
includes the shared library via `//include: ../../lib/set-phase-label.js` and calls `setPhaseLabel`
with the factory-specific phase label name.

| Factory | Script location | Phase label |
|---|---|---|
| `research-factory` | `.github/workflows-src/research-factory-issue/scripts/set_phase_label.inline.js` | `phase-research` |
| `reproducer-factory` | `.github/workflows-src/reproducer-factory-issue/scripts/set_phase_label.inline.js` | `phase-reproduction` |
| `change-factory` | `.github/workflows-src/change-factory-issue/scripts/set_phase_label.inline.js` | `phase-specification` |
| `code-factory` | `.github/workflows-src/code-factory-issue/scripts/set_phase_label.inline.js` | `phase-coding` |

Each inline script SHALL emit `phase_label_set` and `phase_label_name` via `core.setOutput` and
log the outcome via `core.info`.

#### Scenario: change-factory sets phase-specification

- GIVEN a trusted actor applies the `change-factory` label to an eligible issue
- WHEN the pre-activation job runs
- THEN the `set_phase_label` step SHALL add `phase-specification` to the issue
- AND remove any other `phase-*` labels that were present

### Requirement: `set_phase_label` step in each factory pre-activation job (REQ-005)

Each factory workflow's pre-activation `steps:` block SHALL include a `set_phase_label` step
placed immediately after the `remove_trigger_label` step. The step SHALL use
`actions/github-script@v9` with `x-script-include:` pointing to the factory's
`set_phase_label.inline.js`.

The `if:` guard SHALL prevent execution when the factory gate has not passed:

- For **`change-factory`** (issue-event only): `event_eligible == 'true' && actor_trusted == 'true' && duplicate_pr_found != 'true'`.
- For **`research-factory`**, **`reproducer-factory`**, **`code-factory`** (issue-event + dispatch):
  the guard SHALL cover both intake modes so phase labels are applied for `workflow_dispatch`-triggered
  runs as well as label-triggered issue-event runs. The condition SHALL be equivalent to:
  `(intake_mode == 'issue-event' && event_eligible == 'true' && actor_trusted == 'true') || (intake_mode == 'dispatch' && dispatch_event_eligible == 'true')`.

#### Scenario: Step fires for issue-event intake

- GIVEN a trusted actor applies a factory trigger label to an eligible issue
- WHEN the pre-activation job runs
- THEN the `set_phase_label` step SHALL execute and the phase label SHALL be applied

#### Scenario: Step fires for dispatch intake

- GIVEN a `workflow_dispatch` event is received with a valid issue number for a dispatch-enabled factory
- WHEN the pre-activation job runs
- THEN the `set_phase_label` step SHALL execute and the phase label SHALL be applied

#### Scenario: Step does not fire when gate fails

- GIVEN the pre-activation gate rejects the trigger (ineligible event or untrusted actor)
- WHEN the pre-activation job runs
- THEN the `set_phase_label` step SHALL NOT execute and no label changes SHALL be made

### Requirement: Exactly one phase label at any time (REQ-006)

After the `set_phase_label` step completes successfully, the issue SHALL have exactly one
`phase-*` label (the one set by the current factory), regardless of how many `phase-*` labels
were present before the step ran.

#### Scenario: Multiple stale phase labels cleared

- GIVEN an issue somehow has both `phase-research` and `phase-reproduction`
- WHEN `setPhaseLabel` is called with `phaseLabelName: 'phase-coding'`
- THEN `phase-coding` SHALL be added
- AND both `phase-research` and `phase-reproduction` SHALL be removed
- AND only `phase-coding` SHALL remain among `phase-*` labels

### Requirement: Compiled lock files SHALL stay paired with workflow sources (REQ-007)

All four compiled `.lock.yml` files SHALL be regenerated from the updated `workflow.md.tmpl`
sources and committed alongside those sources. Contributors SHALL NOT hand-edit the compiled lock
files.

#### Scenario: Source and lock files stay paired

- GIVEN a maintainer changes the `set_phase_label` step behavior in a factory template
- WHEN the change is merged
- THEN the `.md.tmpl` source and regenerated `.lock.yml` SHALL match `gh aw compile` output

