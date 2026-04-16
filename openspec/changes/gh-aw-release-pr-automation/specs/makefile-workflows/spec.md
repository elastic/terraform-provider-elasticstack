## ADDED Requirements

### Requirement: Release preparation workflow dispatch target
The Makefile SHALL provide a maintainer-facing target that dispatches the release preparation GitHub workflow through `gh workflow run` instead of performing release mutation locally. The target SHALL accept a bump-mode input that supports `patch`, `minor`, and `major`, SHALL default that input to `patch`, and SHALL reject unsupported bump values before invoking `gh`.

#### Scenario: Default bump input dispatches patch release preparation
- **GIVEN** a maintainer runs the release preparation Make target without overriding the bump mode
- **WHEN** the target dispatches the workflow
- **THEN** it SHALL invoke `gh workflow run` using `patch` as the bump input

#### Scenario: Unsupported bump value fails before dispatch
- **GIVEN** a maintainer supplies a bump value other than `patch`, `minor`, or `major`
- **WHEN** the Make target validates its inputs
- **THEN** it SHALL fail before dispatching the workflow

#### Scenario: Make target does not duplicate release logic
- **GIVEN** a maintainer uses the release preparation Make target
- **WHEN** the target runs successfully
- **THEN** it SHALL only dispatch the GitHub workflow rather than editing `Makefile`, editing `CHANGELOG.md`, creating branches, or opening pull requests locally
