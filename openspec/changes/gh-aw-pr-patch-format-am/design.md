## Context

The repository uses GitHub Agentic Workflows (GH AW) to create pull requests from `code-factory` and `change-factory` issue-intake runs. A recent failure showed the safe-output PR creation step attempting to apply a bundle whose prerequisite commit was not present in the safe_outputs job checkout. GH AW documentation provides a supported frontmatter lever, `safe-outputs.create-pull-request.patch-format`, with `am` as a documented alternative to `bundle`.

The generated lock workflows currently contain safe_outputs checkout behavior that is separate from the main agent checkout configuration. Because the documented frontmatter guarantee is the patch transport setting, this change standardizes on `patch-format: am` rather than relying on checkout-depth behavior alone.

## Goals / Non-Goals

### Goals
- Use a documented GH AW configuration lever to avoid bundle prerequisite failures in PR-producing factory workflows.
- Apply the same transport policy consistently to both `code-factory` and `change-factory` issue-intake workflows.
- Capture the behavior in OpenSpec so future workflow edits preserve the safer transport.

### Non-Goals
- Change how duplicate detection, trust checks, or trigger qualification work.
- Rework generated safe_outputs checkout logic.
- Introduce a new workflow or change any provider/resource behavior.

## Decisions

### Use `patch-format: am` for PR creation

Both workflows will configure:

```yaml
safe-outputs:
  create-pull-request:
    patch-format: am
```

This uses the documented patch-based transport for PR creation and avoids requiring bundle prerequisite commits to exist in a later job checkout.

### Capture the behavior as a workflow requirement

The existing workflow capabilities already specify frontmatter requirements such as network policy, MCP server configuration, labels, and linkage behavior. This change adds an explicit requirement to each capability that the authored workflow frontmatter configures `safe-outputs.create-pull-request.patch-format: am`, and that generated workflow artifacts reflect that authored setting.

## Alternatives Considered

### Rely only on `checkout.fetch-depth: 0`

Rejected as the primary fix. The frontmatter docs clearly scope `checkout` to repository checkout behavior for the agent job, while the observed failure occurred in the later safe_outputs job. Increasing checkout depth may still help operationally, but it is not the clearest documented frontmatter control for this failure mode.

### Keep bundle transport and modify generated lock workflows directly

Rejected. The repository treats compiled workflow artifacts as generated output; requirement changes should target the authored workflow source, not hand-edited lock files.

### Pin safe_outputs to the original run SHA

Rejected for this change. That would require compiler/runtime behavior changes or direct workflow surgery beyond the documented frontmatter levers explored here.

## Risks / Trade-offs

- `am` may behave differently from `bundle` in edge cases where exact commit transport was beneficial, but it is the documented and appropriate mitigation for missing bundle prerequisites.
- This change depends on the compiler/runtime honoring the `patch-format` frontmatter setting in generated workflow artifacts.

## Verification

- Inspect the authored workflow sources for both factories to confirm `safe-outputs.create-pull-request.patch-format: am` is declared.
- Regenerate the compiled workflow artifacts and confirm the generated outputs reflect the configured patch format.
- Validate OpenSpec artifacts with `openspec validate --all`.
