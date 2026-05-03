## 1. docker-compose.yml cleanup

- [ ] 1.1 Remove `container_name:` directive from the `elasticsearch` service
- [ ] 1.2 Remove `container_name:` directive from the `kibana_settings` service
- [ ] 1.3 Remove `container_name:` directive from the `kibana` service
- [ ] 1.4 Remove `container_name:` directive from the `fleet_settings` service
- [ ] 1.5 Remove `container_name:` directive from the `fleet` service
- [ ] 1.6 Remove `container_name:` directive from the `acceptance-tests` service
- [ ] 1.7 Remove `container_name:` directive from the `token-acceptance-tests` service

## 2. .env.template

- [ ] 2.1 Create `.env.template` from current `.env`: copy all lines except `*_CONTAINER_NAME`, `ELASTICSEARCH_PORT`, `KIBANA_PORT`, and `ELASTICSEARCH_URL`
- [ ] 2.2 Add root `.env` to `.gitignore` while ensuring `.env.template` is not listed there
- [ ] 2.3 Stop tracking the existing committed root `.env` so the generated per-worktree `.env` does not produce persistent Git diffs

## 3. Makefile port variables

- [ ] 3.1 Add `ELASTICSEARCH_PORT ?= 9200` and `KIBANA_PORT ?= 5601` near the top of the Makefile (alongside other configurable variables)
- [ ] 3.2 Update `testacc-vs-docker` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)` and `localhost:5601` with `localhost:$(KIBANA_PORT)`
- [ ] 3.3 Update `set-kibana-password` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)`
- [ ] 3.4 Update `setup-synthetics` target: replace `localhost:5601` with `localhost:$(KIBANA_PORT)`
- [ ] 3.5 Update `create-es-api-key` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)`
- [ ] 3.6 Update `create-es-bearer-token` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)`
- [ ] 3.7 Update `setup-kibana-fleet` target: replace `localhost:5601` with `localhost:$(KIBANA_PORT)`

## 4. Worktrunk project config

- [ ] 4.1 Create `.config/wt.toml` with `pre-start = "make setup"`
- [ ] 4.2 Add `post-start` hook to `.config/wt.toml`: copy `.env.template` to `.env` then append `ELASTICSEARCH_PORT`, `KIBANA_PORT`, and `ELASTICSEARCH_URL` lines using `hash_port`
- [ ] 4.3 Add `pre-commit = "make check-lint"` to `.config/wt.toml`
- [ ] 4.4 Add `pre-remove = "docker compose down --volumes"` to `.config/wt.toml`

## 5. Developer documentation

- [ ] 5.1 Add a section to `dev-docs/high-level/development-workflow.md` covering worktrunk setup: install shell integration (`wt config shell install`), user config worktree path template, and how to export `.env` before running Makefile port-dependent targets in a feature worktree
