## Why

The `research-factory` workflow currently writes implementation-research output into a gated block inside the issue body, delimited by `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->` markers. This creates a prompt-injection surface: malicious users can embed fake markers in the issue body or human comments, and regex-based extraction by downstream consumers (e.g., `change-factory`) is fragile.

Moving the research output to a dedicated sticky comment authored by `github-actions[bot]` and sanitising HTML comments from all agent input produces a clean security boundary: the only HTML comments the agent ever sees are those it wrote itself.

## What Changes

- **Move `research-factory` output from issue body block to a sticky comment**: the agent emits research via a custom `safe-outputs` script that creates or updates a comment authored by `github-actions[bot]`, identified by a `<!-- gha-research-factory -->` marker.
- **Add a shared HTML-comment sanitisation library** (`.github/workflows-src/lib/`) used by `research-factory`, `change-factory`, and `code-factory` to strip HTML comments from human-authored input before it reaches the agent.
- **Replace `ci-implementation-research-block-format` with `ci-research-factory-comment-format`**: the old body-block capability is deprecated/removed and a new comment-format capability is introduced with the bot-authored comment structure and optional JSON metadata.
- **Update `research-factory` workflow** to use the custom safe-output script, strip HTML comments from issue body and human comments, and no longer rewrite the issue body.
- **Update `change-factory` workflow** to extract prior research from a bot-authored comment (with `findResearchComment` helper) instead of regexing markers from the issue body.
- **Update `code-factory` workflow** to use the shared sanitisation library for its own agent context.
- **Add structured machine-readable JSON metadata** to the research comment, nested inside a collapsible `<details>` element after the References section. The JSON captures typed fields (recommendation spine, confidence, open questions with IDs, affected capabilities, estimated scope, references) that downstream consumers can extract without regex-parsing headings. The human-readable H2/H3/H4 narrative remains unchanged — the JSON is purely accretive.
- **Update `ci-research-factory-issue-intake`** spec to reflect the new output mechanism and input sanitisation requirement.
- **Update `ci-change-factory-issue-intake`** spec to reflect the new research-comment extraction behaviour.

## Capabilities

### New Capabilities
- `ci-html-comment-sanitisation`: Shared deterministic helpers for stripping HTML comments from GitHub issue bodies and human-authored comments before passing them to agents.

### Modified Capabilities
- `ci-research-factory-issue-intake`: Output mechanism changes from `update-issue` to custom `update-research-comment` safe-output script; input sanitisation requirement added.
- `ci-change-factory-issue-intake`: Research extraction changes from regex-based body parsing to comment-based lookup; shared sanitisation applied to issue body and human comments.
- `ci-implementation-research-block-format` **→** `ci-research-factory-comment-format`: Redefined as a bot-authored comment format rather than a body-block format; markers removed; provenance and required subsections retained. **ADDED**: structured JSON metadata block inside a collapsible `<details>` element for machine consumption.

## Impact

- **`.github/workflows-src/research-factory-issue/`**: Workflow template and scripts updated for sticky-comment output and input sanitisation.
- **`.github/workflows-src/change-factory-issue/`**: Updated to extract research from bot comments instead of body markers.
- **`.github/workflows-src/code-factory-issue/`**: Updated to apply HTML-comment sanitisation to agent context.
- **`.github/workflows-src/lib/`**: New shared sanitisation helpers and amended extraction helpers.
- **`openspec/specs/`**: Three delta specs produced (sanitisation, research-factory intake, change-factory intake) and one renamed/refactored spec (research-factory comment format).
- **No provider code changes** — this is a CI workflow and spec-only change.
