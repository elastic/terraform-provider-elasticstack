.DEFAULT_GOAL = help
SHELL := /bin/bash
ROOT_DIR := $(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

# 2024-11-19
# GITHUB_REF ?= refs/heads/main
GITHUB_REF ?= 3757e641278a5186919e35a0f980ac3cda7e8ccd
OAS_URL    := https://raw.githubusercontent.com/elastic/kibana/$(GITHUB_REF)/oas_docs/output/kibana.yaml

.PHONY: all
all: download transform generate ## Fetch, bundle, transform, and generate the API
	

.PHONY: download
download: oas.yaml ## Download the remote schema

oas.yaml: 
	curl -sSfo oas.yaml "$(OAS_URL)"

.PHONY: transform
transform: download ## Transform and filter the schema
	go run $(ROOT_DIR)/transform_schema.go -i ./oas.yaml -o ./oas-filtered.yaml

.PHONY: generate
generate: ## Generate the API
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 -config oapi-config.yaml ./oas-filtered.yaml


.PHONY: clean
clean: ## Remove any downloaded files
	rm -rf oas.yaml oas-filtered.yaml


.PHONY: help
help: ## Show this help
	@ awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m\t%s\n", $$1, $$2 }' $(MAKEFILE_LIST) | column -s$$'\t' -t
