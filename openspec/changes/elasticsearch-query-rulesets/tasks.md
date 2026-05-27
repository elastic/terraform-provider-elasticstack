## 1. Spec

- [ ] 1.1 Keep delta spec aligned with `proposal.md` / `design.md`; run `OPENSPEC_TELEMETRY=0 ./node_modules/.bin/openspec validate elasticsearch-query-rulesets --type change` (or `make check-openspec`) after any spec edits.
- [ ] 1.2 Resolve the open question on minimum ES version guard (8.10 tech-preview vs 8.12 GA); update delta spec with the confirmed version compatibility requirement.
- [ ] 1.3 Verify whether `GET /_query_rules/{ruleset_id}` returns rules in declaration order; if not, document the stable-sort-by-`rule_id` approach in the spec.
- [ ] 1.4 On completion of implementation, **sync** delta into `openspec/specs/elasticsearch-query-rulesets/spec.md` or **archive** the change per project workflow.

## 2. Client wrapper

- [ ] 2.1 Create `internal/clients/elasticsearch/queryrulesets.go` with functions:
  - `PutQueryRuleset(ctx, rulesetID string, rules []QueryRuleModel) error`
  - `GetQueryRuleset(ctx, rulesetID string) (*QueryRulesetModel, error)` — returns `nil, nil` on 404
  - `DeleteQueryRuleset(ctx, rulesetID string) error`
  Using the `go-elasticsearch` typed client (`typedapi/queryrules/putruleset`, `getruleset`, `deleteruleset`).
- [ ] 2.2 Add an `internal/models` or inline model type for `QueryRulesetModel`, `QueryRuleModel`, `QueryRuleCriteriaModel`, and `QueryRuleActionsModel` that map cleanly to/from both the API types and the Terraform schema.

## 3. Resource implementation

- [ ] 3.1 Create `internal/elasticsearch/queryrulesets/` package with:
  - `schema.go` — `Schema()` function defining the full resource schema (see design.md for shape); include `elasticsearch_connection` block, `ruleset_id` with RequiresReplace, `id` as computed, `rules` as ListNestedAttribute.
  - `model.go` — `queryRulesetModel`, `queryRuleModel`, `queryRuleCriteriaModel`, `queryRuleActionsModel`, `queryRuleActionDocModel` types with `tftypes` tags.
  - `resource.go` — implements `resource.Resource`; wire `Create`, `Read`, `Update`, `Delete`, `ImportState`.
  - `data_source.go` — implements `datasource.DataSource`; wire `Read`.
- [ ] 3.2 Implement `Create`: call `PutQueryRuleset`, then `GetQueryRuleset` to populate computed fields; set `id` to `<cluster_uuid>/<ruleset_id>`.
- [ ] 3.3 Implement `Read`: call `GetQueryRuleset`; if 404, call `resp.State.RemoveResource(ctx)` and return; otherwise map API response to state. Handle rule ordering: if the API does not preserve order, sort by `rule_id` to avoid perpetual plan diffs (verify during acceptance tests; document chosen approach).
- [ ] 3.4 Implement `Update`: call `PutQueryRuleset` with the new full rule list; call `GetQueryRuleset` to refresh state.
- [ ] 3.5 Implement `Delete`: call `DeleteQueryRuleset`.
- [ ] 3.6 Implement `ImportState`: parse `<cluster_uuid>/<ruleset_id>` import ID; set `id` and `ruleset_id`; delegate to `Read`.
- [ ] 3.7 Add plan-time validator for `actions` mutual exclusion: exactly one of `ids` or `docs` must be set per rule.
- [ ] 3.8 Add plan-time validator (or schema-level `Validators`) for `criteria.values`: must be a valid JSON array string when `criteria.type != "always"`.
- [ ] 3.9 Add a minimum ES version check for the Query Rules API (see open question 1 in design.md); surface a clear diagnostic if the cluster version is below the minimum.

## 4. Provider registration

- [ ] 4.1 Register `elasticstack_elasticsearch_query_ruleset` resource in `provider/plugin_framework.go` (or equivalent registration file).
- [ ] 4.2 Register `data.elasticstack_elasticsearch_query_ruleset` data source in the same file.

## 5. Documentation

- [ ] 5.1 Add `MarkdownDescription` to every schema attribute (resource and data source).
- [ ] 5.2 Create `templates/resources/elasticsearch_query_ruleset.md.tmpl` (or equivalent docs template) with usage example showing a `pinned` rule and an `exclude` rule.
- [ ] 5.3 Create `templates/data-sources/elasticsearch_query_ruleset.md.tmpl` with a lookup example.
- [ ] 5.4 Run `make generate-docs` and verify output in `docs/`.

## 6. Testing

- [ ] 6.1 Acceptance test — basic CRUD: create a ruleset with one `pinned` rule and one `exclude` rule; assert state matches; update rules (add a third rule); assert state; destroy.
- [ ] 6.2 Acceptance test — rule ordering: create ruleset with multiple rules; assert state preserves declaration order; run plan again; assert no diff.
- [ ] 6.3 Acceptance test — `criteria.values` with numeric values: create a rule with `criteria.type = "gt"` and `values = jsonencode([100])`; assert state round-trips the JSON string correctly.
- [ ] 6.4 Acceptance test — `actions.docs` variant: create a rule using `docs = [{_index = "my-index", _id = "42"}]` instead of `ids`; assert state.
- [ ] 6.5 Acceptance test — `criteria.type = "always"`: create a rule with an `always` criterion (no metadata or values); assert accepted and stored.
- [ ] 6.6 Acceptance test — import: create resource, import by composite ID `<cluster_uuid>/<ruleset_id>`, verify imported state matches, run plan, confirm no diff.
- [ ] 6.7 Acceptance test — data source: create resource, read via `data.elasticstack_elasticsearch_query_ruleset`, assert all attributes match.
- [ ] 6.8 Acceptance test — not-found: delete ruleset outside Terraform; run refresh; assert resource is removed from state.
- [ ] 6.9 Unit tests for `criteria.values` JSON validation and `actions` mutual-exclusion validator logic.
- [ ] 6.10 If a minimum ES version is confirmed (task 1.2), add a `SkipFunc` to all acceptance tests that skips on unsupported cluster versions (mirroring the pattern used in other acceptance tests in this repo).
