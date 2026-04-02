# Tasks: Acceptance Test Config Directory Lint

## 1. Define the lint capability

- [ ] 1.1 Add the `acceptance-test-config-directory-lint` delta spec describing in-scope `resource.TestStep` enforcement for any `_test.go` file that uses `resource.Test` or `resource.ParallelTest`.
- [ ] 1.2 Confirm the spec captures the single accepted inline-config exception: steps that also declare `ExternalProviders`.

## 2. Implement the analyzer

- [ ] 2.1 Add a new custom analyzer package and plugin-module wrapper following the existing `analysis/esclienthelper` structure.
- [ ] 2.2 Detect in-scope `resource.TestStep` composite literals in any `_test.go` file within `resource.Test` and `resource.ParallelTest` acceptance-test flows, without path-based exclusions.
- [ ] 2.3 Report a violation when an in-scope step sets `Config` without also setting `ExternalProviders`.
- [ ] 2.4 Report a violation when an in-scope step sets `ConfigDirectory` with anything other than a direct `acctest.NamedTestCaseDirectory(...)` call.
- [ ] 2.5 Report a violation when an in-scope step sets `ExternalProviders` but does not use inline `Config`, or mixes `ExternalProviders` with `ConfigDirectory`.
- [ ] 2.6 Ensure diagnostics explain the accepted ordinary-step and compatibility-step shapes.

## 3. Wire the rule into repository lint

- [ ] 3.1 Register the analyzer with `golangci-lint` using the repo's existing custom-plugin pattern.
- [ ] 3.2 Update lint configuration so `make check-lint` fails when the acceptance-test config rule is violated.

## 4. Migrate existing in-scope violations

- [ ] 4.1 Convert ordinary in-scope acceptance tests that still use inline `Config` to directory-backed fixtures with `acctest.NamedTestCaseDirectory(...)`, including `provider/**` tests when applicable.
- [ ] 4.2 Replace in-scope `ConfigDirectory: config.TestNameDirectory()` usage with `acctest.NamedTestCaseDirectory(...)` and corresponding fixture directories.
- [ ] 4.3 Preserve legitimate older-provider compatibility coverage by keeping `ExternalProviders` steps on inline `Config`.

## 5. Add regression coverage

- [ ] 5.1 Add analyzer tests for compliant ordinary steps that use `ConfigDirectory: acctest.NamedTestCaseDirectory(...)`.
- [ ] 5.2 Add analyzer tests for compliant compatibility steps that use `ExternalProviders` with inline `Config`.
- [ ] 5.3 Add analyzer tests for invalid ordinary inline `Config`, invalid `ConfigDirectory` helper usage, and invalid mixed `ExternalProviders` plus `ConfigDirectory` shapes.
- [ ] 5.4 Run targeted analyzer tests and repository lint checks to confirm compliant cases pass and violations fail as specified.
