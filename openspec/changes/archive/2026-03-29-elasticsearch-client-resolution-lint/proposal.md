## Why

Elasticsearch entities currently rely on convention to resolve `*clients.APIClient` values before calling Elasticsearch helpers and client methods. That leaves room for bypass paths such as provider-meta casts or ad-hoc construction, which can silently undermine connection-handling consistency across SDK and Framework implementations.

## What Changes

- Introduce a new OpenSpec capability for provider lint rules that enforce helper-derived Elasticsearch client resolution at concrete sink call sites.
- Define the approved client-source helpers, wrapper policies, sink scope, provenance-fact behavior, and diagnostic expectations for the analyzer implemented in `analysis/esclienthelper`.
- Require repository lint execution and analyzer tests to enforce the rule in local development and CI.

## Capabilities

### New Capabilities
- `elasticsearch-client-resolution-lint`: Provider lint requirements for proving that Elasticsearch client values used by entity code come from approved helper-derived sources.

### Modified Capabilities
- _(none)_

## Impact

- **Specs**: new capability under `openspec/changes/elasticsearch-client-resolution-lint/specs/elasticsearch-client-resolution-lint/spec.md`
- **Analyzer implementation**: `analysis/esclienthelper`
- **Lint wiring**: `golangci-lint` integration and `make check-lint`
- **Tests**: analyzer regression coverage for compliant and non-compliant sink usage
