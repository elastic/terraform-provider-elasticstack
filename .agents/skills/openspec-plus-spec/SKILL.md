---
name: openspec-plus-spec
description: "MANDATORY skill that activates whenever the OpenSpec specification phase begins. Triggers: /opsx-new or /opsx-continue runs; openspec-new-change, openspec-continue-change, or openspec-explore is active; `openspec instructions spec` or `openspec instructions specs` is invoked; or the user wants to create, update, review, refine, or discuss an OpenSpec specification."
metadata:
  version: 1.6.0
  priority: high
  disable-user-invocation: true
---

# OpenSpec Plus Spec

## Mission

Strengthen specification quality. Requirements complete, testable, unambiguous, expressed as Gherkin scenarios for behavioral correctness, aligned with approved proposal and any existing design. Improves requirement quality before design (spec-first) or aligns requirements with existing design boundaries (design-first).

---

> **This skill is RIGID. NEVER write any spec file before completing Phase 1, presenting analysis to user, passing pre-write alignment gate, and (after writing) running self-review subagent. Running analysis internally and writing immediately is a skill violation — even if you checked every gate.**

**Red flags — STOP, you are about to violate this skill:**

- "I already checked all the gates"
- "The requirements look complete from the proposal"
- "The user wants this quickly, I'll just write it"
- "I surfaced these gaps in my reasoning, that counts"
- "I'll list all my ambiguities in a table and ask the user to fill them in"
- "I'll ask all questions at once so the user can answer in one message" (unless `settings.questionMode: batch` — then it's the documented procedure, not a shortcut)
- "These ambiguities are related so I'll group them into one question" (outside configured `batch` mode)
- "Free-form scenarios are clearer than Gherkin here"
- "I should read the code to understand the context"
- "The code I need to read is small, I'll just open it directly"
- "I'll skim the spec myself, the reviewer subagent isn't needed"
- "The fixes from the reviewer were minor, I should re-dispatch to confirm"

---

## Inputs Available

OpenSpec allows either order after proposal:

* `proposal → spec → design` (no design yet)
* `proposal → design → spec` (design exists)

When this skill runs:

* ALWAYS read approved proposal — authoritative source of intent
* If design exists, ALWAYS read it — constrains acceptable spec language and bounded behavior
* Spec MUST align with proposal AND any existing design

---

## Workflow

Four phases. NEVER skip or merge.

```text
Phase 0: Schema & Template Resolution
Phase 1: Interactive Analysis
Phase 2: Drift Check
Phase 3: Write & Self-Review
```

---

## Workflow Visibility (MANDATORY)

Display workflow phases via task tool at start; update as each phase completes.

---

## Core Principles

- **Describe What, Not How** — spec defines behavior and outcomes, not implementation; prefer "what must happen" over "how it will be built".
- **Every Requirement Must Be Verifiable** — requirements MUST be testable; avoid vague or subjective language.
- **Surface Missing Requirements** — look for gaps in user behavior, system behavior, permissions, error handling, operational behavior, edge cases.
- **YAGNI — No Speculative Requirements** — every requirement traces to proposal goal or present stakeholder need.
- **Respect Project Standards** — read project instruction files (`AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, equivalents) at Phase 1 start; spec MUST respect project conventions; conflicts → surface to user, never silently override.

## Spec Phase Reads Code Only When Strictly Required — And Only Via Subagent

Spec describes WHAT, not HOW. Default: no code reading — proposal and any existing design provide behavioral context.

When code reading IS required (requirement explicitly references existing behavior to ground), MUST dispatch `explore` subagent. NEVER read code in root context. Prompt MUST scope to observable behavior only, never implementation.

Rigid rule. Root-context code reading during spec phase is a skill violation, regardless of file size. Cannot articulate why code reading required? Don't read code.

When dispatching explore subagent, prompt: "Describe the observable behavior of <X> grounded in <FILES_OR_PATHS>. Return only externally observable: inputs, outputs, state changes, side effects. No implementation, data structures, algorithms, or call paths." Use response only for grounding terminology and current-behavior references.

---

## Phase 0: Schema & Template Resolution

Run FIRST, before any analysis:

1. Run:
   ```bash
   openspec instructions specs --change <name> --json
   ```
   (Use `spec` instead of `specs` if the schema names the artifact `spec` — check `openspec status --change <name>` for the artifact id.) Extract: `template` (structural authority; sections you MUST fill), `instruction` (per-section guidance; what content each section needs), `rules` (project constraints to honor). Parse template sections (H2/H3/H4 headers + HTML comments) — these are your **information requirements**; Phase 1 analysis MUST collect enough substance to fill every section. If the template has sections the 7-step analysis doesn't naturally cover, add targeted questions for them.
2. Read `openspec/.plus/config.yaml` (missing/unreadable/unrecognized → defaults): `settings.questionMode` (`sequential` default; `batch` groups steps 3-7's questions into fewer rounds — see Phase 1 below).

---

## Phase 1: Interactive Analysis (MANDATORY)

**Pre-existing answers:** If recent conversation already answered any analysis points (requirements, ambiguities, edge cases, stakeholder concerns, scenarios) — via prior exploration, a detailed initial request, or any other source — incorporate those into your analysis and SKIP the corresponding question. Ask ONE question at a time only for genuinely unresolved gaps (`batch` mode: collect steps 3-7's questions into as few rounds as possible instead — same real answers, no assumptions). NEVER re-ask questions already answered.

Run every step. Where ambiguities or gaps require user decision, use **question tool** — one question at a time, with options. NEVER write any file during this phase.

Always include your recommended answer with rationale on every question — never a bare question without a recommendation. In `sequential` mode, if the user's answer introduces a new ambiguity or dependent decision, follow that branch before advancing to the next step or gap (`batch`: defer it to the next batch round instead). A gap or ambiguity is only resolved when no sub-decision within it remains open. If a fact can be determined from existing artifacts, project files, or the environment, look it up — do not ask the user for discoverable information.

## 1. Project Context

Before analysis:

1. **Project-level instruction files first** — `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, equivalents at project root, `.claude/`, `.opencode/`, `docs/`. These capture standards: terminology, testing (test types, frameworks, naming), domain language. Spec MUST respect these.
2. Read approved proposal — always present, authoritative for goals, scope, non-goals
3. If design exists, read it — constrains acceptable spec language
4. Read existing spec files for terminology and structural convention

NEVER read source code unless requirement explicitly references current behavior to ground. When code reading IS required, dispatch via `explore` subagent (see Core Principles). Default: no code reading.

## 2. Requirement Extraction

Extract all requirements from proposal (+ design if present). Normalize into clear statements. **Output:** list in plain language.

## 3. Completeness Analysis

Look for gaps across user/system/admin/failure/integration flows. **Output:** gaps found, or "None found" explicitly. Gaps needing user decision → **question tool**, ONE at a time (`batch`: joins the combined round with steps 4-7).

## 4. Ambiguity Analysis

Identify terms with multiple interpretations. For each, use **question tool**: plain language framing, 3-4 concrete options, one "(Recommended)", ONE question at a time. NEVER assume. In `sequential` mode never batch; in `batch` mode present these together with steps 3, 5, 6, 7's questions in as few rounds as the tool allows.

## 5. Edge Case Analysis

Review: invalid input, missing data, permissions, concurrency, retries, partial failures. **Output:** uncovered edges, or "None found." Decisions → **question tool**, ONE at a time (`batch`: joins the combined round).

## 6. Stakeholder Review

Review from: end users, admins, operators, developers, integrators. **Output:** missing or conflicting requirements from any perspective.

## 7. Scenarios In Gherkin

Every functional or behavioral requirement MUST have at least one scenario in Gherkin syntax with uppercase keywords:

```
GIVEN <initial state or precondition>
WHEN <event or action taken>
THEN <observable outcome>
AND <continues prior keyword's category — use only when natural>
BUT <exclusion or counter-outcome — use only when natural>
```

Rules:

* Cover positive, negative, edge case scenarios for each functional/behavioral requirement
* "AND" and "BUT" not mandatory — use only when they are applicable and make scenario clearer
* Multiple scenarios per requirement encouraged when distinct branches exist (happy path, error, edge), never required
* Non-functional requirements (response time, capacity) expressed as plain measurable criteria, not Gherkin

Constructing scenarios surfaces ambiguity — a requirement that can't be expressed as Given/When/Then is not testable.

For each requirement lacking a scenario, with ambiguous scenario, or contradictory scenarios, use **question tool**. ONE question at a time (`batch`: joins the combined round).

---

## Proposal Or Design Conflict

Discover during analysis that proposal — or existing design, if present — is incomplete, contradictory, or inconsistent with what user is asking for:

STOP. Surface the issue. NEVER paper over with spec choices.

Resolutions:

* revise spec within current proposal/design
* revise proposal or design
* split work into separate change

Don't continue spec work until conflict resolved.

---

## Phase 1 Complete — Summarize Decisions

Once all answered, summarize resolved decisions:

* List each ambiguity/gap and user's chosen answer
* Confirm no open questions remain

In `sequential` mode, explicitly ask the user to confirm shared understanding before proceeding to Phase 2 — do NOT advance on silence or implied agreement. In `batch` mode, skip this extra question — the summary was already built from the user's combined replies; proceed to Phase 2.

**Template coverage check:** Verify every template section (from Phase 0) has collected substance to fill it. If any section lacks substance, ask targeted questions until covered (`batch`: combine into one round).

**Rules compliance check:** Review `rules` from Phase 0. If any rule constrains what can be specified (e.g., "all requirements must have Gherkin scenarios", "edge cases must be explicitly documented"), verify the requirements honor those constraints. If a rule is violated, surface the conflict to the user before proceeding (`batch`: fold multiple conflicts into one round).

Mark Phase 1 complete. Update task status. Proceed to Phase 2.

**NEVER write any file until Phase 2 drift check passes.**

---

## Phase 2: Drift Check

Re-read approved proposal (and any existing design). If analysis results drift from proposal scope, goals, or non-goals — STOP. Surface the drift. Resolutions: revise spec, revise proposal/design, or split into separate change. NEVER silently reconcile.

Mark Phase 2 complete. Update task status.

---

## Phase 3: Write & Self-Review

## 3.1 Write the Spec File

Only begin writing after:

1. Phase 1 analysis presented and user has confirmed or resolved all questions
2. Phase 2 drift check passes

Use the `template` and `outputPath` from Phase 0. Use `template` structure EXACTLY. Never improvise sections, restructure, or invent format conventions. Apply `instruction` and `rules` as constraints — do NOT copy them into the file.

**Structure is content — preserve the form.** Gherkin scenarios, requirement lists, and any structured output from Phase 1 must carry through to the spec unchanged. Reducing structured content to prose loses meaning regardless of word count.

**Before writing — 2 mandatory steps:**

**Step 1 — Map Phase 1 to template:** Phase 1 outputs are in context — use them directly. Do NOT extract, summarize, or rephrase. For each template section (from Phase 0), map the full Phase 1 content — requirements, Gherkin scenarios, and any structured output — unchanged. Nothing left unmapped.

**Step 2 — Density check:** Spec must be at least as dense as Phase 1. **Structural fidelity:** every Gherkin scenario from Phase 1 must appear verbatim — prose replacement of any scenario fails this check.

Write from the mapping. Do NOT discard any Phase 1 content.

**CRITICAL — Missing or underrepresented information propagates as blind spots into design, tasks, and implementation.** Every Phase 1 output must appear with full weight and specificity intact.

Write to `outputPath` from Phase 0.

## 3.2 Scenario Format (mandatory)

Every scenario in written spec MUST use Gherkin syntax:

* Uppercase keywords: GIVEN, WHEN, THEN, AND, BUT
* One step per line
* AND continues prior keyword's category, used only when natural
* BUT expresses exclusion or counter-outcome, used only when natural
* Steps describe observable state and behavior, never implementation

Example:

```
GIVEN the user is authenticated
AND the user has admin role
WHEN the user requests the dashboard
THEN the response contains user metrics
AND the response status is 200
BUT no audit logs are exposed
```

## 3.3 Session Context Fidelity Check (MANDATORY)

Inline self-check before dispatching the compliance reviewer.

Re-read the written spec.md in full. Then scan the entire session — Phase 1 analysis, user inputs, resolved ambiguities, gap decisions, scenario discussions, and any pre-phase conversations. For each requirement, constraint, clarification, edge case decision, or behavioral detail surfaced **for the agreed spec**: verify it appears in the written spec.md with its original specificity intact. Discarded alternatives from Phase 1 Q&A are intentionally absent — do NOT flag their omission.

**Pass:** all agreed context captured → proceed to 3.4.  
**Fail:** list each missing item with its session source → fix spec.md inline → re-verify all fixes before proceeding to 3.4.

Do NOT skip. Do NOT proceed to 3.4 with any agreed requirement or decision missing.

## 3.4 Artifact Compliance Review (mandatory, single-shot)

Dispatch subagent of type `general-purpose` (use your subagent/task tool) with reviewer prompt below. Subagent loads spec, proposal, design (if present) into its own context, returns structured findings list, exits.
> type-general-purpose dispatch: Claude Code `Agent(general-purpose)` · Devin/Windsurf `run_subagent(subagent_general)` · OpenCode `@general` · Codex `spawn_agent` (`multi_agent=true`) · Antigravity `invoke_subagent(self)` · Pi `subagent` · unlisted → self-assess; no dispatch tool → execute inline as self-check.

**Discipline:**

* Single-shot only — dispatch once, get findings, fix inline in root, surface for user review
* NEVER re-dispatch reviewer after fixing
* NEVER skip subagent because "spec looks fine to me"
* NEVER reload spec/proposal/design into root for review — defeats the purpose

### Reviewer Subagent Prompt

```
You are a spec document reviewer for an OpenSpec change. Verify the spec is
complete, consistent, and ready for user review.

Inputs:
- Spec file: <SPEC_PATH>
- Proposal file: <PROPOSAL_PATH>
- Design file (path if exists, else "NONE"): <DESIGN_PATH_OR_NONE>
- Template: <TEMPLATE_CONTENT_FROM_PHASE_0>

Read all inputs before reviewing. Check each category:

| Category | What to look for |
|---|---|
| Placeholders | TBD, TODO, "[fill in]", incomplete sections |
| Internal consistency | Two requirements that contradict each other |
| Scenario format | Every scenario uses uppercase GIVEN/WHEN/THEN; AND/BUT only when natural |
| Coverage | Functional/behavioral requirements have positive, negative, and edge scenarios where applicable |
| Implementation leaks | Requirements or scenarios describing HOW instead of WHAT |
| Terminology consistency | Same concept named consistently throughout |
| Proposal alignment | Every requirement traces to a proposal goal; no scope expansion; no relaxed non-goals; no contradictions |
| Design alignment (only if design path provided) | No requirement contradicts a design decision; no design detail leaks into the spec |
| Template compliance | Artifact sections match the provided template — no improvised, missing, or reordered sections |

Calibration: only flag issues that would mislead the user during review or
that would cause downstream phases to build the wrong thing. Minor wording
improvements and stylistic preferences are NOT issues.

Scenario keyword format (GIVEN/WHEN/THEN/AND/BUT) is defined by the skill,
not the template. If the template uses a simpler format (e.g., WHEN/THEN
only), the skill's format takes precedence — do NOT flag this as a template
compliance issue or scenario format issue.

Return format:

Status: Approved | Issues Found

Issues (if any):
- [Category]: [specific finding] — [why it matters]

Recommendations (advisory, do not block):
- [optional improvement suggestions]
```

After receiving reviewer's response:

* Status Approved → mark Phase 3 complete, surface spec for user review
* Status Issues Found → fix each Issue inline in root context, surface for user review (NEVER re-dispatch)

Mark Phase 3 complete. Update task status.

---

## Anti-Patterns

NEVER define architecture, services, components, schemas, APIs. NEVER choose technologies. NEVER create implementation plans or generate tasks. NEVER write free-form prose scenarios in place of Gherkin for functional/behavioral requirements. NEVER describe HOW in a scenario. Implementation details appear — redirect to design phase.

---

## Success Criteria

**Succeeds:** requirements clearer/more complete, ambiguities resolved WITH user, Gherkin scenarios for all functional requirements (positive/negative/edge), scenarios implementation-independent, aligned with proposal/design, code reading via explore subagent only, reviewer dispatched once with findings applied.

**Fails:** spec written before Phase 1 confirmed, drift check skipped, reviewer skipped/re-dispatched, scenarios in prose or describing implementation, code read in root, design/architecture/technology/implementation work appears.
