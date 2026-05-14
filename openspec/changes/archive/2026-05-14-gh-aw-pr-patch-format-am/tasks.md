## 1. Update workflow source configuration
- [x] 1.1 Inspect the authored `code-factory` and `change-factory` workflow sources to find their `safe-outputs.create-pull-request` blocks.
- [x] 1.2 Add `patch-format: am` to both authored workflow sources.
- [x] 1.3 Regenerate the compiled workflow artifacts so the checked-in `.md` and `.lock.yml` outputs match the authored source.

## 2. Update requirements
- [x] 2.1 Add a delta spec for `ci-code-factory-issue-intake` requiring `safe-outputs.create-pull-request.patch-format: am` in the authored workflow frontmatter and generated artifacts.
- [x] 2.2 Add a delta spec for `ci-change-factory-issue-intake` requiring `safe-outputs.create-pull-request.patch-format: am` in the authored workflow frontmatter and generated artifacts.

## 3. Validate
- [x] 3.1 Run `openspec validate --all`.
- [x] 3.2 Verify the generated workflow outputs reflect the `am` patch transport for both workflows.
