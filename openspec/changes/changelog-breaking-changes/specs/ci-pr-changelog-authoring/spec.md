## MODIFIED Requirements

### Requirement: Existing changelog section is validated deterministically
The workflow SHALL parse and validate the `## Changelog` section from the pull request body using `parseChangelogSectionFull` and `validateChangelogSectionFull`. The validator SHALL require `Customer impact` to be exactly one of `none`, `fix`, `enhancement`, or `breaking` (case-sensitive). The validator SHALL require a `Summary` line when `Customer impact` is not `none`. The validator SHALL reject a `### Breaking changes` subsection that is present but empty. The validator SHALL require that subsection when `Customer impact` is `breaking`. The validator SHALL reject a `### Breaking changes` subsection when `Customer impact` is any valid value other than `breaking`; the error message SHALL be: `### Breaking changes section requires Customer impact: breaking; use <!-- /breaking-changes --> as an end marker.`

When validation fails, the workflow SHALL post or update a PR comment identifying the failure reason. When validation passes, the workflow SHALL update any existing failure comment to indicate the check passed.

#### Scenario: Valid changelog section passes the check
- **WHEN** the pull request body contains a `## Changelog` section that satisfies all validation rules
- **THEN** the workflow SHALL succeed, and if a prior failure comment exists it SHALL be updated to a "check passed" message

#### Scenario: Malformed changelog section fails with comment
- **WHEN** the pull request body contains a `## Changelog` section that does not satisfy the validation rules
- **THEN** the workflow SHALL fail and SHALL upsert a PR comment listing each validation error

#### Scenario: Missing changelog section fails with comment
- **WHEN** the pull request body contains no `## Changelog` section and the PR does not carry the `no-changelog` label
- **THEN** the workflow SHALL fail and SHALL upsert a PR comment stating that no `## Changelog` section was found

#### Scenario: Breaking changes with non-breaking impact fails with descriptive error
- **WHEN** the pull request body contains `### Breaking changes` content and `Customer impact` is `fix`, `enhancement`, or `none`
- **THEN** the workflow SHALL fail and SHALL upsert a PR comment containing the error `### Breaking changes section requires Customer impact: breaking; use <!-- /breaking-changes --> as an end marker.`

#### Scenario: Invalid impact value with breaking changes does not emit the breaking-impact-mismatch error
- **WHEN** the pull request body contains `### Breaking changes` content and `Customer impact` is an unsupported value (e.g., `patch`)
- **THEN** the workflow SHALL fail with an "invalid Customer impact" error only; it SHALL NOT also emit the error `### Breaking changes section requires Customer impact: breaking; use <!-- /breaking-changes --> as an end marker.`

### Requirement: Breaking changes subsection may be free-form markdown with optional end marker
Within the `## Changelog` contract, the `### Breaking changes` subsection is permitted only when `Customer impact` is `breaking`. When present, it SHALL allow free-form markdown content, including prose, bullet lists, and fenced code blocks. Validation SHALL treat that subsection as a delimited markdown block rather than a structured schema.

The subsection SHALL end at the first of: (a) an end marker line, (b) the next `##`- or `###`-level heading outside a fenced code block, or (c) the end of the `## Changelog` section. An end marker line is a line whose full content, after stripping any leading and trailing whitespace, satisfies the pattern `<!--\s*/breaking-changes\s*-->` (case-sensitive) — that is, the line contains nothing except the HTML comment with optional internal whitespace around `/breaking-changes`; it is recognised only when it appears outside a fenced code block and after the `### Breaking changes` heading. In implementation terms this corresponds to the anchored regex `/^\s*<!--\s*\/breaking-changes\s*-->\s*$/`. A marker appearing before the `### Breaking changes` heading, or inside a fenced code block, SHALL be ignored.

#### Scenario: Breaking changes block contains fenced code
- **WHEN** the pull request body includes `### Breaking changes` with fenced code blocks or migration prose
- **THEN** the workflow SHALL accept that subsection as valid when the block is non-empty

#### Scenario: End marker stops extraction before the next heading
- **WHEN** the pull request body includes `<!-- /breaking-changes -->` after breaking-change prose but before the next heading or end of changelog section
- **THEN** the parser SHALL include only the content before the marker in `breakingChanges` and SHALL exclude any content after the marker

#### Scenario: End marker inside a fenced code block is ignored
- **WHEN** the pull request body contains `<!-- /breaking-changes -->` inside a backtick- or tilde-fenced code block within the `### Breaking changes` subsection
- **THEN** the parser SHALL NOT treat it as an end marker; extraction SHALL continue past it

#### Scenario: End marker before the breaking changes heading is ignored
- **WHEN** the pull request body contains `<!-- /breaking-changes -->` in the `## Changelog` section but before the `### Breaking changes` heading
- **THEN** the parser SHALL ignore it; it SHALL NOT affect the start or boundary of `### Breaking changes` extraction

#### Scenario: End marker with internal whitespace is recognised
- **WHEN** the pull request body contains `<!--  /breaking-changes  -->` (extra spaces around the tag name) inside the `### Breaking changes` subsection
- **THEN** the parser SHALL treat it as a valid end marker and stop extraction at that line

### Requirement: PR template default state fails the changelog check
The pull request template SHALL pre-fill the `## Changelog` block with placeholder text that is not a valid `Customer impact` value, ensuring contributors must consciously replace it before the PR can pass the changelog check. The placeholder for `Customer impact` SHALL be `<none, fix, enhancement, breaking>` and for `Summary` SHALL be `<single line summary>`.

#### Scenario: Unedited template body fails the check
- **WHEN** a contributor opens a pull request without editing the `## Changelog` section of the template
- **THEN** the workflow SHALL fail because `Customer impact: <none, fix, enhancement, breaking>` is not a valid impact value

### Requirement: PR template documents the breaking example and end marker
The pull request template SHALL include a "Good example" block for the `breaking` impact level that shows: a `Customer impact: breaking` line, a `Summary:` line, a `### Breaking changes` block with a short prose description, and the `<!-- /breaking-changes -->` end marker immediately after the breaking-change content. The template instructions SHALL note that `<!-- /breaking-changes -->` is optional and ends the breaking-changes block early.

#### Scenario: Contributor can copy the breaking example as a starting point
- **WHEN** a contributor opens a pull request intending to document a breaking change
- **THEN** the template SHALL provide a concrete, copyable example of the complete `breaking` format including the end marker

### Requirement: Verifier failure comment documents the end marker
The failure comment posted by the PR changelog check workflow SHALL include `<!-- /breaking-changes -->` in its "Expected format" block, immediately after the `<free-form markdown>` line for the `### Breaking changes` section.

#### Scenario: Failure comment shows end marker in expected format
- **WHEN** the PR changelog check fails and posts a failure comment
- **THEN** the comment's "Expected format" block SHALL show `<!-- /breaking-changes -->` on the line after `<free-form markdown>  (required when Customer impact is "breaking")`
