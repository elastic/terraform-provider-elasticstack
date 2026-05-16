## Why

The `### Breaking changes` parser currently reads until the next `##`-level heading, so any content following the intended breaking-change prose (OpenSpec notes, Macroscope bot summaries, separators) is silently included in the generated changelog. Additionally, nothing prevents a contributor from adding a `### Breaking changes` block to a PR whose `Customer impact` is `fix` or `none`, producing a misleading or unintended changelog entry. The default pre-filled template also allows a PR to accidentally pass as `Customer impact: none` without the author realising they need to make a choice.

## What Changes

- **End marker support** — `extractBreakingChanges` stops at `<!-- /breaking-changes -->` when the marker appears inside the `### Breaking changes` block (outside a code fence), in addition to the existing heading-boundary stop. The marker is whitespace-flexible but case-sensitive (HTML comment semantics).
- **Validation rule: breaking changes require breaking impact** — `validateChangelogSectionFull` rejects a `### Breaking changes` subsection when `Customer impact` is anything other than `breaking`. The error message directs authors to use the end marker if they want extra context in their PR body.
- **PR template: invalid default state** — The pre-filled template body is changed from `Customer impact: none / Summary:` to angle-bracket placeholder text that fails the changelog check, forcing authors to consciously replace it.
- **PR template: breaking example with end marker** — A second "Good example" block is added showing a complete `breaking` entry with a `### Breaking changes` block and the `<!-- /breaking-changes -->` end marker.
- **Verifier comment: document end marker** — The `Expected format` block in the failure comment is updated to show `<!-- /breaking-changes -->` after the breaking-changes prose.

## Capabilities

### New Capabilities
- (none)

### Modified Capabilities
- `ci-pr-changelog-authoring`: new end-marker extraction rule, new validation rule (breaking changes require breaking impact), updated PR template defaults and examples, updated failure comment format block.

## Impact

- `.github/workflows-src/lib/pr-changelog-parser.js` — parser and validator logic
- `.github/workflows-src/lib/pr-changelog-parser.test.mjs` — new test cases
- `.github/workflows-src/lib/pr-changelog-check.js` — failure comment body
- `.github/pull_request_template.md` — default pre-fill and examples
