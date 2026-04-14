## MODIFIED Requirements

### Requirement: Schema-coverage rotation allows repository bootstrap ecosystems
The `schema-coverage-rotation` workflow SHALL declare an AWF network policy that allows the repository bootstrap path to use the default allowlist plus the Node and Go ecosystems, and SHALL allow `elastic.litellm-prod.ai` for the Claude engine's Anthropic-compatible proxy access.

#### Scenario: Workflow frontmatter allows required ecosystems
- **WHEN** maintainers inspect the schema-coverage rotation workflow frontmatter
- **THEN** `network.allowed` SHALL include `defaults`
- **AND** `network.allowed` SHALL include `node`
- **AND** `network.allowed` SHALL include `go`
- **AND** `network.allowed` SHALL include `elastic.litellm-prod.ai`
