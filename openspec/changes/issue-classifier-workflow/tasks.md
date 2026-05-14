## 1. Pre-activation script

- [ ] 1.1 Create `.github/workflows-src/issue-classifier/scripts/compute_issues.inline.js` — determines trigger mode (`event`, `scheduled`, `dispatch`), queries untriaged issues for scheduled/dispatch paths, checks `triaged` label for event path, outputs `mode`, `issues_json` (JSON array of `{number, title}`), `issue_count`, and `gate_reason`

## 2. Workflow source file

- [ ] 2.1 Create `.github/workflows-src/issue-classifier/workflow.md.tmpl` with YAML frontmatter: triggers (`issues: [opened]`, `schedule: daily`, `workflow_dispatch` with optional `issue_number`), engine config (`llm-gateway/claude-sonnet-4-6`, `--effort high`), permissions (`issues: read`), safe-outputs (`add-labels`, `add-comment`, `noop`), network (`defaults`, `elastic.litellm-prod.ai`), and pre-activation step referencing `compute_issues.inline.js`
- [ ] 2.2 Write the agent prompt body (markdown section of `workflow.md.tmpl`): classification rubric for all four categories with clear criteria and examples, per-issue loop instructions, `add_labels` call format (both `triaged` + `needs-*`), `add_comment` template with `<!-- gha-issue-classifier -->` marker and warm explanatory tone, and `noop` condition

## 3. Compile and validate

- [ ] 3.1 Run `gh aw compile issue-classifier` to produce `.github/workflows/issue-classifier.lock.yml` and `.github/workflows/issue-classifier.md`
- [ ] 3.2 Run `npx openspec validate` (or `make check-openspec`) and confirm no spec violations

## 4. Label prerequisites

- [ ] 4.1 Verify that labels `triaged`, `needs-research`, `needs-reproduction`, `needs-spec`, and `needs-human` exist in the repository; create any that are missing via `gh label create`
