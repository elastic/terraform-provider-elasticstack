## Why

The `scripts/compile-workflow-sources` Go template engine and `x-script-include` mechanism add unnecessary build-step complexity to GitHub Actions workflow authoring. `actions/github-script` natively supports `require()` to load external files ([docs](https://github.com/actions/github-script#run-a-separate-file)), which eliminates the need for a custom pre-compilation step. Removing this indirection simplifies maintenance, reduces CI time, and makes workflow behavior discoverable by reading the committed files directly.

## What Changes

- **Remove** `scripts/compile-workflow-sources/` (Go code, tests, and Makefile references)
- **Remove** `.github/workflows-src/` (templates, manifest, and redundant `.inline.js` scripts)
- **Move** reusable JavaScript logic to `.github/scripts/workflows/` as standard CommonJS modules
- **Replace** all `x-script-include:` directives with `require()` calls in workflow YAML/Markdown
- **Add** missing `actions/checkout` steps to workflows that now need repository files on disk
- **Delete** tests that assert on `//include:` expansion behavior
- **Keep** `gh aw compile` for agentic workflows (`.md` → `.lock.yml`); this is unaffected

## Capabilities

### New Capabilities
- `workflow-script-modules`: Establishes `.github/scripts/workflows/` as the canonical home for shared GitHub Actions script modules, defines the module API contract (CommonJS exports, `{github, context, core}` signatures, output conventions), and describes how inline YAML `script:` blocks consume them via `require()`.

### Modified Capabilities
- *(None. No provider resource or data source behavior changes.)*

## Impact

- GitHub Actions workflows (`.github/workflows/*.yml` and `.github/workflows/*.md`)
- Makefile targets: `workflow-generate`, `workflow-test`, `check-workflows`
- Workflow unit tests (`.github/workflows-src/lib/*.test.mjs` → `.github/scripts/workflows/lib/*.test.mjs`)
- Repository root: `scripts/` directory shrinks by one sub-project
