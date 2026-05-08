## Why

The `research-factory` workflow currently writes implementation-research output into a gated block inside the issue body, delimited by `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->` markers. This creates a prompt-injection surface: malicious users can embed fake markers in the issue body or human comments, and regex-based extraction by downstream workflows (change-factory, code-factory) is fragile.

Moving the research output to a dedicated sticky comment authored by `github-actions[bot]` and sanitising HTML comments from all agent input produces a clean security boundary: the only HTML comments the agent ever sees are those it wrote itself.

## What Changes

- **Move `research-factory` output from issue body block to a sticky comment**: the agent emits research via a custom `safe-outputs` script that creates or updates a comment authored by `github-actions[bot]`, identified by a `<!-- gha-research-factory -->` marker.
- **Add a shared HTML-comment sanitisation library** (`.github/workflows-src/lib/`) used by `research-factory`, `change-factory`, and `code-factory` to strip HTML comments from human-authored input before it reaches the agent.
- **Rename `ci-implementation-research-block-format` to `ci-research-factory-comment-format`** and redefine "block" as a bot-authored comment (not a region inside the issue body).
- **Update `research-factory` workflow** to use the custom safe-output script, strip HTML comments from issue body and human comments, and no longer rewrite the issue body.
- **Update `change-factory` workflow** to extract prior research from a bot-authored comment (with `findResearchComment` helper) instead of regexing markers from the issue body.
- **Update `code-factory` workflow** to use the shared sanitisation library for its own agent context.
- **Update `ci-research-factory-issue-intake`** spec to reflect the new output mechanism and input sanitisation requirement.
- **Update `ci-change-factory-issue-intake`** spec to reflect the new research-comment extraction behaviour.

## Capabilities

### New Capabilities
- `ci-html-comment-sanitisation`: Shared deterministic helpers for stripping HTML comments from GitHub issue bodies and human-authored comments before passing them to agents.

### Modified Capabilities
- `ci-research-factory-issue-intake`: Output mechanism changes from `update-issue` to custom `update-research-comment` safe-output script; input sanitisation requirement added.
- `ci-change-factory-issue-intake`: Research extraction changes from regex-based body parsing to comment-based lookup; shared sanitisation applied to issue body and human comments.
- `ci-implementation-research-block-format` **→** `ci-research-factory-comment-format`: Redefined as a bot-authored comment format rather than a body-block format; markers removed; provenance and required subsections retained.

## Impact

- **`.github/workflows-src/research-factory-issue/`**: Workflow template and scripts updated for sticky-comment output and input sanitisation.
- **`.github/workflows-src/change-factory-issue/`**: Updated to extract research from bot comments instead of body markers.
- **`.github/workflows-src/code-factory-issue/`**: Updated to apply HTML-comment sanitisation to agent context.
- **`.github/workflows-src/lib/`**: New shared sanitisation helpers and amended extraction helpers.
- **`openspec/specs/`**: Three delta specs produced (sanitisation, research-factory intake, change-factory intake) and one renamed/refactored spec (research-factory comment format).
- **No provider code changes** — this is a CI workflow and spec-only change.
