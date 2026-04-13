## 1. Provider coverage tests

- [ ] 1.1 Add provider-level tests that enumerate covered `elasticstack_kibana_*` and `elasticstack_fleet_*` entities and assert `kibana_connection` presence.
- [ ] 1.2 Extend those tests to assert exact shared-helper equivalence and non-deprecated schema or block metadata for covered entities.

## 2. Lint enforcement

- [ ] 2.1 Add Kibana/Fleet client-resolution analyzer support that treats approved `kibana_connection` helpers as the required client source at in-scope sinks.
- [ ] 2.2 Wire the analyzer into `.golangci.yaml` and repository lint execution so violations fail `make check-lint`.

## 3. Regression coverage

- [ ] 3.1 Add analyzer tests for compliant helper-derived Kibana/Fleet client usage and for non-compliant bypass paths.
- [ ] 3.2 Run the relevant provider tests and lint checks to verify coverage enforcement and helper-usage diagnostics.
