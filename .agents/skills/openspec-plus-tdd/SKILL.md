---
name: openspec-plus-tdd
description: "MANDATORY skill that activates whenever code is written to implement an OpenSpec change task. Triggers: openspec-plus-apply is active, /opsx-apply is running, the user is implementing tasks from an OpenSpec change, an implementer subagent dispatched by openspec-plus-apply is starting work, or the user invokes phrases like 'TDD for the change', 'implementing change tasks', or 'writing tests for spec scenarios'. Load before any production code is written for an OpenSpec change. Enforces strict RED-GREEN-REFACTOR per test (any test — acceptance, unit, edge case, helper). Iron Law: NO PRODUCTION CODE WITHOUT A FAILING TEST. Gherkin scenarios in spec.md are the canonical source for acceptance tests (every scenario MUST become at least one test); additional unit, edge-case, and helper tests are encouraged and follow the same per-test cycle."
metadata:
  version: 1.6.0
  priority: high
  disable-user-invocation: true
---

# OpenSpec Plus TDD

## Mission

Strict RED-GREEN-REFACTOR per test for OpenSpec change implementation. Every test — whether derived from a Gherkin scenario relevant to the slice in `spec.md`, written for a unit, edge case, helper, or error path — goes through its own atomic cycle before the next test begins. Production code exists only to make a previously-failing test pass. Surgical changes, simplicity first, no speculative abstractions, every changed line traces to a slice task.

Gherkin scenarios in `spec.md` are the **canonical source for acceptance tests**: every scenario relevant to the slice MUST become at least one test. The implementer is encouraged to add additional tests — unit tests for individual functions, edge-case tests, helper tests, error-path tests — when fast-feedback granularity is valuable. Every test follows the same cycle.

Loaded by `openspec-plus-apply` (subagent prompt + inline mode) before any code is written.

---

> **RIGID. NEVER write production code before a test fails for the right reason. NEVER write tests for multiple cases before the first one is GREEN. NEVER skip the REFACTOR assessment. NEVER mark a task `[x]` while a relevant test is failing or skipped. NEVER add comments for non-complex logic. NEVER refactor code outside the slice. NEVER write any code (test or production) before reading the project's referenced coding/testing standards. NEVER ship without covering every Gherkin scenario in spec.md with at least one test. Letter and spirit are the same.**

**Red flags — STOP, you are about to violate this skill:**

- "I'll write the test after, it's faster"
- "Too simple to need a test"
- "Manually verified, that's enough"
- "Gherkin scenario is vague, generic test is fine"
- "Skip this failing test, circle back later"
- "Mark `.skip` to unblock the slice"
- "While I'm here, clean up the adjacent code"
- "Add an interface in case we swap implementations"
- "Short comment explains the obvious"
- "Error handling for cases that can't happen"
- "Test passed first run, must be right"
- "Let me write tests for all the cases first, then implement"
- "Test 1 done — I have a clear picture, let me write all the rest at once"
- "I have a clear picture of all 5 cases — let me write them all"
- "Writing one test at a time is slower"
- "These cases are related, I'll batch them"
- "The code I'm about to write covers test 2 anyway, no need to write its test separately first"
- "Acceptance tests cover the happy path — skip the unit/edge tests"
- "Scenarios covered, no need to add granular tests even though the helper has edge cases"
- "Nothing to refactor, skip the assessment"
- "I know the project conventions, no need to re-read AGENTS.md"
- "AGENTS.md has many rules — I'll apply the ones that feel relevant"

None justify production code without a red test, ignored failures, speculative abstractions, comments on obvious code, or scope expansion.

---

## The Iron Law

```
NO PRODUCTION CODE WITHOUT A FAILING TEST
```

Every test — acceptance, unit, edge case, helper, error path — must be observed to fail for the right reason before the production code that makes it pass is written. Test not observed to fail = test proves nothing. Code written first = delete it, start over. No exceptions without explicit user permission.

### Mandatory Acceptance Coverage

Every Gherkin scenario relevant to the slice in `spec.md` MUST become at least one test. The scenario IS the acceptance contract; the test IS the verification. A slice cannot ship with an uncovered scenario, even if all other tests pass.

### Encouraged Granular Coverage

Beyond acceptance tests, add unit/edge/helper/error tests when valuable (non-trivial branches, null/empty inputs, boundary values, error paths). Same RED-GREEN-REFACTOR cycle — no special handling regardless of test origin.

---

## Inputs

* Slice tasks (`tasks.md`)
* Slice spec requirements + Gherkin scenarios (`spec.md`)
* Slice design decisions (`design.md`)
* Project standards (`AGENTS.md` / `CLAUDE.md` / `GEMINI.md` if in context)
* Existing code in slice's affected files

NEVER read source outside the slice's affected files.

---

## Workflow

```text
Phase 0: Pre-RED — read project's coding/testing conventions; follow strictly

Phase 1+: Plan the test set for the slice:
  Mandatory:  one test per Gherkin scenario relevant to the slice in spec.md
  Encouraged: additional unit / edge-case / helper / error-path tests
              when fast-feedback granularity is valuable

Phase 2+: For each test, ONE AT A TIME (any test, in any order):
  Follow the Per-Test State Machine (digraph below) atomically.
  Do NOT begin the next test until the current one terminates at
  "Test Complete".

After all tests complete AND every Gherkin scenario relevant to the slice is covered, run
the slice's pre-mark gate.
```

### Per-Test State Machine (MANDATORY)

Every test traverses this cycle end-to-end before work begins on the next. Atomic per test — no shortcuts, no batching, no skipping nodes. Applies to all test types (acceptance and granular).

**Per-test cycle:** START → RECORD STATE BEFORE (file path + count + names) → RED (write ONE failing test K) → VERIFY-RED (fails for expected reason? no → fix test, retry) → GREEN (minimum production code for K only) → VERIFY-GREEN (K passes, others green, output pristine? no → fix production code, retry) → REFACTOR ASSESS (needed? yes → act, verify green, revert if broken; no → record "not needed — reason") → RECORD STATE AFTER (count = previous + 1) → TEST K COMPLETE → return to START for K+1 (or end if all done AND all Gherkin scenarios covered).

The cycle forbids:

* Starting test K+1 before K reaches COMPLETE
* Skipping REFACTOR assessment — every test passes through it
* Skipping state recording — audit trail is mandatory
* Ending slice with uncovered Gherkin scenarios

---

## Concrete Pattern: WRONG vs RIGHT

This is the single most-violated rule. Read both examples carefully.

### WRONG — Batching (the model's training default)

```
Implementer opens empty test file.
Writes test 1, test 2, test 3, test 4, test 5 in one pass.
Runs tests — all 5 fail.
Writes production code covering all 5 cases in one pass.
Runs tests — all 5 pass.
Reports DONE.
```

This is NOT TDD. Each test passed immediately when production code arrived. You never observed test 1 failing in isolation. You never refactored after each green. You wrote the whole solution in your head and dumped it onto disk.

The end state (5 tests, 5 features) is identical to RIGHT — but the discipline is absent. The reviewer cannot tell from end state alone, but YOU know you batched.

### RIGHT — One Test At A Time

```
State: 0 tests.

Test 1 (acceptance — Gherkin "valid login"):
  Write test 1 (1 test, 1 failing). RED: "expected `Email required`, got undefined" ✓
  Write minimum production. GREEN: 1 passing, pristine ✓
  Refactor: no duplication, names clear → "not needed."

Test 2 (acceptance — Gherkin "invalid password"):
  Write test 2 (2 tests, 1 failing). RED: "expected `Invalid password`, got `Internal error`" ✓
  Write minimum production. GREEN: 2 passing ✓
  Refactor: extracted `mapAuthError` helper. Tests green.

Test 3 (unit — edge case for `mapAuthError(null)`):
  Write test 3 (3 tests, 1 failing). RED: "expected `Invalid input`, got TypeError" ✓
  Add null guard. GREEN: 3 passing ✓. Refactor: not needed.

[repeat for each test...]
```

Tests 1-2: Gherkin scenarios (mandatory acceptance). Test 3: implementer-initiated edge case (granular). All follow the same per-test cycle.

If you find yourself thinking *"I know all 5 cases, let me write them all at once"* — STOP. That is the violation. Delete what you just wrote. Restart from test 1.

### Why "I'll write all the tests, then implement" is wrong

* Test 2 might pass immediately when you implement test 1 — you'd never know if test 2 actually tests what you think.
* No checkpoint forces you to confront edge cases per test. Edge cases get glossed.
* You miss refactor opportunities that emerge between tests.
* The discipline is the value, not the end state.

---

## Phase 0: Pre-RED — Read Referenced Conventions

Before ANY code (mandatory, once per slice):

1. `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` (or equivalents at project root, `.claude/`, `.opencode/`, `docs/`)
2. Follow references inside those files to other docs (coding standards, testing conventions, patterns)
3. Slice's affected files — absorb local style

These files are the contract — follow every documented rule strictly, end-to-end (no cherry-picking). Re-read per slice (files may have been updated). Do NOT proceed to Phase 1 before reading is done.

---

## Phase 1: RED — Failing Test (One At A Time)

The test you write at this phase is one of two kinds:

**1. Acceptance test — translated from a Gherkin scenario.**

A Gherkin scenario in `spec.md`:

```gherkin
#### Scenario: User logs in with valid credentials
GIVEN a user account exists with email `alice@example.com` and password `correct-pw`
WHEN the user submits the login form with those credentials
THEN the response sets a session cookie
AND the user is redirected to `/dashboard`
```

Translate directly into one minimal acceptance test:

```typescript
test('logs in with valid credentials', async () => {
  await createUser({ email: 'alice@example.com', password: 'correct-pw' });

  const response = await submitLogin({ email: 'alice@example.com', password: 'correct-pw' });

  expect(response.headers['set-cookie']).toMatch(/session=/);
  expect(response.status).toBe(302);
  expect(response.headers.location).toBe('/dashboard');
});
```

**2. Granular test — implementer-initiated for a unit, edge case, helper, or error path**

Example: while implementing the login above, the implementer factors out a `mapAuthError` helper. They add a unit test for it:

```typescript
test('mapAuthError handles null input', () => {
  expect(() => mapAuthError(null)).toThrow('Invalid input');
});
```

This test was not derived from a Gherkin scenario — it was added because the helper's null case needs fast-feedback coverage. It follows the same cycle.

Rules (both kinds):

* One test at a time — never batch. Test name describes behavior, not implementation.
* Real code paths; mocks ONLY when dependency unavailable. Test the OUTCOME, not call sequence.
* Acceptance: translate Gherkin faithfully. Granular: state the unit's contract explicitly.

---

## Phase 2: VERIFY-RED — Watch It Fail Correctly

**MANDATORY. NEVER SKIP.**

Run the test. Confirm:

1. Test FAILS (not errors, not passes).
2. Failure message matches what scenario implies.
3. Failure is because feature is missing — not typo, not missing import, not setup bug.

Test passes immediately → feature exists or test is wrong. Fix the test.
Test errors → fix error, re-run until it fails for the expected reason.

---

## Phase 3: GREEN — Minimum Production Code

Simplest code that passes the test.

* No features beyond what the failing scenario requires
* No abstractions for single-use code
* No flexibility/configuration the scenario didn't ask for
* No error handling for impossible cases
* Match existing patterns

If 200 lines and 50 would do — rewrite.

---

## Phase 4: VERIFY-GREEN — Watch It Pass

**MANDATORY.**

1. Test passes.
2. Other tests in the file still pass.
3. Tests for slice's other scenarios still pass.
4. Output pristine — no warnings, errors, deprecations.

Test fails → fix production code, NOT the test. Test is the source of truth.
Unrelated tests fail → fix now, before next test.

---

## Phase 5: REFACTOR — Mandatory Assessment, Conditional Action

REFACTOR is NOT optional. After every GREEN, you MUST perform an explicit refactor assessment.

### 5.1 Assessment (always)

Apply clean code principles, project conventions (from Phase 0 reading), and look for common code smells in the code introduced by this test's GREEN phase (within the slice's files only).

Answer in the report: **Is refactoring needed?** (yes/no with one-sentence reason)

If **no** → state explicitly: *"Refactor assessment: no refactoring needed — code is minimal, clear, and non-duplicative."* Then move to NEXT.

If **yes** → proceed to 5.2.

### 5.2 Action (only if assessment = yes)

Refactor respecting project conventions. Tests must stay green — run after every edit. NEVER add behavior, touch outside slice, or reformat adjacent code. If tests break → revert.

Record outcome (e.g., *"Extracted `parseAuthHeader` helper; tests green"* or *"No refactoring needed — minimal, clear, non-duplicative"*).

---

## Phase 6: NEXT — Only Now

Only after the current test's REFACTOR assessment is recorded, move to the next test. Return to Phase 1 RED with the next test (the next uncovered Gherkin scenario, OR the next implementer-initiated unit/edge/helper/error test).

NEVER skip ahead. NEVER write the next test while the current one is still in GREEN or REFACTOR phase.

The slice is done when:
* Every Gherkin scenario in spec.md has at least one passing test (mandatory acceptance coverage), AND
* Every test the implementer added (unit, edge, helper, error) is passing, AND
* All tests went through the per-test cycle individually.

---

## Implementation Principles (apply throughout)

| Principle | TDD application |
|---|---|
| Think Before Coding | Read scenario or unit contract carefully. State assumptions. Ambiguous → ASK before writing the test. |
| Simplicity First | One test at a time (acceptance OR granular). Minimum production change. No speculative abstractions. |
| Surgical Changes | Every changed line traces to a slice task. Don't refactor adjacent code. Match existing style. |
| Goal-Driven Execution | The current test (Gherkin scenario OR unit contract) IS the verifiable goal. Loop independently until the test passes. |

---

## Code Style Rules — Code As Documentation

Code explains itself through good names, small focused functions, and clear structure.

* **Self-documenting.** Names describe intent. Functions do one thing. Structure makes flow obvious.
* **Comments only when:** genuinely non-obvious algorithm, external-constraint workaround, or counter-intuitive tradeoff. Refactor before commenting — if better naming/structure removes the need, do that first.
* **Never:** describe obvious behavior in comments; leave commented-out code, TODO, FIXME.
* **Testable** (hard to test = hard to use), **readable** (one mental load per function), **maintainable** (one responsibility per file).

---

## Verification Gates

**Per-test:** Pre-RED done ✓ RED observed ✓ GREEN passes + others green + pristine ✓ Minimum code ✓ No unnecessary comments ✓ No adjacent code touched ✓ REFACTOR assessed ✓ — Cannot check all? Restart from RED.

**Per-slice:** Every Gherkin scenario covered ✓ All tests pass ✓ Each test went through its own cycle (no batching) ✓

---

## When Stuck

| Symptom | Cause | Fix |
|---|---|---|
| Hard to write test | Interface hard to use | Simplify interface before writing test |
| Need to mock half the world | Code too coupled | Dependency injection or split unit |
| Test setup huge | Coupling | Extract helpers; still huge → simplify design |
| Scenario ambiguous | Spec gap | STOP. Ask user. Don't invent |
| Don't know what granular tests to add | Just-write-tests trap | Add tests for: non-trivial branches, null/empty inputs, boundary values, error paths the scenario doesn't cover. Skip if behavior is fully covered by acceptance test. |
| Test passes immediately | Feature exists or test wrong | Investigate; rewrite test |
| Can't make test fail right | Test wrong | Rewrite before any production code |
| Production code keeps growing | Over-engineering | What's the minimum that passes? |

---

## Anti-Patterns

NEVER: skip Pre-RED reading | batch tests | move to next test before current REFACTOR | skip REFACTOR assessment | production code before failing test | test after code works ("for coverage") | `.skip`/`.todo`/`xtest`/comment-out | suppress output | add comments on non-complex logic | leave TODO/FIXME/commented-out code | refactor adjacent code | features the test didn't request | abstractions for single use | error handling for untested cases | skip scenario tests | skip granular tests when edge cases exist | continue with failing tests | commit code.

"Just this once" → STOP. Restart from RED.

---

## Integration With openspec-plus-apply

* **Subagent mode:** implementer subagent uses `skill` tool to load this skill before any code; if unavailable, reads this SKILL.md directly from the skills directory. Each subagent loads fresh in isolated context.
* **Inline mode:** main agent uses `skill` tool once at Phase 2 start; if unavailable, reads this SKILL.md directly from the skills directory. If already loaded in main agent context, reference instead of reload.

The slice's pre-mark gate (lint + format + tests + other on affected files) runs AFTER the TDD cycle completes for all tests AND every Gherkin scenario is covered. Gate failure → return to failing test's TDD cycle. Never bypass.

---

## Success Criteria

**Succeeds:** Pre-RED done, every Gherkin scenario has a passing test, every test observed to fail before passing, full per-test cycle completed before starting next, REFACTOR assessed and recorded for each, minimal production code, pristine output, no unnecessary comments, no adjacent refactoring, gate clean.

**Fails:** Pre-RED skipped, batching, next test before current REFACTOR complete, REFACTOR assessment skipped, production code without red-then-green test, tests skipped/commented, comments on obvious code, adjacent code touched, speculative abstractions, scenario paraphrased instead of translated, uncovered scenarios.
