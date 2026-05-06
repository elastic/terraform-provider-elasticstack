## 1. Spec

- [ ] 1.1 Validate the change with `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-watch-null-metadata-inconsistency --type change`.
- [ ] 1.2 Sync or archive the delta into `openspec/specs/elasticsearch-watch/spec.md` after implementation is verified.

## 2. Fix null-metadata round-trip

- [ ] 2.1 In `internal/elasticsearch/watcher/watch/models.go`, change the nil-metadata branch of `fromAPIModel` from:
  ```go
  if watch.Body.Metadata == nil {
      d.Metadata = jsontypes.NewNormalizedValue(`{}`)
  }
  ```
  to:
  ```go
  if watch.Body.Metadata == nil {
      d.Metadata = jsontypes.NewNormalizedValue(`null`)
  }
  ```

## 3. Acceptance test

- [ ] 3.1 Add a new acceptance test `TestAccResourceWatch_nullMetadata` in
  `internal/elasticsearch/watcher/watch/acc_test.go` that:
  - Creates a watch with `metadata = jsonencode(null)` (the string `"null"`)
  - Asserts that the `metadata` attribute in state equals `"null"` after create
  - Asserts that a subsequent plan is empty (no perpetual diff)
- [ ] 3.2 Add the matching Terraform config in
  `internal/elasticsearch/watcher/watch/testdata/TestAccResourceWatch_nullMetadata/` (or an
  equivalent `NamedTestCaseDirectory`-compatible path) containing a watch resource with
  `metadata = jsonencode(null)`.

## 4. Verification

- [ ] 4.1 Run targeted Go build: `make build`.
- [ ] 4.2 Run the targeted acceptance test for the watcher package against a live Elasticsearch
  stack using the environment described in `dev-docs/high-level/testing.md`.
