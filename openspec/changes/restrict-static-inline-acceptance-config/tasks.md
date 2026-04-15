## 1. Extend analyzer behavior

- [x] 1.1 Update `analysis/acctestconfigdirlintplugin` so `ExternalProviders` steps validate the shape of their `Config` expression instead of accepting any `Config`.
- [x] 1.2 Resolve identifier-based `Config` expressions to their declarations and accept only package-level `string` variables populated by `//go:embed` from `.tf` fixtures under `testdata/`.
- [x] 1.3 Emit actionable diagnostics for rejected compatibility-step config sources, including raw string literals, `fmt.Sprintf` expressions, concatenated strings, helper-returned strings, and non-embedded variables.

## 2. Migrate existing compatibility tests

- [x] 2.1 Identify existing `ExternalProviders` compatibility steps that still define static Terraform config inside Go.
- [x] 2.2 Extract each static compatibility fixture to `.tf` files under `testdata/` and load it through package-level `//go:embed` variables.
- [x] 2.3 Replace helper- or literal-based compatibility-step `Config` usage with embedded fixture vars and move runtime values into `ConfigVariables` where needed.

## 3. Add regression coverage

- [x] 3.1 Add analyzer tests for accepted compatibility steps whose `Config` references package-level embedded fixture variables.
- [x] 3.2 Add analyzer tests for rejected compatibility steps that use raw string literals, formatted strings, helper-returned strings, and non-embedded variables for `Config`.
- [x] 3.3 Preserve or extend existing analyzer tests covering ordinary directory-backed steps and provider-wiring rules so the stricter compatibility check does not regress current behavior.

## 4. Validate rollout

- [x] 4.1 Run targeted analyzer tests for `analysis/acctestconfigdirlintplugin`.
- [x] 4.2 Run repository lint to confirm the stricter rule passes after compatibility-test migrations.
