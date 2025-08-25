# Contributing

This guide explains how to set up your environment, make changes, and submit a PR.

## Development Setup

* Fork and clone the repo.
* Setup your preferred IDE (Intelliji, VSCode, etc.)

Requirements:
* [Terraform](https://www.terraform.io/downloads.html) >= 1.0.0
* [Go](https://golang.org/doc/install) >= 1.25
* Docker (for acceptance tests)

## Development Workflow

* Create a new branch for your changes.
* Make your changes. See [Useful Commands](#useful-commands) and [Debugging](#debugging).
  * GitHub Copilot can be used to help with these changes (See [Using Copilot](#using-copilot))
* Validate your changes
  * Run unit and acceptance tests (See [Running Acceptance Tests](#running-acceptance-tests)).
  * Run `make fmt`, `make lint`.
  * All checks also run automatically on every PR.
* Add a changelog entry in `CHANGELOG.md` under the `Unreleased` section. This will be included in the release notes of the next release.
* Submit your PR for review.

When creating new resources:
* Use the [Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework/getting-started/code-walkthrough) for new resources.
  * Use an existing resource (e.g. `internal/elasticsearch/security/system_user`) as a template.
  * Some resources use the deprecated Terraform SDK, so only resources using the new Terraform Framework should be used as reference.
* Use the generated API clients to interact with the Kibana APIs. (See [Working with Generated API Clients](#working-with-generated-api-clients)
* Add documentation and examples for the resource. Update the generated docs with `make docs-generate`.
* Write unit and acceptance tests.


## Using Copilot

GitHub Copilot can speed up development, but you’re responsible for correctness:
* Create an issue describing the desired change and acceptance criteria.
* Assign the issue to Copilot and iterate with prompts.
* Review outputs carefully; add tests and adjust as needed.
* Example issue: https://github.com/elastic/terraform-provider-elasticstack/issues/1219

### Useful Commands

* `make build`: Build the provider.
* `make install`: Install the provider.
* `make fmt`: Format the code.
* `make lint`: Lint the code.
* `make test`: Run unit tests.
* `make docs-generate`: Generate documentation.

#### Running Acceptance Tests

Acceptance tests spin up Elasticsearch, Kibana, and Fleet with Docker and run tests in a Go container.

Quick start (default stack version from Makefile):
```bash
make docker-testacc
```

Run a single test with terraform debug enabled:
```bash
env TF_LOG=DEBUG make docker-testacc TESTARGS='-run ^TestAccResourceDataStreamLifecycle$$'
```

A way to forward debug logs to a file:
```bash
env TF_ACC_LOG_PATH=/tmp/tf.log TF_ACC_LOG=DEBUG TF_LOG=DEBUG make docker-testacc
```

### Working with Generated API Clients

If your work involves the Kibana API, the API client can be generated directly from the Kibana OpenAPI specs:
- For Kibana APIs, use the generated client in `generated/kbapi`.
- To add new endpoints, see [generated/kbapi/README.md](generated/kbapi/README.md).
- Regenerate clients with:
  ```sh
  make transform generate
  ```

The codebase includes a number of deprecated clients which should not be used anymore:
- `libs/go-kibana-rest`: Fork of an external library, which is not maintained anymore.
- `generated/alerting`
- `generated/connectors`
- `generated/slo`
- `internal/clients/*`: Manually written clients. These should only be used/extended if it is not possible to use the generated clients.

### Updating Documentation

Docs are generated from templates in `templates/` and examples in `examples/`.
* Update or add templates and examples.
* Run `make docs-generate` to produce files under `docs/`.
* Commit the generated files. `make lint` will fail if docs are stale.

### Debugging

Run the provider in debug mode and reattach the provider in Terraform:
* Launch `main.go` with the `-debug` flag from your IDE.
* After launching, the provider will print an env var. Copy the printed `TF_REATTACH_PROVIDERS='{…}'` value.
* Export it in your shell where you run Terraform: `export TF_REATTACH_PROVIDERS='{…}'`.
* Terraform will now talk to your debug instance, and you can set breakpoints.

## Project Structure

A quick overview over what's in each folder:

* `docs/` - Documentation files
  * `data-sources/` - Documentation for Terraform data sources
  * `guides/` - User guides and tutorials
  * `resources/` - Documentation for Terraform resources
* `examples/` - Example Terraform configurations
  * `cloud/` - Examples using the cloud to launch testing stacks
  * `data-sources/` - Data source usage examples
  * `resources/` - Resource usage examples
  * `provider/` - Provider configuration examples
* `generated/` - Auto-generated clients from the `generate-clients` make target
  * `kbapi/` - Kibana API client
  * `alerting/` - (Deprecated) Kibana alerting API client
  * `connectors/` - (Deprecated) Kibana connectors API client
  * `slo/` - (Deprecated) SLO (Service Level Objective) API client
* `internal/` - Internal Go packages
  * `acctest/` - Acceptance test utilities
  * `clients/` - API client implementations
  * `elasticsearch/` - Elasticsearch-specific logic
  * `fleet/` - Fleet management functionality
  * `kibana/` - Kibana-specific logic
  * `models/` - Data models and structures
  * `schema/` - Connection schema definitions for plugin framework
  * `utils/` - Utility functions
  * `versionutils/` - Version handling utilities
* `libs/` - External libraries
  * `go-kibana-rest/` - (Deprecated) Kibana REST API client library
* `provider/` - Core Terraform provider implementation
* `scripts/` - Utility scripts for development and CI
* `templates/` - Template files for documentation generation
  * `data-sources/` - Data source documentation templates
  * `resources/` - Resource documentation templates
  * `guides/` - Guide documentation templates
* `xpprovider/` - Additional provider functionality needed for Crossplane

## Releasing (maintainers)

Releasing is implemented in CI pipeline.

To release a new provider version:

* Create PR which
- updates Makefile with the new provider VERSION (e.g. `VERSION ?= 0.11.13`);
- updates CHANGELOG.md with the list of changes being released.
[Example](https://github.com/elastic/terraform-provider-elasticstack/commit/be866ebc918184e843dc1dd2f6e2e1b963da386d).

* Once the PR is merged, the release CI pipeline can be started by pushing a new release tag to the `main` branch. (`git tag v0.11.13 && git push origin v0.11.13`)
