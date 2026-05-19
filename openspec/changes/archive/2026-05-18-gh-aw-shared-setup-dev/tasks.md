## 1. Create the shared setup component

- [x] 1.1 Create `.github/workflows/shared/setup-dev.yml` with unconditional steps: Setup Go, Setup Terraform CLI, Export Go and Terraform paths for AWF chroot mode, Setup Node.js, Setup repository dependencies (`make setup`)
- [x] 1.2 Verify the YAML is valid and the file has no `on:` field (shared component only)

## 2. Modify agentic workflow templates (import + cleanup)

- [x] 2.1 `.github/workflows-src/changelog-generation/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Go / Export Go paths / Setup Node / Setup repo deps steps
- [x] 2.2 `.github/workflows-src/kibana-spec-impact/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Go / Export Go paths / Setup Node / Setup repo deps steps (keep pre_activation Setup Go for deterministic script execution)
- [x] 2.3 `.github/workflows-src/schema-coverage-rotation/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Go / Export Go paths / Setup Node / Setup repo deps steps
- [x] 2.4 `.github/workflows-src/openspec-verify-label/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Go / Export Go paths / Setup Node / Setup TF / Setup repo deps steps
- [x] 2.5 `.github/workflows-src/ci-deadcode-removal-rotation/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup repository dependencies step (keep pre_activation Go + chroot steps untouched)
- [x] 2.6 `.github/workflows-src/change-factory-issue/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Node.js / Install npm dependencies steps
- [x] 2.7 `.github/workflows-src/research-factory-issue/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Node.js / Install npm dependencies steps
- [x] 2.8 `.github/workflows-src/code-factory-issue/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Go / Export Go paths / Setup Node / Setup TF / Setup Elastic Stack / Get dependencies / Setup Kibana user / Get ES API key / Setup Fleet / Docker compose logs steps
- [x] 2.9 `.github/workflows-src/reproducer-factory-issue/workflow.md.tmpl`: Add `imports: [shared/setup-dev.yml]`, delete agent-phase Setup Go / Export Go paths / Setup Node / Setup TF / Setup Elastic Stack / Get dependencies / Setup Kibana user / Get ES API key / Setup Fleet / Docker compose logs steps

## 3. Regenerate compiled workflow files

- [x] 3.1 Run `make compile-workflows` to regenerate `.github/workflows/*.md` from `.md.tmpl` files
- [x] 3.2 Verify no `*.md` file contains the deleted step patterns anymore (grep for removed step names)
- [x] 3.3 Run `gh aw compile` (or let CI do it via `make check-lint`) to regenerate `.github/workflows/*.lock.yml`

## 4. Verify and validate

- [x] 4.1 Run `make check-lint` and ensure it passes
- [x] 4.2 Inspect at least one regenerated `.lock.yml` to confirm imported steps appear before workflow-specific steps
- [x] 4.3 Confirm no orphaned references to deleted Elastic Stack setup remain in any `.md.tmpl` (grep for `docker-fleet`, `set-kibana-password`, `create-es-api-key`, `setup-kibana-fleet`)
