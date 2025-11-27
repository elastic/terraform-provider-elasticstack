#!/bin/bash

set -euo pipefail

echo "--- Download dependencies"
make vendor

echo "--- Importing GPG key"
echo -n "$GPG_PRIVATE_SECRET" | base64 --decode | gpg --import --batch --yes --passphrase "$GPG_PASSPHRASE_SECRET"

echo "--- Caching GPG passphrase"
echo "$GPG_PASSPHRASE_SECRET" | gpg --armor --detach-sign --passphrase-fd 0 --pinentry-mode loopback

echo "--- Release the binaries"

if [[ -n "${SNAPSHOT}" ]]; then
  make release-snapshot
  echo "--- Uploading snapshot artifacts"
  ls -alh dist/
else
  echo "Would do a real release"
fi

