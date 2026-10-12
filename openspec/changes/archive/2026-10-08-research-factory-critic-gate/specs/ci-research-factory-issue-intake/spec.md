## MODIFIED Requirements

### Requirement: Workflow time-boxes the research session and survives partial completion
The workflow SHALL set a job-level `timeout-minutes` of 60 minutes. The agent prompt SHALL communicate a 50-minute self-budget to the agent, covering all research and critique rounds, and SHALL instruct it to reserve the final minutes of the budget for emitting its research comment. The prompt SHALL further instruct the agent that, if research time runs short, it SHALL prefer emitting a partial-but-valid research comment (with explicit unanswered open questions) over emitting `noop`. If the self-budget runs out before the critique loop finishes, the agent SHALL stop iterating, emit its latest revision, and report a gate outcome of `research-needs-human`.

#### Scenario: Maintainer inspects compiled workflow timeout
- **GIVEN** the compiled `research-factory-issue.md` workflow
- **WHEN** maintainers inspect the agent job definition
- **THEN** the agent job SHALL declare `timeout-minutes: 60`

#### Scenario: Agent prompt communicates the self-budget
- **GIVEN** the `research-factory-issue.md` workflow source
- **WHEN** maintainers inspect the agent prompt body
- **THEN** the prompt SHALL state the 50-minute research self-budget covering all critique rounds
- **AND** the prompt SHALL state that the agent SHALL prefer a partial-but-valid research comment over `noop` when running short on time

#### Scenario: Self-budget expires during the critique loop
- **GIVEN** the agent has produced at least one research draft
- **WHEN** the self-budget expires before the done gate is satisfied
- **THEN** the agent SHALL emit its latest research revision as the research comment
- **AND** the comment's gate outcome SHALL be `research-needs-human`
- **BUT** the agent SHALL NOT emit `noop` solely because the critique loop did not finish

### Requirement: Workflow removes the factory trigger label in pre-activation when the agent proceeds
The workflow SHALL include a deterministic pre-activation step that removes the `research-factory` label from the triggering issue using `actions/github-script@v9` with `x-script-include` to an inline script that delegates to the shared `.github/scripts/workflows/lib/remove-trigger-label.js` helper (parameterized to accept the factory label name and issue number). The step SHALL run only when the workflow would proceed to the research agent (eligible qualifying issue event, trusted actor — when applicable — and concurrency gate satisfied). Whenever the workflow would proceed to the research agent, in either intake mode (issue event or `workflow_dispatch`), pre-activation SHALL also remove any gate-outcome label (`ready-for-change-factory` or `research-needs-human`) left on the issue by a previous research run, and SHALL NOT remove any other label. The workflow SHALL grant `issues: write` to pre-activation where required for label removal. Dispatch-targeted issues SHALL NOT require the `research-factory` label to be present, and the workflow SHALL NOT treat the absence of that label on dispatch-targeted issues as an error.

#### Scenario: Remove step mirrors verify workflow pattern
- **GIVEN** the authored `research-factory` issue-intake workflow
- **WHEN** maintainers inspect its `on.steps`
- **THEN** it SHALL include a remove-label step structurally equivalent to the change-factory and code-factory remove-label steps, with the `Remove trigger label` step using `actions/github-script@v9` and `x-script-include` to include the shared trigger-label removal script/helper path
- **AND** the included script SHALL reuse the shared `remove-trigger-label` library (not a forked copy of the GitHub API logic)

#### Scenario: Label removed only when agent gate passes
- **GIVEN** an issue for which research is requested
- **WHEN** pre-activation determines the research agent SHALL run for the issue
- **THEN** the remove-label step SHALL run and SHALL attempt to remove `research-factory` from that issue
- **AND** pre-activation SHALL remove `ready-for-change-factory` and `research-needs-human` if either is present

#### Scenario: Dispatch-targeted issue has no `research-factory` label
- **GIVEN** an issue that does not carry the `research-factory` label
- **WHEN** the workflow is triggered by `workflow_dispatch` for that issue
- **THEN** the workflow SHALL continue normally and SHALL NOT require trigger-label removal for that run
- **AND** pre-activation SHALL still remove any stale gate-outcome label from that issue

### Requirement: Workflow remains research-only and does not write code
The `research-factory` workflow SHALL NOT implement provider, CI, or documentation behavior, SHALL NOT open pull requests, and SHALL NOT modify repository files. Its only durable outputs SHALL be a single `update_research_comment` safe-output operation executed by the custom `update-research-comment` script, producing a comment conforming to the `ci-research-factory-comment-format` capability, and the gate-outcome label changes made by that script and by pre-activation as defined in this capability. The workflow SHALL NOT enable safe outputs that would permit creating pull requests, creating issues, or posting free-form comments beyond the framework's own `status-comment`.

#### Scenario: Maintainer inspects compiled workflow safe outputs
- **GIVEN** the compiled `research-factory-issue.md` workflow
- **WHEN** maintainers inspect its `safe-outputs:` block
- **THEN** it SHALL include a custom safe-output job named `update-research-comment`
- **AND** it SHALL NOT include `update-issue`, `create-pull-request`, `push-to-pull-request-branch`, `update-pull-request`, or `create-issue`
- **AND** it SHALL NOT include `add-comment`

#### Scenario: Issue requests provider implementation
- **GIVEN** a qualifying issue that describes a Terraform resource, data source, or other provider implementation
- **WHEN** the research agent runs
- **THEN** the agent SHALL produce a research comment describing approaches and open questions
- **AND** the agent SHALL NOT modify provider source, generated clients, or documentation

### Requirement: Workflow does not promote the issue to a downstream factory
The `research-factory` workflow SHALL NOT apply the `change-factory`, `code-factory`, or any other factory trigger label as part of its research run. Promotion of an issue from the research stage to a downstream stage SHALL be performed by a human maintainer or by a separate (future) classifier workflow, not by `research-factory` itself. The only labels the workflow's research output MAY apply are the gate-outcome labels `ready-for-change-factory` and `research-needs-human`, and these SHALL be applied only by the custom `update-research-comment` safe-output script. Neither gate-outcome label SHALL trigger any workflow.

#### Scenario: Research run completes successfully
- **GIVEN** an eligible issue
- **WHEN** the research agent finishes a successful run
- **THEN** the workflow SHALL NOT add the `change-factory`, `code-factory`, or any other factory trigger label to the issue

#### Scenario: Maintainer inspects compiled workflow safe outputs
- **GIVEN** the compiled `research-factory-issue.md` workflow
- **WHEN** maintainers inspect its `safe-outputs:` block
- **THEN** the block SHALL NOT enable `add-labels`
- **AND** the block SHALL NOT enable `remove-labels`

#### Scenario: Research passes the done gate
- **GIVEN** a research run whose gate outcome is `ready-for-change-factory`
- **WHEN** the run completes
- **THEN** the issue SHALL carry the `ready-for-change-factory` label
- **BUT** the issue SHALL NOT carry the `change-factory` label as a result of the run
- **AND** no downstream factory workflow SHALL be started by the run

### Requirement: Agent emits a single research comment via custom safe-output script
When the deterministic gate passes and the agent completes its research, the agent SHALL emit a single `update_research_comment` safe-output operation whose `body` payload contains the research content conforming to the `ci-research-factory-comment-format` capability. The workflow SHALL define a custom `safe-outputs.jobs` entry named `update-research-comment` that creates or updates an issue comment authored by `github-actions[bot]`. If an existing comment by `github-actions[bot]` containing the marker `<!-- gha-research-factory -->` is found on the issue, the script SHALL update that comment; otherwise it SHALL create a new comment. After writing the comment, the script SHALL apply exactly one gate-outcome label, derived from the comment's metadata as defined by the `ci-research-factory-comment-format` capability, and SHALL remove the other gate-outcome label if present. The agent SHALL NOT emit `update_issue`, `add-comment`, or any other safe-output operation as part of its research output.

#### Scenario: Agent produces research on a fresh issue
- **GIVEN** an eligible issue with no prior research comment
- **WHEN** the workflow runs
- **THEN** the agent SHALL emit one `update_research_comment` operation
- **AND** the custom script SHALL create a new comment on the issue
- **AND** that comment SHALL contain `<!-- gha-research-factory -->` as its first line
- **AND** the issue SHALL carry exactly one of `ready-for-change-factory` or `research-needs-human`

#### Scenario: Agent regenerates an existing research comment
- **GIVEN** an eligible issue that already has a research comment by `github-actions[bot]`
- **WHEN** the workflow runs
- **THEN** the agent SHALL emit one `update_research_comment` operation
- **AND** the custom script SHALL update the existing comment in place
- **AND** the issue SHALL NOT gain an additional research comment

#### Scenario: Agent times out before reaching a confident recommendation
- **GIVEN** an eligible issue
- **WHEN** the agent's self-budget expires before research is complete
- **THEN** the agent SHALL emit a partial-but-valid research comment with explicit unanswered open questions
- **AND** the issue SHALL carry the `research-needs-human` label
- **BUT** the agent SHALL NOT emit `noop` solely because research is partial

## ADDED Requirements

### Requirement: Research is refined through a bounded, independent critique loop
Every research run, whether triggered by label or by `workflow_dispatch`, SHALL refine its research through a loop of draft -> critique -> revise before emitting the research comment. Each critique SHALL be produced by a critic role that is separate from the research author and runs on a different model from the author. The critic SHALL evaluate the current draft against the hard checklist defined by the `ci-research-factory-comment-format` capability, SHALL score it from 0 to 100 across completeness, feasibility, compatibility, test coverage, and idiomaticness, and SHALL list any actionable feedback. The loop SHALL run at most 3 rounds, where one round is one critique of one draft. The loop SHALL stop early once the research is converged and every hard-checklist item passes. If the research is converged but a checklist item fails, the loop SHALL continue, with the critic's feedback targeting the failing items, until the round bound is reached. The critique loop SHALL NOT span multiple workflow runs.

#### Scenario: Strong first draft finishes in one round
- **GIVEN** a first research draft that the critic scores at 85 or higher
- **AND** the critic reports no actionable feedback
- **AND** every hard-checklist item passes
- **AND** the draft has no blocking open questions
- **WHEN** the critique round completes
- **THEN** the agent SHALL stop iterating after 1 round
- **AND** the issue SHALL be labelled `ready-for-change-factory`

#### Scenario: Draft improves to the threshold over the remaining rounds
- **GIVEN** a first research draft the critic scores below 85 with actionable feedback
- **AND** the critic remains available for every round
- **WHEN** the author revises the draft twice and the critic scores both revisions at 85 or higher with every checklist item passing
- **AND** the final revision has no blocking open questions
- **THEN** the loop SHALL stop after round 3
- **AND** the issue SHALL be labelled `ready-for-change-factory`

#### Scenario: Converged but a checklist item still fails
- **GIVEN** a draft the critic scores at 90 with no actionable feedback beyond a failing `Testable` item
- **WHEN** the critique round completes before the round bound
- **THEN** the loop SHALL continue to another round

#### Scenario: Round bound reached without convergence
- **GIVEN** a research draft that has not satisfied the convergence rule
- **WHEN** the third critique round completes
- **THEN** the agent SHALL stop iterating
- **AND** the gate outcome SHALL be `research-needs-human`

#### Scenario: Critic is unavailable
- **GIVEN** the critic model cannot be reached or returns no usable critique
- **WHEN** the agent attempts a critique round
- **THEN** the agent SHALL still emit a research comment
- **AND** the comment metadata SHALL record the critic status as not ok
- **AND** the gate outcome SHALL be `research-needs-human`

#### Scenario: Maintainer inspects critic independence
- **GIVEN** the `research-factory-issue.md` workflow source and a produced research comment
- **WHEN** maintainers inspect the workflow configuration and the comment metadata
- **THEN** the critic role SHALL be configured with a model that differs from the author model
- **AND** the comment metadata SHALL record both the author model and the critic model identifiers

### Requirement: Issue carries exactly one gate-outcome label after each research run
After every research run that emits a research comment, the issue SHALL carry exactly one of the labels `ready-for-change-factory` or `research-needs-human`. When a new research run proceeds past pre-activation, any gate-outcome label from a previous run SHALL be removed before the agent starts, so a stale outcome is never visible during or after a re-run. If a run ends without emitting a research comment, the issue SHALL carry neither gate-outcome label. The `ready-for-change-factory` and `research-needs-human` labels SHALL exist in the repository; maintainer documentation SHALL list them alongside the other factory labels. The workflow SHALL NOT add or remove the issue classifier's `needs-human` label.

#### Scenario: Re-run replaces a previous ready outcome
- **GIVEN** an issue carrying `ready-for-change-factory` from a previous research run
- **WHEN** a maintainer re-applies the `research-factory` label and the new run proceeds
- **THEN** `ready-for-change-factory` SHALL be removed before the agent starts
- **AND** after the run the issue SHALL carry exactly the outcome label of the new run

#### Scenario: Run ends without a research comment
- **GIVEN** an issue carrying `research-needs-human` from a previous research run
- **WHEN** a new research run proceeds past pre-activation but the agent job fails before emitting a research comment
- **THEN** the issue SHALL carry neither `ready-for-change-factory` nor `research-needs-human` as a result of the run

#### Scenario: Classifier label is preserved
- **GIVEN** an issue carrying the issue classifier's `needs-human` label
- **WHEN** a research run proceeds and completes with either gate outcome
- **THEN** the issue SHALL still carry `needs-human`

#### Scenario: Run does not proceed
- **GIVEN** an issue carrying `ready-for-change-factory`
- **WHEN** a research trigger is rejected in pre-activation (for example, an untrusted actor)
- **THEN** the issue's gate-outcome label SHALL be unchanged
