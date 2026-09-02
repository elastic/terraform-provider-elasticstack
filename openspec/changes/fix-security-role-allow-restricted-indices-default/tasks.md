## 1. Regression test (write first, confirm it reproduces the bug)

- [ ] 1.1 Add an acceptance test in `internal/elasticsearch/security/role/` that: creates the resource with one `indices` element via a `dynamic "indices"` block (matching the issue's repro HCL), then applies a second time with a second `indices` element appended through the same `dynamic` block, omitting `allow_restricted_indices` on both
- [ ] 1.2 Add the equivalent test for `remote_indices`
- [ ] 1.3 Confirm both tests fail against current `main` with "Provider produced inconsistent result after apply" / "does not correlate with any element in actual" before making the schema change

## 2. Schema fix

- [ ] 2.1 In `internal/elasticsearch/security/role/schema.go`, change `attrAllowRestrictedIndices` on the `indices` `SetNestedBlock` from `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` to `Default: booldefault.StaticBool(false)`
- [ ] 2.2 Apply the same change to `attrAllowRestrictedIndices` on the `remote_indices` `SetNestedBlock`
- [ ] 2.3 Add the `github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault` import; drop the `boolplanmodifier` import if no longer used elsewhere in the file
- [ ] 2.4 Confirm the tests from Task 1 now pass

## 3. Unit test updates

- [ ] 3.1 Review `internal/elasticsearch/security/role/models_test.go` (and any other unit test asserting `allow_restricted_indices` is `types.BoolNull()` for a config-omitted case) and update expectations to reflect the new default-resolved value where the assertion covers plan-time resolution
- [ ] 3.2 Run `go vet ./internal/elasticsearch/security/role/...` and the package's unit tests (no `TF_ACC`) to confirm no other coverage regresses

## 4. Documentation

- [ ] 4.1 Confirm `internal/elasticsearch/security/role/resource-description.md` and `descriptions.go` (`allowRestrictedIndicesDescription`) still accurately describe the attribute's default behavior; update wording if it implies "unset means preserved from prior state" rather than "unset means false"

## 5. Spec sync

- [ ] 5.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-security-role-allow-restricted-indices-default --type change` and resolve any reported issues
- [ ] 5.2 Run `make check-openspec` after implementation to confirm specs and code stay in sync
