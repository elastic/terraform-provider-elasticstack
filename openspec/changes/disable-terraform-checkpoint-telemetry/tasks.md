## 1. Disable checkpoint telemetry for `code-factory`

- [ ] 1.1 Add a workflow-level `env: CHECKPOINT_DISABLE: "1"` block to `.github/workflows/code-factory-issue.md` (it has none today; add one at the workflow-frontmatter level, alongside where `network:` is declared).
- [ ] 1.2 Update the "Test environment" `go test` example in `.github/workflows/code-factory-issue.md` to include `CHECKPOINT_DISABLE=1` alongside `ELASTICSEARCH_ENDPOINTS`, `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD`, `KIBANA_ENDPOINT`, and `TF_ACC`.
- [ ] 1.3 Recompile the workflow and confirm the compiled `.github/workflows/code-factory-issue.lock.yml` carries `CHECKPOINT_DISABLE` into the agent job's environment.

## 2. Disable checkpoint telemetry for `reproducer-factory`

- [ ] 2.1 Add `CHECKPOINT_DISABLE: "1"` to the existing workflow-level `env:` block in `.github/workflows/reproducer-factory-issue.md` (alongside `REPRODUCER_FACTORY_ISSUE_NUMBER`).
- [ ] 2.2 Update the "Test environment" `go test` example in `.github/workflows/reproducer-factory-issue.md` to include `CHECKPOINT_DISABLE=1` alongside the documented `ELASTICSEARCH_ENDPOINTS`/`KIBANA_ENDPOINT` connection parameters.
- [ ] 2.3 Recompile the workflow and confirm the compiled `.github/workflows/reproducer-factory-issue.lock.yml` carries `CHECKPOINT_DISABLE` into the agent job's environment.

## 3. Verify the fix closes the reported failure

- [ ] 3.1 Trigger (or wait for) a fresh `code-factory` or `reproducer-factory` run that exercises a `TF_ACC=1` acceptance test, and confirm the run no longer reports `checkpointapi.hashicorp.com` as a blocked domain.
- [ ] 3.2 Confirm the `host.docker.internal` portion of the original report no longer reproduces on a checkout that includes commit `2162bb4c`; if it still reproduces, capture that as a follow-up rather than expanding this change's scope.
- [ ] 3.3 If job-level `env:` does not propagate into the agent's bash-tool shell (see design.md risk), fall back to relying solely on the documented `go test` example and note the limitation in the workflow's description.
