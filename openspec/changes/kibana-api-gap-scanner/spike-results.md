# Spike results (task 2.3)

Status: input to the go/no-go decision. The decision is the maintainer's; nothing here ticks task 2.3.

Command (repository at `kibana-api-gap-scanner` branch, after task 2.1/2.2):

```
go run ./scripts/kibana-api-gaps discover --memory <absent> --slots 3 \
  --resource elasticstack_kibana_slo \
  --resource elasticstack_kibana_synthetics_monitor \
  --resource elasticstack_fleet_agent_policy
```

Runtime is about 5 seconds for a filtered run. Two runs produce byte-identical `report.json`.

## Headline numbers (three resources)

| | SLO | Synthetics monitor | Fleet agent policy | Total |
|---|---|---|---|---|
| Operations reached | 6 | 3 | 4 | 13 |
| Gap groups | 3 | 2 | 33 | 38 |
| Unused fields in groups | 3 | 2 | 37 | 42 |
| Skipped shapes (unique pointers) | 0 | 5 | 20 | 25 |
| Resources with zero operations | none | none | none | none |

No tool errors, no unmatched json tags.

## Signal versus noise

I classified every group by hand (distinct logical field, counting create and update bodies once).
The 38 groups collapse to 21 logical fields, because 16 fields appear once per operation body (create and update).

- **Plausible true gaps (about 12 logical fields, about 57%).** Fleet: `overrides`, `agentless`, `keep_monitoring_alive`, `min_agent_version`, `download_source_ids`, `advanced_settings.*` (3 of 13 leaves reported), `is_default`, `is_default_fleet_server`, `is_managed`. SLO: `settings.projectRoutings` (a recent Kibana addition, confirmed unused by grep). Synthetics: `certificate_error_spki_allowlist`, HTTP `check`. Grepping the provider for the Go field names of each confirmed that none is set anywhere in the provider.
- **Server-managed, action, or read-only noise (about 9 logical fields, about 43%).** `has_agent_version_conditions`, `package_agent_version_conditions`, `has_fleet_server`, `is_verifier`, update-body `id`, `bumpRevision`, `force` (create, update, delete), GET/PUT `format` query parameter, SLO GET `instanceId`. These are exactly the categories the triage prompt is meant to suppress, so this is acceptable noise for an LLM filter. Query parameters on GET and DELETE operations (3 groups) are mostly low value.
- **Union or passthrough noise: none in these three resources.** Union branches resolve correctly (the fixture-proven path), and no opaque-covered type leaked into a group.
- **Create versus update duplication: 16 of 21 logical fields produce two groups.** The same field yields one group for the create body and one for the update body, so a single missing attribute would take two issue slots and two triage verdicts. This is the largest avoidable noise source seen. It does not violate the spec (the group key is resource plus spec location), but it halves the effective slot budget for Fleet.
- **Skipped shapes** are visible and safe: 20 Fleet `advanced_settings.*` leaves are untyped (`interface{}`) in the Go client, and 5 unique synthetics map-valued fields (`labels`, `params`, `playwright_options`, `proxy_headers`, `response`). These hide possible real gaps (false negatives), never false absence.

## Scale check outside the spike (whole repository, for risk assessment)

The three chosen resources are the quiet end of the distribution. A whole-repo run reports 1,632 groups across 45 analysed resources (40 resources have no mapped operation, all Elasticsearch ones), 110 skipped shapes, 67 query-parameter groups. Two resources dominate:

- `elasticstack_kibana_security_detection_rule`: 583 groups. Cause: the resource fills the shared fields (`enabled`, `tags`, `author`, `actions`, ...) through a reflection helper, `buildCommonRuleProps` in `internal/kibana/security_detection_rule/common_props.go` (`reflect.ValueOf(props).Elem().FieldByName`). Type-based usage analysis cannot see reflective field access, so these are **false absences**, which contradicts the design's "never reports a false absence" guarantee and the spec's "absence is proven" requirement.
- `elasticstack_kibana_alerting_rule`: 506 groups. Cause: the request body is a union of about 36 rule-type variants (each `...create-rule-body-alerting` schema reports the same 14 common fields as unused), and the resource sends a generic body rather than the variant types. This is the union noise the design worried about, at scale, and it also includes fields the resource clearly sets (`name`, `enabled`, `schedule`).

These two alone are 67% of all groups. They would flood the first scheduled runs' triage queue (15 per run, ordered by resource rotation and size) with groups that are false or duplicated, even though triage can suppress them and memory records the suppression.

## The v2 question: fixed-value-only fields

Scope reminder: the scanner counts any reference as "used". The v2 question is how many fields that look unused are really set elsewhere to fixed values (and so are not gaps), and conversely how many "used" fields are only ever set to constants (hidden gaps).

- Looking-unused fields in the three resources that are set elsewhere: **0 of 21 logical fields.** Each Go field name was grepped across `internal/` outside tests and generated code with no hit in the resource, its client helpers, or shared helpers (the one apparent exception, `IsDefault`, is in test support for a different resource).
- Used fields that are only set to fixed values: the reviewed bodies (Fleet agent policy create and update, SLO settings, synthetics monitor `Enabled`) show **no field assigned only a literal**; literals appear as enum members (`monitoring_enabled` logs and metrics), which are values of a used field, not constant-only fields. Sample is small (about 25 fields), so the estimate is **under 5%, low confidence**.
- Conclusion: the constant-only blind spot is not a material source of error in these resources; v2 does not need to be planned yet. The reflection blind spot above is far larger.

## Recommendation

**Conditional GO for the three spike resources, with two required fixes before the scheduled run is enabled. Do not enable the whole-repo schedule as it stands.** The human makes the final call.

Evidence for going ahead: discovery is fast (5 s filtered), deterministic, error-free on real shapes, and about 57% of the Fleet, SLO, and synthetics groups are plausible real gaps with the remainder falling into suppress categories the LLM triage is designed for. A recent Kibana field (`settings.projectRoutings`) is surfaced correctly.

Required before PR 2 (schedule), preferably before PR 1 lands, both small and local to `scripts/kibana-api-gaps`:

1. **Treat reflection as opaque.** A type passed to `reflect.ValueOf`/`reflect.TypeOf` (including through generic helpers such as `buildCommonRuleProps[T]`) should count as opaque-covered, like whole-value JSON passthrough. Otherwise 583 detection-rule groups are false absences. This restores the soundness guarantee in the spec.
2. **Reduce union and duplicate noise.** Either collapse sibling union variants that report identical field sets into one group, or treat a resource that never constructs a variant type as not using that variant (so its 36 alerting-rule variants do not each report), and consider merging create and update bodies for the same field name into one group. Without this, first runs are dominated by duplicates.

Resources that should not be scanned until the above is done: `elasticstack_kibana_security_detection_rule`, `elasticstack_kibana_alerting_rule` (the spec has no per-resource exclusion; the options are to fix, or accept that triage and the 15-group budget will spend several days suppressing them).

If neither fix is wanted, the honest alternative is **NO-GO** on the whole-repo scanner and a narrower v1 limited to an allow-list of resources.
