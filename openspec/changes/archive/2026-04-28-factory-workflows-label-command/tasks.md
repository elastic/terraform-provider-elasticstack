## 1. Shared remove-label helper

- [x] 1.1 Generalize `.github/workflows-src/lib/remove-trigger-label.js` to accept **label name** and **issue number** (issue API), preserving 404-treat-as-success behavior
- [x] 1.2 Update `.github/workflows-src/openspec-verify-label/scripts/remove_trigger_label.inline.js` to call the generalized helper with `verify-openspec` and the PR/issue number so behavior stays unchanged
- [x] 1.3 Add unit tests for the generalized helper (cover factory labels and issue path, plus existing verify behavior)

## 2. Factory workflow templates

- [x] 2.1 In `.github/workflows-src/code-factory-issue/workflow.md.tmpl`, **keep** `on.issues.types: [opened, labeled]` (current triggers); add **`status-comment: true`** to `on:`
- [x] 2.2 Add a **Remove factory trigger label** (or equivalent name) step in **`on.steps`** after gating outputs needed for the agent `if:` are available, using `actions/github-script@v9` and `x-script-include` to a new script that calls the generalized helper with **`code-factory`** and `context.payload.issue.number`; mirror openspec-verify-label lines 24–30 structurally
- [x] 2.3 Add **`issues: write`** to `on.permissions` for pre-activation if not already present, consistent with label removal
- [x] 2.4 Wire **`jobs.pre-activation.outputs`** for remove-step outputs if the compiled workflow needs them (optional; follow verify workflow if useful)
- [x] 2.5 Repeat **2.1–2.4** for `.github/workflows-src/change-factory-issue/workflow.md.tmpl` with **`change-factory`**

## 3. Regenerate and validate

- [x] 3.1 Run `make workflow-generate` (or documented equivalent) and commit regenerated `code-factory-issue.*` and `change-factory-issue.*` under `.github/workflows/`
- [x] 3.2 Run workflow lib tests and fix any `if:` ordering / skipped-step output issues for the remove step
- [x] 3.3 Sync delta specs into `openspec/specs/` when ready; run `openspec validate` / `make check-openspec`
