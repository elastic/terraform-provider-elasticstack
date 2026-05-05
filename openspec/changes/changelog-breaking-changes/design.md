## Context

The PR changelog system has three layers:

1. **`pr-changelog-parser.js`** — parses and validates the `## Changelog` section from a PR body. `extractBreakingChanges` reads the `### Breaking changes` block, stopping at the next `##`/`###` heading or the end of the changelog section. `validateChangelogSectionFull` enforces structural rules including that `### Breaking changes` must be present when `Customer impact: breaking` and must not be empty.

2. **`pr-changelog-check.js`** — builds the comment body that the verifier workflow posts to the PR when the check fails. Contains the "Expected format" documentation block.

3. **`.github/pull_request_template.md`** — provides the default PR body that contributors fill in. Currently pre-fills `Customer impact: none / Summary:` which passes the checker without any author action.

The concrete problem: contributors adding content after their intended breaking change block (OpenSpec notes, Macroscope summaries, separators) have all of that content swept into the released changelog because the parser has no early-stop mechanism within the `### Breaking changes` block.

## Goals / Non-Goals

**Goals:**
- Let authors explicitly terminate the `### Breaking changes` block with an HTML comment end marker
- Reject `### Breaking changes` subsections on PRs where `Customer impact` is not `breaking`
- Make the default PR template state fail the changelog check, forcing conscious authorship
- Document the end marker in the PR template and in the verifier failure comment

**Non-Goals:**
- Changing the changelog renderer (`changelog-renderer.js`) — validation at PR time is sufficient; any pre-existing misuse can be cleaned up manually at release time
- Supporting alternative marker syntaxes or case-insensitive matching — HTML comment semantics are case-sensitive; whitespace flexibility is sufficient
- Retroactively fixing already-merged PRs in release history

## Decisions

### Decision 1: End marker as HTML comment `<!-- /breaking-changes -->`

**Choice:** An HTML comment using a closing-tag convention (`/breaking-changes`), recognised via the anchored regex `/^\s*<!--\s*\/breaking-changes\s*-->\s*$/` (case-sensitive). This allows optional leading/trailing whitespace on the line (indented markers) and optional internal whitespace around the tag name, while requiring the whole line to be nothing but the marker.

**Rationale:** HTML comments are invisible in rendered GitHub Markdown, so the marker doesn't pollute the PR body's display. The `/<name>` convention mirrors common HTML/XML idioms and is visually clear about being a closing delimiter. Full-line anchoring (`^...$`) prevents accidental matches on lines that merely contain the marker alongside other content. Internal whitespace flexibility (`\s*` around the tag name) is a practical allowance for contributors who add spaces by habit; case-sensitivity is preserved to match HTML comment semantics.

**Alternatives considered:**
- `<!-- end-breaking-changes -->` — slightly wordier, same properties. Rejected for brevity.
- A `###` heading as a delimiter — visible in rendered markdown, increases required boilerplate. Rejected.
- Nothing (rely on authors to keep breaking change blocks short) — doesn't address the root cause. Rejected.

### Decision 2: Marker only fires inside `### Breaking changes`, and only outside a code fence

**Choice:** The marker check is placed in `extractBreakingChanges`, guarded by `inBreaking === true` and `fenceType === null`. A marker before `### Breaking changes`, or inside a fenced code block, is silently ignored.

**Rationale:** Ignoring the marker outside the breaking changes block makes it safe to mention `<!-- /breaking-changes -->` in documentation (e.g., the PR template instructions) without accidentally triggering extraction boundaries. The fence guard preserves the existing invariant that content inside fenced blocks is never treated as structural.

### Decision 3: New validation rule in `validateChangelogSectionFull` only, not in the renderer

**Choice:** Add the rule "if `breakingChangesHeadingPresent` and `customerImpact` is a valid value other than `breaking`, emit an error" to `validateChangelogSectionFull`. The renderer (`changelog-renderer.js`) is left unchanged.

**Rationale:** The verifier runs on every PR push and is the enforcement gate before merge. The renderer runs only at release time against already-merged PRs; adding the rule there adds complexity without meaningful protection (any violation would already have been caught at PR time). Any pre-existing merged PRs that violate the new rule can be cleaned up manually.

**Guard against double-errors:** The rule is guarded with `VALID_CUSTOMER_IMPACTS.has(parsed.customerImpact)`. If `customerImpact` is already invalid (e.g., `patch`), the base validator emits an "invalid impact" error and rule C is suppressed — avoiding a confusing second error for the same root cause.

### Decision 4: Error message includes end-marker hint

**Choice:** Error message: `"### Breaking changes section requires Customer impact: breaking; use <!-- /breaking-changes --> as an end marker."`

**Rationale:** The error is directly actionable — it tells the author both what's wrong and the escape hatch if they have extra context they want to preserve in their PR body without it leaking into the changelog.

### Decision 5: PR template default is invalid placeholder text

**Choice:** Replace `Customer impact: none\nSummary:` with `Customer impact: <none, fix, enhancement, breaking>\nSummary: <single line summary>`.

**Rationale:** Angle-bracket placeholders are a widely understood "fill this in" convention. Crucially, `<none, fix, enhancement, breaking>` is not a valid `Customer impact` value, so the changelog check will fail if a contributor submits without replacing it. This creates a "pit of success" — the path of least resistance requires engagement with the format.

### Decision 6: PR template gains a `breaking` example with end marker

**Choice:** Add a second "Good example" block showing a complete `breaking` entry with a short `### Breaking changes` prose block and `<!-- /breaking-changes -->`.

**Rationale:** The breaking + end-marker pattern is the most complex case contributors will author. Showing a concrete example is more effective than prose description alone.

## Risks / Trade-offs

**[Risk] Marker regex complexity** — `<!--\s*\/breaking-changes\s*-->` could theoretically match content inside a tilde-fenced block if the fence-tracking has a bug. → Mitigation: the fence guard (`fenceType === null`) is checked first; existing tests cover fence-tracking correctness.

**[Risk] Contributors may not know about the end marker** — PRs authored before the template update won't have it. → Mitigation: the verifier failure comment's "Expected format" block documents it; the new breaking example in the template provides a model. Adoption will grow incrementally.

**[Risk] Rule C fires late (release time) for pre-existing merged PRs** — The renderer doesn't enforce the new rule, so old PRs with `Customer impact: fix` + `### Breaking changes` will render their breaking block into the changelog until manually cleaned. → Accepted: scope is pre-merge enforcement only; cleanup is manual.

## Migration Plan

No deployment steps beyond merging the PR. The workflow tests (`npm test`) cover the parser and check logic. The PR template change takes effect immediately on merge. No rollback mechanism is needed — reverting the PR reverts all changes.

## Open Questions

None — all decisions made during exploration.
