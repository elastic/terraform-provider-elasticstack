## 1. Add `make fmt` to the dead-code removal agent task

- [ ] 1.1 In `.github/workflows/ci-deadcode-removal-rotation.md`, locate the agent task section's step 5 ("Open a cleanup PR"). Insert a new step **between** step 4 (verification) and step 5 (PR creation):

  - Run `make fmt`.
  - If `make fmt` exits non-zero:
    - Record the attempt as `fmt_failed`:
      ```
      go run ./scripts/ci-deadcode-removal-rotation record \
        --memory /tmp/gh-aw/repo-memory/ci-deadcode-removal-rotation/memory/ci-deadcode-removal-rotation/memory.json \
        --symbol "${{ needs.pre_activation.outputs.symbol }}" \
        --package "${{ needs.pre_activation.outputs.package }}" \
        --reason fmt_failed
      ```
    - Call `noop` with a concise reason.
  - If `make fmt` exits zero, continue to step 5 (open the PR).

- [ ] 1.2 Renumber the subsequent task steps in the markdown to keep the list sequential after inserting the new step.

- [ ] 1.3 Rebuild the compiled workflow lock artifact by running `make workflow-generate` (or the equivalent `workflows generate` command for this repo) and commit the updated `.github/workflows/ci-deadcode-removal-rotation.lock.yml`.
