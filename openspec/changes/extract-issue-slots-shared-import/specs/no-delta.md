# No Spec Changes

This change is a pure structural refactoring of workflow implementation. It introduces no new capabilities and modifies no spec-level requirements.

The following existing capabilities remain fully valid and unchanged; only their internal workflow source structure is consolidated:

- `ci-duplicate-code-detector`
- `ci-schema-coverage-rotation-issue-slots`
- `ci-semantic-refactor-workflow`

All behavioral requirements (issue-slot gating logic, output names, prompt content, labels, caps) are preserved exactly.
