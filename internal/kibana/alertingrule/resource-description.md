Creates a Kibana rule. See the [create rule API documentation](https://www.elastic.co/guide/en/kibana/master/create-rule-api.html) for more details.

Execution status and date are available from the [`elasticstack_kibana_alerting_rules`](../data-sources/kibana_alerting_rules) data source.

**NOTE:** `api_key` authentication is only supported for alerting rule resources from version 8.8.0 of the Elastic stack. Using an `api_key` will result in an error message like:

```
Could not create API key - Unsupported scheme "ApiKey" for granting API Key
```
