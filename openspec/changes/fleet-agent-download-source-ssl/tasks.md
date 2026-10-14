## 1. Pre-implementation probe

- [ ] 1.1 Against a live Kibana ≥ 9.1, create an `elasticstack_fleet_agent_download_source`-equivalent download source with `ssl` fully populated (certificate + CA list + key) via the generated client
- [ ] 1.2 PUT an update that omits `ssl` entirely; independently GET the source afterward; record whether `ssl` persisted or was cleared
- [ ] 1.3 Against a fresh copy of the same starting state, PUT an update sending `ssl = {}`; GET and record the result
- [ ] 1.4 Against three independent fresh copies of the same fully-populated starting state, PUT three separate updates that each omit exactly one child (`certificate` only, `key` only, `certificate_authorities` only); GET each afterward and record whether the omitted child cleared while the retained two persisted
- [ ] 1.5 Record the probe results in design.md "Open questions" / "Decision 4", and decide, per the pre-specified branches, whether `toAPIUpdateModel` needs a `RequiresReplace` plan-modifier fallback (whole-object, per-child, or neither)

## 2. Model and mapping

- [ ] 2.1 Create `internal/fleet/agentdownloadsource/models_ssl.go` adapted from `internal/fleet/output/models_ssl.go`: an `agentDownloadSourceSslModel` (`tfsdk` struct for `certificate_authorities`, `certificate`, `key`), `objectValueToSSL`, and `sslToObjectValue` against the kbapi anonymous `Ssl` struct (no `verification_mode`)
- [ ] 2.2 Update `internal/fleet/agentdownloadsource/models.go`: add `Ssl types.Object` to `model`; wire `objectValueToSSL`/`objectValueToSSLUpdate` into `toAPICreateModel` and `toAPIUpdateModel`; apply the Task 1 probe decision (nil pointer vs. explicit `{}` vs. relying on a plan-modifier fallback) to `toAPIUpdateModel`'s removal path
- [ ] 2.3 Add the attribute-scoped `ssl` version requirement (9.1.0) to `GetVersionRequirements`, following the guarded-closure pattern in `internal/fleet/agentpolicy/version_requirements.go`, appended only when `ssl` is known/configured

## 3. Schema

- [ ] 3.1 Add the `ssl` `schema.SingleNestedAttribute` (Optional) to `internal/fleet/agentdownloadsource/schema.go` with `certificate_authorities` (`schema.ListAttribute`, Optional, string elements), `certificate` (`schema.StringAttribute`, Optional), `key` (`schema.StringAttribute`, Optional, Sensitive)
- [ ] 3.2 Add validators: `certificate`/`key` `stringvalidator.LengthAtLeast(1)`; `certificate_authorities` `listvalidator.SizeAtLeast(1)`, `listvalidator.SizeAtMost(10)`, and `listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))`
- [ ] 3.3 Add an object-level validator on `ssl` requiring at least one of `certificate`, `certificate_authorities`, `key` to be set (reject `ssl = {}`)
- [ ] 3.4 If Task 1's probe requires it, add a `planmodifier.Object` (and/or per-child plan modifiers) implementing `RequiresReplace` for the specific removal transition(s) the API does not honor in place

## 4. Read path

- [ ] 4.1 Update `internal/fleet/agentdownloadsource/read.go`'s `readAndHydrateState` to populate `state.Ssl` from `item.Ssl` via `sslToObjectValue`

## 5. Unit tests

- [ ] 5.1 Add `internal/fleet/agentdownloadsource/models_ssl_test.go` covering: full round-trip (all three fields set), CA-only, certificate+key only, API-returns-all-nil normalizes to `types.ObjectNull`, empty CA list normalizes to `types.ListNull`

## 6. Acceptance tests

- [ ] 6.1 Add `TestAccResourceFleetAgentDownloadSource_SSL` to `internal/fleet/agentdownloadsource/acc_test.go`, version-skipped below 9.1.0 via `versionutils.SkipIfUnsupported`, modeled on `internal/fleet/proxy/acc_test.go`'s `TestAccResourceFleetProxy_TLS`
- [ ] 6.2 Step `with_ssl`: create with `certificate`, `key`, and multiple `certificate_authorities` set; assert round-trip via `TestCheckResourceAttr`
- [ ] 6.3 Step `with_ssl_updated`: change all three fields to different values; assert round-trip and no-diff-after-apply
- [ ] 6.4 Step `clear_ssl`: remove `ssl` entirely; assert the outcome matches the Task 1 probe decision — either `TestCheckNoResourceAttr(..., "ssl")` after an in-place update, or that the step triggers resource replacement
- [ ] 6.5 Per-child removal: three separate transitions from a fully-populated `ssl` to one with a single child dropped (drop `key` → certificate+CA only; drop `certificate_authorities` → certificate+key only; drop `certificate` → key+CA only); assert in-place update or `RequiresReplace` per the Task 1 probe decision for each child
- [ ] 6.6 Explicit `ssl = {}` step: assert `terraform plan`/`apply` fails with the object-level validator error
- [ ] 6.7 Validator-rejection cases, each asserted separately: `certificate = ""`, `key = ""`, `certificate_authorities = []`, and a nonempty CA list containing one empty-string element
- [ ] 6.8 CA-list boundary: exactly 10 entries accepted; 11 entries rejected
- [ ] 6.9 `key` is marked sensitive in plan output
- [ ] 6.10 Below-minimum-version case: assert the attribute-scoped diagnostic fires when `ssl` is set against a pre-9.1 stack, mirroring `internal/entitycore/version_requirements_test.go`
- [ ] 6.11 Import: confirm `ssl` round-trips through the existing `ImportStateIdFunc` pattern

## 7. Docs and examples

- [ ] 7.1 Update `examples/resources/elasticstack_fleet_agent_download_source/resource.tf` to show the `ssl` block
- [ ] 7.2 Regenerate `docs/resources/fleet_agent_download_source.md` via the existing `make` docs target
- [ ] 7.3 Add a CHANGELOG entry following the repo's existing format

## 8. Validation and cleanup

- [ ] 8.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 8.2 Run `make check-openspec` — confirm the change validates
- [ ] 8.3 Run targeted acceptance tests against a real Kibana ≥ 9.1 (per `dev-docs/high-level/testing.md`)
- [ ] 8.4 Run the full unit test suite for `internal/fleet/agentdownloadsource`
- [ ] 8.5 Self-review with the `requirements-verification` skill against this change's specs
