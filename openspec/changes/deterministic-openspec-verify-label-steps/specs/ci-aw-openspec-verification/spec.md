## ADDED Requirements

### Requirement: Deterministic agent setup before verification
The workflow SHALL use deterministic custom workflow steps in the agent job to prepare the repository workspace before agent reasoning begins. At a minimum, it SHALL run repository-standard Node dependency installation at the repository root so `npx openspec` is available to the agent without the prompt having to rediscover setup steps.

#### Scenario: OpenSpec CLI is available before agent reasoning
- **WHEN** the agent job starts for a verification run
- **THEN** deterministic custom steps SHALL install the repository's Node dependencies before the agent uses `npx openspec`

### Requirement: Deterministic gates may skip agent execution
The workflow SHALL use deterministic pre-activation outputs to decide whether the expensive agent job runs. When label verification or change-selection gating determines that the pull request is not eligible for verification, the workflow SHALL skip the agent job rather than starting it only to emit a no-op result. Any terminal cleanup behavior required by the workflow SHALL remain compatible with `needs.agent.result == 'skipped'`.

#### Scenario: Ineligible run skips agent job
- **GIVEN** deterministic pre-activation gating concludes the pull request is not eligible for verification
- **WHEN** downstream job conditions are evaluated
- **THEN** the workflow SHALL skip the agent job

## MODIFIED Requirements

### Requirement: Label trigger
The workflow SHALL run on `pull_request` events of type `labeled`. A deterministic pre-activation step SHALL verify that `github.event.label.name` (or equivalent injected label data) is exactly `verify-openspec` and SHALL publish the result for downstream jobs. The workflow SHALL not perform verification or archive steps when that deterministic check indicates a different label.

#### Scenario: Correct label runs automation
- **GIVEN** a pull request receives the label `verify-openspec`
- **WHEN** GitHub dispatches the `labeled` event
- **THEN** the deterministic pre-activation gate SHALL mark the run eligible for the agentic verification path

#### Scenario: Other labels do not start verification
- **GIVEN** a pull request receives a label other than `verify-openspec`
- **WHEN** the `labeled` event fires
- **THEN** the workflow SHALL not perform verification or archive steps for that event

### Requirement: Discover active change id from PR files
The workflow SHALL use a deterministic pre-activation step to load the pull request changed files list, including each file entry's status (`added`, `modified`, `removed`, `renamed`, and so on). It SHALL consider only paths matching `openspec/changes/<id>/...` where `<id>` is a single path segment and `archive` is not the first segment (that is, exclude `openspec/changes/archive/**`). For each such path, it SHALL record the status of that file entry and SHALL publish pre-activation outputs that include the gate result and, when selection succeeds, the selected active change id.

#### Scenario: Derive change id from path
- **GIVEN** a modified file `openspec/changes/my-feature/tasks.md`
- **WHEN** the deterministic selection step parses paths
- **THEN** the active change id SHALL be recognized as `my-feature`

#### Scenario: Selected change is exposed to the agent
- **GIVEN** exactly one active change satisfies the gating rules
- **WHEN** the deterministic selection step completes
- **THEN** the workflow SHALL expose that change id as a pre-activation output for the later agent job

### Requirement: Noop when change selection rules fail
The workflow SHALL not submit a pull request review and SHALL not archive when the deterministic gating result indicates any of the following:

1. More than one distinct `<id>` has at least one file with status `modified` among paths under `openspec/changes/<id>/` (non-archive).
2. Any file under `openspec/changes/` (non-archive) has status `added`.
3. Zero distinct `<id>` has at least one `modified` file under `openspec/changes/<id>/` (non-archive).
4. Any file under `openspec/changes/<id>/` (non-archive) has a status other than `modified` among the set the workflow cares about (for example, `removed` or `renamed`) if the workflow adopts modified-only strictness for verification.

#### Scenario: Two active changes modified
- **GIVEN** the pull request modifies `openspec/changes/foo/proposal.md` and `openspec/changes/bar/tasks.md`
- **WHEN** deterministic gating runs
- **THEN** the workflow SHALL skip verification and SHALL not submit a review

#### Scenario: New file under active change
- **GIVEN** the pull request adds `openspec/changes/foo/new.md`
- **WHEN** deterministic gating runs
- **THEN** the workflow SHALL skip verification

#### Scenario: Single change updated by modifications only
- **GIVEN** all non-archive `openspec/changes/` paths in the pull request refer to a single `<id>` and every such entry has status `modified`
- **WHEN** deterministic gating runs
- **THEN** the workflow SHALL select that `<id>` and continue to verification

### Requirement: Verification using active OpenSpec tooling
For the selected change id published by the deterministic pre-activation step, the agent SHALL use standard OpenSpec commands and `.agents/skills/openspec-verify-change/SKILL.md`, including where applicable `npx openspec status --change "<id>" --json` and `npx openspec instructions apply --change "<id>" --json`, and SHALL perform verification with context rooted at `openspec/changes/<id>/`. The prompt SHALL consume the selected change id from workflow outputs rather than requiring the agent to rediscover PR files before verification. The verification report SHALL include Summary, Issues by priority (CRITICAL, WARNING, SUGGESTION), and Final assessment per the skill.

#### Scenario: CLI resolves the change
- **GIVEN** `openspec/changes/<id>/` exists on the pull request branch and deterministic setup has completed
- **WHEN** `npx openspec status --change "<id>"` runs in the workflow environment
- **THEN** it SHALL succeed for a well-formed active change
