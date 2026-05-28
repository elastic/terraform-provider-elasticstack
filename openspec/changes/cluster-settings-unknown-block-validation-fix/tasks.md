## 1. Fix `categoryBlockEmpty` to treat unknown blocks as non-empty

- [ ] 1.1 In `internal/elasticsearch/cluster/settings/resource.go`, update `categoryBlockEmpty` to return `false` when `block.IsUnknown()`, instead of `true`. Remove `block.IsUnknown()` from the combined `IsNull() || IsUnknown()` check, keep only `block.IsNull()` in that guard, and add a separate early-return branch for `block.IsUnknown()` immediately after the null check.
- [ ] 1.2 In the same function, update the inner `settingSet` check to return `false` when `settingSet.IsUnknown()`, instead of `true`. Remove `settingSet.IsUnknown()` from the combined check at the end of the function.

## 2. Export `categoryBlockEmpty` for unit testing

- [ ] 2.1 In `internal/elasticsearch/cluster/settings/export_test.go`, add an `ExportedCategoryBlockEmpty(block types.Object) bool` function that calls `categoryBlockEmpty(block)`. This follows the existing export pattern in the file.

## 3. Add unit tests for unknown-block handling

- [ ] 3.1 In `internal/elasticsearch/cluster/settings/helpers_test.go`, add `TestValidateConfigModel_BothUnknown_OK`: call `settings.ExportedValidateConfigModel` with two `types.ObjectUnknown(...)` arguments and assert no error is returned.
- [ ] 3.2 Add `TestValidateConfigModel_OneUnknown_OK`: call `settings.ExportedValidateConfigModel` with one `NullSettingsBlock()` and one `types.ObjectUnknown(...)` argument and assert no error is returned.
- [ ] 3.3 Add `TestCategoryBlockEmpty_Unknown_NotEmpty`: call `settings.ExportedCategoryBlockEmpty` with a `types.ObjectUnknown(...)` argument and assert it returns `false`.
- [ ] 3.4 Add `TestCategoryBlockEmpty_UnknownInnerSet_NotEmpty`: construct a known, non-null block object whose `setting` attribute is `types.SetUnknown(...)`, call `settings.ExportedCategoryBlockEmpty` with it, and assert it returns `false`.

## 4. Verify

- [ ] 4.1 Run `go test ./internal/elasticsearch/cluster/settings/...` and confirm all tests pass, including the four pre-existing tests (`BothNull_Error`, `BothEmpty_Error`, `PersistentSet_OK`, `TransientSet_OK`).
