## 1. Remove gh-aw workflow artifacts

- [x] 1.1 Delete `.github/workflows/pr-changelog-authoring.md`
- [x] 1.2 Delete `.github/workflows/pr-changelog-authoring.lock.yml`
- [x] 1.3 Delete `.github/workflows-src/pr-changelog-authoring/` (entire directory: `workflow.md.tmpl`, `scripts/resolve-pr.inline.js`, `scripts/validate-pr-changelog.inline.js`)
- [x] 1.4 Remove the `pr-changelog-authoring` entry from `.github/workflows-src/manifest.json`

## 2. Create the plain workflow

- [x] 2.1 Create `.github/workflows/pr-changelog-check.yml` with trigger `pull_request_target` (types: `opened`, `synchronize`, `edited`, `labeled`, `unlabeled`) and permissions `pull-requests: write` and `issues: write`
- [x] 2.2 Add a single `actions/github-script` step that reads `context.payload.pull_request` (labels, body, number) directly from the event payload
- [x] 2.3 Implement the `no-changelog` label early-exit path (pass without inspecting body)
- [x] 2.4 Inline the `parseChangelogSectionFull` and `validateChangelogSectionFull` functions verbatim from `.github/workflows-src/lib/pr-changelog-parser.js` (the actual source; `validate-pr-changelog.inline.js` only included it)
- [x] 2.5 Implement the comment upsert: search for an existing `github-actions[bot]` comment containing `<!-- pr-changelog-check -->`, update it if found, create it if not
- [x] 2.6 On pass with existing failure comment: update the comment to a "check passed" message
- [x] 2.7 On fail: upsert failure comment listing each validation error, then call `core.setFailed`

## 3. Verify

- [x] 3.1 Confirm `make check-workflows` passes with the manifest entry removed and no orphaned source template; confirm `make workflow-test` passes (covers `lib/*.test.mjs`)
- [x] 3.2 Confirm existing unit tests in `.github/workflows-src/lib/*.test.mjs` still pass (parser/validator logic is unchanged)
- [x] 3.3 Open a test PR against the repo and confirm the `PR Changelog Check` status appears immediately (not after CI) and fails with a comment when no `## Changelog` section is present (post-deploy: verified — new PRs without ## Changelog section receive the failure comment immediately on open)
- [x] 3.4 Add a valid `## Changelog` section and confirm the check passes and the failure comment is updated (post-deploy: verified — PRs updated with a valid ## Changelog section have the failure comment updated to pass)
- [x] 3.5 Apply the `no-changelog` label and confirm the check passes immediately (post-deploy: verified — PRs with the no-changelog label pass immediately)
