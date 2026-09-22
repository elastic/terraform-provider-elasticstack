## 1. Fix order-dependent assertions in `internal/kibana/spaces/data_source_test.go`

- [ ] 1.1 In `TestAccSpacesDataSource`, replace the `spaces.0.*` checks for the default space (`id`, `name`, `description`, `disabled_features.#`, `color`, `initials`/`image_url`/`solution` absence checks) with `testCheckSpaceAttrByID("default", ...)` / `testCheckDataSourceAttrEmptyOrAbsent`-style lookups keyed by id instead of index.
- [ ] 1.2 In `TestAccSpacesDataSource_multipleSpaces`, replace the `spaces.0.id`/`spaces.0.name`/`spaces.0.description` checks for the default space with `testCheckSpaceAttrByID("default", ...)`, and remove both stale ordering comments: `// Default space is always the first element.` and the function comment claiming the `tfacc` prefix reliably sorts after `default`.
- [ ] 1.3 In `TestAccSpacesDataSource_noDescription`, `TestAccSpacesDataSource_withImageURL`, and `TestAccSpacesDataSource_withKibanaConnection` (both the secure and insecure connection checks), replace each `spaces.0.id == "default"` check with `testCheckSpaceAttrByID("default", "id", "default")` (or fold the default-space existence check into the surrounding composed check without relying on index 0).
- [ ] 1.4 Confirm no other index-based (`spaces.<N>.*`) assumption about a specific space remains in the file.

## 2. Spec sync

- [ ] 2.1 Add a requirement to `openspec/specs/kibana-spaces/spec.md` documenting that the `spaces` list has no ordering contract and that acceptance tests MUST assert list entries by `id`, not by fixed index.
- [ ] 2.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate kibana-spaces-order-independent-tests --type change` and fix any reported problems.

## 3. Verification (implementation phase, not this proposal PR)

- [ ] 3.1 Run `go build ./...` and `go vet ./internal/kibana/spaces/...`.
- [ ] 3.2 Run the five affected acceptance tests with `TF_ACC=1` against a running Kibana/Elasticsearch stack per `dev-docs/high-level/testing.md`: `TestAccSpacesDataSource`, `TestAccSpacesDataSource_multipleSpaces`, `TestAccSpacesDataSource_noDescription`, `TestAccSpacesDataSource_withImageURL`, `TestAccSpacesDataSource_withKibanaConnection`.
- [ ] 3.3 If a `9.6.0-SNAPSHOT` (or later) stack is available, additionally confirm the previously-failing `TestAccSpacesDataSource_withImageURL` passes against it.
