## REMOVED Requirements

### Requirement: Evaluation scope (REQ-001)

The script SHALL evaluate only open, non-draft pull requests in the same repository.

#### Scenario: Draft PR ignored

- GIVEN a draft pull request
- WHEN the script runs
- THEN it SHALL NOT treat that PR as eligible for auto-approval based on scope alone

**Reason**: Redundant with how CI invokes the script; removing it simplifies the canonical spec without changing other requirements.

**Migration**: Rely on workflow/job filters for open vs draft PRs; any normative “scope” behavior should be documented at the workflow level or absorbed into other requirements if needed.

## MODIFIED Requirements

### Requirement: Category routing (REQ-001)

The script SHALL evaluate pull requests against named auto-approve categories; a pull request is eligible only when it matches at least one category selector.

#### Scenario: No matching category

- GIVEN a pull request that matches no category
- WHEN gates are evaluated
- THEN the PR SHALL NOT be approved via category routing

### Requirement: Approval action (REQ-002)

When a pull request satisfies all applicable category and global gates, the script SHALL submit an approving review (`APPROVE`).

#### Scenario: All gates pass

- GIVEN a PR matches a category and passes all gates
- WHEN the script completes evaluation
- THEN it SHALL submit an `APPROVE` review

### Requirement: Approval idempotency (REQ-003)

If the acting account has already submitted an approval review, the script SHALL not create another approval review.

#### Scenario: Already approved

- GIVEN the bot already approved the PR
- WHEN the script runs again
- THEN it SHALL NOT submit a duplicate approval

### Requirement: No approval on failure (REQ-004)

If no category matches or any gate fails, the script SHALL not approve and SHALL exit successfully.

#### Scenario: Gate failure

- GIVEN a gate fails
- WHEN the script exits
- THEN it SHALL NOT approve and SHALL exit successfully

### Requirement: Observability (REQ-005)

The script SHALL emit clear, machine-readable category selection results, gate outcomes, and failure reasons.

#### Scenario: Operator inspection

- GIVEN the script runs
- WHEN output is produced
- THEN category selection, gate results, and failures SHALL be machine-readable

### Requirement: Copilot category selector (REQ-006)

The `copilot` category selector SHALL match pull requests opened by Copilot identities (`github-copilot[bot]` or `Copilot`).

#### Scenario: Copilot identity

- GIVEN a PR opened by an allowed Copilot identity
- WHEN category matching runs
- THEN the `copilot` selector MAY match

### Requirement: Copilot commit authors (REQ-007)

Every commit in a `copilot` category pull request SHALL be authored by an allowed Copilot identity (`github-copilot[bot]` or `Copilot`).

#### Scenario: Foreign commit on Copilot PR

- GIVEN a PR matched as `copilot` but a commit has a non-allowed author
- WHEN gates run
- THEN the PR SHALL NOT be approved via that category

### Requirement: Copilot file allowlist (REQ-008)

Every changed file path in a `copilot` category pull request SHALL match one of: `*_test.go` or `*.tf`.

#### Scenario: Disallowed path

- GIVEN a `copilot` PR changes a file outside the allowlist
- WHEN gates run
- THEN approval SHALL NOT proceed for that category path

### Requirement: Copilot diff threshold (REQ-009)

The total pull request line edits (`additions + deletions`) for a `copilot` category pull request SHALL be strictly less than `300`.

#### Scenario: Large Copilot PR

- GIVEN additions plus deletions are 300 or more
- WHEN the Copilot category is considered
- THEN the threshold gate SHALL fail

### Requirement: Dependabot selector (REQ-010)

The `dependabot` category selector SHALL match pull requests opened by `dependabot[bot]`.

#### Scenario: Dependabot PR

- GIVEN a PR opened by `dependabot[bot]`
- WHEN categories are evaluated
- THEN the `dependabot` selector SHALL match

### Requirement: Dependabot approval policy (REQ-011)

A pull request that matches the `dependabot` category SHALL be auto-approved when global approval gates pass.

#### Scenario: Dependabot with passing global gates

- GIVEN a Dependabot PR matching the category
- WHEN global gates pass
- THEN the script SHALL approve

### Requirement: Category extensibility (REQ-012)

Category selectors and category-specific gates SHALL be structured so that new categories can be added without modifying existing category behavior.

#### Scenario: Additive category

- GIVEN a new category is added alongside existing ones
- WHEN existing category logic runs
- THEN behavior of existing categories SHALL remain unchanged

### Requirement: Testability (REQ-013)

Category routing and gate evaluation logic SHALL be unit tested with table-driven tests that cover passing and failing scenarios for each category and threshold boundary, including shared checks-state gates and self-exclusion behavior.

#### Scenario: Unit test coverage

- GIVEN the unit test suite
- WHEN tests run
- THEN category routing and gates SHALL have table-driven coverage as specified
