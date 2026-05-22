## 1. Conditional validator support

- [ ] 1.1 Update `internal/utils/validators/conditional.go` so `AllowedIf...` evaluation treats unknown dependent values as allowed.
- [ ] 1.2 Add a required options argument to the `AllowedIf` validator constructors so call sites can explicitly allow a null/unset dependent value.
- [ ] 1.3 Add or update unit tests in `internal/utils/validators/conditional_test.go` for matching, null, unknown, and non-matching dependent values.

## 2. Kibana space validation

- [ ] 2.1 Update `internal/kibana/spaces/resource_schema.go` to replace unconditional `ConflictsWith` validation with the options-enabled conditional validator for `disabled_features` and `solution`.
- [ ] 2.2 Update resource-level tests or acceptance coverage for `elasticstack_kibana_space` to cover `solution = "classic"`, omitted `solution`, and known non-`classic` `solution` values when `disabled_features` is configured.

## 3. Validation and verification

- [ ] 3.1 Run targeted validator and Kibana space tests to confirm the new conditional behavior.
- [ ] 3.2 Run the relevant OpenSpec validation/check commands and ensure the change artifacts are apply-ready.
