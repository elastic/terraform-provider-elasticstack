## ADDED Requirements

### Requirement: Workflow uses AM patch transport for safe-output PR creation
The authored `change-factory` issue-intake workflow SHALL configure `safe-outputs.create-pull-request.patch-format: am`. The generated workflow artifacts derived from that source SHALL preserve the same PR creation transport policy.

#### Scenario: Maintainer inspects authored workflow frontmatter
- **WHEN** maintainers inspect the authored `change-factory` issue-intake workflow source
- **THEN** `safe-outputs.create-pull-request.patch-format` SHALL be set to `am`

#### Scenario: Generated workflow preserves authored patch transport
- **WHEN** maintainers regenerate and inspect the compiled `change-factory` workflow artifacts
- **THEN** the generated workflow outputs SHALL preserve the `am` PR patch transport configured by the authored workflow source
