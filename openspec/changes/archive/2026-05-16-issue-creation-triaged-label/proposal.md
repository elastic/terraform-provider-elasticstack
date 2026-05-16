## Why

Four automated analysis workflows — **Schema Coverage Rotation**, **Duplicate Code Detector**,
**Semantic Function Refactor**, and **Flaky Test Catcher** — create fully defined, actionable GitHub
issues without the `triaged` label. Because the **Issue Classifier** workflow uses
`labels.includes('triaged')` as its pre-flight gate, it picks up every one of these automated
issues on its next scan and wastes a classification run on issues that have no routing ambiguity.
All four workflows produce issues with enough context for automated remediation; they do not need
the classifier's routing decision.

## What Changes

Add `triaged` to the `safe-outputs.create-issue.labels` list in the source template of each of the
four producer workflows. The change stamps `triaged` onto issues atomically at creation time so the
classifier's gate sees it immediately.

| Workflow | Source template | Labels line (before) |
|---|---|---|
| Schema Coverage Rotation | `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl:43` | `[testing, acceptance-tests, schema-coverage]` |
| Duplicate Code Detector | `.github/workflows-src/duplicate-code-detector/workflow.md.tmpl:31` | `[duplicate-code, code-quality, automated-analysis]` |
| Semantic Function Refactor | `.github/workflows-src/semantic-function-refactor/workflow.md.tmpl:29` | `[semantic-refactor, refactoring, code-quality, automated-analysis]` |
| Flaky Test Catcher | `.github/workflows-src/flaky-test-catcher/workflow.md.tmpl:31` | `[flaky-test]` |

After each template update, `make workflow-generate` regenerates the paired `.lock.yml` file so
the compiled output stays in sync with the source.

## Capabilities

### Modified Capabilities

- `ci-automated-issue-triaged-label`: Requirements for adding `triaged` at creation time in all
  four producer workflow source templates and regenerating their paired `.lock.yml` files.

## Impact

- **Workflow sources**: Four single-line label additions in `.md.tmpl` files; the four corresponding
  compiled `.lock.yml` files must be regenerated.
- **Issue Classifier**: No changes needed — it already handles `triaged` correctly in its gate logic.
- **Backward compatibility**: Existing open issues are not retroactively labelled; only future issues
  created by these workflows carry the new label.
- **No race condition**: `triaged` is applied atomically by the `create-issue` safe-output at issue
  creation time, before the classifier can scan.
