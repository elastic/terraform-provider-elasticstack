---
name: OpenSpec verify (label)
description: >-
  When maintainers add the verify-openspec label, verifies the PR against exactly one active
  OpenSpec change (modified-only), posts a PR review, on APPROVE archives the change and pushes
  the result to the PR branch, and removes the trigger label before the workflow fully completes.
on:
  pull_request:
    types: [labeled]
    # Hard gate: compiled workflow skips agent jobs unless this label triggered the run.
    names: [verify-openspec]
  steps:
    - name: Verify triggering label
      id: verify_label
      uses: actions/github-script@v8
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const label = context.payload.label?.name;
          if (label !== 'verify-openspec') {
            core.setOutput('label_verified', 'false');
            core.setOutput('label_reason', `Unexpected label: ${label || '(none)'}`);
            core.info(`Label check failed: expected verify-openspec, got ${label || '(none)'}`);
          } else {
            core.setOutput('label_verified', 'true');
            core.setOutput('label_reason', 'Label verified: verify-openspec');
            core.info('Label verified: verify-openspec');
          }
    - name: Select active change from PR files
      id: select_change
      uses: actions/github-script@v8
      with:
        github-token: ${{ secrets.GITHUB_TOKEN }}
        script: |
          const prNumber = context.payload.pull_request?.number;
          if (!prNumber) {
            core.setOutput('selection_status', 'ineligible');
            core.setOutput('selection_reason', 'No pull request number in event payload');
            core.setOutput('selected_change', '');
            return;
          }

          const files = await github.paginate(github.rest.pulls.listFiles, {
            owner: context.repo.owner,
            repo: context.repo.repo,
            pull_number: prNumber,
            per_page: 100,
          });

          const changePattern = /^openspec\/changes\/([^\/]+)\/.+$/;
          const archivePattern = /^openspec\/changes\/archive\//;

          const relevantFiles = files.filter(
            f => changePattern.test(f.filename) && !archivePattern.test(f.filename)
          );

          if (relevantFiles.length === 0) {
            core.setOutput('selection_status', 'ineligible');
            core.setOutput('selection_reason', 'No files under openspec/changes/ (non-archive) found in this PR');
            core.setOutput('selected_change', '');
            return;
          }

          const addedFiles = relevantFiles.filter(f => f.status === 'added');
          if (addedFiles.length > 0) {
            core.setOutput('selection_status', 'ineligible');
            core.setOutput('selection_reason', `Added file(s) under openspec/changes/: ${addedFiles.map(f => f.filename).join(', ')}`);
            core.setOutput('selected_change', '');
            return;
          }

          const nonModifiedFiles = relevantFiles.filter(f => f.status !== 'modified');
          if (nonModifiedFiles.length > 0) {
            core.setOutput('selection_status', 'ineligible');
            core.setOutput('selection_reason', `Non-modified file(s) under openspec/changes/: ${nonModifiedFiles.map(f => `${f.filename} (${f.status})`).join(', ')}`);
            core.setOutput('selected_change', '');
            return;
          }

          const modifiedIds = new Set(
            relevantFiles
              .filter(f => f.status === 'modified')
              .map(f => f.filename.match(changePattern)[1])
          );

          if (modifiedIds.size === 0) {
            core.setOutput('selection_status', 'ineligible');
            core.setOutput('selection_reason', 'No active change id with a modified file found');
            core.setOutput('selected_change', '');
            return;
          }

          if (modifiedIds.size > 1) {
            core.setOutput('selection_status', 'ineligible');
            core.setOutput('selection_reason', `Multiple active change ids with modified files: ${[...modifiedIds].join(', ')}`);
            core.setOutput('selected_change', '');
            return;
          }

          const selectedChange = [...modifiedIds][0];
          core.setOutput('selection_status', 'eligible');
          core.setOutput('selection_reason', `Selected change: ${selectedChange}`);
          core.setOutput('selected_change', selectedChange);
          core.info(`Selected active change: ${selectedChange}`);
    - name: Gate — skip agent when run is ineligible
      id: gate
      if: >-
        steps.verify_label.outputs.label_verified != 'true' ||
        steps.select_change.outputs.selection_status != 'eligible'
      uses: actions/github-script@v8
      with:
        script: |
          const reason = '${{ steps.verify_label.outputs.label_verified }}' !== 'true'
            ? '${{ steps.verify_label.outputs.label_reason }}'
            : '${{ steps.select_change.outputs.selection_reason }}';
          core.info(`Run is ineligible — skipping agent: ${reason}`);
          core.setFailed(`Ineligible: ${reason}`);
engine:
  id: copilot
  model: "gpt-5.4"
permissions:
  contents: read
  pull-requests: read
jobs:
  pre_activation:
    outputs:
      label_verified: ${{ steps.verify_label.outputs.label_verified }}
      label_reason: ${{ steps.verify_label.outputs.label_reason }}
      selected_change: ${{ steps.select_change.outputs.selected_change }}
      selection_status: ${{ steps.select_change.outputs.selection_status }}
      selection_reason: ${{ steps.select_change.outputs.selection_reason }}
  completion_cleanup:
    name: Remove verify-openspec label
    needs:
      - pre_activation
      - activation
      - agent
      - safe_outputs
    if: >-
      always() && github.event_name == 'pull_request' && github.event.action == 'labeled' &&
      github.event.label.name == 'verify-openspec'
    runs-on: ubuntu-slim
    permissions:
      issues: write
    steps:
      - name: Remove verify-openspec label
        uses: actions/github-script@v8
        with:
          github-token: ${{ secrets.GH_AW_GITHUB_TOKEN || secrets.GITHUB_TOKEN }}
          script: |
            const issueNumber = context.payload.pull_request?.number;
            if (!issueNumber) {
              core.info('No pull request number found in the event payload; skipping cleanup.');
              return;
            }

            try {
              await github.rest.issues.removeLabel({
                owner: context.repo.owner,
                repo: context.repo.repo,
                issue_number: issueNumber,
                name: 'verify-openspec',
              });
              core.info('Removed verify-openspec from the triggering pull request.');
            } catch (error) {
              if (error.status === 404) {
                core.info('verify-openspec was already absent on the triggering pull request.');
                return;
              }

              throw error;
            }
tools:
  github:
    toolsets: [repos, pull_requests]
network:
  allowed: [defaults, node]
checkout:
  fetch-depth: 0
# runtimes configures the agent environment. go.version must be kept in sync with go.mod
# because the gh-aw runtimes block does not support file-based version resolution.
# Use `make sync-go-runtime` to update this value and regenerate the compiled workflow,
# and `make check-go-runtime` (run by `make check-lint`) to detect drift.
runtimes:
  go:
    version: "1.26.1"
steps:
  - uses: actions/setup-node@53b83947a5a98c8d113130e565377fae1a50d02f # v6
    with:
      node-version-file: package.json
      cache: npm
      cache-dependency-path: package-lock.json
  # actions/setup-go configures Go in the runner environment for dependency installation
  # and repository bootstrap (make setup). The agent environment uses runtimes.go.version above.
  - uses: actions/setup-go@4b73464bb391d4059bd26b0524d20df3927bd417 # v6
    with:
      go-version-file: go.mod
      cache: true
      cache-dependency-path: go.sum
  - uses: hashicorp/setup-terraform@5e8dbf3c6d9deaf4193ca7a8fb23f2ac83bb6c85 # v4.0.0
    with:
      terraform_wrapper: false
  - name: Setup repository dependencies
    run: make setup
  - name: Install Node dependencies
    run: npm ci
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

You verify a pull request against **one** active OpenSpec change under `openspec/changes/<id>/`, following `.agents/skills/openspec-verify-change/SKILL.md`, submit a **single** pull request review, and **only** after an **APPROVE** review run `openspec archive` and push the branch.

## Pre-activation context

Deterministic pre-activation steps have verified the triggering label and selected the active change. A deterministic setup step has installed Node dependencies before agent reasoning begins.

- **Selected change id**: `${{ needs.pre_activation.outputs.selected_change }}`
- **Gating**: already complete — the workflow reached this point only because exactly one active change with modified-only files was found. Do **not** re-inspect PR files or re-derive the change id.

Use **`npx openspec`** for all OpenSpec CLI invocations.

## Verification (active change)

Let `<id>` be `${{ needs.pre_activation.outputs.selected_change }}`.

1. Run:

   - `npx openspec status --change "${{ needs.pre_activation.outputs.selected_change }}" --json`
   - `npx openspec instructions apply --change "${{ needs.pre_activation.outputs.selected_change }}" --json`

2. Read **`.agents/skills/openspec-verify-change/SKILL.md`** and perform verification **rooted at** `openspec/changes/<id>/` using the skill's steps (status / apply JSON for context files, completeness / correctness / coherence, **Issues by priority**: CRITICAL, WARNING, SUGGESTION, **Final assessment**).

## Structural allowlist and relevance

3. **Structurally in scope** (no per-file relevance classification required):

   - All paths under `openspec/changes/${{ needs.pre_activation.outputs.selected_change }}/`.
   - For each delta spec `openspec/changes/${{ needs.pre_activation.outputs.selected_change }}/specs/<capability>/spec.md`, the matching **`openspec/specs/<capability>/spec.md`** if it appears in the PR.

4. For **every other** changed file in the PR, read the diff and classify vs `openspec/changes/<id>/` artifacts (**proposal**, **design**, **tasks**, delta specs) as **`relevant`**, **`uncertain`**, or **`unassociated`**. Only **`unassociated`** blocks **APPROVE**; when unsure, prefer **`relevant`** or **`uncertain`**.

## Review body, inline comments, and decision

5. **Review body** must include:

   - Summary / scorecard from verification (**Issues by priority**).
   - **Out-of-scope / unassociated changes**: list **`unassociated`** files, summarize **`uncertain`**, note accepted **`relevant`** briefly.

6. Add **line-level** **`create-pull-request-review-comment`** entries for mappable CRITICAL (and other high-signal) issues and for **`unassociated`** hunks where the API allows; avoid spam on large **`relevant`** sets.

7. Submit **exactly one** **`submit-pull-request-review`** for this run:

   - Use **`APPROVE`** **if and only if** there are **zero CRITICAL** issues and **zero `unassociated`** files.
   - Otherwise use **`COMMENT`**.
   - **Never** use **`REQUEST_CHANGES`**. WARNING and SUGGESTION **alone** do **not** block APPROVE.

## Archive and push (APPROVE only)

8. **Only** if the review you submitted in step 7 used **`APPROVE`**:

   - Run **`npx openspec archive "${{ needs.pre_activation.outputs.selected_change }}" --yes`** (non-interactive; add `--skip-specs` only if the change is explicitly doc-only and repository policy allows — default is full archive).
   - If the working tree has changes, **commit** them with a clear message (e.g. `chore(openspec): archive ${{ needs.pre_activation.outputs.selected_change }} via verify-openspec`).
   - Use **`push-to-pull-request-branch`** to update the **triggering** PR branch.

9. If the review was **`COMMENT`**, **do not** run `openspec archive`, **do not** commit for archive purposes, and **do not** call **`push-to-pull-request-branch`**.
