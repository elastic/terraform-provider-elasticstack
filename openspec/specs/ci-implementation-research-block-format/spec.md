# ci-implementation-research-block-format Specification

> **DEPRECATED — REMOVED**
>
> This capability is **deprecated and removed** as of the `research-factory-sticky-comment` change. The implementation-research block format has been replaced by [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md), which stores research output in a dedicated sticky comment authored by `github-actions[bot]` rather than as a gated block inside the issue body.
>
> **Migration**: See [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md) for the new comment-based format, and [`ci-research-factory-issue-intake`](../ci-research-factory-issue-intake/spec.md) for the updated workflow requirements. The old content below is retained for historical reference only.

## Purpose
Define the format of the implementation-research block produced by the `research-factory` workflow as a gated region inside the issue body, delimited by HTML-comment markers. This capability is no longer active.

## Requirements

### Requirement: Capability is deprecated and removed
This capability is deprecated and removed. The repository SHALL NOT produce new implementation-research blocks delimited by `<!-- implementation-research:start -->` and `<!-- implementation-research:end -->` markers. All requirements in this specification are superseded by [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md).

#### Scenario: Body markers are no longer produced
- **WHEN** a maintainer inspects research output authored after this change
- **THEN** the output SHALL NOT contain `<!-- implementation-research:start -->` or `<!-- implementation-research:end -->` markers
- **AND** the output SHALL be found in a comment authored by `github-actions[bot]` instead

#### Scenario: Issue body is untouched across re-runs
- **WHEN** the workflow re-runs for an issue
- **THEN** the issue body SHALL NOT be modified by the research-factory workflow
- **AND** the research comment SHALL be updated in place instead

## Removed requirements (historical)

The following requirements are preserved for historical reference only. They are no longer in effect.

#### Block is delimited by stable HTML-comment markers
**Status**: REMOVED  
**Reason**: Replaced by the `ci-research-factory-comment-format` capability. Research output now lives in a dedicated bot-authored comment rather than a gated block inside the issue body, eliminating the marker-parsing surface entirely.  
**Migration**: Refer to [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md) for the new comment-based format.

- **WHEN** a maintainer inspects research output authored after this change
- **THEN** the output SHALL NOT contain `<!-- implementation-research:start -->` or `<!-- implementation-research:end -->` markers
- **AND** the output SHALL be found in a comment authored by `github-actions[bot]` instead

#### Block opens with a provenance header
**Status**: REMOVED  
**Reason**: Provenance header is migrated to the new comment-format spec.  
**Migration**: See [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md).

- **WHEN** a maintainer reads research output after this change
- **THEN** the provenance header SHALL appear inside the bot-authored comment, not inside the issue body

#### Block contains the mandatory research subsections
**Status**: REMOVED  
**Reason**: Subsection requirements are migrated to the new comment-format spec.  
**Migration**: See [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md).

- **WHEN** a maintainer inspects research output after this change
- **THEN** the mandatory subsections SHALL be inside the bot-authored comment

#### Block is regenerated each run; outside content is preserved
**Status**: REMOVED  
**Reason**: With research output in a separate comment, there is no "outside content" to preserve. The issue body is left untouched by `research-factory`.  
**Migration**: The `research-factory` workflow no longer modifies the issue body at all. See [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md) for the regeneration contract.

- **WHEN** the workflow re-runs for an issue
- **THEN** the issue body SHALL NOT be modified by the research-factory workflow
- **AND** the research comment SHALL be updated in place instead

#### Prior block contents and human-authored comments are read as input on the next run
**Status**: REMOVED  
**Reason**: Input reading is migrated to the new comment-format spec.  
**Migration**: See [`ci-research-factory-comment-format`](../ci-research-factory-comment-format/spec.md).

- **WHEN** the workflow re-runs for an issue that already has a research comment
- **THEN** the prior comment contents SHALL be read as input alongside sanitised issue body and sanitised human comments

---

## Historical specification text (for reference)

The following text preserves the original specification wording before deprecation, for auditors reading older issue bodies that still contain the legacy markers.

> The implementation-research block SHALL be delimited by exactly two HTML comments: an opening marker `<!-- implementation-research:start -->` and a closing marker `<!-- implementation-research:end -->`. The markers SHALL appear at the start of their own lines. Issue bodies SHALL contain at most one such block. Producers (the `research-factory` workflow) and consumers (the `change-factory` workflow) SHALL identify the block by these markers and SHALL NOT rely on framework-generated marker formats such as `gh-aw-begin-*`.
>
> The block SHALL begin (immediately after the opening marker) with a `## Implementation research` heading followed by a provenance header that records the run timestamp, a link to the workflow run that authored the block, and an explicit notice that edits inside the block are read as input on the next run but are not preserved verbatim.
>
> The block SHALL contain the following subsections in order, each introduced by a level-3 heading:
> 1. `### Problem framing`
> 2. `### Approaches considered`
> 3. `### Recommendation`
> 4. `### Open questions`
> 5. `### Out of scope`
> 6. `### References`
>
> The block SHALL be the only mutable region of the issue body managed by the workflow. Each successful research run SHALL replace the entire block contents with a freshly synthesized block. Content of the issue body outside the markers SHALL be preserved byte-for-byte by the producer relative to the pre-run original issue content (the issue body with any prior block removed).
>
> On any run where a prior block exists, the producer SHALL read the prior block contents and the human-authored comment history as additional inputs alongside the original issue content. The producer SHALL NOT treat the prior block as authoritative output to preserve verbatim; rather it SHALL integrate the evidence the prior block carries into the synthesis of the new block.
