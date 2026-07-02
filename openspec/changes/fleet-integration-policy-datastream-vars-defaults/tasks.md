## 1. Fix stream vars comparison: strip `data_stream.*` keys

- [x] 1.1 Add a `serverManagedVarsKeys` package-level variable (or constant slice) in `internal/fleet/integration_policy/` listing `["data_stream.type", "data_stream.dataset"]`
- [x] 1.2 Add a helper `stripServerManagedVarsKeys(vars jsontypes.Normalized) (jsontypes.Normalized, diag.Diagnostics)` in `input_value.go` (or a new adjacent file) that unmarshals the JSON, deletes each key in `serverManagedVarsKeys` from the map if present, and returns the re-marshaled normalized JSON; if the input is null/unknown, return it unchanged
- [x] 1.3 In `compareStreams` (in `input_value.go`), call `stripServerManagedVarsKeys` on both `oldStream.Vars` and `newStream.Vars` before passing them to `StringSemanticEquals`; diags from the strip helper SHALL be appended and, if errored, `compareStreams` SHALL return false
- [x] 1.4 Verify that the existing `applyDefaultsToVars` call in `compareStreams` / `ObjectSemanticEquals` is not affected (the stripping is a post-defaults-application step on the comparison side only)

## 2. Fix `defaults` null ⇄ populated-object transition

- [x] 2.1 In `InputValue.ObjectSemanticEquals` (in `input_value.go`), after extracting `oldInput` and `newInput`, add a check: if either `oldInput.Defaults` or `newInput.Defaults` is null/unknown (i.e. `!typeutils.IsKnown`), treat the `defaults` component as semantically equal (skip the equality check for `defaults` and proceed to compare `vars` and `streams`)
- [x] 2.2 Verify that fully-known `defaults` values on both sides also do not block semantic equality (i.e. `defaults` is treated as semantically equal even when the two sides differ), and that `vars` and `streams` are still compared normally
- [x] 2.3 Confirm no schema change is needed (the `defaults` attribute is already `Computed: true` in the schema; no plan modifier or state upgrader required)

## 3. Unit tests

- [x] 3.1 Add a unit test in `input_value_test.go` for `compareStreams` (via `InputValue.ObjectSemanticEquals` or directly if exported in `export_test.go`): two inputs where the API-side stream `vars` contains `data_stream.type` and `data_stream.dataset` in addition to the user-configured keys, and the plan-side stream `vars` contains only the user-configured keys → result SHALL be semantically equal
- [x] 3.2 Add a unit test for the `stripServerManagedVarsKeys` helper (exported via `export_test.go` or tested via `compareStreams`): JSON with `data_stream.type` and `data_stream.dataset` → both stripped; JSON without those keys → returned unchanged; null input → null returned
- [x] 3.3 Add a unit test in `input_value_test.go` for `InputValue.ObjectSemanticEquals`: old value with `defaults = null`, new value with a populated `defaults` object → SHALL return semantically equal
- [x] 3.4 Add a unit test for the symmetric case: old value with a populated `defaults`, new value with `defaults = null` → SHALL return semantically equal
- [x] 3.5 Add a unit test confirming that two fully-known, differing `defaults` values (different vars content) are still treated as semantically equal (since `defaults` is purely server-managed and the user never configures it — equality of `defaults` should not produce a diff even when the two sides differ)

## 4. Acceptance test verification

- [ ] 4.1 Run `TestAccResourceIntegrationPolicyGCPPubSub` against a 9.5.0-SNAPSHOT (or latest 9.5) stack and confirm it passes without modifying the test configuration
- [ ] 4.2 Run `TestAccResourceIntegrationPolicySecrets` (both `single_valued_secrets` and `multi-valued_secrets` subtests) against a 9.5.0-SNAPSHOT stack and confirm both pass
- [ ] 4.3 If either test fails for a reason other than the two gaps fixed in tasks 1–2, record the additional failure mode and open a follow-up issue or task

## 5. Validation and build

- [ ] 5.1 Run `make build` — confirm no compilation errors
- [ ] 5.2 Run `make check-lint` — fix any lint issues
- [ ] 5.3 Run `make check-openspec` — confirm the change validates cleanly
- [ ] 5.4 Run the unit test suite for the integration policy package: `go test ./internal/fleet/integration_policy/...` — all existing and new tests SHALL pass
