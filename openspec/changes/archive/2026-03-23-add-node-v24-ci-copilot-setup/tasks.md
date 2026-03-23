## 1. Workflow

- [x] 1.1 Add `actions/setup-node` (commit-SHA pin; align with `.github/workflows/test.yml` or bump both deliberately) after checkout with **`node-version-file: package.json`**, **`cache: npm`**, and **`cache-dependency-path: package-lock.json`**, and do **not** set `node-version` so the file drives the version
- [x] 1.2 Place the step before `make setup` (and before other steps that invoke Node/npm or OpenSpec)

## 2. Canonical spec

- [x] 2.1 Merge the delta in `openspec/changes/.../specs/ci-copilot-setup-steps/spec.md` into `openspec/specs/ci-copilot-setup-steps/spec.md` (update the **Toolchain and checkout** requirement text and scenarios; extend the **Schema** YAML sketch to include the Node setup step consistent with the workflow)

## 3. Verification

- [x] 3.1 Run `make check-openspec` (or `openspec validate --all`) and fix any structural issues
- [ ] 3.2 Optionally run the `copilot-setup-steps` workflow on a branch to confirm `node` on `PATH` satisfies `package.json` `engines` (and any higher-precedence fields read by setup-node) before `make setup`
