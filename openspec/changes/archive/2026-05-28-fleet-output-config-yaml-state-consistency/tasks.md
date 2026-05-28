## 1. Schema and types

- [x] 1.1 Add `CustomType: customtypes.NormalizedYamlType{}` to the `config_yaml` schema attribute and update its description to document Fleet's inability to clear via omission.
- [x] 1.2 Change `outputModel.ConfigYaml` from `types.String` to `customtypes.NormalizedYamlValue`.

## 2. Read path

- [x] 2.1 Add `configYamlFromAPI(*string) customtypes.NormalizedYamlValue` helper that folds nil and `""` to null.
- [x] 2.2 Replace `model.ConfigYaml = types.StringPointerValue(d.configYaml)` in `fromAPICommonFields` with a call to the helper.

## 3. Preserve user-removed null across update and refresh

- [x] 3.1 In `fromAPICommonFields`, capture `existingConfigYaml := model.ConfigYaml` before overwriting model fields.
- [x] 3.2 Detect import by checking the existing model's `Name` field (null/unknown = import) so the preservation only applies to update and refresh, not to import.
- [x] 3.3 After populating from the API response, if `existingConfigYaml.IsNull()` and we are not in an import scenario, restore `customtypes.NewNormalizedYamlNull()`.

## 4. Tests

- [x] 4.1 Update `models_config_yaml_test.go` to use `customtypes.NewNormalizedYamlNull()` / `NewNormalizedYamlValue` for the model literal.
- [x] 4.2 Add a unit test covering `configYamlFromAPI` for nil, empty, and non-empty inputs.
- [x] 4.3 Add a unit test covering `fromAPICommonFields` normalization end-to-end.
- [x] 4.4 Add a unit test asserting `NormalizedYamlValue.StringSemanticEquals` returns true for key-reordered YAML.
- [x] 4.5 Convert `issue_1856_acc_test.go` from `ExpectError` to a passing two-step apply (create with config_yaml → update without).
- [x] 4.6 Move the test configs to `testdata/TestAccReproduceIssue1856/{create,update}/` to satisfy `acctestconfigdirlint`.

## 5. Docs and changelog

- [x] 5.1 Regenerate `docs/resources/fleet_output.md` via `make build`.
- [x] 5.2 Add an Unreleased CHANGELOG entry referencing issue #1856.

## 6. Validation

- [x] 6.1 `make build` (fmt + lint + docs render).
- [x] 6.2 `go test ./internal/fleet/output/...` (unit tests; acceptance tests gated by `TF_ACC`).
- [x] 6.3 `make check-openspec`.
