# Tasks: Acceptance Test Config Directory Lint

## 1. Define the lint capability

- [x] 1.1 Add the `acceptance-test-config-directory-lint` delta spec describing in-scope `resource.TestStep` enforcement for inline `resource.TestCase` composite literals passed directly to `resource.Test` or `resource.ParallelTest` from `_test.go` files.
- [x] 1.2 Confirm the spec captures the single accepted inline-config exception: steps that also declare `ExternalProviders`.

## 2. Implement the analyzer

- [x] 2.1 Add a new custom analyzer package and plugin-module wrapper following the existing `analysis/esclienthelper` structure.
- [x] 2.2 Detect in-scope `resource.TestStep` composite literals inside inline `resource.TestCase` composite literals passed directly to `resource.Test` and `resource.ParallelTest`, without path-based exclusions.
- [x] 2.3 Report a violation when an in-scope step sets `Config` without also setting `ExternalProviders`.
- [x] 2.4 Report a violation when an in-scope step sets `ConfigDirectory` with anything other than a direct `acctest.NamedTestCaseDirectory(...)` call.
- [x] 2.5 Report a violation when an in-scope step sets `ExternalProviders` but does not use inline `Config`, or mixes `ExternalProviders` with `ConfigDirectory`.
- [x] 2.6 Ensure diagnostics explain the accepted ordinary-step and compatibility-step shapes.
- [x] 2.7 Report a violation when an in-scope `resource.TestCase` sets `ProtoV6ProviderFactories`.
- [x] 2.8 Report a violation when an in-scope `resource.TestStep` sets neither `ProtoV6ProviderFactories` nor `ExternalProviders`.
- [x] 2.9 Report a violation when an in-scope `resource.TestStep` sets both `ProtoV6ProviderFactories` and `ExternalProviders`.
- [x] 2.10 Ensure diagnostics explain that ordinary steps declare `ProtoV6ProviderFactories` on the step, while backwards-compatibility steps use `ExternalProviders`.

## 3. Wire the rule into repository lint

- [x] 3.1 Register the analyzer with `golangci-lint` using the repo's existing custom-plugin pattern.
- [x] 3.2 Update lint configuration so `make check-lint` fails when the acceptance-test config rule is violated.

## 4. Migrate existing in-scope violations

- [x] 4.1 Convert ordinary in-scope acceptance tests that still use inline `Config` to directory-backed fixtures with `acctest.NamedTestCaseDirectory(...)`, including `provider/**` tests when applicable.
- [x] 4.2 Replace in-scope `ConfigDirectory: config.TestNameDirectory()` usage with `acctest.NamedTestCaseDirectory(...)` and corresponding fixture directories.
- [x] 4.3 Preserve legitimate older-provider compatibility coverage by keeping `ExternalProviders` steps on inline `Config`.
- [x] 4.4 Move in-scope `ProtoV6ProviderFactories` assignments from `resource.TestCase` onto the individual ordinary `resource.TestStep` values that use the current provider.
- [x] 4.5 Update any in-scope step that currently relies on inherited provider factories so each step explicitly chooses either `ProtoV6ProviderFactories` or `ExternalProviders`.

## 5. Add regression coverage

- [x] 5.1 Add analyzer tests for compliant ordinary steps that use `ConfigDirectory: acctest.NamedTestCaseDirectory(...)`.
- [x] 5.2 Add analyzer tests for compliant compatibility steps that use `ExternalProviders` with inline `Config`.
- [x] 5.3 Add analyzer tests for invalid ordinary inline `Config`, invalid `ConfigDirectory` helper usage, and invalid mixed `ExternalProviders` plus `ConfigDirectory` shapes.
- [x] 5.4 Run targeted analyzer tests and repository lint checks to confirm compliant cases pass and violations fail as specified.
- [x] 5.5 Add analyzer tests for compliant ordinary steps that set step-level `ProtoV6ProviderFactories`.
- [x] 5.6 Add analyzer tests for invalid test-case-level `ProtoV6ProviderFactories`, missing step-level provider wiring, and invalid mixed `ProtoV6ProviderFactories` plus `ExternalProviders` shapes.
- [x] 5.7 Run targeted analyzer tests and lint checks for the new provider-wiring scenarios.
