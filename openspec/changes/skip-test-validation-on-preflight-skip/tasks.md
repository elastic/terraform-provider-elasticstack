## 1. Update workflow gating

- [ ] 1.1 Update `.github/workflows-src/test/workflow.yml.tmpl` so `test-validation` runs only when `preflight.outputs.should_run == 'true'` while preserving `always()` behavior for upstream `changes` and `test` results.
- [ ] 1.2 Confirm `auto-approve` still behaves correctly for `ready_for_review` runs when `test-validation` is skipped.

## 2. Align validation helper behavior

- [ ] 2.1 Update `.github/workflows-src/lib/validate-test-result.js` so it only models reachable post-preflight validation states.
- [ ] 2.2 Update `.github/workflows-src/lib/validate-test-result.test.mjs` to remove or rewrite cases that treat preflight-disabled runs as a passed validation result.

## 3. Regenerate and verify

- [ ] 3.1 Regenerate `.github/workflows/test.yml` from the authored workflow sources.
- [ ] 3.2 Run the workflow-focused verification commands, including `make workflow-test`, and address any failures caused by the change.
