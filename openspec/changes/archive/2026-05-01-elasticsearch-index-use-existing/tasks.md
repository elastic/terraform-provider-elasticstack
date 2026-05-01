## 1. Schema and model

- [x] 1.1 Add the optional `use_existing` boolean attribute (default `false`, no plan modifiers) to `internal/elasticsearch/index/index/schema.go` with a description that explains: opt-in, create-time-only, static names only, strict on static settings, full ownership after adopt.
- [x] 1.2 Add the corresponding `UseExisting types.Bool` field to `tfModel` in `internal/elasticsearch/index/index/models.go`.
- [x] 1.3 Regenerate `docs/resources/elasticsearch_index.md` so the new attribute is documented.

## 2. Static-setting strict comparison helper

- [x] 2.1 Add a helper that, given a plan `tfModel` and a `models.Index` returned by Get Index, walks `staticSettingsKeys` and reports a list of mismatches for static settings that are explicitly set in the plan (`Known()` and not null) and differ from the existing index value.
- [x] 2.2 Normalize comparison values so Elasticsearch's stringly-typed responses (`"1"`, `"true"`) compare equal to the typed plan values (`int64(1)`, `true`) for each `staticSettingsKeys` attribute type used in the schema.
- [x] 2.3 Add unit tests for the helper covering: no mismatches, single mismatch, multiple mismatches, plan attribute null/unknown, value-type normalization for ints, bools, strings, and the list/set typed static settings (`sort.field`, `sort.order`).

## 3. Adopt branch in Create

- [x] 3.1 Refactor `Update` in `internal/elasticsearch/index/index/update.go` so the alias, dynamic-setting, and mapping reconciliation logic is callable as package-private helpers taking explicit arguments (synthetic state vs plan). Preserve all existing semantics.
- [x] 3.2 In `internal/elasticsearch/index/index/create.go`, when `use_existing = true`:
  - [x] 3.2.1 If the configured `name` matches `esclient.DateMathIndexNameRe`, add a warning diagnostic explaining `use_existing` is ignored for date math names and fall through to the existing create path.
  - [x] 3.2.2 Otherwise call `elasticsearch.GetIndex` for the configured `name`. If the response is `nil` (404), fall through to the existing create path.
  - [x] 3.2.3 If the response returns an existing index, build a synthetic `tfModel` from the API response using the existing `populateFromAPI` flow.
  - [x] 3.2.4 Run the static-setting strict comparison; on any mismatch, return an error diagnostic listing each mismatch and stop without mutating the cluster.
  - [x] 3.2.5 Otherwise call the alias, dynamic-setting, and mapping reconciliation helpers with the synthetic state vs the plan.
  - [x] 3.2.6 Compute `id` from the cluster UUID and the existing concrete index name, set `ConcreteName`, perform a final read, and add a warning diagnostic stating that the existing index was adopted (including the concrete name).
- [x] 3.3 When `use_existing = false`, ensure the create path is byte-for-byte the same as today (no warning, no extra API call).

## 4. Acceptance and unit coverage

- [x] 4.1 Add an acceptance test that creates an index out-of-band (or via an index template that pre-creates a concrete index), then applies a Terraform configuration with `use_existing = true` and verifies: success, the warning diagnostic, no spurious follow-up plan, and that subsequent updates run through the normal update path.
- [x] 4.2 Add an acceptance test where the existing index has a different `number_of_shards` than config and verify the apply fails with an error diagnostic listing the mismatch and no mutation occurs (existing index settings remain unchanged).
- [x] 4.3 Add an acceptance test for `use_existing = true` plus a date math `name` that asserts the warning is emitted and the create proceeds normally.
- [x] 4.4 Add unit tests covering the create handler's branching (existence check skipped for date math, 404 falls through, 200 with mismatching static settings errors, 200 with matching static settings runs reconciliation and emits the warning).

  Branches in **task 4.4** are fully covered in acceptance tests (`TestAccResourceIndexUseExistingDateMath`, `TestAccResourceIndexUseExistingFallthrough`, `TestAccResourceIndexUseExistingMismatch`, `TestAccResourceIndexUseExistingAdopt`, plus `…AdoptAliasReconcile`, `…TemplateNoMappingDrift`); unit tests cover helper extraction only—injectable ES client for `Create` is deferred (see change `design.md`, section “Testing: use_existing create-branch coverage”).

## 5. Changelog and references

- [x] 5.1 Add a `CHANGELOG.md` entry under "Unreleased" describing the new `use_existing` attribute and referencing issue #966.
- [x] 5.2 Cross-link the new attribute in the resource description / schema docs so practitioners encountering the "already exists" error during replacement can discover the workaround.
