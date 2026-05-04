## 1. Author the semantic refactor workflow source

- [x] 1.1 Add `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl` derived from `https://github.com/github/gh-aw/blob/main/.github/workflows/semantic-function-refactor.md` and include an explicit upstream baseline reference.
- [x] 1.2 Register the new workflow source and generated output path in `.github/workflows-src/manifest.json`.
- [x] 1.3 Configure the workflow engine to use Claude through LiteLLM with model `llm-gateway/gpt-5.5`, the Elastic LiteLLM base URL, and `CLAUDE_LITELLM_PROXY_API_KEY`.
- [x] 1.4 Add deterministic pre-activation issue-slot gating for `ISSUE_SLOTS_LABEL=semantic-refactor` and `ISSUE_SLOTS_CAP=3`, reusing the existing issue-slot helper pattern from duplicate-code detector.
- [x] 1.5 Remove upstream's close-existing-`[refactor]` behavior from the local workflow contract and prompt.

## 2. Define prompt and safe-output behavior

- [x] 2.1 Configure `create-issue` safe outputs with title prefix `[semantic-refactor] `, labels `semantic-refactor`, `refactoring`, `code-quality`, and `automated-analysis`, and max `3`.
- [x] 2.2 Add prompt context for `open_issues`, `issue_slots_available`, and `gate_reason`, and instruct the agent not to query issue capacity itself.
- [x] 2.3 Update the semantic refactor prompt to create one issue per distinct actionable opportunity or tightly related refactor cluster, capped by `issue_slots_available`.
- [x] 2.4 Constrain analysis to non-test Go source files and exclude tests, generated files, workflow files, vendored dependencies, and non-Go files from issue findings.
- [x] 2.5 Ensure each issue template requires concrete file or symbol evidence, impact, and actionable refactoring guidance.

## 3. Generate artifacts and add tests

- [x] 3.1 Run the repository workflow generator to produce `.github/workflows/semantic-function-refactor.md` and `.github/workflows/semantic-function-refactor.lock.yml`.
- [x] 3.2 Add workflow-source tests covering the upstream baseline reference, generated artifact pairing, `semantic-refactor` issue-slot gate, safe-output labels and cap, and prompt issue-creation contract.
- [x] 3.3 Add or update tests that assert the generated lock file preserves the LiteLLM model, base URL, and secret-backed API key for agent execution.
- [x] 3.4 Confirm existing issue-slot helper tests cover the `semantic-refactor` bucket behavior or extend them if needed.

## 4. Validate the change

- [x] 4.1 Run `make workflow-generate` and confirm generated workflow artifacts are up to date.
- [x] 4.2 Run `make workflow-test`.
- [x] 4.3 Run `make check-workflows`.
- [x] 4.4 Run OpenSpec validation for `semantic-refactor-workflow`.
