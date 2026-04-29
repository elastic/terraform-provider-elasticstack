## Verification Report: remove-7x-support — Task 1 (Documentation and CI Matrix)

### Summary Scorecard

| Dimension    | Status                                      |
|--------------|---------------------------------------------|
| Completeness | 4/4 tasks complete                          |
| Correctness  | All spec requirements and scenarios covered |
| Coherence    | Design decisions followed                   |

### Artifacts Reviewed

- `openspec/changes/remove-7x-support/proposal.md`
- `openspec/changes/remove-7x-support/design.md`
- `openspec/changes/remove-7x-support/tasks.md`
- `openspec/changes/remove-7x-support/specs/ci-build-lint-test/spec.md`
- `openspec/changes/remove-7x-support/specs/makefile-workflows/spec.md`
- `README.md`
- `.github/workflows-src/test/workflow.yml.tmpl`
- `.github/workflows/test.yml`

### Issues by Priority

#### CRITICAL
*None.*

#### WARNING
*None.*

#### SUGGESTION
*None.*

### Task-by-Task Verification

| Task | Description | Status | Evidence |
|------|-------------|--------|----------|
| 1.1 | Update `README.md` so the documented minimum supported Elastic Stack version is `8.0` or higher. | ✅ Complete | `README.md:12` reads `__The provider supports Elastic Stack versions 8.0+__`. No other support-floor statement exists in the file. |
| 1.2 | Remove the `7.17.13` entry from `.github/workflows-src/test/workflow.yml.tmpl`. | ✅ Complete | The `7.17.13` matrix entry is absent from the template. `grep -r "7\.17"` across the workflows directory returns no matches. |
| 1.3 | Regenerate `.github/workflows/test.yml` with `make workflow-generate`. | ✅ Complete | `make check-workflows` exits `0`, confirming the generated file is in sync with its source template. Running `make workflow-generate` produces no diff. The generated header clearly states it was compiled from the template. |
| 1.4 | Verify the generated workflow acceptance matrix contains no Elastic Stack 7.x entries. | ✅ Complete | Matrix versions are `8.0.1` through `8.19.9` and `9.0.8` through `9.4.0-SNAPSHOT`. No entry starts with `7.`. |

### Spec Requirement Mapping

**`ci-build-lint-test` — REQ-009–REQ-014**
> *"The configured stack versions SHALL NOT include Elastic Stack versions below `8.0.0`."*

- **Status**: Satisfied. The lowest matrix version is `8.0.1` (≥ 8.0.0).

**`ci-build-lint-test` — Scenario: Matrix excludes 7.x stack versions**
> *"WHEN the acceptance matrix is evaluated, THEN every configured stack version SHALL be `8.0.0` or higher, except snapshot labels that represent later unreleased stack versions."*

- **Status**: Satisfied. All non-snapshot entries are `8.0.1` or higher.

**`makefile-workflows` — REQ-017**
> *"When `STACK_VERSION` matches `8.0.%` or `8.1.%`, the Makefile SHALL set the Fleet agent image to `elastic/elastic-agent` on Docker Hub... For other versions, Compose SHALL use the default image source..."*

- **Status**: Satisfied in the generated workflow. The `include` block explicitly sets `fleetImage: elastic/elastic-agent` only for `8.0.1` and `8.1.3`. All other entries default to `docker.elastic.co/elastic-agent/elastic-agent`.
  - *Note*: The `Makefile` itself still contains a `7.17.%` fleet-image fallback. That is scoped to **Task 2** (pending) and does not affect the correctness of Task 1.

### Design Adherence

**Decision 3 — Keep generated files in sync through repository generators**
> *".github/workflows-src/test/workflow.yml.tmpl should be edited first and .github/workflows/test.yml regenerated with make workflow-generate."*

- **Status**: Followed. The generated file matches the template (`make check-workflows` passes).

### Final Assessment

**All checks passed. Task 1 is fully compliant with the change proposal, design, and delta specs. No actionable mismatches or missing work. Ready to proceed to Task 2.**
