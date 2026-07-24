---
network:
  allowed:
    - node
    - go
    - terraform
steps:
  - name: Setup Go
    uses: actions/setup-go@v7.0.0
    with:
      go-version-file: go.mod
      cache: false
  - name: Setup Terraform CLI
    uses: hashicorp/setup-terraform@v4.0.1
    with:
      terraform_wrapper: false
  - name: Export Go and Terraform paths for AWF chroot mode
    run: |
      echo "GOROOT=$(go env GOROOT)" >> "$GITHUB_ENV"
      echo "GOPATH=$(go env GOPATH)" >> "$GITHUB_ENV"
      echo "GOMODCACHE=$(go env GOMODCACHE)" >> "$GITHUB_ENV"
      TERRAFORM_BIN=$(which terraform)
      # Copy terraform into the workspace so the AWF chroot container can see it
      # (the chroot mounts $GITHUB_WORKSPACE but not the RUNNER_TEMP path where
      # hashicorp/setup-terraform installs the binary).
      mkdir -p "$GITHUB_WORKSPACE/bin"
      cp "$TERRAFORM_BIN" "$GITHUB_WORKSPACE/bin/terraform"
      echo "TERRAFORM_BIN=$GITHUB_WORKSPACE/bin/terraform" >> "$GITHUB_ENV"
      echo "PATH=$GITHUB_WORKSPACE/bin:$PATH" >> "$GITHUB_ENV"
  - name: Setup Node.js
    uses: actions/setup-node@v7.0.0
    with:
      node-version-file: package.json
  - name: Setup repository dependencies
    run: make setup
---

# Dev environment setup

Shared setup component. Installs Go, Terraform CLI, and Node.js; exports chroot
paths (with the terraform binary staged into the workspace so the AWF sandbox
can discover it); and runs `make setup`.
