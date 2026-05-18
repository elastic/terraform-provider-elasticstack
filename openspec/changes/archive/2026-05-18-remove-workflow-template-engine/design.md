## Context

The repository currently uses a custom Go template compiler (`scripts/compile-workflow-sources`) to generate GitHub Actions workflow files. The compiler performs two operations:

1. **`x-script-include:` expansion** — replaces a pseudo-YAML key with an inline `script: |` block loaded from a `.inline.js` file, recursively expanding nested `//include:` directives.
2. **`//include:` expansion** — concatenates shared JavaScript libraries into inline scripts.

This adds build-time complexity:
- A Go module, tests, and Makefile targets dedicated to workflow generation
- A `.github/workflows-src/` tree with `.tmpl` files, `.inline.js` wrappers, and a `manifest.json`
- Generated `.yml`/`.md` files carry "Do not edit directly" headers
- Two compilation steps in sequence (`compile-workflow-sources` → `gh aw compile`)

`actions/github-script@v9` supports `require()` for external files when the repository is checked out. This makes the entire template engine unnecessary.

## Goals / Non-Goals

**Goals:**
- Eliminate `scripts/compile-workflow-sources` and `.github/workflows-src/`
- Make `.github/workflows/*.yml` and `.github/workflows/*.md` the source of truth
- Centralize reusable workflow script logic under `.github/scripts/workflows/`
- Preserve all existing workflow behavior (no functional changes to CI gates, issue intake, etc.)
- Keep `gh aw compile` for agentic workflows (`.md` → `.lock.yml`)

**Non-Goals:**
- Changing the behavior of any workflow logic (gate rules, issue eligibility, changelog logic, etc.)
- Migrating from `actions/github-script` to composite actions or JavaScript actions
- Upgrading `gh aw` or `actions/github-script` versions
- Modifying `.lock.yml` files directly (they remain compiled artifacts)

## Decisions

### 1. Module location: `.github/scripts/workflows/`

**Rationale**: Keeps CI scripts co-located with `.github/workflows/` while avoiding confusion with per-workflow YAML files. The nested `workflows/` subdirectory signals these are workflow-specific scripts, distinct from other `.github/scripts/` that might exist.

**Alternative considered**: Top-level `.github/scripts/` — rejected because it could collide with other non-workflow scripts and lacks clear namespacing.

### 2. Module signature: CommonJS default export accepting `{github, context, core}`

**Rationale**: Matches `actions/github-script`'s injected globals. Wrapping modules as async functions makes them composable and testable without `eval()` tricks.

```js
// .github/scripts/workflows/provider/classify-changes.js
const { classifyChanges } = require('../lib/classify-changes.js');

module.exports = async function({ github, context, core }) {
  // orchestration: event routing, API calls, env mapping
  // ...
  core.setOutput('provider_changes', result.providerChanges);
};
```

YAML consumption:
```yaml
script: |
  const classify = require('${{ github.workspace }}/.github/scripts/workflows/provider/classify-changes.js');
  await classify({ github, context, core });
```

**Alternative considered**: Pure CommonJS exports consumed in YAML directly (e.g. `core.setOutput(..., require('...').classifyChanges(...))`) — rejected because it pushes too much orchestration into YAML, making the script harder to read and test.

### 3. Keep `lib/` modules under `.github/scripts/workflows/lib/`

**Rationale**: Pure logic functions (e.g. `classifyChanges`, `gateProvider`) are already unit-tested via the existing `.test.mjs` suite. Moving them preserves those tests with minimal path updates.

### 4. Redundant wrappers are deleted; substantial scripts are kept and moved

**Rationale**: Scripts with non-trivial orchestration (event routing, API pagination, state machine logic) deserve their own file. Trivial glue (read 4 env vars, call a pure function, set output) folds into the YAML.

**Example — trivial (delete, inline in YAML)**:
```js
const { gateProvider } = require('../../lib/gate-provider.js');
const result = gateProvider({ classifyResult: '${{ needs... }}', ... });
```

**Example — substantial (keep, move to module)**:
```js
// ~150 lines of event-driven issue classification
```

### 5. Add `actions/checkout` where missing

**Rationale**: `require()` needs files on disk. Three `.yml` workflows (`provider.yml`, `workflows.yml`, `pr-changelog-check.yml`) run `actions/github-script` before any checkout step. Adding a lightweight checkout (with `persist-credentials: false`) ensures `github.workspace` resolves correctly.

### 6. `//include:` replaced by `require()`

**Rationale**: The `//include:` directive was pre-processor sugar. In a module system, `require()` is the native mechanism. This eliminates the include-expansion test suite (`code-factory-inline-scripts.test.mjs`, `research-factory-inline-scripts.test.mjs`).

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Module path mismatch between `require()` and filesystem | All paths use `${{ github.workspace }}` prefix + relative path; validated by running affected workflows in a branch |
| `github.workspace` undefined in workflows without checkout | Add checkout step (see Decision 5); only 3 workflows affected |
| Lost test coverage from deleted `//include:` expansion tests | `lib/*.test.mjs` tests preserved and updated; integration test is the workflow run itself |
| Accidental behavioral drift during hand-editing | Generate diff between old `.yml`/`.md` and new hand-edited versions for human review; CI validates via `check-workflows` Makefile target removal |
| `gh aw compile` scripts still reference deleted `.inline.js` paths in comments | `.lock.yml` files are compiled artifacts; re-running `gh aw compile` after `.md` edits regenerates them cleanly |

## Migration Plan

1. **Bootstrap** `.github/scripts/workflows/` tree with moved modules
2. **Delete** `scripts/compile-workflow-sources/` and `.github/workflows-src/`
3. **Update** Makefile targets
4. **Edit** `.github/workflows/*.yml` — add checkout where needed, replace inline `script:` with `require()`
5. **Edit** `.github/workflows/*.md` — replace `x-script-include:` with `require()`
6. **Update** unit test paths and delete expansion tests
7. **Build** to verify no Go regressions
8. **Run** affected workflows on a test branch
9. **Regenerate** `.lock.yml` files via `gh aw compile`

Rollback: Restore deleted files from git. The change is purely deletive/relocating — no database migrations or API changes.

## Open Questions

- None at this time. All decisions confirmed during exploration.
