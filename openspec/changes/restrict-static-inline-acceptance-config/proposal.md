## Why

The current acceptance-test config lint rule still treats any `ExternalProviders` step with `Config` as compliant, even when the Terraform module is defined directly in Go as a raw string, a `fmt.Sprintf` heredoc, or a helper that returns static HCL. That leaves a readability and maintainability gap in the exact compatibility-test pattern that recently required a manual cleanup in the watcher acceptance suite.

The repository already prefers checked-in Terraform fixtures for ordinary acceptance steps, and compatibility steps can follow the same fixture discipline without losing the ability to pass `Config` directly to `resource.TestStep`. Tightening this exception now lets the analyzer catch the mechanically detectable class of static inline config that should instead live in `testdata/.../main.tf` and be loaded through package-level `go:embed`.

## What Changes

- Tighten the existing acceptance-test config lint capability so `ExternalProviders` steps no longer accept arbitrary `Config` expressions.
- Require compatibility-step `Config` values to resolve to package-level string variables populated by `//go:embed` from a checked-in `testdata/.../main.tf` fixture.
- Reject raw string literals, `fmt.Sprintf` expressions, concatenated strings, and helper-function-returned Terraform modules when used as `Config` for `ExternalProviders` compatibility steps.
- Keep ordinary current-provider steps on `ConfigDirectory: acctest.NamedTestCaseDirectory(...)`; this change only narrows the compatibility-step exception path.
- Add analyzer diagnostics and regression coverage for allowed embedded-fixture configs and rejected static inline-config patterns.
- Migrate existing compatibility tests that still define static Terraform config in Go to the embedded-fixture pattern before enabling the stricter rule.

## Capabilities

### New Capabilities
- _(none)_

### Modified Capabilities
- `acceptance-test-config-directory-lint`: compatibility steps that use `ExternalProviders` and `Config` must source that config from package-level embedded `main.tf` fixtures instead of static Terraform strings defined in Go

## Impact

- **Analyzer implementation**: `analysis/acctestconfigdirlintplugin` will need declaration-aware validation for `resource.TestStep.Config` expressions.
- **Acceptance tests**: SDK/backwards-compatibility tests that currently build static Terraform config in Go will need fixture extraction to `testdata/.../main.tf` plus package-level `//go:embed` variables.
- **Regression coverage**: analyzer tests will need positive coverage for embedded config vars and negative coverage for raw literals, `fmt.Sprintf`, and helper-returned strings.
- **Lint workflow**: `make check-lint` will fail on new compatibility-step regressions once the stricter rule is wired in.
