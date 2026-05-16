## 1. Mapping semantic equality update

- [x] 1.1 Add scalar-aware semantic equality coverage in `internal/elasticsearch/index/mappings_value_test.go` for boolean, numeric, and negative-difference cases
- [x] 1.2 Update `internal/elasticsearch/index/mappings_value.go` so scalar leaf values compare semantically equal when Elasticsearch returns the equivalent stringified scalar echo
- [x] 1.3 Run `go test ./internal/elasticsearch/index/...` to verify mapping semantic equality behavior

## 2. Index settings null canonicalization

- [x] 2.1 Add unit coverage in the relevant custom type tests for settings values where practitioner-authored JSON `null` must compare equal to Elasticsearch `"null"` echoes
- [x] 2.2 Update `internal/utils/customtypes/index_settings_value.go` so flattened settings canonicalize JSON `null` consistently with Elasticsearch string echoes
- [x] 2.3 Run `go test ./internal/utils/customtypes/...` to verify settings semantic equality behavior

## 3. Resource regression coverage

- [x] 3.1 Add focused component-template acceptance coverage for boolean-valued `template.mappings` that must remain non-drifting after apply
- [x] 3.2 Add focused component-template acceptance coverage for `template.settings` containing JSON `null` values that must remain non-drifting after apply
- [x] 3.3 If an Elasticsearch stack is available, run targeted acceptance tests for `internal/elasticsearch/index/componenttemplate/...`

## 4. Requirements and final validation

- [ ] 4.1 Verify the updated delta specs for `elasticsearch-index-component-template` and `elasticsearch-index-template` match the implemented behavior
- [ ] 4.2 Run `make check-openspec` or `make check-lint` to validate OpenSpec artifacts
- [ ] 4.3 Run `make build` to ensure the provider still compiles after the custom type and test updates
