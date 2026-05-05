# ci-workflow-shared-imports Specification

## Purpose
TBD - created by archiving change extract-issue-slots-shared-import. Update Purpose after archive.
## Requirements
### Requirement: Shared pre-activation script is consolidated in `lib/`

When two or more Agentic Workflow templates include an identical pre-activation script that resolves a shared library via the custom compiler's `//include:` directive, the repository SHALL maintain one canonical script under `.github/workflows-src/lib/` and each consumer SHALL reference it via `x-script-include:` using a relative path. The per-workflow copies of that script SHALL be removed, along with any now-empty `scripts/` directories beneath the consumer workflow directories.

#### Scenario: Issue-slot script is consolidated

- **GIVEN** three workflow templates (`duplicate-code-detector`, `semantic-function-refactor`, and `schema-coverage-rotation`) each contain an identical `scripts/compute_issue_slots.inline.js` that starts with `//include: issue-slots.js`
- **WHEN** the shared script is moved to `.github/workflows-src/lib/compute_issue_slots.inline.js`
- **THEN** each consumer template SHALL update its `x-script-include:` header to `../lib/compute_issue_slots.inline.js`
- **AND** the three per-workflow `scripts/compute_issue_slots.inline.js` files and their now-empty `scripts/` directories SHALL be removed
- **AND** the generated workflow artifacts remain functionally identical to the previous duplicated-script design

