## 1. Workflow revert

- [ ] 1.1 Remove the `jobs.copilot-setup-steps.env` credential defaults from `.github/workflows/copilot-setup-steps.yml` and restore the pre-`6501f04` step environment wiring while preserving later unrelated workflow updates
- [ ] 1.2 Confirm the workflow consumes the GitHub-managed credential variables needed by `make docker-fleet`, `make set-kibana-password`, `make create-es-api-key`, and `make setup-kibana-fleet`

## 2. Spec update

- [ ] 2.1 Merge the delta requirements in `openspec/changes/ci-copilot-setup-use-github-environment/specs/ci-copilot-setup-steps/spec.md` into `openspec/specs/ci-copilot-setup-steps/spec.md`
- [ ] 2.2 Remove the existing self-contained default-credential wording from the canonical `ci-copilot-setup-steps` spec so it documents GitHub repository environment settings as the credential source of truth

## 3. Verification

- [ ] 3.1 Run `./node_modules/.bin/openspec validate ci-copilot-setup-use-github-environment --type change` or `./node_modules/.bin/openspec validate --all`
- [ ] 3.2 Validate the Copilot setup workflow with repository environment settings in place so bootstrap, Kibana password setup, API key creation, and Fleet setup still succeed without workflow-local defaults
