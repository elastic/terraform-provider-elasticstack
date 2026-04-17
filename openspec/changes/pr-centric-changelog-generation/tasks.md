## 1. PR changelog contract helpers

- [ ] 1.1 Add repository-authored helpers to parse and validate the PR-body `## Changelog` contract, including `Customer impact`, `Summary`, and boundary-based extraction of an optional free-form `### Breaking changes` subsection.
- [ ] 1.2 Add fixtures and unit tests for valid sections, malformed structured fields, `Customer impact: none`, and free-form breaking-changes markdown with lists and fenced code blocks.

## 2. PR-time agentic workflow

- [ ] 2.1 Add the new PR changelog authoring workflow source and compiled artifacts, triggered from `workflow_run` on `Build/Lint/Test` completion for pull-request events.
- [ ] 2.2 Implement deterministic pull-request resolution, `no-changelog` skip logic, and format validation so the agent only runs when a required changelog section is missing.
- [ ] 2.3 Implement the metadata-only agent prompt and PR body update path so missing `## Changelog` sections are drafted from the PR title and description without checking out or executing PR code.

## 3. Deterministic changelog assembly workflow

- [ ] 3.1 Refactor the changelog-generation workflow source and compiled outputs to remove merged-history agent synthesis and instead gather merged PR bodies and labels for the authoritative release range.
- [ ] 3.2 Implement deterministic rendering from parsed PR-body changelog sections, excluding `no-changelog` PRs and `Customer impact: none`, while preserving optional `### Breaking changes` blocks under the top-level breaking-changes section.
- [ ] 3.3 Keep output normalization minimal by standardizing only bullet/citation/whitespace shape and breaking-change placement, without semantically rewriting author-provided content.

## 4. Branch and PR management

- [ ] 4.1 Replace GH AW safe-output PR management in scheduled/manual mode with normal GitHub Actions logic that updates the `generated-changelog` branch, reuses an existing PR when present, and creates the PR when absent.
- [ ] 4.2 Implement release-mode PR metadata refresh using the triggering release PR number from event metadata while updating only the target `prep-release-*` branch.

## 5. Verification and rollout

- [ ] 5.1 Add or update tests for workflow gating, parser/renderer behavior, singleton generated-changelog PR reuse, and release-PR update logic.
- [ ] 5.2 Regenerate compiled workflow artifacts and run the relevant workflow/unit test suite plus `make check-openspec`, fixing any resulting issues.
- [ ] 5.3 After maintainer review of the workflow ergonomics, make the PR-time changelog authoring workflow a required pull-request check.
