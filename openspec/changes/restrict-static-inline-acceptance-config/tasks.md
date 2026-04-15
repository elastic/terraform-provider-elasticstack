## 1. Extend analyzer behavior

- [ ] 1.1 Update `analysis/acctestconfigdirlintplugin` so `ExternalProviders` steps validate the shape of their `Config` expression instead of accepting any `Config`.
- [ ] 1.2 Resolve identifier-based `Config` expressions to their declarations and accept only package-level `string` variables populated by `//go:embed` from `testdata/.../main.tf`.
- [ ] 1.3 Emit actionable diagnostics for rejected compatibility-step config sources, including raw string literals, `fmt.Sprintf` expressions, concatenated strings, helper-returned strings, and non-embedded variables.

## 2. Migrate existing compatibility tests

- [ ] 2.1 Identify existing `ExternalProviders` compatibility steps that still define static Terraform config inside Go.
- [ ] 2.2 Extract each static compatibility fixture to `testdata/.../main.tf` and load it through package-level `//go:embed` variables.
- [ ] 2.3 Replace helper- or literal-based compatibility-step `Config` usage with embedded fixture vars and move runtime values into `ConfigVariables` where needed.

## 3. Add regression coverage

- [ ] 3.1 Add analyzer tests for accepted compatibility steps whose `Config` references package-level embedded fixture variables.
- [ ] 3.2 Add analyzer tests for rejected compatibility steps that use raw string literals, formatted strings, helper-returned strings, and non-embedded variables for `Config`.
- [ ] 3.3 Preserve or extend existing analyzer tests covering ordinary directory-backed steps and provider-wiring rules so the stricter compatibility check does not regress current behavior.

## 4. Validate rollout

- [ ] 4.1 Run targeted analyzer tests for `analysis/acctestconfigdirlintplugin`.
- [ ] 4.2 Run repository lint to confirm the stricter rule passes after compatibility-test migrations.
