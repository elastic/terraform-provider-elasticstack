.DEFAULT_GOAL = help
SHELL := /bin/bash

VERSION ?= 0.12.2

NAME = elasticstack
BINARY = terraform-provider-${NAME}
MARCH = "$$(go env GOOS)_$$(go env GOARCH)"

ACCTEST_PARALLELISM ?= 10
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

KIBANA_SYSTEM_USERNAME ?= kibana_system
KIBANA_SYSTEM_PASSWORD ?= password
KIBANA_API_KEY_NAME ?= kibana-api-key

FLEET_NAME ?= terraform-elasticstack-fleet
FLEET_ENDPOINT ?= https://$(FLEET_NAME):8220

export GOBIN = $(shell pwd)/bin


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
	@ ELASTICSEARCH_ENDPOINTS=http://localhost:9200 KIBANA_ENDPOINT=http://localhost:5601 ELASTICSEARCH_USERNAME=$(ELASTICSEARCH_USERNAME) ELASTICSEARCH_PASSWORD=$(ELASTICSEARCH_PASSWORD) make testacc

.PHONY: testacc
testacc: ## Run acceptance tests
	TF_ACC=1 go tool gotestsum --format testname --rerun-fails=3 --packages="-v ./..." -- -count $(ACCTEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

.PHONY: test
test: ## Run unit tests
	go test -v $(TEST) $(TESTARGS) -timeout=5m -parallel=4

CURL_OPTS = -sS --retry 5 --retry-all-errors -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json"

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
docker-elasticsearch: ## Start Elasticsearch single node cluster in docker container
	@ docker compose -f $(COMPOSE_FILE) up --quiet-pull -d elasticsearch

.PHONY: docker-kibana
docker-kibana:  ## Start Kibana node in docker container
	@ docker compose -f $(COMPOSE_FILE) up --quiet-pull -d kibana

.PHONY: docker-fleet
docker-fleet: ## Start Fleet node in docker container
	@ docker compose -f $(COMPOSE_FILE) up --quiet-pull -d fleet

.PHONY: set-kibana-password
set-kibana-password: ## Sets the ES KIBANA_SYSTEM_USERNAME's password to KIBANA_SYSTEM_PASSWORD. This expects Elasticsearch to be available at localhost:9200
	@ curl $(CURL_OPTS) http://localhost:9200/_security/user/$(KIBANA_SYSTEM_USERNAME)/_password -d '{"password":"$(KIBANA_SYSTEM_PASSWORD)"}'

.PHONY: create-es-api-key
create-es-api-key: ## Creates and outputs a new API Key. This expects Elasticsearch to be available at localhost:9200
	@ curl $(CURL_OPTS) http://localhost:9200/_security/api_key -d '{"name":"$(KIBANA_API_KEY_NAME)"}'

.PHONY: create-es-bearer-token
create-es-bearer-token: ## Creates and outputs a new OAuth bearer token. This expects Elasticsearch to be available at localhost:9200
	@ curl $(CURL_OPTS) http://localhost:9200/_security/oauth2/token -d '{"grant_type":"client_credentials"}'

.PHONY: setup-kibana-fleet
setup-kibana-fleet: ## Creates the agent and integration policies required to run Fleet. This expects Kibana to be available at localhost:5601
	curl $(CURL_OPTS) -H "kbn-xsrf: true" http://localhost:5601/api/fleet/fleet_server_hosts -d '{"name":"default","host_urls":["$(FLEET_ENDPOINT)"],"is_default":true}'
	curl $(CURL_OPTS) -H "kbn-xsrf: true" http://localhost:5601/api/fleet/agent_policies -d '{"id":"fleet-server","name":"Fleet Server","namespace":"default","monitoring_enabled":["logs","metrics"]}'
	curl $(CURL_OPTS) -H "kbn-xsrf: true" http://localhost:5601/api/fleet/package_policies -d '{"name":"fleet-server","namespace":"default","policy_id":"fleet-server","enabled":true,"inputs":[{"type":"fleet-server","enabled":true,"streams":[],"vars":{}}],"package":{"name":"fleet_server","version":"1.5.0"}}'

.PHONY: docker-clean
docker-clean: ## Try to remove provisioned nodes and assigned network
	@ docker compose -f $(COMPOSE_FILE) down -v

.PHONY: copy-kibana-ca
copy-kibana-ca: ## Copy Kibana CA certificate to local machine
	@ docker compose -f $(COMPOSE_FILE) cp kibana:/certs/rootCA.pem ./kibana-ca.pem

.PHONY: docs-generate
docs-generate: tools ## Generate documentation for the provider
	@ go tool github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name terraform-provider-elasticstack


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
	@[[ -f $(GOBIN)/golangci-lint ]] || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v2.6.2

.PHONY: golangci-lint
golangci-lint:
	@ $(GOBIN)/golangci-lint run --max-same-issues=0 $(GOLANGCIFLAGS) ./internal/...


.PHONY: lint
lint: setup golangci-lint fmt docs-generate ## Run lints to check the spelling and common go patterns

.PHONY: check-lint
check-lint: setup golangci-lint check-fmt check-docs

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
setup: tools vendor ## Setup the dev environment

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

.PHONY: generate-alerting-client
generate-alerting-client: ## generate Kibana alerting client
	@ docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli:v7.0.1 generate \
		-i /local/generated/alerting/bundled.yaml \
		--skip-validate-spec \
		--git-repo-id terraform-provider-elasticstack \
		--git-user-id elastic \
		-p isGoSubmodule=true \
		-p packageName=alerting \
		-p generateInterfaces=true \
		-g go \
		-o /local/generated/alerting
	@ rm -rf generated/alerting/go.mod generated/alerting/go.sum generated/alerting/test
	@ go fmt ./generated/alerting/...

.PHONY: generate-slo-client
generate-slo-client: tools ## generate Kibana slo client
	@ rm -rf generated/slo
	@ docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli:v7.0.1 generate \
		-i /local/generated/slo-spec.yml \
		--git-repo-id terraform-provider-elasticstack \
		--git-user-id elastic \
		-p isGoSubmodule=true \
		-p packageName=slo \
		-p generateInterfaces=true \
		-p useOneOfDiscriminatorLookup=true \
		-g go \
		-o /local/generated/slo \
		 --type-mappings=float32=float64
	@ rm -rf generated/slo/go.mod generated/slo/go.sum generated/slo/test
	@ go fmt ./generated/slo/...

.PHONY: generate-clients
generate-clients: generate-alerting-client generate-slo-client ## generate all clients
