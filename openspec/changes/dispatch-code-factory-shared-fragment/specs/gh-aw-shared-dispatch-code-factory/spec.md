# gh-aw-shared-dispatch-code-factory — Shared dispatch-code-factory fragment

Shared gh-aw workflow fragment at `.github/workflows/shared/dispatch-code-factory.md`.

## ADDED Requirements

### Requirement: Shared fragment defines the canonical dispatch-code-factory job

The shared fragment SHALL define a `safe-outputs.jobs.dispatch-code-factory` entry that downloads the safe-outputs temporary-ID artifact and dispatches `code-factory-issue.lock.yml` once for each issue created in the current workflow run.

#### Scenario: Consumer workflow inherits dispatch job via import

- **WHEN** a workflow source file declares `imports: [shared/dispatch-code-factory.md]`
- **THEN** the compiled lock SHALL contain a `dispatch-code-factory` job that reads the safe-outputs artifact and dispatches `code-factory-issue.lock.yml` for each created issue

#### Scenario: Shared fragment replaces inline copies

- **WHEN** the shared fragment is created and consumers are updated to use the import
- **THEN** no inline `safe-outputs.jobs.dispatch-code-factory` blocks SHALL remain in `flaky-test-catcher.md`, `semantic-function-refactor.md`, or `schema-coverage-rotation.md`

### Requirement: SOURCE_WORKFLOW is derived dynamically from the calling workflow's name

The shared fragment SHALL derive `SOURCE_WORKFLOW` at runtime from the calling workflow's display name using a bash transformation:

```bash
SOURCE_WORKFLOW=$(echo "${{ github.workflow }}" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
```

The derivation SHALL produce the correct slug for all four consumer workflows:
- `"Flaky Test Catcher"` → `flaky-test-catcher`
- `"Semantic Function Refactor"` → `semantic-function-refactor`
- `"Schema Coverage Rotation"` → `schema-coverage-rotation`
- `"Duplicate Code Detector"` → `duplicate-code-detector`

#### Scenario: Workflow name slug matches expected consumer identifier

- **WHEN** the `dispatch-code-factory` job runs inside a consumer workflow
- **THEN** `SOURCE_WORKFLOW` SHALL equal the lowercase-hyphenated form of that workflow's `name:` field

### Requirement: Shared fragment follows the established shared fragment pattern

The shared fragment SHALL be placed at `.github/workflows/shared/dispatch-code-factory.md` following the `shared/setup-dev.md` precedent. It SHALL contain a YAML frontmatter block with the job definition and a brief markdown description.

#### Scenario: Fragment is importable by multiple workflows

- **WHEN** any number of producer workflows add `imports: [shared/dispatch-code-factory.md]` to their frontmatter
- **THEN** each importing workflow's compiled lock SHALL include the `dispatch-code-factory` job

### Requirement: Dispatch job runs after safe-outputs completion

The `dispatch-code-factory` job in the shared fragment SHALL declare `needs: safe_outputs` so it executes only after the safe-outputs processing job completes.

#### Scenario: Dispatch runs after issue creation

- **WHEN** the safe-outputs processing job completes having created one or more issues
- **THEN** the `dispatch-code-factory` job SHALL start and dispatch `code-factory-issue.lock.yml` for each created issue

#### Scenario: Dispatch still runs when no issues were created

- **WHEN** the safe-outputs processing job completes without creating any issues
- **THEN** the `dispatch-code-factory` job SHALL still run; the `producer-dispatch.js` script SHALL not dispatch any runs when the temporary-ID map is empty
