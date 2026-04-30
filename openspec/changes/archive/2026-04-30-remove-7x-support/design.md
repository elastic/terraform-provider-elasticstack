## Context

The provider still advertises Elastic Stack `7.x+` support in `README.md` and runs acceptance tests against `7.17.13`. Several implementation paths and OpenSpec requirements preserve behavior for Elasticsearch versions below `8.0.0`, including transform setting gates, transform timeout handling, ILM `total_shards_per_node` gating, and Docker Fleet image fallback for `7.17.%`.

Elastic Stack 7.x is now outside the intended support target. The change should remove intentional 7.x support without adding a blanket runtime rejection that would deliberately break incidental compatibility.

## Goals / Non-Goals

**Goals:**

- Set the documented minimum supported Elastic Stack version to `8.0.0`.
- Remove `7.17.13` from the GitHub Actions acceptance matrix.
- Remove 7.x-specific compatibility branches where the supported 8.0+ range makes them redundant.
- Keep 8.x and 9.x feature gates intact where APIs or fields were introduced after `8.0.0`.
- Regenerate generated artifacts that are expected to track source changes.

**Non-Goals:**

- Do not introduce a provider-wide runtime failure for all 7.x clusters.
- Do not remove schema attributes solely because they existed before 8.0 if they remain valid in supported stack versions.
- Do not rewrite unrelated historical changelog entries or archived OpenSpec changes.

## Decisions

### 1. Treat 8.0 as the support floor, not a global connection gate

The provider documentation and tests will define the support floor. Resource logic will not add a broad `serverVersion < 8.0.0` diagnostic.

Alternative considered: enforce `>= 8.0.0` centrally in provider configuration. That would be clearer but intentionally breaks users who may still have working 7.x workflows, which is outside the requested scope.

### 2. Remove only pre-8.0 compatibility gates

Version checks with minimums below `8.0.0` should be removed when they exist only to support 7.x behavior. Examples include transform feature minimum `7.2.0`, transform timeout minimum `7.17.0`, transform field gates below `8.0.0`, and ILM `total_shards_per_node` minimum `7.16.0`.

Version gates at `8.x` or `9.x` should remain because they still describe feature boundaries inside the supported version range.

### 3. Keep generated files in sync through repository generators

The workflow source template is authoritative for the test workflow, so `.github/workflows-src/test/workflow.yml.tmpl` should be edited first and `.github/workflows/test.yml` regenerated with `make workflow-generate`. Terraform provider docs should be regenerated after schema description changes with `make docs-generate`.

### 4. Leave historical records alone

Historical changelog entries and archived OpenSpec changes may continue to mention 7.x because they describe past behavior. Current specs, docs, workflow sources, generated workflow output, and live implementation should reflect the new support floor.

## Risks / Trade-offs

- Removing 7.x matrix coverage may hide incidental regressions for users who still run 7.x. Mitigation: the change explicitly documents 8.0+ as the support floor rather than promising continued 7.x behavior.
- Removing compatibility gates can change behavior when users connect to 7.x despite the unsupported status. Mitigation: avoid a global hard failure and only remove branches whose purpose is pre-8.0 support.
- Generated workflow and docs can drift if only source files are edited. Mitigation: include generator and validation steps in the tasks.
- Some 7.x strings are historical or test fixture values, not support promises. Mitigation: tasks should distinguish current support surface from historical records and unrelated literals.
