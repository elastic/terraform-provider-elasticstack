## 1. Regression tests (write first)

- [ ] 1.1 Add a `resource.UnitTest` in `internal/elasticsearch/security/role/` (no `TF_ACC`) that points the provider at a mock Elasticsearch: GET/PUT/DELETE `/_security/role/{name}` plus cluster info, echoing omitted `allow_restricted_indices` as `false`. Config matches the #4759 HCL (`dynamic "indices"`, field omitted). Step 1 creates one element (`logs-*`); step 2 appends a second (`metrics-*`).
- [ ] 1.2 Confirm the mock test fails against the unfixed schema with "Provider produced inconsistent result after apply" / planned `allow_restricted_indices` null / "does not correlate with any element in actual".
- [ ] 1.3 Add a live acceptance test for the same create-then-append sequence on `indices`.
- [ ] 1.4 Add the equivalent live acceptance test for `remote_indices`, with `SkipFunc: versionutils.CheckIfVersionIsUnsupported(role.MinSupportedRemoteIndicesVersion)` (8.10.0), matching existing remote_indices acc steps.

## 2. Schema fix

- [ ] 2.1 In `internal/elasticsearch/security/role/schema.go`, change `attrAllowRestrictedIndices` on the `indices` `SetNestedBlock` from `PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()}` to `Default: booldefault.StaticBool(false)`.
- [ ] 2.2 Apply the same change to `attrAllowRestrictedIndices` on the `remote_indices` `SetNestedBlock`.
- [ ] 2.3 Add the `github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault` import; drop the `boolplanmodifier` import if no longer used elsewhere in the file. Do not bump `CurrentSchemaVersion`.
- [ ] 2.4 Confirm the mock UnitTest from 1.1 now passes, then the live acc tests from 1.3–1.4.

## 3. Package checks

- [ ] 3.1 Run `go test ./internal/elasticsearch/security/role/` (no `TF_ACC`) and `go vet ./internal/elasticsearch/security/role/...`. Do not rewrite `models_test.go` hand-built `types.BoolNull()` fixtures; those are not plan-time values.

## 4. Documentation

- [ ] 4.1 Run `make docs-generate` and commit any generated resource-doc updates that surface the new `false` default.
- [ ] 4.2 Skim `descriptions/allow_restricted_indices.md` only to confirm it does not claim “unset means preserve prior state”; leave it unchanged if it does not.

## 5. Spec sync

- [ ] 5.1 When syncing the delta into `openspec/specs/elasticsearch-security-role/spec.md`, update the Schema HCL sketch so both `indices.allow_restricted_indices` and `remote_indices.allow_restricted_indices` read `<optional, computed, bool, default false>`.
- [ ] 5.2 Run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate fix-security-role-allow-restricted-indices-default --type change` and resolve any reported issues.
- [ ] 5.3 Run `make check-openspec` after implementation to confirm specs and code stay in sync.
