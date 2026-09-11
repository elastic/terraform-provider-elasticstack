---
name: openspec-plus-tasks
description: "MANDATORY skill that activates whenever the OpenSpec tasks phase begins. Triggers: /opsx-new or /opsx-continue runs; openspec-new-change, openspec-continue-change, or openspec-explore is active; `openspec instructions tasks` is invoked; or the user wants to create, update, review, refine, or discuss an OpenSpec tasks.md file."
metadata:
  version: 1.6.0
  priority: high
  disable-user-invocation: true
---

# OpenSpec Plus Tasks

## Mission

Enrich OpenSpec's tasks.md generation with vertical-slice discipline and subagent review. Format is owned by the schema template (resolved in Phase 0) — this skill provides thinking discipline before the file is written and a subagent review after.

Two anchors that shape every task group:

1. **Each task group = a vertical slice** — end-to-end through whatever layers are needed, not horizontal layer-by-layer
2. **Each task group completion = a testable outcome** — after the last task in a group is checked, the user can verify something works

---

> **This skill is RIGID. NEVER write tasks.md before completing Phase 1 vertical-slice discovery. NEVER skip the subagent review. Reading source code, writing implementation code, running TDD, including exact commit messages, or restructuring the template format is a skill violation.**

**Red flags — STOP, you are about to violate this skill:**

- "I'll group by layer (DB / API / UI) to keep similar tech together"
- "This task group is preparatory work, it doesn't need to be testable on its own"
- "I'll add a 'Files' or 'Dependencies' section to the tasks.md"
- "I should peek at the code to figure out what files to touch"
- "Let me put exact code or shell commands in the task description"
- "The reviewer subagent isn't needed, the tasks look fine to me"
- "Reviewer flagged minor issues, I'll re-dispatch after fixing to confirm"
- "I'll add commit instructions per task"
- "Let me write a TDD-ordered set of sub-tasks"

---

## Inputs

* **Proposal** — intent, scope boundaries, non-goals
* **Specs** — requirements with Gherkin scenarios
* **Design** — architecture decisions, file structure, integration points
* **Existing tasks.md files** in `openspec/changes/` — for naming convention only
* **NEVER read source code** — design phase already grounded patterns

If the design is missing detail you need to slice the work, the design is incomplete. Surface the gap to the user — do not paper over with code reading.

---

## Workflow

Three phases. NEVER skip or merge.

```text
Phase 0: Schema & Template Resolution
Phase 1: Vertical Slicing & Discovery
Phase 2: Write & Artifact Compliance Review
```

---

## Workflow Visibility (MANDATORY)

Display workflow phases via task tool at start; update as each phase completes.

---

## Core Principles

- **Vertical Slice** — each `## N.` heading is a vertical slice: a feature or capability that, when all its tasks are done, produces one end-to-end testable outcome expressible in a single sentence. The slice heading implies the outcome. A slice spans whatever layers are needed (DB, API, UI, CLI). Slices are dependency-ordered — each slice's prerequisites are satisfied by earlier slices.
- **Task** — each `- [ ]` item is a capability the implementer delivers in one work session. A task operates at the spec **requirement** level, not the **scenario** level. It names the capability and summarizes the behavioral facets the implementer must deliver. The spec's Gherkin scenarios are the implementer's test cases during TDD — the task points to the capability; the spec holds the individual test cases. "Done" is unambiguous from the task description alone.
- **WHAT, not HOW** — tasks state what behavior to deliver. The design document holds how: method signatures, struct fields, trait members, wiring steps, data layouts, file creation sequences. The implementer reads the design for architectural guidance and discovers code through TDD. File paths may appear inline to orient (e.g., "Add stream validation to `src/kinesis/handler.rs`") but never as step-by-step choreography.
- **Behavior only** — each task describes a behavior the system exhibits after implementation. Not infrastructure to create before any behavior uses it. Not a process step to run (gate, test suite, build, codegen command). Infrastructure emerges as the implementer delivers behavior through TDD. Verification is the implementer's responsibility per the TDD cycle.
- **Traceability** — every task traces to a spec requirement, design decision, or stakeholder need. No speculative tasks. Follow project conventions (read `AGENTS.md`, `CLAUDE.md`, equivalents at Phase 1 start; conflicts → surface to user, never silently override).

---

## Phase 0: Schema & Template Resolution

Run FIRST, before any slicing:

1. Run:
   ```bash
   openspec instructions tasks --change <name> --json
   ```
   Extract: `template` (structural authority; format you MUST follow), `instruction` (guidance on what content each section needs), `rules` (project constraints to honor). Parse template structure — if it includes metadata fields (priority, effort, labels, categories) beyond the default `## N. Name` + `- [ ] N.M desc` format, Phase 1 must collect that information during slicing; discovery questions MUST satisfy template requirements.
2. Read `openspec/.plus/config.yaml` (missing/unreadable/unrecognized → defaults): `settings.questionMode` (`sequential` default; `batch` groups step 5's clarifying questions into one round — see Phase 1 below).

---

## Phase 1: Pre-Write Discovery & Vertical Slicing (MANDATORY)

**Pre-existing answers:** If recent conversation already identifies vertical slices, names testable outcomes, or establishes dependency ordering — via prior exploration, a detailed initial request, or any other source — incorporate those and SKIP redundant clarification. Ask ONE question at a time only for unresolved slicing decisions (`batch` mode: present them together instead — see step 5). NEVER re-ask what's already established.

### 1. Read Inputs

Read in this order:

1. **Project-level instruction files first** — `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, equivalents at project root, `.claude/`, `.opencode/`, `docs/`. These capture conventions for commit style, branching, file organization, naming. Tasks MUST follow these.
2. Proposal, specs, design.
3. Existing tasks.md files in `openspec/changes/` for naming convention only.

NEVER read source code. If grounding is missing, the design is incomplete — surface that to the user.

### 2. Identify Vertical Slices

From spec requirements and design components, identify slices that deliver testable outcomes. Each slice spans whatever layers (DB, API, UI, CLI, background jobs) are needed.

For each candidate slice ask: "When all tasks in this group are done, what can the user observe or verify?" If "nothing observable", it's not a slice — re-slice or merge.

### 3. Name The Testable Outcome For Each Slice

For each slice, write the testable outcome in one plain sentence. Example: "User can log in with email/password and see the dashboard".

Cannot name in one sentence? The slice is wrong — re-slice or surface a design gap.

### 4. Order Slices By Dependency

Some slices depend on others (e.g., "user signup" before "user login"). Order so prerequisites are satisfied by earlier slices.

### 5. Clarifying Questions

If slicing is ambiguous, use **question tool**, ONE question at a time:

* "Which capability should be the first vertical slice?"
* "Is X part of slice A, or its own slice?"

In `sequential` mode (default), NEVER batch. NEVER assume. Always include your recommended slice boundary with rationale on every question — never a bare question without a recommendation. If the user's answer raises a follow-on slicing decision, resolve that branch before moving on (`batch`: next batch round).

In `batch` mode (`settings.questionMode: batch`), present all unresolved slicing questions together in one round instead — same real answers and recommendations.

### 6. Rules Compliance Check

Review `rules` from Phase 0. If any rule constrains task breakdown, verify the slices honor those constraints. If a rule is violated, surface the conflict to the user before proceeding (`batch`: fold multiple conflicts into one round).

Mark Phase 1 complete. Update task status. Proceed to Phase 2.

---

## Phase 2: Write & Artifact Compliance Review

### 2.1 Write tasks.md

Use the `template` and `outputPath` from Phase 0. Use the structure exactly. Do not restructure, add sections, or change numbering convention.

**Before writing — 2 mandatory steps:**

**Step 1 — Map Phase 1 slices to template:** Phase 1 slices and task descriptions are in context — use them directly. Do NOT extract, summarize, or rephrase. Map each slice and its tasks to the corresponding `## N.` group. Nothing left unmapped.

**Step 2 — Density check:** Tasks must be at least as dense as Phase 1.

Write from the mapping. Do NOT discard any Phase 1 slice or task.

**CRITICAL — Missing or underrepresented information becomes missing or wrong work in implementation.** Every Phase 1 slice must appear with full weight intact.

Following the template from Phase 0, write:

```markdown
## 1. <First Vertical Slice — testable outcome implied by name>

- [ ] 1.1 <Task — references file paths inline where useful>
- [ ] 1.2 <Task — verifiable, small enough for one work session>
- [ ] 1.3 <Task>

## 2. <Second Vertical Slice>

- [ ] 2.1 <Task>
- [ ] 2.2 <Task>
```

Every task must satisfy the Core Principles above. Additionally:

* Preserve template format — no custom sections (Files, Dependencies, Notes, Acceptance)
* No code snippets, shell commands, commit messages, or expected output in task descriptions

Write to `outputPath` from Phase 0.

### 2.2 Artifact Compliance Review (MANDATORY — single-shot)

Dispatch subagent of type `general-purpose` (use your subagent/task tool) with reviewer prompt below. Subagent loads tasks.md plus inputs into its own context, returns structured findings list, exits.
> type-general-purpose dispatch: Claude Code `Agent(general-purpose)` · Devin/Windsurf `run_subagent(subagent_general)` · OpenCode `@general` · Codex `spawn_agent` (`multi_agent=true`) · Antigravity `invoke_subagent(self)` · Pi `subagent` · unlisted → self-assess; no dispatch tool → execute inline as self-check.

**Discipline:**

* Single-shot — dispatch once, get findings, fix inline in root, surface for user review
* NEVER re-dispatch after fixing
* NEVER skip subagent because "tasks look fine to me"
* NEVER reload tasks.md / proposal / specs / design into root for review — defeats the purpose

#### Reviewer Subagent Prompt

```
You are a tasks.md reviewer for an OpenSpec change. Verify the tasks file is
complete, well-sliced, and ready for user review.

Inputs:
- Tasks file: <TASKS_PATH>
- Proposal file: <PROPOSAL_PATH>
- Specs file(s): <SPECS_PATHS>
- Design file: <DESIGN_PATH>
- Template: <TEMPLATE_CONTENT_FROM_PHASE_0>

Read all inputs before reviewing. Check each category:

| Category | What to look for |
|---|---|
| Slice discipline | Every `## N.` group is a vertical slice whose heading implies a one-sentence testable outcome. Not a layer (DB/API/UI), phase (setup/polish), or category (tests/docs). Slices are dependency-ordered |
| Task discipline | Every task operates at the spec requirement level, not the scenario level — it names a capability and summarizes its behavioral facets. A task that restates a single Gherkin scenario is too granular; the spec already holds that detail. The design holds HOW; tasks state WHAT |
| Coverage | Every spec requirement and design decision maps to at least one task; no speculative tasks for unrequested features |
| Naming & format | Same component named consistently across tasks. Artifact structure matches template — no custom sections, no extra nesting. No code, shell commands, commit messages, TBD/TODO placeholders |

Calibration: only flag issues that would mislead implementation, lose
coverage, or break the slicing discipline. Minor wording polish is NOT an issue.

Return format:

Status: Approved | Issues Found

Issues (if any):
- [Category]: [specific finding] — [why it matters]

Recommendations (advisory, do not block):
- [optional improvement suggestions]
```

After receiving reviewer's response:

* Status Approved → mark Phase 2 complete, surface tasks.md for user review
* Status Issues Found → fix each Issue inline in root context, surface for user review (NEVER re-dispatch)

Mark Phase 2 complete. Update task status.

---

## Anti-Patterns

NEVER group tasks by layer (DB / API / UI separately) — vertical slice only. NEVER create groups for setup, build, polish, refactor, or other phase categories. NEVER create groups for "tests", "docs", or "config" — those activities live inside their respective vertical slices.

NEVER write code, exact shell commands, expected output, or commit messages in task descriptions. NEVER add custom sections (Files, Dependencies, Notes, Acceptance) — preserve template format. NEVER include TDD ordering inside a task group — that's the implementor's job. If implementation discipline appears, redirect to the implementor phase.

---

## Success Criteria

**Succeeds:** every group satisfies the Slice principle, every task satisfies the Task + WHAT-not-HOW + Behavior-only principles, spec and design coverage complete, template format preserved, reviewer dispatched once with findings applied.

**Fails:** any principle violated, reviewer skipped or re-dispatched, coverage gaps, source code read in root context.