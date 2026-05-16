## 1. Author the scheduled dead-code rotation workflow and deterministic pre-activation

- [x] 1.1 Add the authored GitHub Agentic Workflow source and compiled workflow artifacts for scheduled and manual dead-code cleanup rotation.
- [x] 1.2 Implement deterministic pre-activation logic to run `go tool deadcode ./...` and `go tool deadcode -test ./...`, intersect the candidate sets, apply cooldown filtering, and select exactly one candidate per run using stable deterministic ordering.
- [x] 1.3 Implement deterministic reference analysis with `gopls references` to collect unique referring files and classify whether local companion test cleanup is eligible.
- [x] 1.4 Implement deterministic acceptance-test exclusion for local companion test cleanup using the `acc_*test.go` filename rule.
- [x] 1.5 Add focused tests for candidate parsing, candidate intersection, cooldown filtering, deterministic selection, reference-classification edge cases, and reason-code classification/aggregation helpers.

## 2. Implement agent cleanup behavior, verification, and cooldown memory

- [x] 2.1 Write the agent prompt and workflow contract for removing the target dead function, aborting on invalid acceptance-style candidates, and removing companion tests only when the single-file local eligibility rule has already been satisfied.
- [x] 2.2 Implement the acceptance-style backstop in the agent contract using `resource.Test` / `resource.ParallelTest` detection before any test deletion.
- [x] 2.3 Implement verification with `make build` using a 10 minute timeout and unit tests for the impacted package or packages.
- [x] 2.4 Implement cooldown-only memory updates for every attempted candidate regardless of outcome, including deterministic outcome reason codes and small structured context, and stop the run after invalid-candidate detection or verification failure.

## 3. Add maintainer-facing outcome summaries, open PRs, and document workflow expectations

- [x] 3.1 Implement compact periodic aggregation of recent attempt outcomes by deterministic reason code and sticky package/area.
- [x] 3.2 Expose the outcome summary through the chosen maintainer-facing surface such as a markdown artifact and/or marker-based issue or comment.
- [x] 3.3 Configure successful verified runs to open a cleanup PR containing one dead-code candidate removal.
- [x] 3.4 Document the workflow scope, candidate eligibility rules, verification behavior, cooldown memory model, deterministic reason codes, outcome summaries, and the expectation that maintainers review, merge, or close PRs manually.
- [x] 3.5 Rebuild generated workflow artifacts and run relevant OpenSpec and targeted validation for the new rotation workflow.
