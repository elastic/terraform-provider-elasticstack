## MODIFIED Requirements

### Requirement: Preflight gate (REQ-023–REQ-027)

The workflow SHALL evaluate whether to execute CI jobs via a dedicated preflight gate job that emits a `should_run` output.

For `push` events, the preflight gate SHALL set `should_run=true` when either:

* No open pull request exists for the pushed branch in the same repository 
* All commits in the push event were authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`)

For `push` events where **neither** of the above holds, the preflight gate SHALL set `should_run=false`.

For non-`push` events (`pull_request` and `workflow_dispatch`), the preflight gate SHALL set `should_run=true`, except for `pull_request` events of type `ready_for_review` where it SHALL set `should_run=false`.

The `build`, `lint`, and matrix acceptance `test` jobs SHALL only execute when the preflight gate outputs `should_run=true`.

#### Scenario: Push without open PR

- GIVEN a push to a branch with no open PR in the same repository
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and all commits by Copilot agent

- GIVEN a push to a branch that has an open PR from the same repo
- AND every commit in the push event was authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `true`

#### Scenario: Push with open PR and a commit not by Copilot agent

- GIVEN a push to a branch that has an open PR from the same repo
- AND at least one commit in the push event was not authored by Copilot coding agent (`198982749+Copilot@users.noreply.github.com`)
- WHEN preflight runs
- THEN `should_run` SHALL be `false` and downstream jobs SHALL be skipped
