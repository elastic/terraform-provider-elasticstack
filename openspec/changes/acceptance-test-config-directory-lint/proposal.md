## Why

Provider acceptance tests that use `resource.Test` or `resource.ParallelTest` are moving toward directory-backed Terraform fixtures so each `resource.TestStep` loads its full configuration from checked-in testdata instead of assembling HCL inline in Go. That convention improves readability, makes fixture reuse easier, and keeps step intent separate from Terraform text, but today it is enforced only socially.

The repo already contains ordinary acceptance steps that still use inline `Config`, and a few `ConfigDirectory` steps that bypass `acctest.NamedTestCaseDirectory(...)` in favor of lower-level helpers. The one intentional exception is compatibility coverage that provisions state with an older external provider version. Those steps must continue to use inline `Config` because they do not load fixtures through the current provider's test-case directory convention.

The same tests also still rely on `ProtoV6ProviderFactories` inherited from the enclosing `resource.TestCase`, which hides whether a specific step is exercising the current provider or a backwards-compatibility path. To make that intent explicit, provider wiring must move to the individual `resource.TestStep`: ordinary steps declare `ProtoV6ProviderFactories`, while backwards-compatibility steps declare `ExternalProviders`.

## What Changes

- Introduce a new OpenSpec capability for lint rules that enforce how acceptance-test `resource.TestStep` values source Terraform configuration in any Go test file that uses `resource.Test` or `resource.ParallelTest`.
- Add a custom Go analyzer, following the existing `analysis/esclienthelper` pattern, rather than relying on `gocritic` ruleguard rules.
- Require ordinary in-scope acceptance steps to use `ConfigDirectory` with `acctest.NamedTestCaseDirectory(...)`.
- Allow inline `Config` only for steps that also declare `ExternalProviders`, which marks the older-provider compatibility path.
- Treat `ExternalProviders` steps as the explicit exception path: they use inline `Config` and do not use `ConfigDirectory`.
- Forbid `ProtoV6ProviderFactories` on the enclosing `resource.TestCase`; provider wiring must be declared per step instead of inherited at the test-case level.
- Require every in-scope `resource.TestStep` to declare exactly one provider-wiring mode: `ProtoV6ProviderFactories` for ordinary coverage or `ExternalProviders` for backwards-compatibility coverage.
- Treat `ExternalProviders` as a backwards-compatibility-only marker, not as a general substitute for ordinary provider-managed steps.
- Wire the analyzer into repository lint execution and add regression coverage for compliant and non-compliant step shapes.
- Migrate existing in-scope tests that currently violate the convention so the new rule can be enabled without leaving known failures behind.

## Capabilities

### New Capabilities
- `acceptance-test-config-directory-lint`: Provider lint requirements for how acceptance-test `resource.TestStep` values in `resource.Test` and `resource.ParallelTest` flows supply Terraform configuration.

### Modified Capabilities
- _(none)_

## Impact

- **Specs**: new capability under `openspec/changes/acceptance-test-config-directory-lint/specs/acceptance-test-config-directory-lint/spec.md`
- **Analyzer implementation**: new custom analyzer package following the `analysis/esclienthelper` / plugin-module pattern
- **Lint wiring**: `golangci-lint` integration and `make check-lint`
- **Acceptance tests**: in-scope tests anywhere in the repo that currently use inline `Config`, raw `config.TestNameDirectory()`, or test-case-level `ProtoV6ProviderFactories` will need fixture- and step-wiring-oriented updates
- **Regression coverage**: analyzer tests for valid directory-backed steps, valid external-provider exception steps, valid step-local `ProtoV6ProviderFactories`, and invalid mixed, inherited, or bypassed configurations
