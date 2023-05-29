.DEFAULT_GOAL = help
SHELL := /bin/bash

VERSION ?= 0.6.0

NAME = elasticstack
BINARY = terraform-provider-${NAME}
MARCH = "$$(go env GOOS)_$$(go env GOARCH)"

ACCTEST_PARALLELISM ?= 10
ACCTEST_TIMEOUT = 120m
ACCTEST_COUNT = 1
TEST ?= ./...
SWAGGER_VERSION ?= 8.7

GOVERSION ?= 1.19

STACK_VERSION ?= 8.6.0

ELASTICSEARCH_NAME ?= terraform-elasticstack-es
ELASTICSEARCH_ENDPOINTS ?= http://$(ELASTICSEARCH_NAME):9200
ELASTICSEARCH_USERNAME ?= elastic
ELASTICSEARCH_PASSWORD ?= password
ELASTICSEARCH_NETWORK ?= elasticstack-network
ELASTICSEARCH_MEM ?= 1024m

KIBANA_NAME ?= terraform-elasticstack-kb
KIBANA_ENDPOINT ?= http://$(KIBANA_NAME):5601
KIBANA_SYSTEM_USERNAME ?= kibana_system
KIBANA_SYSTEM_PASSWORD ?= password

SOURCE_LOCATION ?= $(shell pwd)

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


.PHONY: testacc
testacc: ## Run acceptance tests
	TF_ACC=1 go test -v ./... -count $(ACCTEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)

.PHONY: test
test: ## Run unit tests
	go test -v $(TEST) $(TESTARGS) -timeout=5m -parallel=4

# Retry command - first argumment is how many attempts are required, second argument is the command to run
# Backoff starts with 1 second and double with next itteration
retry = until [ $$(if [ -z "$$attempt" ]; then echo -n "0"; else echo -n "$$attempt"; fi) -ge $(1) ]; do \
		backoff=$$(if [ -z "$$backoff" ]; then echo "1"; else echo "$$backoff"; fi); \
		sleep $$backoff; \
		$(2) && break; \
		attempt=$$((attempt + 1)); \
		backoff=$$((backoff * 2)); \
	done

# To run specific test (e.g. TestAccResourceActionConnector) execute `make docker-testacc TESTARGS='-run ^TestAccResourceActionConnector$$'`
# To enable tracing (or debugging), execute `make docker-testacc TFLOG=TRACE`
.PHONY: docker-testacc
docker-testacc: docker-elasticsearch docker-kibana ## Run acceptance tests in the docker container
	@ docker run --rm \
		-e ELASTICSEARCH_ENDPOINTS="$(ELASTICSEARCH_ENDPOINTS)" \
		-e KIBANA_ENDPOINT="$(KIBANA_ENDPOINT)" \
		-e ELASTICSEARCH_USERNAME="$(ELASTICSEARCH_USERNAME)" \
		-e ELASTICSEARCH_PASSWORD="$(ELASTICSEARCH_PASSWORD)" \
		-e TF_LOG="$(TF_LOG)" \
		--network $(ELASTICSEARCH_NETWORK) \
		-w "/provider" \
		-v "$(SOURCE_LOCATION):/provider" \
		golang:$(GOVERSION) make testacc TESTARGS="$(TESTARGS)"

.PHONY: docker-elasticsearch
docker-elasticsearch: docker-network ## Start Elasticsearch single node cluster in docker container
	@ $(call retry, 5, if ! docker ps --format '{{.Names}}' | grep -w $(ELASTICSEARCH_NAME) > /dev/null 2>&1 ; then \
		docker run -d \
		--memory $(ELASTICSEARCH_MEM) \
		-p 9200:9200 -p 9300:9300 \
		-e "discovery.type=single-node" \
		-e "xpack.security.enabled=true" \
		-e "xpack.security.authc.api_key.enabled=true" \
		-e "xpack.watcher.enabled=true" \
		-e "xpack.license.self_generated.type=trial" \
		-e "repositories.url.allowed_urls=https://example.com/*" \
		-e "path.repo=/tmp" \
		-e ELASTIC_PASSWORD=$(ELASTICSEARCH_PASSWORD) \
		--name $(ELASTICSEARCH_NAME) \
		--network $(ELASTICSEARCH_NETWORK) \
		docker.elastic.co/elasticsearch/elasticsearch:$(STACK_VERSION); \
		fi)

.PHONY: docker-kibana
docker-kibana: docker-network docker-elasticsearch set-kibana-password ## Start Kibana node in docker container
	@ $(call retry, 5, if ! docker ps --format '{{.Names}}' | grep -w $(KIBANA_NAME) > /dev/null 2>&1 ; then \
		docker run -d \
		-p 5601:5601 \
		-e SERVER_NAME=kibana \
		-e ELASTICSEARCH_HOSTS=$(ELASTICSEARCH_ENDPOINTS) \
		-e ELASTICSEARCH_USERNAME=$(KIBANA_SYSTEM_USERNAME) \
		-e ELASTICSEARCH_PASSWORD=$(KIBANA_SYSTEM_PASSWORD) \
		-e XPACK_ENCRYPTEDSAVEDOBJECTS_ENCRYPTIONKEY=a7a6311933d3503b89bc2dbc36572c33a6c10925682e591bffcab6911c06786d \
		-e "logging.root.level=debug" \
		--name $(KIBANA_NAME) \
		--network $(ELASTICSEARCH_NETWORK) \
		docker.elastic.co/kibana/kibana:$(STACK_VERSION); \
		fi)

.PHONY: docker-network
docker-network: ## Create a dedicated network for ES and test runs
	@ if ! docker network ls --format '{{.Name}}' | grep -w $(ELASTICSEARCH_NETWORK) > /dev/null 2>&1 ; then \
		docker network create $(ELASTICSEARCH_NETWORK); \
		fi

.PHONY: set-kibana-password
set-kibana-password: ## Sets the ES KIBANA_SYSTEM_USERNAME's password to KIBANA_SYSTEM_PASSWORD. This expects Elasticsearch to be available at localhost:9200
	@ $(call retry, 10, curl -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json" http://localhost:9200/_security/user/$(KIBANA_SYSTEM_USERNAME)/_password -d "{\"password\":\"$(KIBANA_SYSTEM_PASSWORD)\"}" | grep -q "^{}")

.PHONY: docker-clean
docker-clean: ## Try to remove provisioned nodes and assigned network
	@ docker rm -f $(ELASTICSEARCH_NAME) $(KIBANA_NAME) || true
	@ docker network rm $(ELASTICSEARCH_NETWORK) || true


.PHONY: docs-generate
docs-generate: tools ## Generate documentation for the provider
	@ $(GOBIN)/tfplugindocs


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
tools: $(GOBIN) ## Install useful tools for linting, docs generation and development
	@ cd tools && go install github.com/client9/misspell/cmd/misspell
	@ cd tools && go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
	@ cd tools && go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@ cd tools && go install github.com/goreleaser/goreleaser
	@ cd tools && go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen

.PHONY: misspell
misspell:
	@ $(GOBIN)/misspell -error -source go ./internal/
	@ $(GOBIN)/misspell -error -source text ./templates/


.PHONY: golangci-lint
golangci-lint:
	@ $(GOBIN)/golangci-lint run --max-same-issues=0 --timeout=300s $(GOLANGCIFLAGS) ./internal/...


.PHONY: lint
lint: setup misspell golangci-lint check-fmt check-docs ## Run lints to check the spelling and common go patterns

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
setup: tools ## Setup the dev environment


.PHONY: release-snapshot
release-snapshot: tools ## Make local-only test release to see if it works using "release" command
	@ $(GOBIN)/goreleaser release --snapshot --rm-dist


.PHONY: release-no-publish
release-no-publish: tools check-sign-release ## Make a release without publishing artifacts
	@ $(GOBIN)/goreleaser release --skip-publish --skip-announce --skip-validate


.PHONY: release
release: tools check-sign-release check-publish-release ## Build, sign, and upload your release
	@ $(GOBIN)/goreleaser release --rm-dist


.PHONY: check-sign-release
check-sign-release:
ifndef GPG_FINGERPRINT
	$(error GPG_FINGERPRINT is undefined, but required for signing the release)
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
	@ docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
		-i https://raw.githubusercontent.com/elastic/kibana/$(SWAGGER_VERSION)/x-pack/plugins/alerting/docs/openapi/bundled.json \
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

.PHONY: generate-connectors-client
generate-connectors-client: tools ## generate Kibana connectors client
	@ cd tools && go generate
	@ go fmt ./generated/connectors/...

.PHONY: generate-slo-client
generate-slo-client: tools ## generate Kibana slo client
	@ docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli generate \
		-i /local/generated/slo-spec-test.yml \
		--skip-validate-spec \
		--git-repo-id terraform-provider-elasticstack \
		--git-user-id elastic \
		-p isGoSubmodule=true \
		-p packageName=slo \
		-p generateInterfaces=true \
		-g go \
		-o /local/generated/slo
	@ rm -rf generated/slo/go.mod generated/slo/go.sum generated/slo/test
	@ go fmt ./generated/...

.PHONY: generate-clients
generate-clients: generate-alerting-client generate-slo-client generate-connectors-client ## generate all clients
