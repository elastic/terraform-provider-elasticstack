## 1. Provider registration

- [ ] 1.1 Add `managedintegration.NewResource` to `resources()` in `provider/plugin_framework.go` and remove it from `experimentalResources()` in the same edit, leaving `kibanatag.NewResource` as the sole experimental entry; verify with `grep -n 'managedintegration.NewResource' provider/plugin_framework.go` showing exactly one hit, located inside `resources()`.
- [ ] 1.2 Verify `make build` succeeds.

## 2. Drop the experimental wording

- [ ] 2.1 Rewrite the `**This resource is experimental**: ...` sentence in the `MarkdownDescription` in `internal/fleet/managedintegration/schema.go` to `The underlying Fleet managed integrations API requires Kibana 9.5.0 or later.` — the minimum-version statement lives inside the experimental sentence, so the sentence must be rewritten, not deleted — keeping the Elastic Cloud Hosted / Serverless-only constraint that follows it. Verify per the spec scenario "Resource description makes no experimental claim" with a unit assertion that the description does not contain `experimental` and still names `9.5.0`.
- [ ] 2.2 Remove the ` (experimental API)` suffix from the `ErrorMessage` in `GetVersionRequirements` in `internal/fleet/managedintegration/models.go`, keeping the `v%s or later` phrasing that names 9.5.0; verify the message still names the minimum version.
- [ ] 2.3 Update the two assertions of the exact error string in `internal/fleet/managedintegration/entitycore_contract_test.go` to the new wording, keeping full-string equality — do not loosen them to a bare `Contains("9.5.0")`; the exact pin is what guards against experimental wording creeping back. Verify `go test ./internal/fleet/managedintegration/...` passes.
- [ ] 2.4 Fix the stale comments across the package: in `models.go`, change `experimental, added in Kibana 9.5.0` to `GA, added in Kibana 9.5.0` (only the word "experimental" is stale — the API genuinely was added in 9.5.0); and in `schema.go`, `models.go`, `resource.go`, `acc_test.go`, and `entitycore_contract_test.go`, repoint the `openspec/changes/fleet-managed-integration...` references at the archived location `openspec/changes/archive/2026-07-22-fleet-managed-integration/`. Verify with `grep -rn 'openspec/changes/fleet-managed-integration' internal/fleet/managedintegration/` finding no hits (the archive path does not match this pattern) and `grep -rin experimental internal/fleet/managedintegration/` finding no remaining hits.

## 3. Registration tests

- [ ] 3.1 `git mv provider/plugin_framework_experimental_test.go provider/plugin_framework_registration_test.go`; verify `git status` shows a rename.
- [ ] 3.2 Replace `TestProvider_stableResourcesExcludeManagedIntegrationTypes` and `TestProvider_experimentalResourcesIncludesManagedIntegration` with a stable-registration test asserting `managedIntegrationResourceType` is in `p.resources(ctx)` and absent from `experimentalResources(ctx)`, and that `removedAgentlessPolicyType` is absent from both; verify the new test passes.
- [ ] 3.3 Retarget `TestProvider_managedIntegrationNotRegisteredInDefaultProvider` and `TestProvider_managedIntegrationRegisteredWhenExperimentalEnvEnabled` to `elasticstack_kibana_tag` (adding a `kibanaTagResourceType` constant) so the env-var gate keeps coverage, rename them accordingly, and add an assertion that a provider with `p.version == AccTestVersion` registers `elasticstack_kibana_tag` — that keeps the `acctest` sentinel branch of the gate under failure-detecting coverage once 3.4 makes the managed-integration assertion pass under any version. Verify the tests pass with the env var respectively unset and set to `true`.
- [ ] 3.4 Rework `TestProvider_managedIntegrationRegisteredWithExperimentalAccTestVersion` into a stable-availability test renamed to drop `Experimental` (for example `TestProvider_managedIntegrationRegisteredWithoutOptIn`), asserting the resource is registered under a released version string with `t.Setenv(IncludeExperimentalEnvVar, "")` — dropping `t.Parallel()`, which is incompatible with `t.Setenv`, and guarding against dev environments that export the variable — as well as under `AccTestVersion`, covering the spec scenario "Resource available without the experimental opt-in"; verify it passes.
- [ ] 3.5 Add a test asserting `elasticstack_fleet_managed_integration` is registered exactly once when `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, counting occurrences across `p.Resources(ctx)` rather than using the deduplicating name map; verify it passes, then temporarily re-add the resource to `experimentalResources()`, confirm the test fails, and revert.
- [ ] 3.6 Verify the whole package passes with `go test ./provider/... -count=1`.

## 4. Documentation

- [ ] 4.1 Run `make docs-generate` and verify `docs/resources/fleet_managed_integration.md` is created.
- [ ] 4.2 Review the generated page for wording that assumed the resource was experimental or hidden, and check no other page changed unexpectedly; verify with `git diff --stat docs/`. Add a `templates/resources/fleet_managed_integration.md.tmpl` only if the generated page reads poorly, then re-run `make docs-generate`.
- [ ] 4.3 Remove the paragraph in `examples/resources/elasticstack_fleet_managed_integration/README.md` that instructs exporting `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true` before plan, apply, or import; verify with `grep -rn TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL examples/resources/elasticstack_fleet_managed_integration/` finding no hits.

## 5. Validation

- [ ] 5.1 Verify `make lint` and `make build` pass.
- [ ] 5.2 Verify `make check-openspec` passes (`openspec validate --all`).
- [ ] 5.3 Run the managed-integration acceptance tests against a live stack with `TF_ACC=1` to confirm CRUD behaviour is unchanged (see `dev-docs/high-level/testing.md`). Note: the acceptance harness instantiates the provider with `AccTestVersion`, which registers experimental resources regardless of the env var, so this run cannot prove opt-in-free availability — that coverage lives in the unit tests from tasks 3.2 and 3.4.
- [ ] 5.4 For whoever archives this change: because the delta is REMOVED + ADDED, archive appends **Resource type and stable registration** to the end of `openspec/specs/fleet-managed-integration/spec.md` and deletes the old first requirement. The appended position is acceptable — requirement order is not normative, matching the kibana-dashboard graduation precedent; verify the archived spec still passes `make check-openspec`.
