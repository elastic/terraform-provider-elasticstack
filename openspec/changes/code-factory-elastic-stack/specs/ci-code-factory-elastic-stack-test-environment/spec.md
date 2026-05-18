## ADDED Requirements

### Requirement: Elastic Stack SHALL be reachable through proxy services
Both `code-factory` and `reproducer-factory` issue-intake workflows SHALL expose the Elastic Stack to the agentic sandbox via GH AW `services:` proxy containers. An `es-proxy` service SHALL forward port `9201` to `host.docker.internal:9200`, and a `kb-proxy` service SHALL forward port `5602` to `host.docker.internal:5601`. The proxy containers SHALL use the `backplane/socat-forward` image with env vars (`LISTEN_PORT`, `DEST_PORT`, `DEST_HOST`) passed via `options`, and SHALL include `--add-host host.docker.internal:host-gateway` so `host.docker.internal` resolves on Linux runners.

#### Scenario: Elasticsearch reachable from agent sandbox
- **WHEN** the workflow starts the elastic-stack shared services
- **THEN** Elasticsearch SHALL be reachable from the agentic sandbox at `http://host.docker.internal:9201` (proxied through `es-proxy` to the stack on port 9200)

#### Scenario: Kibana reachable from agent sandbox
- **WHEN** the workflow starts the elastic-stack shared services
- **THEN** Kibana SHALL be reachable from the agentic sandbox at `http://host.docker.internal:5602` (proxied through `kb-proxy` to the stack on port 5601)

#### Scenario: Proxy services use env-var configuration
- **WHEN** maintainers inspect the compiled workflow lock file
- **THEN** the `es-proxy` and `kb-proxy` services SHALL be defined in the agent job `services:` block
- **AND** they SHALL use `backplane/socat-forward` with env vars passed through `options: >-`
- **AND** they SHALL NOT use a `command:` key (which GH AW rejects)

### Requirement: Docker Compose bind address SHALL be configurable per service
The repository's `docker-compose.yml` SHALL support optional `ELASTICSEARCH_BIND` and `KIBANA_BIND` environment variables for the `elasticsearch` and `kibana` service port mappings. The default value SHALL be `127.0.0.1` for safe local development, and CI workflows SHALL be able to override them to `0.0.0.0` so that `host.docker.internal` can reach the services from the agentic sandbox.

#### Scenario: Default bind address for local development
- **WHEN** a developer runs `docker compose up` without setting `ELASTICSEARCH_BIND` or `KIBANA_BIND`
- **THEN** Elasticsearch and Kibana ports SHALL bind to `127.0.0.1` (localhost-only)

#### Scenario: CI override to bind all interfaces
- **WHEN** a workflow sets `ELASTICSEARCH_BIND=0.0.0.0` and `KIBANA_BIND=0.0.0.0`
- **THEN** Elasticsearch and Kibana ports SHALL bind to `0.0.0.0` so that connections from the Docker bridge network (e.g. `host.docker.internal`) are accepted

### Requirement: Stack setup steps SHALL be defined in a shared workflow
The stack setup logic (`make docker-fleet`, `make set-kibana-password`, `make create-es-api-key`, `make setup-kibana-fleet`, and failure-time `docker compose logs`) SHALL be defined once in `.github/workflows/shared/elastic-stack.md` and imported by both `code-factory-issue` and `reproducer-factory-issue` workflows. This ensures both workflows use identical stack configuration and that future fixes apply to both consumers.

#### Scenario: code-factory imports shared stack setup
- **WHEN** the `code-factory-issue` workflow is compiled
- **THEN** the compiled output SHALL include the stack setup steps in the agent job

#### Scenario: reproducer-factory imports shared stack setup
- **WHEN** the `reproducer-factory-issue` workflow is compiled
- **THEN** the compiled output SHALL include the stack setup steps in the agent job
- **AND** it SHALL NOT duplicate those steps inline in its source template

### Requirement: Terraform CLI SHALL be discoverable inside the agentic sandbox
The `shared/setup-dev.md` workflow component SHALL stage the Terraform binary into the tracked workspace so both `code-factory` and `reproducer-factory` agentic sandboxes can discover and execute it, because the AWF container does not mount the GitHub Actions toolcache where `hashicorp/setup-terraform` installs the binary.

#### Scenario: Agent discovers Terraform during verification
- **WHEN** the implementation agent runs `terraform` as part of acceptance tests
- **THEN** the binary SHALL be discoverable within the agentic sandbox's PATH at `$GITHUB_WORKSPACE/.bin/terraform`

#### Scenario: Terraform is staged before agent activation
- **WHEN** the workflow runs the `Setup Terraform CLI` step
- **THEN** a subsequent step SHALL copy the Terraform binary into `$GITHUB_WORKSPACE/.bin/terraform`
- **AND** it SHALL prepend `$GITHUB_WORKSPACE/.bin` to `PATH`

### Requirement: AWF network policy SHALL allow Terraform ecosystem access
Both `code-factory` and `reproducer-factory` workflows SHALL include `terraform` in their AWF network allowlist (either directly or via the shared `elastic-stack.md` component) so the agentic sandbox can download providers and modules from `registry.terraform.io` and `releases.hashicorp.com`.

#### Scenario: Terraform registry reachable from sandbox
- **WHEN** the agent runs `terraform init` or provider downloads
- **THEN** connections to `registry.terraform.io` and `releases.hashicorp.com` SHALL be allowed by the AWF firewall

### Requirement: New capabilities SHALL have corresponding spec delta directories
New capabilities introduced by this change (`ci-shared-elastic-stack` and `ci-shared-setup-dev`) SHALL each have a corresponding spec delta directory under `openspec/changes/code-factory-elastic-stack/specs/` so that shared workflow contracts are captured on archive.

#### Scenario: Maintainer archives the change
- **WHEN** the change is archived and its delta specs are synced to main specs
- **THEN** each new capability SHALL have a spec directory under `openspec/changes/code-factory-elastic-stack/specs/`
- **AND** the delta spec status SHALL be `ADDED` (not `MODIFIED`) for net-new capabilities
