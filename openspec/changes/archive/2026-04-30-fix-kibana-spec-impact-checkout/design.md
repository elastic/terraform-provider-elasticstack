## Context

The `kibana-spec-impact` workflow is compiled from `.github/workflows-src/kibana-spec-impact/workflow.md.tmpl` into `.github/workflows/kibana-spec-impact.lock.yml` by `scripts/compile-workflow-sources`.

The pre-activation `steps:` block in the template performs these actions:
1. Check out the main repository (`fetch-depth: 0`)
2. Set up Go
3. **Check out the `memory/kibana-spec-impact` branch** to an absolute path: `path: /tmp/gh-aw/repo-memory/kibana-spec-impact`
4. Run the `pre-activation` Go command with `--memory` pointing at that absolute path
5. Upload the report as an artifact

The `actions/checkout@v4` action validates that `path` is relative and under `$GITHUB_WORKSPACE`. An absolute `/tmp` path causes the step to fail with:
```
Error: Repository path is not under the workspace
```

This failure blocks the entire workflow because `activation` and `agent` jobs depend on `pre_activation` outputs (`run_agent`, `issue_cap`, etc.).

The `agent` job itself already handles `/tmp` paths fine — it uses `clone_repo_memory_branch.sh` (a `run:` step) to clone the memory branch into `/tmp/gh-aw/repo-memory/kibana-spec-impact`. The `actions/checkout` restriction only applies to the pre-activation job.

## Goals / Non-Goals

**Goals:**
- Fix the `pre_activation` job so the workflow completes successfully
- Keep the existing deterministic `go run ... pre-activation` logic unchanged
- Keep the artifact upload/download contract between pre-activation and the agent intact
- Regenerate the compiled lockfile from the updated template

**Non-Goals:**
- Changing the `agent` job (it already works fine)
- Changing the memory branch name or structure
- Changing the `kibana-spec-impact` Go tool
- Introducing new workflow behaviour (e.g. merging pre-activation into the agent job)

## Decisions

### Decision: Use workspace-relative checkout path in pre-activation

**Chosen:** Change `path: /tmp/gh-aw/repo-memory/kibana-spec-impact` to a relative path under the workspace (e.g. `gh-aw-repo-memory/kibana-spec-impact`).

**Rationale:**
- Minimal change — only two lines in the template (checkout `path` and `--memory` flag value)
- No new dependencies or scripts needed
- The compiled workflow keeps the same structure; only the path changes
- Pre-activation runs on a different runner (`ubuntu-slim`) than the agent (`ubuntu-latest`), so there is no conflict with the agent's own `/tmp` memory clone even if the paths differ

**Alternatives considered:**

| Approach | Why rejected |
|----------|-------------|
| Replace `actions/checkout` with a manual `git clone` run step | Works, but adds ~10 lines of shell and re-authentication boilerplate. Option C (relative path) is simpler. |
| Fetch the memory file via GitHub API (`repos.getContent`) | API call overhead, must handle nested path encoding and branch-not-exists logic. Over-engineered for this problem. |
| Move deterministic check into the agent job | Removes `pre_activation` entirely, which breaks the gh-aw framework (pre-activation also emits `activated`, `setup-trace-id`, team-membership gate). |

### Decision: Keep `--memory` path in sync with checkout `path`

**Chosen:** Update the `--memory` flag in the `go run pre-activation` command to match the new checkout location exactly.

**Rationale:**
- The Go command reads the memory file to resolve the baseline SHA and suppress duplicates
- If the path is wrong, `ensurePreActivationMemory` will bootstrap from seed instead of loading the live branch, silently resetting state

## Risks / Trade-offs

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Compiled lockfile is out of sync with template edit | Low | Include explicit `go run ./scripts/compile-workflow-sources` step and CI check |
| Agent job and pre-activation job memory paths diverge silently | Very low | Paths are each self-contained; the agent downloads and re-closes its own memory. The report artifact is the only contract between jobs. |
| `continue-on-error: true` on the checkout step is lost during copy-paste | Low | Review the template diff carefully; the new checkout step keeps the same `continue-on-error` attribute |
