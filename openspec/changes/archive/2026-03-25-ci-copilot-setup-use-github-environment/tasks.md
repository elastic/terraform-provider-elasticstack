## 1. Workflow revert

- [x] 1.1 Remove the `jobs.copilot-setup-steps.env` credential defaults from `.github/workflows/copilot-setup-steps.yml` while preserving later unrelated workflow updates
- [x] 1.2 Confirm the workflow relies on existing repository `.env` and Makefile defaults for `make docker-fleet`, `make set-kibana-password`, and `make create-es-api-key`, while keeping `FLEET_NAME` explicitly set for `make setup-kibana-fleet`

## 2. Spec update

- [x] 2.1 Merge the delta requirements in `openspec/changes/ci-copilot-setup-use-github-environment/specs/ci-copilot-setup-steps/spec.md` into `openspec/specs/ci-copilot-setup-steps/spec.md`
- [x] 2.2 Remove the old self-contained and explicitly wired credential wording from the canonical `ci-copilot-setup-steps` spec so it documents the current reliance on repository defaults and the explicit Fleet name override

## 3. Verification

- [x] 3.1 Run `./node_modules/.bin/openspec validate ci-copilot-setup-use-github-environment --type change` or `./node_modules/.bin/openspec validate --all`
- [x] 3.2 Validate the Copilot setup workflow with repository defaults still supporting bootstrap, Kibana password setup, API key creation, and Fleet setup without workflow-local defaults beyond `FLEET_NAME`
