## 1. Pre-implementation probes (resolve open questions before committing to clearing/hydration behavior)

- [x] 1.1 Against a live Kibana ≥ 9.1, create an `elasticstack_fleet_server_host`-equivalent server host with `ssl` fully populated (`certificate`, `key`, `certificate_authorities`, `client_auth`) via the generated client
- [x] 1.2 PUT an update that omits `ssl` entirely; independently GET the host afterward; record whether `ssl` persisted or was cleared
- [x] 1.3 Against a fresh copy of the same starting state, PUT an update sending `ssl = {}`; GET and record the result
- [x] 1.4 Against another fresh copy of the same starting state, send a PUT body containing explicit JSON `"ssl": null` (not an omitted field or a nil generated-client pointer with `omitempty`); GET and record the result
- [x] 1.5 Against four independent fresh copies of the same fully-populated starting state, PUT four separate updates that each omit exactly one child (`certificate` only, `key` only, `certificate_authorities` only, `client_auth` only); GET each afterward and record whether the omitted child cleared while the retained children persisted
- [x] 1.6 Record whether the create/update/GET responses echo `key` verbatim or redact/omit it
- [x] 1.7 Confirm the minimum Kibana version that accepts a non-null `ssl` body on `fleet_server_hosts` create/update (expected 9.1.0 per Kibana release notes kibana#208091; verify against the `kibana-openapi-spec` changelog pinned in this repo if available)
- [x] 1.8 Record all probe results in design.md "Open questions" / "Decision 3" / "Decision 4"; decide, per the pre-specified branches in the delta spec, whether `toAPIUpdateModel`'s removal path needs a `RequiresReplace` plan-modifier fallback (whole-object, per-child, or neither), and whether `populateFromAPI` needs to preserve a configured `key` instead of hydrating it from the response
- [x] 1.9 Create a host with `client_auth` omitted; inspect the create and subsequent GET responses for a server-defaulted value, and record the result before finalizing the schema and hydration strategy

## 2. Model and mapping

- [x] 2.1 Create `internal/fleet/serverhost/models_ssl.go` adapted from `internal/fleet/output/models_ssl.go`: a `serverHostSslModel` embedding `sslcore.Values` plus a `ClientAuth types.String` sibling field, `objectValueToSSL`, and `sslToObjectValue` against the kbapi anonymous `Ssl` struct on `PostFleetFleetServerHostsJSONBody`/`PutFleetFleetServerHostsItemidJSONBody`/`ServerHost`
- [x] 2.2 Update `internal/fleet/serverhost/models.go`: add `Ssl types.Object` to `serverHostModel`; wire `objectValueToSSL`/`objectValueToSSLUpdate` into `toAPICreateModel` and `toAPIUpdateModel`; apply the Task 1 probe decision to `toAPIUpdateModel`'s removal path; update `populateFromAPI` to hydrate `Ssl` from the response, applying the Task 1 probe decision for `key` (hydrate verbatim, or preserve the configured/prior value if redacted) and `client_auth` (Optional-only if no server default, otherwise a plan-consistent Optional+Computed strategy). Ensure omitted/unknown `client_auth` is omitted from the create request.
- [x] 2.3 Add `GetVersionRequirements(_ context.Context) ([]entitycore.VersionRequirement, diag.Diagnostics)` to `serverHostModel` (does not exist today), returning an attribute-scoped requirement on `path.Root("ssl")` at the version confirmed in Task 1.7, appended only when `ssl` is known/configured — following `internal/fleet/agentdownloadsource/models.go`'s `GetVersionRequirements` pattern
- [x] 2.4 Before any PUT that would replace `ssl`, inspect the current server-side SSL object for non-null `agent_*`/`es_*` fields. If present and the update cannot preserve them, return an actionable diagnostic before mutation; do not silently drop these deferred fields.
- [x] 2.5 If Task 1.4 finds explicit JSON `ssl: null` is the only in-place clearing form, implement a request path that can serialize the explicit null (rather than passing a nil generated `Ssl` pointer with `omitempty`) and add a serialized-payload assertion for `"ssl": null`. (Not applicable: probe 1.4 showed `{}` clears in place, so no explicit-null path is needed.)

## 3. Schema

- [x] 3.1 Add the `ssl` `schema.SingleNestedAttribute` (Optional) to `internal/fleet/serverhost/schema.go` with `certificate` (`schema.StringAttribute`, Optional), `key` (`schema.StringAttribute`, Optional, Sensitive), `certificate_authorities` (`schema.ListAttribute`, Optional, string elements), `client_auth` (`schema.StringAttribute`, Optional)
- [x] 3.2 Add validators: `certificate`/`key` `stringvalidator.LengthAtLeast(1)`; `certificate_authorities` `listvalidator.SizeAtLeast(1)`, `listvalidator.SizeAtMost(10)`, and `listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))`; `client_auth` `stringvalidator.OneOf("none", "optional", "required")`
- [x] 3.3 Add `internal/fleet/serverhost/ssl_validator.go`: an object-level validator on `ssl` requiring at least one of `certificate`, `key`, `certificate_authorities`, `client_auth` to be set (reject `ssl = {}`), adapted from `internal/fleet/agentdownloadsource/ssl_validator.go` with `client_auth` added to the checked sub-attribute list
- [x] 3.4 If Task 1's probe requires it, add a `planmodifier.Object` (and/or per-child plan modifiers) implementing `RequiresReplace` for the specific removal transition(s) the API does not honor in place (Not applicable: every removal transition is honored in place per the Task 1 probes.)

## 4. CRUD wiring

- [x] 4.1 Confirm `create.go`/`update.go`/`read.go` call through `models.go`'s `toAPICreateModel`/`toAPIUpdateModel`/`populateFromAPI` without needing resource-specific changes beyond what Task 2 adds (the resource already converges through these three functions for its other attributes)

## 5. Unit tests

- [x] 5.1 Add `internal/fleet/serverhost/models_ssl_test.go` covering: full round-trip (all four fields set), CA-only, certificate+key only, client_auth-only, API-returns-all-nil normalizes to `types.ObjectNull`, empty CA list normalizes to `types.ListNull`
- [x] 5.2 Add a schema unit test asserting `ssl` exposes exactly `certificate`, `key` (sensitive), `certificate_authorities`, and `client_auth`, and has the validators from Task 3.2-3.3
- [x] 5.3 Add `GetVersionRequirements` unit tests for `ssl` absent, configured, and unknown (plan-time unknown value) states

## 6. Acceptance tests

- [x] 6.1 Add `TestAccResourceFleetServerHost_SSL` to `internal/fleet/serverhost/acc_test.go`, version-skipped below the Task 1.7-confirmed floor via `versionutils.SkipIfUnsupported`, modeled on `internal/fleet/agentdownloadsource/acc_test.go`'s SSL tests
- [x] 6.2 Step `with_ssl`: create with `certificate`, `key`, `certificate_authorities`, and `client_auth` set; assert round-trip via `TestCheckResourceAttr`; assert `ssl.key` is a sensitive value via `plancheck.ExpectSensitiveValue(..., tfjsonpath.New("ssl").AtMapKey("key"))`
- [x] 6.3 Step `with_ssl_updated`: change all four fields to different values; assert round-trip and no-diff-after-apply
- [x] 6.4 Step `clear_ssl`: remove `ssl` entirely; assert the outcome matches the Task 1 probe decision — either `TestCheckNoResourceAttr(..., "ssl")` after an in-place update, or that the step triggers resource replacement; additionally confirm via a subsequent read that Kibana itself no longer reports `ssl` as set
- [x] 6.5 Per-child removal: four separate transitions from a fully-populated `ssl` to one with a single child dropped (`drop_key`, `drop_cas`, `drop_cert`, `drop_client_auth`); assert in-place update or `RequiresReplace` per the Task 1 probe decision for each child
- [x] 6.6 Explicit `ssl = {}` step: assert `terraform plan`/`apply` fails with the object-level validator error
- [x] 6.7 Validator-rejection cases, each asserted separately: `certificate = ""`, `key = ""`, `certificate_authorities = []`, a nonempty CA list containing one empty-string element, and an invalid `client_auth` value
- [x] 6.8 CA-list boundary: exactly 10 entries accepted; 11 entries rejected
- [x] 6.9 Import: confirm `ssl` round-trips through the resource's existing composite/plain import ID handling, applying the Task 1.5 probe decision for `key`
- [x] 6.10 Version-gating cases, explicit rather than skip-only: (a) successful `ssl` create/update on a stack at or above the Task 1.7-confirmed floor; (b) applying an `ssl`-configured create or update against a stack below the floor produces the attribute-specific `GetVersionRequirements` error before any Fleet mutation request; (c) a config that omits `ssl` entirely still applies successfully below the floor, confirming the gate is per-attribute and not resource-wide (Cases (b) and (c) require a stack below 9.1.0; only 9.5.5 was available, so they are implemented in `TestAccResourceFleetServerHost_SSLRejectedBelowMinimumVersion` and skip on newer stacks. Below-floor gating is covered by unit tests of `GetVersionRequirements`.)
- [x] 6.11 Upgrade compatibility: verify that an unchanged configuration omitting `ssl` produces no `ssl` diff after provider upgrade when the existing host has no server-side `ssl`; separately seed a host with pre-existing server-side `ssl` and verify that the unchanged configuration surfaces the removal/replacement plan, confirming the documented requirement to add the existing settings to configuration to adopt them
- [x] 6.12 Seed a host whose `ssl` object contains non-null `agent_*`/`es_*` fields, attempt a supported-SSL update that would replace the object, and verify the provider returns an actionable diagnostic without sending a mutation or changing the seeded values
- [x] 6.13 Apply and refresh a configuration with `ssl` set but `client_auth` omitted; assert no inconsistent-result error and no subsequent diff. If the Task 1.9 probe observes a default, also assert the value is represented using the selected plan-consistent schema strategy while the initial request omitted it

## 7. Docs and examples

- [x] 7.1 Update `examples/resources/elasticstack_fleet_server_host/resource.tf` to show the `ssl` block
- [x] 7.2 Regenerate `docs/resources/fleet_server_host.md` via the existing `make` docs target
- [x] 7.3 Add a CHANGELOG entry following the repo's existing format
- [x] 7.4 Document in the resource's generated docs that changing SSL settings on an enrolled Fleet Server host can disconnect connected Elastic Agents and may require re-enrollment
- [x] 7.5 Document the required configuration migration for hosts with pre-existing server-side `ssl`: add the existing SSL settings to configuration when upgrading the provider

## 8. Validation and cleanup

- [x] 8.1 Run `make build` and `make check-lint` — fix any issues
- [x] 8.2 Run `make check-openspec` — confirm the change validates
- [x] 8.3 Run targeted acceptance tests against a real Kibana at or above the Task 1.7-confirmed floor (per `dev-docs/high-level/testing.md`)
- [x] 8.4 Run the full unit test suite for `internal/fleet/serverhost`
- [x] 8.5 Self-review with the `requirements-verification` skill against this change's specs
