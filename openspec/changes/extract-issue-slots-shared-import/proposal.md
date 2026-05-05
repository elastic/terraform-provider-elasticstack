## Why

Three of the repository's Agentic Workflows (`duplicate-code-detector`, `semantic-function-refactor`, and `schema-coverage-rotation`) each carry an identical copy-pasted script file: `scripts/compute_issue_slots.inline.js`. Each copy inlines `lib/issue-slots.js` via the custom compiler's `//include:` directive and is otherwise byte-for-byte identical. Any logic change to this script must be made in three places. Extracting it to a single canonical file in `lib/` removes that maintenance burden and drift risk.

## What Changes

- Create `.github/workflows-src/lib/compute_issue_slots.inline.js` as the single canonical script (includes `issue-slots.js` via `//include:`)
- Update the three consumer workflow templates to point `x-script-include:` at `../lib/compute_issue_slots.inline.js` instead of their local `scripts/compute_issue_slots.inline.js`
- Delete the three per-workflow `scripts/compute_issue_slots.inline.js` files and their now-empty `scripts/` directories

No other changes. The frontmatter `jobs:`, `steps:`, `## Pre-activation context` body, and all workflow behavior remain exactly as-is.

## Capabilities

### New Capabilities
- (none — this is an internal infrastructure refactoring)

### Modified Capabilities
- (none — spec-level requirements are unchanged; only the script source location is consolidated)

## Impact

- **`.github/workflows-src/lib/`** — new `compute_issue_slots.inline.js` file added
- **`.github/workflows-src/duplicate-code-detector/workflow.md.tmpl`** — `x-script-include:` path updated; `scripts/` directory deleted
- **`.github/workflows-src/semantic-function-refactor/workflow.md.tmpl`** — same
- **`.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl`** — same
- **`.github/workflows/`** — regenerated consumer workflow artifacts (functionally identical)
- **`manifest.json`** — unchanged
- **`kibana-spec-impact`** — intentionally unaffected; out of scope
