---
steps:
  - name: Setup Go
    uses: actions/setup-go@v6
    with:
      go-version-file: go.mod
      cache: false
  - name: Setup Terraform CLI
    uses: hashicorp/setup-terraform@v4
    with:
      terraform_wrapper: false
  - name: Export Go and Terraform paths for AWF chroot mode
    run: |
      echo "GOROOT=$(go env GOROOT)" >> "$GITHUB_ENV"
      echo "GOPATH=$(go env GOPATH)" >> "$GITHUB_ENV"
      echo "GOMODCACHE=$(go env GOMODCACHE)" >> "$GITHUB_ENV"
      if [ -x "$(which terraform)" ]; then
        echo "TERRAFORM_BIN=$(which terraform)" >> "$GITHUB_ENV"
        TERRAFORM_DIR=$(dirname "$(which terraform)")
        echo "PATH=${TERRAFORM_DIR}:${PATH}" >> "$GITHUB_ENV"
      fi
  - name: Setup Node.js
    uses: actions/setup-node@v6
    with:
      node-version-file: package.json
  - name: Setup repository dependencies
    run: make setup
---

# Dev environment setup

Shared setup component. Installs Go, Terraform CLI, and Node.js; exports chroot paths; and runs `make setup`.
