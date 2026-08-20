## 1. Schema

- [ ] 1.1 Add `attrForceMergeOnClone = "force_merge_on_clone"` to `internal/elasticsearch/index/ilm/constants.go`
- [ ] 1.2 Add `force_merge_on_clone` (`schema.BoolAttribute`, Optional+Computed, `Default: booldefault.StaticBool(true)`) to `blockSearchableSnapshot()` in `internal/elasticsearch/index/ilm/schema_actions.go`
- [ ] 1.3 Add the same attribute to `blockSearchableSnapshotInFrozenPhase()` in the same file
- [ ] 1.4 Add `attrForceMergeOnClone: types.BoolType` to `searchableSnapshotObjectType()` in `internal/elasticsearch/index/ilm/attr_types.go`

## 2. Version gating and expand

- [ ] 2.1 Add a package-level `SearchableSnapshotForceMergeOnCloneMinSupportedVersion = version.Must(version.NewVersion("9.2.0"))` (or equivalent local var) near the other version constants used by `ilmActionSettingOptions`
- [ ] 2.2 Register `attrForceMergeOnClone: {def: true, minVersion: <9.2.0 var>}` in `ilmActionSettingOptions` in `internal/elasticsearch/index/ilm/expand.go`
- [ ] 2.3 Add `attrForceMergeOnClone` to the `ilmActionSearchableSnapshot` case's `expandAction(...)` call in `expandPhase`
- [ ] 2.4 Confirm flatten/read-side mapping for `searchable_snapshot` already passes the new key through generically (no action-specific flatten logic expected, but verify against `internal/models/index.go`)

## 3. `PutIlm` transport change

- [ ] 3.1 In `internal/clients/elasticsearch/ilm.go`, change `PutIlm` to call `typedClient.Ilm.PutLifecycle(policy.Name).Raw(bytes.NewReader(policyBytes)).Do(ctx)` instead of unmarshaling into `putlifecycle.Request` and calling `.Request(&req)`
- [ ] 3.2 Remove the now-unused `putlifecycle.Request` unmarshal step; drop the `putlifecycle` import if nothing else in the file uses it
- [ ] 3.3 Double check `bytes` is imported (already used by `ClearILMPolicyFromIndices` in the same file)
- [ ] 3.4 Confirm error handling around the `.Raw()` call still surfaces Elasticsearch's error body the same way `.Request()` did (compare against `ClearILMPolicyFromIndices`'s existing `.Raw()` error handling for consistency)

## 4. Tests

- [ ] 4.1 Add a new `SkipFunc: versionutils.CheckIfVersionIsUnsupported(<9.2.0 var>)`-gated `resource.TestStep` to `TestAccResourceILMSearchableSnapshotPhases` in `internal/elasticsearch/index/ilm/acc_test.go` (do not add a new test function), exercising `force_merge_on_clone = false` on at least one phase's `searchable_snapshot` block
- [ ] 4.2 Add any `testdata` config fixtures the new step needs (new `ConfigDirectory`, or extend the existing `create`/`update` fixtures if the step reuses them)
- [ ] 4.3 Add/extend a unit test asserting that a non-default `force_merge_on_clone` value against an unsupported (pre-9.2) server version produces the "Unsupported ILM setting" diagnostic, following the pattern used for `attrAllowWriteAfterShrink`
- [ ] 4.4 Run `make build`, `go vet ./...`, and the ILM package's unit tests locally (acceptance tests require a live 9.2+ Elasticsearch cluster and are out of scope for this proposal's authoring step)

## 5. Docs and spec sync

- [ ] 5.1 Update `openspec/specs/elasticsearch-index-lifecycle/spec.md`: extend the version-gated settings requirement (REQ-022–REQ-025) or add a new requirement documenting `force_merge_on_clone` and its `9.2.0` gate, and update the `searchable_snapshot` schema block shown in the "Action block shapes" section
- [ ] 5.2 Regenerate/update provider documentation for `elasticstack_elasticsearch_index_lifecycle` to describe the new attribute
- [ ] 5.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate ilm-force-merge-on-clone --type change` and fix any reported issues
- [ ] 5.4 After implementation lands and this change is archived, run `make check-openspec` to confirm the main spec absorbed the delta cleanly

## 6. Follow-up (not blocking this change)

- [ ] 6.1 File an issue against `elastic/elasticsearch-specification` requesting `force_merge_on_clone` support in the generated `SearchableSnapshotAction` type, so `PutIlm` can eventually revert from `.Raw()` back to the typed `.Request()` path
