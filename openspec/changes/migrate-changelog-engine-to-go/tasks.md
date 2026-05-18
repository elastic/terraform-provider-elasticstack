## 1. Bootstrap Go module skeleton

- [x] 1.1 Create directory `scripts/changelog/` with `main.go` exposing a no-op subcommand router (`gather-evidence`, `run-engine`, `manage-unreleased-pr`, `refresh-release-pr`, `validate-pr-section`) that prints usage and exits non-zero on unknown subcommands.
- [x] 1.2 Create `scripts/changelog/internal/githubx/` with `client.go` (oauth2 + `go-github/v86` client construction following `scripts/auto-approve/main.go`), `event.go` (parse `$GITHUB_EVENT_PATH`), and `output.go` (write `$GITHUB_OUTPUT` with heredoc framing).
- [x] 1.3 Pin `github.com/google/go-github/v86` in `go.mod` to the same version `scripts/auto-approve/` uses; run `make vendor`.
- [x] 1.4 Add `go test ./scripts/changelog/...` to the existing `Workflows CI` test path (via `Makefile` `workflow-test` target or `go test ./scripts/...` umbrella).
- [x] 1.5 Verify `go build ./scripts/changelog` succeeds and `go test ./scripts/changelog/...` passes (empty pass).

## 2. Implement shared section parser

- [x] 2.1 Implement `internal/section/parse.go` with `Parse([]byte) (Section, error)` that reproduces the fenced-code-aware extraction from `lib/pr-changelog-parser.js` (`` ``` `` and `~~~` fences, `## ` terminator outside fences).
- [x] 2.2 Implement `internal/section/customer_impact.go` with the `{none, fix, enhancement, breaking}` enum, JSON marshalling, and a `RequiresSummary()` predicate.
- [x] 2.3 Implement `internal/section/breaking.go` to extract the optional `### Breaking changes` subsection verbatim.
- [x] 2.4 Implement `internal/section/render.go` with `Render(Section, RenderOpts) string` producing the `CHANGELOG.md`-entry shape currently emitted by `lib/changelog-renderer.js`.
- [x] 2.5 Port every test case from `lib/pr-changelog-parser.test.mjs` into `internal/section/parse_test.go` table tests + `testdata/` golden files.
- [x] 2.6 Port the renderer invariants from `lib/changelog-renderer.test.mjs` (and any sibling renderer test files) into `internal/section/render_test.go`.

## 3. Implement CHANGELOG.md rewriter

- [x] 3.1 Implement `internal/rewriter/section.go` reproducing `lib/changelog-rewriter.js`'s `rewriteChangelogSection` (replace `## [Unreleased]` or `## [x.y.z] - <date>` block in place, preserving surrounding sections).
- [x] 3.2 Implement `internal/rewriter/links.go` reproducing the link-table maintenance from `lib/changelog-rewriter.js` (preserve sort order, dedupe).
- [x] 3.3 Port `lib/changelog-rewriter.test.mjs` cases into `internal/rewriter/section_test.go` and `internal/rewriter/links_test.go`, including idempotence checks.

## 4. Implement engine modes

- [x] 4.1 Implement `internal/semver/tags.go` with `ListReleaseTags(execer Execer) ([]Tag, error)` wrapping the `git tag --list "v[0-9]*..." --sort=-version:refname` command from `lib/changelog-engine-factory.js`.
- [x] 4.2 Implement `internal/semver/select.go` reproducing `selectPreviousTag` and `buildCompareRange` selection logic.
- [x] 4.3 Implement `internal/engine/engine.go` orchestrating the `unreleased` and `release` modes: validate inputs, list PRs in compare range, assemble per-PR sections via `internal/section`, write `CHANGELOG.md` via `internal/rewriter`.
- [x] 4.4 Implement `internal/engine/validate.go` reproducing `validateModeAndTargetVersion` (mode whitelist + semver target check).
- [x] 4.5 Wire `run-engine` subcommand in `main.go` to `internal/engine.Run(...)` with flags `--mode`, `--target-version`, plus env-derived repo context.
- [x] 4.6 Port test cases from `lib/changelog-engine.test.mjs`, `lib/changelog-engine-factory.test.mjs` (mode validation, semver selection, assembly failure messages) into `internal/engine/engine_test.go`.

## 5. Implement evidence gathering

- [x] 5.1 Implement `internal/evidence/gather.go` reproducing `lib/changelog-pr-evidence.js` and `changelog/gather-pr-evidence.js`: list PRs in a release range, parse each `## Changelog` section via `internal/section`, accumulate an evidence manifest.
- [x] 5.2 Implement `internal/evidence/manifest.go` reproducing the JSON manifest format from `lib/changelog-evidence-manifest.js`.
- [x] 5.3 Wire `gather-evidence` subcommand to `internal/evidence.Gather(...)`; emit the manifest path/payload via `$GITHUB_OUTPUT`.
- [x] 5.4 Port test cases from `lib/changelog-pr-evidence.test.mjs` and `lib/changelog-evidence-manifest.test.mjs` into `internal/evidence/*_test.go`.

## 6. Implement PR management subcommands

- [ ] 6.1 Implement `internal/prmgmt/release_context.go` reproducing `lib/changelog-release-context.js`.
- [ ] 6.2 Implement `internal/prmgmt/manage_unreleased.go` reproducing the unreleased-PR portion of `lib/changelog-pr-management.js` (open/update PR on the `generated-changelog` branch).
- [ ] 6.3 Implement `internal/prmgmt/refresh_release.go` reproducing the release-PR portion of `lib/changelog-pr-management.js`.
- [ ] 6.4 Wire `manage-unreleased-pr` and `refresh-release-pr` subcommands in `main.go`.
- [ ] 6.5 Port test cases from `lib/changelog-pr-management.test.mjs` and `lib/changelog-release-context.test.mjs` into `internal/prmgmt/*_test.go`.

## 7. Implement PR-body validator

- [ ] 7.1 Implement `internal/prcheck/validate.go` reproducing `lib/pr-changelog-check.js` + `pr-changelog-check/check.js`: load PR body via go-github, parse via `internal/section`, apply `no-changelog` label suppression, format the verdict payload.
- [ ] 7.2 Wire `validate-pr-section` subcommand in `main.go`; emit the `result_json` (or equivalent existing key) via `$GITHUB_OUTPUT` so downstream YAML expressions remain valid.
- [ ] 7.3 Port test cases from `lib/pr-changelog-check.test.mjs` and any `pr-changelog-check/*.test.mjs` into `internal/prcheck/validate_test.go` table tests covering each customer-impact variant, missing/empty summary, and `no-changelog` suppression.

## 8. Switch workflows over and delete JS

- [ ] 8.1 Update `.github/workflows/changelog-generation.yml`: replace each `actions/github-script` step that calls a `lib/changelog-*` or `changelog/*` module with a `run: go run ./scripts/changelog <subcommand>` step; preserve `env:` keys for inputs; remove now-unused `actions/github-script` blocks.
- [ ] 8.2 Update `.github/workflows/pr-changelog-check.yml`: replace its `actions/github-script` step with `run: go run ./scripts/changelog validate-pr-section`; preserve `env:` plumbing and output key names.
- [ ] 8.3 Add `actions/setup-go` to each workflow above if not already present (re-using the version pin from `provider.yml`).
- [ ] 8.4 Run `make workflow-generate` to regenerate any agentic `.lock.yml` files that referenced the deleted JS modules.
- [ ] 8.5 Delete `.github/scripts/workflows/lib/changelog-*.js`, `.github/scripts/workflows/lib/changelog-*.test.mjs`, `.github/scripts/workflows/changelog/*.js`, `.github/scripts/workflows/lib/pr-changelog-*.js`, `.github/scripts/workflows/lib/pr-changelog-*.test.mjs`, and `.github/scripts/workflows/pr-changelog-check/*.js`.
- [ ] 8.6 Remove the deleted `.test.mjs` paths from any explicit Makefile enumeration in `workflow-test`.
- [ ] 8.7 Verify `git grep "lib/changelog-"` and `git grep "pr-changelog-check"` return no hits outside `openspec/` and `CHANGELOG.md` itself.

## 9. Behavioural-parity verification

- [ ] 9.1 Run `make workflow-test` and confirm the suite exits zero with no JS test enumeration errors.
- [ ] 9.2 Run `go test ./scripts/changelog/... -count=1` and confirm all ported tests pass.
- [ ] 9.3 On a feature branch, fabricate a PR body with each customer-impact variant (`none`, `fix`, `enhancement`, `breaking`) and confirm `validate-pr-section` produces the same verdict output as the (pre-migration) JS version on the equivalent body.
- [ ] 9.4 Run the `changelog-generation` workflow against a controlled merge and diff the produced `CHANGELOG.md` against a recent JS-produced reference; confirm byte-equivalent output (or document any intentional formatting differences).
- [ ] 9.5 Confirm `make build` and `make check-lint` still pass.

## 10. Documentation and finalisation

- [ ] 10.1 Add a short README at `scripts/changelog/README.md` documenting subcommands, env-var contract, and invocation examples — mirroring the style of `scripts/kibana-spec-impact/`'s docs.
- [ ] 10.2 If `dev-docs/high-level/repo-structure.md` enumerates the JS clusters being deleted, update it to point at `scripts/changelog/` instead.
- [ ] 10.3 Run `npx openspec validate migrate-changelog-engine-to-go --strict` and resolve any issues.
- [ ] 10.4 Open a PR using the standard commit message conventions; include before/after `git diff --stat` highlighting LOC reduction.
