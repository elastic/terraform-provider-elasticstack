## 1. Watch read-state preservation

- [ ] 1.1 Add a watch-local helper that recursively merges API `actions` JSON with prior Terraform `actions` JSON, preserving only redacted string leaves when a prior concrete value exists at the same path
- [ ] 1.2 Update watch read/state mapping so refreshed `actions` uses the merged JSON while all other watch fields keep their existing read behavior
- [ ] 1.3 Keep imported or first-read watches without prior concrete `actions` values on the raw API response path for redacted leaves

## 2. Regression coverage

- [ ] 2.1 Add unit tests for the redacted-actions merge helper, including nested objects, arrays, mismatched paths, and no-prior-value cases
- [ ] 2.2 Extend watch acceptance coverage with a scenario where actions contain a redacted secret and an unrelated update still succeeds
- [ ] 2.3 Verify non-redacted action fields from the API continue to refresh into state when redacted leaves are being preserved

## 3. Spec and verification

- [ ] 3.1 Align the watch implementation with the new read/state preservation requirements for redacted `actions`
- [ ] 3.2 Validate the OpenSpec change artifacts for `preserve-redacted-watch-actions`
- [ ] 3.3 Run focused watch tests that exercise the new preservation behavior and address any failures
