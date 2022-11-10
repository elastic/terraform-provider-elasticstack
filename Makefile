.DEFAULT_GOAL = help
SHELL := /bin/bash

VERSION ?= 0.4.0

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
