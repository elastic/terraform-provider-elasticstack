#!/bin/bash
set -euo pipefail

# Serena MCP server bootstrap for Go projects.
#
# The upstream ghcr.io/github/serena-mcp-server:latest image ships with
# Go 1.25 and gopls v0.21.0, which cannot initialise against modules that
# declare go 1.26.1 (the version used by this repository).
#
# This script is executed as the Serena container entrypoint. It downloads
# and installs the Go version declared in go.mod, installs a compatible
# gopls, and then starts the Serena MCP server.

GO_MOD="${GITHUB_WORKSPACE:-/workspace}/go.mod"
GO_VERSION="1.26.1"
if [ -f "$GO_MOD" ]; then
    GO_VERSION=$(grep -E '^go [0-9]+\.[0-9]+(\.[0-9]+)?' "$GO_MOD" | awk '{print $2}')
    echo "Detected Go version from go.mod: ${GO_VERSION}"
fi

GO_INSTALL_DIR="/usr/local/go"
GOPATH="/tmp/go"

# ---------------------------------------------------------------------------
# 1. Install Go if missing or wrong version
# ---------------------------------------------------------------------------
install_go=0
if command -v go >/dev/null 2>&1; then
    CURRENT_GO=$(go version | awk '{print $3}')
    if [ "$CURRENT_GO" != "go${GO_VERSION}" ]; then
        echo "Go version mismatch: ${CURRENT_GO}, need go${GO_VERSION}. Re-installing..."
        install_go=1
    else
        echo "Go ${GO_VERSION} already present."
    fi
else
    echo "Go not found. Installing go${GO_VERSION}..."
    install_go=1
fi

if [ "$install_go" -eq 1 ]; then
    # python:3.11-slim does not include curl/wget, but Python is available.
    python3 -c "
import urllib.request
import sys
url = 'https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz'
print(f'Downloading {url} ...', file=sys.stderr)
urllib.request.urlretrieve(url, '/tmp/go.tar.gz')
print('Download complete.', file=sys.stderr)
"
    rm -rf "${GO_INSTALL_DIR}"
    tar -C /usr/local -xzf /tmp/go.tar.gz
    rm -f /tmp/go.tar.gz
fi

# ---------------------------------------------------------------------------
# 2. Export toolchain paths
# ---------------------------------------------------------------------------
export PATH="${GO_INSTALL_DIR}/bin:${PATH}"
export GOROOT="${GO_INSTALL_DIR}"
export GOPATH="${GOPATH}"
mkdir -p "${GOPATH}/bin"
export PATH="${GOPATH}/bin:${PATH}"

# ---------------------------------------------------------------------------
# 3. Install / verify gopls
# ---------------------------------------------------------------------------
if ! command -v gopls >/dev/null 2>&1; then
    echo "Installing gopls..."
    go install golang.org/x/tools/gopls@latest
fi

echo "Go toolchain ready:"
go version
gopls version || true

# ---------------------------------------------------------------------------
# 4. Start Serena MCP server
# ---------------------------------------------------------------------------
exec serena start-mcp-server --context claude-code --project "${GITHUB_WORKSPACE:-/workspace}"
