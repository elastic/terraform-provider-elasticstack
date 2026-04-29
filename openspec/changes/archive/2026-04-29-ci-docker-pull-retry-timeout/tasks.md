## 1. Workflow Template Changes

- [ ] 1.1 Add a "Pre-pull fleet image" step to `.github/workflows-src/test/workflow.yml.tmpl`
  - Conditional: `if: matrix.fleetImage`
  - `timeout-minutes: 5`
  - Retry loop with `timeout 90 docker pull` and 3 attempts
- [ ] 1.2 Add `timeout-minutes: 10` to the existing "Start stack with docker compose" step

## 2. Regenerate and Verify

- [ ] 2.1 Run `make workflow-generate` to regenerate `.github/workflows/test.yml`
- [ ] 2.2 Verify the generated workflow matches the template changes
- [ ] 2.3 Ensure `make check-workflows` passes (or equivalent lint)
