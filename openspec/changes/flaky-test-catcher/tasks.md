## 1. Pre-activation JS Script

- [x] 1.1 Create `.github/workflows-src/flaky-test-catcher/scripts/check_ci_failures.inline.js` — queries workflow runs on `main` in the last 3 days, filters `conclusion == 'failure'`, outputs `has_ci_failures`, `failed_run_ids` (JSON), `total_run_count`, `open_issues`, `issue_slots_available`, `gate_reason`
- [x] 1.2 Add shared JS logic to `.github/workflows-src/lib/` if any reusable issue-slot or CI-run query logic is extracted (otherwise inline in the script)
- [x] 1.3 Write unit tests for the pre-activation script logic in `.github/workflows-src/lib/flaky-test-catcher.test.mjs`

## 2. Workflow Template

- [x] 2.1 Create `.github/workflows-src/flaky-test-catcher/workflow.md.tmpl` with YAML frontmatter: trigger (`workflow_dispatch` + daily schedule), `engine`, `permissions` (`contents: read`, `issues: write`, `actions: read`), `safe-outputs` (`create-issue` with labels `flaky-test` + `code-factory`, cap 3; `noop`), `network` (`defaults`), pre-activation job wiring, agent `if` gate
- [x] 2.2 Write the agent prompt (markdown body after `---`) covering: skill reference, pre-activation context usage, log fetching via `gh api`, `--- FAIL:` extraction, fail-rate classification (broken = 100%, flaky ≥ 20%), base-test-name grouping via `TestAcc[^_]+`, commit analysis (messages + changed file paths), dedup against existing `flaky-test` issues, issue creation rules and body format, `noop` conditions

## 3. Agent Skill Document

- [x] 3.1 Create `.agents/skills/flaky-test-catcher/SKILL.md` defining the analysis protocol: how to query `.github/workflows/test.yml` runs via `gh api`, how to fetch job logs (`gh api .../jobs/{id}/logs`), the `--- FAIL:` extraction pattern, the fail-rate formula and thresholds (broken = 100%, flaky ≥ 20%), the base-test-name grouping rule (`TestAcc[^_]+`), the commit analysis steps, and the required issue body sections

## 4. Compiled Output and Manifest

- [x] 4.1 Run `make compile-workflows` (or `go run ./scripts/compile-workflow-sources`) to generate `.github/workflows/flaky-test-catcher.md` from the template
- [x] 4.2 Add the new entry to `.github/workflows-src/manifest.json`: `{ "template": ".github/workflows-src/flaky-test-catcher/workflow.md.tmpl", "output": ".github/workflows/flaky-test-catcher.md" }`
- [x] 4.3 Verify the compiled `.github/workflows/flaky-test-catcher.md` matches the template (no inline JS discrepancies, correct `x-script-include` references resolved)

## 5. Validation

- [x] 5.1 Run `make workflow-test` and confirm pre-activation JS unit tests pass
- [x] 5.2 Run `make check-lint` to confirm the new workflow and manifests pass all lint checks
- [ ] 5.3 Trigger the workflow manually via `workflow_dispatch` on a branch and confirm pre-activation outputs are correct in the Actions log
  > Note: `workflow_dispatch` requires the workflow to exist on the default branch. This step is verified post-merge.
