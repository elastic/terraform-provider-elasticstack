## Why

The `openspec-verify-label` workflow now routes verification through the `claude` engine in GitHub Agentic Workflows. The OpenSpec proposal should reflect the implemented contract: Claude-compatible model traffic is sent through the Elastic LiteLLM proxy at `https://elastic.litellm-prod.ai`, with the model set to `llm-gateway/gpt-5.4`.

## What Changes

- Update the `openspec-verify-label` workflow contract so the verification job uses `engine.id: claude` while configuring Anthropic-compatible LiteLLM endpoint settings for Claude traffic.
- Change the workflow model selection from `gpt-5.4` to `llm-gateway/gpt-5.4`.
- Require the workflow network policy to allow the LiteLLM host in addition to the existing review-environment dependencies.
- Document the secret-backed Claude auth contract needed for the LiteLLM proxy without changing the workflow's GitHub-native review, archive, or safe-output behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci-aw-openspec-verification`: Change the verification-engine requirements so `openspec-verify-label` uses the `claude` engine and routes Anthropic-compatible model traffic through the LiteLLM endpoint `https://elastic.litellm-prod.ai` using model `llm-gateway/gpt-5.4`.

## Impact

- Authored workflow source under `.github/workflows-src/openspec-verify-label/`
- Generated workflow artifacts under `.github/workflows/openspec-verify-label.*`
- Claude engine environment and AWF network policy for the verification job
- OpenSpec requirements for `ci-aw-openspec-verification`
