---
name: openspec-plus-design
description: "MANDATORY skill that activates whenever the OpenSpec design phase begins. Triggers: /opsx-new or /opsx-continue runs; openspec-new-change, openspec-continue-change, or openspec-explore is active; `openspec instructions design` is invoked; or the user wants to create, update, review, refine, or discuss an OpenSpec design document."
metadata:
  version: 1.6.0
  priority: high
  disable-user-invocation: true
---

# OpenSpec Plus Design

## Mission

Strengthen OpenSpec design with brainstorming patterns. Prevent premature architectural convergence — explore alternatives, compare tradeoffs, recommend, get user selection, build incrementally, validate vs proposal (and any existing spec) before writing design.md.

---

> **This skill is RIGID. Do NOT write any design file before all workflow phases complete through Phase 3 section approvals. After writing, reviewer subagent is MANDATORY — skipping review is a skill violation.**

**Red flags — STOP, you are about to violate this skill:**

- "There's only one obvious approach"
- "I'll generate the full design and let the user review at the end"
- "The user is in a hurry, I'll skip selection"
- "I'll group all sections into one approval question"
- "I have approval for the direction, so I can write the whole document now"
- "The proposal (and spec, if it exists) is fresh, I don't need the reviewer subagent"
- "Spec doesn't exist yet, I can sneak some requirements into the design"
- "The design looks fine to me, I'll skip the reviewer subagent"
- "Reviewer flagged minor issues, I'll re-dispatch after fixing to confirm"

---

## Inputs Available

OpenSpec allows either order after proposal:

* `proposal → spec → design` (spec exists)
* `proposal → design → spec` (no spec yet)

When this skill runs:

* ALWAYS read approved proposal — authoritative for intent and capabilities
* If spec exists, ALWAYS read it — design MUST align with it
* If no spec yet, design works from proposal capabilities only — do NOT capture detailed requirements (that's spec phase, runs after design)
* Design MUST align with proposal AND any existing spec

---

## Workflow

Five phases. NEVER skip, merge, or reorder.

```text
Phase 0: Schema & Template Resolution
Phase 1: Explore Design Approaches
Phase 2: Select Design Direction
Phase 3: Build Design Sections
Phase 4: Generate Design
```

---

## Workflow Visibility (MANDATORY)

Display workflow phases via task tool at start; update as each phase completes.

---

## Core Principles

- **Explore Before Converging** — never commit to first approach; always explore alternatives.
- **User Owns The Final Direction** — model recommends, user decides.
- **Build Design Collaboratively** — in `sequential` mode, NEVER generate full design at once; build incrementally, review each section with user (`batch` mode: generate all, one consolidated review — see Phase 3).
- **Design Before Documentation** — validate design (including alignment with proposal and any existing spec) before writing design.md.
- **YAGNI — Complexity Must Be Justified** — every abstraction, layer, indirection MUST serve a real present need.
- **Improve, Don't Refactor** — targeted improvements to existing code only when they serve the change; never propose unrelated refactoring.
- **Respect Project Standards** — read project instruction files (`AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, equivalents) at Phase 1 start; design MUST respect project conventions; conflicts → surface to user, never silently override.

## Use Stable Libraries Before Custom

Survey ecosystem before proposing custom code. For each significant component (parsing, validation, HTTP, retries, scheduling, queuing, auth, crypto, dates, money, observability, etc.):

1. Identify 2-3 candidate libraries
2. Evaluate by: maintenance, stability (semver, breaking changes), popularity, license, security (CVEs, audits), compatibility (runtime, version, existing deps)
3. Library fits → propose with rationale; don't write custom for the same purpose
4. Choosing custom → justify why no library fits (domain-specific, license, dep tree, no maintained option, perf constraints unmet)

Respect any approved-library list from project instruction files. Don't propose libraries the project rejected.

**Human-In-The-Loop Approval:** For ANY new library introduction or replacement, surface to user via question tool: 2-3 candidates evaluated, recommended choice + rationale, accepted tradeoffs, alternatives rejected + why. Do NOT commit a library in the design until user approves. Libraries already in use don't need re-approval.

- **Stay At Architecture Level** — design captures HOW (architecture, patterns, integration); NEVER capture WHAT (detailed requirements, scenarios, acceptance criteria) — that's spec phase.

---

## Phase 0: Schema & Template Resolution

Run FIRST, before any exploration:

1. Run:
   ```bash
   openspec instructions design --change <name> --json
   ```
   Extract: `template` (structural authority; sections you MUST fill), `instruction` (per-section guidance; what content each section needs), `rules` (project constraints to honor). Parse template sections (H2/H3 headers + HTML comments) — these are your **information requirements**; Phase 3 design concerns MUST produce enough substance to fill every section. If the template has sections the hardcoded 5+1 concerns don't cover, add dynamic concerns during Phase 3.
2. Read `openspec/.plus/config.yaml` (missing/unreadable/unrecognized → defaults): `settings.questionMode` (`sequential` default; `batch` groups Phase 1/Library Resolution questions and collapses Phase 3 into one generate-all-then-review round — see Phase 1 and Phase 3 below).

---

## Phase 1: Explore Design Approaches

**Pre-existing answers:** If recent conversation already covers architectural decisions, approaches discussed, constraints noted, or integration seams — via prior exploration, a detailed initial request, or any other source — incorporate those into your analysis and SKIP redundant clarification. Ask ONE question at a time only for genuinely unresolved areas (`batch` mode: present all unresolved areas together instead — same rule as proposal's Phase 1). NEVER re-ask questions already answered in recent conversation.

## Project Context

Before generating approaches:

1. **Project-level instruction files first** — `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, equivalents at project root, `.claude/`, `.opencode/`, `docs/`. These capture standards (tech stack, patterns, testing, architecture, naming). Design MUST respect them unless user approves deviation.
2. Re-read approved proposal — always present (intent, capabilities, scope, non-goals).
3. If spec exists, re-read it — design must align. If no spec yet, skip; spec phase runs after design and aligns with what we produce here.
4. Read code referenced by those artifacts — design needs pattern-level knowledge.
5. Note conventions, similar implementations, integration seams.

Proposal answered "why" + capabilities. Spec (if present) answered "what". Project files answer "how this project does things". Design needs "how does the existing code look". Reuse, don't duplicate.

If no spec exists and you find yourself defining detailed requirements, STOP — that's spec-phase work.

Identify: architectural decisions that must be made and areas where multiple solutions exist. If a fact can be determined from code, project files, or existing artifacts, look it up — do not ask the user for discoverable information. For decisions that genuinely require user input, use question tool — ONE question at a time in `sequential` mode (default), always with your recommended answer and rationale; in `batch` mode, present all genuinely-required decisions together, one combined reply. If the user's answer introduces a dependent decision or leaves a branch unresolved, follow it before advancing (`batch`: next batch round). Do not generate approaches until all pre-approach decisions are resolved.

Generate TWO OR THREE viable design approaches (two when only two meaningful directions exist; three when three materially different solutions exist; never invent a weak approach). Each approach MUST satisfy proposal capabilities and spec (if present). For each provide: Approach Name, Core Idea, Advantages, Disadvantages, Complexity Level, Key Assumptions. Approaches must be meaningfully different — no cosmetic variations.

Recommend ONE approach: explain why preferred, what tradeoffs accepted, why alternatives less suitable.

Mark Phase 1 complete. Update task status.

---

## Phase 2: Select Design Direction

Use question tool. Present all approaches. Mark recommended.

Requirements:

* Present all approaches
* Mark recommended option
* Ask exactly one selection question
* Wait for user selection
* Don't continue until choice made

User may select any option. Two or three options.

Mark Phase 2 complete. Update task status.

## Library Resolution (before Phase 3)

With the approach selected, identify every component that may require a new library (parsing, validation, HTTP, auth, crypto, queuing, scheduling, etc.). For each, apply Use Stable Libraries Before Custom: survey 2-3 candidates, evaluate, surface recommended choice + rationale via question tool, wait for approval — one library per question in `sequential` mode, all libraries together in one round in `batch` mode. Do NOT begin Phase 3 until all new library choices are approved. Already-installed libraries do not need re-approval. No new libraries → proceed immediately.

---

## Phase 3: Build Design Sections

After direction selected, build incrementally. In `sequential` mode (default), NEVER generate full design at once — generate ONE section at a time. Scale each section to its complexity — simple concerns need a few sentences, complex ones several paragraphs; length serves clarity, never appearance of rigor.

After each section, ask via question tool: `Does this section look correct? A. Continue (Recommended) | B. Revise Section | C. Revisit Design Direction` — wait for response before proceeding.

In `batch` mode: generate ALL sections up front (same rigor, concerns 1-7 all still fully worked through), then ask ONE consolidated question: `Continue (Recommended) | Revise section(s) — specify which | Revisit Design Direction`. Revision regenerates only the flagged section(s), re-presented for one more consolidated approval.

---

## Design Section Rules

Phase 3 "sections" are DESIGN CONCERNS, all worked through in substance regardless of mode — approval is per-concern in `sequential`, one consolidated approval in `batch` (see Phase 3) — NOT the same as template sections (Phase 0). Phase 4 maps concerns into template sections.

### Concrete Design Concerns (Phase 3 walk-through)

Five + one conditional + one mandatory. Walk through each — approval per concern (`sequential`) or one consolidated approval (`batch`). Skip only if genuinely N/A.

1. **Architecture** — shape, layers, boundaries, external integration points
2. **Component Structure** — units, responsibilities, interfaces
3. **Data Flow** — how data moves, where state lives, mutation ownership
4. **Error Handling / Failure Modes** — what fails, propagation, retry, partial-failure, observability
5. **Testing Approach** — design-level strategy: test pyramid (unit/integration/contract/e2e), environments, hard-to-test areas needing design accommodation, infrastructure, alignment with project standards from `AGENTS.md`/`CLAUDE.md`
6. **Migration / Rollout** (when applicable) — rollout without breaking existing
7. **Additional Concerns** (mandatory) — before Phase 4, run both checks:
   - **Template gaps:** compare Phase 0 template sections against concerns 1-6; each section without substance becomes a concern.
   - **Change-nature gaps:** given this specific change and design, are concerns like security, performance, observability, backwards compatibility, compliance, or resilience relevant but unaddressed by 1-6? Add those that apply.
   Walk all additions through with the user the same way before proceeding to Phase 4.

NEVER skip Error Handling or Testing for non-trivial changes. "Testing Approach" is DESIGN-level (what/where/infrastructure) — NOT test cases (→ spec Gherkin) or TDD ordering (→ implementor).

### Mapping Concerns To Template

Map concerns into template sections at write time (default `spec-driven` schema — adapt if template differs):

| Phase 3 concern                | Lands in template section |
|--------------------------------|---|
| Architecture                   | Decisions (architecture choice + rationale) and any dedicated Architecture section |
| Component Structure            | Decisions (component breakdown + interfaces) |
| Data Flow                      | Decisions (flow + state ownership) |
| Error Handling / Failure Modes | Decisions (failure-handling strategy) and Risks (specific failure modes + mitigations) |
| Testing Approach               | Decisions (test boundary + infrastructure choices) |
| Migration / Rollout            | Migration Plan section |
| Additional Concerns            | Any section NOT covered by the 5+1 concerns above |

CONCERNS shape substance. TEMPLATE shapes structure.

---

## Section Refinement

User requests changes? Revise current section, re-present, request approval again. Don't continue until current section accepted.

---

## Revisit Design Direction

User chose "Revisit Design Direction"? Return to Phase 2. Allow another selection. Restart section generation.

---

## Proposal Or Spec Conflict

Discover during section building that proposal — or existing spec, if present — is incomplete, contradictory, or wrong:

STOP. Surface the issue. NEVER paper over with design choices.

Resolutions:

* If spec exists: revise design within current spec, revise spec, or revise proposal
* If no spec: revise design within current proposal, revise proposal, or split into separate change

Don't continue until conflict resolved.

---

## Section Completion

In `sequential` mode, continue ONE section at a time until all required sections reviewed and accepted (`batch`: all generated up front, one consolidated approval — see Phase 3).

**Rules compliance check:** Review `rules` from Phase 0. If any rule constrains design choices (e.g., "testing approach section is mandatory", "migration plan required for breaking changes"), verify the design sections honor those constraints. If a rule is violated, surface the conflict to the user before proceeding (`batch`: fold multiple conflicts into one round).

Mark Phase 3 complete. Update task status.

---

## Phase 4: Generate Design

## 4.1 Write Design Document

Only after approaches explored, direction selected, and all sections reviewed/accepted.

Final design must: reflect only the selected approach (skip the discarded approaches), incorporate approved refinements, preserve tradeoffs, align with proposal/spec, stay at architecture level. Use `template` and `outputPath` from Phase 0. Use template structure EXACTLY. Apply `instruction` and `rules` as constraints — do NOT copy them into the file.

**Structure is content — preserve the form.** Whatever representation information took during Phase 1-3 — table, diagram, matrix, comparison, enumeration, flow, or any other non-prose structure — that form must carry through to the artifact unchanged. Reducing any structured representation to prose loses the structure and therefore the meaning, regardless of word count.

**Before writing — 3 mandatory steps:**

**Step 1 — Map Phase 3 sections to template:** Your Phase 3 section outputs are already in context — use them directly. Do NOT extract, summarize, or rephrase. For each Phase 3 section, state which template section(s) it maps to and confirm the full content — including all tables, diagrams, and structured elements — will appear in the artifact unchanged. Nothing left unmapped.

**Step 2 — Density check:** Artifact must be at least as dense as the sum of Phase 1-3 sections. **Structural fidelity sub-check:** every structured representation from Phase 1-3 must appear in the artifact in its original form — prose replacement of any structured content fails this check even if word count is similar.

Write from the mapping. Do NOT discard any section content.

**CRITICAL — Missing or underrepresented information propagates as blind spots into tasks and implementation.** Every section must appear with full weight and specificity intact.

Write to `outputPath` from Phase 0.

---

## 4.2 Session Context Fidelity Check (MANDATORY)

Inline self-check before dispatching the compliance reviewer.

Re-read the written design.md in full. Then scan the entire session — Phase 1-3 discussions, user inputs, and any pre-phase conversations. For each design decision, tradeoff, rationale, constraint, or integration detail surfaced **for the selected approach**: verify it appears in the written design.md with its original specificity intact. Discarded approaches from Phase 1-2 are intentionally absent — do NOT flag their omission.

**Pass:** all context captured → proceed to 4.3.  
**Fail:** list each missing item with its session source → fix design.md inline → re-verify all fixes before proceeding to 4.3.

Do NOT skip. Do NOT proceed to 4.3 with any material context missing.

---

## 4.3 Artifact Compliance Review (MANDATORY — single-shot)

Dispatch subagent of type `general-purpose` (use your subagent/task tool) with reviewer prompt below. Subagent loads written design.md, proposal, spec (if present) into its own context, returns structured findings list, exits.
> type-general-purpose dispatch: Claude Code `Agent(general-purpose)` · Devin/Windsurf `run_subagent(subagent_general)` · OpenCode `@general` · Codex `spawn_agent` (`multi_agent=true`) · Antigravity `invoke_subagent(self)` · Pi `subagent` · unlisted → self-assess; no dispatch tool → execute inline as self-check.

**Discipline:**

* Single-shot — dispatch once, get findings, fix inline in root, surface for user review
* NEVER re-dispatch after fixing
* NEVER skip subagent because "design looks fine to me"
* NEVER reload design/proposal/spec into root for review — defeats the purpose

### Reviewer Subagent Prompt

```
You are a design document reviewer for an OpenSpec change. Verify the design is
complete, consistent, and ready for user review.

Inputs:
- Design file: <DESIGN_PATH>
- Proposal file: <PROPOSAL_PATH>
- Spec file (path if exists, else "NONE"): <SPEC_PATH_OR_NONE>
- Template: <TEMPLATE_CONTENT_FROM_PHASE_0>

Read all inputs before reviewing. Check each category:

| Category | What to look for |
|---|---|
| Proposal alignment | Every design decision traces to a proposal capability; no scope expansion; no relaxed non-goals; no contradictions |
| Spec alignment (only if spec exists) | Every spec requirement addressed by design; no design decision contradicts a spec requirement; no design quietly redefines a spec term |
| Scope drift | No design decision adds capabilities outside proposal scope; no design decision captures detailed requirements (those belong to spec phase) |
| Architecture level | Design stays at HOW (architecture, patterns, integration) — never WHAT (detailed requirements, scenarios, acceptance criteria) |
| Template compliance | Artifact sections match the provided template — no improvised, missing, or reordered sections |
| Internal consistency | No two decisions that contradict each other |
| Terminology consistency | Same concept named consistently throughout |
| YAGNI | Every abstraction, layer, or indirection serves a real present need — no "for the future" |
| Implementation leaks | No task lists, code snippets, WBS, or execution plans |
| Testing Strategy | Layers, tooling choices, and boundary definitions are architectural decisions that belong in design and intentionally constrain the tasks phase. Flag only if specific test code, mock implementations, or file structure appear |
| Rejected approaches | Design must contain ONLY the selected approach. If rejected/alternative approaches appear, flag — they belong in the decision process, not the output |
| Placeholders | TBD, TODO, "[fill in]", incomplete sections |

Calibration: only flag issues that would mislead the user during review or
cause downstream phases to build the wrong thing. Minor wording improvements
and stylistic preferences are NOT issues.

Return format:

Status: Approved | Issues Found

Issues (if any):
- [Category]: [specific finding] — [why it matters]

Recommendations (advisory, do not block):
- [optional improvement suggestions]
```

After receiving reviewer's response:

* Status Approved → proceed to 4.4
* Status Issues Found → fix each Issue inline in root context (NEVER re-dispatch)

**Same density rule applies.** If the fixed design is significantly shorter than the sum of Phase 1-3 discussions, re-expand before proceeding.

---

## 4.4 Mark Phase 4 Complete

Surface design for user review. Mark Phase 4 complete. Present final workflow completion.

---

## Anti-Patterns

Implementation plans, WBS, estimates, code, deployment → later phases. Detailed requirements, scenarios, acceptance criteria → spec phase. Never produce either here.

---

## Success Criteria

**Succeeds:** alternatives explored, tradeoffs considered, recommendation provided, user selected, every concern approved (per-section in `sequential`, consolidated in `batch`), reviewer dispatched once with findings applied, design at architecture level.

**Fails:** alternatives/selection/concern-approval skipped, full design generated with no approval gate at all, reviewer skipped/re-dispatched, requirements or implementation planning appears.
