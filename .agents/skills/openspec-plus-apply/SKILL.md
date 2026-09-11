---
name: openspec-plus-apply
description: "MANDATORY skill that activates whenever the OpenSpec apply phase begins. Triggers: /opsx-apply runs, the openspec-apply-change vanilla skill is referenced or active, `openspec instructions apply` is invoked, or the user asks to implement, apply, execute, or build out an OpenSpec change ('implement the change', 'apply tasks', 'execute change', 'build out the change'). Takes over only vanilla step 6 (implementation loop) and emulates step 7 output (final status)."
metadata:
  version: 1.6.0
  priority: high
  disable-user-invocation: true
---

# OpenSpec Plus Apply

## Mission

Wrap vanilla `/opsx-apply` step 6 (task implementation loop) with subagent orchestration, two-stage review, strict TDD via `openspec-plus-tdd`, four implementation principles, conditional parallelism over a dependency-validated graph, per-slice + final lint/format/test/other gates, never-ignore-failures rule.

Vanilla owns:

* Step 1 — Select change
* Step 2 — `openspec status --change <name> --json`
* Step 3 — `openspec instructions apply --change <name> --json`
* Step 4 — Read context files
* Step 5 — Show progress
* Step 7 — Final status (this skill EMULATES output, not invokes vanilla)

This skill takes over Step 6 only.

---

> **RIGID. NEVER skip Phase 0 mode question when `settings.apply.executionMode` is `ask`. NEVER read affected source code files into the main agent's context — pass PATHS to subagents who read them. NEVER mark a slice `[x]` while a relevant test is failing or skipped. NEVER run code-quality review before spec-compliance ✅. NEVER dispatch parallel implementer subagents within a change without dependency analysis + user confirmation. NEVER modify spec/design from inside the implementer — escalate to plus-design / plus-spec. NEVER commit code. NEVER take over vanilla steps 1-5 or invoke openspec-verify-change or auto-trigger /opsx-archive.**

**Red flags — STOP, you are about to violate this skill:**

- "Re-read the artifacts to be safe" — vanilla loaded them, in context
- "Let me read the affected source files before dispatching the subagent" — NO (subagent mode). Pass paths. Subagent reads. (Inline mode is different — main agent reads directly.)
- "Critique the spec/design before starting" — done; trust them
- "Fix the spec/design from here while implementing" — escalate, don't edit
- "Mark the test `.skip` to move on" — never. Failing tests block progress
- "Slice is small, skip the spec-compliance reviewer" — never
- "Implementer's report says DONE, good enough" — verify with reviewer subagents
- "Both reviewers found issues, fix simultaneously" — spec-compliance first, re-review, then code-quality
- "Reviewing is generic, I'll write the review prompt myself" — NO. Every reviewer dispatch reads its `*-prompt.md` and sends the body verbatim; improvising it is the #1 failure mode
- "This prompt is long, I'll tighten it / drop the obvious parts when passing it on" — NO. Copy the template's `prompt:` block verbatim, substitute only `{PLACEHOLDER}`s. Every bullet is load-bearing; simplifying silently strips instructions from the subagent
- "Should I continue to the next slice?" — never ask. Continuous unless BLOCKED
- "Three fixes failed, try a fourth" — STOP. Artifacts wrong. Pause and exit
- "Run all tests instead of just affected" — wastes time; per-slice gate is scoped
- "Commit as I go" — vanilla doesn't commit; we don't either
- "Run /opsx-archive once everything is done" — only suggest, never auto-trigger
- "Run openspec-verify-change as part of final gate" — out of scope
- "Implementer wrote all tests upfront, that's fine" — TDD violation. One test at a time.
- "Implementer skipped REFACTOR because nothing to refactor" — refactor assessment is mandatory; skipping the assessment is the violation, not the skipping of action
- "AGENTS.md has many rules, implementer applied the ones that felt relevant" — cherry-picking the project's documented rules. Follow them strictly, end-to-end.
- "Implementer did test 1 RED-GREEN-REFACTOR correctly, then batched the rest" — the per-test state machine applies to EVERY test, not just the first
- "Implementer only wrote acceptance tests (one per Gherkin scenario), skipped all unit/edge tests" — acceptance coverage is mandatory but granular tests are encouraged for non-trivial branches, edges, error paths

None justify re-reading artifacts, reading affected code into main context, editing spec/design, ignored failures, skipped reviewers, batched fixes, per-slice check-ins, fix-attempt #4, full-suite tests, commits, archive auto-trigger, verify-change handoff, batched scenario tests, skipped REFACTOR assessment, or cherry-picked project rules.

---

## Inputs

Already in conversation context (loaded by vanilla `/opsx-apply` steps 2-4):

* `proposal.md`, `spec.md` (or `specs/**/*.md`), `design.md`, `tasks.md`
* Output of `openspec status --change <name> --json`
* Output of `openspec instructions apply --change <name> --json`

Project standards (`AGENTS.md` / `CLAUDE.md` / `GEMINI.md`) — opencode auto-loads as project instructions; authoritative for code style, file organization, naming, build/test/lint commands.

Build files (`package.json`, `Cargo.toml`, `go.mod`, etc.) — source of lint/format/test commands for gates.

If artifacts NOT in context, surface error:

```
OpenSpec change artifacts are not in context. Run `/opsx-apply <change>` to load
them, then I will take over the implementation loop.
```

NEVER read source code outside what slices touch.

---

## Workflow

Four phases (0-3) plus hand-back. NEVER skip, merge, or reorder.

```text
[ ] Phase 0: Pre-Flight (verify inputs, discover commands, choose mode)
[ ] Phase 1: Dependency analysis & parallelism check
[ ] Phase 2: Per-slice execution loop
[ ] Phase 3: Whole-change review & final gate
```

**Flow:** Phase 0 → Phase 1 → (blocked? → suggest /opsx-continue, exit) → (all_done? → suggest /opsx-archive, exit) → Phase 2 per-slice: implementer → (DONE → spec-compliance review → code-quality review → per-slice gate → mark [x]) | (NEEDS_CONTEXT → provide, re-dispatch) | (BLOCKED → 4-step assessment → recoverable: retry | fundamental/3+ retries: pause and exit). After all slices → Phase 3: whole-change review + final gate → emulate vanilla step 7.

Reviewers cap at 3 fix cycles. Beyond 3 → STOP, artifacts wrong, escalate.

---

## Workflow Visibility (MANDATORY)

Display workflow phases via todowrite at start; update as phases complete. Add per-slice todos under Phase 2 once slicing is finalized.

---

## Core Principles

- **Take Over Step 6 Only** — vanilla owns steps 1-5 and 7; after Phase 3, emulate step 7 output and stop.
- **Trust The Artifacts** — earlier phases validated them; NEVER re-critique; implementation reveals a gap → escalate (pause/exit, suggest plus-design/plus-spec); never edit from inside this skill.

### Main Agent Context Hygiene (subagent mode ONLY)

In subagent mode, the main agent NEVER reads slice's affected source files — pass PATHS only. Subagents read files in their isolated context:

* Implementer: pass affected file paths. Implementer reads them.
* Spec-compliance reviewer: pass paths to changed files. Reviewer reads them.
* Code-quality reviewer: pass changed file paths. Reviewer diffs via `git diff HEAD`; falls back to reading files if git unavailable.
* Final reviewer: pass artifact paths + all changed file paths. Same diff strategy.

Artifacts (`proposal.md`, `spec.md`, `design.md`, `tasks.md`) are already in main context from vanilla — using them for subagent prompt excerpts is fine. Source code files are NOT.

- **Inline Mode Reading** (inline mode only) — main agent IS the implementer; reads files directly; reads project standards + affected files for Pre-RED; edits directly during TDD; self-reviews own changes; reviews cumulative diff at Phase 3. Each file read once.
- **Pre-RED Reading** (both modes) — before any code, implementer reads project standards (`AGENTS.md`/`CLAUDE.md`/`GEMINI.md` + referenced docs); enforced by `openspec-plus-tdd` Phase 0; follow every documented rule strictly, end-to-end (no cherry-picking).

### Strict TDD Through openspec-plus-tdd

**Iron Law: NO PRODUCTION CODE WITHOUT A FAILING TEST.** One test at a time — full cycle (RED → VERIFY-RED → GREEN → VERIFY-GREEN → REFACTOR) before starting next. Applies to all tests: acceptance (mandatory per Gherkin scenario) and granular (encouraged). REFACTOR assessment mandatory per test — skipping assessment is the violation, not skipping action.

### Implementation Principles For Both Modes

Apply in every slice:

1. **Think Before Coding** — surface assumptions, ask when uncertain, never silently pick between interpretations.
2. **Simplicity First** — minimum code that solves the problem; nothing speculative.
3. **Surgical Changes** — every changed line traces to a task in this slice; never improve adjacent code.
4. **Goal-Driven Execution** — each test—Gherkin or granular—is a verifiable goal; loop independently until it passes.

- **Code Style** — canonical rules in `openspec-plus-tdd` § Code Style Rules; code is self-documenting; comments are exceptional last resort.
- **Never Ignore Failing Tests** — `.skip`, `.todo`, `xtest`, `it.skip`, commented-out tests, or `--bail` shortcuts are anti-patterns; pre-mark gate blocks on any failure.
- **Continuous Execution** — do NOT pause between slices; pause only on BLOCKED (unrecoverable), 3+ failed fix cycles, or 3+ NEEDS_CONTEXT in a row.
- **No Commits** — vanilla doesn't commit; this skill doesn't either.
- **Resumability Is Free** — `openspec instructions apply` is the source of truth; re-running picks up at next pending slice.
- **Respect Project Standards** — `AGENTS.md`/`CLAUDE.md`/`GEMINI.md` capture lint/format/test commands, code style, file organization, naming conventions.

---

## Phase 0: Pre-Flight

### 0.0 Resolve Apply Config

Read `openspec/.plus/config.yaml` once (missing/unreadable/unrecognized value → default, zero behavior change): `settings.apply.executionMode` (`ask` default | `subagent` | `inline`, anything else → `ask` — skips 0.3 only for `subagent`/`inline`), `settings.apply.parallelism` (`ask` default | `always` | `never`, anything else → `ask` — skips 1.4 only for `always`/`never`).

### 0.1 Verify In-Context Artifacts

Confirm vanilla's steps 2-4 produced these in conversation context:

* Output of `openspec instructions apply --change <name> --json` showing `state`, `tasks`, `contextFiles`
* Content of `proposal.md`, `spec.md` / `specs/**/*.md`, `design.md`, `tasks.md`

If absent → surface error above, stop.

State handling:

* `"blocked"` → suggest `/opsx-continue`, exit
* `"all_done"` → suggest `/opsx-archive`, exit
* `"ready"` → continue

### 0.2 Discover Pre-Mark Gate Commands

Identify ONCE at Phase 0 (used by per-slice + final gates):

* **Lint** — e.g., `bun run lint`
* **Format** — e.g., `bun run format`
* **Test** — e.g., `bun test`
* **Other** — type-check (`tsc --noEmit`), schema-check, build-check, etc.

Discovery order:

1. `package.json` `"scripts"` (or `Cargo.toml`, `go.mod`, `pyproject.toml`)
2. `AGENTS.md` / `CLAUDE.md` / `GEMINI.md` documented commands
3. Project README or build file conventions

Missing or ambiguous → ask user ONCE via question tool.

### 0.3 Mode Question (MANDATORY, blocking)

If `settings.apply.executionMode` ≠ `ask` (from 0.0), skip this question — use the configured mode directly, state it in one line, proceed to Phase 1.

Otherwise, use question tool with ONE question. **Subagent mode is the recommended default for ALL changes — small, medium, and large.** Inline mode exists only as a fallback for environments where subagent dispatch is unavailable. Phrase the question so this preference is unmistakable.

```text
How should the change be implemented?

1. Subagent mode (Strongly Recommended — default for all changes) —
   fresh subagent per slice with two-stage review (spec-compliance then
   code-quality), isolated context per slice, main agent stays clean.
   This produces the highest quality output and is the recommended
   choice unless subagent dispatch is unavailable in your environment.

2. Inline mode (Fallback only) — main agent runs each slice with the
   same TDD discipline, implementation principles, and self-review
   checklists. Reviews are self-checks rather than independent
   reviewer subagents. Use only when subagent dispatch is unavailable
   or when you have an explicit reason to avoid it.
```

Both modes apply SAME discipline (TDD, principles, never-ignore-failures, per-slice gate, final gate). Subagent mode adds isolated reviewers; inline mode runs reviews as self-check checklists.

If the user does not pick explicitly and provides no reason to prefer Inline, default to Subagent mode and proceed.

Record choice. Display Phase 0-3 todowrite plan.

---

## Phase 1: Dependency Analysis

### 1.1 Build Slice Graph

For each slice (`## N. <Name>`) in `tasks.md`:

* **Affected files** — infer from task descriptions + `design.md` "Component Structure" + spec terminology
* **Depends-on slices** — from spec/design ordering + task descriptions

Build directed graph: `slice_n -> slice_m` if `n` depends on `m`.

### 1.2 Validate File Existence

For each inferred file path:

```bash
git log --oneline -- <inferred-file>
```

Detect:

* File exists (recent commits) → safe to modify
* File doesn't exist → slice will create; note for implementer
* File touched by multiple slices → collision; mark for sequential

### 1.3 Identify Parallelizable Groups

Slices A, B parallelizable when:

* No dependency edge between them
* Disjoint affected-file sets

### 1.4 Ask About Parallelism (only if independent group exists)

If `settings.apply.parallelism` ≠ `ask` (from 0.0), skip this question — `always` dispatches all independent groups in parallel, `never` runs sequentially, without asking.

```text
Slices [A, B, ...] appear independent (no shared files, no dependency
between them).

1. Dispatch in parallel (Recommended) — faster, isolated subagents
2. Run sequentially — one at a time
```

Parallel: dispatch all members concurrently. Wait for ALL to return before reviewing any. Parallel slice BLOCKED while siblings still running → wait, then handle BLOCKED.

Sequential: dependency order.

### 1.5 Final Slice Order

Output: `[slice_1, slice_2, ..., (parallel_group_X), ..., slice_N]`. Add per-slice todos under Phase 2 in todowrite.

---

## Phase 2: Per-Slice Execution Loop

For each slice (or parallel group):

### Subagent Dispatch Contract (MANDATORY — 2.A.1, 2.A.3, 2.A.4, 3.1)

Before EVERY dispatch (implementer AND reviewers): **READ the referenced `*-prompt.md`** (resolve `./` next to this SKILL.md, not the CWD). The subagent prompt is that file's fenced `prompt:` block — **copy it verbatim**, replacing ONLY `{PLACEHOLDER}` tokens. You are transcribing a payload, not re-authoring one: do NOT summarize, condense, shorten, paraphrase, reorder, or drop any line, bullet, or section — the length is deliberate and load-bearing. NEVER reconstruct a prompt from memory — reviewer prompts are NOT generic. The urge to "tighten" or "drop the obvious parts" is the failure mode. The placeholder lists below say WHAT to substitute, not what to send.

### 2.A Subagent Mode

#### 2.A.1 Dispatch Implementer

Per the Dispatch Contract, **read `./implementer-prompt.md` NOW, send verbatim.** Fill placeholders:

* `{SLICE_NUMBER}`, `{SLICE_NAME}`
* `{TASKS_TEXT}` — full text of all `N.M` tasks
* `{SPECS_TEXT}` — full text of slice-relevant scenarios from specs (requirements + Gherkin)
* `{DESIGN_PATH}` — **path only** to `design.md` (from `contextFiles` in vanilla step 2-4 output). Implementer reads it end-to-end.
* `{AFFECTED_FILES}` — **paths only** (e.g., `src/auth/login.ts`, `tests/auth/login.test.ts`). Implementer reads them.
* `{PROJECT_STANDARDS_PATHS}` — **paths only** to `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, and any docs they reference (e.g., `docs/coding-standards.md`, `docs/testing.md`, `CONTRIBUTING.md`, `docs/patterns/*`). Implementer reads them in Step 0 (Pre-RED).
* `{LINT_CMD}`, `{FORMAT_CMD}`, `{TEST_CMD}`, `{OTHER_CHECKS}` — discovered at Phase 0
* `{WORKING_DIR}`

NEVER make subagent read `tasks.md` and `specs/**/*.md` — paste full slice text. NEVER inherit session history. For design/affected/standards files, pass paths only; implementer reads them.

#### 2.A.2 Handle Implementer Status

* **DONE** → spec-compliance review.
* **DONE_WITH_CONCERNS** → read concerns. Correctness/scope → address before review. Otherwise note + proceed.
* **NEEDS_CONTEXT** → provide missing context, re-dispatch same model. 3+ in a row → escalate to user (fundamental gap, pause and exit).
* **BLOCKED** → 4-step assessment:
  1. Context → re-dispatch same model.
  2. Reasoning → re-dispatch capable model.
  3. Slice too large → break into per-task subagents.
  4. Fundamental → pause and exit, suggest plus-design / plus-spec.

Never silently retry without changing something.

#### 2.A.3 Spec-Compliance Review

Per the Dispatch Contract, **read `./spec-compliance-reviewer-prompt.md` NOW, send verbatim.** Fill placeholders:

* `{SLICE_NUMBER}`, `{SLICE_NAME}`
* `{TASKS_TEXT}` — full text of all `N.M` tasks for this slice
* `{IMPLEMENTER_REPORT}` — what the implementer reported after completing the slice
* `{CHANGED_FILE_PATHS}` — **paths only** to files the implementer modified or created
* `{PROPOSAL_PATH}` — **path only** to `proposal.md` (from `contextFiles` in vanilla step 2-4 output)
* `{SPECS_TEXT}` — full text of slice-relevant scenarios from specs (only those containing requirements and Gherkin scenarios that map to this slice's tasks)
* `{DESIGN_PATH}` — **path only** to `design.md` (from `contextFiles` in vanilla step 2-4 output)

NEVER read artifact or changed files in main context — pass paths only; reviewer reads in isolated context.

**OpenSpec-contract scope only** — does NOT verify project-rule compliance, naming style, comments, or lint/format. Those belong to code-quality reviewer (2.A.4).

Returns ✅ Compliant OR ❌ Issues: `Task-Incomplete`, `Missing-Requirement`, `Missing-Scenario`, `Out-of-Scope`, `Design-Violation`, `TDD-Discipline`. ❌ → implementer fixes → run per-slice gate (2.A.5) on affected files → re-dispatch reviewer. Cap 3 cycles → STOP, pause and exit.

#### 2.A.4 Code-Quality Review

ONLY after spec-compliance ✅. Per the Dispatch Contract, **read `./code-quality-reviewer-prompt.md` NOW, send verbatim.** Fill placeholders:

* `{SLICE_NUMBER}`, `{SLICE_NAME}`
* `{TASKS_TEXT}` — full text of all `N.M` tasks for this slice
* `{IMPLEMENTER_REPORT}` — what the implementer reported after completing the slice
* `{DESIGN_PATH}` — **path only** to `design.md` (from `contextFiles` in vanilla step 2-4 output); reviewer reads fully for architecture, file structure, naming, patterns
* `{SPECS_TEXT}` — full text of slice-relevant scenarios from specs (only those containing requirements and Gherkin scenarios that map to this slice's tasks)
* `{CHANGED_FILE_PATHS}` — **paths only** to files the implementer modified or created; reviewer diffs via `git diff HEAD` (no git → reads files directly)
* `{STANDARDS_DOC_PATHS}` — **paths only** to `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, and referenced docs

NEVER read artifact, changed, or standards files in main context — pass paths only; reviewer reads in isolated context.

Applies implementation principles + code-quality concerns + project-rule compliance. Returns Strengths / Issues (Critical/Important/Minor) / Assessment. ❌ → implementer fixes → run per-slice gate (2.A.5) on affected files → re-dispatch reviewer. Cap 3 cycles → STOP, pause and exit.

#### 2.A.5 Per-Slice Gate

Run for affected files of THIS slice:

```bash
{LINT_CMD}     <affected-files>
{FORMAT_CMD}   <affected-files>
{TEST_CMD}     --filter <tests-affected-by-slice>
{OTHER_CHECKS} <affected-files>
```

ANY failure:

* NEVER `.skip`/`.todo`/`xtest`/`it.skip`/comment-out to bypass
* Dispatch implementer with specific failure
* Re-run gate

Cap 3 cycles → STOP, pause and exit.

#### 2.A.6 Mark Slice Complete

Edit `tasks.md`:

```diff
- - [ ] N.1 <task>
- - [ ] N.2 <task>
+ - [x] N.1 <task>
+ - [x] N.2 <task>
```

NEVER mark `[x]` before all gates pass.

### 2.B Inline Mode

Same gates and discipline as subagent mode, run by the main agent. The main agent IS the implementer — it reads files directly (no path-passing indirection, no subagent isolation, no duplicate reads).

#### 2.B.0 Pre-RED Reading (MANDATORY)

Before any code, read: (1) project standards (`AGENTS.md` / `CLAUDE.md` / `GEMINI.md` + referenced docs), (2) slice's affected files for local patterns. After reading, MUST enumerate in TodoWrite every convention relevant to this slice's scope; MUST apply all of them — no selective skipping. NEVER proceed before reading is done.

#### 2.B.1 Load TDD Skill

Use `skill` tool to load `openspec-plus-tdd` once at Phase 2 start. If already loaded in session, reference instead of reload. If no skill tool is available, read its SKILL.md directly from your skills directory.

#### 2.B.2 Per-Slice TDD Cycle (ONE TEST AT A TIME)

Follow `openspec-plus-tdd` Per-Test State Machine: pick ONE test (acceptance from Gherkin = mandatory, or granular = encouraged), complete full cycle (RED → VERIFY-RED → GREEN → VERIFY-GREEN → REFACTOR assessment → NEXT) before starting next. NEVER batch. Record per-test refactor outcome. Done when every Gherkin scenario has a passing test AND all granular tests pass. Apply four implementation principles + code-as-documentation throughout.

#### 2.B.2.1 Cross-Task Refactor (MANDATORY)

AFTER all tests pass, BEFORE self-reviews. Scan ALL code across ALL tasks in this slice: duplicate logic → extract shared utility; naming drift → unify; shared abstractions emerged → consolidate; dead code from earlier tasks superseded → remove. Run tests after every refactor. Record outcome (or "nothing found — reason").

#### 2.B.3 Inline Self-Reviews

Run three checklists in order (mirroring subagent reviewers). ALWAYS re-read `proposal.md`, slice-relevant `spec.md`/`specs/**/*.md`, and `design.md` before each self-check — do not rely on memory.

* **Spec-compliance self-check** — re-read proposal, slice-relevant spec file(s), and design end-to-end. Verify: every spec requirement implemented; every Gherkin scenario has a passing test; no work beyond proposal scope; design decisions honored; file structure matches design; naming consistent with artifact terminology.
* **TDD-Discipline self-check** — verify each test completed the full per-test cycle individually (no batching); per-test refactor outcomes recorded with reasoning; all Gherkin scenarios covered. Violation → redo the slice with strict per-test discipline.
* **Code-quality self-check** — re-read design for architecture, file structure, naming, and patterns. Verify against every convention enumerated in 2.B.0 — no selective skipping. Check: cross-task refactoring done (no duplicate logic, naming drift, missed shared abstractions, or dead code across tasks); surgical changes (every line traces to slice tasks); no speculative abstractions; **code is self-documenting (comments only for genuinely non-obvious logic / external-constraint workaround / counter-intuitive tradeoff; refactor attempted first)**; no commented-out code or TODO/FIXME markers; single responsibility; naming clear; decomposition matches design; per-test and cross-task refactor outcomes recorded.

Issues found → fix inline → run per-slice gate (2.B.4) on affected files → re-check. Cap 3 cycles per checklist → STOP, pause and exit.

#### 2.B.4 Per-Slice Gate

Same as 2.A.5. Never ignore failures.

#### 2.B.5 Mark Slice Complete

Same as 2.A.6.

### 2.C Continuous Execution

After marking `[x]`, IMMEDIATELY proceed to next slice (or parallel group). NEVER ask "should I continue?".

Pause only on:

* BLOCKED with no recovery path
* 3+ failed fix cycles on any reviewer or gate
* 3+ NEEDS_CONTEXT in a row from same implementer

Each pause exits Phase 2, surfaces context, suggests plus-* update.

---

## Phase 3: Whole-Change Final Review

### 3.1 Whole-Change Code Review

After all slices `[x]`:

Per the Dispatch Contract, **read `./final-review-prompt.md` NOW, send verbatim.** Pass: change name/schema, all artifact paths, all changed file paths. Main agent does NOT pre-read the diff. Reviewer diffs via `git diff HEAD` (no git → reads files directly). Verifies cross-slice integration, surfaces design issues (advisory), applies principles whole-change. Does NOT do spec/design alignment (that's `openspec-verify-change`).

Issues: small/local → fix inline; large/cross-slice → re-dispatch implementer. Cap 3 cycles → STOP, pause and exit.

### 3.2 Final Cumulative Gate

Run for ALL files affected by ALL slices:

```bash
{LINT_CMD}     <all-affected-files>
{FORMAT_CMD}   <all-affected-files>
{TEST_CMD}     --filter <tests-affected-by-any-slice>
{OTHER_CHECKS} <all-affected-files>
```

ANY failure:

* NEVER skip or ignore
* Dispatch implementer (or fix inline) for specific failure
* Re-run final gate

Cap 3 cycles → STOP, pause and exit (likely cross-slice integration problem).

### 3.3 On Phase 3 Success

Proceed to hand-back.

---

## Hand-Back: Emulate Vanilla Step 7

Produce step 7 output after Phase 3. Do NOT invoke vanilla.

### Success format

```text
## Implementation Complete

**Change:** <change-name>
**Schema:** <schema-name>
**Progress:** N/N tasks complete ✓

### Completed This Session
- [x] <task 1.1>
- [x] <task 1.2>
- [x] <task 2.1>
...

All tasks complete. Run `/opsx-archive` to sync specs and finalize the change.
```

### Pause format

```text
## Implementation Paused

**Change:** <change-name>
**Schema:** <schema-name>
**Progress:** M/N tasks complete

### Pause Reason
<specific reason — BLOCKED / 3+ failed fixes / 3+ NEEDS_CONTEXT / fundamental artifact gap>

### Suggested Next Step
<plus-design / plus-spec / plus-tasks update / user clarification>

### Resume
Re-run `/opsx-apply <change-name>` after addressing the cause. OpenSpec
state preserved via task checkboxes; next pending slice picks up.
```

NEVER auto-trigger `/opsx-archive`. NEVER invoke `openspec-verify-change`.

---

## Failure Recovery & Resumability

| Failure | Recovery |
|---|---|
| Artifacts not in context at Phase 0 | Surface error, ask user to run `/opsx-apply <change>` first, exit |
| Implementer NEEDS_CONTEXT | Provide context, re-dispatch (same model). 3+ in a row → escalate |
| Implementer BLOCKED `context` | Provide more context, re-dispatch same model |
| Implementer BLOCKED `reasoning` | Re-dispatch with capable model |
| Implementer BLOCKED `too-large` | Break into per-task subagents |
| Implementer BLOCKED `fundamental` | Pause and exit, suggest plus-design / plus-spec |
| Spec-compliance ❌ × 1-2 | Implementer fixes, re-review |
| Spec-compliance ❌ × 3 | Pause and exit, suggest plus-spec / plus-design (or plus-tasks if a task description was unclear/wrong) |
| Code-quality ❌ × 1-2 | Implementer fixes, re-review |
| Code-quality ❌ × 3 | Pause and exit, suggest plus-design |
| Per-slice gate fails | Implementer fixes, re-run. NEVER skip/ignore |
| Per-slice gate ❌ × 3 | Pause and exit |
| Final reviewer (Phase 3) fails | Fix inline (small) or re-dispatch implementer (large), re-review |
| Final reviewer ❌ × 3 | Pause and exit, cross-slice integration problem |
| Final cumulative gate fails | Implementer fixes, re-run gate |
| Final gate ❌ × 3 | Pause and exit |
| Test fails repeatedly | Apply systematic-debugging Phase 1 (root cause), not more fixes |
| Verification not run | STOP per `verification-before-completion` (no claims without fresh evidence) |
| Session interrupted mid-implementation | Re-run `/opsx-apply` — state preserved via checkboxes |
| Wrong slice manually marked `[x]` | User edits tasks.md back, re-runs |
| Parallel slice BLOCKED while siblings run | Wait for siblings, then handle BLOCKED |

Resumability is free via OpenSpec CLI — Phase 0 always reads state fresh.

---

## Anti-Patterns

NEVER:

* Take over vanilla steps 1-5 or 7 (emulate 7) | invoke `openspec-verify-change` | auto-trigger `/opsx-archive` | skip mode question
* **Subagent mode:** read source files or project standards into main context — pass paths only
* **Inline mode:** re-read files already in context this slice
* Run code-quality before spec-compliance ✅ | mark `[x]` before gate passes | self-review replacing reviewer subagent
* `.skip`/`.todo`/`xtest`/comment-out tests | batch-write tests | skip REFACTOR assessment | production code before failing test
* Cherry-pick project rules — follow all strictly | add comments on non-complex logic | refactor adjacent code | speculative abstractions
* Make subagent read `tasks.md` (paste text) | modify spec/design from implementer | continue past 3+ failures | commit code
* Dispatch parallel slices without dependency analysis + user confirmation | run full test suite for per-slice gate

---

## Success Criteria

**Succeeds:** Phase 0 verified artifacts + discovered gates; mode chosen explicitly; dependency-validated ordering; each slice: spec-compliance ✅ + code-quality ✅ + gate ✅ before `[x]`; no skipped tests; no artifact modifications; Phase 3 reviewer approved + final gate clean; vanilla step 7 output emulated.

**Fails:** code without verified artifacts; `[x]` before gates; 3+ fixes without questioning artifacts; wrong review order; implementer modified spec/design; auto-triggered archive; invoked verify-change; skipped tests; skipped final gate; invoked vanilla instead of emulating.

---

## What This Skill Does NOT Do

* Modify vanilla `/opsx-apply` command or `openspec-apply-change` skill (forbidden)
* Verify branch isolation (allowed on main/master)

---

## Integration

* **`openspec-plus-tdd`** — loaded via `skill` tool by implementer subagent + main agent (inline) before any code
