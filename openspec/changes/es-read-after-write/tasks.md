## 1. Envelope: add read-after-write to writeFromPlan

- [ ] 1.1 In `writeFromPlan`, after a successful concrete callback invocation, call `r.readFunc` with the returned `writtenModel` as prior state
- [ ] 1.2 If `readFunc` reports `found == false`, append an error diagnostic using `r.component` and `r.resourceName` to identify the resource type
- [ ] 1.3 Set state from the model returned by `readFunc`, not from the concrete callback's return value
- [ ] 1.4 Add unit tests for: read-after-write happy path (create), read-after-write happy path (update), not-found-after-create error, not-found-after-update error, readFunc error after create, readFunc error after update
- [ ] 1.5 Update existing write happy-path tests to assert state is set from the `readFunc` result

## 2. Concrete callbacks: remove inline read-after-write

- [ ] 2.1 `internal/elasticsearch/security/role/update.go` — remove `readRole` call and not-found handling; ensure composite ID is still set on the returned model before returning
- [ ] 2.2 `internal/elasticsearch/cluster/script/update.go` — remove `readScriptPayload` call, field-carrying block, and not-found check; ensure composite ID is still set on the returned model before returning
- [ ] 2.3 `internal/elasticsearch/security/rolemapping/update.go` — remove `readRoleMappingResource` call and nil check; return the plan model directly (rolemapping's `readFunc` computes the composite ID itself so no ID setup is needed)
- [ ] 2.4 `internal/elasticsearch/security/systemuser/update.go` — remove `readSystemUser` call and not-found handling; ensure composite ID is still set on the returned model before returning

## 3. Spec and godoc

- [ ] 3.1 Update the `ElasticsearchCreateFunc` and `ElasticsearchUpdateFunc` type godocs to document the narrowed contract: callbacks call the API and return the written model with composite ID set (where readFunc carries it through) and any create-only fields populated; they must not call readFunc
- [ ] 3.2 Run `make check-openspec` to confirm the delta spec passes validation
