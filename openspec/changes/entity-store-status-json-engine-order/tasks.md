## 1. Shared normalization helper

- [x] 1.1 Add a helper (e.g. `normalizeStatusJSON(rawBody []byte) []byte`) to
      `internal/kibana/security_entity_store/helpers.go`: decode `rawBody` into
      `map[string]json.RawMessage`; if the `"engines"` key is absent or fails to decode into
      `[]json.RawMessage`, return `rawBody` unchanged; otherwise decode each element's `"type"`
      field, `sort.SliceStable` the elements by `type`, re-encode the sorted slice back into the
      map under `"engines"`, and `json.Marshal` the map. On any decode/marshal error, return the
      original `rawBody` unchanged (normalization is cosmetic; it must never turn into a hard Read
      failure).
- [x] 1.2 Unit tests for the helper: otherwise-identical bodies with two engines reversed
      (`user`,`generic` vs `generic`,`user`) normalize to the same byte output; a single-engine body
      and a `not_installed` body with no/empty `engines` preserve all fields semantically, including
      unmodeled and nested fields; malformed bodies and other explicit passthrough cases are returned
      byte-for-byte unchanged rather than panicking or erroring.

## 2. Wire the helper into both Read paths

- [x] 2.1 `internal/kibana/security_entity_store/read.go`: in `readEntityStore`, change
      `model.StatusJSON = jsontypes.NewNormalizedValue(string(rawBody))` to normalize `rawBody`
      through the new helper first.
- [x] 2.2 `internal/kibana/security_entity_store/data_source_read.go`: in
      `readEntityStoreDataSource`, apply the same change to its
      `model.StatusJSON = jsontypes.NewNormalizedValue(string(rawBody))` call.

## 3. Verify the fix against the originally-failing test

- [x] 3.1 Run `TestAccResourceKibanaSecurityEntityStore_import` against a Kibana `9.6.0-SNAPSHOT`
      (or any stack reproducing the engine-order flip) and confirm `ImportStateVerify` no longer
      reports a `status_json` diff.
- [x] 3.2 Run the full `internal/kibana/security_entity_store` unit test suite and confirm no
      regression in existing `flattenStatus`/`flattenEngines` coverage.

## 4. Spec sync

- [x] 4.1 Run
      `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate entity-store-status-json-engine-order --type change`
      and resolve any reported issues.
- [ ] 4.2 After implementation lands, sync the delta specs into
      `openspec/specs/kibana-security-entity-store/` and
      `openspec/specs/kibana-security-entity-store-status/` (or archive the change) per
      `openspec-sync-specs` / `openspec-archive-change`.
