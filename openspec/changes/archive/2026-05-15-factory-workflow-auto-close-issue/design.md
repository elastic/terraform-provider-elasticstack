## Context

Both `change-factory` and `reproducer-factory` already treat their pull requests as linked, non-closing issue references. Their prompts require `Related to #N` and explicitly prohibit `Closes`/`Fixes` language because these PRs either propose work or reproduce a bug rather than resolving the underlying issue.

However, gh-aw's `create-pull-request` safe output defaults `auto-close-issue` to `true` for issue-triggered workflows. That default appends `Fixes #N` when no closing keyword is already present, which overrides the workflows' intended non-closing linkage semantics. The fix is configuration-only: set `auto-close-issue: false` on the relevant safe outputs in the repository-authored workflow sources and regenerate the compiled workflow artifacts.

## Goals / Non-Goals

**Goals:**
- Make `change-factory` PR creation honor the existing non-closing `Related to #N` contract.
- Make `reproducer-factory` PR creation honor the existing non-closing `Related to #N` contract.
- Capture the non-closing behavior explicitly in the affected OpenSpec workflow requirements.
- Keep the implementation minimal and aligned with gh-aw's documented configuration.

**Non-Goals:**
- Changing branch naming, labels, patch transport, or duplicate-PR detection logic.
- Altering provider code, Terraform acceptance tests, or OpenSpec application behavior.
- Introducing custom post-processing to strip closing keywords from PR bodies.

## Decisions

### Set `auto-close-issue: false` directly on each `create-pull-request` safe output
This is the documented gh-aw mechanism for preventing automatic `Fixes #N` injection on issue-triggered PR creation. It aligns runtime behavior with the workflows' current prompts and existing requirements.

Alternative considered: rely only on prompt instructions to avoid closing keywords. Rejected because the undesired `Fixes #N` text is injected by the safe-output handler, not authored by the agent.

### Treat this as a requirements modification, not a new capability
The existing specs already state that these PRs must use `Related to #N` and must not auto-close the source issue. The change tightens implementation requirements so the workflow configuration actually enforces the documented behavior.

Alternative considered: create a new standalone capability for PR auto-close policy. Rejected because the behavior belongs to the existing workflow-intake capabilities and does not introduce a separate user-visible feature.

### Regenerate compiled workflow artifacts from repository-authored sources
The generated `.github/workflows/*.md` and `.lock.yml` files are not hand-edited. The authored sources under `.github/workflows-src/` remain the source of truth, and generated artifacts must be updated to preserve consistency.

Alternative considered: patch only generated artifacts. Rejected because generated files are not canonical and would be overwritten by the next generation run.

## Risks / Trade-offs

- **Missed authored source location** → Update the repository-authored workflow source first, then regenerate compiled artifacts and verify the generated files also contain `auto-close-issue: false`.
- **Spec drift between workflow behavior and OpenSpec requirements** → Update both affected specs so the implementation contract matches the workflow configuration.
- **Assuming prompt text alone is sufficient** → Use the gh-aw configuration flag so the behavior is enforced even if prompt wording changes later.
