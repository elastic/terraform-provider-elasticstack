## 1. Pin the target ref and reproduce both breakages

- [ ] 1.1 Record one immutable Kibana `main` commit SHA as `github_ref` in `generated/kbapi/Makefile`
      (replacing `23436cdb1b7104ea2973e3eee33e3d9c481ce15a`)
- [ ] 1.2 Run `make -C generated/kbapi all` against that pinned ref and confirm the `N0`/`N1` compile
      failure and the `kbapi.SLOs*` → undefined-type build failures both reproduce, capturing the actual
      downloaded `oas.yaml` for the remaining steps
- [ ] 1.3 Diff the downloaded spec's `Cases_attachment_type_filter` parameter and its inline array
      schema against the existing `Cases_string`/`Cases_string_array` components (item type,
      `maxItems`/`minItems`, uniqueness, `style`/`explode`), and record whether reuse is safe or a
      dedicated array component is needed
- [ ] 1.4 Diff the downloaded spec's renamed SLO component names against today's `SLOs_*` names to
      confirm the `KibanaHTTPAPIsSLO*` prefix assumption and enumerate every type this proposal's code
      references

## 2. Fix the Cases_attachment_type_filter build break (Problem 1)

- [ ] 2.1 Add `fixCasesAttachmentTypeFilterParameter` to `generated/kbapi/transform_schema.go`,
      pointing the parameter's `oneOf` branches at `Cases_string`/`Cases_string_array` (or a new
      dedicated array component, per 1.3), and register it in the `transformers` slice
- [ ] 2.2 Add `TestFixCasesAttachmentTypeFilterParameter` to `generated/kbapi/transform_schema_test.go`,
      asserting the fixture's inline `oneOf` becomes the expected refs and that the array branch's
      `maxItems` bound is preserved (or intentionally different, per 1.3)
- [ ] 2.3 Regenerate and confirm the `N0`/`N1` compile errors are gone

## 3. Add the regeneration-time schema-drift guard

- [ ] 3.1 Add a guard step (run first in `transformers`, or as a pre-transform check invoked from
      `transform_schema.go`'s `main`) that fails `make -C generated/kbapi all` if the downloaded spec is
      missing the source node for `fixCasesAttachmentTypeFilterParameter`, `fixSloFloatFormats`, or
      `fixSloResponseArtifacts`
- [ ] 3.2 Verify the guard actually fails when a targeted path is renamed (e.g. temporarily rename a
      target path in a local test run) and passes against the current pinned spec

## 4. Re-target the SLO-specific transforms (part of Problem 2)

- [ ] 4.1 Update `fixSloFloatFormats` to set `format: double` on the renamed equivalents of
      `objective.target`, `objective.timesliceTarget`, the histogram good/total `from`/`to` fields, APM
      latency `threshold`, and the timeslice metric threshold/percentile fields
- [ ] 4.2 Update `fixSloResponseArtifacts` to add the `artifacts` ref on the renamed SLO response schema
- [ ] 4.3 Add `TestFixSloFloatFormats` and `TestFixSloResponseArtifacts` fixture tests to
      `generated/kbapi/transform_schema_test.go`
- [ ] 4.4 Regenerate and confirm (by reading `kibana.gen.go`) that the renamed objective/threshold
      fields are `float64`, not `float32`, and that the SLO response type includes `Artifacts`

## 5. Regenerate the client and migrate SLO domain/model code

- [ ] 5.1 Run `make -C generated/kbapi all` against the pinned ref with both fixes in place and confirm
      `generated/kbapi/kibana.gen.go` builds cleanly on its own
- [ ] 5.2 Rename straightforward `kbapi.SLOs*` → `kbapi.KibanaHTTPAPIsSLO*` references with unchanged
      shape across `internal/models/slo.go`, `internal/kibana/slo/*.go`, and
      `internal/clients/kibanaoapi/slo.go`
- [ ] 5.3 Rebuild the `TimeWindow` construction in `internal/kibana/slo/models.go` (`toAPIModel`) against
      the renamed type's actual shape
- [ ] 5.4 Rebuild `kqlTFFormToKqlWithFiltersUnion` and `kqlWithFiltersAPIToTFFormFilter` in
      `internal/kibana/slo/models_kql_custom_indicator.go` against the renamed `KqlWithFilters` union's
      actual accessor set
- [ ] 5.5 Rewrite APM latency indicator mapping (`models_apm_latency_indicator.go`) to construct/decode
      params through the renamed type's union wrapper accessors instead of a plain struct literal
- [ ] 5.6 Rewrite APM availability indicator mapping (`models_apm_availability_indicator.go`) the same
      way as 5.5
- [ ] 5.7 Confirm `internal/kibana/slo/schema.go` and `resource.go` need no changes beyond the renamed
      type references (or make the changes if the regenerated shape requires it)
- [ ] 5.8 Confirm the `SLOsArtifacts`/dashboards anonymous-struct shape (with its `Id` field and
      `(nolint/redacted):revive` converters in `internal/kibana/slo/models.go`) is unchanged after the
      rename, or update the conversion if it is not

## 6. Update and extend unit test coverage

- [ ] 6.1 Update `internal/kibana/slo/models_api_mapping_test.go` and every per-indicator `*_test.go`
      file to compile against the renamed types
- [ ] 6.2 Update `internal/clients/kibanaoapi/slo_test.go` to compile against the renamed types
- [ ] 6.3 Add or extend unit tests covering the reshaped `TimeWindow` construction, the reshaped
      `KqlWithFilters` union round-trip (string and object-form KQL inputs), and the reshaped APM
      latency/availability union construction and decoding
- [ ] 6.4 Add a float-precision regression test asserting `objective.target` and
      `objective.timesliceTarget` round-trip as `float64` with no precision loss (guards against a
      repeat of issue #2396)

## 7. Build and acceptance validation

- [ ] 7.1 Run `make build` and confirm it succeeds with no remaining `kbapi.SLOs*` references under
      `internal/`
- [ ] 7.2 Run the full `internal/kibana/slo` and `internal/clients/kibanaoapi` unit test suites
- [ ] 7.3 Run `internal/kibana/slo/acc_test.go`'s existing acceptance suite against a stack version that
      supports the regenerated spec, confirming every indicator type, `group_by` variant, `settings`,
      and `artifacts`/dashboards case still produces an identical Terraform plan/state versus
      pre-migration behavior, with no unexpected resource replacement
- [ ] 7.4 Confirm the existing SLO version-gating tests (`models_version_requirements_test.go`,
      `acc_test.go`'s `versionutils.CheckIfVersionMeetsConstraints` skips) still pass unchanged

## 8. Close out the Renovate bump

- [ ] 8.1 Close or supersede Renovate PR #5464 once this change merges and CI is green
- [ ] 8.2 Confirm Renovate resumes bumping `generated/kbapi/Makefile`'s `github_ref` cleanly afterward

## 9. Out of scope — tracked separately, not part of this change

- [ ] 9.1 File a follow-up issue for `sli.synthetics.availability` indicator support, including its
      unresolved minimum-version and wire-shape open questions from `design.md`
