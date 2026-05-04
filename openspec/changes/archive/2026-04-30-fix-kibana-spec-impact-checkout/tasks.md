## 1. Fix pre-activation checkout path in workflow template

- [x] 1.1 Edit `.github/workflows-src/kibana-spec-impact/workflow.md.tmpl`
  - Change the `Checkout repo-memory branch` step `path` from `/tmp/gh-aw/repo-memory/kibana-spec-impact` to a workspace-relative path (`gh-aw-repo-memory/kibana-spec-impact`)
  - Update the `--memory` flag in the `Compute kibana spec impact` step to match the new relative path
  - Preserve `continue-on-error: true` on the checkout step
- [x] 1.2 Verify the template compiles cleanly by running the workflow compiler

## 2. Regenerate compiled lockfile

- [x] 2.1 Run `go run ./scripts/compile-workflow-sources` to regenerate `.github/workflows/kibana-spec-impact.lock.yml`
- [x] 2.2 Inspect the diff to confirm only the pre-activation checkout path and `--memory` flag changed; no other jobs or steps were altered

## 3. Verify agent steps remain untouched

- [x] 3.1 Confirm the `agent` job's `Download kibana spec impact report` and `Clone repo-memory branch` steps still reference `/tmp/gh-aw/...` paths (these are correct and must not change)
- [x] 3.2 Confirm the repo-memory tool config (`branch-name: memory/kibana-spec-impact`) matches the checkout `ref` in pre-activation
