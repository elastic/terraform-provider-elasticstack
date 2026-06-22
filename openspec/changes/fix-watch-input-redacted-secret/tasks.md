## 1. Extend `fromAPIModel` for `input` redaction preservation

- [ ] 1.1 Add `priorInput jsontypes.Normalized` parameter to `fromAPIModel` in
  `internal/elasticsearch/watcher/watch/models.go`
- [ ] 1.2 In the `input` unmarshalling block (lines 138–147), when `priorInput` is known, unmarshal
  it to `map[string]any` and call `mergePreserveRedactedLeaves(apiInputMap, priorInputMap)` before
  marshalling back to the state JSON string — mirroring the `mergedActions` path for `actions`
  (lines 163–179)
- [ ] 1.3 When `priorInput` is unknown or null, keep the existing behavior of marshalling the raw
  API response directly into `d.Input`

## 2. Update `readWatch` call site

- [ ] 2.1 In `internal/elasticsearch/watcher/watch/read.go`, update the `fromAPIModel` call to pass
  `state.Input` as the new `priorInput` argument alongside the existing `state.Actions` argument

## 3. Unit test coverage

- [ ] 3.1 Add unit tests for the `input` redaction-preservation path; cover:
  - HTTP basic auth `password` redacted with a prior string value → prior value preserved
  - HTTP basic auth `password` redacted with no prior value → sentinel stored as-is
  - HTTP basic auth `password` redacted with a prior non-string value (e.g. an object) → prior
    value preserved
  - Non-redacted `input` fields unchanged regardless of prior value
  - Nil/empty `input` from API falls back to `{"none":{}}` as before

## 4. Acceptance test coverage

- [ ] 4.1 Add an acceptance test in `internal/elasticsearch/watcher/watch/acc_test.go` that:
  - Creates a watch with an HTTP `input` using basic auth and a sensitive password variable
  - Verifies that `terraform apply` succeeds (no `.input: inconsistent values for sensitive
    attribute` error)
  - Verifies that a subsequent `terraform plan` produces an empty plan (no perpetual diff)
- [ ] 4.2 Add any required HCL testdata fixtures under
  `internal/elasticsearch/watcher/watch/testdata/`

## 5. Spec update

- [ ] 5.1 Update the delta spec in `openspec/changes/fix-watch-input-redacted-secret/specs/elasticsearch-watch/spec.md` to:
  - Add **REQ-030** under a new `### Requirement: Input redaction preservation (REQ-030)` section
    specifying that the resource SHALL preserve prior known Terraform values at nested `input` paths
    where the API returns `::es_redacted::`, mirroring REQ-014–016 for `actions`
  - Add scenarios for:
    - Redacted `input` secret preserved on read-after-write after create/update
    - Redacted `input` secret preserved on refresh
    - No prior `input` value (import/first read) — sentinel stored as-is
  - Update the narrative in the **REQ-023–REQ-027** `JSON field mapping — read/state` section to
    explicitly include `input` alongside `actions` in the redaction-preservation description
- [ ] 5.2 After implementation is verified, sync the delta into `openspec/specs/elasticsearch-watch/spec.md` (or archive the change) per project workflow

## 6. OpenSpec validation

- [ ] 6.1 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-watch-input-redacted-secret --type change` and resolve any reported problems
- [ ] 6.2 Confirm `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec status --change "fix-watch-input-redacted-secret" --json` shows no `blocked` state
