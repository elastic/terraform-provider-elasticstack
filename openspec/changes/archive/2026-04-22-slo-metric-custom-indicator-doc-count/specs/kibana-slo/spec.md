## ADDED Requirements

### Requirement: Mapping — `metric_custom_indicator` doc_count aggregation (REQ-035)

When a `metric_custom_indicator` metric entry uses `aggregation = "doc_count"`, the `field` attribute SHALL be optional and the provider SHALL NOT send `field` to the Kibana API. For all other aggregation types, `field` SHALL be provided by the user. The schema SHALL declare `field` as optional (not required) for `metric_custom_indicator.{good,total}.metrics`.

On write, when `aggregation = "doc_count"` the provider SHALL serialise the metric using the no-field API variant (`Metrics1`). For all other aggregations the provider SHALL use the field-bearing API variant (`Metrics0`). After a successful read-back, when the API returns a `doc_count` metric the provider SHALL store `field = null` in state.

This aligns `metric_custom_indicator` with the already-correct `timeslice_metric_indicator` behavior (REQ already present in that indicator's schema definition).

Schema change (replaces lines 83–88 and 93–98 of the schema block):

```hcl
    good {                                        # exactly 1 block
      equation = <required, string>
      metrics {                                   # at least 1 block
        name        = <required, string>
        aggregation = <required, string>
        field       = <optional, string>          # required for all aggregations except doc_count; must NOT be set for doc_count
        filter      = <optional, string>
      }
    }

    total {                                       # exactly 1 block
      equation = <required, string>
      metrics {                                   # at least 1 block
        name        = <required, string>
        aggregation = <required, string>
        field       = <optional, string>          # required for all aggregations except doc_count; must NOT be set for doc_count
        filter      = <optional, string>
      }
    }
```

#### Scenario: doc_count metric written without field

- **WHEN** a `metric_custom_indicator` good or total metric has `aggregation = "doc_count"` and `field` is not set
- **THEN** the provider SHALL send the metric to the Kibana API without a `field` key

#### Scenario: doc_count metric read back with null field

- **WHEN** the Kibana API returns a `metric_custom_indicator` metric with `aggregation = "doc_count"`
- **THEN** the provider SHALL store `field = null` in state for that metric

#### Scenario: non-doc_count metric still requires field

- **WHEN** a `metric_custom_indicator` metric has `aggregation != "doc_count"` and a non-null `field`
- **THEN** the provider SHALL send `field` in the API request as before
