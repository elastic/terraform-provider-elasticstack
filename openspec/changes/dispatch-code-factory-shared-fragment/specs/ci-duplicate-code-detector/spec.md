# ci-duplicate-code-detector — dispatch-code-factory delta

Delta spec for capability `ci-duplicate-code-detector`. Adds the code-factory dispatch requirement that was present in the other three producer workflows but missing from the duplicate-code detector.

## ADDED Requirements

### Requirement: Created duplicate-code issues are explicitly dispatched to `code-factory`

After safe-output issue creation completes, the workflow SHALL explicitly dispatch the `code-factory` workflow once for each duplicate-code issue created in the current run, using the `dispatch-code-factory` shared fragment imported via `imports: [shared/dispatch-code-factory.md]`.

#### Scenario: One created duplicate-code issue dispatches one implementation run

- **WHEN** the workflow creates one duplicate-code issue in a run
- **THEN** it SHALL dispatch exactly one `code-factory` workflow run for that issue

#### Scenario: Multiple created duplicate-code issues dispatch multiple implementation runs

- **WHEN** the workflow creates multiple duplicate-code issues in a run
- **THEN** it SHALL dispatch exactly one independent `code-factory` workflow run per created issue

#### Scenario: No issues created means no dispatches

- **WHEN** the workflow completes without creating any issues (noop or no-signal run)
- **THEN** it SHALL NOT dispatch `code-factory` runs

### Requirement: Agent prompt instructs dispatch after issue creation

The `duplicate-code-detector.md` agent prompt SHALL include a `## Dispatch` section that instructs the agent to call the `dispatch_code_factory` safe output tool once after all issues have been created (or after determining no issues will be created).

#### Scenario: Agent calls dispatch_code_factory at end of run

- **WHEN** the agent completes issue creation (or determines no issues are needed)
- **THEN** it SHALL call the `dispatch_code_factory` tool exactly once, consistent with the instruction in the prompt's `## Dispatch` section

### Requirement: Dispatch block is imported from the shared fragment, not inlined

The `duplicate-code-detector.md` workflow source SHALL obtain the `dispatch-code-factory` safe-outputs job via `imports: [shared/dispatch-code-factory.md]` and SHALL NOT contain an inline copy of the job block.

#### Scenario: Compiled lock contains dispatch_code_factory job

- **WHEN** the workflow source imports `shared/dispatch-code-factory.md`
- **THEN** the compiled `duplicate-code-detector.lock.yml` SHALL contain the `dispatch_code_factory` tool declaration and the `dispatch-code-factory` job descriptor

#### Scenario: Test suite asserts dispatch presence

- **WHEN** `duplicate-code-detector.test.mjs` runs
- **THEN** it SHALL assert that the workflow source references `dispatch_code_factory`, the compiled lock contains `dispatch_code_factory`, and the lock contains the `dispatch-code-factory` job descriptor string
