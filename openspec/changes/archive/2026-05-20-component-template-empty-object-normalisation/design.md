## Context

`elasticstack_elasticsearch_component_template` uses `flattenTemplateBlock`
(`internal/elasticsearch/index/componenttemplate/flatten.go`) to map the API response onto Terraform
state. The function checks `t.Mappings != nil` and `t.Settings != nil` to decide whether to emit a
non-null value. When Elasticsearch returns `"mappings": {}` or `"settings": {}` in the GET
response, Go's JSON decoder sets the field to a non-nil empty `map[string]any{}`. The function then
emits `MappingsValue("{}")` and `IndexSettingsValue("{}")` instead of their null equivalents.

The Terraform Plugin Framework performs a post-apply consistency check by calling
`StringSemanticEquals` on the planned value (`null`) and the state value (`"{}"`). When the plan
had no `mappings` or `settings` block, the planned value is `null`. `MappingsValue.StringSemanticEquals`
returns `false` for `null vs "{}"`, causing the "Provider produced inconsistent result after apply"
error.

### SDK-era errors (resolved in v0.15.0)

| Error | SDK root cause | PF fix |
|---|---|---|
| `version null → 0` | Typed decoder set zero-value int | `flattenToData` checks `Version != nil` |
| `settings format change` | No semantic equality; ES normalises short-form keys | `IndexSettingsValue.StringSemanticEquals` compares canonical form |

The remaining edge case (empty-object response) is addressed in this change.

## Goals / Non-Goals

**Goals**:
- Treat `"mappings": {}` and `"settings": {}` API responses as equivalent to absent.
- Add regression acceptance tests confirming no drift after create for the exact original issue
  scenario.

**Non-Goals**:
- Changing `MappingsValue.StringSemanticEquals` to explicitly handle `null vs {}` — the flatten
  layer normalisation makes that unnecessary.
- Addressing index templates (`elasticstack_elasticsearch_index_template`) — they were not reported
  and have a different flatten path.
- Addressing the intermittent "Root object was present, but now absent" error reported by `@breml`
  (likely an Elasticsearch eventual-consistency race on the post-write read; tracked separately).
- Making `version` `Optional + Computed` (not required for the original bug; separate issue if
  needed).

## Decisions

- **Normalise at the flatten layer**: change `!= nil` to `len(...) > 0` in `flattenTemplateBlock`
  for both mappings and settings. This is the minimal targeted fix; a nil map and an empty map both
  have `len == 0`, so the guard handles both.
- **Regression test**: add a `PlanOnly: true, ExpectNonEmptyPlan: false` step after the existing
  create step in `TestAccResourceComponentTemplate`, and add a dedicated
  `TestAccResourceComponentTemplateIssue609NoDrift` test with the exact reporter config (short-form
  settings, alias present, no mappings). The `PlanOnly` step is the authoritative no-drift signal;
  the dedicated test pins the specific HCL shape from the issue.
- **Scope**: no changes to provider schema, no new attributes.

## Risks / Trade-offs

- [Low risk] Changing `!= nil` to `len > 0` is a safe narrowing: a non-nil non-empty map is still
  a non-null value. The only behaviour change is for non-nil empty maps.
- [Low risk] If Elasticsearch does NOT return `"mappings": {}` for a given ES version, the
  `PlanOnly` acceptance test still provides regression coverage for the `null` case.
- [Non-blocking question] Does Elasticsearch 8.13.x (the original reporter's version) include
  `"mappings": {}` in GET responses for templates without mappings? Acceptance tests in CI will
  answer this definitively once the no-drift step is added.

## Open questions

- Does Elasticsearch 8.13.x return `"mappings": {}` in `GET /_component_template` responses for
  templates that have no mappings? This determines whether Approach B's empty-map normalisation is
  needed for an active production bug or only as a defensive guard for future ES versions.
  (Acceptance tests in CI will answer this if the no-drift test step is added first.)
- Is the intermittent "Root object was present, but now absent" error from `@breml` a separate
  issue worth filing independently (likely an ES eventual-consistency race on the post-write read)?
- Should `version` be changed to `Optional + Computed` to handle external tools writing `version: 0`
  as a default? (Not required for the original bug, but more future-proof.)
