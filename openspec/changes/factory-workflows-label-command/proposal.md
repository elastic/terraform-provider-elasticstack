## Why

Factory issue-intake should keep today’s **`issues` trigger model** (`opened` and `labeled`) while still giving maintainers **one-shot label semantics** (re-apply the factory label to re-run) and **run visibility** on the issue. GitHub Agentic Workflows [status comments](https://github.github.com/gh-aw/reference/triggers/#status-comments-status-comment) cover the workflow-run link without a custom “post URL” step; deterministic label removal can mirror the proven pattern used by **OpenSpec verify (label)** (pre-activation `github-script` + shared helper).

## What Changes

- **Keep** the existing **`issues: types: [opened, labeled]`** (or equivalent) trigger configuration for both **Code Factory** and **Change Factory** issue-intake workflows—no switch to `label_command`.
- Add **`status-comment: true`** to each workflow’s top-level `on:` so the framework posts/updates a status comment on the triggering issue with the workflow run link.
- Add a **deterministic pre-activation step** (same structural placement as `.github/workflows-src/openspec-verify-label/workflow.md.tmpl` lines 24–30: `actions/github-script@v9` + `x-script-include`) that **removes the factory trigger label** from the issue when the run is allowed to proceed to the agent, **reusing the shared `remove-trigger-label` helper** as far as practical (generalize parameters for issue number and label name rather than duplicating API logic).
- Update OpenSpec requirements for **`ci-code-factory-issue-intake`** and **`ci-change-factory-issue-intake`** to document `status-comment` and deterministic label removal; extend or add tests for the generalized helper.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `ci-code-factory-issue-intake`: Add requirements for `status-comment: true` and for deterministic removal of the `code-factory` label in pre-activation when the agent gate passes; triggers remain `issues.opened` / `issues.labeled` as today.
- `ci-change-factory-issue-intake`: Same for `change-factory`.

## Impact

- `.github/workflows-src/code-factory-issue/workflow.md.tmpl` and `.github/workflows-src/change-factory-issue/workflow.md.tmpl` (`on:` gains `status-comment: true`; `on.steps` gains remove-label step; `on.permissions` may need **`issues: write`** for label removal on issues, matching the verify workflow’s pattern).
- `.github/workflows-src/lib/remove-trigger-label.js` (or adjacent module) generalized for **label name + issue number**, with **openspec-verify-label**’s inline script updated to call the generalized API so behavior stays single-sourced.
- New workflow-local `scripts/remove_factory_trigger_label.inline.js` (or shared path) for each factory, following the verify workflow’s `x-script-include` pattern.
- Regenerated `.github/workflows/*.md` / `*.lock.yml`; tests under `.github/workflows-src/lib/`.
- Canonical specs via delta sync/archive.
