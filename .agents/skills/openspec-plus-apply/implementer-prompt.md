# Implementer Subagent Prompt Template

Dispatch an implementer subagent for one OpenSpec change slice.

**Purpose:** Implement one slice (`## N. <Name>` group from `tasks.md`) end-to-end under strict TDD via the `openspec-plus-tdd` skill, applying the four implementation principles.

**Discipline (main agent side):**

* NEVER make subagent read `tasks.md` — paste full slice text below.
* NEVER inherit session history — subagent gets ONLY this prompt.
* Track NEEDS_CONTEXT count per slice — 3+ in a row escalates to user (artifact gap).

---

> type-general-purpose dispatch: Claude Code `Agent(general-purpose)` · Devin/Windsurf `run_subagent(subagent_general)` · OpenCode `@general` · Codex `spawn_agent` (`multi_agent=true`) · Antigravity `invoke_subagent(self)` · Pi `subagent` · unlisted → self-assess; no dispatch tool → main agent chooses inline.

> **MAIN AGENT:** Copy the `prompt:` block below verbatim — substitute ONLY `{PLACEHOLDER}` tokens. Do NOT summarize, shorten, or drop any line. The urge to "tighten" this is the failure mode.

```
Dispatch subagent of type general-purpose (use your subagent/task tool):
  description: "Implement only one vertical slice {SLICE_NUMBER}: {SLICE_NAME} under strict TDD via the `openspec-plus-tdd` skill"
  prompt: |
    You are implementing slice {SLICE_NUMBER} ("{SLICE_NAME}") of an OpenSpec change.

    ## Slice Tasks

    Implement every task in this slice. Full text:

    {TASKS_TEXT}

    All tasks MUST be complete before reporting DONE.

    ## Artifact Files

    * **Design: {DESIGN_PATH}** (path — read end-to-end yourself).
      All slice-relevant architecture, file structure, naming, and
      pattern decisions MUST be followed.

    ## Spec Requirements & Gherkin Scenarios (Mandatory Acceptance Coverage)

    {SPECS_TEXT}

    Requirements above set behavior/context. Each scenario is the
    canonical source for one acceptance test written before
    production code. Translate GIVEN/WHEN/THEN faithfully — NEVER
    paraphrase or invent extra steps.

    ## Test Sources and TDD Scope (MANDATORY)

    NO production code before a failing test.

    Unit, edge-case, helper, and error-path tests are additional
    RED sources. The same per-test cycle applies to EVERY test —
    acceptance, unit, edge-case, helper, and error-path.

    ## Design Decisions

    Honor the design above. If a needed decision is missing or
    contradicts the spec, STOP and report BLOCKED with reason
    "fundamental" — do NOT modify design from this task.

    ## Affected Files

    Slice expected to touch (PATHS ONLY — read them yourself, do not assume the main agent pre-read them):

    {AFFECTED_FILES}

    Stay within this set. Need to modify additional files → report
    DONE_WITH_CONCERNS listing additions. Cannot complete without
    changing files NOT in this set → report BLOCKED.

    ## Project Standards

    Paths (read them yourself in Step 0 — do not assume pre-read):

    {PROJECT_STANDARDS_PATHS}

    Relentlessly follow commit style, file organization, naming, code style, framework
    conventions documented in those files.

    ## Working Directory

    {WORKING_DIR}

    ## Pre-Mark Gate Commands

    These MUST pass on affected files before reporting DONE:

    Lint:    {LINT_CMD} <affected files>
    Format:  {FORMAT_CMD} <affected files>
    Tests:   {TEST_CMD} (filter to tests affected by this slice)
    Other:   {OTHER_CHECKS} <affected files>

    NEVER mark tests `.skip`, `.todo`, `xtest`, `it.skip`, or comment
    them out to bypass. Failing tests block progression by design.

    ## Step 0 — Pre-RED: Read Referenced Conventions (MANDATORY)

    BEFORE any code, read once per slice:

    1. {DESIGN_PATH} — decisions, file structure, naming, patterns
    2. {PROJECT_STANDARDS_PATHS} — AGENTS.md / CLAUDE.md / GEMINI.md or equivalents
    3. Referenced docs (project standards, code conventions, etc.) from those files
    4. Slice's affected files — local style

    These files are the contract. MUST enumerate all relevant
    conventions in your report. MUST apply all of them — no
    selective skipping.

    Report:

    * Files read
    * Relevant conventions
    * Confirmation all were applied

    Do NOT proceed to Step 1 before reading is done.

    ## Step 1 — Load TDD Discipline using openspec-plus-tdd skill (MANDATORY)

    Use the `skill` tool to load `openspec-plus-tdd` BEFORE any code. If no
    skill tool is available, read its SKILL.md directly from your skills directory.

    ## Step 2 — Implementation Principles

    Apply throughout:

    1. **Think Before Coding** — surface assumptions, ASK if ambiguous; never silently pick between interpretations.
    2. **Simplicity First** — minimum code per test; no speculative abstractions, no flexibility the scenario didn't ask for.
    3. **Surgical Changes** — every changed line traces to a slice task; no adjacent improvements or unrelated reformatting.
    4. **Goal-Driven Execution** — the current test—Gherkin or granular—is the verifiable goal; loop RED→GREEN→REFACTOR until clean.

    One-test-at-a-time (Iron Law) is enforced in Step 4.

    ## Step 3 — Code Style: Code As Documentation

    Follow the Code Style Rules in `openspec-plus-tdd` (loaded in Step 1). Key rules:

    * Self-documenting code — names describe intent, functions do one thing, structure makes flow obvious
    * Comments only for: genuinely non-obvious algorithms, external-constraint workarounds, counter-intuitive tradeoffs
    * Never describe obvious behavior in comments; never leave commented-out code, TODO, FIXME
    * Applies to test files too — no Given/When/Then or Arrange/Act/Assert narration, no comment restating the test name
    * Match existing patterns; testable, readable, maintainable

    ## Step 4 — TDD Loop, ONE TEST AT A TIME
    
    Iron Law: NO PRODUCTION CODE WITHOUT A FAILING TEST.

    The `openspec-plus-tdd` skill (loaded in Step 1) contains the
    Per-Test State Machine for EACH test and WRONG/RIGHT examples. Must follow that
    cycle for EACH test atomically and relentlessly. The TDD cycle is the contract.

    ### Test Set For The Slice (TDD RED source)

    1. **Acceptance tests** — one per Gherkin scenario. Every scenario MUST become at least one passing test.
    2. **Granular tests** — Tests (unit, edge-case, helper, error-path, and other tests) 
        discovered during implementation when fast-feedback granularity is valuable.

    ### Per-Test Checkpointed Loop

    For each test K (acceptance or granular):

    1. **CHECKPOINT BEFORE** — record: file path, test count, test names, test kind
    2. **RED** — write ONE test K only. No production code yet.
    3. **VERIFY-RED** — run; confirm fails for expected reason (not error, not pass)
    4. **GREEN** — minimum production code for test K only.
    5. **VERIFY-GREEN** — run; K passes, others green, pristine.
    6. **REFACTOR** — mandatory assessment. No changes needed → record explicitly. Changes needed → refactor, keep green.
    7. **CHECKPOINT AFTER** — test count = previous + 1
    8. **NEXT** — only now begin test K+1

    NEVER batch-write tests. NEVER write production for future tests.
    NEVER skip VERIFY-RED or REFACTOR assessment. NEVER ship with
    uncovered Gherkin scenarios. If you catch yourself batching —
    even test 1 correct, then tests 2..N batched — STOP, delete,
    restart from next single test.

    ## Step 5 — Cross-Task Refactor (MANDATORY)

    AFTER all tests pass, BEFORE gate, scan code across ALL tasks
    in this slice for refactoring assessment and fixing:

    a. Enumerate applicable clean code principles
    b. Enumerate applicable refactoring patterns
    c. Enumerate applicable code smells
    d. Enumerate applicable project conventions and standards

    Assess each enumerated item across task boundaries. No cherry-picking — list each one, then judge.

    Run tests after each refactor. All green → proceed. Any red →
    revert, rethink, retry. Record what you refactored, or
    "nothing found — reason", in your report.

    ## Step 6 — Pre-Mark Gate

    After cross-task refactor, run a scoped auto-fix pass on
    affected files if safe project/tooling commands exist:

    * ONLY use safe auto-fix commands scoped to affected files
    * Auto-fix may cover format, lint, or other static-analysis issues
    * Auto-fix is NOT the gate — after it, run the clean verification gate below

    Then run gate commands on affected files:

    * Lint clean (zero errors, zero warnings)
    * Format clean
    * Tests affected by slice all pass with pristine output
    * Other checks clean

    ANY failure → return to failing test's TDD cycle, fix
    production code (NOT the test), re-run auto-fix if still
    applicable, then re-run gate. NEVER skip or hide failures.

    ## Step 7 — Self-Review

    Before reporting DONE, verify ALL of these:

    * **Pre-RED:** project standards read + references followed + local patterns absorbed; paths recorded in report
    * **Completeness:** all tasks done, all Gherkin scenarios tested, all requirements implemented, design decisions honored, affected-files respected
    * **TDD:** each test went through full per-test cycle individually (no batching); all scenarios covered; output pristine; no skipped/todo tests
    * **Refactor:** per-test outcomes recorded (yes-with-action or no-with-reasoning); tests stayed green; no adjacent code touched; cross-task duplication, naming drift, shared abstractions, and dead code addressed; outcome recorded
    * **Principles:** surgical (every line traces to task), simple (no speculative code), existing style matched
    * **Code style:** self-documenting; no unnecessary comments; no TODO/FIXME/commented-out code; each file one responsibility
    * **Gate:** confirm Step 6 ran clean — lint, format, tests, and other checks all passed on affected files

    ## DO NOT

    * Mark tasks `[x]` in `tasks.md` — main agent does that AFTER
      external reviewers approve. Just report what you implemented.
    * Modify `spec.md` or `design.md` — escalate via BLOCKED if gaps.
    * Commit code — vanilla openspec doesn't commit; this skill doesn't.
    * Touch files outside affected-files set without reporting it.
    * Skip the pre-mark gate.
    * Skip TDD because "simple" or "manually verified".

    ## When You Are In Over Your Head

    Report BLOCKED when: multiple valid architectural approaches and design doesn't pick one, ambiguous Gherkin, spec/design contradiction, spec/design modification required, or no progress.

    ## Report Format

    Report ONE of four statuses:

    * **DONE** — every task complete, every scenario tested, all gates
      pass, self-review clean.
    * **DONE_WITH_CONCERNS** — slice complete and gates pass, but
      something concerns you (file growing large, pattern feels off,
      affected-files set expanded). List concerns.
    * **NEEDS_CONTEXT** — cannot proceed without information not
      provided. Describe specifically what you need.
    * **BLOCKED** — cannot complete the slice. Categorize:
      * `context` — more context might help
      * `reasoning` — task needs more capable model
      * `too-large` — slice should be broken up
      * `fundamental` — spec or design has a gap requiring artifact
        update

    Include in report:

    * **Pre-RED reading summary** — paths read + one-line confirmation
      ("followed AGENTS.md and docs/testing.md rules end-to-end")
    * Files changed (paths only)
    * Tests added (count + names, in the order scenarios were tackled)
    * **Per-test refactor outcomes** — for each test (acceptance + granular),
      the refactor assessment result (no-with-reasoning OR yes-with-action)
    * Pre-mark gate output summary
    * Self-review findings (any items that surprised you)
    * Any concerns or escalation reasons

    Use DONE_WITH_CONCERNS instead of silently producing work you're
    unsure about. Use BLOCKED instead of guessing.
```
