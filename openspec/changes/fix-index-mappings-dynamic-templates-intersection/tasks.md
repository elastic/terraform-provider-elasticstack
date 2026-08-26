## 1. Package consolidation

- [x] 1.1 Move `internal/elasticsearch/index/indexmappings/intersect.go` into `internal/elasticsearch/index` (e.g. `internal/elasticsearch/index/intersect.go`); move `intersect_test.go` alongside it
- [x] 1.2 Rename `intersectMappings` to exported `IntersectMappings`; keep `intersectProperties` unexported
- [x] 1.3 Update the call site in `internal/elasticsearch/index/indexmappings/read.go:73` to `index.IntersectMappings(apiMap, stateMap)`
- [x] 1.4 Remove the now-empty `indexmappings/intersect.go` and fix package declarations/imports in the moved files

## 2. Name-keyed `dynamic_templates` intersection

- [x] 2.1 Add a `dynamicTemplatesKey = "dynamic_templates"` constant (parallel to the existing `propertiesKey`) in the relocated file
- [x] 2.2 In `IntersectMappings`'s "API key absent" early return (`if !ok { ... }`), add a `dynamicTemplatesKey` case that `continue`s without writing to `result` (drop, not keep-`stateVal`) — this covers the API omitting `dynamic_templates` entirely, e.g. every declared template removed out-of-band (decision 2)
- [x] 2.3 Implement `intersectDynamicTemplates(apiVal, stateVal any) (templates []any, ok bool)`: parse both sides via `dynamicTemplatesByName`; on failure for either side, return `ok = false` (decision 3, passthrough); on success, walk the **state's** original `[]any` order, keep each declared name's entry using the API's definition when present in `dynamicTemplatesByName(apiVal)`, and omit names absent from the API (decision 2)
- [x] 2.4 Add a `dynamic_templates` branch in `IntersectMappings` (after the `apiVal, ok := apiMappings[key]` lookup succeeds) parallel to the existing `properties` branch: on `ok == true`, set `result[key]` to the filtered templates only if non-empty (mirroring the `properties` `len(intersected) > 0` guard), then `continue`; on `ok == false`, fall through to the existing `FieldSemanticallyEqual`/passthrough logic unchanged
- [x] 2.5 Confirm the state's declared order is preserved in the returned slice (do not iterate the map from `dynamicTemplatesByName`, which has no defined order)

## 3. Order-sensitive semantic equality

- [x] 3.1 Extend `dynamicTemplatesSemanticallyEqual` (`mappings_value.go`) to additionally check that the relative order of the user's declared template names (from `userRaw`) matches their relative order within `apiRaw`, after filtering `apiRaw` down to just the declared names (ignoring API-only extras)
- [x] 3.2 Add unit test: user declares `[alpha, beta]`, API returns `[beta, alpha]` (both present, semantically equal definitions, reordered) → `dynamicTemplatesSemanticallyEqual` returns `false`
- [x] 3.3 Add unit test: user declares `[alpha, beta]`, API returns `[extra, alpha, beta]` (index-template-injected extra interleaved, declared order preserved among declared names) → `dynamicTemplatesSemanticallyEqual` returns `true`

## 4. Unit tests (intersection)

- [x] 4.1 Add unit test: state declares one template, API returns that template plus an index-template-injected extra → result contains only the declared template
- [x] 4.2 Add unit test: state declares a template name absent from the API → result omits that name (decision 2)
- [x] 4.3 Add unit test: state declares one or more templates, API omits the `dynamic_templates` key entirely → result omits the key (decision 2, covers task 2.2)
- [x] 4.4 Add unit test: state or API `dynamic_templates` has a duplicate template name → whole-array passthrough for that key (decision 3)
- [x] 4.5 Add unit test: state or API `dynamic_templates` entry value is not an object → whole-array passthrough for that key (decision 3)
- [x] 4.6 Add unit test: multiple declared templates → result preserves the state's declared order, not the API's array order
- [x] 4.7 Run `go test ./internal/elasticsearch/index/...` and confirm the relocated `properties`-intersection unit tests still pass unchanged

## 5. Acceptance test tightening

- [x] 5.1 Change `checkStateMappingsDynamicTemplates(minCount int)` in `acc_test.go` to an exact-name-set assertion (e.g. `checkStateMappingsDynamicTemplates(wantNames ...string)`) that fails if the state array's name set differs from `wantNames`
- [x] 5.2 Update **both** existing call sites — `TestAccResourceIndexMappings_allTopLevelKeys` (`acc_test.go:179`) and `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` (`acc_test.go:210`) — to pass each fixture's expected declared name(s); the latter additionally asserts `template_default` (the out-of-band-injected name) is absent
- [x] 5.3 Add a step to `TestAccResourceIndexMappings_dynamicTemplatesFromIndexTemplate` that deletes a declared template name out-of-band (via `setDynamicTemplatesOutOfBand` omitting it) and asserts the next read/plan reflects its removal from state rather than pinning the stale value

## 6. Spec sync

- [ ] 6.1 Extend `openspec/specs/elasticstack-elasticsearch-index-mappings/spec.md` REQ-004 to document `dynamic_templates` name-keyed filtering, drop-on-missing (including the fully-absent-key case), and passthrough-on-unparseable-shape semantics, with scenarios
- [ ] 6.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-index-mappings-dynamic-templates-intersection --type change` and resolve any reported issues
- [ ] 6.3 After implementation and merge, run `make check-openspec` / archive the change per the project's OpenSpec workflow so the delta is folded into the main spec
