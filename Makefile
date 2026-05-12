# Preserve environment and command-line variable values across .env inclusion.
# The .env file (auto-created from .env.template) may contain defaults that
# would otherwise silently override values set via workflow matrices or local shell.
define _env_guard_save
_$(1)_ORIGIN := $(origin $(1))
_$(1)_VALUE  := $($(1))
endef

define _env_guard_restore
ifneq ($(filter environment command line,$(_$(1)_ORIGIN)),)
  override $(1) := $(_$(1)_VALUE)
endif
endef

# Guard variables present in both .env.template and CI/local usage.
_ENV_GUARD_VARS := STACK_VERSION FLEET_IMAGE ELASTICSEARCH_PASSWORD KIBANA_PASSWORD

$(foreach v,$(_ENV_GUARD_VARS),$(eval $(call _env_guard_save,$v)))

-include .env

$(foreach v,$(_ENV_GUARD_VARS),$(eval $(call _env_guard_restore,$v)))

.DEFAULT_GOAL = help
SHELL := /bin/bash

VERSION ?= 0.14.5

NAME = elasticstack
BINARY = terraform-provider-${NAME}
MARCH = "$$(go env GOOS)_$$(go env GOARCH)"

ACCTEST_PARALLELISM ?= 10
ACCTEST_PACKAGE_PARALLELISM ?= 6
ACCTEST_TOTAL_SHARDS ?= 1
ACCTEST_SHARD_INDEX ?= 0
ACCTEST_TIMEOUT = 120m
ACCTEST_COUNT = 1
TEST ?= ./...

USE_TLS ?= 0
COMPOSE_FILE := docker-compose.yml
ifeq ($(USE_TLS),1)
	COMPOSE_FILE := docker-compose.tls.yml
endif

ELASTICSEARCH_USERNAME ?= elastic
ELASTICSEARCH_PASSWORD ?= password
ELASTICSEARCH_PORT ?= 9200
KIBANA_PORT ?= 5601

export ELASTICSEARCH_PORT KIBANA_PORT

# Auto-create .env from template so docker-compose and Make targets work
# when the repo is checked out without a committed .env (e.g. CI, fresh clone).
# Worktrunk worktrees generate their own .env, so this only runs once.
.env:
	@test -f $@ || cp .env.template $@

KIBANA_SYSTEM_USERNAME ?= kibana_system
KIBANA_SYSTEM_PASSWORD ?= password
KIBANA_API_KEY_NAME ?= kibana-api-key

FLEET_NAME ?= fleet
FLEET_ENDPOINT ?= https://$(FLEET_NAME):8220

# Fleet Server image repository. Some older stack versions (notably 8.0.x, 8.1.x)
# do not publish elastic-agent images to docker.elastic.co, so fall back to Docker Hub.
ifneq (,$(filter 8.0.% 8.1.%,$(STACK_VERSION)))
FLEET_IMAGE := elastic/elastic-agent
endif

RERUN_FAILS ?= 5
RERUN_FAILS_MAX_FAILURES ?= 20

export GOBIN = $(shell pwd)/bin

# OpenSpec CLI (see package.json); installed via `make setup-openspec`
OPENSPEC_BIN := $(CURDIR)/node_modules/.bin/openspec

$(GOBIN): ## create bin/ in the current directory
	mkdir -p $(GOBIN)

## Downloads all the Golang dependencies.
vendor:
	@ go mod download

.PHONY: build-ci
build-ci: ## build the terraform provider
	go build -o ${BINARY}

.PHONY: build
build: lint build-ci ## build the terraform provider

# run acceptance tests against the docker container that has been started with `make docker-kibana` (or `make docker-elasticsearch`)
# To run specific test (e.g. TestAccResourceActionConnector) execute `make testacc-vs-docker TESTARGS='-run ^TestAccResourceKibanaConnectorBedrock$$'`
.PHONY: testacc-vs-docker
testacc-vs-docker:
	@ ELASTICSEARCH_ENDPOINTS=http://localhost:$(ELASTICSEARCH_PORT) KIBANA_ENDPOINT=http://localhost:$(KIBANA_PORT) ELASTICSEARCH_USERNAME=$(ELASTICSEARCH_USERNAME) ELASTICSEARCH_PASSWORD=$(ELASTICSEARCH_PASSWORD) make testacc

.PHONY: testacc
testacc: ## Run acceptance tests
	TF_ACC=1 go tool gotestsum --format testname --rerun-fails=$(RERUN_FAILS) --rerun-fails-max-failures=$(RERUN_FAILS_MAX_FAILURES) --packages="$(shell go list ./... | sort | awk '(NR-1) % $(ACCTEST_TOTAL_SHARDS) == $(ACCTEST_SHARD_INDEX)')" -- -p $(ACCTEST_PACKAGE_PARALLELISM) -v -count $(ACCTEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

.PHONY: hook-test
hook-test: ## Run hook JavaScript unit tests
	@ node --test .agents/hooks/*.test.mjs

.PHONY: test
test: workflow-test hook-test ## Run unit tests and JS tests
	go test -v $(TEST) $(TESTARGS) -timeout=5m -parallel=4 -count=1

CURL_BASE_OPTS = -sS --retry 5 --retry-all-errors -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json"
CURL_OPTS = $(CURL_BASE_OPTS) -X POST
FLEET_DEFAULT_DOWNLOAD_SOURCE_ID = terraform-acc-fleet-default-download-source
FLEET_DEFAULT_DOWNLOAD_SOURCE_NAME = Terraform Acceptance Default Agent Download Source
FLEET_DEFAULT_DOWNLOAD_SOURCE_HOST = https://artifacts.elastic.co/downloads/elastic-agent

# To run specific test (e.g. TestAccResourceActionConnector) execute `make docker-testacc TESTARGS='-run ^TestAccResourceActionConnector$$'`
# To enable tracing (or debugging), execute `make docker-testacc TF_LOG=TRACE`
.PHONY: docker-testacc
docker-testacc: docker-fleet ## Run acceptance tests in the docker container
	@ docker compose -f $(COMPOSE_FILE) --profile acceptance-tests up --quiet-pull acceptance-tests

# To run specific test (e.g. TestAccResourceActionConnector) execute `make docker-testacc TESTARGS='-run ^TestAccResourceActionConnector$$'`
# To enable tracing (or debugging), execute `make docker-testacc TF_LOG=TRACE`
.PHONY: docker-testacc-with-token
docker-testacc-with-token: docker-fleet
	@ export ELASTICSEARCH_BEARER_TOKEN=$(shell $(MAKE) create-es-bearer-token | jq -r .access_token); \
	docker compose -f $(COMPOSE_FILE) --profile token-acceptance-tests up --quiet-pull token-acceptance-tests;

.PHONY: docker-elasticsearch
docker-elasticsearch: .env ## Start Elasticsearch single node cluster in docker container
	@ docker compose -f $(COMPOSE_FILE) up --quiet-pull -d elasticsearch

.PHONY: docker-kibana
docker-kibana: .env  ## Start Kibana node in docker container
	@ docker compose -f $(COMPOSE_FILE) up --quiet-pull -d kibana

.PHONY: docker-fleet
docker-fleet: .env ## Start Fleet node in docker container
	@ export KIBANA_CONFIG_FILE=$$(if [ "$(STACK_VERSION)" = "9.4.0" ]; then echo "kibana-9.4.yml"; else echo "kibana.yml"; fi); \
	docker compose -f $(COMPOSE_FILE) up --quiet-pull -d fleet

.PHONY: set-kibana-password
set-kibana-password: ## Sets the ES KIBANA_SYSTEM_USERNAME's password to KIBANA_SYSTEM_PASSWORD. This expects Elasticsearch to be available at localhost:9200 (set via ELASTICSEARCH_PORT env var)
	@ curl $(CURL_OPTS) http://localhost:$(ELASTICSEARCH_PORT)/_security/user/$(KIBANA_SYSTEM_USERNAME)/_password -d '{"password":"$(KIBANA_SYSTEM_PASSWORD)"}'

.PHONY: setup-synthetics
setup-synthetics: ## Creates the synthetics policy required to run Synthetics. This expects Kibana to be available at localhost:5601 (set via KIBANA_PORT env var)
	@ curl $(CURL_OPTS) -H "kbn-xsrf: true" http://localhost:$(KIBANA_PORT)/api/fleet/epm/packages/synthetics/1.2.2 -d '{"force": true}'

.PHONY: create-es-api-key
create-es-api-key: ## Creates and outputs a new API Key. This expects Elasticsearch to be available at localhost:9200 (set via ELASTICSEARCH_PORT env var)
	@ curl $(CURL_OPTS) http://localhost:$(ELASTICSEARCH_PORT)/_security/api_key -d '{"name":"$(KIBANA_API_KEY_NAME)"}'

.PHONY: create-es-bearer-token
create-es-bearer-token: ## Creates and outputs a new OAuth bearer token. This expects Elasticsearch to be available at localhost:9200 (set via ELASTICSEARCH_PORT env var)
	@ curl $(CURL_OPTS) http://localhost:$(ELASTICSEARCH_PORT)/_security/oauth2/token -d '{"grant_type":"client_credentials"}'

.PHONY: setup-kibana-fleet
setup-kibana-fleet: ## Creates the agent and integration policies required to run Fleet. This expects Kibana to be available at localhost:5601 (set via KIBANA_PORT env var)
	curl $(CURL_OPTS) -H "kbn-xsrf: true" http://localhost:$(KIBANA_PORT)/api/fleet/fleet_server_hosts -d '{"name":"default","host_urls":["$(FLEET_ENDPOINT)"],"is_default":true}'
	curl $(CURL_OPTS) -H "kbn-xsrf: true" http://localhost:$(KIBANA_PORT)/api/fleet/agent_policies -d '{"id":"fleet-server","name":"Fleet Server","namespace":"default","monitoring_enabled":["logs","metrics"]}'
	curl $(CURL_OPTS) -H "kbn-xsrf: true" http://localhost:$(KIBANA_PORT)/api/fleet/package_policies -d '{"name":"fleet-server","namespace":"default","policy_id":"fleet-server","enabled":true,"inputs":[{"type":"fleet-server","enabled":true,"streams":[],"vars":{}}],"package":{"name":"fleet_server","version":"1.5.0"}}'
	@ download_sources="$$(mktemp)"; \
	trap 'rm -f "$$download_sources"' EXIT; \
	status="$$(curl $(CURL_BASE_OPTS) -o "$$download_sources" -w '%{http_code}' -H "kbn-xsrf: true" http://localhost:$(KIBANA_PORT)/api/fleet/agent_download_sources)"; \
	case "$$status" in \
		2*) ;; \
		404) exit 0 ;; \
		*) echo "Unexpected response listing Kibana agent download sources: HTTP $$status" >&2; exit 1 ;; \
	esac; \
	if jq -e '.items[]? | select(.is_default == true and (.host // "") != "")' "$$download_sources" >/dev/null; then \
		exit 0; \
	fi; \
	status="$$(curl $(CURL_OPTS) -o /dev/null -w '%{http_code}' -H "kbn-xsrf: true" http://localhost:$(KIBANA_PORT)/api/fleet/agent_download_sources -d '{"id":"$(FLEET_DEFAULT_DOWNLOAD_SOURCE_ID)","name":"$(FLEET_DEFAULT_DOWNLOAD_SOURCE_NAME)","host":"$(FLEET_DEFAULT_DOWNLOAD_SOURCE_HOST)","is_default":true}')"; \
	case "$$status" in \
		2*) ;; \
		400|409) \
			status="$$(curl $(CURL_BASE_OPTS) -X PUT -o /dev/null -w '%{http_code}' -H "kbn-xsrf: true" http://localhost:$(KIBANA_PORT)/api/fleet/agent_download_sources/$(FLEET_DEFAULT_DOWNLOAD_SOURCE_ID) -d '{"name":"$(FLEET_DEFAULT_DOWNLOAD_SOURCE_NAME)","host":"$(FLEET_DEFAULT_DOWNLOAD_SOURCE_HOST)","is_default":true}')"; \
			case "$$status" in 2*) ;; *) echo "Unexpected response ensuring Kibana agent download source: HTTP $$status" >&2; exit 1 ;; esac ;; \
		*) echo "Unexpected response creating Kibana agent download source: HTTP $$status" >&2; exit 1 ;; \
	esac

.PHONY: docker-clean
docker-clean: .env ## Try to remove provisioned nodes and assigned network
	@ docker compose -f $(COMPOSE_FILE) --profile acceptance-tests down --volumes

.PHONY: copy-kibana-ca
copy-kibana-ca: .env ## Copy Kibana CA certificate to local machine
	@ docker compose -f $(COMPOSE_FILE) cp kibana:/certs/rootCA.pem ./kibana-ca.pem

.PHONY: docs-generate
docs-generate: tools ## Generate documentation for the provider
	@ terraform_version="$$(tr -d '[:space:]' < .terraform-version)"; \
	TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=false go tool github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name terraform-provider-elasticstack --tf-version "$$terraform_version"

.PHONY: workflow-generate
workflow-generate: ## Generate workflow markdown sources
	@ go run ./scripts/compile-workflow-sources --manifest .github/workflows-src/manifest.json
	@ gh aw compile

.PHONY: workflow-test
workflow-test: ## Run unit tests for workflow source generation
	@ go test ./scripts/compile-workflow-sources -run 'TestCompileWorkflow'
	@ go test ./scripts/kibana-spec-impact/... -count=1
	@ node --test .github/workflows-src/lib/*.test.mjs

.PHONY: check-workflows
check-workflows: ## Check generated workflow markdown sources
	@ go run ./scripts/compile-workflow-sources --manifest .github/workflows-src/manifest.json --check --verbose

.PHONY: gen
gen: docs-generate ## Generate the code and documentation
	@ go generate ./...

.PHONY: clean
clean: ## Remove built binary
	rm -f ${BINARY}

.PHONY: install
install: build ## Install built provider into the local terraform cache
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/elastic/${NAME}/${VERSION}/${MARCH}
	mv ${BINARY} ~/.terraform.d/plugins/registry.terraform.io/elastic/${NAME}/${VERSION}/${MARCH}

.PHONY: tools
tools: $(GOBIN)  ## Download golangci-lint locally if necessary.
	@[[ -f $(GOBIN)/golangci-lint ]] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/main/install.sh | sh -s -- -b $(GOBIN) v2.12.2

.PHONY: golangci-lint-custom
golangci-lint-custom: tools
	@ [[ -f $(GOBIN)/golangci-lint-custom ]] || $(GOBIN)/golangci-lint custom

.PHONY: golangci-lint
golangci-lint: golangci-lint-custom
	@ $(GOBIN)/golangci-lint-custom run --max-same-issues=0 $(GOLANGCIFLAGS) ./...

LINT_PERF_DIR := $(CURDIR)/analysis/lint-perf-output/$(shell date +%Y%m%dT%H%M%S)

.PHONY: lint-perf
lint-perf: golangci-lint-custom ## Measure isolated custom-linter performance and write timing/profile artifacts
	@ mkdir -p $(LINT_PERF_DIR)
	@ echo "Writing per-run artifacts to $(LINT_PERF_DIR)"
	@ echo "--- acctestconfigdirlint (golangci isolated run) ---"
	@ { time $(GOBIN)/golangci-lint-custom run --enable-only=acctestconfigdirlint --concurrency=1 \
		--cpu-profile-path=$(LINT_PERF_DIR)/acctestconfigdirlint-golangci-cpu.prof \
		--mem-profile-path=$(LINT_PERF_DIR)/acctestconfigdirlint-golangci-mem.prof \
		--trace-path=$(LINT_PERF_DIR)/acctestconfigdirlint-golangci-trace.out \
		./... ; } 2>&1 | tee $(LINT_PERF_DIR)/acctestconfigdirlint-lint.txt || true
	@ echo "--- analyzer benchmarks ---"
	@ go test ./analysis/acctestconfigdirlint/... -bench=. -benchmem \
		-cpuprofile=$(LINT_PERF_DIR)/acctestconfigdirlint-cpu.prof \
		-memprofile=$(LINT_PERF_DIR)/acctestconfigdirlint-mem.prof \
		-trace=$(LINT_PERF_DIR)/acctestconfigdirlint-trace.out \
		-run='^$$' 2>&1 | tee $(LINT_PERF_DIR)/acctestconfigdirlint-bench.txt || true
	@ echo "Artifacts written to $(LINT_PERF_DIR)/"

.PHONY: lint
lint: GOLANGCIFLAGS += --fix
lint: setup golangci-lint fmt docs-generate ## Run lints to check the spelling and common go patterns

.PHONY: check-lint
check-lint: setup check-openspec golangci-lint check-workflows check-fmt gen check-docs

.PHONY: setup-openspec
setup-openspec: node_modules/.openspec-stamp ## Install Node dependencies (OpenSpec CLI via npm ci)

node_modules/.openspec-stamp: package-lock.json package.json
	@ command -v npm >/dev/null 2>&1 || { echo "npm not found; install Node.js 24.x for OpenSpec" >&2; exit 1; }
	npm ci
	@ touch $@

.PHONY: check-openspec
check-openspec: ## Validate OpenSpec specs and change proposals (structural); requires `make setup` or `make setup-openspec`
	@ test -x $(OPENSPEC_BIN) || { echo "OpenSpec CLI missing; run 'make setup' or 'make setup-openspec'" >&2; exit 1; }
	@ OPENSPEC_TELEMETRY=0 $(OPENSPEC_BIN) validate --all

.PHONY: renovate-post-upgrade
renovate-post-upgrade: vendor notice
	@ make -C generated/kbapi all

.PHONY: notice
notice: vendor
	@ go list -m -json all | go tool go.elastic.co/go-licence-detector  -noticeOut=NOTICE -noticeTemplate ./.NOTICE.tmpl -includeIndirect -rules .notice_rules.json -overrides .notice_overrides.ndjson

.PHONY: fmt
fmt: ## Format code
	go fmt ./...
	terraform fmt --recursive

.PHONY:check-fmt
check-fmt: fmt ## Check if code is formatted
	@if [ "`git status --porcelain `" ]; then \
	  echo "Unformatted files were detected. Please run 'make fmt' to format code, and commit the changes" && echo `git status --porcelain docs/` && exit 1; \
	fi

.PHONY: check-docs
check-docs: docs-generate  ## Check uncommitted changes on docs
	@if [ "`git status --porcelain docs/`" ]; then \
	  echo "Uncommitted changes were detected in the docs folder. Please run 'make docs-generate' to autogenerate the docs, and commit the changes" && echo `git status --porcelain docs/` && exit 1; \
	fi

.PHONY: setup
setup: tools vendor setup-openspec ## Setup the dev environment

.PHONY: prep-release
prep-release: ## Dispatch the release preparation workflow (BUMP=patch|minor|major, default: patch)
	@ BUMP="$(or $(BUMP),patch)"; \
	  case "$$BUMP" in \
	    patch|minor|major) ;; \
	    *) echo "BUMP must be patch, minor, or major (got: $$BUMP)" >&2; exit 1 ;; \
	  esac; \
	  gh workflow run prep-release.yml --field bump="$$BUMP"

.PHONY: release-snapshot
release-snapshot: tools ## Make local-only test release to see if it works using "release" command
	@ go tool github.com/goreleaser/goreleaser/v2 release --snapshot --clean


.PHONY: release-no-publish
release-no-publish: tools check-sign-release ## Make a release without publishing artifacts
	@ go tool github.com/goreleaser/goreleaser/v2 release --skip=publish,announce,validate  --parallelism=2


.PHONY: release
release: tools check-sign-release check-publish-release ## Build, sign, and upload your release
	@ go tool github.com/goreleaser/goreleaser/v2 release --clean  --parallelism=4


.PHONY: check-sign-release
check-sign-release:
ifndef GPG_FINGERPRINT_SECRET
	$(error GPG_FINGERPRINT_SECRET is undefined, but required for signing the release)
endif


.PHONY: check-publish-release
check-publish-release:
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is undefined, but required to make build and upload the released artifacts)
endif


.PHONY: release-notes
release-notes: ## greps UNRELEASED notes from the CHANGELOG
	@ awk '/## \[Unreleased\]/{flag=1;next}/## \[.*\] - /{flag=0}flag' CHANGELOG.md


.PHONY: help
help: ## this help
	@ awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m\t%s\n", $$1, $$2 }' $(MAKEFILE_LIST) | column -s$$'\t' -t

