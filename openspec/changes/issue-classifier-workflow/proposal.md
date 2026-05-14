## Why

Untriaged issues accumulate in the backlog without clear next steps, making it hard for contributors and automation (research-factory, reproducer-factory) to know which issues are ready to act on. Automated triage at filing time and daily backlog sweeps ensures every issue has a routing label within hours of creation.

## What Changes

- **New workflow** `issue-classifier` (`workflow.md.tmpl` + compiled `.lock.yml`) that triggers on `issues: opened`, `schedule: daily`, and `workflow_dispatch`
- Pre-activation step queries untriaged issues, gates the agent when nothing to do
- Agent classifies each issue into one of four categories and emits `add-labels` + `add-comment` safe-outputs
- Labels applied: `triaged` (always) + one of `needs-research`, `needs-reproduction`, `needs-spec`, `needs-human`
- Classification comment explains the routing decision and invites correction

## Capabilities

### New Capabilities

- `ci-issue-classifier-workflow`: The gh-aw workflow definition, pre-activation scripts, and agent prompt covering the full classify-label-comment lifecycle for a single issue or a batch of up to 5 backlog issues

### Modified Capabilities

<!-- none -->

## Impact

- New files under `.github/workflows-src/issue-classifier/` and compiled `.github/workflows/issue-classifier.lock.yml`
- No changes to existing workflows, provider code, or Go source
- Requires labels `triaged`, `needs-research`, `needs-reproduction`, `needs-spec`, `needs-human` to exist in the repository
