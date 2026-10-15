## 1. PUT-clearing semantics (resolved; confirm before merge)

- [x] 1.1 Create a download source with `ssl` fully populated (certificate + CA list + key) — confirmed via live probe against Kibana 9.4.0, recorded in design.md Decision 4
- [x] 1.2 Confirm omitting `ssl` entirely from `PUT` leaves the prior value unchanged — confirmed
- [x] 1.3 Confirm `PUT` sending `ssl: {}` clears it, with a subsequent `GET` returning `"ssl": {}` — confirmed
- [x] 1.4 Confirm omitting a single child (`certificate`, `key`, or `certificate_authorities`) while sending the other two clears only the omitted child — confirmed
- [ ] 1.5 Before considering this change implementation-complete, re-run probe 1.4 specifically for the `key`/`secrets.ssl.key` combination this design uses (the recorded probe's wording describes the all-plain-`ssl` shape; confirm omitting `secrets.ssl.key` from a `PUT` that still includes `ssl.certificate`/`ssl.certificate_authorities` clears only the secret and leaves the other two untouched) — see design.md's residual "Open questions" entry

## 2. Model and mapping

- [ ] 2.1 Create `internal/fleet/agentdownloadsource/models_ssl.go` adapted from `internal/fleet/output/models_ssl.go`: an `agentDownloadSourceSslModel` (`tfsdk` struct for `certificate_authorities`, `certificate`, `key`), and `objectValueToSSL`/`sslToObjectValue` helpers covering `certificate`/`certificate_authorities` against the kbapi plain `Ssl` struct (no `verification_mode`); `key` is handled separately per 2.2, not inside these helpers
- [ ] 2.2 Update `internal/fleet/agentdownloadsource/models.go`: add `Ssl types.Object` to `model`; wire `certificate`/`certificate_authorities` into `toAPICreateModel`/`toAPIUpdateModel`'s plain `Ssl` field via the Task 2.1 helpers; wire `key` into `body.Secrets.Ssl.Key` via the generated union's string-arm constructor (`FromPostFleetAgentDownloadSourcesJSONBodySecretsSslKey1`/`FromPutFleetAgentDownloadSourcesSourceidJSONBodySecretsSslKey1`) whenever `ssl.key` is configured — never set the plain `Ssl.Key` field (design.md Decision 1a); for removal, mirror `authForUpdate`'s existing pattern and send an explicit, non-nil, all-nil-fields `Ssl{}` (and omit `Secrets.Ssl`) when `ssl` was previously set and is now absent from config, rather than a nil pointer (design.md Decision 4)
- [ ] 2.3 Add the attribute-scoped `ssl` version requirement (9.1.0) to `GetVersionRequirements`, following the guarded-closure pattern in `internal/fleet/agentpolicy/version_requirements.go`, appended only when `ssl` is known/configured

## 3. Schema

- [ ] 3.1 Add the `ssl` `schema.SingleNestedAttribute` (Optional) to `internal/fleet/agentdownloadsource/schema.go` with `certificate_authorities` (`schema.ListAttribute`, Optional, string elements), `certificate` (`schema.StringAttribute`, Optional), `key` (`schema.StringAttribute`, Optional, Sensitive)
- [ ] 3.2 Add validators: `certificate_authorities` only — `listvalidator.SizeAtMost(10)` and `listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1))`. Do **not** add `stringvalidator.LengthAtLeast(1)` on `certificate`/`key`, and do **not** add `listvalidator.SizeAtLeast(1)` on `certificate_authorities` — empty/omitted values are accepted, matching `auth` (design.md Decision 2, revised)
- [ ] 3.3 Do **not** add an object-level "at least one of" validator on `ssl` — `ssl = {}` is accepted (design.md Decision 2, revised; supersedes the original plan)
- [ ] 3.4 Do **not** add a `RequiresReplace` plan modifier for `ssl` or any of its children — every removal transition is an in-place update (design.md Decision 4, resolved)

## 4. Read path

- [ ] 4.1 Update `internal/fleet/agentdownloadsource/read.go`'s `readAndHydrateState` to populate `state.Ssl.certificate`/`certificate_authorities` from `item.Ssl` via `sslToObjectValue`, and `state.Ssl.key` by carrying the prior Terraform state value forward via the existing `preserveRedactedSecret` helper (never reading `item.Ssl.Key` or `item.Secrets.Ssl.Key` directly) — mirrors the existing `auth.password`/`auth.api_key` treatment in the same file

## 5. Unit tests

- [ ] 5.1 Add `internal/fleet/agentdownloadsource/models_ssl_test.go` covering: full round-trip for `certificate`/`certificate_authorities`, CA-only, certificate-only, API-returns-all-nil normalizes to `types.ObjectNull`/`types.ListNull`, and an explicitly-empty `certificate = ""`/`certificate_authorities = []` configuration round-trips as configured (not rejected, not normalized away)
- [ ] 5.2 Add a model/read unit test asserting `key` is carried forward from prior state across a read regardless of what `item.Ssl.Key`/`item.Secrets.Ssl.Key` contain (including both nil and non-nil API values), confirming the provider never persists an API-read value into `key`
- [ ] 5.3 Add a schema unit test asserting `ssl` exposes only `certificate_authorities`, `certificate`, and sensitive `key` (with no `verification_mode`), has `certificate_authorities`'s `SizeAtMost(10)`/element-nonempty validators, and has **no** object-level or per-field nonemptiness validators

## 6. Acceptance tests

- [ ] 6.1 Add `TestAccResourceFleetAgentDownloadSource_SSL` to `internal/fleet/agentdownloadsource/acc_test.go`, version-skipped below 9.1.0 via `versionutils.SkipIfUnsupported`, modeled on `internal/fleet/proxy/acc_test.go`'s `TestAccResourceFleetProxy_TLS`
- [ ] 6.2 Step `with_ssl`: create with `certificate`, `key`, and multiple `certificate_authorities` set; assert `certificate`/`certificate_authorities` round-trip via `TestCheckResourceAttr`; assert `key` stays equal to the configured value in state across a refresh even though the API never echoes it back
- [ ] 6.3 Step `with_ssl_updated`: change all three fields to different values; assert round-trip and no-diff-after-apply, including a changed `key` being detected as a plan diff by ordinary config comparison (no custom change-detection needed)
- [ ] 6.4 Step `clear_ssl`: remove `ssl` entirely; assert an in-place update (`plancheck.ExpectResourceAction(..., ResourceActionUpdate)`, not replace) and that a subsequent read shows `ssl` as null; independently GET the source and assert `secrets.ssl.key` and plain `ssl.key` are absent and certificate/CA fields are cleared
- [ ] 6.5 Per-child removal: three separate transitions from a fully-populated `ssl` to one with a single child dropped (drop `key` → certificate+CA only; drop `certificate_authorities` → certificate+key only; drop `certificate` → key+CA only); assert each is an in-place update and the dropped child reads back as null/absent while the other two persist; independently GET the source to verify the dropped plain field or `secrets.ssl.key` reference is absent, retained certificate/CA values are unchanged, any retained key still has a secret reference, and plain `ssl.key` remains absent
- [ ] 6.6 Explicit `ssl = {}` step: assert this is **accepted** (no validator error) and produces an in-place update that clears any previously-set SSL values while preserving the configured empty object in Terraform state; independently GET the source and assert `secrets.ssl.key` and plain `ssl.key` are absent and certificate/CA fields are cleared
- [ ] 6.7 Explicitly-empty-value cases, each asserted as **accepted** (not rejected): `certificate = ""`, `key = ""`, `certificate_authorities = []`
- [ ] 6.8 CA-list boundary: exactly 10 entries accepted; 11 entries rejected; a CA list containing one empty-string element among non-empty entries rejected
- [ ] 6.9 `key` is marked sensitive in plan output
- [ ] 6.10 Below-minimum-version case: assert the attribute-scoped diagnostic fires when `ssl` is set against a pre-9.1 stack, mirroring `internal/entitycore/version_requirements_test.go`
- [ ] 6.11 Import: confirm `certificate`/`certificate_authorities` round-trip through the existing `ImportStateIdFunc` pattern, and `key` is null/unset immediately post-import (no private-state or server-side source for it), matching the existing `auth.password`/`auth.api_key` import treatment

## 7. Docs and examples

- [ ] 7.1 Update `examples/resources/elasticstack_fleet_agent_download_source/resource.tf` to show the `ssl` block, including a comment noting `key` is stored by Fleet as a managed secret (sent via `secrets.ssl.key` internally) and is never read back from the API
- [ ] 7.2 Regenerate `docs/resources/fleet_agent_download_source.md` via the existing `make` docs target
- [ ] 7.3 Add a CHANGELOG entry following the repo's existing format
- [ ] 7.4 Update the `ssl` shape in the canonical schema sketch at `openspec/specs/fleet-agent-download-source/spec.md` when syncing this delta; the delta spec's "Auth and secrets not in v1 schema" requirement is renamed/narrowed to allow `ssl` while continuing to exclude the `{id}`-reference form of `secrets.*` (see the change's own delta spec for the exact wording)

## 8. Validation and cleanup

- [ ] 8.1 Run `make build` and `make check-lint` — fix any issues
- [ ] 8.2 Run `make check-openspec` — confirm the change validates
- [ ] 8.3 Run targeted acceptance tests against a real Kibana ≥ 9.1 (per `dev-docs/high-level/testing.md`)
- [ ] 8.4 Run the full unit test suite for `internal/fleet/agentdownloadsource`
- [ ] 8.5 Self-review with the `requirements-verification` skill against this change's specs
