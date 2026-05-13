## 1. Promote the resource registration

- [x] 1.1 In `provider/plugin_framework.go`, remove `dashboard.NewResource` from `Provider.experimentalResources` so it only returns `streams.NewResource`.
- [x] 1.2 In the same file, add `dashboard.NewResource` to `Provider.resources` alongside the other Kibana resources (preserve alphabetical/area grouping consistent with neighboring entries).
- [x] 1.3 Confirm the `internal/kibana/dashboard` import statement in `provider/plugin_framework.go` is still needed (it is, because it is now used from `resources`).

## 2. Update the capability spec

- [x] 2.1 Apply the `kibana-dashboard` delta from `openspec/changes/graduate-kibana-dashboard/specs/kibana-dashboard/spec.md` to `openspec/specs/kibana-dashboard/spec.md`, adding the new `Provider registration (REQ-040)` requirement with its three scenarios.
- [x] 2.2 Re-run `make check-openspec` (or `./node_modules/.bin/openspec validate --strict`) and ensure the change validates cleanly.

## 3. Regenerate provider documentation

- [x] 3.1 Run `make docs-generate` to produce `docs/resources/kibana_dashboard.md` from the resource schema.
- [x] 3.2 Inspect the generated `docs/resources/kibana_dashboard.md` page for obvious rendering issues (heading hierarchy, example block, nested attribute tables) and ensure it lists schema attributes such as `title`, `space_id`, `time_range`, and `panels`.
- [x] 3.3 Commit the generated docs file (no manual edits; future regenerations should be idempotent).

## 4. Verification

- [x] 4.1 Run `make build` to ensure the provider compiles after the registration move.
- [x] 4.2 Run `go test ./provider/...` so `TestPluginFrameworkResourcesEmbedEntityCoreResourceBase` continues to pass under `AccTestVersion` (which includes both the standard and experimental sets, so this catches accidental duplicate registrations).
- [x] 4.3 Manually verify by running `go test -run '^TestPluginFrameworkResourcesEmbedEntityCoreResourceBase$' ./provider/...` and that the resource set returned by a non-`AccTestVersion` provider (e.g. via a small ad-hoc test or by inspecting `(*Provider).Resources` output) contains `elasticstack_kibana_dashboard` without `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL` set.
- [x] 4.4 Run a representative dashboard acceptance test (e.g. `go test -tags=acceptance -run TestAccResourceKibanaDashboardMarkdownPanel ./internal/kibana/dashboard/...`) against a local stack to confirm runtime behavior is unchanged.

## 5. Practitioner-facing notes

- [x] 5.1 Add a CHANGELOG entry under the unreleased section noting that `elasticstack_kibana_dashboard` is no longer experimental and no longer requires `TF_ELASTICSTACK_INCLUDE_EXPERIMENTAL=true`, following the existing `CHANGELOG.md` format.
- [x] 5.2 Confirm no other markdown under `docs/` or `dev-docs/` references the dashboard resource as experimental; update any stale references discovered.
