## 1. Workflow source and compilation

- [x] 1.1 Add the GitHub Agentic Workflow markdown under `.github/workflows/` (align name with the spec, e.g. `openspec-verify-label.md`, or update the spec to match).
- [x] 1.2 Author frontmatter: `on.pull_request.types: [labeled]`; ensure the run only proceeds when label is **`verify-openspec`**; grant **`contents: write`** and **`pull-requests: write`** (or compiler-equivalent); declare **`safe-outputs`** for `create-pull-request-review-comment`, `submit-pull-request-review`, and **`push-to-pull-request-branch`** with limits and `checkout`/`fetch` settings per gh-aw docs; set `engine`, `tools`, and `network` for `npm ci` / `npx openspec` / `openspec archive` as needed.
- [x] 1.3 Write agent instructions: PR file list + gating (single `<id>`, **modified-only** under `openspec/changes/<id>/`, **noop** on **added** files under non-archive change paths or multiple ids); verification via **openspec-verify-change** + `openspec status` / `openspec instructions apply` for the selected id; structural allowlist + relevance review; review body and inline comments; **APPROVE** vs **COMMENT**; **only after APPROVE**, run **`openspec archive <id>`** (or equivalent) then commit and **`push-to-pull-request-branch`**.
- [x] 1.4 Run `gh aw compile` and commit the `.lock.yml`; update [`.github/aw/actions-lock.json`](.github/aw/actions-lock.json) if required.

## 2. Repository settings and documentation

- [x] 2.1 Document label **`verify-openspec`**, required permissions (including Actions approval/push), and that the workflow **archives on APPROVE** so maintainers expect branch updates.
- [x] 2.2 If CI does not run on bot-pushed commits, configure **`github-token-for-extra-empty-commit`** or PAT per [Triggering CI](https://github.github.io/gh-aw/reference/triggering-ci/) as needed.

## 3. Validation

- [x] 3.1 Run `npx openspec validate aw-openspec-verification --strict` and fix any issues.
- [x] 3.2 Run `make build` per [`AGENTS.md`](./AGENTS.md) when the implementation touches non-workflow code.

## 4. Ship and spot-check

- [x] 4.1 Apply **`verify-openspec`** on a PR that **only modifies** files under one `openspec/changes/<id>/`: expect review; if **APPROVE**, expect archive + push.
- [x] 4.2 PR touching two active change dirs or **adding** files under `openspec/changes/<id>/`: expect **`noop`**.
- [x] 4.3 **COMMENT** review: confirm **no** archive and **no** push.
