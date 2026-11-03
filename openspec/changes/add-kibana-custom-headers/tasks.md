## 1. Provider-configured headers reach Kibana-derived requests

- [x] 1.1 Expose the sensitive provider-level Kibana `headers` string map and carry normalized values into Kibana and Fleet clients while preserving existing behavior for null, empty, or omitted maps.
- [x] 1.2 Send provider-configured custom headers on Kibana and derived Fleet requests while preserving native authentication precedence.
- [x] 1.3 Publish the provider-level custom-header contract in generated documentation.

## 2. Entity-local headers survive scoped client lifecycles

- [x] 2.1 Expose the same sensitive `headers` string map on entity-local `kibana_connection` and normalize null, empty, omitted, and configured values consistently with provider configuration.
- [x] 2.2 Carry entity-local headers through scoped Kibana and Fleet clients and preserve them across lifecycle snapshot restoration.
- [x] 2.3 Publish the entity-local custom-header contract in generated resource and data-source documentation.
