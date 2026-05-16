## 1. Author the scheduled dead-code rotation workflow and deterministic pre-activation

- [ ] 1.1 Add the authored GitHub Agentic Workflow source and compiled workflow artifacts for scheduled and manual dead-code cleanup rotation.
- [ ] 1.2 Implement deterministic pre-activation logic to run `go tool deadcode ./...` and `go tool deadcode -test ./...`, intersect the candidate sets, apply cooldown filtering, and select exactly one candidate per run using stable deterministic ordering.
- [ ] 1.3 Implement deterministic reference analysis with `gopls references` to collect unique referring files and classify whether local companion test cleanup is eligible.
- [ ] 1.4 Implement deterministic acceptance-test exclusion for local companion test cleanup using the `acc_*test.go` filename rule.
- [ ] 1.5 Add focused tests for candidate parsing, candidate intersection, cooldown filtering, deterministic selection, and reference-classification edge cases.

## 2. Implement agent cleanup behavior, verification, and cooldown memory

- [ ] 2.1 Write the agent prompt and workflow contract for removing the target dead function, aborting on invalid acceptance-style candidates, and removing companion tests only when the single-file local eligibility rule has already been satisfied.
- [ ] 2.2 Implement the acceptance-style backstop in the agent contract using `resource.Test` / `resource.ParallelTest` detection before any test deletion.
- [ ] 2.3 Implement verification with `make build` using a 10 minute timeout and unit tests for the impacted package or packages.
- [ ] 2.4 Implement cooldown-only memory updates for every attempted candidate regardless of outcome, and stop the run after invalid-candidate detection or verification failure.

## 3. Open PRs and document maintainer workflow

- [ ] 3.1 Configure successful verified runs to open a cleanup PR containing one dead-code candidate removal.
- [ ] 3.2 Document the workflow scope, candidate eligibility rules, verification behavior, cooldown memory model, and the expectation that maintainers review, merge, or close PRs manually.
- [ ] 3.3 Rebuild generated workflow artifacts and run relevant OpenSpec and targeted validation for the new rotation workflow.
