## 1. Helper contract

- [x] 1.1 Extend `scripts/kibana-spec-impact/` so the repository-local Go helper can perform the deterministic pre-activation flow: initialize memory when needed, compute the report, derive gate outputs, and write the report file used by the workflow.
- [x] 1.2 Add or update focused Go tests covering the pre-activation gate/report contract, including the emitted gate fields and report-file behavior.

## 2. Workflow orchestration

- [x] 2.1 Update `.github/workflows-src/kibana-spec-impact/workflow.md.tmpl` so `on.steps` checks out the repository before Go setup, declares the repo-memory `branch-name` explicitly, and initializes repo memory with a dedicated checkout/init step against that same branch before computing Kibana spec impact.
- [x] 2.2 Replace the current pre-activation report handoff with an explicit artifact upload in pre-activation and artifact download into `/tmp/gh-aw/agent` in the agent job.
- [x] 2.3 Rewrite the agent instructions so report and issued-file references point at `/tmp/gh-aw/agent`, while repo-memory persistence continues to use the configured `/tmp/gh-aw/repo-memory/...` path.

## 3. Generated artifacts and verification

- [x] 3.1 Remove or retire obsolete inline workflow helper code if it is no longer referenced after the Go-based pre-activation flow lands.
- [x] 3.2 Regenerate `.github/workflows/kibana-spec-impact.md` and `.github/workflows/kibana-spec-impact.lock.yml` from the updated source template.
- [x] 3.3 Run `make workflow-test` and `make check-workflows`, then confirm the generated workflow contains the new checkout, repo-memory initialization, artifact handoff, and `/tmp/gh-aw/agent` prompt paths.
