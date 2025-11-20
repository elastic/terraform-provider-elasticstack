.DEFAULT_GOAL = help
SHELL := /bin/bash
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# POC-only Streams client:
# - github_repo/github_ref are intentionally configurable so the POC can track a
#   Kibana branch that exposes the Streams APIs and OAS.
github_repo ?= rStelmach/kibana
github_ref  ?= c72d4e73064dcabdcdc24ac1cae64c7f594378ed
oas_url     := https://raw.githubusercontent.com/$(github_repo)/$(github_ref)/oas_docs/output/kibana.yaml

.PHONY: all
all: clean download transform generate ## Fetch, filter, and generate the Streams API client

.PHONY: download
download: oas.yaml ## Download the remote schema

oas.yaml:
	curl -sSfo oas.yaml "$(oas_url)"

.PHONY: transform
transform: download ## Filter schema down to Streams-only paths
	go run $(ROOT_DIR)/transform_schema.go -i ./oas.yaml -o ./oas-streams.yaml

.PHONY: generate
generate: ## Generate the Streams API client
	go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config oapi-config.yaml ./oas-streams.yaml

.PHONY: clean
clean: ## Remove any downloaded or transformed OAS files
	rm -rf oas.yaml oas-streams.yaml

.PHONY: help
help: ## Show this help
	@ awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m\t%s\n", $$1, $$2 }' $(MAKEFILE_LIST) | column -s$$'\t' -t


