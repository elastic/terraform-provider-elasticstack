## 1. Resource schema and model updates

- [x] 1.1 Extend `elasticstack_fleet_output` type validation to accept `remote_elasticsearch` and add schema fields for required service-token auth plus optional TLS/mTLS settings.
- [x] 1.2 Mark remote auth/key material as sensitive and configure plan/state behavior for fields that may be redacted by Fleet read APIs.
- [x] 1.3 Update internal output models and request builders so create/update calls include remote Elasticsearch-specific payload fields.
- [x] 1.4 Expose `sync_integrations`, `sync_uninstalled_integrations`, and `write_to_logs_streams` for `remote_elasticsearch`, with conditional schema validation and API mapping; extend docs and examples.

## 2. CRUD/state mapping behavior

- [x] 2.1 Extend read/type-dispatch logic to map `OutputRemoteElasticsearch` responses into Terraform state for both common and remote-specific fields.
- [x] 2.2 Implement secret-preserving read behavior for remote output fields when Fleet omits or redacts stored secret values.
- [x] 2.3 Ensure create/read/update/delete operations for remote outputs preserve existing identity and Kibana space-context behavior.

## 3. Test coverage

- [x] 3.1 Add/extend unit tests for schema validation (accepted type, required remote auth, TLS/mTLS combinations, sensitive attributes).
- [x] 3.2 Add/extend unit tests for model conversion and read mapping of `OutputRemoteElasticsearch`, including redacted-secret scenarios.
- [x] 3.3 Add/extend acceptance tests for remote Elasticsearch output lifecycle and update behavior where environment support is available.

## 4. Documentation and verification

- [x] 4.1 Update `docs/resources/fleet_output.md` with `remote_elasticsearch` examples and field descriptions, including remote-output limitations/prerequisites.
- [x] 4.2 Run targeted tests and provider build checks (`go test` for touched packages and `make build`) and fix regressions.
- [x] 4.3 Verify the OpenSpec change is apply-ready (`openspec status --change fleet-output-remote-elasticsearch`) and adjust artifacts if validation feedback requires it.
