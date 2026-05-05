## Context

The repository maintains three Agentic Workflows that use a "issue slot" pattern to limit GitHub issue creation. Each workflow today carries an identical block of frontmatter YAML:

- A `steps` entry that runs `actions/github-script@v9` with a wrapper script (`compute_issue_slots.inline.js`) that counts open issues by label and computes `cap - open`
- A `jobs.pre-activation` definition that exposes `open_issues`, `issue_slots_available`, and `gate_reason` as outputs
- An `if` gate on the agent job (`needs.pre_activation.outputs.issue_slots_available != '0'`)
- A `## Pre-activation context` body section repeated in each workflow with only the label name and cap value differing

There are three identical copies of `scripts/compute_issue_slots.inline.js` — one per workflow directory — each including the shared `lib/issue-slots.js` via the custom compiler's `//include:` directive. The customization per workflow is purely the `ISSUE_SLOTS_LABEL` and `ISSUE_SLOTS_CAP` env vars passed to `actions/github-script`.

GH AW supports parameterized shared imports (`import-schema` + `uses`/`with`) that merge `steps`, `jobs`, `safe-outputs`, `mcp-servers`, and body content. This is the canonical mechanism for deduplicating workflow components, documented at `https://github.github.com/gh-aw/reference/imports/` and demonstrated by upstream examples such as `github/gh-aw`'s `shared/mcp/serena.md`.

The custom workflow-source compiler (`scripts/compile-workflow-sources/`) expands `x-script-include:` into inline `script: |` blocks when generating `.github/workflows/*.md` files. This still works for shared components because the compiler processes each template independently before `gh aw compile` merges imports.

## Goals / Non-Goals

**Goals:**
- Centralize the entire issue-slot pre-activation mechanism into a single shared workflow component
- Parameterize the label and cap so each consumer passes only those two values
- Keep the shared component's body (`## Pre-activation context`) prepended to every importing workflow's agent prompt
- Retain the existing `lib/issue-slots.js` as the unit-tested helper; no logic changes
- Eliminate the three triplicated `scripts/compute_issue_slots.inline.js` files

**Non-Goals:**
- Changing the `kibana-spec-impact` workflow (it uses a different Go-tool pre-activation pattern)
- Changing any workflow behavior (caps, labels, gating logic, output names) — this is a pure structural refactor
- Replacing the custom compiler with a different build system
- Extracting the agent-prompt body for `kibana-spec-impact` (out of scope)

## Decisions

### Decision: Use `import-schema` with `label` and `cap` parameters

**Rationale:** These are the only two values that vary across the three consumers. Everything else (GitHub script logic, output names, job structure, prompt wording) is identical.

**Alternatives considered:**
- Hardcode cap in the shared import and only parameterize label. Rejected: `schema-coverage-rotation` uses `max: 3` in `safe-outputs` today but parameterizing cap future-proofs against tuning.
- Extract only the `steps`/`jobs` frontmatter and leave the prompt body in each workflow. Rejected: the `## Pre-activation context` prompt text is also identical aside from label name interpolation; leaving it in workflows would leave one source of duplication.

### Decision: Keep `lib/issue-slots.js` as the canonical logic, include from shared script via `//include:`

**Rationale:** `issue-slots.js` already has unit tests (`issue-slots.test.mjs`). The shared script should inline the same logic, not duplicate it.

**Path resolution:**
```
.github/workflows-src/shared/scripts/compute_issue_slots.inline.js
  → //include: ../../lib/issue-slots.js
```
This is two parent directories up into `workflows-src/lib/`, which is correct.

### Decision: Place the shared component under `.github/workflows-src/shared/`

**Rationale:** This mirrors the existing convention where `.github/workflows/shared/` holds compiled GH AW shared workflows. The custom compiler outputs to `.github/workflows/shared/issue-slots.md`.

### Decision: Add the shared component to `manifest.json` with an explicit `output`

**Rationale:** The custom compiler currently produces one output per manifest entry. Shared imports still need to be compiled to `.github/workflows/shared/*.md` before `gh aw compile` can resolve them. By adding an entry to `manifest.json`, `make workflow-generate` produces the compiled shared workflow in the right place.

**Manifest entry:**
```json
{
  "template": ".github/workflows-src/shared/issue-slots.md.tmpl",
  "output": ".github/workflows/shared/issue-slots.md"
}
```

### Decision: Use `steps` + `jobs` in the shared import (not `needs`)

**Rationale:** GH AW merges `steps` from imports into the pre-activation job and merges top-level `jobs` into the compiled lock file. The `pre-activation` job name is reserved/conventional. By having the shared file define both `steps` and `jobs.pre-activation`, the importing workflow gets them automatically without needing its own `jobs:` block for gating.

## Risks / Trade-offs

- **[Risk] The custom compiler doesn't know about GH AW imports and might mis-handle `x-script-include` in the shared template.** → Mitigation: The custom compiler treats each template independently. It expands `x-script-include` before `gh aw compile` processes imports. No compiler changes needed; the shared template is just another source file.

- **[Risk] Consumer workflows might accidentally introduce conflicting `steps` or `jobs` names.** → Mitigation: The importing workflows currently define their own `steps` and `jobs.pre-activation` with the same ID. After the change, they will remove those blocks. GH AW import merging will prepend/merge without conflict since the consumers will no longer define duplicates. Risk is transient during the change only.

- **[Risk] `gh aw compile` fails if the shared workflow is not generated before it runs.** → Mitigation: The manifest already drives generation order. `make workflow-generate` runs the custom compiler first, then `gh aw compile`. The shared entry will be generated before consumers are compiled.

- **[Risk] Prompt ordering — shared body is prepended before consumer body.** → Accepted behavior. The `## Pre-activation context` should precede the `## Task` section in the agent prompt. GH AW's merge behavior prepends imported body content, which matches the desired ordering.

## Migration Plan

1. Create `.github/workflows-src/shared/` directory structure
2. Write `shared/issue-slots.md.tmpl` with frontmatter (`import-schema`, `steps`, `jobs`) and body
3. Write `shared/scripts/compute_issue_slots.inline.js` with `//include: ../../lib/issue-slots.js`
4. Update `manifest.json` with the new entry
5. Simplify the three consumer templates: remove duplicated `steps`, `jobs`, and `## Pre-activation context`; add `imports: - uses: shared/issue-slots.md` with `with: label/cap`
6. Delete the three old `scripts/compute_issue_slots.inline.js` files and their parent `scripts/` directories
7. Run `make workflow-generate` to compile everything
8. Verify the generated `.github/workflows/shared/issue-slots.md` and updated consumer `.md` files look correct
9. Verify `make check-workflows` passes
10. Commit

**Rollback:** Revert the commit. The old triplicated files are deleted but can be restored from git history.
