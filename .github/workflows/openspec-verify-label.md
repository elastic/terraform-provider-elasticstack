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
engine:
  id: copilot
  model: "gpt-5.4"
permissions:
  contents: read
  pull-requests: read
jobs:
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

## Trigger and first gate

1. The workflow is compiled so it runs the agent **only** when a `pull_request` **`labeled`** event applies the label **`verify-openspec`** (see frontmatter `names:`). **Before any other step**, if the injected event shows a different label name than `verify-openspec`, call **`noop`** and **stop** (defensive; should not occur when the compiler’s label gate is in effect).

## Pull request files and change selection

2. Load the pull request **changed files** list for the triggering PR (GitHub API: each entry has **path** and **status**: `added`, `modified`, `removed`, `renamed`, etc.).

3. Filter paths to those under `openspec/changes/` **excluding** `openspec/changes/archive/**`. For each such path, derive `<id>` as the **first path segment** after `openspec/changes/` (so `openspec/changes/foo/bar.md` → `foo`; skip if the segment is `archive`).

4. **Gating — call `noop` and stop** (no review, no archive, no push) if **any** of these hold:

   - **Any** non-archive path under `openspec/changes/` has status **`added`**.
   - **More than one** distinct `<id>` has **at least one** file with status **`modified`** among those paths.
   - **Zero** distinct `<id>` has **at least one** **`modified`** file under `openspec/changes/<id>/` (non-archive).
   - **Any** file under `openspec/changes/<id>/` (non-archive) that appears in the PR has a status **other than** **`modified`** (treat **`removed`**, **`renamed`**, and any other non-`modified` status as **noop**).

5. If gating passes, you have **exactly one** selected change id `<id>`. Record it and continue.

## Repository setup for OpenSpec

6. Check out the **full** repository at the PR head ref (the triggering pull request branch) with enough history for a clean working tree.

7. Install Node dependencies so OpenSpec CLI is available, using repository practice:

   - Run `npm ci` at the repository root (requires `network: node`).

8. Use **`npx openspec`** for all OpenSpec CLI invocations below.

## Verification (active change)

9. For the selected `<id>`, run:

   - `npx openspec status --change "<id>" --json`
   - `npx openspec instructions apply --change "<id>" --json`

10. Read **`.agents/skills/openspec-verify-change/SKILL.md`** and perform verification **rooted at** `openspec/changes/<id>/` using the skill’s steps (status / apply JSON for context files, completeness / correctness / coherence, **Issues by priority**: CRITICAL, WARNING, SUGGESTION, **Final assessment**).

## Structural allowlist and relevance

11. **Structurally in scope** (no per-file relevance classification required):

    - All paths under `openspec/changes/<id>/`.
    - For each delta spec `openspec/changes/<id>/specs/<capability>/spec.md`, the matching **`openspec/specs/<capability>/spec.md`** if it appears in the PR.

12. For **every other** changed file in the PR, read the diff and classify vs `openspec/changes/<id>/` artifacts (**proposal**, **design**, **tasks**, delta specs) as **`relevant`**, **`uncertain`**, or **`unassociated`**. Only **`unassociated`** blocks **APPROVE**; when unsure, prefer **`relevant`** or **`uncertain`**.

## Review body, inline comments, and decision

13. **Review body** must include:

    - Summary / scorecard from verification (**Issues by priority**).
    - **Out-of-scope / unassociated changes**: list **`unassociated`** files, summarize **`uncertain`**, note accepted **`relevant`** briefly.

14. Add **line-level** **`create-pull-request-review-comment`** entries for mappable CRITICAL (and other high-signal) issues and for **`unassociated`** hunks where the API allows; avoid spam on large **`relevant`** sets.

15. Submit **exactly one** **`submit-pull-request-review`** for this run:

    - Use **`APPROVE`** **if and only if** there are **zero CRITICAL** issues and **zero `unassociated`** files.
    - Otherwise use **`COMMENT`**.
    - **Never** use **`REQUEST_CHANGES`**. WARNING and SUGGESTION **alone** do **not** block APPROVE.

## Archive and push (APPROVE only)

16. **Only** if the review you submitted in step 15 used **`APPROVE`**:

    - Run **`npx openspec archive "<id>" --yes`** (non-interactive; add `--skip-specs` only if the change is explicitly doc-only and repository policy allows — default is full archive).
    - If the working tree has changes, **commit** them with a clear message (e.g. `chore(openspec): archive <id> via verify-openspec`).
    - Use **`push-to-pull-request-branch`** to update the **triggering** PR branch.

17. If the review was **`COMMENT`** (or you exited via **`noop`** earlier), **do not** run `openspec archive`, **do not** commit for archive purposes, and **do not** call **`push-to-pull-request-branch`**.

## Noop completion

18. Whenever you exit early with **`noop`**, include a **clear, short** message explaining which gate failed (wrong label, multiple change ids, added file under `openspec/changes/`, non-modified status, etc.).

