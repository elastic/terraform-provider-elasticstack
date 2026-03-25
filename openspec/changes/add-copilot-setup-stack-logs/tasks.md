## 1. Workflow update

- [x] 1.1 Add a failure-path step to `.github/workflows/copilot-setup-steps.yml` that runs `docker compose logs --no-color`
- [x] 1.2 Guard the log collection step so it runs only when the job has failed, without changing the successful setup path

## 2. Spec update

- [x] 2.1 Merge the delta requirement in `openspec/changes/add-copilot-setup-stack-logs/specs/ci-copilot-setup-steps/spec.md` into `openspec/specs/ci-copilot-setup-steps/spec.md`

## 3. Verification

- [x] 3.1 Run `./node_modules/.bin/openspec validate --all` or `make check-openspec`
- [ ] 3.2 Optionally trigger the Copilot setup workflow or equivalent validation path to confirm failed runs emit stack logs and successful runs do not
