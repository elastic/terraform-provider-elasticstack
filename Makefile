.DEFAULT_GOAL = help
SHELL := /bin/bash


VERSION ?= 0.11.9

NAME = elasticstack
BINARY = terraform-provider-${NAME}
MARCH = "$$(go env GOOS)_$$(go env GOARCH)"

ACCTEST_PARALLELISM ?= 10
ACCTEST_TIMEOUT = 120m
ACCTEST_COUNT = 1
TEST ?= ./...
SWAGGER_VERSION ?= 8.7

GOVERSION ?= $(shell grep -e '^go' go.mod | cut -f 2 -d ' ')

STACK_VERSION ?= 8.15.2

ELASTICSEARCH_NAME ?= terraform-elasticstack-es
ELASTICSEARCH_ENDPOINTS ?= http://$(ELASTICSEARCH_NAME):9200
ELASTICSEARCH_USERNAME ?= elastic
ELASTICSEARCH_PASSWORD ?= password
ELASTICSEARCH_NETWORK ?= elasticstack-network
ELASTICSEARCH_MEM ?= 2048m

KIBANA_NAME ?= terraform-elasticstack-kb
KIBANA_ENDPOINT ?= http://$(KIBANA_NAME):5601
KIBANA_SYSTEM_USERNAME ?= kibana_system
KIBANA_SYSTEM_PASSWORD ?= password
KIBANA_API_KEY_NAME ?= kibana-api-key

FLEET_NAME ?= terraform-elasticstack-fleet
FLEET_ENDPOINT ?= https://$(FLEET_NAME):8220

SOURCE_LOCATION ?= $(shell pwd)
, := ,

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

# Retry command - first argument is how many attempts are required, second argument is the command to run
# Backoff starts with 1 second and double with next iteration
retry = until [ $$(if [ -z "$$attempt" ]; then echo -n "0"; else echo -n "$$attempt"; fi) -ge $(1) ]; do \
		backoff=$$(if [ -z "$$backoff" ]; then echo "1"; else echo "$$backoff"; fi); \
		sleep $$backoff; \
		$(2) && break; \
		attempt=$$((attempt + 1)); \
		backoff=$$((backoff * 2)); \
	done

# To run specific test (e.g. TestAccResourceActionConnector) execute `make docker-testacc TESTARGS='-run ^TestAccResourceActionConnector$$'`
# To enable tracing (or debugging), execute `make docker-testacc TF_LOG=TRACE`
.PHONY: docker-testacc
docker-testacc: docker-elasticsearch docker-kibana docker-fleet ## Run acceptance tests in the docker container
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

# To run specific test (e.g. TestAccResourceActionConnector) execute `make docker-testacc TESTARGS='-run ^TestAccResourceActionConnector$$'`
# To enable tracing (or debugging), execute `make docker-testacc TF_LOG=TRACE`
.PHONY: docker-testacc-with-token
docker-testacc-with-token:
	@ docker run --rm \
		-e ELASTICSEARCH_ENDPOINTS="$(ELASTICSEARCH_ENDPOINTS)" \
		-e KIBANA_ENDPOINT="$(KIBANA_ENDPOINT)" \
		-e ELASTICSEARCH_BEARER_TOKEN="$(ELASTICSEARCH_BEARER_TOKEN)" \
		-e KIBANA_USERNAME="$(ELASTICSEARCH_USERNAME)" \
		-e KIBANA_PASSWORD="$(ELASTICSEARCH_PASSWORD)" \
		-e TF_LOG="$(TF_LOG)" \
		--network $(ELASTICSEARCH_NETWORK) \
		-w "/provider" \
		-v "$(SOURCE_LOCATION):/provider" \
		golang:$(GOVERSION) make testacc TESTARGS="$(TESTARGS)"

.PHONY: docker-elasticsearch
docker-elasticsearch: docker-network ## Start Elasticsearch single node cluster in docker container
	@ docker rm -f $(ELASTICSEARCH_NAME) &> /dev/null || true
	@ $(call retry, 5, if ! docker ps --format '{{.Names}}' | grep -w $(ELASTICSEARCH_NAME) > /dev/null 2>&1 ; then \
		docker run -d \
		--memory $(ELASTICSEARCH_MEM) \
		-p 9200:9200 -p 9300:9300 \
		-e "discovery.type=single-node" \
		-e "xpack.security.enabled=true" \
		-e "xpack.security.authc.api_key.enabled=true" \
		-e "xpack.security.authc.token.enabled=true" \
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
	@ docker rm -f $(KIBANA_NAME)  &> /dev/null || true
	@ $(call retry, 5, if ! docker ps --format '{{.Names}}' | grep -w $(KIBANA_NAME) > /dev/null 2>&1 ; then \
		docker run -d \
		-p 5601:5601 \
		-e SERVER_NAME=kibana \
		-e ELASTICSEARCH_HOSTS=$(ELASTICSEARCH_ENDPOINTS) \
		-e ELASTICSEARCH_USERNAME=$(KIBANA_SYSTEM_USERNAME) \
		-e ELASTICSEARCH_PASSWORD=$(KIBANA_SYSTEM_PASSWORD) \
		-e XPACK_ENCRYPTEDSAVEDOBJECTS_ENCRYPTIONKEY=a7a6311933d3503b89bc2dbc36572c33a6c10925682e591bffcab6911c06786d \
		-e LOGGING_ROOT_LEVEL=debug \
		--name $(KIBANA_NAME) \
		--network $(ELASTICSEARCH_NETWORK) \
		docker.elastic.co/kibana/kibana:$(STACK_VERSION); \
		fi)

.PHONY: docker-kibana-with-tls
docker-kibana-with-tls: docker-network docker-elasticsearch set-kibana-password
	@ docker rm -f $(KIBANA_NAME)  &> /dev/null || true
	@ mkdir -p certs
	@ CAROOT=certs mkcert localhost $(KIBANA_NAME)
	@ mv localhost*.pem certs/

	@ $(call retry, 5, if ! docker ps --format '{{.Names}}' | grep -w $(KIBANA_NAME) > /dev/null 2>&1 ; then \
		docker run -d \
		-p 5601:5601 \
		-v $(shell pwd)/certs:/certs \
		-e SERVER_NAME=kibana \
		-e ELASTICSEARCH_HOSTS=$(ELASTICSEARCH_ENDPOINTS) \
		-e ELASTICSEARCH_USERNAME=$(KIBANA_SYSTEM_USERNAME) \
		-e ELASTICSEARCH_PASSWORD=$(KIBANA_SYSTEM_PASSWORD) \
		-e XPACK_ENCRYPTEDSAVEDOBJECTS_ENCRYPTIONKEY=a7a6311933d3503b89bc2dbc36572c33a6c10925682e591bffcab6911c06786d \
		-e SERVER_SSL_CERTIFICATE=/certs/localhost+1.pem \
		-e SERVER_SSL_KEY=/certs/localhost+1-key.pem \
		-e SERVER_SSL_ENABLED=true \
		-e LOGGING_ROOT_LEVEL=debug \
		--name $(KIBANA_NAME) \
		--network $(ELASTICSEARCH_NETWORK) \
		docker.elastic.co/kibana/kibana:$(STACK_VERSION); \
		fi)

.PHONY: docker-fleet
docker-fleet: docker-network docker-elasticsearch docker-kibana setup-kibana-fleet ## Start Fleet node in docker container
	@ docker rm -f $(FLEET_NAME)  &> /dev/null || true
	@ $(call retry, 5, if ! docker ps --format '{{.Names}}' | grep -w $(FLEET_NAME) > /dev/null 2>&1 ; then \
		docker run -d \
		-p 8220:8220 \
		-e SERVER_NAME=fleet \
      	-e FLEET_ENROLL=1 \
      	-e FLEET_URL=$(FLEET_ENDPOINT) \
      	-e FLEET_INSECURE=true \
      	-e FLEET_SERVER_ENABLE=1 \
      	-e FLEET_SERVER_POLICY_ID=fleet-server \
      	-e FLEET_SERVER_ELASTICSEARCH_HOST=$(ELASTICSEARCH_ENDPOINTS) \
      	-e FLEET_SERVER_ELASTICSEARCH_INSECURE=true \
      	-e FLEET_SERVER_INSECURE_HTTP=true \
      	-e KIBANA_HOST=$(KIBANA_ENDPOINT) \
      	-e KIBANA_FLEET_SETUP=1 \
      	-e KIBANA_FLEET_USERNAME=$(ELASTICSEARCH_USERNAME) \
      	-e KIBANA_FLEET_PASSWORD=$(ELASTICSEARCH_PASSWORD) \
		--name $(FLEET_NAME) \
		--network $(ELASTICSEARCH_NETWORK) \
		docker.elastic.co/beats/elastic-agent:$(STACK_VERSION); \
		fi)


.PHONY: docker-network
docker-network: ## Create a dedicated network for ES and test runs
	@ if ! docker network ls --format '{{.Name}}' | grep -w $(ELASTICSEARCH_NETWORK) > /dev/null 2>&1 ; then \
		docker network create $(ELASTICSEARCH_NETWORK); \
		fi

.PHONY: set-kibana-password
set-kibana-password: ## Sets the ES KIBANA_SYSTEM_USERNAME's password to KIBANA_SYSTEM_PASSWORD. This expects Elasticsearch to be available at localhost:9200
	@ $(call retry, 10, curl -sS -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json" http://localhost:9200/_security/user/$(KIBANA_SYSTEM_USERNAME)/_password -d '{"password":"$(KIBANA_SYSTEM_PASSWORD)"}' | grep -q "^{}")

.PHONY: create-es-api-key
create-es-api-key: ## Creates and outputs a new API Key. This expects Elasticsearch to be available at localhost:9200
	@ $(call retry, 10, curl -sS -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json" http://localhost:9200/_security/api_key -d '{"name":"$(KIBANA_API_KEY_NAME)"}')

.PHONY: create-es-bearer-token
create-es-bearer-token: ## Creates and outputs a new OAuth bearer token. This expects Elasticsearch to be available at localhost:9200
	@ $(call retry, 10, curl -sS -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json" http://localhost:9200/_security/oauth2/token -d '{"grant_type":"client_credentials"}')

.PHONY: setup-kibana-fleet
setup-kibana-fleet: ## Creates the agent and integration policies required to run Fleet. This expects Kibana to be available at localhost:5601
	@ $(call retry, 10, curl -sS --fail-with-body -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json" -H "kbn-xsrf: true" http://localhost:5601/api/fleet/fleet_server_hosts -d '{"name":"default"$(,)"host_urls":["$(FLEET_ENDPOINT)"]$(,)"is_default":true}')
	@ $(call retry, 10, curl -sS --fail-with-body -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json" -H "kbn-xsrf: true" http://localhost:5601/api/fleet/agent_policies -d '{"id":"fleet-server"$(,)"name":"Fleet Server"$(,)"namespace":"default"$(,)"monitoring_enabled":["logs"$(,)"metrics"]}')
	@ $(call retry, 10, curl -sS --fail-with-body -X POST -u $(ELASTICSEARCH_USERNAME):$(ELASTICSEARCH_PASSWORD) -H "Content-Type: application/json" -H "kbn-xsrf: true" http://localhost:5601/api/fleet/package_policies -d '{"name":"fleet-server"$(,)"namespace":"default"$(,)"policy_id":"fleet-server"$(,)"enabled":true$(,)"inputs":[{"type":"fleet-server"$(,)"enabled":true$(,)"streams":[]$(,)"vars":{}}]$(,)"package":{"name":"fleet_server"$(,)"version":"1.5.0"}}')

.PHONY: docker-clean
docker-clean: ## Try to remove provisioned nodes and assigned network
	@ docker rm -f $(ELASTICSEARCH_NAME) $(KIBANA_NAME) $(FLEET_NAME) || true
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
	@ cd tools && go install github.com/goreleaser/goreleaser/v2
	@ cd tools && go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	@ cd tools && go install go.uber.org/mock/mockgen

.PHONY: misspell
misspell:
	@ $(GOBIN)/misspell -error -source go ./internal/
	@ $(GOBIN)/misspell -error -source text ./templates/


.PHONY: golangci-lint
golangci-lint:
	@ $(GOBIN)/golangci-lint run --max-same-issues=0 $(GOLANGCIFLAGS) ./internal/...


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
	@ $(GOBIN)/goreleaser release --snapshot --clean


.PHONY: release-no-publish
release-no-publish: tools check-sign-release ## Make a release without publishing artifacts
	@ $(GOBIN)/goreleaser release --skip=publish,announce,validate  --parallelism=2


.PHONY: release
release: tools check-sign-release check-publish-release ## Build, sign, and upload your release
	@ $(GOBIN)/goreleaser release --clean  --parallelism=4


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

.PHONY: generate-data-views-client
generate-data-views-client: ## generate Kibana data-views client
	@ docker run --rm -v "${PWD}:/local" openapitools/openapi-generator-cli:v7.0.1 generate \
		-i /local/generated/data_views/bundled.yaml \
		--skip-validate-spec \
		--git-repo-id terraform-provider-elasticstack \
		--git-user-id elastic \
		-p isGoSubmodule=true \
		-p packageName=data_views \
		-p generateInterfaces=true \
		-g go \
		-o /local/generated/data_views
	@ rm -rf generated/data_views/go.mod generated/data_views/go.sum generated/data_views/test
	@ go fmt ./generated/data_views/...

.PHONY: generate-connectors-client
generate-connectors-client: tools ## generate Kibana connectors client
	@ cd tools && go generate
	@ go fmt ./generated/connectors/...

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
	@ go fmt ./generated/...

.PHONY: generate-clients
generate-clients: generate-alerting-client generate-slo-client generate-data-views-client generate-connectors-client ## generate all clients
