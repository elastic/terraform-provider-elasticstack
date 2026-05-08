## 1. Shared library: comment-history capture

- [x] 1.1 Add a `factoryFetchIssueComments` (or similarly named) helper to `.github/workflows-src/lib/factory-issue-shared.js` that paginates `github.rest.issues.listComments`, returns chronological human-authored comments only (excluding `github-actions[bot]`, `dependabot[bot]`, and any sender whose login ends with `[bot]`), and caps results at a sensible upper bound (e.g. 200 comments) with a truncation flag in the output.
- [x] 1.2 Add unit tests for the helper in `.github/workflows-src/lib/factory-issue-shared.test.mjs` (or a new test file) covering: empty comment list, only-bot comments, mixed bot+human comments, pagination across pages, truncation behavior at the cap.
- [x] 1.3 Export the helper from `.github/workflows-src/lib/factory-issue-module.gh.js` so it is available to inline scripts via the same module shape used by the other shared helpers.

## 2. Shared library: comment serialization for prompt context

- [x] 2.1 Add a `serializeIssueComments` helper that takes the captured comment array and returns a deterministic markdown rendering suitable for embedding in an agent prompt (each comment prefixed with author + UTC timestamp, newlines escaped consistently, total length truncated and marked if it would exceed an agent-context budget).
- [x] 2.2 Add unit tests covering ordering, single-comment rendering, multi-comment rendering, length truncation with marker, and stable formatting for empty input.

## 3. Workflow source skeleton

- [x] 3.1 Create directory `.github/workflows-src/research-factory-issue/` with subdirectories `scripts/` and a `workflow.md.tmpl` placeholder.
- [x] 3.2 Create `intake-constants.js` in the new directory mirroring change-factory's structure (e.g. `FACTORY_LABEL = 'research-factory'`, `ISSUE_BRANCH_PREFIX` is unused but include a placeholder comment explaining its omission since this workflow does not branch).
- [x] 3.3 Add the new workflow to `.github/workflows-src/manifest.json` with template path and output path `.github/workflows/research-factory-issue.md`.

## 4. Workflow inline scripts

- [x] 4.1 Add `scripts/qualify_trigger.inline.js` mirroring change-factory's qualify trigger (using `FACTORY_LABEL = 'research-factory'`).
- [x] 4.2 Add `scripts/validate_dispatch_inputs.inline.js` mirroring code-factory's dispatch validator (parsing `issue_number` and optional `source_workflow`).
- [x] 4.3 Add `scripts/fetch_live_issue.inline.js` mirroring code-factory's fetch-live-issue script.
- [x] 4.4 Add `scripts/check_actor_trust.inline.js` mirroring code-factory's actor trust script (issue-event mode only).
- [x] 4.5 Add `scripts/fetch_issue_comments.inline.js` that calls the new `factoryFetchIssueComments` shared helper and emits the captured comments as a serialized GitHub Actions output (using the heredoc/EOF pattern already used by code-factory's `normalize_context` step for multi-line outputs).
- [x] 4.6 Add `scripts/remove_trigger_label.inline.js` reusing the shared `removeTriggerLabel` helper with `labelName: 'research-factory'`.
- [x] 4.7 Add `scripts/finalize_gate.inline.js` adapted from change-factory's finalize gate (no duplicate-PR concept; gate covers event eligibility, actor trust, and concurrency).
- [x] 4.8 Add unit tests for any new inline scripts that contain non-trivial logic (e.g. `fetch_issue_comments.inline.js`) under `.github/workflows-src/lib/`.

## 5. Workflow template (`workflow.md.tmpl`)

- [x] 5.1 Author the workflow frontmatter `on:` section with `issues.opened`, `issues.labeled`, and `workflow_dispatch` (with `issue_number: number` and optional `source_workflow: string`).
- [x] 5.2 Set `on.status-comment: true` and the pre-activation `on.permissions:` block (`contents: read`, `issues: write`, `pull-requests: read`).
- [x] 5.3 Author the `on.steps` pipeline: `determine_intake_mode` → `qualify_trigger` (issue-event only) → `capture_issue_context` (issue-event only) → `validate_dispatch_inputs` (dispatch only) → `fetch_live_issue` (dispatch only) → `check_actor_trust` (issue-event only) → `fetch_issue_comments` → `remove_trigger_label` (issue-event only) → `normalize_context` → `finalize_gate`.
- [x] 5.4 Author the `normalize_context` step as a Bash step that fans the issue-event vs dispatch outputs into a unified set of normalized outputs (`intake_mode`, `issue_number`, `issue_title`, `issue_body`, `issue_comments`, `event_eligible`, `event_eligible_reason`, `actor_trusted`, `actor_trusted_reason`, `trigger_label_removed`, `trigger_label_removed_reason`, `source_workflow`), patterned on code-factory's normalize_context step. Use heredoc-EOF for the multi-line `issue_body` and `issue_comments` outputs.
- [x] 5.5 Add the workflow-level `concurrency:` block keyed `research-factory-issue-${{ github.event.issue.number || inputs.issue_number }}` with `cancel-in-progress: false`.
- [x] 5.6 Add the workflow-level top-level `if:` gate that requires `event_eligible == 'true'`, `actor_trusted == 'true'`, and `issue_number != ''`.
- [x] 5.7 Add the agent job's `steps:` for `actions/setup-node@v6` (using `node-version-file: package.json`) and `npm ci`. Do **not** add Go, Terraform, Elastic Stack, Fleet, or API-key setup steps.
- [x] 5.8 Set `timeout-minutes: 35` on the agent job.
- [x] 5.9 Configure the `engine:` block (Claude via litellm proxy) identically to change-factory.
- [x] 5.10 Configure `permissions:` for the agent job: `contents: read`, `issues: read`, `pull-requests: read` (issues:write is **not** granted to the agent — `update-issue` lives in safe-outputs).
- [x] 5.11 Configure `tools.github.toolsets: [issues, repos]` (no `pull_requests` toolset since this workflow doesn't author PRs).
- [x] 5.12 Configure `network.allowed: [defaults, node, elastic.litellm-prod.ai, www.elastic.co]`.
- [x] 5.13 Configure `mcp-servers.elastic-docs` pointing to `https://www.elastic.co/docs/_mcp/`.
- [x] 5.14 Configure `checkout: { fetch-depth: 0 }`.
- [x] 5.15 Configure `safe-outputs:` with `update-issue: { body: , target: triggering, max: 1 }` and `noop: { max: 1, report-as-issue: false }`. Do **not** include `create-pull-request`, `add-comment`, `add-labels`, or any code-writing safe outputs.
- [x] 5.16 Author the agent prompt body as documented in tasks 6.1–6.10.

## 6. Agent prompt content

- [x] 6.1 Open the prompt with role framing: "You author the implementation-research block for a GitHub issue labeled `research-factory`. Your only durable output is a single update to the issue body."
- [x] 6.2 Render the pre-activation context section (gate reason, intake mode, issue number, title, body, normalized comment history, repository, triggered-by, run link).
- [x] 6.3 Document the time budget: "You have ~25 minutes of agentic work. Reserve the last ~3 minutes for emitting your `update_issue`. The job hard-kills at 35 minutes."
- [x] 6.4 Document the partial-output preference: "If you run short on time, prefer emitting a partial-but-valid block with explicit unanswered open questions over emitting `noop`."
- [x] 6.5 Document the elastic-docs MCP availability and the expectation to use `search_docs` / `find_related_docs` / `get_document_by_url` when researching unfamiliar API surface.
- [x] 6.6 Document the comparison requirement: "You SHALL compare at least two distinct candidate approaches under `### Approaches considered`."
- [x] 6.7 Document the block schema in detail (markers, mandatory subsections in order, provenance header, social contract notice).
- [x] 6.8 Document the body-rewrite contract: "Emit exactly one `update_issue` operation with `operation: replace`. The new body SHALL preserve all content outside `<!-- implementation-research:* -->` markers byte-for-byte from the pre-block original issue content. Strip any prior block before composing the new one. The new body SHALL contain exactly one block."
- [x] 6.9 Document free-will semantics: "Edits a user has made inside the prior block are read as input but are not preserved verbatim. Synthesize the next block from: original issue content + chronological comment history + prior block contents (as draft input)."
- [x] 6.10 Document hard guardrails: SHALL NOT modify repository files, SHALL NOT open pull requests, SHALL NOT post free-form comments, SHALL NOT add labels (including `change-factory`), SHALL NOT call `update_issue` more than once, SHALL NOT re-check intake gates.

## 7. change-factory awareness of the research block

- [x] 7.1 Edit `.github/workflows-src/change-factory-issue/workflow.md.tmpl` agent prompt to add a new section explaining the `<!-- implementation-research:start --> ... <!-- implementation-research:end -->` markers, the location of `### Recommendation` / `### Open questions`, the rule that a present block is the exclusive scope source, and the rule that an absent block falls back to today's title-and-body behavior.
- [x] 7.2 Edit the change-factory prompt to add explicit instructions that, when a block is present: adopt `### Recommendation` as the proposal spine, copy `### Open questions` into `design.md` as `## Open questions`, and treat `### Approaches considered` as already-evaluated context.
- [x] 7.3 Add explicit instruction to change-factory: SHALL NOT modify the implementation-research block (no `update-issue` against it; no rewriting of the markers).

## 8. Generation and lock files

- [x] 8.1 Run `make workflow-generate` (or the repository-equivalent command) to produce `.github/workflows/research-factory-issue.md` and `.github/workflows/research-factory-issue.lock.yml`.
- [x] 8.2 Re-run generation for change-factory to refresh `.github/workflows/change-factory-issue.md` and `.github/workflows/change-factory-issue.lock.yml` with the prompt changes from task 7.
- [x] 8.3 Verify generated files compile cleanly with `gh aw compile` (or whatever the repository tooling wraps) and contain the configuration sections asserted by the new spec scenarios.

## 9. Build and test

- [x] 9.1 Run `make build` and confirm it succeeds with the new and modified workflow source files in place.
- [x] 9.2 Run `make check-openspec` (or `./node_modules/.bin/openspec validate research-factory-workflow --type change`) and resolve any reported problems.
- [x] 9.3 Run the workflow-source unit tests (e.g. `npm test` against the `.github/workflows-src/lib/` test files) and confirm all green.
- [x] 9.4 Confirm the linter (e.g. `make check-lint`) is clean.

## 10. Documentation and labels

- [x] 10.1 Coordinate the creation of the `research-factory` GitHub label in the repository (color and description aligned with `change-factory` and `code-factory`). Document the label in the workflow's authored description.
- [x] 10.2 Add a brief operator-facing note in `dev-docs/` (or wherever the existing change-factory / code-factory operator notes live) describing how to trigger research-factory, what the gated section looks like, and the social contract for editing the block / posting comments.
