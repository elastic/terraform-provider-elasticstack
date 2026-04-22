## Why

`make docs-generate` currently invokes `tfplugindocs` without pinning a Terraform CLI version. `terraform-plugin-docs` falls back to whatever `terraform` binary is on a developer's `PATH`, and if none is present it may download the latest Terraform release. That makes documentation generation depend on local machine state instead of repository-owned configuration.

This repo already pins other tooling versions in repository-controlled files (for example Go via `go.mod`, Node via contributor docs/setup, and `golangci-lint` in `Makefile`). Docs generation should follow the same pattern so contributors and CI regenerate the same docs from the same provider schema extraction behavior.

## What Changes

- **Introduce** a root `.terraform-version` file as the repository-owned source of truth for the Terraform CLI version used by docs generation and relevant CI validation jobs.
- **Update** `make docs-generate` to read `.terraform-version` and pass that version to `tfplugindocs`.
- **Document** the pinned Terraform version policy and where it is configured.
- **Align** CI Terraform setup for docs/lint validation with `.terraform-version` so local and CI docs generation use the same CLI version.
- **Refresh** the `tfplugindocs` tool dependency to a release that carries HashiCorp's updated release-signing key material, because the existing dependency chain fails while downloading the pinned Terraform CLI with `openpgp: key expired`.
- **Adopt** the current latest stable Terraform release as the initial pinned value, with future updates managed through Renovate's built-in `.terraform-version` support.

## Capabilities

### New Capabilities

- (none)

### Modified Capabilities

- `makefile-workflows`: Documentation generation becomes deterministic with respect to Terraform CLI selection. The repository, via `.terraform-version`, not the developer workstation, decides which Terraform version `tfplugindocs` uses.
- `ci-build-lint-test`: Lint/docs validation jobs that exercise docs generation use the same `.terraform-version`-pinned Terraform CLI version.

## Impact

- `.terraform-version` — add the repository-owned Terraform CLI version file, initially pinned to the current latest stable release.
- `Makefile` — update `docs-generate` to read `.terraform-version` and use it in `tfplugindocs` invocation.
- `go.mod` / `go.sum` — refresh the `terraform-plugin-docs` tool dependency (and its transitive toolchain dependencies) so docs generation can still download and verify the pinned Terraform CLI release.
- `dev-docs/high-level/documentation.md` and possibly `dev-docs/high-level/contributing.md` — explain the pinned Terraform version policy and contributor expectations.
- `.github/workflows/test.yml` and/or generated workflow sources — align Terraform setup in lint/docs-related jobs with `.terraform-version`.
- `renovate.json` — verify the repository's Renovate configuration allows built-in `.terraform-version` updates as the ongoing maintenance path.
- OpenSpec specs — update `makefile-workflows` (and CI specs if needed) to reflect deterministic Terraform version selection for docs generation and CI.
