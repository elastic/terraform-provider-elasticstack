## Purpose

Provides one shared gh-aw workflow import that gives any producer workflow the canonical job for dispatching `research-factory` per created issue, so producers do not copy dispatch logic.

## ADDED Requirements

### Requirement: Shared fragment defines the canonical dispatch-research-factory job
A shared fragment at `.github/workflows/shared/dispatch-research-factory.md` SHALL define a `safe-outputs.jobs.dispatch-research-factory` entry with a required boolean `dispatch` input. The job SHALL read the safe-outputs temporary-ID artifact and dispatch `research-factory-issue.lock.yml` once for each issue created in the current workflow run.

#### Scenario: Consumer workflow inherits dispatch job via import
- **GIVEN** a workflow source file that declares `imports: [shared/dispatch-research-factory.md]`
- **WHEN** it is compiled
- **THEN** the compiled lock SHALL expose `dispatch_research_factory` with a required boolean `dispatch` input
- **AND** it SHALL contain a job that reads the safe-outputs artifact and dispatches `research-factory-issue.lock.yml` for each created issue

#### Scenario: Consumer requests dispatch
- **GIVEN** a producer workflow importing the fragment
- **WHEN** it completes issue creation
- **THEN** its agent instructions SHALL require one `dispatch_research_factory` call with `dispatch: true`

#### Scenario: Fragment is importable by multiple workflows
- **GIVEN** several producer workflows that import the fragment
- **WHEN** they are compiled
- **THEN** each importing workflow's compiled lock SHALL include the `dispatch-research-factory` job

### Requirement: Research dispatch identifies the calling workflow
The fragment SHALL derive the `source_workflow` value passed to `research-factory` at runtime from the calling workflow's display name, as its lowercase, hyphen-separated form.

#### Scenario: Workflow name slug matches the consumer
- **GIVEN** a consumer workflow named "Kibana API Gap Scanner"
- **WHEN** its `dispatch-research-factory` job runs
- **THEN** each dispatch SHALL pass `source_workflow` `kibana-api-gap-scanner`

### Requirement: Research dispatch job runs after safe-outputs completion
The `dispatch-research-factory` job SHALL run only after the safe-outputs processing job completes.

#### Scenario: Dispatch runs after issue creation
- **GIVEN** a safe-outputs processing job that created one or more issues
- **WHEN** it completes
- **THEN** the `dispatch-research-factory` job SHALL start and dispatch `research-factory-issue.lock.yml` for each created issue

#### Scenario: Dispatch still runs when no issues were created
- **GIVEN** a safe-outputs processing job that created no issues
- **WHEN** it completes
- **THEN** the `dispatch-research-factory` job SHALL still run
- **BUT** it SHALL NOT dispatch any `research-factory` runs
