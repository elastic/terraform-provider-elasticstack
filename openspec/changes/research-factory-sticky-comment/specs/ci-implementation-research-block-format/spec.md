# ci-implementation-research-block-format Specification

## REMOVED Requirements

### Requirement: Block is delimited by stable HTML-comment markers
**Reason**: Replaced by the `ci-research-factory-comment-format` capability. Research output now lives in a dedicated bot-authored comment rather than a gated block inside the issue body, eliminating the marker-parsing surface entirely.
**Migration**: Refer to `openspec/changes/research-factory-sticky-comment/specs/ci-research-factory-comment-format/spec.md` for the new comment-based format. The canonical path `openspec/specs/ci-research-factory-comment-format/spec.md` will be created when this change is synced.

#### Scenario: Body markers are no longer produced
- **WHEN** a maintainer inspects research output authored after this change
- **THEN** the output SHALL NOT contain `<!-- implementation-research:start -->` or `<!-- implementation-research:end -->` markers
- **AND** the output SHALL be found in a comment authored by `github-actions[bot]` instead

### Requirement: Block opens with a provenance header
**Reason**: Provenance header is migrated to the new comment-format spec.
**Migration**: See `openspec/changes/research-factory-sticky-comment/specs/ci-research-factory-comment-format/spec.md`. The canonical path will be created when this change is synced.

#### Scenario: Provenance lives in the research comment
- **WHEN** a maintainer reads research output after this change
- **THEN** the provenance header SHALL appear inside the bot-authored comment, not inside the issue body

### Requirement: Block contains the mandatory research subsections
**Reason**: Subsection requirements are migrated to the new comment-format spec.
**Migration**: See `openspec/changes/research-factory-sticky-comment/specs/ci-research-factory-comment-format/spec.md`. The canonical path will be created when this change is synced.

#### Scenario: Subsections are in the research comment
- **WHEN** a maintainer inspects research output after this change
- **THEN** the mandatory subsections SHALL be inside the bot-authored comment

### Requirement: Block is regenerated each run; outside content is preserved
**Reason**: With research output in a separate comment, there is no "outside content" to preserve. The issue body is left untouched by `research-factory`.
**Migration**: The `research-factory` workflow no longer modifies the issue body at all. See `openspec/changes/research-factory-sticky-comment/specs/ci-research-factory-comment-format/spec.md` for the regeneration contract. The canonical path will be created when this change is synced.

#### Scenario: Issue body is untouched across re-runs
- **WHEN** the workflow re-runs for an issue
- **THEN** the issue body SHALL NOT be modified by the research-factory workflow
- **AND** the research comment SHALL be updated in place instead

### Requirement: Prior block contents and human-authored comments are read as input on the next run
**Reason**: Input reading is migrated to the new comment-format spec.
**Migration**: See `openspec/changes/research-factory-sticky-comment/specs/ci-research-factory-comment-format/spec.md`. The canonical path will be created when this change is synced.

#### Scenario: Prior research comment is read as input
- **WHEN** the workflow re-runs for an issue that already has a research comment
- **THEN** the prior comment contents SHALL be read as input alongside sanitised issue body and sanitised human comments
