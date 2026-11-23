## Why

`elasticstack_fleet_elastic_defend_integration_policy` never reads or writes the Kibana package-policy field `additional_datastreams_permissions` on either the create (`POST /api/fleet/package_policies`) or update package-policy request. The field lets a user grant the Elastic Agent policy write access to data streams beyond the ones the Endpoint/Elastic Defend package itself owns (for example `logs-nginx-*`). `elasticstack_fleet_integration_policy` and `elasticstack_fleet_managed_integration` already implement this field end-to-end; the Defend resource is the outlier missing it, as flagged by the Kibana API gap scanner in [#5301](https://github.com/elastic/terraform-provider-elasticstack/issues/5301).

This closes that gap by adding `additional_datastreams_permissions` to the Defend resource, following the request/response mapping already proven in `integration_policy`.

A human review comment on the issue (2026-10-16) resolved the research's open questions and changed one piece of the recommended design: the attribute is **Optional + Computed with no revoke-on-unset**, rather than `integration_policy`'s Optional (non-Computed) attribute that sends an empty array to actively revoke permissions when unset. This avoids a one-time migration hazard for Defend users who already have out-of-band permissions configured in Kibana — removing the attribute from Terraform configuration leaves server-side permissions untouched, and an explicit `[]` is how a user clears them. See design.md's "Deviation from research recommendation" section.

## What Changes

- Add `additional_datastreams_permissions` to the Defend resource schema as an `Optional` + `Computed` `ListAttribute` of strings, with `UseStateForUnknown` so an unset attribute does not show a perpetual diff. Unlike `integration_policy`'s `SizeAtLeast(1)` validator, an explicit empty list (`[]`) is accepted here — it is how a user clears previously granted permissions.
- Add the field to `elasticDefendIntegrationPolicyModel` and map it in both `buildBootstrapRequest` and `buildFinalizeRequest` (the plain `Update` path also calls `buildFinalizeRequest`), so both the two-phase create and the update path send the same configured value. A known value (including an empty list) is always sent as `*[]string`; a null/unknown value is always omitted (no revoke-on-unset).
- Map the field back from the API response in `populateModelFromAPI`: a non-empty API list becomes a list value; an API value that is absent or empty maps to an empty list only when state/plan already holds a known empty list, otherwise null.
- Add an attribute-scoped minimum-version requirement (9.1.0, mirroring `integration_policy`'s `MinVersionAdditionalDatastreamsPermissions`) via `entitycore.NewAttributeVersionRequirement`, evaluated only when the attribute is known and non-null. Restructure `elasticDefendIntegrationPolicyModel.GetVersionRequirements` (`models.go:51-54`) to accumulate requirements, since it currently returns early (`nil, nil`) whenever `agent_policy_ids` is unknown, which would otherwise skip the new requirement.
- Document that removing the attribute from configuration does not revoke previously granted permissions server-side, and that setting `[]` is the supported way to clear them.

## Capabilities

### Modified Capabilities

- `fleet-elastic-defend-integration-policy`: Adds the Optional+Computed `additional_datastreams_permissions` list(string) attribute, wired into both the bootstrap and finalize/update request builders and read-back mapping, gated on Elastic Stack >= 9.1.0, with no revoke-on-unset semantics and explicit-empty-list support for clearing.

## Impact

- **Changed code**: `internal/fleet/elastic_defend_integration_policy/schema.go` (new attribute), `models.go` (new model field, restructured `GetVersionRequirements`), `common.go` (`buildBootstrapRequest`, `buildFinalizeRequest`, `populateModelFromAPI`), `resource.go` (new `MinVersionAdditionalDatastreamsPermissions` constant alongside the existing `MinVersionPolicyIDs`).
- **New/updated acceptance tests**: `internal/fleet/elastic_defend_integration_policy/acc_test.go` and its `testdata/` — mirroring `TestAccResourceIntegrationPolicyAdditionalDatastreamsPermissions`, adapted for no-revoke-on-unset and explicit-empty-list-clears semantics, plus a version-gate test parallel to the existing `agent_policy_ids` gate test.
- **New/updated unit tests**: request-body cases (populated, empty list, null/omitted) for both `buildBootstrapRequest` and `buildFinalizeRequest`; `populateModelFromAPI` round-trips including the empty-list state-dependent rule; `GetVersionRequirements` cases for the attribute alone, `agent_policy_ids` alone, both, and neither.
- **Docs**: `docs/resources/fleet_elastic_defend_integration_policy.md` and the resource's example will need the new attribute and a note that unsetting it does not revoke permissions (not produced in this proposal-only change).
- **No `generated/kbapi` changes**: `AdditionalDatastreamsPermissions *[]string` already exists on `kbapi.PackagePolicyCreateRequestTypedInputs`, `kbapi.PackagePolicyUpdateRequestTypedInputs`, and `kbapi.PackagePolicy`; no regeneration is required.
- **Compatibility**: Purely additive. The new attribute is `Optional` + `Computed`, so existing configurations that don't reference it see no diff and no behavior change — unlike `integration_policy`'s shipped behavior, nothing is ever sent for this field unless the user configures it, so there is no first-apply migration hazard for existing Defend policies with out-of-band permissions.
