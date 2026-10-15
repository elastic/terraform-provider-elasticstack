## 1. `writeonlyhash` helper fix

- [ ] 1.1 Fix `internal/utils/writeonlyhash/hasher.go`'s `passwordBytes` (`hasher.go:97-102`): when `len(salt)+1+len(value)` exceeds 72 bytes, compute SHA-256 over the salted bytes and base64-encode the digest before returning; otherwise return the salted bytes unchanged
- [ ] 1.2 Add `internal/utils/writeonlyhash/hasher_test.go` covering: a value ≤72 bytes hashes identically before/after the fix (regression guard for `elasticsearch_connector`'s existing stored hashes); a PEM-sized (>72 byte) value round-trips through `Compute`/`Matches` without error; two different >72-byte values that share a common ~72-byte prefix after the salt do NOT falsely match

## 2. Shared envelope: thread private state into Kibana writes

- [ ] 2.1 Add `Private PrivateStateStorage` to `entitycore.KibanaWriteRequest[T]` in `internal/entitycore/kibana_resource_envelope.go`, mirroring `entitycore.WriteRequest[T].Private` in `internal/entitycore/resource_envelope.go:87-95`
- [ ] 2.2 Stop discarding the real `PrivateStateStorage` value via `_` in the write-adapter wiring inside `kibana_resource_envelope.go`; pass the framework response's `Private` through into the constructed `KibanaWriteRequest[T]`
- [ ] 2.3 Confirm no existing `KibanaWriteFunc[T]` callback breaks: the field is additive, so existing resources that don't reference `req.Private` are unaffected; run the package's existing unit tests for `internal/entitycore`

## 3. Schema: write-only attributes and consolidated auth validator

- [ ] 3.1 Add `key_wo` (`WriteOnly`, `Sensitive`, string) to the `ssl` nested attribute in `internal/fleet/agentdownloadsource/schema.go` (alongside the existing `key` at `schema.go:107-114`), and make `key` and `key_wo` mutually exclusive because they map to the same API field
- [ ] 3.2 Add `password_wo` and `api_key_wo` (`WriteOnly`, `Sensitive`, string) to the `auth` nested attribute in `schema.go` (alongside the existing `password`/`api_key` at `schema.go:128-143`)
- [ ] 3.3 Remove the per-attribute `stringvalidator.AlsoRequires`/`ConflictsWith` validators currently on `auth.username`/`password`/`api_key` (`schema.go:121-143`) and replace with one object-level `validator.Object` on the `auth` `SingleNestedAttribute` implementing: `username` and either `password` or `password_wo` must both be configured or both be absent; `password`/`password_wo` mutually exclusive; `api_key`/`api_key_wo` mutually exclusive; password family conflicts with api_key family — add this as a new validator type in `internal/fleet/agentdownloadsource/validators.go` (or extend the existing validator file if one is more appropriate), following the shape of `sslAtLeastOneSubattrValidator` for the object-level-validator plumbing
- [ ] 3.4 Confirm `TestResource_schemaOmitsSecretsReferences` (`entitycore_contract_test.go`) still passes unmodified: the new attributes are `_wo` siblings, not a literal `secrets` schema key, so the assertion's premise holds; add a new contract test asserting the `_wo` attributes are present and marked `WriteOnly`+`Sensitive`
- [ ] 3.5 Extend `sslAtLeastOneSubattrValidator` and its apply-time `AllNil` guard in `models_ssl.go` to recognize `key_wo`, including unknown-value deferral; test `ssl = { key_wo = "..." }` without any other SSL child, as well as rejection when both `key` and `key_wo` are set

## 4. Models: secret-reference request types and hash-aware write branching

- [ ] 4.1 In `internal/fleet/agentdownloadsource/models_ssl.go`, add model plumbing for `key_wo`, and route both the existing plain `key` and new `key_wo` through `secrets.ssl.key` (not plain `ssl.key`); send the plain key as a literal and use the literal or `{Id: string}` arm for `key_wo` per Decision 3's hash/id branching
- [ ] 4.2 In `internal/fleet/agentdownloadsource/models.go`, add `authModel.PasswordWO`/`APIKeyWO` fields and extend `authRequest`/`authForUpdate` (`models.go:146-157,198-206`) so configured `password_wo`/`api_key_wo` map to `kbapi...SecretsAuthPassword{0,1}`/`...SecretsAuthApiKey{0,1}` using the same literal-vs-`{id}` branching as 4.1, while plain `password`/`api_key` continue to map to the existing plain wire fields unchanged
- [ ] 4.3 Add a `secretWriteBranch` (or similarly named) helper shared by 4.1/4.2 that, given a write-only value, its `writeonlyhash` hash, the stored prior hash (from private state), and the pre-PUT GET's secret id for that field, returns one of: literal (first apply / hash changed / no GET id available), or `{id: <existing>}` (hash unchanged and GET id available) — implementing Decision 3 once, not duplicated per field

## 5. Read: GET-before-PUT secret id lookup

- [ ] 5.1 In `internal/fleet/agentdownloadsource/read.go`, add a helper that performs the pre-PUT GET by `source_id` and extracts whatever secret id Fleet currently has for `secrets.ssl.key`/`secrets.auth.password`/`secrets.auth.api_key` (decoding the `...0`-shaped `{Id string}` response arm from `GetFleetAgentDownloadSources200JSONResponseBody_Items_Secrets_*` types), returning `nil`/absent per field when Fleet reports no id (secret storage unavailable)
- [ ] 5.2 Wire this helper into the Update path in `internal/fleet/agentdownloadsource/write.go` (`updateAgentDownloadSource`, called from `writeAgentDownloadSource` at `write.go:38-41`) so the GET happens immediately before building the PUT body via 4.3's branching helper
- [ ] 5.3 Update read hydration to distinguish configured plain credentials from write-only credentials: never copy an echoed `key_wo` value from API `ssl.key` into stateful `ssl.key`; preserve a configured plain `ssl.key` when the API redacts it; retain an `ssl` object when `key_wo` is its only configured child
- [ ] 5.4 Add post-write and refresh tests for the three read safeguards in 5.3, including the storage-unavailable response that echoes a `key_wo` literal as `ssl.key`

## 6. Write: persist and clear hashes via private state

- [ ] 6.1 Update `writeAgentDownloadSource` (`write.go:31-42`) to accept and forward `req.Private` (now available per task 2) into the create/update model-building calls
- [ ] 6.2 After a successful create or update, for each configured write-only attribute (`ssl.key_wo`, `auth.password_wo`, `auth.api_key_wo`), compute its `writeonlyhash` hash and store it via `req.Private.SetKey` under `secret_hash:ssl.key_wo` / `secret_hash:auth.password_wo` / `secret_hash:auth.api_key_wo` (per `Hasher.PrivateStateKey`), JSON-wrapping the raw hash bytes first (they contain `$`), following `internal/elasticsearch/connector/resource/secret.go`'s `encodeSecretHashForPrivateState` pattern — add the encode/decode helpers in a new `internal/fleet/agentdownloadsource/secret_hash.go`
- [ ] 6.3 Construct the resource's `writeonlyhash.Hasher` via `writeonlyhash.New("elasticstack_fleet_agent_download_source")` as a package-level var in `secret_hash.go`, mirroring `internal/elasticsearch/connector/resource/secret.go`'s `secretHasher`
- [ ] 6.4 Read write-only literals from `req.Config` when building create/update writes and computing hashes; do not rely on `req.Plan` for those fields because write-only values are null there. Continue using planned values for ordinary attributes
- [ ] 6.5 When configuration removes a write-only value with a stored hash, build the explicit clearing update and retain the hash until the remote update succeeds; only then clear the corresponding private-state key. A subsequent re-add must be treated as a first write and send a literal

## 7. ModifyPlan and Delete

- [ ] 7.1 Add `internal/fleet/agentdownloadsource/modify_plan.go` with `func (r *Resource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse)`, mirroring `internal/elasticsearch/connector/resource/modify_plan.go`: decode config and state for the three write-only attributes, compare configured values' `writeonlyhash` hashes against stored `secret_hash:*` private-state entries via `req.Private.GetKey` + the decode helper from 6.2, and force an update if a hash differs or a stored hash exists while the configured value is null. The null transition must plan the clearing update without deleting the hash before the write succeeds
- [ ] 7.2 In `internal/fleet/agentdownloadsource/resource.go`, add `_ resource.ResourceWithModifyPlan = (*Resource)(nil)` to the interface assertions (alongside the existing `resource.Resource`/`resource.ResourceWithConfigure`/`resource.ResourceWithImportState` assertions)
- [ ] 7.3 Add a `Delete` override on `*Resource` in `resource.go` that calls the embedded `KibanaResource[model].Delete`, then clears the three `secret_hash:*` private-state keys via `resp.Private.SetKey(ctx, key, nil)`, mirroring `internal/elasticsearch/connector/resource/resource.go:54-61`

## 8. Acceptance tests

- [ ] 8.1 Add `ssl.key` and `ssl.key_wo` create/rotate/no-op cases to `internal/fleet/agentdownloadsource/acc_test.go`: verify both attributes write to `secrets.ssl.key` (the plain key as a literal, and `key_wo` as literal or `{id}` per Decision 3); verify unchanged `key_wo` preserves the id, changed `key_wo` sends a new literal, and configuring both attributes is rejected
- [ ] 8.2 Repeat 8.1 for `auth.password_wo` and `auth.api_key_wo` independently
- [ ] 8.3 Add a mixed-attribute validator rejection case: `auth.username` with neither `password` nor `password_wo` set — error; `password`+`password_wo` both set — error; `api_key`+`api_key_wo` both set — error; `password_wo`+`api_key` (cross-family) — error
- [ ] 8.4 Add a secret-storage-unavailable case (against a Fleet Server version below the runtime storage-gating floor, or a test double simulating a GET response with no secret id): assert the resource resends the literal value rather than attempting an `{id}` reference, does not error, and never copies an echoed write-only literal into the plain sibling during post-write read or refresh
- [ ] 8.5 Add a PEM-sized `ssl.key_wo` value acceptance case exercising the task-1 bcrypt fix end-to-end (create, no-op refresh, then a value-only change triggers an update)
- [ ] 8.6 Add a delete case asserting the resource's three `secret_hash:*` private-state keys are cleared (directly, if the test harness can inspect private state, or indirectly via re-create-with-same-config-not-erroring)
- [ ] 8.7 Add an import case: `ssl.key_wo`/`auth.password_wo`/`auth.api_key_wo` are null after import (no API representation to import from), mirroring the existing `AuthBasicImport` pattern
- [ ] 8.8 Add auth validator coverage for `password` without `username` and `password_wo` without `username`, and for `username` without either password form
- [ ] 8.9 Add a `key_wo`-only SSL create/refresh case asserting the parent `ssl` object remains present while the write-only value is not persisted in `ssl.key`
- [ ] 8.10 Add a write-only removal/re-add lifecycle case: removing an unchanged `key_wo` with no ordinary attribute changes forces a PUT that clears the remote secret and clears the private hash only after success; re-adding it sends a literal and stores a new private-state hash

## 9. Spec sync

- [ ] 9.1 After implementation, sync the delta spec in `openspec/changes/fleet-agent-download-source-secrets/specs/fleet-agent-download-source/spec.md` into `openspec/specs/fleet-agent-download-source/spec.md` (openspec-sync-specs skill) and archive the change (openspec-archive-change skill)

## 10. Docs (follow-up, not required for this proposal)

- [ ] 10.1 Add `ssl.key_wo`/`auth.password_wo`/`auth.api_key_wo` examples to `examples/resources/elasticstack_fleet_agent_download_source/`, regenerate `docs/resources/fleet_agent_download_source.md`, and add the Terraform CLI 1.11 minimum-version note reused from the other `_wo` resources
