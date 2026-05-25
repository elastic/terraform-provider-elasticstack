## 1. Runtime guard in `convertGlobalDataTags`

- [ ] 1.1 In `internal/fleet/agentpolicy/models.go`, update the `convertGlobalDataTags`
  method's value-conversion callback to use explicit `IsNull()` / `IsUnknown()` guards
  instead of `ValueStringPointer() != nil`. The new logic must:
  - Check `!item.StringValue.IsNull() && !item.StringValue.IsUnknown()` first; if true,
    call `value.FromAgentPolicyGlobalDataTagsItemValue0(item.StringValue.ValueString())`.
  - Else check `!item.NumberValue.IsNull() && !item.NumberValue.IsUnknown()`; if true,
    call `value.FromAgentPolicyGlobalDataTagsItemValue1(item.NumberValue.ValueFloat32())`.
  - Else (both null/unknown): call `diags.AddAttributeError` with summary
    `"Invalid global_data_tags entry"` and detail
    `"Each entry in global_data_tags must have exactly one of string_value or number_value set."`,
    then return `kbapi.AgentPolicyGlobalDataTagsItem{}`.

## 2. Schema `AtLeastOneOf` validators

- [ ] 2.1 In `internal/fleet/agentpolicy/schema.go`, add
  `stringvalidator.AtLeastOneOf(path.MatchRelative().AtParent().AtName("string_value"), path.MatchRelative().AtParent().AtName("number_value"))`
  to the `Validators` slice of the `string_value` attribute inside the `global_data_tags`
  nested object.
- [ ] 2.2 Add the equivalent `float32validator.AtLeastOneOf(...)` to the `Validators` slice
  of `number_value`, referencing the same two paths.
- [ ] 2.3 Confirm both `stringvalidator` and `float32validator` packages are already imported
  in `schema.go` (they are — used by existing `ConflictsWith` validators); no new import
  needed.

## 3. Unit test

- [ ] 3.1 In `internal/fleet/agentpolicy/models_test.go`, add a test
  `TestConvertGlobalDataTags_NullNullEntry` that:
  - Constructs an `agentPolicyModel` with a `GlobalDataTags` map containing one entry whose
    `string_value` and `number_value` are both null (use `types.StringNull()` and
    `types.Float32Null()`).
  - Calls `model.convertGlobalDataTags(ctx, features{SupportsGlobalDataTags: true})`.
  - Asserts the returned diagnostics contain an error (`diags.HasError() == true`).
  - Asserts no panic occurs (the test passing without `t.FailNow()` from a recovered panic
    is sufficient; alternatively, do not use `recover` — if the code panics the test fails).

## 4. Verify

- [ ] 4.1 Run `make build` to confirm the provider compiles after the changes.
- [ ] 4.2 Run `go test ./internal/fleet/agentpolicy/... -run TestConvertGlobalDataTags` to
  confirm the new unit test passes.
- [ ] 4.3 Optionally run `go test ./internal/fleet/agentpolicy/... -run TestMerge` and
  `TestConvertHostName` to confirm no existing unit tests regressed.
