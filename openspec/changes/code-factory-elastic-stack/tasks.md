## 1. docker-compose.yml changes

- [ ] 1.1 Update `elasticsearch` service `ports` to use `${BIND_ADDRESS:-127.0.0.1}:${ELASTICSEARCH_PORT}:9200`
- [ ] 1.2 Update `kibana` service `ports` to use `${BIND_ADDRESS:-127.0.0.1}:${KIBANA_PORT}:5601`
- [ ] 1.3 Verify `make docker-fleet` still works locally with default `BIND_ADDRESS` (localhost-only)
- [ ] 1.4 Verify `make docker-fleet BIND_ADDRESS=0.0.0.0` binds to all interfaces

## 2. Workflow template changes

- [ ] 2.1 Add `env` block to `Setup Elastic Stack` step in `.github/workflows-src/code-factory-issue/workflow.md.tmpl`:
  - `BIND_ADDRESS: "0.0.0.0"`
  - `ELASTICSEARCH_PORT: "8080"`
  - `KIBANA_PORT: "80"`
- [ ] 2.2 Add `Stage Terraform for agent` step in `.github/workflows-src/code-factory-issue/workflow.md.tmpl`:
  - Copy `$(which terraform)` to `.bin/terraform`
  - Ensure `.bin/` is in `PATH` or documented for the agent
- [ ] 2.3 Update the `## Test environment` section in the agent prompt (`workflow.md.tmpl`):
  - Change ES endpoint from `host.docker.internal:9200` to `host.docker.internal:8080`
  - Change KB endpoint from `host.docker.internal:5601` to `http://host.docker.internal` (port 80)
- [ ] 2.4 Update the `## Verification tasks` section in the agent prompt (`workflow.md.tmpl`):
  - Change acceptance test port references to `8080` and `80`

## 3. Compile and validate workflows

- [ ] 3.1 Run `gh aw compile` (or project-specific compilation command) against `.github/workflows-src/code-factory-issue/workflow.md.tmpl`
- [ ] 3.2 Verify the generated `.github/workflows/code-factory-issue.lock.yml` includes the env vars and updated prompt text
- [ ] 3.3 Verify `make check-lint` passes
- [ ] 3.4 Commit changes to branch `code-factory-elastic-stack`

## 4. End-to-end validation

- [ ] 4.1 Trigger `code-factory-issue` workflow via `workflow_dispatch` with a test issue number
- [ ] 4.2 Verify in the run logs that Elasticsearch binds to `0.0.0.0:8080` and Kibana to `0.0.0.0:80`
- [ ] 4.3 Verify the agent sandbox can `curl http://host.docker.internal:8080` and `curl http://host.docker.internal`
- [ ] 4.4 Verify the agent sandbox can run `terraform --version`
- [ ] 4.5 If any step fails, capture logs and iterate
