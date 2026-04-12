## 1. Add deterministic repository bootstrap to schema-coverage rotation

- [x] 1.1 Update `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl` to install Go from `go.mod`, export `GOROOT`, `GOPATH`, and `GOMODCACHE`, install Node from `package.json`, and run `make setup` before agent reasoning begins.
- [x] 1.2 Update the schema-coverage rotation workflow frontmatter so `network.allowed` includes `defaults`, `node`, and `go`, and ensure the prompt relies on the preconfigured toolchain rather than instructing the agent to install it.

## 2. Rebuild generated workflow artifacts

- [x] 2.1 Recompile `.github/workflows/schema-coverage-rotation.md` from the authored workflow source.
- [x] 2.2 Recompile `.github/workflows/schema-coverage-rotation.lock.yml` and verify the generated artifacts reflect the new bootstrap steps and network policy.

## 3. Validate the change

- [x] 3.1 Run the relevant OpenSpec validation for `schema-coverage-rotation-toolchain-bootstrap` and address any structural issues in the proposal artifacts.
- [x] 3.2 Run the relevant workflow compilation or repository checks needed to confirm the schema-coverage rotation workflow remains valid after the bootstrap changes.
