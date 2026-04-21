## Why

The current changelog workflow tries to synthesize release notes from merged pull requests at release-assembly time, which has led to title-shaped entries, weak customer-impact filtering, and output that drifts from the established `CHANGELOG.md` voice. Moving changelog authorship closer to each pull request lets authors and reviewers supply the intent while keeping release assembly deterministic and consistent.

## What Changes

- Add a new PR-time agentic workflow that runs after `Build/Lint/Test` completes, validates whether a pull request already contains a valid `## Changelog` section, and only invokes the agent when the PR is missing that section and is not labeled `no-changelog`.
- Define a structured PR-body changelog contract with required `Customer impact` and `Summary` fields plus an optional free-form `### Breaking changes` markdown subsection that may include prose, lists, and fenced code blocks.
- Refactor changelog release assembly so scheduled/manual and release-preparation runs deterministically parse merged PR bodies and labels, rebuild changelog sections from those structured PR entries, and stop relying on agentic PR-history summarization.
- Replace GH AW safe-output PR management in changelog release assembly with normal GitHub Actions branch and pull-request update logic, including singleton `generated-changelog` PR reuse and release-PR updates from event metadata.
- Preserve authoritative full-section regeneration for `## [Unreleased]` and concrete release sections, while allowing any merged-PR-triggered follow-on workflow only as a convenience rerun of the same deterministic rebuild logic rather than as a separate append-only path.

## Capabilities

### New Capabilities

- `ci-pr-changelog-authoring`: A PR-time workflow that ensures pull requests either contain a valid `## Changelog` section or explicitly opt out with `no-changelog`, and that can draft the missing PR-body changelog section from the PR title and description.

### Modified Capabilities

- `ci-changelog-generation`: Replace agentic merged-history summarization with deterministic assembly from merged PR-body changelog sections, preserve optional breaking-change blocks, and manage generated/release PR updates with standard GitHub Actions logic instead of GH AW safe outputs.

## Impact

- **Workflow sources:** new PR-time workflow under `.github/workflows-src/` plus the compiled workflow under `.github/workflows/`; significant changes to `.github/workflows-src/changelog-generation/workflow.md.tmpl` and its compiled outputs.
- **Deterministic helpers:** new or updated repository-authored parsers/validators for PR-body changelog sections, merged PR aggregation, breaking-change block extraction, and branch/PR update logic.
- **Specifications:** new canonical spec for the PR-time workflow capability and updates to `openspec/specs/ci-changelog-generation/spec.md`; possibly minor alignment in related CI specs where they reference changelog generation behavior.
- **Repository process:** pull requests gain a required changelog check after `Build/Lint/Test`, and changelog intent moves from release-time inference to PR-time authoring/review.
