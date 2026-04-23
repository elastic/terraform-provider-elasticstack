## 1. Author the duplicate-code detector workflow contract

- [ ] 1.1 Add the authored duplicate-code detector workflow source under `.github/workflows-src/`, with the source derived from and traceable to `https://github.com/github/gh-aw/blob/main/.github/workflows/duplicate-code-detector.md`, and register or retain its generated output in the workflow-source manifest.
- [ ] 1.2 Define deterministic pre-activation issue-slot gating for the `duplicate-code` label and workflow-configured issue cap before agent analysis starts.
- [ ] 1.3 Write the workflow prompt contract for duplicate-detection scope, significance thresholds, and one-issue-per-pattern reporting.

## 2. Add deterministic helper logic and generated artifacts

- [ ] 2.1 Implement or refine shared helper logic and inline scripts needed to compute issue-slot availability for the workflow.
- [ ] 2.2 Generate and commit `.github/workflows/duplicate-code-detector.md`, `.github/workflows/duplicate-code-detector.lock.yml`, and any related lock metadata updates.
- [ ] 2.3 Add or update focused workflow-source tests covering the issue-slot gate and generated workflow expectations.

## 3. Validate and document the workflow

- [ ] 3.1 Run the relevant workflow generation, workflow-source tests, and OpenSpec validation for the duplicate-code detector change.
- [ ] 3.2 Ensure maintainer-facing workflow behavior is documented through the upstream source reference, issue labels, title prefix, cap, and actionable issue-content contract captured by the change.
