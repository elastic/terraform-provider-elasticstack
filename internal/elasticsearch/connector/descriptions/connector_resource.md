Manages an [Elasticsearch content connector](https://www.elastic.co/docs/reference/search-connectors) via the [connector APIs](https://www.elastic.co/docs/api/doc/elasticsearch/group/endpoint-connector). Requires Elasticsearch **8.12.0** or later and the `manage_connector` cluster privilege for write operations (`monitor_connector` or `manage_connector` for read-only access).

## Lifecycle

Create and update use a narrow `POST /_connector` or `PUT /_connector/{connector_id}` envelope for identity fields, then fan out to per-aspect partial-update endpoints (`_pipeline`, `_scheduling`, `_features`, `_configuration`, `_name`, `_index_name`, `_service_type`, `_native`, `_api_key_id`) for everything else, and finish with `GET /_connector/{connector_id}` to refresh state. Delete calls `DELETE /_connector/{connector_id}` (404 is treated as success).

## Configuration values

`configuration_values` is written with `PUT /_connector/{connector_id}/_configuration`. The connector **service** must be running and have registered its per-`service_type` configuration schema before values can be applied—if `GET` returns an empty `configuration` object, the provider returns a structured error and does not call `_configuration`.

Removing a key from `configuration_values` stops Terraform from managing that field but does **not** unset it on the server.

Use `secret_value` (write-only, sensitive) for credentials. Drift is detected via a bcrypt hash in resource private state; after `terraform import`, the first refresh does not signal drift, and the first apply baselines the hash.

## Runtime telemetry

This resource omits read-only runtime fields (`status`, `last_synced`, filtering, and so on). Use the companion `data.elasticstack_elasticsearch_connector` data source to inspect live connector state.
