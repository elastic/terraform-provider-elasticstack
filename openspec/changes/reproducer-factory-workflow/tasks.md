## 1. Repository prerequisites

- [ ] 1.1 Create `internal/acctest/reproductions/` Go package with a `.gitkeep` or package stub so the fallback test-file path is valid
- [ ] 1.2 Verify or create the `reproducer-factory` GitHub label in the repository

## 2. Workflow source scaffolding

- [ ] 2.1 Create `.github/workflows-src/reproducer-factory-issue/scripts/` directory and `intake-constants.js` with `FACTORY_LABEL = 'reproducer-factory'`, `ISSUE_BRANCH_PREFIX = 'reproducer-factory/issue-'`, `DUPLICATE_LINKAGE_MODE = 'related-literal'`
- [ ] 2.2 Extend `.github/workflows-src/lib/factory-issue-shared.js` to support a `related-literal` duplicate-linkage mode: add a third branch to `bodyPattern` and `linkagePhrase` that matches `Related to #N` (analogous to the existing `closes-literal` branch for `Closes #N`)
- [ ] 2.3 Copy shared inline scripts from `research-factory-issue/scripts/` without modification: `check_actor_trust.inline.js`, `fetch_issue_comments.inline.js`, `fetch_live_issue.inline.js`, `finalize_gate.inline.js`, `qualify_trigger.inline.js`, `remove_trigger_label.inline.js`, `validate_dispatch_inputs.inline.js`
- [ ] 2.4 Copy `check_duplicate_pr.inline.js` from `code-factory-issue/scripts/` (uses `intake-constants.js` for branch prefix via shared lib — no change needed beyond the `intake-constants.js` values from 2.1)
- [ ] 2.5 Write `fetch_prior_reproducer_comment.inline.js` — adapts `research-factory-issue/scripts/fetch_prior_research_comment.inline.js` to search for the `<!-- gha-reproducer-factory -->` marker instead of `<!-- gha-research-factory -->`
- [ ] 2.6 Write `write_context_files.inline.js` — adapts the research-factory version to write to `/tmp/reproducer-factory-context/` and name the prior-comment file `prior_reproducer_comment.md`
- [ ] 2.7 Write `update_reproducer_comment.job.js` — adapts `research-factory-issue/scripts/update_research_comment.job.js` to use the `<!-- gha-reproducer-factory -->` marker and the `REPRODUCER_FACTORY_ISSUE_NUMBER` env var

## 3. Agent prompt (workflow.md.tmpl)

- [ ] 3.1 Write `workflow.md.tmpl` YAML frontmatter: name, 65-minute timeout, `on:` (issues + workflow_dispatch + status-comment), pre-activation job outputs, `tools:` (github: [issues, pull_requests, repos]), `network:` ([defaults, node, go, elastic.litellm-prod.ai, www.elastic.co]), `mcp-servers:` (elastic-docs), `checkout:` (fetch-depth: 0), `safe-outputs:` (`update-reproducer-comment` max 1 with step invoking `update_reproducer_comment.job.js`, `create-pull-request` max 1 with labels [reproducer-factory], `noop` max 1 report-as-issue false), pre-activation `jobs:` block (mirrors research-factory structure plus duplicate-PR check step from code-factory, outputs include `duplicate_pr_found` and `duplicate_pr_url`)
- [ ] 3.2 Write the agent prompt body (after the `---` separator): pre-activation context section, time-budget section (55-minute self-budget, hard-kill at 65 minutes, partial-output preference), test environment section (same `host.docker.internal` params as code-factory), Elastic documentation section (elastic-docs MCP instructions and fallback), task section covering the three-outcome decision tree, test file placement rules (resource-package when confidently identified vs `internal/acctest/reproductions/` fallback), outcome-A instructions (write test with `ExpectError`/`ExpectNonEmptyPlan`, run test, confirm pass, emit `update-reproducer-comment` + `create-pull-request`), outcome-B instructions (3 avenues, each must name a file path or Go symbol), outcome-C instructions (run without `ExpectError`, confirm clean pass, git archaeology, emit `update-reproducer-comment` only), pull-request contract section (branch `reproducer-factory/issue-{n}`, body includes `Related to #N` — not `Closes #N` — since the reproduction confirms the bug rather than resolving it), and guardrails section

## 4. Build and validation

- [ ] 4.1 Run `make workflow-generate` (or equivalent) to compile `reproducer-factory-issue.lock.yml` and verify the compiled file is produced without errors
- [ ] 4.2 Run `make check-openspec` (or `npx openspec validate`) to confirm the new specs pass structural validation
- [ ] 4.3 Run the existing workflow-lib test suite (`npm test` or `make test-workflows`) to confirm no regressions in shared library code
