## 1. Schema

- [ ] 1.1 Add `attrForceMergeOnClone = "force_merge_on_clone"` to `internal/elasticsearch/index/ilm/constants.go`
- [ ] 1.2 Add `force_merge_on_clone` (`schema.BoolAttribute`, Optional+Computed, `Default: booldefault.StaticBool(true)`) to `blockSearchableSnapshot()` in `internal/elasticsearch/index/ilm/schema_actions.go`
- [ ] 1.3 Add the same attribute to `blockSearchableSnapshotInFrozenPhase()` in the same file
- [ ] 1.4 Add `attrForceMergeOnClone: types.BoolType` to `searchableSnapshotObjectType()` in `internal/elasticsearch/index/ilm/attr_types.go`

## 1a. `force_merge_index = false` interaction validator

Elasticsearch's `SearchableSnapshotAction` constructor rejects any non-null `force_merge_on_clone` when `force_merge_index = false`. This repo's `TestAccResourceILMSearchableSnapshotPhases` acceptance fixtures already configure `force_merge_index = false`, so this must be handled before the attribute can ship, not treated as an edge case.

- [ ] 1a.1 In `internal/utils/validators/conditional.go`, extend the dependent-value evaluation (`evaluateStaticPath`/`evaluatePathExpression`, currently hardwired to read the dependent as `types.String`) to support a `types.Bool` dependent compared against a `bool`, alongside the existing string case — do not replace the string path, since other resources depend on it. Alternatively, add a small standalone `validator.Bool`/`validator.Object` constructor built on the same `Condition` plumbing, keyed on the sibling's boolean value (shaped like the existing presence-only `ForbiddenIfDrilldownVariantSiblingNestedPresent` at `conditional.go:662`, but value-conditional instead of presence-conditional)
- [ ] 1a.2 Attach the new bool-conditional validator to `force_merge_on_clone` in both `blockSearchableSnapshot()` and `blockSearchableSnapshotInFrozenPhase()` (`schema_actions.go`), forbidding a non-null value when the sibling `force_merge_index` is `false`. Do not reach for `objectvalidator.ConflictsWith`/`AlsoRequires` here — both are presence-only (any non-null sibling value triggers them) and cannot express "conflicts only when the sibling equals `false`"
- [ ] 1a.3 Add a unit test for the new validator/constructor scoped to this `force_merge_on_clone`/`force_merge_index` usage (existing string-dependent call sites of `conditional.go` are unaffected and don't need re-testing)

## 2. Version gating and expand

- [ ] 2.1 Add a package-level `SearchableSnapshotForceMergeOnCloneMinSupportedVersion = version.Must(version.NewVersion("9.2.1"))` (or equivalent local var) near the other version constants used by `ilmActionSettingOptions` — **9.2.1, not 9.2.0**: the field lands in [elastic/elasticsearch#137375](https://github.com/elastic/elasticsearch/pull/137375), labeled `v9.2.1`/`v9.3.0`; `9.2.0` servers do not parse it and will reject the payload
- [ ] 2.2 Register `attrForceMergeOnClone: {def: true, minVersion: <9.2.1 var>}` in `ilmActionSettingOptions` in `internal/elasticsearch/index/ilm/expand.go`
- [ ] 2.3 Add `attrForceMergeOnClone` to the `ilmActionSearchableSnapshot` case's `expandAction(...)` call in `expandPhase`, and omit it from the constructed action map whenever `force_merge_index` resolves to `false` (regardless of `force_merge_on_clone`'s own configured/computed value) — Elasticsearch's `SearchableSnapshotAction` constructor throws `[force_merge_on_clone] is not allowed when [force_merge_index] is [false]`, so the Computed `true` default must never be sent in that combination
- [ ] 2.4 Add a `case ilmActionSearchableSnapshot` in `internal/elasticsearch/index/ilm/flatten.go` (currently falls through to the generic `default:` case at `flatten.go:92-94`) that backfills `force_merge_on_clone = true` when the key is absent from the read action map and `force_merge_index` is `true`, mirroring the existing `ilmActionShrink`/`allow_write_after_shrink` fallback at `flatten.go:85-90`. Elasticsearch only echoes `force_merge_on_clone` back when it was explicitly set (it's a `@Nullable` field server-side, unlike the always-present `force_merge_index`), so without this backfill a Computed attribute left at its default would read back as `null`. Do not backfill when `force_merge_index` is `false` (no valid default applies there).

## 3. `PutIlm` transport change

- [ ] 3.1 In `internal/clients/elasticsearch/ilm.go`, change `PutIlm` to call `typedClient.Ilm.PutLifecycle(policy.Name).Raw(bytes.NewReader(policyBytes)).Do(ctx)` instead of unmarshaling into `putlifecycle.Request` and calling `.Request(&req)`
- [ ] 3.2 Remove the now-unused `putlifecycle.Request` unmarshal step; drop the `putlifecycle` import if nothing else in the file uses it
- [ ] 3.3 Double check `bytes` is imported (already used by `ClearILMPolicyFromIndices` in the same file)
- [ ] 3.4 Confirm error handling around the `.Raw()` call still surfaces Elasticsearch's error body the same way `.Request()` did (compare against `ClearILMPolicyFromIndices`'s existing `.Raw()` error handling for consistency)

## 3a. `GetIlm` read-path change (required for the write-path fix to actually round-trip)

`PutIlm`'s `.Raw()` swap alone is not sufficient: `GetIlm` decodes the GET Lifecycle response via `.Do(ctx)` into a typed `*types.Lifecycle`, whose `SearchableSnapshotAction` field has the identical gap (no `ForceMergeOnClone`), so a value written successfully would be silently dropped again on the very next read — including the mandatory read-after-write in create/update. Without this section, the feature would write correctly but never persist in state.

- [ ] 3a.1 In `internal/clients/elasticsearch/ilm.go`, change `GetIlm` to call `typedClient.Ilm.GetLifecycle().Policy(policyName).Perform(ctx)` instead of `.Do(ctx)`, following the existing `.Perform(ctx)` + manual decode pattern already used by `GetIndicesWithILMPolicy` in the same file (there for the analogous reason: the typed `Lifecycle` struct doesn't expose `in_use_by`)
- [ ] 3a.2 Decode the subset of the raw response body the provider needs — `modified_date`, `policy._meta`, and each phase's `min_age`/`actions` — into local structs, leaving `actions` as `map[string]map[string]any` per phase rather than a typed `IlmActions`/`SearchableSnapshotAction`
- [ ] 3a.3 Update `GetIlm`'s signature/return type and its caller (`readILM` in `read.go`, `readPolicyIntoModel` in `policy.go`) to match; `readPolicyIntoModel`'s existing re-marshal-then-unmarshal step (`policy.go:102-113`) that builds a generic `actions map[string]map[string]any` from `phase.Actions` may become unnecessary once `GetIlm` already returns actions in that shape
- [ ] 3a.4 Confirm error handling (404-as-not-found, other HTTP errors) is preserved equivalently to the current `.Do(ctx)`-based `GetIlm` (compare against `GetIndicesWithILMPolicy`'s existing `.Perform(ctx)` error handling for consistency)

## 4. Tests

- [ ] 4.1 Add a new `SkipFunc: versionutils.CheckIfVersionIsUnsupported(<9.2.1 var>)`-gated `resource.TestStep` to `TestAccResourceILMSearchableSnapshotPhases` in `internal/elasticsearch/index/ilm/acc_test.go` (do not add a new test function; note the existing `SkipFunc` convention at `acc_test.go:425` is used by `TestAccResourceILMFrozenPhase`, not this test — apply the same convention here, not a copy of that test's exact gating), exercising `force_merge_on_clone = false` on at least one phase's `searchable_snapshot` block
- [ ] 4.2 Add any `testdata` config fixtures the new step needs (new `ConfigDirectory`, or extend the existing `create`/`update` fixtures if the step reuses them)
- [ ] 4.3 Ensure the new acceptance step includes a plan-after-apply check (e.g. `ExpectNonEmptyPlan: false` / a trailing `PlanOnly` step, matching existing conventions in this test file) that asserts no diff on `force_merge_on_clone` after apply — this specifically catches the read-path regression described in section 3a if it's ever reintroduced
- [ ] 4.4 Add table-driven unit test coverage for `expandAction`'s `ilmActionSearchableSnapshot` case covering at least these branches, not just the non-default/unsupported-version one: (a) default (`true`) on an unsupported (pre-9.2.1) server → omitted from payload, no error; (b) non-default (`false`) on an unsupported server → "Unsupported ILM setting" diagnostic, following the pattern used for `attrAllowWriteAfterShrink`; (c) `force_merge_index = false` with `force_merge_on_clone` configured (any value) → rejected by the 1a.1/1a.2 validator; (d) `force_merge_index = false` with `force_merge_on_clone` left unset → no error, field omitted from payload
- [ ] 4.5 Run `make build`, `go vet ./...`, and the ILM package's unit tests locally (acceptance tests require a live 9.2.1+ Elasticsearch cluster and are out of scope for this proposal's authoring step)

## 5. Docs and spec sync

- [ ] 5.1 Update `openspec/specs/elasticsearch-index-lifecycle/spec.md`: extend the version-gated settings requirement (REQ-022–REQ-025) or add a new requirement documenting `force_merge_on_clone` and its `9.2.1` gate, and update the `searchable_snapshot` schema block shown in the "Action block shapes" section
- [ ] 5.2 Regenerate/update provider documentation for `elasticstack_elasticsearch_index_lifecycle` to describe the new attribute
- [ ] 5.3 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate ilm-force-merge-on-clone --type change` and fix any reported issues
- [ ] 5.4 After implementation lands and this change is archived, run `make check-openspec` to confirm the main spec absorbed the delta cleanly

## 6. Follow-up (not blocking this change)

- [ ] 6.1 File an issue against `elastic/elasticsearch-specification` requesting `force_merge_on_clone` support in the generated `SearchableSnapshotAction` type, so `PutIlm` can eventually revert from `.Raw()` back to the typed `.Request()` path
