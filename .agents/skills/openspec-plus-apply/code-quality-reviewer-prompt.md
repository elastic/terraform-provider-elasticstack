# Code-Quality Reviewer Prompt Template

Dispatch a code-quality reviewer subagent for an OpenSpec change slice.

**Purpose:** Verify implementation is well-built — clean, surgical, simple, tested, maintainable.

**Order:** Dispatch ONLY after spec-compliance reviewer returns ✅. Never run code-quality before spec-compliance is clean.

**Reviewer reads the diff itself.** Main agent passes changed file paths; reviewer runs `git diff HEAD` in its isolated context, falling back to reading files directly if git unavailable.

---

> type-general-purpose dispatch: Claude Code `Agent(general-purpose)` · Devin/Windsurf `run_subagent(subagent_general)` · OpenCode `@general` · Codex `spawn_agent` (`multi_agent=true`) · Antigravity `invoke_subagent(self)` · Pi `subagent` · unlisted → self-assess; no dispatch tool → main agent chooses inline.

> **MAIN AGENT:** Copy the `prompt:` block below verbatim — substitute ONLY `{PLACEHOLDER}` tokens. Do NOT summarize, shorten, or drop any line. The urge to "tighten" this is the failure mode.

```
Dispatch subagent of type general-purpose (use your subagent/task tool):
  description: "Review code quality for slice {SLICE_NUMBER}"
  prompt: |
    You are reviewing code quality of slice {SLICE_NUMBER}
    ("{SLICE_NAME}") of an OpenSpec change. Spec compliance has already
    been confirmed; do not re-review correctness against the spec.

    ## Slice Tasks

    {TASKS_TEXT}

    ## What Implementer Claims They Built

    {IMPLEMENTER_REPORT}

    ## Artifact Files (paths only — read them yourself)

    * Design: {DESIGN_PATH}

    * **Design** — read fully, end to end; its structure is dynamic
      (template-driven). Every section that touches this slice's
      architecture, file structure, component boundaries, naming,
      or patterns MUST be used for verification — do NOT skip or
      deprioritize any detail.

    ## Spec Requirements & Gherkin Scenarios (This Slice Only)

    {SPECS_TEXT}

    Use these as the scope boundary for surgical/simplicity checks —
    nothing beyond what requirements and scenarios asked for.

    ## Diff

    Changed files: {CHANGED_FILE_PATHS}

    Run `git diff HEAD -- {CHANGED_FILE_PATHS}` yourself. Every
    claim MUST reference actual code. If git is unavailable, read
    each file directly and note the fallback in your report.

    ## Project Standards Documents (Read Independently)

    The project's source of truth for code style, file organization, naming, testing conventions, lint/format/test commands, and framework idioms.

    {STANDARDS_DOC_PATHS}

    Read each one, including referenced docs. MUST enumerate all
    conventions relevant to this slice's scope, MUST verify all of
    them with no selective skipping, and in your report MUST list
    the conventions verified and cite source docs for violations.

    ## What To Check

    Review discipline: work through each concern section below one at a
    time. For EACH section, you MUST:

    1. **Write out** every item listed in that section as a numbered or
       bulleted list BEFORE assessing any of them.
    2. For each item, **write out** whether it is relevant to this slice,
       then **write out** PASS or FAIL with a one-line reason.
    3. Only after all items in the section are assessed, move to the next.

    Do NOT batch, summarize, or assess silently — the per-section
    enumeration and per-item verdict MUST appear in your written output.
    Every item must be visited. Apply to both production and test code.
    This walkthrough is working analysis only; your final answer is
    ONLY the Return Format below.

    1. **Standard concerns:**

       * **File responsibility** — each file one clear responsibility,
         well-defined interface
       * **Decomposition** — each unit understandable and testable
         independently
       * **File growth** — did this slice create new files already large,
         or significantly grow existing files? Don't flag pre-existing
         sizes — focus on what THIS slice contributed
       * **Naming** — match what things do, not how; consistent with
         project's naming conventions
       * **Error handling** — matches codebase's existing patterns
       * **Test quality** — tests exercise real behavior not mock
         behavior; comprehensive for slice's scenarios; same comment
         convention as production (no Given/When/Then or Arrange/Act/
         Assert narration, no comment restating the test name)

    2. **Cross-task refactoring concerns:**

       Write out ALL that apply to this slice (do not skip any category):

       a. Applicable clean code principles — list them, then for each: PASS or FAIL
       b. Applicable refactoring patterns — list them, then for each: PASS or FAIL
       c. Applicable code smells — list them, then for each: PASS or FAIL
       d. Applicable project conventions and standards — list them, then for each: PASS or FAIL

       No cherry-picking — every enumerated item gets a review and written verdict.

    3. **Implementation principles concerns:**

       * **Surgical Changes** — every changed line traces to a slice task? Adjacent code refactored without being asked?
       * **Simplicity First** — abstractions for single-use code? Configurability slice didn't request? Would a senior engineer call this overcomplicated?
       * **No Speculative Features** — anything beyond what slice tasks and Gherkin scenarios required?

    4. **Code-style concerns:**

       * **Comments** — comments on non-complex logic are noise; flag them. If a comment exists, ask: could a better name/structure remove it? If yes, flag as Important (refactor-first rule).
       * **Testability** — hard to test = hard to use, usually a sign of coupling.
       * **Readability** — does each function fit one mental load?
       * **Maintainability** — would a new contributor understand the slice's code shape in a few minutes?
       * **No commented-out code, no TODO, no FIXME, no "explained later" markers.**

    5. **Project standards compliance (MANDATORY):**

       Read every rule in the project standards docs. Verify the implementation honors each applicable rule — naming, structure, framework idioms, testing conventions, type annotations, file placement. Cherry-picking is itself a violation. Cite source doc for each violation (e.g., "AGENTS.md §Naming: `usrCtrl` should be `userController`").

       Severity: **Important** for most violations; **Critical** only when the rule breaks the build; **Minor** for "should"-level stylistic rules.

    ## Calibration

    Severity:

    * **Critical** — bug, broken behavior, security issue, regression,
      rule violation that breaks the project.
    * **Important** — design concern, scope violation (surgical/
      simplicity), unclear naming, missing test coverage for a real
      edge case, project-rule violation (naming/structure/framework
      idioms/testing conventions).
    * **Minor** — style preference, suggestion, observation,
      "should"-level rule violation. Note but don't block.

    Don't invent issues. Slice clean → say so. Cite the source doc
    for every project-rule violation.

    ## Return Format

    Send only this as your final answer. Leave the analysis above out.

    **Strengths**
    - <brief, specific strengths backed by file:line>

    **Issues**

    Critical:
    - <issue> — file:line — <why it matters>
    - ...

    Important:
    - <issue> — file:line — <why it matters>
    - ...

    Minor:
    - <issue> — file:line — <observation>
    - ...

    **Assessment**

    <One sentence: ready to proceed | needs Critical/Important fixes>
```

---

## Cap And Escalation (main agent side)

| Cycle | Action |
|---|---|
| 1-2 | Implementer fixes Critical + Important, reviewer re-dispatched |
| 3 | STOP. Pause and exit. Architecture is wrong, not implementation. Suggest `plus-design` artifact update. |

NEVER attempt cycle 4. Minor issues are noted but never block progression.
