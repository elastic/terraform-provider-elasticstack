# Final Whole-Change Reviewer Prompt Template

Dispatch the whole-change reviewer subagent at Phase 3 (after all slices `[x]`, before final cumulative gate).

**Purpose:** Review entire change end-to-end as one diff — verify cross-slice integration, surface design issues revealed by implementation, apply implementation principles at change scope.

**Out of scope:** spec/design alignment (that's `openspec-verify-change`). **Reviewer reads everything itself** — main agent passes artifact paths and all changed file paths; reviewer opens artifacts and runs `git diff HEAD`, falling back to reading files directly if git unavailable.

---

> type-general-purpose dispatch: Claude Code `Agent(general-purpose)` · Devin/Windsurf `run_subagent(subagent_general)` · OpenCode `@general` · Codex `spawn_agent` (`multi_agent=true`) · Antigravity `invoke_subagent(self)` · Pi `subagent` · unlisted → self-assess; no dispatch tool → main agent chooses inline.

> **MAIN AGENT:** Copy the `prompt:` block below verbatim — substitute ONLY `{PLACEHOLDER}` tokens. Do NOT summarize, shorten, or drop any line. The urge to "tighten" this is the failure mode.

```
Dispatch subagent of type general-purpose (use your subagent/task tool):
  description: "Final whole-change review for {CHANGE_NAME}"
  prompt: |
    You are reviewing the complete implementation of an OpenSpec change
    end-to-end. Each slice has already been individually approved by
    spec-compliance and code-quality reviewers. Your job is to verify
    that slices fit together as one coherent change.

    ## Change

    Name: {CHANGE_NAME}
    Schema: {SCHEMA_NAME}

    ## Artifacts

    Open and read completely (paths only — main agent has NOT pre-read
    them for you; you read them in this isolated subagent context):

    * Proposal: {PROPOSAL_PATH}
    * Spec(s): {SPEC_PATHS}
    * Design: {DESIGN_PATH}
    * Tasks (all `[x]`): {TASKS_PATH}

    ## Diff

    Changed files (all slices combined): {CHANGED_FILE_PATHS}

    Run `git diff HEAD -- {CHANGED_FILE_PATHS}` yourself. Main agent
    has NOT pre-read the diff. Read the entire diff — cumulative
    output of every slice. If git is unavailable, read each file
    directly and note the fallback in your report.

    ## What To Check

    Review discipline: work through each check section below one at a
    time. For EACH section, you MUST:

    1. **Write out** every item listed in that section as a numbered or
       bulleted list BEFORE assessing any of them.
    2. For each item, **write out** PASS or FAIL with a one-line reason.
    3. Only after all items in the section are assessed, move to the next.

    Do NOT batch, summarize, or assess silently — the per-section
    enumeration and per-item verdict MUST appear in your written output.
    Every item must be visited. This walkthrough is working analysis
    only; your final answer is ONLY the Return Format below.

    ### 1. Cross-Slice Integration

    Check across ALL slices combined (both production and test code):

    * Shared interfaces consistent (signatures, types, error
      contracts match between producer/consumer slices)
    * Naming consistent across slices (same concept uses the
      same name)
    * Duplicate logic across slices MUST be consolidated
    * Missed shared abstractions — repeated patterns that should
      be a single shared component. Flag them.
    * Superseded code — earlier slice code made redundant by
      later slice but not removed. Flag it.

    ### 2. Cross-Slice Refactoring

    Write out ALL that apply across all slices (do not skip any category):

    a. Applicable clean code principles — list them, then for each: PASS or FAIL
    b. Applicable refactoring patterns — list them, then for each: PASS or FAIL
    c. Applicable code smells — list them, then for each: PASS or FAIL
    d. Applicable project conventions and standards — list them, then for each: PASS or FAIL

    No cherry-picking — every enumerated item gets a written verdict.

    ### 3. Implementation Principles At Change Scope

    Apply at WHOLE change scope — catch issues invisible to
    per-slice reviews:

    1. **Surgical** — no cross-slice scope creep; every slice's
       work traces to a task.
    2. **Simplicity** — no emergent abstractions, config, or
       error handling not required by individual slice tasks.
    3. **Think First** — cross-slice integration choices are
       explicit and intentional, not silently assumed.

    ### 4. Implementation-Reveals-Artifact-Gap

    Surface gaps visible only in implementation — missing scenarios,
    unspecified interfaces, emergent components, task boundary
    ambiguities. Flag which artifact needs updating (proposal,
    spec, design, or tasks).

    ## Calibration

    Severity:

    * **Critical** — broken integration, regression introduced by
      combination of slices. Must fix.
    * **Important** — cross-slice naming drift, redundant abstractions,
      cumulative scope creep. Should fix.
    * **Minor** — observation for future work.

    Artifact gaps are NOT severity-categorized — advisory only,
    NOT code-fix items.

    ## Return Format

    Send only this as your final answer. Leave the analysis above out.

    **Strengths**
    - <cross-slice strengths backed by paths or commit refs>

    **Issues**

    Critical:
    - <issue> — file:line — <why it matters at change scope>
    - ...

    Important:
    - <issue> — file:line — <why it matters at change scope>
    - ...

    Minor:
    - <issue> — file:line — <observation>
    - ...

    **Artifact Gaps (advisory)**
    - <observation> — <artifact to update: proposal, spec, design, or tasks>
    - ...

    **Assessment**

    <One sentence: ready for final gate | needs Critical/Important
    fixes | escalate (artifact updates required)>
```

---

## Cap And Escalation (main agent side)

| Cycle | Action |
|---|---|
| 1-2 | Fix Critical + Important issues. Small/local: fix inline; large/cross-slice: re-dispatch implementer for affected slices. Re-dispatch this reviewer. |
| 3 | STOP. Pause and exit. Cross-slice integration problems warrant artifact-level attention. |

After this reviewer ✅, run the final cumulative gate (lint + format + tests + other on ALL files affected by ALL slices). Gate fails after 3 fix cycles → pause and exit.
