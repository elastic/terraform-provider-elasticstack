## 1. Changelog-generation workflow template

- [ ] 1.1 Add `GH_AW_CI_TRIGGER_TOKEN` environment variable to the release-mode push step in `.github/workflows-src/changelog-generation/workflow.yml.tmpl`, and add the empty-commit CI trigger logic after the `git push`
- [ ] 1.2 Add `GH_AW_CI_TRIGGER_TOKEN` environment variable to the unreleased-mode push step in `.github/workflows-src/changelog-generation/workflow.yml.tmpl`, and add the empty-commit CI trigger logic (with `--force`) after the `git push`

## 2. Prep-release workflow template

- [ ] 2.1 Add `GH_AW_CI_TRIGGER_TOKEN` environment variable to the release-branch push step in `.github/workflows-src/prep-release/workflow.yml.tmpl`, and add the empty-commit CI trigger logic after the `git push`

## 3. Recompile and verify

- [ ] 3.1 Run `go run ./scripts/compile-workflow-sources/main.go` to regenerate both compiled workflow YAML files
- [ ] 3.2 Verify the compiled `.github/workflows/changelog-generation.yml` contains the CI trigger logic in both push steps
- [ ] 3.3 Verify the compiled `.github/workflows/prep-release.yml` contains the CI trigger logic in its push step
- [ ] 3.4 Run `make build` to confirm the project still compiles cleanly
