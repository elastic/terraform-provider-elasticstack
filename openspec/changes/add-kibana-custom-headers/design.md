## Context

Kibana connections currently carry endpoint, authentication, and TLS settings through a canonical Terraform schema, typed configuration conversion, scoped client construction, and a shared Kibana/Fleet HTTP transport. Elasticsearch connections already support sensitive `map(string)` custom headers. Kibana-derived Fleet clients share the Kibana client configuration and transport, while entity-local connections replace provider-level connection values and are preserved in private lifecycle snapshots.

This change follows the existing Elasticsearch contract and the current Kibana client architecture. It introduces no dependency or new transport abstraction.

## Goals / Non-Goals

**Goals:**

- Expose the same sensitive `map(string)` header interface on provider-level and entity-local Kibana connections.
- Carry normalized headers through provider and scoped client construction.
- Apply headers consistently to generated and direct Kibana requests and to Fleet requests derived from the effective Kibana connection.
- Preserve native Kibana authentication precedence and lifecycle snapshot behavior.
- Cover schema, conversion, transport, inheritance, and snapshot behavior with focused tests.

**Non-Goals:**

- Add custom headers to the separate provider-level `fleet` block.
- Add a dedicated environment variable for Kibana headers.
- Support repeated values for one header name.
- Introduce a reusable transport abstraction or a new dependency.
- Change existing HTTP debug logging or add provider-wide secret redaction.

## Decisions

1. **Mirror the Elasticsearch connection contract.** The canonical Kibana connection specification will define `headers` as an optional sensitive map with string elements. Header names and values will be trimmed during configuration conversion, matching Elasticsearch behavior. Using the canonical connection specification keeps provider-level and entity-local block shapes aligned automatically.

2. **Carry headers in the shared Kibana client configuration.** Both provider-level and entity-local conversion paths will decode the Terraform map into the client configuration. Omitted or null maps produce no headers and preserve existing behavior. The existing Kibana-to-Fleet configuration conversion will carry the map to Fleet clients, preserving the current unified connection semantics.

3. **Apply headers in the existing shared HTTP transport.** Header injection belongs in the transport used by both generated OpenAPI requests and direct HTTP requests. The transport will set configured custom headers before resolving native authentication precedence. If native username/password, API key, or bearer-token authentication is present, its `Authorization` value wins; otherwise a custom `Authorization` value remains.

4. **Preserve headers in lifecycle snapshots.** The Kibana connection snapshot will serialize and restore normalized headers so a client reconstructed for a later lifecycle operation uses the same effective connection.

5. **Test each integration boundary.** Schema tests will verify the sensitive string-map contract; configuration tests will verify normalization, null handling, provider/scoped conversion, and Fleet inheritance; transport tests will verify request injection and authentication precedence; snapshot tests will verify round trips. Generated documentation, OpenSpec validation, targeted Go tests, lint, and build checks provide the cumulative gate.

The existing shared transport remains the universal request boundary, and Fleet continues to derive its request configuration from the effective Kibana connection.

## Risks / Trade-offs

- **Kibana headers are inherited by Fleet clients derived from that connection** → This is intentional and matches existing unified Kibana/Fleet connection semantics; a distinct Fleet header interface remains out of scope.
- **A custom `Authorization` value can conflict with native authentication** → Native authentication deterministically wins when configured, matching the established authentication contract.
- **A string map cannot represent repeated HTTP header values** → This intentionally matches Elasticsearch and covers the access-proxy authentication use cases in issue #440.
- **Sensitive headers may appear in debug HTTP dumps** → This matches current Elasticsearch and Kibana authentication logging behavior; comprehensive redaction requires a separate provider-wide change.
- **Existing state does not contain the new map** → The optional attribute decodes as null or empty, so existing configurations and state retain current behavior without migration.
