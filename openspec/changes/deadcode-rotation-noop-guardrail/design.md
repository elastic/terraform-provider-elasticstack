## Context

The dead-code removal rotation workflow selects one dead-code candidate per run (deterministically, in pre-activation), then hands off to an LLM agent that removes the symbol, optionally removes companion tests, verifies with `make build` and `go test`, formats with `make fmt`, and opens a cleanup PR via the `create-pull-request` safe output. Every gh-aw agentic workflow run is expected to end with exactly one safe-output call; a run with none is treated by the framework as a failure and files a GitHub issue against this repository (this happened for issue #4767).

Reading the current agent task in `.github/workflows/ci-deadcode-removal-rotation.md`:

- Step 3's companion-test backstop (`resource.Test` / `resource.ParallelTest` guard) already pairs its "abort" instruction with an explicit "Then call `noop` with a concise reason."
- Step 4's three verification-failure branches (`make build` fails, `make build` times out, `go test` fails for impacted packages) each say only "Record the attempt... Stop without creating a PR." — no explicit `noop` call is mentioned in the branch itself.
- The task list ends with a catch-all sentence: "If you cannot safely proceed at any point, record the appropriate reason code and call `noop`." This is meant to cover every unmentioned failure branch, but it is easy for step 4's explicit "Stop without creating a PR" wording to read as the terminal instruction for that branch, leaving the catch-all sentence unreached.

This is consistent with the observed symptom: the agent recorded a memory attempt (satisfying "record the appropriate reason code") but the run produced no safe output at all, which matches a run that executed step 4's stop instruction literally and never got to the closing catch-all.

## Goals / Non-Goals

**Goals:**
- Every branch in the agent task that stops the run without opening a PR explicitly instructs the agent to call `noop`, inline with that branch's memory-recording instruction — no branch should depend solely on the closing catch-all sentence.
- Add a run-level guardrail, in the same style as the existing "Never modify more than one dead symbol per run" guardrails, stating the run must end with exactly one safe output.

**Non-Goals:**
- Changing the pre-activation selection algorithm, the memory schema, or the set of valid attempt-reason codes in `scripts/ci-deadcode-removal-rotation/memory.go`.
- Changing the `create-pull-request` / PR-opening path, which already works correctly.
- Addressing the `make fmt` failure branch's conformance to the existing `REQ-FMT-001` requirement (tracked separately — see Open Questions).
- Changing the job-level `if: needs.pre_activation.outputs.found == 'true'` gate that skips the entire agent job when pre-activation finds no eligible candidate. That gate causes the job to be skipped (not run-with-no-output), which is a distinct mechanism from the agent stopping mid-task, and is outside the "agent instructions" scope named in this issue.

## Decisions

### 1. Make each verification-failure branch call `noop` explicitly, rather than relying on the closing catch-all

Each of the three branches in step 4 ("Verify") gets a "Then call `noop` with a concise reason" instruction appended directly after its existing memory-recording instruction, matching the phrasing already used in step 3's backstop branch.

Why:
- Explicit, local instructions are more reliable for an LLM agent following a numbered task list than a single generic sentence at the end of the document, especially when an earlier, more specific instruction ("stop") could read as terminal for that branch.
- This mirrors the pattern already used elsewhere in this same document (step 3) and in other rotation workflows in this repository (e.g. `schema-coverage-rotation.md`'s "you MUST call `noop` with a short reason").

Alternatives considered:
- Rely solely on strengthening the closing catch-all sentence (e.g., making it more emphatic): rejected because the failure mode is specifically that a branch-local "stop" instruction pre-empts the closing sentence; strengthening wording at the end of the document does not fix that ordering problem.
- Restructure the task into a single "on any failure" subroutine referenced by every branch: rejected as a larger structural change than needed to close the gap; inline repetition is consistent with the document's existing style and easier to verify per-branch.

### 2. Add a run-level guardrail requiring exactly one terminal safe output

Add a bullet to the existing "Guardrails" section: "Every run MUST end with exactly one safe-output call — either `create-pull-request` or `noop`. Never stop mid-task without calling one of them."

Why:
- The existing Guardrails section is the natural place a reader (human or agent) checks for cross-cutting constraints that apply regardless of which task branch was taken.
- This gives the requirement a second, structurally distinct place to be found beyond the per-branch instructions, reducing the chance that a future edit to step 4 silently drops the inline `noop` instruction again.

## Risks / Trade-offs

- [Repetition of the same "call noop" instruction across three branches increases document length] — acceptable; the existing document already repeats similar instructions (e.g., the `record` command invocation) per branch, and clarity for an LLM-followed task list matters more than terseness here.
- [The workflow's compiled lock file (`ci-deadcode-removal-rotation.lock.yml`) must be regenerated after the source `.md` changes, or the deployed workflow will not reflect the fix] — captured explicitly in `tasks.md` for the implementer.

## Open Questions

- Should the `make fmt` failure branch (step 5) also gain an inline "call `noop`" instruction to fully satisfy the already-approved `REQ-FMT-001` requirement in `openspec/specs/deadcode-removal-rotation/spec.md`, which mandates recording reason `fmt_failed` and calling `noop`? The current `.md` does not implement failure handling for this step at all, and `fmt_failed` is not currently a valid attempt reason in `scripts/ci-deadcode-removal-rotation/memory.go`. This is a related but pre-existing gap, not introduced by this change; flagging it here so the implementer or a follow-up issue can decide whether to fold it into this change's implementation or track it separately.
