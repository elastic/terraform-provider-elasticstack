## 1. Preserve Phase 1 reproduction coverage

- [x] 1.1 Create a worktree for the investigation.
- [x] 1.2 Add `TestAccResourceIndexTemplateNoMappingDrift` with no index `mappings`, no `ignore_changes`, and a second step asserting `plancheck.ExpectEmptyPlan()`.
- [x] 1.3 Add `TestAccResourceIndexTemplateUserMappingNoDrift` with user-owned index `mappings`, template-injected extras, no `ignore_changes`, and a second step asserting `plancheck.ExpectEmptyPlan()`.
- [x] 1.4 Run targeted acceptance coverage against the local Elastic Stack and record current behavior: no-config mapping case passes; user-owned mapping case fails with provider inconsistent result after apply.
- [x] 1.5 Temporarily skip `TestAccResourceIndexTemplateUserMappingNoDrift` with `t.Skip()` so the reproduced bug does not block CI before the fix lands.

## 2. Implement shared mapping semantics

- [x] 2.1 Extract `mappingsPlanModifier.modifyMappings` logic into a shared helper in `internal/elasticsearch/index/index/mappings_walker.go`.
- [x] 2.2 Generalize the helper beyond `properties` so it can classify top-level API-only keys such as `dynamic_templates` as non-drifting template-owned extras.
- [x] 2.3 Add table-driven unit tests for the helper, covering matching mappings, template-injected `properties`, template-injected `dynamic_templates`, removed fields retained by Elasticsearch, and incompatible user-owned type changes.

## 3. Fix semantic equality for `mappings`

- [x] 3.1 Add `mappingsType` / `mappingsValue` in `internal/elasticsearch/index/index/mappings_type.go`, string-backed by `jsontypes.Normalized`.
- [x] 3.2 Implement `StringSemanticEquals` so refreshed/API mappings that are a non-drifting superset of prior user intent compare equal.
- [x] 3.3 Switch the `mappings` attribute in `internal/elasticsearch/index/index/schema.go` from `jsontypes.NormalizedType{}` to the custom mappings type.
- [x] 3.4 Simplify `mappingsPlanModifier` so it uses the shared helper only for `RequiresReplace` decisions and no longer copies state-only mapping values into the plan.

## 4. Remove workaround and enable acceptance target

- [x] 4.1 Remove `t.Skip()` from `TestAccResourceIndexTemplateUserMappingNoDrift`.
- [x] 4.2 Remove `lifecycle { ignore_changes = [mappings] }` from `internal/elasticsearch/index/index/testdata/TestAccResourceIndexWithTemplate/create/index.tf`.
- [x] 4.3 Adjust the `TestAccResourceIndexWithTemplate` expected `mappings` assertion to match the new user-intent-preserving state behavior.

## 5. Document and verify

- [x] 5.1 Add a `CHANGELOG.md` Fixed entry referencing GitHub issue #563.
- [x] 5.2 Update any user-facing resource description or generated docs that mention the mapping/template workaround.
- [x] 5.3 Run `go test ./internal/elasticsearch/index/index -run TestAccResourceIndexTemplate -count=0` for compile coverage.
- [x] 5.4 Run targeted acceptance tests: `TestAccResourceIndexTemplateNoMappingDrift`, `TestAccResourceIndexTemplateUserMappingNoDrift`, `TestAccResourceIndexWithTemplate`, and `TestAccResourceIndexRemovingField`.
- [x] 5.5 Run `make build`.
