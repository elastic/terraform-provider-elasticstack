## 1. Guard `toAPICreateRequest` against Unknown/Null `role_descriptors`

- [ ] 1.1 In `internal/elasticsearch/security/api_key/models.go`, in the
  `toAPICreateRequest()` method, wrap the `model.RoleDescriptors.Unmarshal(...)` block with
  a `typeutils.IsKnown(model.RoleDescriptors)` guard (matching the existing pattern used
  for `Metadata`). When the guard is false, leave `roleDescriptors` as nil and skip
  `req.RoleDescriptors` assignment.
- [ ] 1.2 Apply the same `typeutils.IsKnown(model.RoleDescriptors)` guard to
  `toUpdateAPIRequest()` in the same file, which has an identical unconditional
  `Unmarshal` call.

## 2. Guard `validateRestrictionSupport` against Unknown/Null `role_descriptors`

- [ ] 2.1 In `internal/elasticsearch/security/api_key/create.go`, in
  `validateRestrictionSupport()`, add a guard at the top: if
  `model.RoleDescriptors.IsNull() || model.RoleDescriptors.IsUnknown()`, return immediately
  with no diagnostics (no restrictions can be present when there are no role descriptors).

## 3. Add acceptance test for creating an API key without `role_descriptors`

- [ ] 3.1 Create directory
  `internal/elasticsearch/security/api_key/testdata/TestAccResourceSecurityAPIKeyNoRoleDescriptors/create/`.
- [ ] 3.2 In that directory, add `main.tf` containing a minimal
  `elasticstack_elasticsearch_security_api_key` resource with only `name` set (no
  `role_descriptors`, no `expiration`), and a `variable "api_key_name"` input.
- [ ] 3.3 In `internal/elasticsearch/security/api_key/acc_test.go`, add
  `TestAccResourceSecurityAPIKeyNoRoleDescriptors` that:
  - generates a random `api_key_name`
  - uses `acctest.NamedTestCaseDirectory("create")` pointing at the new testdata
  - asserts `name`, `api_key`, `encoded`, `id`, and `key_id` are set
  - uses `SkipFunc: versionutils.CheckIfVersionIsUnsupported(apikey.MinVersion)`

## 4. Verification

- [ ] 4.1 `make build` passes.
- [ ] 4.2 `make check-lint` passes.
- [ ] 4.3 `make check-openspec` passes.
- [ ] 4.4 Unit tests for the `apikey` package pass (`go test ./internal/elasticsearch/security/api_key/...`).
- [ ] 4.5 Acceptance test `TestAccResourceSecurityAPIKeyNoRoleDescriptors` passes against a
  running Elasticsearch stack.
