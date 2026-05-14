## Why

The `code-factory` and `change-factory` GH AW workflows can fail during PR creation when the safe-output transport relies on Git bundle prerequisites that are not present in a later job checkout. We need a documented, frontmatter-level configuration that avoids this failure mode without relying on generated workflow internals.

## What Changes

- Update the requirements for the `code-factory` issue-intake workflow so PR creation uses the safe-outputs `am` patch transport instead of the bundle transport.
- Update the requirements for the `change-factory` issue-intake workflow so PR creation uses the safe-outputs `am` patch transport instead of the bundle transport.
- Specify that the authored workflow frontmatter is the source of truth for this setting so generated workflow artifacts inherit the safer PR creation behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `ci-code-factory-issue-intake`: Require the authored workflow to configure `safe-outputs.create-pull-request.patch-format: am`.
- `ci-change-factory-issue-intake`: Require the authored workflow to configure `safe-outputs.create-pull-request.patch-format: am`.

## Impact

Affected systems are the authored GH AW workflow sources and generated workflow artifacts for `code-factory` and `change-factory` issue intake. This changes PR creation transport behavior for those workflows but does not change provider APIs or Terraform runtime behavior.
