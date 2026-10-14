# ci-research-factory-dispatch-fanout Specification

## Purpose

Lets issue-producing analysis workflows hand each newly created issue to `research-factory` automatically, deterministically, and exactly once per issue.

## Requirements

### Requirement: Producer workflows dispatch `research-factory` once per created issue
Repository-authored issue-producing analysis workflows that hand off to `research-factory` SHALL, after issue creation, dispatch the `research-factory` workflow once for each issue created in the current run, passing that issue's number and a `source_workflow` identifying the producer.

#### Scenario: Multiple created issues are fanned out deterministically
- **GIVEN** a producer workflow that creates multiple issues in one run
- **WHEN** issue creation completes
- **THEN** the workflow SHALL dispatch one independent `research-factory` run per created issue
- **AND** each dispatch SHALL target exactly one issue number
- **AND** each dispatch SHALL carry the producer's `source_workflow` identifier

#### Scenario: No issues created means no dispatches
- **GIVEN** a producer workflow run that creates no issues
- **WHEN** the run completes
- **THEN** it SHALL NOT dispatch `research-factory`

### Requirement: Research fan-out uses the safe-output temporary issue ID map
The dispatch phase SHALL use the safe-output temporary issue ID map as the source of truth for created issue identities, reading the repository slug and issue number for each entry rather than inferring newly created issues by search, title, or label matching.

#### Scenario: Temporary issue ID map contains two created issues
- **GIVEN** safe-output artifacts with two temporary issue ID map entries
- **WHEN** the dispatch phase runs
- **THEN** it SHALL dispatch `research-factory` exactly twice using those repository and issue-number values

#### Scenario: Temporary issue ID map is missing or malformed
- **GIVEN** a dispatch phase that cannot read a valid temporary issue ID map
- **WHEN** it attempts fan-out
- **THEN** the workflow SHALL fail or stop fan-out rather than guessing which issues were created

### Requirement: Research fan-out remains deterministic and repository-authored
The logic that determines which issues to dispatch to `research-factory` SHALL be deterministic repository-authored workflow jobs or scripts and SHALL NOT depend on agent-emitted safe outputs to determine which created issues to dispatch. It SHALL run only after safe-output issue creation has completed.

#### Scenario: Maintainer inspects dispatch handoff implementation
- **GIVEN** a producer workflow source that hands off to `research-factory`
- **WHEN** maintainers review it
- **THEN** the logic that parses created issue identities and dispatches `research-factory` SHALL be repository-authored and deterministic
- **AND** it SHALL run after safe-output issue creation has completed

### Requirement: `research-factory` accepts dispatches from `github-actions[bot]`
The `research-factory` workflow SHALL allow `github-actions[bot]` through its pre-activation team-membership check (via the `on.bots` allow-list, compiled to `GH_AW_ALLOWED_BOTS`), so that producer dispatches executed with the workflow `GITHUB_TOKEN` are activated rather than skipped for lacking repository permission. This SHALL NOT change which issue events qualify for human-triggered intake.

#### Scenario: Producer dispatch runs as `github-actions[bot]`
- **GIVEN** a producer workflow dispatches `research-factory` using its `GITHUB_TOKEN`
- **WHEN** the `research-factory` pre-activation job checks the triggering actor
- **THEN** the actor `github-actions[bot]` SHALL be allowed
- **AND** activation and the agent job SHALL run
