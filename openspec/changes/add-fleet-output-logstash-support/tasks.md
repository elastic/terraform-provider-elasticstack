## 1. Resource behavior for Logstash type

- [ ] 1.1 Inspect `internal/fleet/output` request builders and add/complete `type = "logstash"` create/update payload mapping.
- [ ] 1.2 Ensure read/state mapping handles Fleet `logstash` output responses without unknown-type diagnostics.
- [ ] 1.3 Verify CRUD code paths (including space-aware context handling) operate correctly for Logstash outputs.

## 2. Tests

- [ ] 2.1 Add or update unit tests for request/state mapping branches specific to `logstash` output type.
- [ ] 2.2 Add acceptance coverage for Logstash output create/read/update behavior.
- [ ] 2.3 Add or extend import verification for a Logstash output resource.

## 3. Documentation and validation

- [ ] 3.1 Update Fleet output resource documentation/examples to include `logstash` configuration.
- [ ] 3.2 Run targeted Fleet output tests and ensure new Logstash coverage passes.
- [ ] 3.3 Run repository validation steps required for the change (`make build` and relevant lint/test checks).
