# Contributing

This guide explains how to set up your environment, make changes, and submit a PR.

## Development Setup

* Fork and clone the repo.
* Setup your preferred IDE (IntelliJ, VSCode, etc.)

Requirements:
* [Terraform](https://www.terraform.io/downloads.html) >= 1.0.0
* [Go](https://golang.org/doc/install) >= 1.25
* [Node.js](https://nodejs.org/) 24.x (for OpenSpec; installed via `make setup` / `npm ci`)
* Docker (for acceptance tests)

OpenSpec requirements specs live under `openspec/specs/`; see [`openspec-requirements.md`](./openspec-requirements.md). PR review automation (e.g. the **`verify-openspec`** label workflow) is documented in [`code-review.md`](./code-review.md).

## Development Workflow

* Create a new branch for your changes.
* Make your changes. See [`development-workflow.md`](./development-workflow.md) for the typical change loop and common make targets, and see [Debugging](#running--debugging-the-provider) for local runs.
* Validate your changes
  * Run unit and acceptance tests (see [`testing.md`](./testing.md)).
  * Run `make lint` to check linting and formatting. For this check to succeed, all changes must have been committed.
  * All checks also run automatically on every PR.
* Submit your PR for review.

The `## [Unreleased]` section of `CHANGELOG.md` is maintained automatically. The `changelog-generation` GitHub Actions workflow runs on a schedule and regenerates the `## [Unreleased]` section from merged PR history. It opens a PR from the `generated-changelog` branch that is auto-merged once checks pass. You do not need to manually add changelog entries after your PR is merged.

When creating new resources:
* Use the [Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework/getting-started/code-walkthrough) for new resources.
  * Use an existing resource (e.g. `internal/elasticsearch/security/systemuser`) as a template.
  * Some resources use the deprecated Terraform SDK, so only resources using the new Terraform Framework should be used as reference.
* Use the generated API clients to interact with the Kibana APIs (see [`generated-clients.md`](./generated-clients.md)).
* Add documentation templates and examples for the resource (see [`documentation.md`](./documentation.md)).
* Write unit and acceptance tests (see [`testing.md`](./testing.md)).

### Running & Debugging the Provider

You can run the currently checked-out code for local testing and use it with Terraform.

Also see [Terraform docs on debugging](https://developer.hashicorp.com/terraform/plugin/debugging#starting-a-provider-in-debug-mode).

Run the provider in debug mode and reattach the provider in Terraform:
* Launch `main.go` with the `-debug` flag from your IDE.
  * Or launch it with `go run main.go -debug` from the command line.
* After launching, the provider will print an env var. Copy the printed `TF_REATTACH_PROVIDERS='{…}'` value.
* Export it in your shell where you run Terraform: `export TF_REATTACH_PROVIDERS='{…}'`.
* Terraform will now talk to your debug instance, and you can set breakpoints.

### Useful commands

See [`development-workflow.md`](./development-workflow.md) for common make targets and the typical change loop.

### Acceptance tests

See [`testing.md`](./testing.md) for how to run acceptance tests (including required environment variables).

### Generated API clients

See [`generated-clients.md`](./generated-clients.md).

### Updating documentation

See [`documentation.md`](./documentation.md).

When you run `make docs-generate`, the command uses the Terraform CLI version pinned in the repository root `.terraform-version` file via `tfplugindocs`. If docs generation behavior changes because Terraform needs to be updated, update `.terraform-version` rather than relying on a locally installed Terraform version.

## Repo structure

See [`repo-structure.md`](./repo-structure.md).

## Releasing (maintainers)

Releasing is implemented in CI pipeline. Release preparation is now automated — do not manually edit `VERSION` or `CHANGELOG.md` release sections.

To release a new provider version:

1. **Dispatch the release preparation workflow** using the `prep-release` Makefile target:

   ```
   make prep-release           # defaults to patch bump (e.g. 0.14.3 → 0.14.4)
   make prep-release BUMP=minor # minor bump (e.g. 0.14.3 → 0.15.0)
   make prep-release BUMP=major # major bump (e.g. 0.14.3 → 1.0.0)
   ```

   This dispatches the `prep-release.yml` GitHub Actions workflow, which:
   - Computes the target version by finding the latest semver release tag (`v*.*.*`) on `main` and applying the requested bump.
   - Creates (or reuses) a `prep-release-x.y.z` branch and opens a pull request with the `VERSION` variable in `Makefile` updated to the target version.

2. **Await the changelog update**. The `changelog-generation` workflow automatically detects the new `prep-release-*` PR and regenerates the concrete `## [x.y.z] - YYYY-MM-DD` section in `CHANGELOG.md`, pushing the result to the `prep-release-*` branch.

3. **Review and merge the release PR**. Once the changelog section is populated and all checks pass, merge the `prep-release-x.y.z` PR.

4. **Tag and release**. After the PR is merged, start the release by pushing the version tag to `main`:

   ```
   git tag v0.14.4 && git push origin v0.14.4
   ```

   The release CI pipeline will then build and publish the provider.
