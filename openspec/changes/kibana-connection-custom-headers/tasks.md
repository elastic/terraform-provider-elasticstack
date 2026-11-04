## 1. Users can configure custom headers and Kibana receives them

- [ ] 1.1 Expose the optional `headers` map through the shared Kibana connection schema, with sensitive,
      write-only, and ephemeral-state behavior matching each generated connection context
- [ ] 1.2 Carry configured headers from provider-level and entity-local connection blocks to GET, HEAD,
      and mutating Kibana requests despite endpoint environment precedence, while omitted or empty maps
      preserve existing request behavior
- [ ] 1.3 Preserve provider-managed authentication and `kbn-xsrf` behavior alongside custom headers,
      including the absence of `Authorization` when no provider authentication is configured
- [ ] 1.4 Retain entity-local managed-resource and data-source headers as sensitive state, exclude
      action headers from state, and preserve Terraform's ephemeral-state semantics
- [ ] 1.5 Honor custom headers across every supported Stack version without introducing a minimum
      Kibana or Stack version gate
- [ ] 1.6 Publish the generated Kibana connection documentation and changelog entry, documenting that
      existing debug formatting masks `Authorization` but does not generally redact arbitrary custom
      proxy-token headers

## 2. Reserved provider-managed headers fail before a Kibana request

- [ ] 2.1 Reject `authorization` and `kbn-xsrf` keys case-insensitively while building the affected
      connection, identifying the configured key in the diagnostic before that connection sends its
      first request
- [ ] 2.2 Preserve non-reserved custom headers while guaranteeing that an invalid provider-level or
      entity-local connection sends no request, regardless of reserved-key casing or diagnostic timing

## 3. Custom credentials remain within the initial Kibana origin

- [ ] 3.1 Allow same-origin redirects when custom headers are configured, preserving those headers on
      the redirected request
- [ ] 3.2 Reject every host, effective-port, or scheme change—including HTTPS-to-HTTP downgrades—before
      the redirect destination receives any request
- [ ] 3.3 Preserve existing redirect behavior for connections without custom headers

## 4. Fleet never receives Kibana-only custom headers

- [ ] 4.1 Exclude Kibana custom headers when deriving the Fleet client configuration, regardless of
      other Fleet connection settings
- [ ] 4.2 Ensure outgoing Fleet requests never contain a Kibana-only custom header
