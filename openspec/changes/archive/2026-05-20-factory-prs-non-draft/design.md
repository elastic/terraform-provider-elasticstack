## Context

The `create-pull-request` safe output from the GitHub Agentic Workflows safeoutputs server defaults to `draft: true`. Both the `change-factory` and `reproducer-factory` issue-intake workflows use this safe output without overriding the draft flag. As a result, every PR they produce lands as a draft, requiring a maintainer to manually convert it to ready-for-review before the PR can receive reviews.

The fix is a one-field addition to the `safe-outputs.create-pull-request` block in each workflow source:

```yaml
safe-outputs:
  create-pull-request:
    draft: false
    ...
```

This is the minimal, lowest-risk change — no logic, no prompt changes, no new steps. The flag is declared in workflow frontmatter, compiled into the lock file, and respected by the safeoutputs layer without any agent involvement.

## Goals / Non-Goals

**Goals:**
- Ensure change-factory proposal PRs are created in ready-for-review state.
- Ensure reproducer-factory reproduction PRs are created in ready-for-review state.
- Update specs to codify this as a requirement.

**Non-Goals:**
- Changing any other PR creation behaviour (labels, auto-close, patch transport, branch naming).
- Applying this change to the `code-factory` or `research-factory` workflows (not mentioned in the issue).
- Adding explicit `draft: false` instructions to the agent prompt (the safe-output declaration is sufficient).

## Decisions

### Set `draft: false` in workflow frontmatter

The `safe-outputs.create-pull-request` block in each workflow source is the canonical place to configure PR creation behaviour. Adding `draft: false` there means the compiled lock file enforces the constraint without depending on the agent to make the right call. No prompt changes are needed.

Alternative considered: instruct the agent to pass `draft: false` when emitting `create-pull-request`.
Rejected because: the safe-output configuration layer is the right place for a workflow-level constraint; agent-side instructions can be ignored or omitted under edge cases.

### Scope limited to change-factory and reproducer-factory

The issue names only these two factory workflows. No other workflows are in scope for this change.

## Open questions

<!-- None identified. The fix is unambiguous: add `draft: false` to the safe-output configuration in the two named workflows. -->
