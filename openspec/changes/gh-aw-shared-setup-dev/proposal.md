## Why

Nine agentic workflows in `.github/workflows-src/` duplicate identical setup boilerplate: configure Go, Terraform, Node.js, export chroot paths, and run `make setup`. This duplication is error-prone (inconsistent chroot export variants, order-of-operations bugs like TF before chroot), hard to maintain (version bumps, new tools), and hides a key insight: independent of the agent's actual task, it always needs the same dev tooling to work with this Go/Terraform repository. Extracting this into a single shared GH-AW import component eliminates ~60 lines of duplicated YAML per workflow.

## What Changes

- **New**: Create `.github/workflows/shared/setup-dev.yml` — an opinionated, zero-option shared workflow that installs Go, Terraform, Node.js, exports GOROOT/GOPATH/GOMODCACHE/TERRAFORM_BIN to `$GITHUB_ENV`, and runs `make setup`.
- **Remove**: Delete the Elastic Stack setup scaffolding (`make docker-fleet`, `make set-kibana-password`, `make create-es-api-key`, `make setup-kibana-fleet`, `docker compose logs`) from `code-factory-issue` and `reproducer-factory-issue` workflows. These services are not accessible within the agent's chroot sandbox; they add dead weight.
- **Remove**: Delete the `npm ci` step from `change-factory-issue` and `research-factory-issue`; `make setup` subsumes it.
- **Remove**: Delete all inline dev-setup steps from the agent phase of `changelog-generation`, `kibana-spec-impact`, `schema-coverage-rotation`, `openspec-verify-label`, `ci-deadcode-removal-rotation` (agent phase only), `change-factory-issue`, `research-factory-issue`, `code-factory-issue`, and `reproducer-factory-issue`.
- **Modify**: Add `imports: [shared/setup-dev.yml]` to the frontmatter of all 9 workflow `.md.tmpl` files.
- **Modify**: Run the workflow compiler (`make compile-workflows`) to regenerate `.md` and `.lock.yml` files.

## Capabilities

### New Capabilities
- `gh-aw-shared-setup-dev`: Shared GH-AW import for installing and exporting Go, Terraform, Node.js, and running `make setup` as pre-agent steps.

### Modified Capabilities
<!-- No existing spec-level behavior changes. This is purely workflow infrastructure refactoring. -->

## Impact

- `.github/workflows-src/*/workflow.md.tmpl` — 9 templates modified (steps deleted, imports added)
- `.github/workflows/shared/setup-dev.yml` — new shared file
- `.github/workflows/*.md` — regenerated via workflow compiler
- `.github/workflows/*.lock.yml` — regenerated via `gh aw compile`
- `code-factory-issue` and `reproducer-factory-issue` lose ~10 lines of non-functional Elastic Stack setup each
