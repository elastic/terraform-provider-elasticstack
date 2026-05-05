## Why

Three of the repository's Agentic Workflows (`duplicate-code-detector`, `semantic-function-refactor`, and `schema-coverage-rotation`) each carry an identical copy-pasted pre-activation mechanism: counting open GitHub issues by label, computing a `cap - open` slot count, exposing outputs, gating the agent job, and injecting a `## Pre-activation context` section into the agent prompt. This redundancy means any change to the slot-gating logic, prompt wording, or script structure must be made in three places. Extracting this into a single parameterized shared import reduces maintenance burden and the risk of drift.

## What Changes

- Create a new shared GH AW workflow component at `.github/workflows-src/shared/issue-slots.md.tmpl` that contains:
  - A parameterized `import-schema` accepting `label` (string, required) and `cap` (number, default: 3)
  - Pre-activation `steps` (`compute_issue_slots` via `actions/github-script@v9` with a single canonical script)
  - Pre-activation `jobs` exposing `open_issues`, `issue_slots_available`, `gate_reason`
  - A `## Pre-activation context` body section interpolated with the label and cap
- Add a `scripts/compute_issue_slots.inline.js` under `.github/workflows-src/shared/scripts/` as the single canonical script that `//include:`s `../../lib/issue-slots.js`
- Update the workflow manifest (`manifest.json`) to include `shared/issue-slots.md.tmpl` with an `output` path under `.github/workflows/shared/`
- Update the three consumer workflows to import `shared/issue-slots.md` via GH AW `imports: - uses:` passing their `label` and `cap`
- Remove the triplicated `compute_issue_slots.inline.js` scripts and in-workflow pre-activation prompt text from the three consumer directories
- Regenerate all workflow artifacts (`make workflow-generate` or equivalent)

No breaking changes. Workflow behavior remains identical: same labels, same caps, same gating outputs.

## Capabilities

### New Capabilities
- (none — this is an internal infrastructure refactoring)

### Modified Capabilities
- (none — spec-level requirements are unchanged; only implementation structure is consolidated)

## Impact

- **`.github/workflows-src/shared/`** — new directory with the shared component and script
- **`.github/workflows-src/duplicate-code-detector/`** — simplified template dropping duplicated frontmatter and prompt text
- **`.github/workflows-src/semantic-function-refactor/`** — same simplification
- **`.github/workflows-src/schema-coverage-rotation/`** — same simplification
- **`.github/workflows-src/manifest.json`** — new entry for the shared component
- **`.github/workflows/shared/issue-slots.md`** — generated shared workflow artifact (created by the custom compiler, then compiled by `gh aw compile`)
- **`.github/workflows/`** — regenerated consumer workflow artifacts
- **Kibana-spec-impact workflow** — intentionally unaffected; its pre-activation uses a separate Go-tool mechanism and is out of scope
