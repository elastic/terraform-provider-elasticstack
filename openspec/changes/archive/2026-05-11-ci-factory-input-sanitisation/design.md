## Context

The three LLM-powered factory workflows (change-factory, research-factory, code-factory) ingest user-authored content from GitHub issues and pass it to a Claude agent. Currently, the `sanitize-context.js` shared library provides a single `stripHtmlComments()` function that uses the regex `/<!--[\s\S]*?(?:-->|$)/g` to remove HTML comments.

This function is applied to:
- `issue_body` — ✅ in all three factories
- `issue_comments` (serialized) — ✅ in all three factories
- `prior_research_comment` — ❌ **not sanitised** in research-factory or change-factory

The missing coverage of `prior_research_comment` is the primary gap, but even where `stripHtmlComments` is applied, invisible Unicode characters and ASCII control characters pass through untouched. These are known prompt-injection vectors: characters like zero-width spaces (`\u200B`), bidirectional marks (`\u200E`, `\u200F`), and null bytes (`\x00`) are invisible to human reviewers but parsable by LLMs.

The gh-aw sandbox provides credential protection (GITHUB_TOKEN is not available to the agent), so this is defence-in-depth. The practical risk is: an attacker with read-only access files an issue with hidden prompt injection, and a write-access collaborator applies the factory label triggering the workflow.

## Goals / Non-Goals

**Goals:**
- Apply consistent sanitisation to all three user-sourced input paths in all factory workflows
- Strip invisible ASCII control characters that have no legitimate use in issue content
- Strip invisible Unicode characters (zero-width, bidirectional marks, BOM) that have no legitimate use in issue content
- Preserve the `stripHtmlComments` behaviour unchanged — it's already proven and tested
- Keep the change scope small: one shared function, one test file update, a handful of inline script changes

**Non-Goals:**
- Natural language prompt-pattern detection (e.g. flagging "ignore previous instructions")
- Markdown structure validation or fenced-code-block escaping
- Content-delimiter wrapping (proven ineffective against LLM prompt injection)
- XML/HTML tag stripping beyond HTML comments (too aggressive, breaks legitimate content)

## Decisions

### Decision 1: Compose sanitisation as a pipeline in `sanitizeUserContent`

Rather than adding separate functions and calling them individually at each call site, compose them into a single `sanitizeUserContent(text)` function that runs all three filters in order:

```
input → stripHtmlComments → stripControlChars → stripInvisibleUnicode → output
```

**Rationale**: The call sites (three inline scripts) currently call `stripHtmlComments()` directly. Adding more individual calls at each site increases maintenance surface. A single function keeps the call sites simple and ensures consistent ordering.

**Alternative considered**: Separate exports called individually at each call site. Rejected because any future addition would require updating multiple call sites.

### Decision 2: Keep `stripHtmlComments` exported for backward compatibility

Export both `sanitizeUserContent` (the composed function) and keep `stripHtmlComments` exported. The existing tests for `stripHtmlComments` continue to pass unchanged.

### Decision 3: Control character ranges

Strip the following ASCII control characters:
- `\x00-\x08` (null, start-of-heading, etc.)
- `\x0B` (vertical tab)
- `\x0C` (form feed)
- `\x0E-\x1F` (shift-out through unit-separator)
- `\x7F` (delete)

Preserve:
- `\x09` (tab)
- `\x0A` (line feed)
- `\x0D` (carriage return)

Also strip Unicode line/paragraph separators `\u2028` and `\u2029` — these are not ASCII but function as invisible control characters in most contexts.

### Decision 4: Invisible Unicode ranges

Strip the following invisible Unicode characters:
- `\u200B` (zero-width space)
- `\u200C` (zero-width non-joiner)
- `\u200D` (zero-width joiner)
- `\u200E` (left-to-right mark)
- `\u200F` (right-to-left mark)
- `\u2060` (word joiner)
- `\u2061` (function application)
- `\u2062` (invisible times)
- `\u2063` (invisible separator)
- `\u2064` (invisible plus)
- `\uFEFF` (BOM / zero-width no-break space)

These characters are never legitimate in Terraform provider issue content (no natural language use case for invisible Unicode formatting controls) and are a known vector for hiding text from code review.

### Decision 5: Sanitise prior_research_comment at the write step

The research comment body is written to disk in two places:
- `.github/workflows-src/research-factory-issue/scripts/write_context_files.inline.js` — writes `prior_research_comment` to `/tmp/research-factory-context/prior_research_comment.md`
- `.github/workflows-src/change-factory-issue/scripts/extract_research_comment.inline.js` — writes `research_comment` to `/tmp/change-factory-context/research_comment.md`

Both write steps will be updated to pass the relevant content through `sanitizeUserContent` before writing.

The GHA marker `<!-- gha-research-factory -->` will be stripped like any other HTML comment. This is fine — the agent receives it as a clearly-named file and the workflow template already tells the agent what it contains.

## Risks / Trade-offs

- **[Low] Over-stripping** — A user who legitimately includes zero-width characters or control characters in their issue would have them removed. In practice, these characters have no legitimate use in Terraform provider issues. Risk: negligible.
- **[Low] False positives from stripControlChars** — Tab, newline, and carriage return are preserved. Other control characters (`\x00-\x08`, `\x0B`, `\x0C`, `\x0E-\x1F`, `\x7F`, `\u2028`, `\u2029`) are extremely unlikely in legitimate issue content. Risk: negligible.
- **[Low] IDEM power** — `sanitizeUserContent` is idempotent (applying it twice produces the same result as applying it once), so existing call sites that already strip HTML comments can safely be updated to call the composed function without double-processing issues.