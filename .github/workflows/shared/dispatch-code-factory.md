---
safe-outputs:
  jobs:
    dispatch-code-factory:
      needs: safe_outputs
      description: "Dispatch code-factory for each created issue"
      permissions:
        actions: write
        contents: read
      runs-on: ubuntu-latest
      steps:
        - name: Checkout repository
          uses: actions/checkout@v7.0.1
          with:
            persist-credentials: false
            sparse-checkout: .github/scripts/workflows/lib
            sparse-checkout-cone-mode: true
            fetch-depth: 1
        - name: Download safe-outputs artifact
          uses: actions/download-artifact@v8.0.1
          with:
            name: safe-outputs-items
            path: /tmp/gh-aw/safe-outputs
            if-no-files-found: warn
        - name: Dispatch code-factory runs
          env:
            GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
            GITHUB_REPOSITORY: ${{ github.repository }}
            GITHUB_WORKFLOW_NAME: ${{ github.workflow }}
          run: |
            SOURCE_WORKFLOW=$(echo "$GITHUB_WORKFLOW_NAME" | tr '[:upper:]' '[:lower:]' | tr ' ' '-')
            node .github/scripts/workflows/lib/producer-dispatch.js \
              /tmp/gh-aw/safe-outputs/temporary-id-map.json \
              "$SOURCE_WORKFLOW"
---

# Dispatch code-factory

Shared safe-outputs job that dispatches `code-factory-issue.lock.yml` once per issue
created in the current workflow run. `SOURCE_WORKFLOW` is derived from the calling
workflow display name (`github.workflow`).
