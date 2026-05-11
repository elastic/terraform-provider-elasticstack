# ci-change-factory-issue-intake — Delta

## MODIFIED Requirements

### Requirement: Workflow activates only for qualifying `change-factory` issue events
The workflow MAY subscribe to GitHub `issues.opened`, `issues.labeled`, and `issue_comment` events, but it SHALL activate the proposal agent only for eligible `change-factory` triggers. Eligible triggers SHALL include `issues.labeled` when the newly applied label is exactly `change-factory`; `issues.opened` when the issue already includes the `change-factory` label at creation time; and `issue_comment` events when the comment is the activation payload of a `slash_command: change-factory` trigger. In gh-aw's event model, `issue_comment` and `pull_request_comment` are distinct event names; only `issue_comment` activations (i.e. comments on issues, not pull requests) are delivered to a workflow that declares `events: [issue_comment]`. The shared `factoryQualifyTriggerEvent` helper SHALL treat `issue_comment` as an automatically eligible event name — returning `event_eligible: true` unconditionally — because gh-aw's routing guarantees that only issue-comment activations reach this handler; actor trust and duplicate checks remain the effective substantive gates.

#### Scenario: Label applied after issue creation
- **WHEN** an `issues.labeled` event is received and `github.event.label.name` is `change-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the proposal agent

#### Scenario: Issue opens with the trigger label already present
- **WHEN** an `issues.opened` event is received and the issue's initial labels include `change-factory`
- **THEN** the workflow SHALL treat the event as eligible to activate the proposal agent

#### Scenario: Non-trigger issue event is ignored
- **WHEN** an `issues` event is received without the `change-factory` label in the qualifying position for that event type
- **THEN** the workflow SHALL NOT activate the proposal agent for that event

#### Scenario: Slash command comment is treated as eligible
- **WHEN** an `issue_comment` event is received as the activation payload of a `slash_command: change-factory` trigger
- **THEN** `factoryQualifyTriggerEvent` SHALL return `event_eligible: true` for that event, because gh-aw's `events: [issue_comment]` routing guarantees the payload originates from an issue comment (not a pull request conversation)

#### Scenario: Pull request comments cannot reach the slash command handler
- **WHEN** a `/change-factory` comment is posted on a pull request conversation
- **THEN** gh-aw routes it under `pull_request_comment`, which is not listed in the workflow's `events:`; the workflow SHALL NOT activate, and `factoryQualifyTriggerEvent` is never called for that payload

### Requirement: Workflow suppresses duplicate linked pull requests
Before agent activation, the workflow SHALL detect whether an open linked `change-factory` pull request already exists for the triggering issue. A pull request SHALL be treated as linked only when it is open, carries the `change-factory` label, uses the deterministic branch name `change-factory/issue-<issue-number>`, and includes explicit issue linkage metadata such as `Closes #<issue-number>`. When a duplicate is found, the workflow SHALL post exactly one comment on the triggering issue explaining the skip and linking to the existing PR URL, then skip agent activation. The comment SHALL instruct the maintainer to close or convert the PR to a draft before retrying.

#### Scenario: Existing linked PR prevents a duplicate run and posts a comment
- **WHEN** the workflow finds an open pull request that satisfies the linked `change-factory` PR criteria for the triggering issue
- **THEN** the workflow SHALL post one comment on the triggering issue referencing the existing PR and instructing the maintainer to close it before retrying
- **AND** the workflow SHALL skip agent activation instead of opening a duplicate pull request

#### Scenario: Unrelated PR does not block issue intake
- **WHEN** an open pull request mentions the issue or has a similar title but does not satisfy the full linked `change-factory` PR criteria
- **THEN** the workflow SHALL NOT treat that pull request as the canonical linked PR for duplicate suppression

### Requirement: Agent uses the implementation-research comment as the authoritative scope baseline when present, subject to human direction override
When the triggering issue has a comment authored by `github-actions[bot]` that contains the marker `<!-- gha-research-factory -->` and a heading `## Implementation research`, the `change-factory` agent SHALL treat that entire comment as the authoritative baseline for proposal scope — unless a non-empty `human_direction` is present, in which case `human_direction` SHALL take precedence as the final say on all design decisions. When `human_direction` is empty, the agent SHALL adopt the comment's `### Recommendation` as the proposal spine, carry the comment's `### Open questions` into `design.md`, and treat `### Approaches considered` as already-evaluated context without re-exploring alternatives. If the sanitised issue body or sanitised human comments contain explicit signals that contradict the research recommendation (and no `human_direction` override is present), the agent SHALL prefer the contradicting signal and note the disagreement in the proposal artifacts. When no research comment exists, the agent SHALL use the issue title and body as the authoritative source regardless of `human_direction`.

#### Scenario: Issue has a research comment and no human direction
- **WHEN** a `change-factory` run starts for an issue that has a bot-authored research comment and `human_direction` is empty
- **THEN** the agent SHALL adopt the comment's `### Recommendation` as the chosen approach and use it as the spine of `proposal.md`
- **AND** the agent SHALL copy the comment's `### Open questions` into the resulting `design.md`
- **AND** the agent SHALL NOT re-explore the alternative approaches enumerated in `### Approaches considered`

#### Scenario: Human direction overrides research recommendation
- **WHEN** a `change-factory` run starts for an issue with a research comment and `human_direction` is non-empty
- **THEN** the agent SHALL treat `human_direction` as the final say on design decisions
- **AND** the agent SHALL NOT follow the research comment's `### Recommendation` if it conflicts with `human_direction`

#### Scenario: Issue has no research comment
- **WHEN** a `change-factory` run starts for an issue that does not have a bot-authored research comment
- **THEN** the agent SHALL author the proposal using the issue title and body as the authoritative source
- **AND** any non-empty `human_direction` SHALL still apply as the final say on design decisions

#### Scenario: Issue has a research comment but later comments contradict the recommendation
- **WHEN** a `change-factory` run starts for an issue that has a research comment and whose visible context contradicts the comment's recommendation, and `human_direction` is empty
- **THEN** the agent SHALL prefer the contradicting signal and SHALL note the disagreement in the proposal artifacts

### Requirement: Agent prompt documents human direction as a design override
The `change-factory` workflow's authored prompt SHALL include a `## Human direction` section that is presented when `human_direction` is non-empty. The section SHALL state that the human direction is the final say on all design decisions for this proposal, that it overrides the research comment's `### Recommendation` and any other design inferences, and that the agent SHALL apply it without second-guessing.

#### Scenario: Maintainer inspects the change-factory prompt for human direction handling
- **WHEN** maintainers inspect the authored `change-factory-issue` workflow prompt
- **THEN** the prompt SHALL include a section for `human_direction` that describes it as the final say on design decisions when non-empty
- **AND** the prompt SHALL explicitly state that it overrides the research comment's `### Recommendation`

#### Scenario: Empty human direction does not change agent behaviour
- **WHEN** `human_direction` is empty (label trigger or bare slash command)
- **THEN** the prompt section SHALL have no effect on agent behaviour and the existing research-recommendation handling SHALL apply unchanged
