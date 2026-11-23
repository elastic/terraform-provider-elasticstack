## 1. Users can configure custom headers and Kibana receives them

- [x] 1.1 Expose the optional `headers` map through the shared Kibana connection schema, with sensitive,
      write-only, and ephemeral-state behavior matching each generated connection context
- [x] 1.2 Carry accepted configured headers from provider-level and entity-local connection blocks to
      GET, HEAD, and mutating Kibana requests despite endpoint environment precedence, while omitted or
      empty maps preserve existing request behavior
- [x] 1.3 Preserve provider-managed authentication and `kbn-xsrf` behavior alongside custom headers,
      including the absence of `Authorization` when no provider authentication is configured
- [x] 1.4 Retain entity-local managed-resource and data-source headers as sensitive state, exclude
      action headers from state, and preserve Terraform's ephemeral-state semantics
- [x] 1.5 Honor accepted custom headers across every supported Stack version without introducing a
      minimum Kibana or Stack version gate
- [x] 1.6 Publish the generated Kibana connection documentation and changelog entry, documenting that
      existing debug formatting masks `Authorization` but does not generally redact arbitrary custom
      proxy-token headers

## 2. Invalid custom header sets fail before a Kibana request

- [x] 2.1 Reject `authorization`, `kbn-xsrf`, `host`, `content-length`, `transfer-encoding`, and
      `trailer` keys case-insensitively while building the affected connection, identifying the
      configured key in the diagnostic before that connection sends its first request
- [x] 2.2 Reject multiple configured keys that differ only by case, identifying both conflicting keys
      before the affected connection sends its first request
- [x] 2.3 Preserve uniquely named, non-reserved custom headers while guaranteeing that an invalid
      provider-level or entity-local connection sends no request, regardless of key casing or
      diagnostic timing

## 3. Custom credentials remain within the initial Kibana origin

- [x] 3.1 Allow same-origin redirects when custom headers are configured, preserving those headers on
      the redirected request
- [x] 3.2 Reject every host, effective-port, or scheme change—including HTTPS-to-HTTP downgrades—before
      the redirect destination receives any request
- [x] 3.3 Preserve existing redirect behavior for connections without custom headers

## 4. Fleet never receives Kibana-only custom headers

- [x] 4.1 Exclude Kibana custom headers when deriving the Fleet client configuration, regardless of
      other Fleet connection settings
- [x] 4.2 Ensure outgoing Fleet requests never contain a Kibana-only custom header
