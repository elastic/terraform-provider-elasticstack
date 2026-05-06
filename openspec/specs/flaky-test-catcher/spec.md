# flaky-test-catcher Specification

## Purpose
TBD - created by archiving change flaky-test-catcher. Update Purpose after archive.
## Requirements
### Requirement: Workflow triggers
The flaky-test-catcher workflow SHALL run on a daily schedule and SHALL support manual `workflow_dispatch` triggering.

#### Scenario: Daily scheduled run
- **WHEN** the daily cron schedule fires
- **THEN** the workflow starts and executes the pre-activation job

#### Scenario: Manual dispatch
- **WHEN** a user triggers the workflow via `workflow_dispatch`
- **THEN** the workflow starts and executes the pre-activation job

---

### Requirement: Pre-activation CI failure scan
The pre-activation job SHALL query the GitHub Actions API for all workflow runs of the `.github/workflows/test.yml` workflow on `main` within the last 3 days and identify those with `conclusion == 'failure'`.

#### Scenario: Failures present
- **WHEN** one or more workflow runs on `main` in the last 3 days have `conclusion == 'failure'`
- **THEN** pre-activation outputs `has_ci_failures = 'true'`, a JSON array of failed run IDs as `failed_run_ids`, and the total run count as `total_run_count`

#### Scenario: No failures present
- **WHEN** all workflow runs on `main` in the last 3 days have `conclusion != 'failure'` (or there are no runs)
- **THEN** pre-activation outputs `has_ci_failures = 'false'` and the agent job is skipped

#### Scenario: SNAPSHOT failures excluded
- **WHEN** a SNAPSHOT matrix cell's acceptance test step fails but the job concludes as `success` due to `continue-on-error: true` on the step
- **THEN** that run is NOT counted as a failure and is excluded from analysis

---

### Requirement: Pre-activation issue slot check
The pre-activation job SHALL count open GitHub issues labelled `flaky-test` (excluding pull requests) and compute available issue slots against a cap of 3.

#### Scenario: Slots available
- **WHEN** fewer than 3 open `flaky-test` issues exist
- **THEN** pre-activation outputs `issue_slots_available` as the remaining capacity and `open_issues` as the current count

#### Scenario: Cap reached
- **WHEN** 3 or more open `flaky-test` issues already exist
- **THEN** pre-activation outputs `issue_slots_available = '0'` and the agent job is skipped via the `if` gate

---

### Requirement: Agent activation gate
The agent job SHALL only run when both `has_ci_failures == 'true'` AND `issue_slots_available != '0'`.

#### Scenario: Both gates pass
- **WHEN** pre-activation reports CI failures and available issue slots
- **THEN** the agent job starts

#### Scenario: Either gate fails
- **WHEN** pre-activation reports no CI failures OR no available issue slots
- **THEN** the agent job is skipped without error

---

### Requirement: Test failure extraction
The agent SHALL fetch job logs for each failed run ID from pre-activation, identify jobs with `conclusion == 'failure'` in the `Matrix Acceptance Test` job group, and extract failing test names matching the Go test output pattern `--- FAIL: <TestName>`.

#### Scenario: Failures found in logs
- **WHEN** a failed run's job logs contain `--- FAIL: TestAccXxx`
- **THEN** the agent records `TestAccXxx` as a failing test for that run

#### Scenario: No recognisable test failures in log
- **WHEN** a failed run's job logs do not contain `--- FAIL:` patterns (e.g., infrastructure failure before tests ran)
- **THEN** the agent skips that run and continues with remaining failed runs

---

### Requirement: Failure rate classification
The agent SHALL compute a fail rate for each test as `fail_count / total_run_count` and classify it as **broken** (rate = 1.0) or **flaky** (rate ≥ 0.20 and < 1.0). Tests with a fail rate below 0.20 SHALL be ignored.

#### Scenario: Broken test
- **WHEN** a test appears in the failure logs of every workflow run in the analysis window
- **THEN** the test is classified as **broken**

#### Scenario: Flaky test
- **WHEN** a test appears in the failure logs of ≥ 20% but < 100% of workflow runs in the analysis window
- **THEN** the test is classified as **flaky** with the computed fail rate recorded

#### Scenario: Noise (below threshold)
- **WHEN** a test appears in the failure logs of fewer than 20% of workflow runs in the analysis window
- **THEN** the test is ignored and no issue is created for it

---

### Requirement: Test grouping by base test name
The agent SHALL group failing tests by their base test name, defined as the portion of the test function name matching `TestAcc[^_]+` (i.e., everything up to but not including the first `_`). Scenario variants of the same test SHALL be grouped into one issue.

#### Scenario: Multiple scenario variants map to one group
- **WHEN** `TestAccResourceAgentConfiguration_alternateEnvironment` and `TestAccResourceAgentConfiguration_minimal` both appear as failing
- **THEN** they are grouped under a single issue titled `[flaky-test] TestAccResourceAgentConfiguration`

#### Scenario: Test with no underscore suffix
- **WHEN** `TestAccResourceAgentConfiguration` appears as failing (no `_scenario` suffix)
- **THEN** it maps to the issue title `[flaky-test] TestAccResourceAgentConfiguration` unchanged

---

### Requirement: Commit-based fix detection
For each affected resource, the agent SHALL inspect commits to `main` since the timestamp of the oldest failing run. It SHALL check both commit messages and changed file paths for references to the resource name, test names, or related keywords (e.g., "fix", "flaky"). If a relevant commit is found, the issue body SHALL note "may already be addressed in `<sha>`" rather than suppressing issue creation.

#### Scenario: Fix commit detected
- **WHEN** a commit after the oldest failing run modifies a file matching the resource's test file path or references the resource/test name in its message
- **THEN** the issue body includes a "may already be addressed" note with the commit SHA

#### Scenario: No fix commit found
- **WHEN** no commits after the oldest failing run reference the resource or its tests
- **THEN** the issue body contains no fix-detection note

---

### Requirement: Issue deduplication
The agent SHALL check for existing open GitHub issues labelled `flaky-test` whose title matches `[flaky-test] <resource_name>` before creating a new issue. If a matching open issue exists, no new issue SHALL be created for that resource.

#### Scenario: Existing open issue
- **WHEN** an open issue titled `[flaky-test] elasticstack_elasticsearch_index` already exists
- **THEN** the agent skips issue creation for that resource

#### Scenario: No existing issue
- **WHEN** no open issue matches the resource title
- **THEN** the agent proceeds with issue creation (subject to issue slot cap)

---

### Requirement: Issue creation
The agent SHALL create one GitHub issue per affected resource (up to `issue_slots_available`). Each issue SHALL be labelled `flaky-test`, and SHALL include: broken test list, flaky test list with fail rates, commit analysis result, a sample failure excerpt, and the affected stack versions.

#### Scenario: Issue created for resource with broken and flaky tests
- **WHEN** `elasticstack_elasticsearch_index` has both broken and flaky tests
- **THEN** an issue titled `[flaky-test] elasticstack_elasticsearch_index` is created with sections for Broken Tests, Flaky Tests, Commit Analysis, Sample Failure Output, and Affected Stack Versions, labelled `flaky-test`

#### Scenario: Issue cap enforced
- **WHEN** the agent has already created `issue_slots_available` issues in this run
- **THEN** no further issues are created regardless of remaining affected resources

### Requirement: No-op when nothing actionable
The agent SHALL call `noop` with a descriptive reason when the analysis completes without creating any issues (e.g., all affected resources already have open issues, or all failures are below the 20% threshold).

#### Scenario: All resources already tracked
- **WHEN** every resource with qualifying failures already has an open `flaky-test` issue
- **THEN** the agent calls `noop` with the reason "all affected resources already have open issues"

### Requirement: Created flaky-test issues are explicitly dispatched to `code-factory`
After safe-output issue creation completes, the workflow SHALL explicitly dispatch the `code-factory` workflow once for each flaky-test issue created in the current run rather than relying on a producer-side `code-factory` label to activate implementation intake.

#### Scenario: One created flaky-test issue dispatches one implementation run
- **WHEN** the workflow creates one flaky-test issue in a run
- **THEN** it SHALL dispatch exactly one `code-factory` workflow run for that issue

#### Scenario: Multiple created flaky-test issues dispatch multiple implementation runs
- **WHEN** the workflow creates multiple flaky-test issues in a run
- **THEN** it SHALL dispatch exactly one independent `code-factory` workflow run per created issue

