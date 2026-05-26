## Context

The `openspec-verify-label.md` workflow uses a `pre_activation` job with injected steps (`verify_label`, `remove_trigger_label`, `classify_and_select`) to gate agent execution. When `classify_and_select` determines the PR is ineligible (e.g., no change files, multiple IDs, unsupported file status), the agent job is skipped silently. PR authors get no feedback.

The `pre_activation` job already holds `pull-requests: write` permission (needed for `remove_trigger_label`), so posting a comment requires no new permission grants.

The `classify-and-select.js` script already publishes `selection_reason` as a step output with a precise human-readable message for every ineligibility scenario.

## Goals / Non-Goals

**Goals:**
- Post a PR comment when `verify-openspec` is applied to an ineligible PR.
- Include the specific ineligibility reason from `classify_and_select`.
- Include "How to fix" remediation guidance in the comment so the author knows what to change.
- Follow the existing inject-steps pattern in `openspec-verify-label.md` frontmatter.

**Non-Goals:**
- Deduplication / idempotency — duplicates are acceptable; deduplication is a follow-on enhancement.
- Commenting when `label_verified == 'false'` (wrong label triggered the workflow — no user action needed).
- Changing eligibility rules in `select-change.js`.
- Modifying agent instructions for the already-eligible path.

## Decisions

**Use a deterministic inject step (Approach A).**
A new `comment_ineligible` step is added to the `.md` frontmatter `steps:` block after `classify_and_select`. This is consistent with `remove_trigger_label` — a sibling deterministic step that already performs a side-effect after classification. No extra job, no extra checkout, no extra runner cost.

Step `if:` condition:
```
steps.verify_label.outputs.label_verified == 'true' &&
steps.classify_and_select.outputs.selection_status == 'ineligible'
```

**New script: `comment-ineligible.js`.**
Pattern matches `remove-trigger-label.js`: a thin wrapper that imports a library function and calls it. The script reads `selection_reason` from step environment (via `core.getInput` or environment variable injection from the workflow step context) and calls `github.rest.issues.createComment` on `context.payload.pull_request.number`.

Alternative: inline script in the workflow YAML — rejected because sibling scripts are stored as files to allow unit testing and code review.

**Comment body includes reason + remediation guidance.**
The comment template:

```
**OpenSpec verify skipped** ⚠️

The \`verify-openspec\` label was applied but this PR is not eligible for verification:

> <selection_reason>

**How to fix**

For the PR to be verified, it must contain exactly one active OpenSpec change directory under \`openspec/changes/<id>/\` where:
- `<id>` is a single path segment (not `archive`).
- All changed files under that path have status `added` or `modified` only (no renames, deletes, etc.).
- No other OpenSpec change directories (non-archive) appear in the PR.

See the [OpenSpec authoring guide](../../dev-docs/high-level/openspec-requirements.md) for details.
```

**Unit tests.**
Add `comment-ineligible.test.mjs` covering: comment posted with correct body, step skipped when `label_verified` is false, step skipped when `selection_status` is eligible. Pattern matches `classify-and-select.test.mjs` and `remove-trigger-label.test.mjs`.

## Risks / Trade-offs

- Repeated label applications produce repeated identical comments — accepted per the human direction: "Nope, duplicates are fine."
- The `comment_ineligible` step runs in `pre_activation`, adding a small GitHub API call on every ineligible run; cost is minimal and only triggered when the PR is ineligible.
- `selection_reason` is passed between steps as a job-output string; size is bounded (all known reason strings are short).

## Migration Plan

1. Add `comment-ineligible.js` script in `.github/scripts/workflows/openspec-verify/`.
2. Add `comment-ineligible.test.mjs` unit tests.
3. Update `.github/workflows/openspec-verify-label.md` to add the `comment_ineligible` inject step after `classify_and_select`.
4. Recompile `.github/workflows/openspec-verify-label.lock.yml` with `gh aw compile`.
5. Update `ci-aw-openspec-verification` delta spec with the new requirement.

## Open Questions

_(None — all open questions from the research comment were resolved via human direction.)_
