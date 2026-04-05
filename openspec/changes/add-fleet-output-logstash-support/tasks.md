## 1. Resource behavior for Logstash type

- [x] 1.1 Inspect `internal/fleet/output` request builders and add/complete `type = "logstash"` create/update payload mapping.
- [x] 1.2 Ensure read/state mapping handles Fleet `logstash` output responses without unknown-type diagnostics.
- [x] 1.3 Verify CRUD code paths (including space-aware context handling) operate correctly for Logstash outputs.

## 2. Tests

- [x] 2.1 Add or update unit tests for request/state mapping branches specific to `logstash` output type.
- [x] 2.2 Add acceptance coverage for Logstash output create/read/update behavior.
- [x] 2.3 Add or extend import verification for a Logstash output resource.
- [x] 2.4 Add test coverage for SSL-enabled and SSL-disabled Logstash configurations, including update transitions between SSL modes.

## 3. Documentation and validation

- [x] 3.1 Update Fleet output resource documentation/examples to include `logstash` configuration.
- [x] 3.2 Run targeted Fleet output tests and ensure new Logstash coverage passes.
- [x] 3.3 Run repository validation steps required for the change (`make build` and relevant lint/test checks).
