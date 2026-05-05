## Context

Three Agentic Workflow templates each contain an identical `scripts/compute_issue_slots.inline.js` file. The script inlines `lib/issue-slots.js` via the custom compiler's `//include:` directive and uses env vars `ISSUE_SLOTS_LABEL` / `ISSUE_SLOTS_CAP` which are already set per-workflow in the frontmatter `steps` block. The only thing that varies across consumers is those env vars — the script body is byte-for-byte identical.

The custom compiler resolves `x-script-include:` paths relative to the template file. Moving the script to `lib/` and updating the path in each template is the minimal change.

## Goals / Non-Goals

**Goals:**
- Single canonical `compute_issue_slots.inline.js` in `lib/`
- Three updated `x-script-include:` references pointing at `../lib/compute_issue_slots.inline.js`
- Three deleted `scripts/` directories

**Non-Goals:**
- Deduplicating the `jobs:`, `steps:`, or `## Pre-activation context` body blocks across consumer templates
- Any GH AW shared import mechanism
- Changing workflow behavior (caps, labels, outputs, gating logic)
- Any changes to `kibana-spec-impact`

## Decisions

### Decision: Place the shared script in `.github/workflows-src/lib/`

**Rationale:** `lib/` already holds `issue-slots.js` and its test. This is the established location for shared workflow-source logic. The `//include:` path from any consumer template becomes `../lib/compute_issue_slots.inline.js` — one level up from the workflow directory, matching the existing `../../lib/issue-slots.js` chain.

**Alternatives considered:**
- A new `shared/scripts/` directory. Rejected: unnecessary new structure; `lib/` already exists and is the right home.

### Decision: Leave frontmatter and body duplication in place

**Rationale:** GH AW's `imports:` merge mechanism doesn't compose `jobs:` correctly for this use case. The `## Pre-activation context` body duplication is tolerable — it's static text, not logic. Scope is limited to the script.

## Risks / Trade-offs

- **Minimal risk.** The `//include:` path is resolved by the custom compiler at build time. The generated `.github/workflows/*.md` files will be identical to today's — the only difference is where the source lives.
- **Verification:** `make workflow-generate` followed by `make check-workflows` confirms the output is correct.
