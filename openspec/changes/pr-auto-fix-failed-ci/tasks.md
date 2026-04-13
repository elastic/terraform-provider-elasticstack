## 1. Author the remediation workflow and deterministic gates

- [ ] 1.1 Add the authored GH AW workflow source and compiled workflow artifacts for failed-CI pull-request remediation triggered from `workflow_run` of `Build/Lint/Test`.
- [ ] 1.2 Implement deterministic pre-activation logic for failed-run gating, same-repository PR resolution, `auto-fix` label checks, and supported failure classification for `lint` and `acceptance`.
- [ ] 1.3 Add focused tests for the deterministic gating and classification logic, including unlabeled, fork, ambiguous-PR, unsupported-failure, and supported-failure scenarios.

## 2. Implement remediation behavior and PR feedback

- [ ] 2.1 Write the agent remediation prompt and workflow contract for lint fixes, acceptance-test analysis, combined supported failures, and clear skip reasons.
- [ ] 2.2 Implement marker-based PR feedback so skipped, unsupported, and analysis-only outcomes create or update a single comment per source workflow run.
- [ ] 2.3 Configure successful remediation runs to push fixes back to the triggering PR branch and support follow-up CI triggering for agent-authored updates.

## 3. Document and verify the feature

- [ ] 3.1 Document maintainer usage of the `auto-fix` label, the supported failure profiles, and the repository configuration needed for CI reruns after agent-authored pushes.
- [ ] 3.2 Rebuild the compiled workflow artifacts and run the relevant OpenSpec, workflow, and targeted test validation for the new remediation flow.
