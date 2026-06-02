## Why

`elasticstack_fleet_output` updates fail with `Provider produced inconsistent result after apply: .config_yaml: inconsistent values for sensitive attribute` whenever the Fleet API echoes a `config_yaml` value in the response that does not match the planned value ([issue #1856](https://github.com/elastic/terraform-provider-elasticstack/issues/1856)). The Fleet API treats an omitted `config_yaml` in an update body as "no change" and serializes the stored value (or an empty string for outputs that never had one) back to the client. Because the attribute is marked Sensitive, Terraform's post-apply consistency check rejects the null↔value mismatch and the apply errors out.

## What Changes

- The shared output reader (`fromAPICommonFields`) normalizes nil and empty-string `config_yaml` from the Fleet API into a null state value, so outputs that never had a `config_yaml` stop oscillating between null and `""`.
- The shared reader also preserves a null `config_yaml` from the existing model across the Fleet API echo on both the update and refresh paths, so removing the attribute from configuration applies cleanly and stays clean on subsequent refresh even when Fleet preserves the previously stored value server-side. Import is excluded from this preservation by checking the existing model's required `name` field, which is null on import but populated after any successful create / read / update.
- The `config_yaml` schema attribute adopts `customtypes.NormalizedYamlType`, so semantically equivalent YAML (whitespace, key reordering, anchor expansion) returned by the API no longer triggers spurious updates or post-apply consistency errors.
- Resource documentation calls out that the Fleet API does not support clearing `config_yaml` via omission; users must delete and re-create the output to fully clear the stored value.

## Capabilities

### New Capabilities
None.

### Modified Capabilities
- `fleet-output`: refine REQ-017 state-mapping behavior to specify Fleet's `config_yaml` echo handling and add the plan-intent preservation requirement.

## Impact

- `internal/fleet/output/schema.go` — `config_yaml` becomes `customtypes.NormalizedYamlType{}`; description documents the Fleet-side limitation.
- `internal/fleet/output/models.go` — `ConfigYaml` field migrates to `customtypes.NormalizedYamlValue`; `fromAPICommonFields` calls a new `configYamlFromAPI` helper that folds empty/nil to null and preserves a user-removed null across the API echo (skipping the preservation on import).
- `internal/fleet/output/update.go` — uses the consolidated reader behavior; no Update-specific preservation needed.
- `internal/fleet/output/models_config_yaml_test.go` and a new `models_config_yaml_unit_test.go` cover normalization and semantic-equality behavior.
- `internal/fleet/output/issue_1856_acc_test.go` flips from `ExpectError` to a passing two-step apply, with the configs moved to `testdata/TestAccReproduceIssue1856/{create,update}/`.
- `CHANGELOG.md` — Unreleased entry referencing issue #1856.
- `docs/resources/fleet_output.md` — regenerated from the updated description.
