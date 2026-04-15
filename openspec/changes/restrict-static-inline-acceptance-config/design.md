## Context

The current `acceptance-test-config-directory-lint` capability enforces relationships between `Config`, `ConfigDirectory`, `ExternalProviders`, and `ProtoV6ProviderFactories` inside inline `resource.TestCase` literals. That catches ordinary inline-config drift, but it still treats every `ExternalProviders` step with `Config` as valid regardless of whether the Terraform module is a checked-in fixture or an ad hoc string assembled in Go.

This leaves a narrow but recurring exception path in SDK/backwards-compatibility tests: contributors can hide static Terraform modules in raw string literals, `fmt.Sprintf` heredocs, or helper functions that return HCL. The desired repository pattern is stricter and still mechanical: compatibility steps may continue using `Config`, but only when `Config` references a package-level string variable populated by `//go:embed` from `testdata/.../main.tf`.

## Goals / Non-Goals

**Goals:**

- Extend the existing analyzer so it validates the declaration shape behind `resource.TestStep.Config` for `ExternalProviders` compatibility steps.
- Allow a single accepted compatibility-step config pattern: `Config` points to a package-level embedded string fixture sourced from `testdata/.../main.tf`.
- Reject mechanically detectable static inline-config patterns defined in Go, including raw literals, `fmt.Sprintf`, concatenation, and helper-returned Terraform modules.
- Preserve the existing ordinary-step rule set: current-provider steps continue using `ConfigDirectory: acctest.NamedTestCaseDirectory(...)`.
- Provide diagnostics that tell contributors to move static Terraform into fixture files and load it through `//go:embed`.
- Add regression tests and migrate existing compatibility tests that still define static Terraform inside Go before enabling the stricter lint behavior.

**Non-Goals:**

- Following full data flow through arbitrary helper stacks, aliases, or interprocedural value propagation beyond the direct declaration referenced by `Config`.
- Validating the semantic contents of the Terraform module beyond checking that the config comes from a checked-in embedded fixture.
- Relaxing the existing rule set to allow new forms of ordinary inline config.
- Reworking acceptance-test APIs outside the analyzer and the affected compatibility tests.

## Decisions

- **Extend the existing `go/analysis` plugin instead of adding a regex rule**: Pattern matching can catch raw heredocs, but it becomes noisy and incomplete once `Config` values flow through identifiers. The existing analyzer already has typed access to the relevant `resource.TestStep` literals, so adding declaration-aware checks there gives more precise diagnostics with less false-positive risk.

- **Use declaration-aware validation, not full SSA/data-flow**: The analyzer should inspect the `Config` expression shape and, when it is an identifier, resolve that identifier to its declaration. For v1, the accepted declaration shape is a package-level `string` variable with an attached `//go:embed` directive that targets a `testdata/.../main.tf` fixture. This captures the real repository pattern without taking on the maintenance cost of whole-program flow analysis.

- **Treat non-identifier `Config` expressions as invalid for compatibility steps**: Raw string literals, `fmt.Sprintf(...)`, concatenation, selector-based builders, and helper-function calls should all produce diagnostics. The fix is always the same: extract the static module to `testdata/.../main.tf`, embed it at package scope, and reference that variable from `Config`.

- **Require direct package-scope ownership for the embedded fixture variable**: The accepted config source should be declared in the same test package as the acceptance test, not passed through helper layers. That keeps the fixture reference auditable next to the test and avoids analyzer complexity around imported symbols or transitive aliases.

- **Preserve existing `ExternalProviders` semantics while narrowing the source contract**: Compatibility steps still use `ExternalProviders` plus `Config`, and they still must not use `ConfigDirectory` or `ProtoV6ProviderFactories`. This change narrows only where the `Config` string may originate.

- **Migrate existing violations before enabling enforcement**: There are already compatibility tests in the repo that use static Terraform strings in Go. Those tests need conversion to embedded fixtures as part of rollout so the stricter analyzer does not fail immediately on known debt.

## Risks / Trade-offs

- **[Risk] The accepted source shape is stricter than the minimum needed to load equivalent Terraform** -> Mitigation: document the exact approved pattern in the spec and diagnostics so contributors know to use `//go:embed` plus `testdata/.../main.tf` instead of inventing variants.

- **[Risk] Some compatibility tests may currently rely on light parameterization inside helper functions** -> Mitigation: move runtime values into `ConfigVariables` where possible and keep the embedded fixture static.

- **[Risk] A declaration-only check may miss exotic aliasing patterns** -> Mitigation: accept that narrower scope in v1 because it targets the real repository convention and keeps the analyzer predictable. Revisit broader data-flow only if concrete false negatives emerge.

- **[Risk] Tightening the rule may require touching several acceptance tests in one rollout** -> Mitigation: stage the implementation with targeted searches for `ExternalProviders` plus `Config` and convert existing static inline configs before turning the rule on in failing lint paths.

## Migration Plan

1. Update the `acceptance-test-config-directory-lint` delta spec to describe the new compatibility-step config-source restriction and expected diagnostics.
2. Extend `analysis/acctestconfigdirlintplugin` so `ExternalProviders` steps validate the origin of `Config`.
3. Add analyzer regression tests for accepted embedded fixture vars and rejected static inline-config patterns.
4. Convert existing compatibility tests that still define Terraform modules in Go to `testdata/.../main.tf` plus package-level `//go:embed`.
5. Run targeted analyzer tests and repository lint to verify the stricter rule is enforceable without repository-wide failures.

## Open Questions

- None for the initial change; the v1 contract is intentionally narrow and based on the repository’s preferred fixture pattern.
