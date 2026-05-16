# `issue-creation-triaged-label` — Apply `triaged` label at issue creation in automated producer workflows

Workflow implementation: four `workflow.md.tmpl` source templates under
`.github/workflows-src/<workflow>/`, each compiled into a paired `.lock.yml` under
`.github/workflows/`.

## Purpose

Define requirements for stamping `triaged` onto GitHub issues at creation time in the automated
producer workflows that generate fully defined, actionable issues. This prevents the Issue
Classifier from wasting a classification run on issues that require no routing decision.

## Affected workflows

| Workflow | Source template | Compiled lock file |
|---|---|---|
| Schema Coverage Rotation | `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` | `.github/workflows/schema-coverage-rotation.lock.yml` |
| Duplicate Code Detector | `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl` | `.github/workflows/duplicate-code-detector.lock.yml` |
| Semantic Function Refactor | `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl` | `.github/workflows/semantic-function-refactor.lock.yml` |
| Flaky Test Catcher | `.github/workflows-src/flaky-test-catcher/workflow.md.tmpl` | `.github/workflows/flaky-test-catcher.lock.yml` |

## ADDED Requirements

### Requirement: `triaged` applied at issue creation time (REQ-001)

Each of the four automated producer workflows SHALL include `triaged` in the
`safe-outputs.create-issue.labels` list of its source template. The `triaged` label SHALL be
applied as part of issue creation through that configuration.

#### Scenario: Newly created automated issue bypasses classifier

- GIVEN the Schema Coverage Rotation (or Duplicate Code Detector, Semantic Function Refactor, or
  Flaky Test Catcher) workflow creates a new issue
- WHEN the Issue Classifier runs its next scan
- THEN the new issue SHALL have the `triaged` label present
- AND the Issue Classifier pre-flight gate SHALL skip the issue without routing it

#### Scenario: `triaged` co-exists with workflow-specific labels

- GIVEN any of the four producer workflows creates an issue
- WHEN the issue's labels are inspected
- THEN both the workflow-specific labels (e.g. `schema-coverage`, `flaky-test`) AND `triaged`
  SHALL be present on the issue

### Requirement: Label declaration in source template (REQ-002)

The `triaged` label SHALL be declared in the `safe-outputs.create-issue.labels` field of each
affected source template. It SHALL NOT be added via a separate workflow step or a reactive
`issues: labeled` trigger.

The required label additions by template (append `triaged` to each existing list):

```yaml
# schema-coverage-rotation/workflow.md.tmpl
labels: [testing, acceptance-tests, schema-coverage, triaged]

# duplicate-code-detector/workflow.md.tmpl
labels: [duplicate-code, code-quality, automated-analysis, triaged]

# semantic-function-refactor/workflow.md.tmpl
labels: [semantic-refactor, refactoring, code-quality, automated-analysis, triaged]

# flaky-test-catcher/workflow.md.tmpl
labels: [flaky-test, triaged]
```

#### Scenario: Label list is visible in source template

- GIVEN a contributor reads a producer workflow's source template
- WHEN they inspect the `safe-outputs.create-issue.labels` field
- THEN `triaged` SHALL appear alongside the workflow-specific labels

### Requirement: Compiled lock files reflect label change (REQ-003)

After updating the source templates, all four paired `.lock.yml` files SHALL be regenerated via
`make workflow-generate` and committed alongside the template changes. The `GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG`
environment variable in each compiled lock file SHALL reflect the updated label list. Hand-editing
of lock files is not permitted.

#### Scenario: Lock file and source template stay in sync

- GIVEN a maintainer updates a producer template's `labels:` list
- WHEN `make workflow-generate` is run
- THEN the paired `.lock.yml` file SHALL be regenerated with the new label configuration
- AND the diff between the old and new lock file SHALL show only the expected label addition in
  the `GH_AW_SAFE_OUTPUTS_HANDLER_CONFIG` value

### Requirement: `triaged` label must exist in the repository (REQ-004)

The `triaged` GitHub label SHALL exist in the repository before the updated workflows are deployed.
Because the Issue Classifier already applies this label, it is expected to exist; this requirement
is a verification gate rather than a provisioning requirement.

#### Scenario: Label pre-exists from classifier usage

- GIVEN the Issue Classifier has previously run and applied `triaged` to classified issues
- WHEN the producer workflow templates are updated and deployed
- THEN the `triaged` label SHALL already be present in the repository's label set
- AND no manual label creation step SHALL be required
