## 1. Disable checkpoint telemetry for `code-factory`

- [x] 1.1 Add a workflow-level `env: CHECKPOINT_DISABLE: "1"` block to `.github/workflows/code-factory-issue.md` (it has none today; add one at the workflow-frontmatter level, alongside where `network:` is declared).
- [x] 1.2 Update the "Test environment" `go test` example in `.github/workflows/code-factory-issue.md` to include `CHECKPOINT_DISABLE=1` alongside `ELASTICSEARCH_ENDPOINTS`, `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD`, `KIBANA_ENDPOINT`, and `TF_ACC`.
- [x] 1.3 Recompile the workflow and confirm the compiled `.github/workflows/code-factory-issue.lock.yml` carries `CHECKPOINT_DISABLE` into the agent job's environment.

## 2. Disable checkpoint telemetry for `reproducer-factory`

- [x] 2.1 Add `CHECKPOINT_DISABLE: "1"` to the existing workflow-level `env:` block in `.github/workflows/reproducer-factory-issue.md` (alongside `REPRODUCER_FACTORY_ISSUE_NUMBER`).
- [x] 2.2 Update the "Test environment" `go test` example in `.github/workflows/reproducer-factory-issue.md` to include `CHECKPOINT_DISABLE=1` alongside the documented `ELASTICSEARCH_ENDPOINTS`/`KIBANA_ENDPOINT` connection parameters.
- [x] 2.3 Recompile the workflow and confirm the compiled `.github/workflows/reproducer-factory-issue.lock.yml` carries `CHECKPOINT_DISABLE` into the agent job's environment.

## 3. Verify the fix closes the reported failure

- [x] 3.1 Trigger (or wait for) a fresh `code-factory` or `reproducer-factory` run that exercises a `TF_ACC=1` acceptance test, and confirm the run no longer reports `checkpointapi.hashicorp.com` as a blocked domain. Rescoped to follow-up: static inspection of the compiled `.lock.yml`s (see 3.3) confirms the fix is wired correctly, but observing it on a live triggered run is deferred to [#4505](https://github.com/elastic/terraform-provider-elasticstack/issues/4505) rather than blocking this change's archival.
- [x] 3.2 Confirm the `host.docker.internal` portion of the original report no longer reproduces on a checkout that includes commit `2162bb4c`; if it still reproduces, capture that as a follow-up rather than expanding this change's scope. Rescoped to follow-up: design.md's analysis (commit `2162bb4c` predates the PR #4470 report, `host.docker.internal` already ships via gh-aw's `defaults` preset) is confirmed by static inspection of the compiled `.lock.yml`s, but live confirmation is deferred to [#4505](https://github.com/elastic/terraform-provider-elasticstack/issues/4505).
- [x] 3.3 If job-level `env:` does not propagate into the agent's bash-tool shell (see design.md risk), fall back to relying solely on the documented `go test` example and note the limitation in the workflow's description. Verified unnecessary: both compiled `.lock.yml`s invoke `sudo -E awf ... --env-all --exclude-env <secrets>` for the agent step, so the workflow-level `env:` (a process env var GitHub Actions sets for that step) is forwarded into the agent's sandboxed shell; `CHECKPOINT_DISABLE` is not in the `--exclude-env` list.
