SHELL := /bin/bash

SOURCE_LOCATION ?= $(shell pwd)

-include $(SOURCE_LOCATION)/.env

vault :=$(SOURCE_LOCATION)/scripts/retry.sh 5 vault

# BUILD_ID is present during run on Jenkins machine, but not on dev box, hence using it here to distinguish between those cases
ifndef VAULT_TOKEN
	ifdef BUILD_ID
		VAULT_TOKEN = $(shell $(vault) write -address=$(VAULT_ADDR) -field=token auth/approle/login role_id=$(VAULT_ROLE_ID) secret_id=$(VAULT_SECRET_ID))
	else
		VAULT_TOKEN = $(shell $(vault) write -address=$(VAULT_ADDR) -field=token auth/github/login token=$(GITHUB_TOKEN))
	endif
endif

.PHONY: clean-secrets
clean-secrets: ## Delete credentials
	@ rm -f .apikey .gpg_private .gpg_passphrase .github_token .gpg_fingerprint

.apikey:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -field=apikey secret/devops-ci/terraform-provider-ec > .apikey

.gpg_private:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -field=gpg_private secret/devops-ci/terraform-provider-ec  | base64 -d > .gpg_private

.gpg_passphrase:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -field=gpg_passphrase secret/devops-ci/terraform-provider-ec > .gpg_passphrase

.gpg_fingerprint:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -field=gpg_fingerprint secret/devops-ci/terraform-provider-ec > .gpg_fingerprint

.github_token:
	@ VAULT_TOKEN=$(VAULT_TOKEN) $(vault) read -field=gh_personal_access_token secret/devops-ci/terraform-provider-ec > .github_token

import-gpg-key:
	@ cat .gpg_passphrase | gpg --import --batch --yes --passphrase-fd 0 .gpg_private

cache-gpg-passphrase:
	@ cat .gpg_passphrase | gpg --armor --detach-sign --passphrase-fd 0 --pinentry-mode loopback
