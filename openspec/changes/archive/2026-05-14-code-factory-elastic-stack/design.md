## Context

The `code-factory` workflow runs an AWF (AI Workflow Framework) agent inside a sandboxed Docker container orchestrated by `awf`. The agent needs to run acceptance tests against an Elasticsearch and Kibana stack that is started via `make docker-fleet` (which uses Docker Compose) in the workflow steps that run on the GitHub Actions runner.

Two infrastructure problems prevent the agent from reaching the stack:

1. **AWF firewall port restriction**: The AWF sandbox only allows outbound host ports `80`, `443`, and `8080` via `--allow-host-ports`. Elasticsearch (9200) and Kibana (5601) are blocked.
2. **Bind address isolation on Linux**: The `docker-compose.yml` explicitly binds ports to `127.0.0.1` (`127.0.0.1:9200:9200` and `127.0.0.1:5601:5601`). On Linux GitHub Actions runners, `host.docker.internal` resolves to the Docker bridge gateway (e.g. `172.17.0.1`), not loopback. Even if the AWF firewall allowed 9200/5601, the services would not accept connections from the bridge gateway because they only listen on `127.0.0.1`.

A third problem: the Terraform binary installed by `hashicorp/setup-terraform` is placed in the GitHub Actions toolcache (`/opt/hostedtoolcache`), which is **not** mounted into the AWF agent container. The agent's PATH augmentation searches those directories but they are empty from the container's perspective.

## Goals / Non-Goals

**Goals:**
- Make the Elastic Stack reachable from the `code-factory` agentic sandbox on Linux runners
- Preserve the existing `localhost`-only security default for local developer runs of `docker-compose.yml`
- Ensure the Terraform CLI is discoverable inside the agentic sandbox for acceptance tests
- Keep the change minimal and self-contained (no provider code changes)

**Non-Goals:**
- Changing the AWF firewall allowlist (controlled by the `gh-aw` compiler, not repo-authored code)
- Rewriting the orchestration from Docker Compose to GH AW `services:`
- Changing the `provider.yml` CI workflow or local `make` developer workflows
- Adding new Terraform resources or data sources

## Decisions

### Decision 1: Remap ES to port 8080 and Kibana to port 80 (instead of requesting firewall changes)

**Rationale:** The AWF `--allow-host-ports` flag is injected by the `gh-aw` compiler. There is no repo-authored frontmatter option to extend it. Adding `9200` and `5601` would require a framework-level change at GitHub, which is out of scope and uncertain timeline. Ports `80` and `8080` are already in the allowlist.

**Alternative considered:** Use `services:` in the frontmatter. Rejected because Docker Compose gives us deterministic orchestration (`depends_on: condition: service_healthy`, named volumes, config file mounts) that would need to be reimplemented as explicit wait loops with `services:`.

### Decision 2: Introduce a `BIND_ADDRESS` env var defaulting to `127.0.0.1`

**Rationale:** Hard-coding `0.0.0.0` would open ports to the network on developer machines, which is a security regression. An env var with a safe default preserves local behavior while allowing CI to opt-in.

**Format:** `ports: ["${BIND_ADDRESS:-127.0.0.1}:${ELASTICSEARCH_PORT}:9200"]`

**Alternative considered:** Use a second compose file (`docker-compose.ci.yml`) or Docker host networking. Rejected: a second compose file fragments the setup and would need to keep both in sync. Host networking is not portable to Docker Desktop.

### Decision 3: Stage Terraform into the workspace, not the runtime PATH

**Rationale:** The AWF container mounts the workspace (`${GITHUB_WORKSPACE}`) but not `/opt/hostedtoolcache`. Copying Terraform into a `.bin/` directory inside the workspace guarantees the agent can find it without relying on implicit mount behavior that may change with AWF updates.

**Alternative considered:** Install Terraform inside the AWF container at runtime. Rejected: the AWF container is a minimal sandbox without `apt`, `brew`, or internet access to download binaries. The `hashicorp/setup-terraform` action already handles installation; we just need visibility.

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Port 80/8080 may already be in use on some runner configurations | Test in a `workflow_dispatch` run before relying on it; fallback is to use `80` only for Kibana since it returns 302s (stateless) and `8080` for ES since it handles high load |
| Kibana returning 302 redirects from `/` could confuse health checks in the agent prompt | The agent prompt already specifies API paths like `/api/fleet/...`; Kibana's 302 on `/` is handled correctly by the provider tests |
| `BIND_ADDRESS` env var is a new, untested convention in this repo | Documented in the workflow step and in the spec; other CI workflows (`provider.yml`) do not set it, so they keep the `127.0.0.1` behavior |
| Terraform binary copied into workspace can be accidentally committed | Add `.bin/` to `.gitignore` if not already present; the file is created at runtime on fresh CI checkouts and the workspace is ephemeral |

## Migration Plan

1. Update `docker-compose.yml` with `BIND_ADDRESS` substitution
2. Update `.github/workflows-src/code-factory-issue/workflow.md.tmpl` with:
   - `env` vars on `Setup Elastic Stack` step
   - `Stage Terraform for agent` step
   - Updated agent prompt text (test environment URLs, verification task)
3. Compile the workflow: `gh aw compile` (or equivalent project script)
4. Run a `workflow_dispatch` test of `code-factory` with a dummy issue to verify the agent can reach the stack and run `terraform --version`
5. Rollback: revert the commit; no stateful changes

## Open Questions

- Should the `change-factory` workflow get the same treatment (it also needs to run tests)?
- Does `gh aw compile` emit a diff that includes the `--allow-host-ports` lock value, confirming the framework has the right allowlist?
