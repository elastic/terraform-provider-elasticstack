## 1. Schema

- [ ] 1.1 Add `auth` `SingleNestedAttribute` to `internal/fleet/agentdownloadsource/schema.go`: `username` (optional string), `password` (optional string, `Sensitive: true`), `api_key` (optional string, `Sensitive: true`), `headers` (optional `ListNestedAttribute` of `{key, value}` both required strings, `listvalidator.SizeAtMost(100)`)
- [ ] 1.2 Add `stringvalidator.AlsoRequires` between `auth.username` and `auth.password` (relative-path, each direction), matching `internal/schema/es_connection_shared.go:51,59`
- [ ] 1.3 Add `stringvalidator.ConflictsWith` between `auth.password` and `auth.api_key` (relative-path), matching `internal/schema/es_connection_shared.go:68,78`
- [ ] 1.4 Confirm no validator ties `headers` to the other three fields (headers-only `auth` block must remain valid)

## 2. Models — create/update mapping

- [ ] 2.1 In `internal/fleet/agentdownloadsource/models.go`, add an `authModel` struct (`Username`, `Password`, `APIKey types.String`, `Headers types.List`) and thread it onto `model` as `Auth types.Object`
- [ ] 2.2 Implement `toAPIAuthModel` building `kbapi...JSONRequestBody.Auth` from `authModel` (headers converted to the generated anonymous struct slice)
- [ ] 2.3 Update `toAPICreateModel` to set `Auth` from the planned `auth` block when present, nil when absent
- [ ] 2.4 Update `toAPIUpdateModel` per Decision 3: when planned `auth` is null/absent AND `prior.Auth` was non-null, set `Auth` to an explicit empty struct (`&struct{...}{}`, marshals to `{}`); when planned `auth` is non-null, populate fully; when both planned and prior are null/absent, leave `Auth` nil
- [ ] 2.5 Add the 9.4.0 version requirement in `GetVersionRequirements`, firing only when `auth` is non-null in the plan, using `entitycore.NewAttributeVersionRequirement`/`addMinimum` pattern from `internal/fleet/agentpolicy/version_requirements.go:32-43`; keep `agentDownloadSourceMinVersion` (8.13.0) as the resource baseline

## 3. Read hydration

- [ ] 3.1 Change `readAndHydrateState` signature in `internal/fleet/agentdownloadsource/read.go` to accept a `fallbackAuth types.Object` (or equivalent) parameter alongside the existing preserved space/connection params
- [ ] 3.2 Decode `item.Auth.Username` verbatim (`types.StringPointerValue`)
- [ ] 3.3 Decode `item.Auth.Headers` verbatim into a `types.List`, keeping null (no headers) distinct from an explicit empty list (`headers: []` in the response), matching `fromAPIKafkaModel`'s headers handling (`internal/fleet/output/models_kafka.go:584-598`)
- [ ] 3.4 Apply the `kafka.password` nil-pointer preservation condition independently to `item.Auth.Password` and `item.Auth.ApiKey`: when the API pointer is nil, fall back to the corresponding field on `fallbackAuth`; otherwise use the API value — mirror `internal/fleet/output/models_kafka.go:566-581`
- [ ] 3.5 When `item.Auth` itself is nil in the response, set state's `auth` to null (no credentials/headers configured server-side)
- [ ] 3.6 Update `readAgentDownloadSource` (the plain refresh/Read path) to pass `prior.Auth` as the fallback, mirroring kafka's `configuredPassword` capture
- [ ] 3.7 Update `finalizeWrite` in `internal/fleet/agentdownloadsource/write.go` to pass `plan.Auth` (not prior state) as the fallback into `readAndHydrateState`, so a just-written/rotated credential survives the mandatory post-write read

## 4. Acceptance tests

- [ ] 4.1 Add `testdata/TestAccResourceFleetAgentDownloadSource*/with_auth_basic/main.tf` (username+password) and corresponding test case in `acc_test.go`
- [ ] 4.2 Add `with_auth_api_key` test case
- [ ] 4.3 Add `with_auth_headers_only` test case (no username/password/api_key, headers only)
- [ ] 4.4 Add `auth_update_rotate_password` test case asserting the rotated value survives post-update hydration (exercises the plan-fallback path from 3.7)
- [ ] 4.5 Add `auth_switch_basic_to_api_key` test case
- [ ] 4.6 Add `auth_remove` test case: `auth` block set then removed in a subsequent apply; assert the follow-up read shows no credentials (exercises the `auth: {}` clearing path from 2.4)
- [ ] 4.7 Add `auth_empty_object` test case (`auth = {}` in config)
- [ ] 4.8 Add `auth_headers_empty_list` test case (`auth = { headers = [] }`) asserting state keeps `headers` as an empty list, not null
- [ ] 4.9 Add `auth_headers_max` (exactly 100 header entries, accepted) and `auth_headers_over_max` (101 entries, `ExpectError` matching the API's 400 `array size is [101]...` message or the provider's own size validator diagnostic) test cases
- [ ] 4.10 Add validator `ExpectError` cases: `username`+`password`+`api_key` together, and `username` without `password`
- [ ] 4.11 Add `auth_import_redacted`: create a source with `auth.password`/`auth.api_key` set, import it, and use a scoped `ImportStateVerifyIgnore` (`[]string{"auth.password", "auth.api_key"}`) plus a custom `ImportStateCheck` asserting both are null post-import, following `internal/apm/source_map/acc_test.go:159-167,210` / `internal/elasticsearch/security/user/acc_test.go:293-304`
- [ ] 4.12 Add a negative version-gate case: a configuration setting `auth` against a Kibana below 9.4.0 is rejected, while an otherwise-identical configuration omitting `auth` still succeeds at the resource's 8.13.0 baseline; use the provider's existing version-skip/version-gate test helpers

## 5. Spec sync

- [ ] 5.1 After implementation, sync the delta spec in `openspec/changes/fleet-agent-download-source-auth/specs/fleet-agent-download-source/spec.md` into `openspec/specs/fleet-agent-download-source/spec.md` (openspec-sync-specs skill) and archive the change (openspec-archive-change skill)

## 6. Docs (follow-up, not required for this proposal)

- [ ] 6.1 Add an `auth` example to `examples/resources/elasticstack_fleet_agent_download_source/` and regenerate `docs/resources/fleet_agent_download_source.md`
