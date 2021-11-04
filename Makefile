.DEFAULT_GOAL = help
SHELL := /bin/bash

VERSION ?= 0.1.0

NAME = elasticstack
BINARY = terraform-provider-${NAME}
MARCH = "$$(go env GOOS)_$$(go env GOARCH)"

ACCTEST_PARALLELISM ?= 10
ACCTEST_TIMEOUT = 120m
ACCTEST_COUNT = 1
TEST ?= ./...

export GOBIN = $(shell pwd)/bin


$(GOBIN): ## create bin/ in the current directory
	mkdir -p $(GOBIN)


.PHONY: build
build: lint ## build the terraform provider
	go build -o ${BINARY}


.PHONY: testacc
testacc: lint ## Run acceptance tests
	TF_ACC=1 go test ./... -v -count $(ACCTEST_COUNT) -parallel $(ACCTEST_PARALLELISM) $(TESTARGS) -timeout $(ACCTEST_TIMEOUT)


.PHONY: test
test: lint ## Run unit tests
	go test $(TEST) $(TESTARGS) -timeout=5m -parallel=4


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


.PHONY: misspell
misspell:
	@ $(GOBIN)/misspell -error -source go ./internal/
	@ $(GOBIN)/misspell -error -source text ./templates/


.PHONY: golangci-lint
golangci-lint:
	@ $(GOBIN)/golangci-lint run --max-same-issues=0 --timeout=300s $(GOLANGCIFLAGS) ./internal/...


.PHONY: lint
lint: setup misspell golangci-lint ## Run lints to check the spelling and common go patterns


.PHONY: setup
setup: tools ## Setup the dev environment


.PHONY: help
help: ## this help
	@ awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m\t%s\n", $$1, $$2 }' $(MAKEFILE_LIST) | column -s$$'\t' -t


include .ci/Makefile.ci
