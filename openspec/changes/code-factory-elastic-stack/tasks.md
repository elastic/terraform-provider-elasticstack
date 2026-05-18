## 1. docker-compose.yml changes

- [x] 1.1 Update `elasticsearch` service `ports` to use `${ELASTICSEARCH_BIND:-127.0.0.1}:${ELASTICSEARCH_PORT}:9200`
- [x] 1.2 Update `kibana` service `ports` to use `${KIBANA_BIND:-127.0.0.1}:${KIBANA_PORT}:5601`
- [x] 1.3 Verify `make docker-fleet` still works locally with default bind addresses (localhost-only)
- [x] 1.4 Verify `make docker-fleet` with `ELASTICSEARCH_BIND=0.0.0.0 KIBANA_BIND=0.0.0.0` binds to all interfaces

## 2. Shared workflow: setup-dev.md

- [x] 2.1 Update the "Export Go and Terraform paths for AWF chroot mode" step in `.github/workflows/shared/setup-dev.md`:
  - Copy `$(which terraform)` into `$GITHUB_WORKSPACE/.bin/terraform`
  - Export `TERRAFORM_BIN=$GITHUB_WORKSPACE/.bin/terraform`
  - Export `PATH=$GITHUB_WORKSPACE/.bin:$PATH`

## 3. Shared workflow: elastic-stack.md (new)

- [x] 3.1 Create `.github/workflows/shared/elastic-stack.md` with frontmatter:
  - `services:` containing `es-proxy` and `kb-proxy` using `backplane/socat-forward`
  - `network:` allowed list including `terraform` ecosystem
  - `steps:` for stack setup (`make docker-fleet`, `make set-kibana-password`, `make create-es-api-key`, `make setup-kibana-fleet`, `docker compose logs` on failure)
- [x] 3.2 Ensure proxy services use `--add-host host.docker.internal:host-gateway`
- [x] 3.3 Ensure proxy env vars (`LISTEN_PORT`, `DEST_PORT`, `DEST_HOST`) are passed via `options: >- -e ...`

## 4. Workflow template: code-factory-issue

- [x] 4.1 Update `imports:` to include `shared/elastic-stack.md`
- [x] 4.2 Remove inline `network:` frontmatter key (now provided by shared file)
- [x] 4.3 Update agent prompt `## Test environment` section to remove the "acceptance tests are currently blocked" warning

## 5. Workflow template: reproducer-factory-issue

- [x] 5.1 Add `imports: [shared/setup-dev.md, shared/elastic-stack.md]`
- [x] 5.2 Remove inline dev-setup steps (`Setup Go`, `Export Go paths`, `Setup Node.js`, `Setup Terraform CLI`, `Export Go+Terraform paths`, `Get dependencies`)
- [x] 5.3 Remove inline stack-setup steps (`Setup Elastic Stack`, `Setup Kibana user`, `Get ES API key`, `Setup Fleet`, `Docker compose logs`)
- [x] 5.4 Remove inline `network:` and `services:` frontmatter keys (now provided by shared file)
- [x] 5.5 Keep the `Download issue context artifact` inline step

## 6. Compile and validate workflows

- [x] 6.1 Run `make workflow-generate`
- [x] 6.2 Verify both `.lock.yml` files contain the proxy services (`es-proxy`, `kb-proxy`) and `terraform` in allowed domains
- [x] 6.3 Verify both `.lock.yml` files contain the stack setup steps in the agent job
- [x] 6.4 Verify both `.lock.yml` files contain the Terraform workspace copy step

## 7. End-to-end validation

- [ ] 7.1 Merge workflow infrastructure changes to `main` (required for `safe_outputs` on agent-created PRs)
- [ ] 7.2 Trigger `reproducer-factory-issue` workflow via `workflow_dispatch` with a test issue number
- [ ] 7.3 Verify the agent sandbox can `curl http://host.docker.internal:9200` and `curl http://host.docker.internal:5601`
- [ ] 7.4 Verify the agent sandbox can run `terraform --version`
- [ ] 7.5 Verify acceptance tests (`TF_ACC=1`) can execute successfully against the stack
- [ ] 7.6 If any step fails, capture logs and iterate
