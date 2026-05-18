## MODIFIED Requirements

### Requirement: Existing changelog section is validated deterministically

The workflow SHALL parse and validate the `## Changelog` section from the pull request body. The validator SHALL require `Customer impact` to be exactly one of `none`, `fix`, `enhancement`, or `breaking` (case-sensitive). The validator SHALL require a `Summary` line when `Customer impact` is not `none`. The validator SHALL reject a `### Breaking changes` subsection that is present but empty. The validator SHALL require that subsection when `Customer impact` is `breaking`. The validator SHALL reject a `### Breaking changes` subsection when `Customer impact` is any valid value other than `breaking`; the error message SHALL be: `### Breaking changes section is only allowed when Customer impact: breaking; change to Customer impact: breaking or remove the ### Breaking changes heading.`

The parser and validator SHALL be the single canonical implementation owned by the `changelog-tooling` capability (`scripts/changelog/internal/section/` and `scripts/changelog/internal/prcheck/`). No parallel validator implementation SHALL exist elsewhere in the repository.

When validation fails, the workflow SHALL post or update a PR comment identifying the failure reason. When validation passes, the workflow SHALL update any existing failure comment to indicate the check passed.

#### Scenario: Valid changelog section passes the check
- **WHEN** the pull request body contains a `## Changelog` section that satisfies all validation rules
- **THEN** the workflow SHALL succeed, and if a prior failure comment exists it SHALL be updated to a "check passed" message

#### Scenario: Malformed changelog section fails with comment
- **WHEN** the pull request body contains a `## Changelog` section that does not satisfy the validation rules
- **THEN** the workflow SHALL fail and SHALL upsert a PR comment listing each validation error

#### Scenario: Missing changelog section fails with comment
- **WHEN** the pull request body has no `## Changelog` section and the `no-changelog` label is not applied
- **THEN** the workflow SHALL fail and SHALL upsert a PR comment instructing the author to add the section or apply the `no-changelog` label

#### Scenario: Validator is shared with the changelog engine
- **WHEN** the `validate-pr-section` subcommand of `scripts/changelog/` and any other subcommand that parses PR-body `## Changelog` sections are invoked on the same input
- **THEN** they SHALL produce the same parsed representation of customer impact, summary, and breaking-changes content, because they share the parser defined by the `changelog-tooling` capability
