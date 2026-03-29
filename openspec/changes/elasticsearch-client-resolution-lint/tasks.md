## 1. Analyzer enforcement

- [x] 1.1 Update `analysis/esclienthelper` to identify in-scope sinks in `internal/elasticsearch/**` for both `internal/clients/elasticsearch` function parameters and `*clients.APIClient` receiver calls.
- [x] 1.2 Enforce approved helper-derived client origins at sink usage, rejecting provider-meta casts, direct construction, and other bypass flows when provenance cannot be proven.
- [x] 1.3 Add or update wrapper handling so explicitly allowlisted functions and fact-proven wrappers or factories are accepted as helper-derived sources.

## 2. Lint integration

- [x] 2.1 Ensure the analyzer exports and imports provenance facts needed for interprocedural helper-derivation checks.
- [x] 2.2 Wire the analyzer into repository lint execution so violations fail `make check-lint`.
- [x] 2.3 Verify violation diagnostics identify non-helper-derived sink usage and point developers to the approved helper sources.

## 3. Regression coverage

- [x] 3.1 Add analyzer tests for compliant helper-derived usage at `internal/clients/elasticsearch` function-call sinks.
- [x] 3.2 Add analyzer tests for compliant and non-compliant `*clients.APIClient` receiver-call usage, including wrapper allowlist and fact-proven wrapper cases.
- [x] 3.3 Run targeted analyzer tests and repository lint checks to confirm compliant cases pass and violations fail as specified.
