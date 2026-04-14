## ADDED Requirements

### Requirement: Schema-coverage rotation uses the Claude engine through the Anthropic-compatible proxy
The `schema-coverage-rotation` workflow SHALL set `engine.id` to `claude` and SHALL route Claude traffic through `ANTHROPIC_BASE_URL` set to `https://elastic.litellm-prod.ai/`. Any configured `ANTHROPIC_API_KEY` value SHALL be sourced from a GitHub Actions secret-backed expression rather than from a checked-in literal.

#### Scenario: Authored workflow selects Claude
- **WHEN** maintainers inspect the authored `schema-coverage-rotation` workflow source
- **THEN** `engine.id` SHALL be `claude`

#### Scenario: Claude traffic is routed through the Elastic LiteLLM endpoint
- **WHEN** maintainers inspect the workflow engine environment
- **THEN** `ANTHROPIC_BASE_URL` SHALL be `https://elastic.litellm-prod.ai/`

#### Scenario: Anthropic authentication is secret-backed
- **WHEN** maintainers inspect the authored workflow source
- **THEN** any configured `ANTHROPIC_API_KEY` value SHALL come from a GitHub Actions secret expression rather than a literal API key value committed to the repository

### Requirement: Schema-coverage rotation sets an explicit Claude execution budget
The `schema-coverage-rotation` workflow SHALL set an explicit per-tool execution timeout of 300 seconds so the repository-local bootstrap and schema-analysis commands are not constrained by the Claude engine's shorter default tool-call timeout.

#### Scenario: Workflow frontmatter defines the Claude tool timeout
- **WHEN** maintainers inspect the schema-coverage rotation workflow frontmatter
- **THEN** `tools.timeout` SHALL be `300`
