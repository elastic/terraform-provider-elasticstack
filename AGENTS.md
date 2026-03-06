# Agent guide (start here)

This repo is the Terraform provider for Elastic Stack, written in Go.

## Before making changes

- Follow the project’s coding conventions in [`coding-standards.md`](./dev-docs/high-level/coding-standards.md).
- For contributor workflow, setup, and release notes, see [`contributing.md`](./dev-docs/high-level/contributing.md).

## High-level dev docs

- Repo orientation and where code lives: [`dev-docs/high-level/repo-structure.md`](./dev-docs/high-level/repo-structure.md)
- Common workflows and “what to do when”: [`dev-docs/high-level/development-workflow.md`](./dev-docs/high-level/development-workflow.md)
- Testing (unit + acceptance) and required env: [`dev-docs/high-level/testing.md`](./dev-docs/high-level/testing.md)
- Generated clients (Kibana `kbapi`) and regeneration: [`dev-docs/high-level/generated-clients.md`](./dev-docs/high-level/generated-clients.md)
- Documentation generation: [`dev-docs/high-level/documentation.md`](./dev-docs/high-level/documentation.md)

## After making changes

- Ensure the project builds - `make build`
- Ensure any new/updated acceptance tests pass (via `go test`). Check the [testing](./dev-docs/high-level/testing.md) docs for an example of running targeted tests. Check if the Elastic stack is available using the default variables in [testing](./dev-docs/high-level/testing.md) before trying to create new Stack services.