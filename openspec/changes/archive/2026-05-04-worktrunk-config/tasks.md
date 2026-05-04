## 1. docker-compose.yml cleanup

- [x] 1.1 Remove `container_name:` directive from the `elasticsearch` service
- [x] 1.2 Remove `container_name:` directive from the `kibana_settings` service
- [x] 1.3 Remove `container_name:` directive from the `kibana` service
- [x] 1.4 Remove `container_name:` directive from the `fleet_settings` service
- [x] 1.5 Remove `container_name:` directive from the `fleet` service
- [x] 1.6 Remove `container_name:` directive from the `acceptance-tests` service
- [x] 1.7 Remove `container_name:` directive from the `token-acceptance-tests` service
- [x] *(review fix)* Remove `container_name:` directive from the `kibana_certs` service in `docker-compose.tls.yml`
- [x] *(review fix)* Update `.buildkite/scripts/update-kibana-client.sh` to use `docker compose logs` instead of hardcoded container names

## 2. .env.template

- [x] 2.1 Create `.env.template` from current `.env`: copy all lines except `*_CONTAINER_NAME`, `ELASTICSEARCH_PORT`, `KIBANA_PORT`, and `ELASTICSEARCH_URL`
- [x] 2.2 Add root `.env` to `.gitignore` while ensuring `.env.template` is not listed there
- [x] 2.3 Stop tracking the existing committed root `.env` so the generated per-worktree `.env` does not produce persistent Git diffs

## 3. Makefile port variables

- [x] 3.1 Add `ELASTICSEARCH_PORT ?= 9200` and `KIBANA_PORT ?= 5601` near the top of the Makefile (alongside other configurable variables)
- [x] 3.2 Update `testacc-vs-docker` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)` and `localhost:5601` with `localhost:$(KIBANA_PORT)`
- [x] 3.3 Update `set-kibana-password` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)`
- [x] 3.4 Update `setup-synthetics` target: replace `localhost:5601` with `localhost:$(KIBANA_PORT)`
- [x] 3.5 Update `create-es-api-key` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)`
- [x] 3.6 Update `create-es-bearer-token` target: replace `localhost:9200` with `localhost:$(ELASTICSEARCH_PORT)`
- [x] 3.7 Update `setup-kibana-fleet` target: replace `localhost:5601` with `localhost:$(KIBANA_PORT)`

## 4. Worktrunk project config

- [x] 4.1 Create `.config/wt.toml` with `pre-start = "make setup"`
- [x] 4.2 Add `post-start` hook to `.config/wt.toml`: copy `.env.template` to `.env` then append `ELASTICSEARCH_PORT`, `KIBANA_PORT`, and `ELASTICSEARCH_URL` lines using `hash_port`
- [x] 4.3 Add `pre-commit = "make check-lint"` to `.config/wt.toml`
- [x] 4.4 Add `pre-remove = "docker compose down --volumes"` to `.config/wt.toml`
- [x] *(review fix)* Make `post-start` copy conditional with `cp -n`, add `set -e`, use distinct hash inputs for ES vs KB ports, and use explicit port in `ELASTICSEARCH_URL`

## 5. Developer documentation

- [x] 5.1 Add a section to `dev-docs/high-level/development-workflow.md` covering worktrunk setup: install shell integration (`wt config shell install`), user config worktree path template, and how to export `.env` before running Makefile port-dependent targets in a feature worktree
- [x] *(review fix)* Correct user config path, clarify worktree location inside bare repo, expand list of port-dependent targets, explain .env provenance, add cleanup note
