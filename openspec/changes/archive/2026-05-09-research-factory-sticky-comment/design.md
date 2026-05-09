# Design: Research-Factory Sticky Comment

## Context

The `research-factory` workflow currently writes implementation-research output into a gated block inside the issue body, delimited by `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->` markers. The `change-factory` workflow then regex-extracts this block from the issue body to use as the authoritative scope source.

This creates two problems:
1. **Prompt-injection surface**: malicious users can embed fake markers in the issue body or human comments. The agent and downstream consumers parse untrusted content to find markers.
2. **Body-rewrite fragility**: the agent must preserve all issue body content outside the markers "byte-for-byte," which is cognitively expensive and risks overwriting concurrent human edits.

This change moves research output to a dedicated sticky comment authored by `github-actions[bot]`, and adds shared HTML-comment sanitisation across all factory workflows.

## Goals / Non-Goals

**Goals:**
- Eliminate the HTML-comment marker parsing attack surface in factory workflow inputs
- Separate bot-generated research output from human-authored issue content
- Preserve the issue body as immutable community-sourced content
- Allow downstream consumers (`change-factory`, `code-factory`) to locate research output unambiguously by author + marker

**Non-Goals:**
- Changing the research subsections or provenance format (only the container changes)
- Changing the trigger, trust, or concurrency logic of any factory workflow
- Enabling `add-comment` or other new safe outputs for `research-factory`
- Addressing prompt injection via non-HTML-comment vectors (e.g., markdown tricks, unicode homoglyphs)
- Replacing the human-readable narrative with JSON (both coexist; the JSON is accretive)

## Decisions

### Use a custom `safe-outputs.scripts` handler for the sticky comment

**Decision**: Instead of `update-issue`, define a `safe-outputs.scripts` entry named `update-research-comment` that reads the agent output from `$GH_AW_AGENT_OUTPUT` and uses the GitHub REST API to create or update a comment.

**Rationale**: The gh-aw framework does not provide a built-in `update-comment` safe output. Custom scripts run synchronously inside the consolidated safe-outputs job handler loop with no extra job allocation overhead. This keeps the agent's durable output inside the framework's safe-output validation boundary while implementing the create-or-update logic.

**Alternatives considered:**
- **Post-activation `actions/github-script` step** (outside safe-outputs): Rejected because it bypasses the framework's output validation and audit trail. The agent would write a file, and a post-step would blindly post it.
- **Built-in `add-comment` with `max: 1`**: Rejected because `add-comment` always creates new comments; it does not update existing ones. We'd produce comment spam on re-runs.
- **Built-in `create-issue` with the research as the issue body**: Rejected — discussions are separate entities, harder to link visually, and create noise in the Discussions tab.

### Strip HTML comments deterministically before agent context

**Decision**: A shared helper `stripHtmlComments(text)` lives in `.github/workflows-src/lib/` and is called by pre-activation workflow steps before writing `issue_body.md` and `issue_comments.md`.

**Rationale**: Deterministic stripping in workflow steps is auditable and testable. It means the agent literally never receives HTML comments from human-authored content, which is a stronger guarantee than instructing the agent to ignore them.

**Alternatives considered:**
- **Strip in the agent prompt**: Rejected — relies on agent compliance, which is unpredictable across runs and models.
- **Strip server-side in the LLM proxy or gateway**: Rejected — not under our control; also the same content is read by deterministic downstream steps (change-factory pre-activation) that also need clean input.
- **Validate markers instead of stripping**: Rejected — validating that markers are "real" is harder than removing all HTML comments; users can still embed other malicious content inside comments.

### Keep `<!-- gha-research-factory -->` as a lightweight filter marker inside the comment

**Decision**: The comment begins with `<!-- gha-research-factory -->` as its first line. Downstream consumers use this marker plus `author: github-actions[bot]` to find the research comment.

**Rationale**: The `gh-aw` framework posts its own status/activation comments by `github-actions[bot]`. Filtering by author alone is insufficient. A hidden marker is the simplest unambiguous discriminator. Using a framework-generated workflow-id marker is tempting, but that couples us to framework internals and the marker format could change.

**Alternatives considered:**
- **No marker, just search for `## Implementation research` heading**: Rejected — heading text could change; we want a stable machine-readable key.
- **Framework `gh-aw-workflow-id` marker**: Rejected — framework markers are internal implementation details and may not reliably distinguish our research comment from the framework's own status comment.
- **Comment metadata or reaction**: Rejected — GitHub comments have no custom metadata fields; reactions are not queryable in a practical way for this use case.

### Embed structured JSON metadata inside a collapsible `<details>` block

**Decision**: After the `### References` section, the research comment includes an HTML `<details>` element containing a fenced JSON code block with a typed machine-readable representation of the research: `recommendation` (spine, confidence, approach index), `open_questions` (with IDs and `blocking` flags), `affected_capabilities`, `estimated_scope`, and `references`.

**Rationale**: The comment becomes a **dual-interface artifact**: humans read the H2/H3/H4 narrative; machines read the JSON. Downstream consumers (`change-factory` today, a classifier workflow tomorrow) can extract fields without regex-parsing headings, enabling typed validation, auto-promotion gates, and cross-run reference by `open_question.id`. The `<details>` element hides the JSON from human readers by default, so the UX is unchanged. At the moment the workflow generates the comment, the JSON is synthesized from the same reasoning context as the narrative, making inconsistency unlikely. However, if a maintainer edits the comment between runs, the narrative and JSON can diverge until the next workflow run regenerates the entire comment. The regeneration contract therefore re-establishes consistency at each run boundary.

**Alternatives considered:**
- **Separate JSON artifact uploaded via `upload-artifact`**: Rejected — fragments the research output across two locations; downstream consumers would need to download and correlate an artifact with a comment.
- **Schema-enforced YAML frontmatter at the top of the comment**: Rejected — raw YAML frontmatter is ugly for humans and interferes with the provenance narrative at the top of the comment.
- **GitHub comment reactions as metadata signaling**: Rejected — reactions are limited to a fixed emoji set, cannot carry structured data, and are not discoverable in a practical way.
- **Store JSON in a hidden HTML comment**: Rejected — we just went to great lengths to strip HTML comments from human input; embedding machine data in them would be inconsistent.

### Apply sanitisation to all three factory workflows

**Decision**: `research-factory`, `change-factory`, and `code-factory` all apply `stripHtmlComments` to human-authored input.

**Rationale**: The prompt-injection surface exists for any workflow that reads issue bodies and comments. Consistent sanitisation across all factories is a defense-in-depth measure. Even though `code-factory` does not currently consume the research block, it still reads the issue body and comments for implementation context.

## Risks / Trade-offs

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Custom safe-output script fails due to GitHub API rate limiting or network issue | Low | High | The script uses `core.setFailed()` on API errors; the workflow run will fail visibly and can be retried. The research comment is not a critical path for issue triage. |
| StripHtmlComments over-strips legitimate content that happens to contain `<!--` | Low | Medium | The regex targets the standard HTML comment pattern `<!-- ... -->`. Markdown code fences are unlikely to contain accidental `<!--` sequences in this domain. If discovered, the regex can be made more conservative. |
| Concurrent research-factory runs for the same issue race on comment update | Low | Medium | Concurrency is already enforced per issue (`concurrency: group: research-factory-issue-${{ ... }}`). The custom script updates the existing comment by ID, which is atomic. |
| Downstream consumers break if they still expect body-block markers | Medium | High | Change-factory spec is updated in this same change to read from comments. We'll validate the compiled workflow before deploying. |
| Research comment hits the 65,536-character limit | Low | Medium | Same limit as issue body. Research blocks are currently well under this. If growth becomes a concern, we can split or truncate. |
| JSON schema drift between narrative and metadata | Low | High | The agent generates both from the same reasoning context in a single pass, making inconsistency unlikely. We can add a deterministic validation step later. |

## Migration Plan

1. **Implement shared library** (`sanitize-context.js` + tests)
2. **Update research-factory workflow template** to add the custom safe-output script, remove `update-issue`, add sanitisation steps
3. **Regenerate compiled workflow** (`make workflow-generate`) and verify diff
4. **Update change-factory workflow template** to extract research from comments instead of body markers
5. **Regenerate compiled workflow** and verify diff
6. **Update code-factory workflow template** to apply sanitisation
7. **Regenerate compiled workflow** and verify diff
8. **Test on a staging issue**: apply `research-factory` label, verify comment creation, re-apply label, verify comment update
9. **Test change-factory integration**: verify it reads the research comment correctly
10. **Update canonical specs**: sync delta specs into `openspec/specs/`

**Rollback**: The old body-block format is no longer produced once deployed. If a critical issue arises, we can manually restore the old workflow source from git history. Existing issues with body-block research would still work since change-factory falls back to title+body when no research comment is found.

## Open Questions

- Should `code-factory` also consume the research comment for scope, or is title+body sufficient for its use case? (Currently out of scope; can be added later.)
- Should the shared sanitisation library also strip other injection vectors (e.g., `@everyone` mentions, certain Unicode homoglyphs)? The current change limits scope to HTML comments.
