## Why

The `alerts_filter` attribute on `actions` in `elasticstack_kibana_security_detection_rule` has been broken since its introduction ([#3110](https://github.com/elastic/terraform-provider-elasticstack/issues/3110), originally reported in [#1446](https://github.com/elastic/terraform-provider-elasticstack/issues/1446#issuecomment-4477817637)). The attribute is typed as `schema.MapAttribute{ElementType: types.StringType}`, a flat string-to-string map. Kibana's Detection Engine API expects a nested JSON object:

```json
{
  "query": {
    "kql": "event.action : \"test_case_a\"",
    "filters": []
  },
  "timeframe": {
    "days": [1, 2, 3, 4, 5],
    "hours": { "start": "00:00", "end": "24:00" },
    "timezone": "UTC"
  }
}
```

**Read-path bug**: the code iterates the top-level response map keys and stringifies each value with `fmt.Sprintf("%v", v)`. For the `"query"` key (a nested map), Go produces `"map[filters:[] kql:event.action : \"test_case_a\"]"` — a Go map literal, not JSON.

**Write-path bug**: the flat string map is re-assembled and sent to the API, which rejects it with `400: [actions.0.alertsFilter.query]: could not parse object value from json input`.

## What Changes

Replace the `MapAttribute(string)` `alerts_filter` with a **structured nested block** that explicitly models `query` (with `kql` and `filters_json` attributes) and an optional `timeframe` sub-block, aligned with the `elasticstack_kibana_alerting_rule` resource's existing `alerts_filter` block.

### Schema sketch

```hcl
resource "elasticstack_kibana_security_detection_rule" "example" {
  # ... other fields ...

  actions {
    id             = elasticstack_kibana_action_connector.foo.connector_id
    action_type_id = ".slack"
    params         = jsonencode({ message = "alert fired" })

    alerts_filter {
      query {
        kql          = "event.action : \"test_case_a\""
        filters_json = jsonencode([])   # JSON-normalized array of Kibana filter objects
      }
      timeframe {                       # optional
        days        = [1, 2, 3, 4, 5]
        timezone    = "Europe/London"
        hours_start = "08:00"
        hours_end   = "17:00"
      }
    }
  }
}
```

Key schema decisions:
- `alerts_filter` becomes a `schema.SingleNestedBlock` (from `schema.MapAttribute`)
- `query` is a nested block with `kql` (optional string) and `filters_json` (optional `jsontypes.Normalized` JSON string)
- The `filters_json` attribute name (rather than `filters`) allows a future typed `filters` list to be added without conflict
- `timeframe` is an optional nested block with `days`, `timezone`, `hours_start`, `hours_end` — identical to the `elasticstack_kibana_alerting_rule` shape
- Schema version bumps from 1 → 2 with a no-op `StateUpgrade` (the feature has been broken; no valid state exists to migrate)

### Modified capabilities

- `kibana-security-detection-rule`: Replace broken `alerts_filter` map attribute with structured nested block (REQ-080–REQ-087)

## Impact

- **Specs**: Delta under `openspec/changes/kibana-security-detection-rule-alerts-filter/specs/kibana-security-detection-rule/spec.md` until synced/archived.
- **Implementation** (future): `internal/kibana/security_detection_rule/` (schema, models, expand/flatten helpers, state migration), documentation, acceptance tests.
- **Breaking schema change**: practitioners must rewrite any `alerts_filter` configuration from the flat map form to the new nested block form. Since the feature was functionally broken, no valid state exists to migrate non-destructively.
