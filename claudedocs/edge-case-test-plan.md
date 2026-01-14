# Integration Policy Edge Case Test Plan

**Date**: 2026-01-14
**Purpose**: Comprehensive testing to ensure PR #1616 fix addresses all scenarios without regressions

## Critical Edge Cases to Test

### 1. vars_json Sanitization (THE FIX)

**Test Scenario**: Verify internal metadata is stripped before API calls

```hcl
# Test 1.1: vars_json with various data types
vars_json = jsonencode({
  string_var = "test"
  number_var = 123
  bool_var   = true
  array_var  = ["a", "b", "c"]
  object_var = {
    nested = "value"
  }
})
```

**Expected**: All values preserved, `__tf_provider_context` not sent to API

```hcl
# Test 1.2: Empty vars_json
vars_json = jsonencode({})
```

**Expected**: Works without errors

```hcl
# Test 1.3: vars_json with secrets (already tested in acc_test.go)
vars_json = jsonencode({
  secret_key = "sensitive"
})
```

**Expected**: Secrets properly handled, metadata stripped

### 2. Integration Version Changes

**Test Scenario**: Verify version updates work correctly

```hcl
# Test 2.1: Version upgrade (already in TestAccResourceIntegrationPolicy_VersionUpdate)
integration_version = "1.16.0" → "1.17.0"
```

**Expected**: Update succeeds, agent_policy_id preserved

```hcl
# Test 2.2: Version downgrade
integration_version = "1.17.0" → "1.16.0"
```

**Expected**: Update succeeds (if Kibana allows), agent_policy_id preserved

```hcl
# Test 2.3: Manual Kibana upgrade + Terraform reconcile (REAL BUG SCENARIO)
# 1. Create with Terraform
# 2. Manually upgrade in Kibana UI
# 3. Run terraform apply
```

**Expected**: No HTTP 400 error (this is what the fix addresses)

### 3. agent_policy_id vs agent_policy_ids Handling

**Test Scenario**: Verify framework handles these correctly without our intervention

```hcl
# Test 3.1: agent_policy_id only (standard case)
agent_policy_id = "policy-123"
```

**Expected**: Works, field preserved across updates

```hcl
# Test 3.2: agent_policy_ids only (8.15+ only)
agent_policy_ids = ["policy-123", "policy-456"]
```

**Expected**: Works on 8.15+, field preserved across updates

```hcl
# Test 3.3: Switching between fields (Tobio mentioned this causes error)
# Step 1: agent_policy_id = "policy-123"
# Step 2: Change to agent_policy_ids = ["policy-123"]
```

**Expected**: Proper error or handled by RequiresReplace modifier

### 4. inputs Configuration Edge Cases

**Test Scenario**: Verify inputs preservation logic works correctly

```hcl
# Test 4.1: No inputs configured (baseline)
# No inputs block
```

**Expected**: No inputs added by populateFromAPI

```hcl
# Test 4.2: Inputs configured
inputs = {
  "tcp-tcp" = {
    enabled = true
  }
}
```

**Expected**: Inputs preserved and updated correctly

```hcl
# Test 4.3: Inputs configured then removed
# Step 1: Has inputs block
# Step 2: Remove inputs block
```

**Expected**: Inputs removed, no inconsistent state error

### 5. Computed Fields Handling

**Test Scenario**: Verify UseStateForUnknown() plan modifiers work without manual preservation

```hcl
# Test 5.1: id and policy_id handling
# These should be automatically preserved by framework
```

**Expected**: Framework preserves them, no manual code needed

### 6. space_ids Handling (Multi-space support from PR #1390)

**Test Scenario**: Verify space-aware behavior works correctly

```hcl
# Test 6.1: No space_ids (single default space)
# No space_ids attribute
```

**Expected**: Works in default space

```hcl
# Test 6.2: Multiple spaces (if supported)
space_ids = ["default", "space-2"]
```

**Expected**: Policy visible in multiple spaces

### 7. output_id Handling (Version-dependent feature)

**Test Scenario**: Verify version constraints work

```hcl
# Test 7.1: output_id on 8.16+ (already tested in TestAccResourceIntegrationPolicyWithOutput)
output_id = "output-123"
```

**Expected**: Works on 8.16+

```hcl
# Test 7.2: output_id on < 8.16
output_id = "output-123"
```

**Expected**: Error with clear message about version requirement

## Regression Tests

### Test All Existing Acceptance Tests Pass

Run full acceptance test suite:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/fleet/integration_policy/...
```

**Expected**: All tests pass including:
- TestAccResourceIntegrationPolicyMultipleAgentPolicies
- TestAccResourceIntegrationPolicyWithOutput
- TestAccResourceIntegrationPolicy
- TestAccResourceIntegrationPolicySecretsFromSDK
- TestAccResourceIntegrationPolicySecrets
- TestAccIntegrationPolicyAzureMetrics
- TestAccIntegrationPolicyInputs
- TestAccResourceIntegrationPolicyGCPVertexAI
- TestAccResourceIntegrationPolicy_VersionUpdate (updated)

## Manual Testing with User Scenario

**Critical Test**: Reproduce user's exact error scenario

1. Create integration policy with Terraform
2. Manually upgrade integration version in Kibana UI
3. Run `terraform apply`

**Before Fix**: HTTP 400 error: `"Variable :__tf_provider_context not found"`
**After Fix**: Should reconcile successfully without errors

## Test Execution Plan

1. ✅ Run unit tests: `go test -v ./internal/fleet/integration_policy/`
2. ⏳ Run acceptance tests: `TF_ACC=1 go test -v -timeout 120m ./internal/fleet/integration_policy/`
3. ⏳ Manual edge case testing with Docker environment
4. ⏳ User validation in actual DCM5 environment

## Success Criteria

- ✅ All unit tests pass
- ⏳ All acceptance tests pass
- ⏳ No new test failures introduced
- ⏳ vars_json sanitization verified (no __tf_provider_context sent to API)
- ⏳ Version updates work correctly
- ⏳ No inconsistent state errors
- ⏳ User confirms fix works in their environment

## Known Limitations

- **Manual Kibana upgrade scenario**: Difficult to test in automated acceptance test (requires external UI interaction)
- **Multi-space behavior**: Requires specific Kibana configuration
- **Version-specific features**: Tests conditionally skipped on unsupported versions
