## 1. Add deterministic issue-slot gating

- [x] 1.1 Add `.github/workflows-src/schema-coverage-rotation/` and author the workflow there using the same templated pattern as `.github/workflows-src/openspec-verify-label/`.
- [x] 1.2 Extract the issue-slot calculation into a helper module under `.github/workflows-src/lib/`, add a thin inline wrapper for `actions/github-script`, and cover the helper with unit tests.
- [x] 1.3 Update the templated schema-coverage rotation workflow source to run the helper in pre-activation and publish the open-count, slot-count, and gate-reason outputs.

## 2. Gate the agent job and simplify the prompt

- [x] 2.1 Update the workflow job conditions so the schema-coverage agent path is skipped entirely when `issue_slots_available` is `0`.
- [x] 2.2 Remove issue-counting instructions from the agent prompt and replace them with references to the precomputed slot outputs.

## 3. Rebuild and verify workflow artifacts

- [x] 3.1 Recompile `.github/workflows/schema-coverage-rotation.lock.yml` from the markdown workflow source.
- [x] 3.2 Run the relevant OpenSpec and workflow validation checks for the new pre-activation gating path and skipped-agent behavior.
