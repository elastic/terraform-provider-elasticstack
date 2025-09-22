Creates or updates a Kibana synthetics monitor. See [API docs](https://www.elastic.co/guide/en/kibana/current/add-monitor-api.html)

## Supported monitor types
 * `http`
 * `tcp`
 * `icmp`
 * `browser`

The monitor type is determined by the fields in the `suite` block. See the [API docs](https://www.elastic.co/guide/en/kibana/current/add-monitor-api.html#add-monitor-api-request-body) for more details on which fields are required for each monitor type.

**NOTE:** Due-to nature of partial update API, reset values to defaults is not supported.
In case you would like to reset an optional monitor value, please set it explicitly or delete and create new monitor.