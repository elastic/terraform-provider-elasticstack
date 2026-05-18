## Why

Both `code-factory` and `reproducer-factory` issue-intake workflows need the Elastic Stack available inside the AWF agentic sandbox so that acceptance tests can run during verification. The stack was already started by `reproducer-factory`, but the agent could not reach it because:

1. **AWF firewall blocks non-standard host ports**: Elasticsearch (9200) and Kibana (5601) are outside the default `80, 443, 8080` allowlist.
2. **Linux Docker bridge isolation**: `docker-compose.yml` binds to `127.0.0.1` by default. On Linux runners, `host.docker.internal` resolves to the Docker bridge gateway, which cannot reach loopback-bound ports.
3. **Terraform CLI invisible in sandbox**: `hashicorp/setup-terraform` installs to `RUNNER_TEMP`, which is **not** mounted into the AWF container. The agent could not run `terraform` commands.

Additionally, the stack setup logic and dev-tooling setup were duplicated (or missing) across the two workflows, making maintenance fragile.

## What Changes

- **`docker-compose.yml`**: Make bind addresses configurable via `${ELASTICSEARCH_BIND:-127.0.0.1}` and `${KIBANA_BIND:-127.0.0.1}` for the `elasticsearch` and `kibana` service ports, preserving the safe `localhost`-only default for local development.
- **`.github/workflows/shared/setup-dev.md`**: Update the "Export Go and Terraform paths for AWF chroot mode" step to copy the Terraform binary from `RUNNER_TEMP` into `$GITHUB_WORKSPACE/.bin/terraform` and prepend that directory to `PATH`, so the agentic sandbox can discover it.
- **`.github/workflows/shared/elastic-stack.md`** (new shared workflow): Extract reusable Elastic Stack infrastructure containing:
  - `services:` (`es-proxy` and `kb-proxy` using `backplane/socat-forward`) that bridge ports `9201→9200` and `5602→5601` to `host.docker.internal`.
  - `network:` additions (`terraform` ecosystem) so the AWF firewall allows Terraform registry downloads.
  - `steps:` for stack setup (`make docker-fleet` with `0.0.0.0` bind, Kibana user password, ES API key, Fleet setup, Docker compose logs on failure).
- **`.github/workflows/code-factory-issue.md`**: Import `shared/elastic-stack.md` (in addition to existing `shared/setup-dev.md`); update the agent prompt `## Test environment` section to remove the warning that acceptance tests are blocked and instead describe how to run them.
- **`.github/workflows/reproducer-factory-issue.md`**: Import `shared/setup-dev.md` and `shared/elastic-stack.md`; remove all inline dev-setup and stack-setup steps (they are now provided by the shared files); remove inline `network:` and `services:` frontmatter keys.
- **Regenerate compiled workflows**: Run `make workflow-generate` to produce updated `.github/workflows/code-factory-issue.md`, `.github/workflows/reproducer-factory-issue.md`, and their respective `.lock.yml` files.

## Capabilities

### New Capabilities
- `ci-shared-elastic-stack`: Defines the shared Elastic Stack test environment (socat proxies, network rules, and stack setup steps) consumed by both intake workflows.
- `ci-shared-setup-dev`: Defines reusable dev-tooling setup (Go, Terraform, Node.js, repo dependencies) with the Terraform workspace copy fix.

### Modified Capabilities
- `ci-code-factory-issue-intake`: Now provisions the Elastic Stack and exposes it to the agentic sandbox so acceptance tests can execute as part of verification. The agent prompt no longer warns that tests are blocked.
- `ci-reproducer-factory-issue-intake`: Delegates dev-setup and stack-setup to shared components instead of inline steps, ensuring both workflows stay in sync.

## Impact

- Affects `.github/workflows/code-factory-issue.md`, `.github/workflows/reproducer-factory-issue.md`, and their compiled `.lock.yml` outputs.
- Affects `.github/workflows/shared/setup-dev.md` (updated) and `.github/workflows/shared/elastic-stack.md` (new).
- Affects `docker-compose.yml` (must remain backward-compatible for local dev and other CI workflows).
- No impact on existing `provider.yml` CI workflow or local developer workflows (defaults remain `127.0.0.1`).
- No Terraform resource or data source code changes.
