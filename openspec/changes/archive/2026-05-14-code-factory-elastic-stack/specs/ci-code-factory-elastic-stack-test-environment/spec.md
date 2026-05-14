## ADDED Requirements

### Requirement: Elastic Stack ports SHALL map to AWF-allowed host ports
The `code-factory` issue-intake workflow SHALL publish Elasticsearch on a port included in the AWF `--allow-host-ports` allowlist, and SHALL publish Kibana on a port included in the same allowlist, so the agentic sandbox can reach the stack via `host.docker.internal`.

#### Scenario: Elasticsearch on allowed port
- **WHEN** the `code-factory` workflow starts the Elastic Stack
- **THEN** Elasticsearch SHALL be reachable from the agentic sandbox at `http://host.docker.internal` on a port that is in the AWF `--allow-host-ports` allowlist such as `8080`

#### Scenario: Kibana on allowed port
- **WHEN** the `code-factory` workflow starts the Elastic Stack
- **THEN** Kibana SHALL be reachable from the agentic sandbox at `http://host.docker.internal` on a port that is in the AWF `--allow-host-ports` allowlist such as `80`

### Requirement: Docker Compose bind address SHALL be configurable per environment
The repository's `docker-compose.yml` SHALL support an optional `BIND_ADDRESS` environment variable for the `elasticsearch` and `kibana` service port mappings. The default value SHALL be `127.0.0.1` for safe local development, and CI workflows SHALL be able to override it to `0.0.0.0` so that `host.docker.internal` can reach the services from the agentic sandbox.

#### Scenario: Default bind address for local development
- **WHEN** a developer runs `docker compose up` without setting `BIND_ADDRESS`
- **THEN** Elasticsearch and Kibana ports SHALL bind to `127.0.0.1` (localhost-only)

#### Scenario: CI override to bind all interfaces
- **WHEN** the `code-factory` workflow sets `BIND_ADDRESS=0.0.0.0`
- **THEN** Elasticsearch and Kibana ports SHALL bind to `0.0.0.0` so that connections from the Docker bridge network (e.g. `host.docker.internal`) are accepted

### Requirement: The code-factory agent SHALL use remapped ports for acceptance tests
The agent prompt in the `code-factory` workflow SHALL instruct the agent to connect to Elasticsearch and Kibana using the remapped, AWF-allowed ports, and the verification task for acceptance tests SHALL use those same remapped ports.

#### Scenario: Agent runs acceptance tests with remapped ports
- **WHEN** the implementation agent executes acceptance tests during a `code-factory` run
- **THEN** the agent SHALL use `ELASTICSEARCH_ENDPOINTS=http://host.docker.internal:8080` and `KIBANA_ENDPOINT=http://host.docker.internal` for the test environment

#### Scenario: Agent prompt reflects reachable test environment
- **WHEN** the implementation agent reads the test environment instructions in the agent prompt
- **THEN** the prompt SHALL describe the Elastic Stack as reachable on port `8080` for Elasticsearch and port `80` for Kibana via `host.docker.internal`

### Requirement: Terraform CLI SHALL be discoverable inside the agentic sandbox
The `code-factory` workflow SHALL stage the Terraform binary into the tracked workspace so the agentic sandbox can discover and execute it during acceptance test runs, because the AWF container does not mount the GitHub Actions toolcache where `hashicorp/setup-terraform` installs the binary.

#### Scenario: Agent discovers Terraform during verification
- **WHEN** the implementation agent runs `terraform` as part of acceptance tests
- **THEN** the binary SHALL be discoverable within the agentic sandbox's PATH or at a known workspace-relative path

#### Scenario: Terraform is staged before agent activation
- **WHEN** the workflow runs the `Setup Terraform CLI` step
- **THEN** a subsequent step SHALL copy or link the Terraform binary into the workspace so it is available inside the sandboxed agent container
