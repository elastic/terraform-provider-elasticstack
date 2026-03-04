# Contributing

This guide explains how to set up your environment, make changes, and submit a PR.

## Development Setup

* Fork and clone the repo.
* Setup your preferred IDE (IntelliJ, VSCode, etc.)

Requirements:
* [Terraform](https://www.terraform.io/downloads.html) >= 1.0.0
* [Go](https://golang.org/doc/install) >= 1.25
* Docker (for acceptance tests)

## Development Workflow

* Create a new branch for your changes.
* Make your changes. See [`development-workflow.md`](./development-workflow.md) for the typical change loop and common make targets, and see [Debugging](#running--debugging-the-provider) for local runs.
* Validate your changes
  * Run unit and acceptance tests (see [`testing.md`](./testing.md)).
  * Run `make lint` to check linting and formatting. For this check to succeed, all changes must have been committed.
  * All checks also run automatically on every PR.
* Submit your PR for review.
* Add a changelog entry in `CHANGELOG.md` under the `Unreleased` section. This will be included in the release notes of the next release. The changelog entry references the PR, so it has to be added after the PR has been opened.

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

## Repo structure

See [`repo-structure.md`](./repo-structure.md).

## Releasing (maintainers)

Releasing is implemented in CI pipeline.

To release a new provider version:

* Create PR which
- updates Makefile with the new provider VERSION (e.g. `VERSION ?= 0.11.13`);
- updates CHANGELOG.md with the list of changes being released.
[Example](https://github.com/elastic/terraform-provider-elasticstack/commit/be866ebc918184e843dc1dd2f6e2e1b963da386d).

* Once the PR is merged, the release CI pipeline can be started by pushing a new release tag to the `main` branch. (`git tag v0.11.13 && git push origin v0.11.13`)
