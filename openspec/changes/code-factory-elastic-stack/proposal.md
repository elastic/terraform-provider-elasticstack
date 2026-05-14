## Why

The `code-factory` issue-intake workflow fails to run acceptance tests because the agentic sandbox cannot reach the Elastic Stack services (Elasticsearch on port 9200, Kibana on port 5601) from inside the AWF firewall. Two problems compound: the AWF `--allow-host-ports` allowlist only includes `80, 443, 8080`, and the `docker-compose.yml` binds exposed ports to `127.0.0.1` which `host.docker.internal` cannot reach on Linux runners. This makes the acceptance test verification step in the agent prompt impossible, causing implementation PRs to be created without actual test validation.

## What Changes

- **`docker-compose.yml`**: Make the bind address configurable via `${BIND_ADDRESS:-127.0.0.1}` for the `elasticsearch` and `kibana` service ports, preserving the safe `localhost`-only default for local development.
- **`.github/workflows-src/code-factory-issue/workflow.md.tmpl`**: Add environment variables (`BIND_ADDRESS=0.0.0.0`, `ELASTICSEARCH_PORT=8080`, `KIBANA_PORT=80`) to the `Setup Elastic Stack` step so the stack binds to `0.0.0.0` on allowed ports.
- **`.github/workflows-src/code-factory-issue/workflow.md.tmpl`**: Update the agent prompt's test environment instructions and verification tasks to use ports `8080` and `80` with `host.docker.internal`.
- **`.github/workflows-src/code-factory-issue/workflow.md.tmpl`**: Add a `Stage Terraform for agent` step that copies the Terraform binary from the runner toolcache into the workspace so the agentic sandbox can discover it.
- **`.github/workflows/code-factory-issue.lock.yml`**: Regenerate from the updated source template via `gh aw compile` (or the project's workflow compilation process).

## Capabilities

### New Capabilities
- `ci-code-factory-elastic-stack-test-environment`: Defines how the Elastic Stack test services are exposed to the `code-factory` agentic sandbox, including port mapping to the AWF allowlist and bind-address configuration.

### Modified Capabilities
- `ci-code-factory-issue-intake`: The workflow SHALL expose the Elastic Stack on ports that are reachable from within the AWF agentic sandbox so that acceptance tests can be executed as part of the verification tasks.

## Impact

- Affects `.github/workflows-src/code-factory-issue/workflow.md.tmpl` and `.github/workflows/code-factory-issue.lock.yml`.
- Affects `docker-compose.yml` (must remain backward-compatible for local dev and other CI workflows).
- No impact on existing `provider.yml` CI workflow or local developer workflows.
- No Terraform resource or data source code changes.
