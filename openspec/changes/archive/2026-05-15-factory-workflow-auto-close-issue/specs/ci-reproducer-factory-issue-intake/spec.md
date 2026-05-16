## MODIFIED Requirements

### Requirement: Agent creates exactly one linked pull request when reproduction succeeds
When the reproduction test passes, the agent SHALL create exactly one pull request on branch `reproducer-factory/issue-{n}` labeled `reproducer-factory`. The PR body SHALL include `Related to #N` to establish deterministic linkage for duplicate-PR suppression on future runs. The PR SHALL contain only the reproduction test file and no other changes. `Related to` is used rather than a closing keyword because the reproduction does not resolve the issue — it confirms it.

#### Scenario: Eligible issue-event intake creates a reproduction PR
- **WHEN** the reproduction test passes for an eligible issue event
- **THEN** the agent SHALL emit `create-pull-request` with branch `reproducer-factory/issue-{n}` and body containing `Related to #N`

#### Scenario: Safe-output configuration prevents automatic closing references
- **WHEN** maintainers inspect the authored `reproducer-factory` workflow safe-output configuration
- **THEN** `safe-outputs.create-pull-request.auto-close-issue` SHALL be set to `false`
- **AND** generated workflow artifacts derived from that source SHALL preserve the same non-closing PR policy

#### Scenario: No PR is created when reproduction fails
- **WHEN** the agent reaches outcome B (cannot reproduce) or outcome C (appears fixed)
- **THEN** the agent SHALL NOT emit `create-pull-request`
