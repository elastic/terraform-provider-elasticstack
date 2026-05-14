## 1. Update workflow source configuration
- [ ] 1.1 Inspect the authored `code-factory` and `change-factory` workflow sources to find their `safe-outputs.create-pull-request` blocks.
- [ ] 1.2 Add `patch-format: am` to both authored workflow sources.
- [ ] 1.3 Regenerate the compiled workflow artifacts so the checked-in `.md` and `.lock.yml` outputs match the authored source.

## 2. Update requirements
- [ ] 2.1 Add a delta spec for `ci-code-factory-issue-intake` requiring `safe-outputs.create-pull-request.patch-format: am` in the authored workflow frontmatter and generated artifacts.
- [ ] 2.2 Add a delta spec for `ci-change-factory-issue-intake` requiring `safe-outputs.create-pull-request.patch-format: am` in the authored workflow frontmatter and generated artifacts.

## 3. Validate
- [ ] 3.1 Run `openspec validate --all`.
- [ ] 3.2 Verify the generated workflow outputs reflect the `am` patch transport for both workflows.
