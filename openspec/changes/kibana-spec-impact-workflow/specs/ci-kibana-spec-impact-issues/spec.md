## ADDED Requirements

### Requirement: Workflow runs for Kibana spec-impact analysis

The repository SHALL provide a GitHub workflow for Kibana spec-impact analysis that can run automatically for Kibana generated-client changes on the default branch and can also be started manually. The workflow SHALL evaluate a persisted baseline revision so that analysis compares the current target revision to the last processed Kibana spec-impact baseline rather than only to the immediately previous commit.

#### Scenario: Push to default branch triggers analysis
- **WHEN** a push to the default branch includes changes to the Kibana spec-impact inputs for this workflow
- **THEN** the workflow SHALL evaluate those changes against the stored baseline revision

#### Scenario: Maintainer reruns analysis manually
- **WHEN** a maintainer starts the workflow manually
- **THEN** the workflow SHALL run the same baseline-aware analysis flow without requiring a new Kibana spec change to arrive first

### Requirement: Deterministic helper emits entity impact evidence

The workflow SHALL use deterministic repository-local helper tooling to derive the canonical Kibana entity inventory from provider registrations, diff tracked Kibana generated-client inputs against the selected baseline, and emit structured impact evidence for matching entities. In V1, the helper SHALL only claim high-confidence impact for entities that can be matched through `generated/kbapi` and `internal/clients/kibanaoapi` usage.

#### Scenario: Helper discovers registered entities from provider code
- **WHEN** the helper prepares entity inventory for a workflow run
- **THEN** it SHALL derive Kibana resources and data sources from the repository's registered provider implementations rather than from a handwritten entity manifest

#### Scenario: Helper reports matched entities with evidence
- **WHEN** the helper finds changed Kibana client symbols referenced by a supported entity
- **THEN** it SHALL emit structured evidence that includes the impacted entity name, entity type, matched implementation path, and the changed methods or types that produced the match

#### Scenario: Unsupported client surfaces remain out of high-confidence scope
- **WHEN** a changed Kibana-facing entity depends only on unsupported V1 client surfaces such as legacy `go-kibana-rest` or `generated/slo`
- **THEN** the helper SHALL NOT classify that entity as a high-confidence impacted entity in V1

### Requirement: Agent creates one issue per impacted entity from deterministic evidence

The workflow agent SHALL use deterministic helper output as the primary source of truth for impacted entities and SHALL create at most one issue per impacted entity in a run. Each created issue SHALL summarize the affected Terraform entity, the evidence that caused the match, and the likely impact in terms of new fields, widened options, or newly exposed capability when that can be inferred from the deterministic evidence and local repository context.

#### Scenario: High-confidence match creates a focused issue
- **WHEN** the helper reports a high-confidence impacted entity
- **THEN** the agent SHALL create one issue for that entity summarizing the matched evidence and likely provider follow-up areas

#### Scenario: Weak generic evidence does not force an issue
- **WHEN** the helper output indicates only broad or weak evidence that is not actionable for a single entity
- **THEN** the agent SHALL refrain from creating an issue for that weak match

### Requirement: Workflow suppresses duplicate impact issues

The workflow SHALL persist repo-memory state for processed Kibana spec-impact runs and SHALL use that state to suppress duplicate issue creation for equivalent entity impacts. Duplicate suppression SHALL be based on deterministic impact identity such as baseline revision, target revision, and entity-level impact fingerprint rather than only on issue-title matching.

#### Scenario: Equivalent impact is not reopened
- **WHEN** a workflow run encounters an entity impact that matches a previously processed deterministic impact identity
- **THEN** the workflow SHALL NOT create a duplicate issue for that entity

#### Scenario: New target revision can create a fresh issue
- **WHEN** a later workflow run reaches a new target revision or a changed entity-level impact fingerprint for the same entity
- **THEN** the workflow SHALL treat that impact as eligible for a new issue

### Requirement: Issue cap does not suppress never-filed entities

The workflow SHALL cap the number of new issues created per run while ensuring duplicate-suppression fingerprints are recorded **only** for entities that actually received an issue in that run. Entities that were eligible but not issued due to the cap SHALL remain eligible in a future run until an issue is created for them.

#### Scenario: Capped entity is not fingerprinted
- **WHEN** more high-confidence impacted entities exist than the per-run issue cap allows to be filed
- **THEN** repo memory SHALL NOT record a dedupe fingerprint for entities that did not receive an issue in that run

### Requirement: Baseline advances after successful analysis

The helper/workflow SHALL advance the persisted last-analyzed target revision after each successful analysis completion, including when **zero** new high-confidence issues are created or when all matches were duplicate-suppressed, so future diffs are not stuck on stale baselines.

#### Scenario: Analysis with no new issues still advances baseline
- **WHEN** a workflow run completes successfully and no new issues are created
- **THEN** the persisted baseline SHALL still advance to the analyzed target revision

#### Scenario: Partial issuance still advances baseline
- **WHEN** some high-confidence entities receive issues under the cap but others do not
- **THEN** the baseline SHALL advance and memory SHALL record fingerprints only for entities that were actually issued

### Requirement: Generated kbapi file lifecycle is tolerantly diffed

The helper SHALL tolerate `generated/kbapi/kibana.gen.go` being absent at the baseline revision, target revision, or both (introduce/rename/remove lifecycle) without failing the entire impact report; it SHALL treat missing content as an empty kbapi surface for that side of the diff.

#### Scenario: Baseline revision lacks generated kbapi file
- **WHEN** `git show` for the baseline revision cannot resolve `generated/kbapi/kibana.gen.go` because the path does not exist in that commit
- **THEN** the helper SHALL still emit an impact report (treating that side as an empty kbapi surface) instead of failing the entire command

### Requirement: `memory-record-from-report` issuance contract

The `memory-record-from-report` helper SHALL require an `--issued` JSON file path **when** the report contains one or more `high_confidence_impacts` entries, so accidental omission cannot advance the baseline without explicit issuance intent (including an empty issuance list via `[]` when no issues were intentionally created despite impacts).

#### Scenario: High-confidence report requires issued file
- **WHEN** the impact report includes at least one `high_confidence_impacts` entry
- **THEN** the memory-record command SHALL fail unless `--issued` is provided

### Requirement: Root internal/kibana entity scan robustness

For Terraform entities implemented in the root `internal/kibana` package, the helper SHALL fall back to scanning the full package directory when no filename prefix mapping exists, so new entities do not fail inventory impact scanning solely for lack of a prefix table entry.

#### Scenario: New root-package Kibana entity without prefix mapping
- **WHEN** a Terraform entity uses the root `internal/kibana` Go package and has no dedicated filename-prefix mapping yet
- **THEN** the helper SHALL still select implementation sources under that package directory for impact matching without returning an inventory error
