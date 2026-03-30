## 1. Move gating into deterministic workflow steps

- [x] 1.1 Update `.github/workflows/openspec-verify-label.md` to add deterministic `on.steps` that verify the triggering label and inspect PR files under `openspec/changes/`.
- [x] 1.2 Implement deterministic change-selection outputs for the workflow, including the selected change id and a clear gate result / skip reason for downstream jobs.
- [x] 1.3 Gate the agent job on the deterministic pre-activation outputs so ineligible runs skip the expensive agent execution path.
- [x] 1.4 Update the agent instructions so they consume pre-activation outputs instead of rediscovering label state or PR file selection logic.

## 2. Prepare the agent workspace deterministically

- [x] 2.1 Add deterministic pre-agent custom `steps:` to install repository Node dependencies with `npm ci` so `npx openspec` is available in the agent job.
- [x] 2.2 Remove redundant setup instructions from the markdown prompt while keeping verification, relevance review, review submission, and archive-on-approve behavior intact.

## 3. Regenerate and verify workflow artifacts

- [x] 3.1 Recompile `.github/workflows/openspec-verify-label.lock.yml` from the markdown source with `gh aw compile`.
- [x] 3.2 Verify cleanup and other terminal jobs still behave correctly when `needs.agent.result == 'skipped'`.
- [x] 3.3 Run the relevant OpenSpec and workflow validation checks to confirm the deterministic-step contract, selected-change outputs, skipped-agent path, and compiled workflow stay aligned.
