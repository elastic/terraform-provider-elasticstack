# Kibana issue draft (not filed)

Copy into [elastic/kibana](https://github.com/elastic/kibana/issues/new). Related Terraform provider issue: https://github.com/elastic/terraform-provider-elasticstack/issues/4902

---

**Title:** Dashboard API: ES|QL XY Y-metric `color: {type: static}` is accepted then replaced with `{type: auto}`

**Labels:** `bug`, `Team:Visualizations` (adjust to the current Lens/Dashboard API team)

## Summary

`POST /api/dashboards` accepts an ES|QL XY data-layer Y metric with `color: { "type": "static", "color": "#54B399" }` (the published `KibanaHTTPAPIsXyLayerESQL.Y.Color` union includes `StaticColor`) and returns **201**. The **same create response** already has `color: { "type": "auto" }`. A subsequent GET repeats `{type: auto}`.

This is not a client drop: the request body contains the static color. The same layer's `breakdown_by.color` categorical mapping **is** persisted.

## Environment

- Kibana `9.6.0` snapshot (`build_snapshot: true`, observed build `daf212d53e88`)
- Elasticsearch `9.6.0-SNAPSHOT`

## Reproduce

Create a dashboard whose first panel is a typed Lens XY `vis` with an ES|QL data layer. Send a Y metric like:

```json
"y": [
  {
    "color": {
      "color": "#54B399",
      "type": "static"
    },
    "column": "system.cpu.user.pct",
    "format": { "type": "number" }
  }
]
```

Observed **201** body for that metric:

```json
"y": [
  {
    "column": "system.cpu.user.pct",
    "format": {
      "type": "number",
      "decimals": 2,
      "compact": false
    },
    "axis": "y",
    "color": { "type": "auto" }
  }
]
```

(The `axis` / format extras are separate injected defaults. The color replacement is the bug.)

A full Terraform acceptance repro is `TestAccResourceDashboardXYChart_layers` in [elastic/terraform-provider-elasticstack](https://github.com/elastic/terraform-provider-elasticstack).

## Expected

Either:

1. Persist `color: { type: static, color: "#54B399" }` on read-back (schema allows it), or
2. Reject the request with 400 if ES|QL XY Y metrics do not support static color (and remove `StaticColor` from the published union).

Silent rewrite on 201 breaks API clients that compare plan vs read-back (Terraform: `Provider produced inconsistent result after apply`).

## Actual

201 success; Y-metric static color replaced with `{type: auto}`. Same-layer `breakdown_by` colors persist.

## Why this looks like a Kibana bug

- OpenAPI/`kbapi` models `KibanaHTTPAPIsXyLayerESQL.Y.Color` as a union that includes `KibanaHTTPAPIsStaticColor`.
- The request is accepted, not schema-rejected.
- Only the Y-metric static color is dropped; breakdown colors on the same layer survive.

## Impact

Blocks Terraform provider acceptance `TestAccResourceDashboardXYChart_layers` on 9.6 against a practitioner-set ES|QL Y static color. The provider will **not** treat `{type:static}` as equivalent to `{type:auto}`.
