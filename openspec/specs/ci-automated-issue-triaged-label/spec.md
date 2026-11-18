# ci-automated-issue-triaged-label Specification

## Purpose
TBD - created by archiving change issue-creation-triaged-label. Update Purpose after archive.
## Requirements
### Requirement: `triaged` applied at issue creation time (REQ-001)

Each of the five automated producer workflows (Schema Coverage Rotation, Duplicate Code Detector, Semantic Function Refactor, Flaky Test Catcher, and Kibana API Gap Scanner) SHALL include `triaged` in the
`safe-outputs.create-issue.labels` list of its workflow source. The `triaged` label SHALL be
applied as part of issue creation through that configuration.

#### Scenario: Newly created automated issue bypasses classifier

- GIVEN the Schema Coverage Rotation (or Duplicate Code Detector, Semantic Function Refactor,
  Flaky Test Catcher, or Kibana API Gap Scanner) workflow creates a new issue
- WHEN the Issue Classifier runs its next scan
- THEN the new issue SHALL have the `triaged` label present
- AND the Issue Classifier pre-flight gate SHALL skip the issue without routing it

#### Scenario: `triaged` co-exists with workflow-specific labels

- GIVEN any of the five producer workflows creates an issue
- WHEN the issue's labels are inspected
- THEN both the workflow-specific labels (e.g. `schema-coverage`, `flaky-test`, `kibana-api-gap`) AND `triaged`
  SHALL be present on the issue

### Requirement: Label declaration in source template (REQ-002)

The `triaged` label SHALL be declared in the `safe-outputs.create-issue.labels` field of each
affected workflow source under `.github/workflows/`. It SHALL NOT be added via a separate workflow step or a reactive
`issues: labeled` trigger.

The required label lists by workflow source:

```yaml
# schema-coverage-rotation.md
labels: [testing, acceptance-tests, schema-coverage, triaged]

# duplicate-code-detector.md
labels: [duplicate-code, code-quality, automated-analysis, triaged]

# semantic-function-refactor.md
labels: [semantic-refactor, refactoring, code-quality, automated-analysis, triaged]

# flaky-test-catcher.md
labels: [flaky-test, triaged]

# kibana-api-gap-scanner.md
labels: [kibana-api-gap, triaged]
```

#### Scenario: Label list is visible in source template

- GIVEN a contributor reads a producer workflow's source
- WHEN they inspect the `safe-outputs.create-issue.labels` field
- THEN `triaged` SHALL appear alongside the workflow-specific labels

### Requirement: Compiled lock files reflect label change (REQ-003)

After updating the workflow sources, all five paired `.lock.yml` files SHALL be regenerated via
`make workflow-generate` and committed alongside the source changes. The `GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG`
environment variable in each compiled lock file SHALL reflect the updated label list. Hand-editing
of lock files is not permitted.

#### Scenario: Lock file and source template stay in sync

- GIVEN a maintainer updates a producer workflow's `labels:` list
- WHEN `make workflow-generate` is run
- THEN the paired `.lock.yml` file SHALL be regenerated with the new label configuration
- AND the diff between the old and new lock file SHALL show only the expected label addition in
  the `GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG` value

#### Scenario: New producer workflow is compiled with the label

- GIVEN the newly added Kibana API Gap Scanner workflow source declares `triaged`
- WHEN `make workflow-generate` is run
- THEN its new `.lock.yml` file SHALL include `triaged` in the `GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG` label list

### Requirement: `triaged` label must exist in the repository (REQ-004)

The `triaged` GitHub label SHALL exist in the repository before the updated workflows are deployed.
Because the Issue Classifier already applies this label, it is expected to exist; this requirement
is a verification gate rather than a provisioning requirement.

#### Scenario: Label pre-exists from classifier usage

- GIVEN the Issue Classifier has previously run and applied `triaged` to classified issues
- WHEN the producer workflow templates are updated and deployed
- THEN the `triaged` label SHALL already be present in the repository's label set
- AND no manual label creation step SHALL be required

