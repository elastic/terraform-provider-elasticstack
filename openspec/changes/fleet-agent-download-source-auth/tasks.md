## 1. Credential modes

- [ ] 1.1 Add the `auth` `SingleNestedAttribute` to `internal/fleet/agentdownloadsource/schema.go` with optional `username` (string), `password` and `api_key` (sensitive strings), and optional `headers` (list of `{key, value}` with required strings)
- [ ] 1.2 Add relative-path `stringvalidator.AlsoRequires` validators in both directions between `auth.username` and `auth.password`, plus `stringvalidator.ConflictsWith` between `auth.password` and `auth.api_key`, following `internal/schema/es_connection_shared.go:51,59,68,78`; do not couple `headers` to credentials
- [ ] 1.3 Add `authModel` in `internal/fleet/agentdownloadsource/models.go`, populate the generated request `Auth` from it, and map configured credential fields on create and update
- [ ] 1.4 Extend `readAndHydrateState` in `internal/fleet/agentdownloadsource/read.go` to decode `username` and preserve redacted `password`/`api_key` from a caller-supplied fallback independently, mirroring `internal/fleet/output/models_kafka.go:566-581`
- [ ] 1.5 Add acceptance cases for basic auth and API-key auth, plus validator errors for username without password, password without username, and password with api_key

## 2. Headers and empty auth

- [ ] 2.1 Add the `listvalidator.SizeAtMost(100)` validator to `auth.headers` and confirm no validator prevents headers-only or empty `auth = {}` blocks
- [ ] 2.2 Map headers to the generated request and response types while preserving null versus empty: if configured as an empty list, construct a non-nil empty generated slice so the request sends `"headers":[]` rather than omitting `headers`; decode an API `headers: []` as an empty `types.List`, and absent headers as null, following `internal/fleet/output/models_kafka.go:584-598`
- [ ] 2.3 Add acceptance cases for headers-only auth, `auth = {}`, and `auth = { headers = [] }` (assert the empty list remains non-null in state)
- [ ] 2.4 Add `auth_headers_max` (100 entries accepted) and `auth_headers_over_max` (101 entries rejected by the schema validator before any API call) acceptance cases

## 3. Update removal and secret preservation

- [ ] 3.1 Update `toAPIUpdateModel`: send an explicit empty `Auth` object when planned `auth` is absent but `prior.Auth` was set; populate the full object when planned `auth` is set; leave `Auth` nil only when absent in both plan and prior state
- [ ] 3.2 Pass `prior.Auth` as the fallback from the plain refresh/Read path and `plan.Auth` from `finalizeWrite` into `readAndHydrateState`, so refresh preserves known secrets and post-write reads preserve newly set or rotated secrets
- [ ] 3.3 Add acceptance cases for password rotation, switching from basic auth to API key, and removing previously configured auth (assert the removal update sends `{}` and clears server-side auth)

## 4. Import redaction

- [ ] 4.1 Ensure the import read path has no prior-auth fallback: redacted `password` and `api_key` remain null while `username` and `headers` hydrate from the API response
- [ ] 4.2 Add a basic-auth import case that verifies `auth.password` is null after import, using an `ImportStateVerifyIgnore` scoped to `auth.password` and an `ImportStateCheck` for null
- [ ] 4.3 Add a separate API-key import case that verifies `auth.api_key` is null after import, using an `ImportStateVerifyIgnore` scoped to `auth.api_key` and an `ImportStateCheck` for null; follow `internal/apm/source_map/acc_test.go:159-167,210` / `internal/elasticsearch/security/user/acc_test.go:293-304`

## 5. Version gating

- [ ] 5.1 Add the 9.4.0 version requirement in `GetVersionRequirements`, firing only when `auth` is non-null in the plan, using `entitycore.NewAttributeVersionRequirement`/`addMinimum` from `internal/fleet/agentpolicy/version_requirements.go:32-43`; keep the 8.13.0 resource baseline unchanged
- [ ] 5.2 Add a negative version-gate case: auth is rejected below 9.4.0, while an otherwise-identical configuration without auth succeeds at the resource's 8.13.0 baseline, using existing version-skip/version-gate test helpers

## 6. Spec sync

- [ ] 6.1 After implementation, sync the delta spec in `openspec/changes/fleet-agent-download-source-auth/specs/fleet-agent-download-source/spec.md` into `openspec/specs/fleet-agent-download-source/spec.md` (openspec-sync-specs skill) and archive the change (openspec-archive-change skill)

## 6. Docs (follow-up, not required for this proposal)

- [ ] 6.1 Add an `auth` example to `examples/resources/elasticstack_fleet_agent_download_source/` and regenerate `docs/resources/fleet_agent_download_source.md`
