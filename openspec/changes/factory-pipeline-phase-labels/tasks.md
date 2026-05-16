## 1. Add shared phase-label helper library

- [x] 1.1 Create `.github/workflows-src/lib/set-phase-label.js` exporting `setPhaseLabel({ github, context, issueNumber, phaseLabelName })`. The function SHALL:
  - Call `github.rest.issues.addLabels` to add `phaseLabelName` to the issue.
  - Call `github.rest.issues.listLabelsOnIssue` (per_page: 100) and collect all labels whose name starts with `phase-` and is not `phaseLabelName`.
  - Call `github.rest.issues.removeLabel` for each stale `phase-*` label found; treat 404 as success.
  - Return a result object with `phase_label_set: boolean`, `phase_label_name: string`, `stale_labels_removed: string[]`, and `reason: string`.
- [x] 1.2 Add unit tests for `set-phase-label.js` alongside the existing lib tests (e.g. `set-phase-label.test.mjs`), covering: label added successfully, stale labels removed, addLabels failure, removeLabel 404 treated as success, missing issueNumber guard.

## 2. Add per-factory inline scripts

- [x] 2.1 Create `.github/workflows-src/research-factory-issue/scripts/set_phase_label.inline.js` using `//include: ../../lib/set-phase-label.js`, invoking `setPhaseLabel` with `phaseLabelName: 'phase-research'`.
- [x] 2.2 Create `.github/workflows-src/reproducer-factory-issue/scripts/set_phase_label.inline.js` using `//include: ../../lib/set-phase-label.js`, invoking `setPhaseLabel` with `phaseLabelName: 'phase-reproduction'`.
- [x] 2.3 Create `.github/workflows-src/change-factory-issue/scripts/set_phase_label.inline.js` using `//include: ../../lib/set-phase-label.js`, invoking `setPhaseLabel` with `phaseLabelName: 'phase-specification'`.
- [x] 2.4 Create `.github/workflows-src/code-factory-issue/scripts/set_phase_label.inline.js` using `//include: ../../lib/set-phase-label.js`, invoking `setPhaseLabel` with `phaseLabelName: 'phase-coding'`.

Each inline script SHALL output `phase_label_set` and `phase_label_name` via `core.setOutput` and log the outcome via `core.info`.

## 3. Update factory workflow templates

For each factory, add a `set_phase_label` step immediately after the `remove_trigger_label` step in
the pre-activation `steps:` block. The step must use `actions/github-script@v9` with
`x-script-include:` pointing to the factory's `set_phase_label.inline.js` and set an appropriate
`if:` guard.

- [x] 3.1 Update `.github/workflows-src/change-factory-issue/workflow.md.tmpl`: add step with `if:` condition matching `remove_trigger_label` (`event_eligible == 'true' && actor_trusted == 'true' && duplicate_pr_found != 'true'`).
- [x] 3.2 Update `.github/workflows-src/research-factory-issue/workflow.md.tmpl`: add step with `if:` condition covering both intake modes — issue-event (eligible + trusted) and dispatch (dispatch eligible).
- [x] 3.3 Update `.github/workflows-src/reproducer-factory-issue/workflow.md.tmpl`: add step with `if:` condition covering both intake modes.
- [x] 3.4 Update `.github/workflows-src/code-factory-issue/workflow.md.tmpl`: add step with `if:` condition covering both intake modes.

Each step SHALL expose `phase_label_set` and `phase_label_name` as step outputs (mirroring the pattern of `remove_trigger_label` and `trigger_label_removed`/`trigger_label_removed_reason` outputs). The pre-activation job outputs block of each factory SHOULD expose these outputs if needed by the agent job condition or logging.

## 4. Recompile lock files

- [x] 4.1 Run the GH AW compiler (`gh aw compile`) for all four factory workflows and commit the updated `.lock.yml` files:
  - `.github/workflows/research-factory-issue.lock.yml`
  - `.github/workflows/reproducer-factory-issue.lock.yml`
  - `.github/workflows/change-factory-issue.lock.yml`
  - `.github/workflows/code-factory-issue.lock.yml`
- [x] 4.2 Verify that only the expected step additions appear in the compiled diff; no unintended template changes.

## 5. Label provisioning and documentation

- [ ] 5.1 Create the four `phase-*` labels in the GitHub repository manually (or document the step for the maintainer to execute):
  - `phase-research`
  - `phase-reproduction`
  - `phase-specification`
  - `phase-coding`
- [x] 5.2 Add a note to the relevant contributor documentation (e.g. a comment in each factory workflow source or a section in the contributing guide) listing the required `phase-*` labels and their purpose.

## 6. Validation

- [x] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate factory-pipeline-phase-labels --type change` and resolve any reported issues.
- [x] 6.2 Run `make build` if any non-workflow source files were modified.
- [ ] 6.3 Manually verify on a test issue: apply a factory label, confirm the correct `phase-*` label appears and any prior `phase-*` labels are removed.
