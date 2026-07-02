## Why

Kibana 9.5 changed its dashboard API behavior: when a dashboard is created or updated without
a `description`, the API now persists and returns `description: ""` (empty string) instead of
omitting the field or returning `null` as it did in 8.x and 9.4.

The `elasticstack_kibana_dashboard` resource declares `description` as an Optional `schema.StringAttribute`.
When a practitioner omits `description`, Terraform plans it as `null`. After apply, the Read path calls
`typeutils.StringishPointerValue(data.Data.Description)` which returns `types.StringValue("")`
when Kibana echoes back an empty string — not `null`. Terraform then flags an inconsistent result:

```
Error: Provider produced inconsistent result after apply

When applying changes to elasticstack_kibana_dashboard.test, provider
"provider[\"registry.terraform.io/hashicorp/elasticstack\"]" produced an
unexpected new value: .description: was null, but now cty.StringVal("").
```

This breaks approximately 14 acceptance tests on 9.5.0-SNAPSHOT and affects real users on 9.5 who
omit the `description` attribute. The fix is a localized normalization in the Read path with
intent preservation: when the planned value was `null`, map an API-returned empty string back to
`null` in state; when the user explicitly set `description = ""`, preserve `""`.

## What Changes

Add intent-preserving null/empty-string normalization to the dashboard `description` read path.
Specifically, the `models.go` `FromAPI` function SHALL use a plan-aware check: when the prior
Terraform state/plan had `description` as null, and the Kibana API returns `""`, store `null` in
state rather than `""`. This aligns with the REQ-009 null-preservation pattern already applied to
other optional fields.

No schema changes are needed. No migration is required.

## Capabilities

### Modified Capabilities

- `kibana-dashboard`: extend REQ-008 (read behavior) and REQ-009 (state preservation) to cover
  the root-level `description` attribute — when the API returns an empty-string `description` and
  prior state had `description` null, the provider SHALL store `null` in state, not `""`.

## Impact

- `internal/kibana/dashboard/models.go` — change `m.Description` assignment from
  `typeutils.StringishPointerValue(...)` to intent-preserving logic that substitutes
  `types.StringNull()` when the API returns `""` and prior intent was `null`.
- `openspec/changes/dashboard-description-null-normalization/specs/kibana-dashboard/spec.md` —
  delta spec extending REQ-008 and REQ-009 to cover root-level `description` normalization.
- Acceptance tests covering: omitted `description` round-trips as null on 9.5+; explicit
  `description = ""` is preserved; explicit non-empty `description` round-trips unchanged.
