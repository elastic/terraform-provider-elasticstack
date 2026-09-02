## 1. Add explicit `noop` calls to the verification-failure branches

- [ ] 1.1 In `.github/workflows/ci-deadcode-removal-rotation.md`, step 4 ("Verify"), `make build` failure/timeout branch: after the existing `go run ./scripts/ci-deadcode-removal-rotation record ...` instruction, add "Then call `noop` with a concise reason." before "Stop without creating a PR."
- [ ] 1.2 In the same step 4, `go test` failure branch: after the "Record the attempt as `tests_failed`" instruction, add "Then call `noop` with a concise reason." before "Stop without creating a PR."
- [ ] 1.3 Confirm the companion-test backstop branch (step 3) already has this phrasing and needs no change; use it as the reference wording for 1.1/1.2.

## 2. Add the run-level "must terminate with a safe output" guardrail

- [ ] 2.1 In the "Guardrails" section of `.github/workflows/ci-deadcode-removal-rotation.md`, add: "Every run MUST end with exactly one safe-output call — either `create-pull-request` or `noop`. Never stop mid-task without calling one of them."

## 3. Regenerate the compiled workflow

- [ ] 3.1 Rebuild the compiled workflow lock artifact by running `make workflow-generate` (or the equivalent `workflows generate` command for this repo) and commit the updated `.github/workflows/ci-deadcode-removal-rotation.lock.yml`.

## 4. Update the capability spec

- [ ] 4.1 Apply the delta in `openspec/changes/deadcode-rotation-noop-guardrail/specs/deadcode-removal-rotation/spec.md` to `openspec/specs/deadcode-removal-rotation/spec.md` when this change is archived, adding the new `REQ-NOOP-001` requirement alongside the existing `REQ-FMT-001`.
