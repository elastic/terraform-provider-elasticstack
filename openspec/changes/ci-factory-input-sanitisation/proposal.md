## Why

The change-factory, research-factory, and code-factory workflows ingest user-authored content from GitHub issues (issue body, issue comments, and prior research comments) and feed it into LLM agents. Currently, the only sanitisation applied is HTML comment stripping via `stripHtmlComments()`, and this is not applied uniformly to all input sources. This leaves the agent pipelines vulnerable to prompt injection via invisible or parser-confusing characters — hidden instructions embedded in zero-width Unicode or ASCII control characters that are invisible to human reviewers but visible to LLMs.

## What Changes

- **Extend `stripHtmlComments` into a broader `sanitizeUserContent` function** in `lib/sanitize-context.js` that composes:
  - HTML comment stripping (existing, unchanged)
  - Control character stripping (new: removes non-printable ASCII control chars except `\n`, `\r`, `\t`)
  - Invisible Unicode stripping (new: removes zero-width spaces, joiners, directional marks, BOM, and related invisible characters)
- **Fix the `prior_research_comment` gap** in the research-factory and change-factory workflows: apply `sanitizeUserContent` to the prior research comment body before writing it to disk, bringing it in line with the other input sources.
- **Update tests** in `sanitize-context.test.mjs` to cover the new sanitisation functions.

## Capabilities

### New Capabilities
- `ci-factory-comment-sanitisation`: rules for sanitising user-sourced content in factory issue intake workflows

### Modified Capabilities
<!-- No existing spec changes — this is a new defensive layer, not a behaviour change. -->

## Impact

- `.github/workflows-src/lib/sanitize-context.js` — core sanitisation logic extended
- `.github/workflows-src/research-factory-issue/scripts/write_context_files.inline.js` — apply sanitisation to prior research comment
- `.github/workflows-src/change-factory-issue/scripts/sanitize_context.inline.js` — unchanged (already applies stripHtmlComments), but will use the new composed function
- `.github/workflows-src/change-factory-issue/scripts/extract_research_comment.inline.js` — apply sanitisation to the prior research comment before writing it to disk
- `.github/workflows-src/code-factory-issue/scripts/sanitize_context.inline.js` — unchanged, same as above
- `.github/workflows-src/lib/sanitize-context.test.mjs` — new test coverage
- `.github/workflows-src/lib/factory-issue-comments.js` — the `serializeIssueComments` function handles prompt-budget truncation; no change needed but worth noting it already limits context size