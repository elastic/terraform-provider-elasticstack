## Context

The `code-factory` and `reproducer-factory` workflows run an AWF (AI Workflow Framework) agent inside a sandboxed Docker container. The agent needs to run acceptance tests against an Elasticsearch and Kibana stack started via `make docker-fleet` (Docker Compose) on the GitHub Actions runner.

Three infrastructure problems prevented the agent from reaching the stack and using Terraform:

1. **AWF firewall blocks ES/Kibana ports**: The AWF sandbox `--allow-host-ports` allowlist only covers `80`, `443`, and `8080` by default. Elasticsearch (9200) and Kibana (5601) are blocked from within the agent.
2. **Bind address isolation on Linux**: `docker-compose.yml` explicitly binds ports to `127.0.0.1`. On Linux runners, `host.docker.internal` resolves to the Docker bridge gateway (e.g. `172.17.0.1`), not loopback. Even if the firewall allowed 9200/5601, the services would refuse connections from the bridge gateway.
3. **Terraform binary not mounted**: `hashicorp/setup-terraform` installs to `RUNNER_TEMP` (`/opt/hostedtoolcache`), which is **not** mounted into the AWF container. The agent's runtime PATH searches directories that are empty inside the sandbox.

A fourth organisational problem: the dev-tooling setup and stack setup were either duplicated inline (`reproducer-factory`) or missing/outdated (`code-factory`), creating maintenance drift.

## Goals / Non-Goals

**Goals:**
- Make the Elastic Stack reachable from both `code-factory` and `reproducer-factory` agentic sandboxes on Linux runners
- Preserve the existing `localhost`-only security default for local developer runs of `docker-compose.yml`
- Ensure the Terraform CLI is discoverable inside both agentic sandboxes
- Extract shared setup components so both workflows stay in sync
- Keep the change minimal and self-contained (no provider code changes)

**Non-Goals:**
- Changing the AWF `--allow-host-ports` allowlist (requires a framework-level change at GitHub, out of scope)
- Publishing a custom Docker image for the proxies
- Rewriting the orchestration from Docker Compose to GH AW `services:`
- Changing the `provider.yml` CI workflow or local `make` developer workflows
- Adding new Terraform resources or data sources

## Decisions

### Decision 1: Use socat proxy `services` instead of remapping to ports 80/8080

**Rationale:** An earlier design remapped ES to port `8080` and Kibana to port `80` because those ports are in the AWF allowlist by default. This was rejected because:
- Port `80` may conflict with other services and Kibana's 302 behaviour adds confusion.
- The GH AW compiler detects `services:` definitions and opens those ports in the firewall automatically, so dedicated proxy ports (`9201` and `5602`) are cleaner and don't conflict.
- `backplane/socat-forward` accepts configuration via env vars injected through `options`, avoiding the broken `command:` key that caused `invalid reference format` errors in GH AW validation.
- `--add-host host.docker.internal:host-gateway` fixes Linux `host.docker.internal` resolution.

**Alternative considered:** Remap to `8080`/`80`. Rejected: conflicts, confusing redirects, and harder to debug.

### Decision 2: Introduce `ELASTICSEARCH_BIND` and `KIBANA_BIND` env vars defaulting to `127.0.0.1`

**Rationale:** Hard-coding `0.0.0.0` would open ports on developer machines. Separate env vars (replacing a single `BIND_ADDRESS`) allow each service to be configured independently while preserving safe defaults.

**Format:**
```yaml
ports:
  - ${ELASTICSEARCH_BIND:-127.0.0.1}:${ELASTICSEARCH_PORT}:9200
  - ${KIBANA_BIND:-127.0.0.1}:${KIBANA_PORT}:5601
```

**Alternative considered:** Use a second compose file (`docker-compose.ci.yml`). Rejected: fragments the setup and requires keeping both in sync.

### Decision 3: Stage Terraform into the workspace, not the runtime PATH

**Rationale:** The AWF container mounts the workspace (`${GITHUB_WORKSPACE}`) but not `/opt/hostedtoolcache` or `RUNNER_TEMP`. Copying Terraform into a `bin/` directory inside the workspace guarantees the agent can find it.

**Alternative considered:** Install Terraform inside the AWF container at runtime. Rejected: the AWF container is a minimal sandbox without `apt`, `brew`, or internet access to download binaries.

### Decision 4: Extract shared workflow components

**Rationale:** `reproducer-factory` had the correct stack setup inline, while `code-factory` lacked it entirely. The dev-tooling setup (`Setup Go`, `Setup Terraform`, etc.) was also duplicated. Extracting `shared/setup-dev.md` (dev tools) and `shared/elastic-stack.md` (stack infrastructure) means both workflows import the same definitions. Future fixes apply in one place.

**Shared component contracts:**
- `.github/workflows/shared/setup-dev.md` frontmatter provides `steps:` for Go, Terraform, Node.js, and `make setup`.
- `.github/workflows/shared/elastic-stack.md` frontmatter provides `services:` (proxies), `network:` (allowed domains/ecosystems), and `steps:` for stack setup.
- The GH AW compiler merges frontmatter keys (`steps`, `services`, `network`) from imported shared files into the consuming workflow.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| GH AW `safe_outputs` rejects workflow file changes on branches that have them but `main` does not | The infrastructure changes must be merged to `main` before agent-created PRs on these branches will pass `safe_outputs` (`protect_top_level_dot_folders: true`). |
| `host.docker.internal` may behave differently on non-Linux runners | The `--add-host` flag is only needed on Linux; Docker Desktop handles this automatically. The workflow only runs on GitHub-hosted `ubuntu-latest`. |
| Terraform binary copied into workspace can be accidentally committed | The `bin/` directory is created at runtime in CI and is outside the source tree; it is ephemeral on fresh checkouts. |
| Shared workflow file changes affect multiple workflows | This is intentional (DRY), but means shared-file edits must be validated against both consumers (`code-factory` and `reproducer-factory`). |

## Migration Plan

1. Update `docker-compose.yml` with `ELASTICSEARCH_BIND` and `KIBANA_BIND` substitution.
2. Update `.github/workflows/shared/setup-dev.md` with workspace Terraform copy fix.
3. Create `.github/workflows/shared/elastic-stack.md` with proxies, network rules, and stack steps.
4. Update both workflow templates to import the shared files and remove inline equivalents.
5. Run `make workflow-generate` and verify compilation succeeds.
6. Merge the infrastructure changes to `main` (required for `safe_outputs` on future agent runs).
7. Trigger a test run of `reproducer-factory-issue` via `workflow_dispatch` to verify the agent can reach the stack and run `terraform --version`.

## Open Questions

- Should the `change-factory` or `research-factory` workflows also import `shared/elastic-stack.md`?
- Should `shared/setup-dev.md` be imported by any other workflows currently duplicating its steps?
