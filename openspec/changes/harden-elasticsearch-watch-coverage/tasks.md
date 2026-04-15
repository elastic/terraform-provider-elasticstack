## 1. Acceptance Coverage

- [ ] 1.1 Add watch acceptance coverage for import using the composite `<cluster_uuid>/<watch_id>` identifier
- [ ] 1.2 Add acceptance coverage for omitted `active` and omitted `throttle_period_in_millis` defaults
- [ ] 1.3 Add an acceptance step that removes a previously configured `transform` and verifies it stays absent after refresh

## 2. Watch State Synchronization Fix

- [ ] 2.1 Update watch read behavior so a missing API `transform` clears the Terraform state attribute
- [ ] 2.2 Keep the change scoped to existing watch behavior without altering resource type, ID format, or import format

## 3. Verification

- [ ] 3.1 Validate the OpenSpec change artifacts for `harden-elasticsearch-watch-coverage`
- [ ] 3.2 Run focused watch tests to confirm the new coverage and `transform` fix behave as specified
