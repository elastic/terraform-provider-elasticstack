## ADDED Requirements

### Requirement: Elastic Stack SHALL be reachable through proxy services
The `shared/elastic-stack.md` workflow component SHALL expose the Elastic Stack to the agentic sandbox via GH AW `services:` proxy containers. An `es-proxy` service SHALL forward port `9201` to `host.docker.internal:9200`, and a `kb-proxy` service SHALL forward port `5602` to `host.docker.internal:5601`. The proxy containers SHALL use the `backplane/socat-forward` image with env vars (`LISTEN_PORT`, `DEST_PORT`, `DEST_HOST`) passed via `options`, and SHALL include `--add-host host.docker.internal:host-gateway` so `host.docker.internal` resolves on Linux runners.

#### Scenario: Elasticsearch reachable from agent sandbox
- **WHEN** a workflow imports `shared/elastic-stack.md`
- **THEN** Elasticsearch SHALL be reachable from the agentic sandbox at `http://host.docker.internal:9201` (proxied through `es-proxy` to the stack on port 9200)

#### Scenario: Kibana reachable from agent sandbox
- **WHEN** a workflow imports `shared/elastic-stack.md`
- **THEN** Kibana SHALL be reachable from the agentic sandbox at `http://host.docker.internal:5602` (proxied through `kb-proxy` to the stack on port 5601)

#### Scenario: Proxy services use env-var configuration
- **WHEN** maintainers inspect a compiled workflow lock file for a consumer of `shared/elastic-stack.md`
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

### Requirement: AWF network policy SHALL allow Terraform ecosystem access
Workflows that import `shared/elastic-stack.md` SHALL include `terraform` in their AWF network allowlist (provided by the shared component) so the agentic sandbox can download providers and modules from `registry.terraform.io` and `releases.hashicorp.com`.

#### Scenario: Terraform registry reachable from sandbox
- **WHEN** the agent runs `terraform init` or provider downloads
- **THEN** connections to `registry.terraform.io` and `releases.hashicorp.com` SHALL be allowed by the AWF firewall
