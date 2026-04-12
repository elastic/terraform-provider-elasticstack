## Why

The `openspec-verify-label` workflow currently runs GitHub Agentic Workflows with a GitHub-hosted Copilot model selection baked into the workflow frontmatter. The repository now wants that verification path to keep using the Copilot engine while routing model inference through the Elastic OpenAI-compatible LiteLLM proxy at `https://elastic.litellm-prod.ai/v1`, with the model set to `llm-gateway/gpt-5.4`.

## What Changes

- Update the `openspec-verify-label` workflow contract so the verification job keeps `engine.id: copilot` while configuring Copilot CLI BYOK provider settings for an OpenAI-compatible LiteLLM endpoint.
- Change the workflow model selection from `gpt-5.4` to `llm-gateway/gpt-5.4`.
- Require the workflow network policy to allow the LiteLLM host in addition to the existing review-environment dependencies.
- Document the secret-backed provider-auth contract needed for the LiteLLM proxy without changing the workflow's GitHub-native review, archive, or safe-output behavior.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `ci-aw-openspec-verification`: Change the verification-engine requirements so `openspec-verify-label` keeps the Copilot engine but routes model traffic through the OpenAI-compatible LiteLLM endpoint `https://elastic.litellm-prod.ai/v1` using model `llm-gateway/gpt-5.4`.

## Impact

- Authored workflow source under `.github/workflows-src/openspec-verify-label/`
- Generated workflow artifacts under `.github/workflows/openspec-verify-label.*`
- Copilot engine environment and AWF network policy for the verification job
- OpenSpec requirements for `ci-aw-openspec-verification`
