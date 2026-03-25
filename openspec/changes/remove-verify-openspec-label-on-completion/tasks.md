## 1. Update workflow contract and source

- [ ] 1.1 Update `.github/workflows/openspec-verify-label.md` so the workflow removes `verify-openspec` during a final completion phase for approve, comment, noop, and failure outcomes.
- [ ] 1.2 Add or narrow workflow permissions so the cleanup logic can remove the trigger label without broadening unrelated access.

## 2. Regenerate and verify workflow artifacts

- [ ] 2.1 Recompile `.github/workflows/openspec-verify-label.lock.yml` from the markdown source with `gh aw compile`.
- [ ] 2.2 Run the repository's OpenSpec validation checks and confirm the updated workflow contract matches the new `ci-aw-openspec-verification` requirements.
