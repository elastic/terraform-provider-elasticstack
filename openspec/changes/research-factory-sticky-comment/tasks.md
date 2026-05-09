## 1. Shared sanitisation library

- [ ] 1.1 Create `.github/workflows-src/lib/sanitize-context.js` with `stripHtmlComments(text)` function
- [ ] 1.2 Create `.github/workflows-src/lib/sanitize-context.test.mjs` with unit tests for `stripHtmlComments`
- [ ] 1.3 Export `findResearchComment(comments, marker)` helper from `sanitize-context.js`
- [ ] 1.4 Add tests for `findResearchComment` (filters by author + marker)
- [ ] 1.5 Run `make workflow-test` (or `node --test .github/workflows-src/lib/*.test.mjs`) to verify new tests pass alongside existing tests

## 2. Research-factory workflow

- [ ] 2.1 Update `.github/workflows-src/research-factory-issue/workflow.md.tmpl` to define `safe-outputs.scripts.update-research-comment`
- [ ] 2.2 Remove `update-issue` from `safe-outputs` in the research-factory workflow template
- [ ] 2.3 Add deterministic pre-activation step to strip HTML comments from issue body before writing `issue_body.md`
- [ ] 2.4 Add deterministic pre-activation step to strip HTML comments from each human comment before writing `issue_comments.md`
- [ ] 2.5 Update agent prompt to instruct `update_research_comment` instead of `update_issue`
- [ ] 2.6 Update agent prompt to remove "preserve body byte-for-byte" and "marker delimiters" instructions
- [ ] 2.7 Update agent prompt to reference `ci-research-factory-comment-format` instead of `ci-implementation-research-block-format`
- [ ] 2.8 Update agent prompt time-budget wording: "issue-body update" → "research comment"
- [ ] 2.9 Update agent prompt to include JSON metadata schema instructions and `<details>` element formatting
- [ ] 2.10 Update agent prompt to instruct the agent to ensure JSON metadata is internally consistent with human-readable subsections
- [ ] 2.11 Regenerate compiled workflow with `make workflow-generate`
- [ ] 2.12 Verify compiled `.github/workflows/research-factory-issue.md` contains `safe-outputs.scripts.update-research-comment`
- [ ] 2.13 Verify compiled workflow no longer contains `update-issue` in safe-outputs

## 3. Change-factory workflow

- [ ] 3.1 Update `.github/workflows-src/change-factory-issue/workflow.md.tmpl` pre-activation to fetch issue comments
- [ ] 3.2 Add deterministic step to extract research comment using `findResearchComment` from shared library
- [ ] 3.3 Add deterministic step to strip HTML comments from issue body and human comments
- [ ] 3.4 Update agent prompt: replace body-block extraction instructions with comment-based extraction
- [ ] 3.5 Update agent prompt: describe `<!-- gha-research-factory -->` marker + `## Implementation research` heading
- [ ] 3.6 Update agent prompt: remove references to `<!-- implementation-research:start/end -->` markers
- [ ] 3.7 Update agent prompt: add instructions for extracting the JSON metadata block from the `<details>` element when present
- [ ] 3.8 Update agent prompt: note that JSON metadata is a future enhancement area and the agent should not depend on it today
- [ ] 3.9 Regenerate compiled workflow with `make workflow-generate`
- [ ] 3.10 Verify compiled `.github/workflows/change-factory-issue.md` reflects comment-based extraction

## 4. Code-factory workflow

- [ ] 4.1 Update `.github/workflows-src/code-factory-issue/workflow.md.tmpl` to strip HTML comments from issue body and human comments
- [ ] 4.2 Regenerate compiled workflow with `make workflow-generate`
- [ ] 4.3 Verify compiled `.github/workflows/code-factory-issue.md` applies sanitisation

## 5. Testing and validation

- [ ] 5.1 Validate all compiled workflows with `make check-lint` (or equivalent CI lint job)
- [ ] 5.2 Test `stripHtmlComments` edge cases: nested comments, no comments, empty string, markdown code fences
- [ ] 5.3 Test `findResearchComment` with: no matching comment, matching comment, multiple bot comments
- [ ] 5.4 End-to-end test on a staging issue: apply `research-factory` label, verify sticky comment is created
- [ ] 5.5 End-to-end test: re-apply `research-factory` label, verify comment is updated in place (not duplicated)
- [ ] 5.6 End-to-end test: apply `change-factory` label, verify it reads the research comment correctly
- [ ] 5.7 Verify issue body is untouched after both research-factory runs
- [ ] 5.8 Verify research comment contains a collapsed `<details>` element with valid JSON after the References section
- [ ] 5.9 Verify JSON `recommendation.spine` matches the human-readable Recommendation section
- [ ] 5.10 Verify JSON `open_questions` IDs are stable across re-runs (where questions are unchanged)

## 6. Documentation and spec sync

- [ ] 6.1 Review all delta spec files for consistency with implemented behavior
- [ ] 6.2 Sync `ci-html-comment-sanitisation` delta spec into `openspec/specs/ci-html-comment-sanitisation/spec.md`
- [ ] 6.3 Sync `ci-research-factory-comment-format` delta spec into `openspec/specs/ci-research-factory-comment-format/spec.md`
- [ ] 6.4 Sync `ci-research-factory-issue-intake` modifications into `openspec/specs/ci-research-factory-issue-intake/spec.md`
- [ ] 6.5 Sync `ci-change-factory-issue-intake` modifications into `openspec/specs/ci-change-factory-issue-intake/spec.md`
- [ ] 6.6 Sync `ci-implementation-research-block-format` removals (mark as deprecated/removed in canonical spec)
- [ ] 6.7 Run `make check-openspec` and fix any validation errors
- [ ] 6.8 Archive the change with `openspec archive-change "research-factory-sticky-comment"`
