---
imports: [shared/setup-dev.md]
name: Security role docs drift
description: >-
  Detects drift between documented Kibana feature privileges and the live Kibana
  features API, then opens a self-healing PR with guide and JSON updates when needed.
on:
  workflow_dispatch:
  schedule:
    - cron: weekly on monday
  push:
    branches:
      - main
    paths:
      - 'generated/kbapi/**'
model: "llm-gateway/gpt-5.5"
engine:
  id: claude
  version: 2.1.98
  env:
    ANTHROPIC_BASE_URL: "https://elastic.litellm-prod.ai/"
    ANTHROPIC_API_KEY: ${{ secrets.CLAUDE_LITELLM_PROXY_API_KEY }}
permissions:
  contents: read
  issues: read
  pull-requests: read
  actions: read
tools:
  cli-proxy: true
  github:
    mode: gh-proxy
    toolsets: [repos, pull_requests]
  timeout: 300
safe-outputs:
  create-pull-request:
    labels: [automated-analysis, documentation]
    max: 1
    patch-format: am
  noop:
    max: 1
    report-as-issue: false
network:
  allowed: [defaults, node, go, terraform, elastic.litellm-prod.ai]
checkout:
  fetch-depth: 0
steps:
  - name: Checkout repository
    uses: actions/checkout@v7.0.1
    with:
      fetch-depth: 0
      persist-credentials: false
  - name: Setup Go
    uses: actions/setup-go@v7.0.0
    with:
      go-version-file: go.mod
      cache: false
  - name: Run security role docs pre-activation
    id: pre_activation
    run: |
      mkdir -p /tmp/gh-aw/agent
      go run ./scripts/security-role-docs pre-activation \
        --features-path scripts/security-role-docs/kibana-features.json \
        --report-path /tmp/gh-aw/agent/drift-report.json
---

# Security role docs drift worker

Pre-activation has computed drift against the live Kibana features API. Read the drift report at `/tmp/gh-aw/agent/drift-report.json` and open exactly one pull request when drift exists.

## Pre-activation context

- **Report path**: `/tmp/gh-aw/agent/drift-report.json`

## Task

1. Read `/tmp/gh-aw/agent/drift-report.json`.
2. Update `scripts/security-role-docs/kibana-features.json` so it reflects the live Kibana feature set from the report:
   - add unknown features to `documented` when they should appear in the guide
   - remove features from `documented` when they appear in `removed_features`
   - if a feature should stay undocumented, reviewers can move it from `documented` to `skip`
3. Update the Kibana feature privilege reference in `templates/guides/security-roles.md.tmpl` to match the feature changes.
4. Run `make docs-generate` so `docs/guides/security-roles.md` matches the updated template.
5. Create exactly one pull request with the resulting changes.

## Pull request requirements

- Explain the drift found using `unknown_features` and `removed_features` from the report.
- State that reviewers may move newly added features from `documented` to `skip` when the feature should remain excluded from the guide.
- Do not create issues.
- If the report contains no actionable drift after inspection, call `noop` instead of opening a pull request.