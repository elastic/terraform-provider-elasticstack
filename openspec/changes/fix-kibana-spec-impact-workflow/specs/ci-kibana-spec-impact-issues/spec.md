## MODIFIED Requirements

### Requirement: Workflow runs for Kibana spec-impact analysis

The repository SHALL provide a GitHub workflow for Kibana spec-impact analysis that can run automatically for Kibana generated-client changes on the default branch and can also be started manually. The workflow SHALL evaluate a persisted baseline revision so that analysis compares the current target revision to the last processed Kibana spec-impact baseline rather than only to the immediately previous commit. Before deterministic pre-activation analysis runs, the workflow SHALL initialize the repository checkout and the configured repo-memory context needed for baseline-aware computation. The repo-memory configuration SHALL declare the memory branch explicitly, and pre-activation initialization SHALL use that same configured branch.

#### Scenario: Push to default branch triggers analysis
- **WHEN** a push to the default branch includes changes to the Kibana spec-impact inputs for this workflow
- **THEN** the workflow SHALL evaluate those changes against the stored baseline revision

#### Scenario: Maintainer reruns analysis manually
- **WHEN** a maintainer starts the workflow manually
- **THEN** the workflow SHALL run the same baseline-aware analysis flow without requiring a new Kibana spec change to arrive first

#### Scenario: Pre-activation computes against initialized repo memory
- **WHEN** the workflow enters its pre-activation phase to determine whether agent analysis should run
- **THEN** the workflow SHALL have already checked out the repository and initialized the configured repo-memory path used for Kibana spec-impact baseline state

#### Scenario: Pre-activation init uses explicit repo-memory branch
- **WHEN** the workflow configures repo memory for Kibana spec-impact analysis
- **THEN** it SHALL declare the repo-memory `branch-name` explicitly and the pre-activation checkout/init step SHALL target that same branch rather than relying on an implicit default

### Requirement: Deterministic helper emits entity impact evidence

The workflow SHALL use deterministic repository-local helper tooling to derive the canonical Kibana entity inventory from provider registrations, diff tracked Kibana generated-client inputs against the selected baseline, emit structured impact evidence for matching entities, and derive workflow gate outputs from that deterministic report. In V1, the helper SHALL only claim high-confidence impact for entities that can be matched through `generated/kbapi` and `internal/clients/kibanaoapi` usage.

#### Scenario: Helper discovers registered entities from provider code
- **WHEN** the helper prepares entity inventory for a workflow run
- **THEN** it SHALL derive Kibana resources and data sources from the repository's registered provider implementations rather than from a handwritten entity manifest

#### Scenario: Helper reports matched entities with evidence
- **WHEN** the helper finds changed Kibana client symbols referenced by a supported entity
- **THEN** it SHALL emit structured evidence that includes the impacted entity name, entity type, matched implementation path, and the changed methods or types that produced the match

#### Scenario: Helper derives workflow gate from deterministic report
- **WHEN** pre-activation computes the Kibana spec-impact report
- **THEN** the repository-local helper SHALL also emit the gate values needed by the workflow, including whether the agent should run, the issue cap, the high-confidence count, and the gate reason

#### Scenario: Unsupported client surfaces remain out of high-confidence scope
- **WHEN** a changed Kibana-facing entity depends only on unsupported V1 client surfaces such as `generated/slo`
- **THEN** the helper SHALL NOT classify that entity as a high-confidence impacted entity in V1

## ADDED Requirements

### Requirement: Deterministic report is handed to the agent via artifact

The workflow SHALL preserve the deterministic Kibana spec-impact report across job boundaries by uploading it as a workflow artifact during pre-activation and downloading it into `/tmp/gh-aw/agent` for the agent job. The agent SHALL read the report from the downloaded artifact location and SHALL write any run-local JSON support files for this flow under `/tmp/gh-aw/agent`, while continuing to persist durable repo-memory state under the configured `/tmp/gh-aw/repo-memory/...` path.

#### Scenario: Pre-activation uploads the deterministic report
- **WHEN** pre-activation successfully generates the Kibana spec-impact report
- **THEN** the workflow SHALL upload that report as an artifact for later jobs instead of relying on a repo-root workspace file surviving job handoff

#### Scenario: Agent consumes downloaded report artifact
- **WHEN** the agent job starts after a successful pre-activation run
- **THEN** the workflow SHALL download the report artifact into `/tmp/gh-aw/agent`, and the agent instructions SHALL reference that downloaded path as the source of truth for report input

#### Scenario: Agent writes issued entities beside the downloaded report
- **WHEN** the agent records which high-confidence entities actually received issues in the current run
- **THEN** it SHALL write the issued-entities JSON file under `/tmp/gh-aw/agent` and use that file when invoking repo-memory persistence
