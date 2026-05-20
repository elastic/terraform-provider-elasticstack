## 1. Add `make fmt` to the dead-code removal agent task

- [ ] 1.1 In `.github/workflows/ci-deadcode-removal-rotation.md`, locate the agent task section's step 5 ("Open a cleanup PR"). Insert a new step **between** step 4 (verification) and step 5 (PR creation):

  - Run `make fmt`.

- [ ] 1.2 Renumber the subsequent task steps in the markdown to keep the list sequential after inserting the new step.

- [ ] 1.3 Rebuild the compiled workflow lock artifact by running `make workflow-generate` (or the equivalent `workflows generate` command for this repo) and commit the updated `.github/workflows/ci-deadcode-removal-rotation.lock.yml`.
