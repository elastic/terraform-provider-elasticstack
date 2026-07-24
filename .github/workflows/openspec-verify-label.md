---
imports: [shared/setup-dev.md]
name: OpenSpec verify (label)
description: >-
  When maintainers add the verify-openspec label to a pull request, verifies the PR against exactly
  one active OpenSpec change (added and/or modified files only), posts a PR review, and on APPROVE
  (only when the run is approval-eligible and archive/push is allowed for same-repository PRs)
  archives the change and pushes the result to the PR branch. Uses pull_request_target with a
  deterministic verify_label step; a deterministic script step removes verify-openspec after
  activation; the agent does not use remove-labels safe output.
on:
  pull_request_target:
    types: [labeled]
  permissions:
    issues: write
    pull-requests: write
    contents: read
  steps:
    - name: Checkout repository
      uses: actions/checkout@v7.0.0
      with:
        persist-credentials: false
        fetch-depth: 1
    - name: Verify trigger label
      id: verify_label
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/openspec-verify/verify-label.js');
          await fn({ github, context, core });
    - name: Remove trigger label
      id: remove_trigger_label
      if: steps.verify_label.outputs.label_verified == 'true'
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/openspec-verify/remove-trigger-label.js');
          await fn({ github, context, core });
    - name: Classify pull request and select active change
      id: classify_and_select
      if: steps.verify_label.outputs.label_verified == 'true'
      uses: actions/github-script@v9.0.0
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/openspec-verify/classify-and-select.js');
          await fn({ github, context, core });
    - name: Comment on ineligible PR
      id: comment_ineligible
      if: >-
        steps.verify_label.outputs.label_verified == 'true' &&
        steps.classify_and_select.outputs.selection_status == 'ineligible'
      uses: actions/github-script@v9.0.0
      env:
        SELECTION_REASON: ${{ steps.classify_and_select.outputs.selection_reason }}
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const fn = require('${{ github.workspace }}/.github/scripts/workflows/openspec-verify/comment-ineligible.js');
          await fn({ github, context, core });
if: >-
  needs.pre_activation.outputs.label_verified == 'true' &&
  needs.pre_activation.outputs.selection_status == 'eligible'
steps: []
engine:
  id: claude
  model: "llm-gateway/claude-opus-4-8"
  args:
    - "--effort"
    - "high"
  env:
    ANTHROPIC_BASE_URL: "https://elastic.litellm-prod.ai"
    ANTHROPIC_API_KEY: ${{ secrets.CLAUDE_LITELLM_PROXY_API_KEY }}
permissions:
  contents: read
  pull-requests: read
jobs:
  pre-activation:
    outputs:
      label_verified: ${{ steps.verify_label.outputs.label_verified }}
      label_verified_reason: ${{ steps.verify_label.outputs.label_verified_reason }}
      trigger_label_removed: ${{ steps.remove_trigger_label.outputs.trigger_label_removed }}
      trigger_label_removed_reason: ${{ steps.remove_trigger_label.outputs.trigger_label_removed_reason }}
      selection_status: ${{ steps.classify_and_select.outputs.selection_status }}
      selection_reason: ${{ steps.classify_and_select.outputs.selection_reason }}
      selected_change: ${{ steps.classify_and_select.outputs.selected_change }}
      review_disposition: ${{ steps.classify_and_select.outputs.review_disposition }}
      disposition_reason: ${{ steps.classify_and_select.outputs.disposition_reason }}
      archive_push_allowed: ${{ steps.classify_and_select.outputs.archive_push_allowed }}
      archive_push_allowed_reason: ${{ steps.classify_and_select.outputs.archive_push_allowed_reason }}
tools:
  cli-proxy: true
  github:
    mode: gh-proxy
    toolsets: [repos, pull_requests]
network:
  allowed: [defaults, node, go, elastic.litellm-prod.ai]
checkout:
  fetch-depth: 0
safe-outputs:
  create-pull-request-review-comment:
    max: 25
    target: triggering
    side: RIGHT
  noop:
    max: 1
    report-as-issue: false
  submit-pull-request-review:
    max: 1
    target: triggering
  push-to-pull-request-branch:
    target: triggering
    max: 1
---

# OpenSpec verify, archive, and clean up (label-gated)

You verify a pull request against **one** active OpenSpec change under `openspec/changes/<id>/`, following `.agents/skills/openspec-verify-change/SKILL.md`, submit a **single** pull request review, and **only** after an **APPROVE** review and when archive/push is allowed run `openspec archive` and push the branch. The **`verify-openspec`** label is removed by a deterministic script step before this agent runs; do **not** emit **`remove-labels`** safe outputs.

## Pre-activation context

Deterministic pre-activation steps have classified the pull request and selected the active change.

- **Selected change id**: `${{ needs.pre_activation.outputs.selected_change }}`
- **Review disposition** (do not infer from PR files): `${{ needs.pre_activation.outputs.review_disposition }}` — either `approval-eligible` or `comment-only`
- **Disposition reason** (authoritative; echo or paraphrase in the review body when relevant): ${{ needs.pre_activation.outputs.disposition_reason }}
- **Archive/push allowed**: `${{ needs.pre_activation.outputs.archive_push_allowed }}` — either `true` (same-repository PR) or `false` (fork PR)
- **Archive/push allowed reason**: ${{ needs.pre_activation.outputs.archive_push_allowed_reason }}
- **Gating**: already complete — the workflow reached this point only because exactly one active non-archive change was found with file statuses limited to `added` and `modified`. Do **not** re-inspect PR files, re-derive the change id, guess approval eligibility from the diff, or re-derive whether archive/push is allowed.

Use **`npx openspec`** for all OpenSpec CLI invocations.

## Verification (active change)

Let `<id>` be `${{ needs.pre_activation.outputs.selected_change }}`.

1. Run:

   - `npx openspec status --change "${{ needs.pre_activation.outputs.selected_change }}" --json`
   - `npx openspec instructions apply --change "${{ needs.pre_activation.outputs.selected_change }}" --json`

2. Read **`.agents/skills/openspec-verify-change/SKILL.md`** and perform verification **rooted at** `openspec/changes/<id>/` using the skill's steps (status / apply JSON for context files, completeness / correctness / coherence, **Issues by priority**: CRITICAL, WARNING, SUGGESTION, **Final assessment**).

Whilst reviewing the implementation, do not raise issues covered by CI. For example, do not:

- Consider syntactic correctness of code.
- Run tests.

Instead, check the results of Github actions runs for the PR.

## Structural allowlist and relevance

1. **Structurally in scope** (no per-file relevance classification required):

   - All paths under `openspec/changes/${{ needs.pre_activation.outputs.selected_change }}/`.
   - For each delta spec `openspec/changes/${{ needs.pre_activation.outputs.selected_change }}/specs/<capability>/spec.md`, the matching **`openspec/specs/<capability>/spec.md`** if it appears in the PR.

2. For **every other** changed file in the PR (outside the structural allowlist in step 3), read the diff and classify vs `openspec/changes/<id>/` artifacts (**proposal**, **design**, **tasks**, delta specs) as **`relevant`**, **`uncertain`**, or **`unassociated`**. This step covers **relevance classification only** for out-of-scope files: among those outcomes, **`unassociated`** is what blocks **APPROVE** on the relevance axis. It is **not** the full approval gate—**CRITICAL** issues from verification (steps 1–2) still block **APPROVE** per step 7. When unsure, prefer **`relevant`** or **`uncertain`**.

## Review body, inline comments, and decision

1. **Review body** must include:

   - Summary / scorecard from verification (**Issues by priority**).
   - **Out-of-scope / unassociated changes**: list **`unassociated`** files, summarize **`uncertain`**, note accepted **`relevant`** briefly.
   - When **`${{ needs.pre_activation.outputs.review_disposition }}`** is **`comment-only`** (net-new spec change material under the selected change): explain that the review is limited to **`COMMENT`** because it introduces a net-new spec change (added files under the active change), **including when the normal approval criteria are otherwise satisfied**. Do **not** imply the pull request met those criteria if verification reported **CRITICAL** issues; still describe the net-new **`COMMENT`** limitation. Tie this to the deterministic **Disposition reason** above.

2. Add **line-level** **`create-pull-request-review-comment`** entries for mappable CRITICAL (and other high-signal) issues and for **`unassociated`** hunks where the API allows; avoid spam on large **`relevant`** sets.

3. Submit **exactly one** **`submit-pull-request-review`** for this run:

   - Use **`APPROVE`** **if and only if** **`${{ needs.pre_activation.outputs.review_disposition }}`** is **`approval-eligible`** **and** there are **zero CRITICAL** issues and **zero `unassociated`** files.
   - Use **`COMMENT`** when **`${{ needs.pre_activation.outputs.review_disposition }}`** is **`comment-only`**, **including** when verification finds zero CRITICAL issues and zero **`unassociated`** files.
   - Otherwise (blocking issues / unassociated files) use **`COMMENT`**.
   - **Never** use **`REQUEST_CHANGES`**. WARNING and SUGGESTION **alone** do **not** block **`APPROVE`** for **`approval-eligible`** runs.
   - **`${{ needs.pre_activation.outputs.archive_push_allowed }}`** being `false` does **not** force **`COMMENT`** — the review decision is based solely on verification results and review disposition.

## Archive and push (APPROVE only, approval-eligible only, archive-push-allowed only)

1. **Only** if the review you submitted in step 7 used **`APPROVE`** **and** **`${{ needs.pre_activation.outputs.review_disposition }}`** is **`approval-eligible`** **and** **`${{ needs.pre_activation.outputs.archive_push_allowed }}`** is **`true`**:

   - Run **`npx openspec archive "${{ needs.pre_activation.outputs.selected_change }}" --yes`** (non-interactive; add `--skip-specs` only if the change is explicitly doc-only and repository policy allows — default is full archive).
   - If the working tree has changes, **commit** them with a clear message (e.g. `chore(openspec): archive ${{ needs.pre_activation.outputs.selected_change }} via verify-openspec`).
   - Use **`push-to-pull-request-branch`** to update the **triggering** PR branch.

2. If the review was **`COMMENT`**, or the run was **`comment-only`**, or **`${{ needs.pre_activation.outputs.archive_push_allowed }}`** is **`false`**, **do not** run `openspec archive`, **do not** commit for archive purposes, and **do not** call **`push-to-pull-request-branch`**.
