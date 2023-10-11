#!/bin/bash

set -euo pipefail

cleanup() {
    ARG=${?}
    echo  "--- Clean up"

    unset GPG_FINGERPRINT_SECRET
    # unset GITHUB_TOKEN
    rm -rf dist bin
    exit ${ARG}
}

trap cleanup EXIT

echo "--- Download dependencies"
make vendor

echo "--- Import gpg key"

GITHUB_ORGANIZATION=elastic
REPO_NAME=terraform-provider-elasticstack
VAULT_PATH=secret/ci/${GITHUB_ORGANIZATION}-${REPO_NAME}

GPG_PRIVATE_SECRET=$(vault read -field=gpg_private ${VAULT_PATH}  | base64 -d)

GPG_PASSPHRASE_SECRET=$(vault read -field=gpg_passphrase ${VAULT_PATH})

cat ${GPG_PASSPHRASE_SECRET} | gpg --import --batch --yes --passphrase-fd 0 ${GPG_PRIVATE_SECRET}

echo "--- Cache GPG key and release the binaries"

cat ${GPG_PASSPHRASE_SECRET} | gpg --armor --detach-sign --passphrase-fd 0 --pinentry-mode loopback

echo "--- Release the binaries"

# 'make release' calls 'goreleaser' that needs GPG_FINGERPRINT_SECRET and GITHUB_TOKEN env vars
export GPG_FINGERPRINT_SECRET=$(vault read -field=gpg_fingerprint ${VAULT_PATH} | xargs)

## TODO
## goreleaser needs GH token to publish binaries to GH
## it's commented out while the BK pipeline is being tested 
# export GITHUB_TOKEN=$(vault read -field=github_release_token ${VAULT_PATH} | xargs)

make release
