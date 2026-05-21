## 1. Add the `comment-ineligible` script

- [ ] 1.1 Create `.github/scripts/workflows/openspec-verify/comment-ineligible.js` that reads `selection_reason` from the step environment (available as the `SELECTION_REASON` env var passed from the workflow step) and calls `github.rest.issues.createComment` on the triggering PR number with a comment body that includes the ineligibility reason and "How to fix" remediation guidance.
- [ ] 1.2 Create `.github/scripts/workflows/openspec-verify/comment-ineligible.test.mjs` with unit tests covering: (a) comment is posted with the correct body when `selection_status` is `ineligible` and `label_verified` is `true`, (b) the function short-circuits gracefully when the PR number is absent, (c) the comment body includes the `selection_reason` string verbatim.

## 2. Update the workflow source

- [ ] 2.1 In `.github/workflows/openspec-verify-label.md`, add a new inject step `comment_ineligible` to the frontmatter `steps:` block immediately after `classify_and_select`:
  ```yaml
  - name: Comment on ineligible PR
    id: comment_ineligible
    if: >-
      steps.verify_label.outputs.label_verified == 'true' &&
      steps.classify_and_select.outputs.selection_status == 'ineligible'
    uses: actions/github-script@v9
    env:
      SELECTION_REASON: ${{ steps.classify_and_select.outputs.selection_reason }}
    with:
      github-token: ${{ secrets.GITHUB_TOKEN }}
      script: |
        const fn = require('${{ github.workspace }}/.github/scripts/workflows/openspec-verify/comment-ineligible.js');
        await fn({ github, context, core });
  ```

## 3. Recompile the lock file

- [ ] 3.1 Run `gh aw compile .github/workflows/openspec-verify-label.md` and commit the updated `.github/workflows/openspec-verify-label.lock.yml`.

## 4. Update the delta spec

- [ ] 4.1 Ensure `openspec/changes/verify-openspec-ineligible-pr-comment/specs/ci-aw-openspec-verification/spec.md` is aligned with the requirement added in this change (see delta spec).

## 5. Validate

- [ ] 5.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate verify-openspec-ineligible-pr-comment --type change` and resolve any issues.
- [ ] 5.2 Run the unit test suite for the new script: `node --test .github/scripts/workflows/openspec-verify/comment-ineligible.test.mjs`.
